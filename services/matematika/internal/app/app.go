package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/internal/cache"
	"github.com/IbadT/business_bank_back/services/matematika/internal/database"
	"github.com/IbadT/business_bank_back/services/matematika/internal/models"
	"github.com/IbadT/business_bank_back/services/matematika/internal/repository"
	"github.com/IbadT/business_bank_back/services/matematika/internal/service"
	httptransport "github.com/IbadT/business_bank_back/services/matematika/internal/transport/http"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/redis"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"

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
	generatorService service.GeneratorService
	httpHandler      *httptransport.Handler
	echo             *echo.Echo
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
		log.Printf("Warning: Failed to initialize Redis: %v", err)
		log.Println("Redis will not be available, cache will be disabled")
		// Не прерываем запуск приложения, если Redis недоступен
	}

	// 4. Инициализация зависимостей (Repository -> Service -> Handler)
	a.initDependencies()

	// 5. Инициализация HTTP сервера
	a.initHTTPServer()

	// 6. Graceful shutdown
	return a.startServer()
}

// initEnvironment загружает переменные окружения
func (a *App) initEnvironment() error {
	configLoaded := true
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
		configLoaded = false
	}

	if configLoaded {
		os.Setenv("CONFIG_LOADED", "true")
	}

	return nil
}

// initDatabase инициализирует подключение к базе данных и выполняет миграции
func (a *App) initDatabase() error {
	db, err := database.InitDB()
	if err != nil {
		return err
	}
	a.db = db
	log.Println("✓ Database connected successfully")

	// Автоматическая миграция моделей
	if err := db.AutoMigrate(
		&models.User{},
		&models.GenerationRequest{},
		&models.GeneratedTransaction{},
		&models.FinancialSummaryDB{},
		&models.DailyBalanceV2{},
		&models.TransactionTemplateDB{},
		&models.DefaultCustomerDB{},
		&models.Holiday{},
		&models.GenerationState{},
		&models.UserGateway{},
	); err != nil {
		log.Printf("❌ Failed to migrate models: %v", err)
		return err
	}

	log.Println("✓ Database migrations completed")
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
	breakdownService := service.NewBreakdownService(transactionRepo)

	// TransactionService (нужен для BalanceAdjustmentService)
	transactionService := service.NewTransactionService(transactionRepo)

	// BalanceAdjustmentService (нужен для GeneratorService)
	balanceAdjustmentService := service.NewBalanceAdjustmentService(transactionRepo, transactionService, generationRequestRepo)

	// HolidayService (нужен для GeneratorService)
	cacheRepo := cache.NewRepository(a.redis)
	cacheService := cache.New(cacheRepo)
	holidayService := service.NewHolidayService(holidayRepo, cacheService)

	// GatewayService (нужен для GeneratorService)
	gatewayService := service.NewGatewayService(gatewayRepo, configRepo)

	// BaseAmountService (нужен для GeneratorService)
	baseAmountService := service.NewBaseAmountService(stateRepo)

	// GeneratorService
	genService, err := service.NewGeneratorService(configRepo, stateRepo, userRepo, holidayRepo, gatewayRepo, holidayService, gatewayService, baseAmountService, breakdownService, balanceAdjustmentService, generationRequestRepo, transactionRepo)
	if err != nil {
		log.Printf("Warning: Failed to initialize GeneratorService: %v", err)
		log.Println("GeneratorService will not be available")
	} else {
		a.generatorService = genService
		log.Println("✓ GeneratorService initialized successfully")
	}

	// UserService
	userService := service.NewUserService(userRepo)

	// SeedService
	seedService := service.NewSeedService(a.db)

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
	log.Println("✓ Redis connected successfully")
	return nil
}

// startServer запускает HTTP сервер и обрабатывает graceful shutdown
func (a *App) startServer() error {
	port := a.config.Port

	if port == "" {
		port = database.GetEnv("PORT", "8080")
	}

	// Запускаем HTTP сервер в goroutine
	go func() {
		log.Printf("✓ HTTP server starting on port %s", port)
		if err := a.echo.Start(":" + port); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// go func() {
	// 	log.Printf("✓ GRPC server starting on port %s", port)
	// 	if err := a.grpcServer.Serve(grpcPort); err != nil && err != grpc.ErrServerStopped {
	// 		log.Fatalf("GRPC server error: %v", err)
	// 	}
	// }()

	// Канал для сигналов остановки
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Ждем сигнал остановки
	<-quit

	log.Println("Shutting down server...")

	// Контекст с таймаутом для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Останавливаем HTTP сервер
	if err := a.echo.Shutdown(ctx); err != nil {
		log.Printf("Error during HTTP server shutdown: %v", err)
	}

	log.Println("✓ Server stopped gracefully")
	return nil
}
