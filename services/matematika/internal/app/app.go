package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/internal/cache"
	"github.com/IbadT/business_bank_back/services/matematika/internal/database"
	"github.com/IbadT/business_bank_back/services/matematika/internal/repository"
	balanceservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/balance"
	baseamountservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/base"
	breakdownservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/breakdown"
	gatewayservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/gateway"
	generatorservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/generator"
	holidayservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/holiday"
	seedservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/seed"
	transactionservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/transaction"
	userservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/user"
	transportgrpc "github.com/IbadT/business_bank_back/services/matematika/internal/transport/grpc"
	httptransport "github.com/IbadT/business_bank_back/services/matematika/internal/transport/http"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/redis"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	// "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Config - конфигурация приложения
type Config struct {
	Port     string
	GRPCPort string
}

// App - основное приложение
type App struct {
	config           *Config
	db               *gorm.DB
	redis            *redis.RDS
	generatorService generatorservice.GeneratorService
	httpHandler      *httptransport.Handler
	grpcHandler      *transportgrpc.Handler
	echo             *echo.Echo
	grpcServer       *grpc.Server
}

// NewApp создает новое приложение с заданной конфигурацией
func NewApp(cfg *Config) *App {
	return &App{
		config: cfg,
	}
}

// Run запускает приложение
func (a *App) Run() error {
	// 1. Инициализация окружения
	if err := a.initEnvironment(); err != nil {
		return err
	}

	// 2. Инициализация базы данных
	if err := a.initDatabase(); err != nil {
		return err
	}

	// 3. Инициализация Redis
	if err := a.initRedis(); err != nil {
		logrus.WithError(err).Warn("Warning: Failed to initialize Redis")
		logrus.Warn("Redis will not be available, cache will be disabled")
		// Не прерываем запуск приложения, если Redis недоступен
	}

	// 4. Инициализация зависимостей (Repository -> Service -> Handler)
	a.initDependencies()

	// 5. Инициализация HTTP сервера
	a.initHTTPServer()

	// 6. Запуск серверов
	return a.startServers()
}

// initEnvironment загружает переменные окружения
func (a *App) initEnvironment() error {
	configLoaded := true
	if err := godotenv.Load(); err != nil {
		logrus.Info("No .env file found, using system environment variables")
		configLoaded = false
	}

	if configLoaded {
		os.Setenv("CONFIG_LOADED", "true")
	}

	return nil
}

// initDatabase инициализирует подключение к базе данных
// Миграции выполняются вручную через команду: make migrate-up-matematika
func (a *App) initDatabase() error {
	db, err := database.InitDB()
	if err != nil {
		return err
	}
	a.db = db
	logrus.Info("✓ Database connected successfully")

	// Автоматические миграции отключены для контроля версий схемы БД
	// Используйте: make migrate-up-matematika для применения миграций из migrations/
	return nil
}

// initDependencies инициализирует зависимости (Repository -> Service -> Handler)
func (a *App) initDependencies() {
	// ConfigRepository для GeneratorService
	configPath := database.GetEnv("CONFIG_PATH", "./config")
	configRepo := repository.NewConfigRepository(configPath)

	// ========================= REPOSITORIES =========================

	// StateRepository для сохранения состояния между генерациями
	stateRepo := repository.NewStateRepository(a.db)

	// HolidayRepository для HolidayService
	holidayRepo := repository.NewHolidayRepository(a.db)

	// TransactionRepository для TransactionService
	transactionRepo := repository.NewTransactionRepository(a.db)

	// UserRepository
	userRepo := repository.NewUserRepository(a.db)

	// GatewayRepository
	gatewayRepo := repository.NewGatewayRepository(a.db)

	// GenerationRequestRepository для сохранения запросов генерации
	generationRequestRepo := repository.NewGenerationRequestRepository(a.db)

	// ========================= SERVICES =========================
	// BreakdownService (нужен для GeneratorService)
	breakdownService := breakdownservice.NewBreakdownService(transactionRepo)

	// TransactionService (нужен для BalanceAdjustmentService)
	transactionService := transactionservice.NewTransactionService(transactionRepo)

	// BalanceAdjustmentService (нужен для GeneratorService)
	balanceAdjustmentService := balanceservice.NewBalanceAdjustmentService(transactionRepo, transactionService, generationRequestRepo)

	// HolidayService (нужен для GeneratorService)
	cacheRepo := cache.NewRepository(a.redis)
	cacheService := cache.New(cacheRepo)
	holidayService := holidayservice.NewHolidayService(holidayRepo, cacheService)

	// GatewayService (нужен для GeneratorService)
	gatewayService := gatewayservice.NewGatewayService(gatewayRepo, configRepo, cacheService)

	// BaseAmountService (нужен для GeneratorService)
	baseAmountService := baseamountservice.NewBaseAmountService(stateRepo)

	// GeneratorService
	genService, err := generatorservice.NewGeneratorService(configRepo, stateRepo, userRepo, holidayRepo, gatewayRepo, holidayService, gatewayService, baseAmountService, breakdownService, balanceAdjustmentService, generationRequestRepo, transactionRepo)
	if err != nil {
		logrus.WithError(err).Warn("Warning: Failed to initialize GeneratorService")
		logrus.Warn("GeneratorService will not be available")
	} else {
		a.generatorService = genService
		logrus.Info("✓ GeneratorService initialized successfully")
	}

	// UserService
	userService := userservice.NewUserService(userRepo)

	// SeedService
	seedService := seedservice.NewSeedService(a.db)

	// ========================= HTTP TRANSPORT HANDLER =========================
	// HTTP Transport Handler
	httpHandler := httptransport.NewHandler(seedService,
		a.generatorService,
		userService,
		holidayService,
		transactionService,
		gatewayService,
		breakdownService,
		baseAmountService,
		balanceAdjustmentService,
	)
	a.httpHandler = httpHandler

	// ========================= gRPC TRANSPORT HANDLER =========================
	// gRPC Transport Handler
	grpcHandler := transportgrpc.NewHandler(
		a.generatorService,
		userService,
		transactionService,
		holidayService,
		gatewayService,
		baseAmountService,
		breakdownService,
		balanceAdjustmentService,
	)
	a.grpcHandler = grpcHandler
}

// initHTTPServer инициализирует HTTP сервер
func (a *App) initHTTPServer() {
	e := a.httpHandler.Init()
	a.echo = e
}

// initRedis инициализирует подключение к Redis
func (a *App) initRedis() error {
	redisClient := database.InitRedis()

	// Проверяем подключение
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		return err
	}

	// Создаем RDS обертку из готового клиента
	a.redis = redis.NewFromClient(redisClient)
	logrus.Info("✓ Redis connected successfully")
	return nil
}

// startServers запускает HTTP и gRPC серверы и обрабатывает graceful shutdown
func (a *App) startServers() error {
	httpPort := a.config.Port
	if httpPort == "" {
		httpPort = database.GetEnv("PORT", "8080")
	}

	grpcPort := a.config.GRPCPort
	if grpcPort == "" {
		grpcPort = database.GetEnv("GRPC_PORT", "9090")
	}

	// Запускаем HTTP сервер в goroutine
	go func() {
		logrus.Infof("✓ HTTP server starting on port %s", httpPort)
		if err := a.echo.Start(":" + httpPort); err != nil && err != http.ErrServerClosed {
			logrus.WithError(err).Fatal("HTTP server error")
		}
	}()

	// Запускаем gRPC сервер
	logrus.Infof("✓ gRPC server starting on port %s", grpcPort)
	grpcServer, err := transportgrpc.RunGRPCServer(a.grpcHandler, grpcPort)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to start gRPC server")
	}
	a.grpcServer = grpcServer

	// Канал для сигналов остановки
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Ждем сигнал остановки
	<-quit

	logrus.Info("Shutting down servers...")

	// Контекст с таймаутом для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Останавливаем HTTP сервер
	if err := a.echo.Shutdown(ctx); err != nil {
		logrus.WithError(err).Error("Error during HTTP server shutdown")
	}

	// Останавливаем gRPC сервер
	if a.grpcServer != nil {
		a.grpcServer.GracefulStop()
		logrus.Info("✓ gRPC server stopped gracefully")
	}

	logrus.Info("✓ All servers stopped gracefully")
	return nil
}
