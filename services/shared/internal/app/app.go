package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log"

	"github.com/IbadT/business_bank_back/services/shared/internal/database"
	httptransport "github.com/IbadT/business_bank_back/services/shared/internal/transport/http"
	"github.com/IbadT/business_bank_back/services/shared/pkg/kafka"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Config struct {
	Port string
	kafkaBrokers []string
	GRPCPort string
}

type app struct {
	config *Config
	db *gorm.DB
	kafkaProducer kafka.Producer
	// sevice
	httpHandler *httptransport.Handler
	echo *echo.Echo
}

type App interface {
	Run() error
	initEnvironment() error
	initDatabase() error
	initKafka() error
	initDependencies()
	startKafkaConsumer()
	initHTTPServer()
	startServer() error
}

func NewApp(cfg *Config) *app {
	return &app{
		config: cfg,
	}
}

func (a *app) Run() error {
	// 1. Инициализация окружения
	if err := a.initEnvironment(); err != nil {
		return err
	}

	if err := a.initDatabase(); err != nil {
		return err
	}

	if err := a.initKafka(); err != nil {
		return err
	}

	a.initDependencies()

	a.startKafkaConsumer()

	a.initHTTPServer()

	return a.startServer()
}

func (a *app) initEnvironment() error {
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

func (a *app) initDatabase() error {
	db, err := database.InitDB()
	if err != nil {
		return err
	}

	a.db = db
	log.Println("✓ Database connected successfully")

	// Автоматическая миграция моделей
	if err := db.AutoMigrate(); err != nil {
		return err
	}

	log.Println("✓ Database migrations completed")
	return nil
}

func (a *app) initKafka() error {

	return nil
}

func (a *app) initDependencies() {
	return
}

func (a *app) startKafkaConsumer() {
	go func() {

	}()
}

func (a *app) initHTTPServer() {
	e := a.httpHandler.Init()
	a.echo = e
}

func (a *app) startServer() error {
	port := a.config.Port

	if port == "" {
		port = database.GetEnv("PORT", "8083")
	}

	go func() {
		log.Printf("✓ HTTP server starting on port %s", port)
		if err := a.echo.Start(":" + port); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// ДОБАВИТЬ GRPC SERVER

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.echo.Shutdown(ctx); err != nil {
		log.Printf("Error during HTTP server shutdown: %v", err)
	}

	if a.kafkaProducer != nil {
		if err := a.kafkaProducer.Close(); err != nil {
			log.Printf("Error closing Kafka producer: %v", err)
		}
	}

	log.Println("✓ Server stopped gracefully")
	return nil
}