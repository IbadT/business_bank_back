package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/internal/database"
	"github.com/IbadT/business_bank_back/services/matematika/internal/models"
	"github.com/IbadT/business_bank_back/services/matematika/internal/repository"
	"github.com/IbadT/business_bank_back/services/matematika/internal/service"
	httptransport "github.com/IbadT/business_bank_back/services/matematika/internal/transport/http"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Config - конфигурация приложения
type Config struct {
	Port     string
	GRPCPort string
}

// App - основное приложение
type App struct {
	config         *Config
	db             *gorm.DB
	generatorService service.GeneratorService
	httpHandler    *httptransport.Handler
	echo           *echo.Echo
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

	// 3. Инициализация зависимостей (Repository -> Service -> Handler)
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

	// Применяем миграцию для переименования password_hash -> password перед AutoMigrate
	migrationSQL := `
	DO $$
	BEGIN
		IF EXISTS (
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_schema = 'public'
			AND table_name = 'users' 
			AND column_name = 'password_hash'
		) THEN
			IF EXISTS (
				SELECT 1 
				FROM information_schema.columns 
				WHERE table_schema = 'public'
				AND table_name = 'users' 
				AND column_name = 'password'
			) THEN
				ALTER TABLE users DROP COLUMN password;
			END IF;
			ALTER TABLE users RENAME COLUMN password_hash TO password;
			RAISE NOTICE 'Renamed password_hash to password';
		END IF;
	END $$;
	`
	if err := db.Exec(migrationSQL).Error; err != nil {
		log.Printf("⚠️  Migration warning (may be already applied): %v", err)
	}

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
	
	// StateRepository для сохранения состояния между генерациями
	stateRepo := repository.NewStateRepository(a.db)

	// HolidayRepository для HolidayService
	holidayRepo := repository.NewHolidayRepository(a.db)
	
	// GeneratorService
	genService, err := service.NewGeneratorService(configRepo, stateRepo, holidayRepo)
	if err != nil {
		log.Printf("Warning: Failed to initialize GeneratorService: %v", err)
		log.Println("GeneratorService will not be available")
	} else {
		a.generatorService = genService
		log.Println("✓ GeneratorService initialized successfully")
	}

	// UserRepository
	userRepo := repository.NewUserRepository(a.db)
	// UserService
	userService := service.NewUserService(userRepo)

	// HolidayService
	holidayService := service.NewHolidayService(holidayRepo)
	// HTTP Transport Handler
	httpHandler := httptransport.NewHandler(a.generatorService, userService, holidayService)
	a.httpHandler = httpHandler
}

// initHTTPServer инициализирует HTTP сервер
func (a *App) initHTTPServer() {
	e := a.httpHandler.Init()
	a.echo = e
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
