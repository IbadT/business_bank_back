package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/internal/calculation"
	"github.com/IbadT/business_bank_back/services/matematika/internal/database"
	"github.com/IbadT/business_bank_back/services/matematika/internal/kafka"
	"github.com/joho/godotenv"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	// ========================================================================
	// 1. ИНИЦИАЛИЗАЦИЯ ОКРУЖЕНИЯ
	// ========================================================================

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// ========================================================================
	// 2. ИНИЦИАЛИЗАЦИЯ БАЗЫ ДАННЫХ
	// ========================================================================

	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	log.Println("✓ Database connected successfully")

	// ========================================================================
	// 3. ИНИЦИАЛИЗАЦИЯ KAFKA PRODUCER
	// ========================================================================

	// Получаем Kafka брокеры из env (формат: "broker1:9092,broker2:9092")
	kafkaBrokers := strings.Split(database.GetEnv("KAFKA_BROKERS", "localhost:9092"), ",")

	// Создаем producer с production настройками
	producerConfig := kafka.DefaultProducerConfig(kafkaBrokers)
	kafkaProducer, err := kafka.NewProducer(producerConfig, log.Default())
	if err != nil {
		log.Fatalf("Failed to initialize Kafka producer: %v", err)
	}
	defer kafkaProducer.Close() // Закрываем при завершении
	log.Println("✓ Kafka producer connected successfully")

	// ========================================================================
	// 4. DEPENDENCY INJECTION (Repository -> Service -> Handler)
	// ========================================================================

	// Repository - работа с БД
	calcRepo := calculation.NewCalculationRepository(db)

	// Validator - валидация входных данных
	validator := calculation.NewRequestValidator()

	// Service - бизнес-логика + Kafka producer (ЗДЕСЬ подключаем Kafka!)
	calcService := calculation.NewCalculationServiceWithKafka(calcRepo, kafkaProducer)

	// Handler - HTTP обработчики
	calcHandler := calculation.NewCalculationHandler(calcService, validator)

	// ========================================================================
	// 5. ЗАПУСК KAFKA CONSUMER (в отдельной goroutine)
	// ========================================================================

	// Consumer будет слушать топик и выводить сообщения в консоль
	go func() {
		// Даем время Kafka полностью запуститься
		time.Sleep(5 * time.Second)

		ctx := context.Background()
		if err := calcService.StartConsumer(ctx); err != nil {
			log.Printf("Consumer error: %v", err)
		}
	}()

	// ========================================================================
	// 6. НАСТРОЙКА HTTP СЕРВЕРА (Echo)
	// ========================================================================

	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Routes
	e.GET("/health", calcHandler.HealthCheck)
	e.POST("/generate-statement", calcHandler.GenerateStatement)
	e.GET("/statement/:id/status", calcHandler.GetStatementStatusByID)
	e.GET("/statement/:id/result", calcHandler.GetStatementResultByID)

	// ========================================================================
	// 7. GRACEFUL SHUTDOWN
	// ========================================================================

	// Запускаем HTTP сервер в goroutine
	port := database.GetEnv("PORT", "8080")
	go func() {
		log.Printf("✓ HTTP server starting on port %s", port)
		if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

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
	if err := e.Shutdown(ctx); err != nil {
		log.Printf("Error during HTTP server shutdown: %v", err)
	}

	// Kafka producer закроется через defer
	log.Println("✓ Server stopped gracefully")
}
