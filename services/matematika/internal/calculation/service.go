package calculation

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/IbadT/business_bank_back/services/matematika/internal/database"
	"github.com/IbadT/business_bank_back/services/matematika/internal/helpers"
	"github.com/IbadT/business_bank_back/services/matematika/internal/kafka"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================================================
// ИНТЕРФЕЙС СЕРВИСА
// ============================================================================

// CalculationService - интерфейс бизнес-логики расчетов
// Определяет контракт для работы с выписками
type CalculationService interface {
	// HealthCheck проверяет здоровье сервиса
	// Параметры:
	//   - ctx: Контекст запроса
	//
	// Возвращает HealthCheckResponse если сервис готов к работе
	HealthCheck(ctx context.Context) (*helpers.HealthCheckResponse, error)

	// GenerateStatement генерирует новую выписку и публикует результат в Kafka
	GenerateStatementToKafka(ctx context.Context, req *helpers.GenerateStatementRequest) (*helpers.GenerateStatementResponse, error)

	// GenerateStatement(ctx context.Context, req *helpers.StatementStateRequest) (*helpers.GenerateStatementResponse, error)
	GenerateStatement(ctx context.Context, req *helpers.StatementStateRequest) (*GenerateStatementResponse, error)

	// GetStatementStatusByID получает статус выписки по ID
	GetStatementStatusByID(ctx context.Context, id string) (interface{}, error)

	// GetStatementResultByID получает результаты расчетов по ID
	GetStatementResultByID(ctx context.Context, id string) (interface{}, error)

	// StartConsumer запускает Kafka consumer для чтения сообщений
	StartConsumer(ctx context.Context) error

	// GetAdminConfig получает конфигурацию системы
	GetAdminConfig(ctx context.Context) (*helpers.AdminConfigResponse, error)

	// GetTransactions получает список транзакций
	GetTransactions(ctx context.Context) ([]Transaction, error)

	// GetBusinessRules получает список бизнес-правил
	GetBusinessRules(ctx context.Context) ([]BusinessRule, error)

	// GetDailyBalances получает список дневных балансов
	GetDailyBalances(ctx context.Context) ([]DailyBalance, error)

	// GetStatements получает список выписок
	GetStatements(ctx context.Context) ([]Statement, error)
}

// ============================================================================
// РЕАЛИЗАЦИЯ СЕРВИСА
// ============================================================================

// calculationService - конкретная реализация CalculationService
// Содержит зависимости:
// - calcRepo: для работы с БД (Repository pattern)
// - kafkaProducer: для публикации событий в Kafka (Event-driven architecture)
type calculationService struct {
	calcRepo      CalculationRepository // Repository для доступа к данным
	kafkaProducer kafka.Producer        // Kafka producer для публикации событий
	kafkaBrokers  []string              // Kafka брокеры
	db            *gorm.DB              // База данных
	config        *ServiceConfig        // Конфигурация сервиса
}

// ServiceConfig - конфигурация для health check
type ServiceConfig struct {
	Version      string // Версия сервиса
	ServiceName  string // Имя сервиса
	ConfigLoaded bool   // Загружен ли конфиг
}

// NewCalculationService создает новый сервис БЕЗ Kafka (для обратной совместимости)
// DEPRECATED: Используйте NewCalculationServiceWithKafka для production
func NewCalculationService(calcRepo CalculationRepository, kafkaProducer kafka.Producer, kafkaBrokers []string, db *gorm.DB) CalculationService {
	return &calculationService{
		calcRepo:      calcRepo,
		kafkaProducer: kafkaProducer,
		kafkaBrokers:  kafkaBrokers,
		db:            db,
		config:        loadServiceConfig(),
	}
}

// NewCalculationServiceWithKafka создает новый сервис С Kafka producer
// Рекомендуемый способ для production
// Параметры:
//   - calcRepo: Repository для работы с БД
//   - kafkaProducer: Producer для публикации событий в Kafka
//   - kafkaBrokers: Kafka брокеры
//   - db: База данных
//
// Возвращает готовый к использованию сервис
func NewCalculationServiceWithKafka(calcRepo CalculationRepository, kafkaProducer kafka.Producer, kafkaBrokers []string, db *gorm.DB) CalculationService {
	return &calculationService{
		calcRepo:      calcRepo,
		kafkaProducer: kafkaProducer, // Внедряем Kafka через Dependency Injection
		kafkaBrokers:  kafkaBrokers,
		db:            db,
		config:        loadServiceConfig(),
	}
}

// loadServiceConfig загружает конфигурацию сервиса из переменных окружения
func loadServiceConfig() *ServiceConfig {
	// Проверяем, загружен ли .env файл (через проверку переменных)
	configLoaded := os.Getenv("CONFIG_LOADED") == "true" || os.Getenv("POSTGRES_HOST") != ""

	return &ServiceConfig{
		Version:      database.GetEnv("SERVICE_VERSION", "1.0.0"),
		ServiceName:  database.GetEnv("SERVICE_NAME", "matematika"),
		ConfigLoaded: configLoaded,
	}
}

// ============================================================================
// МЕТОДЫ СЕРВИСА
// ============================================================================

// HealthCheck проверяет здоровье сервиса
// Параметры:
//   - ctx: Контекст запроса
//
// Возвращает error если сервис не готов к работе
func (s *calculationService) HealthCheck(ctx context.Context) (*helpers.HealthCheckResponse, error) {
	// Проверяем подключение к базе данных
	dbStatus := database.HealthCheckDB(s.db)

	// Проверяем подключение к Kafka
	kafkaStatus := kafka.HealthCheckKafka(s.kafkaBrokers)

	// Проверяем подключение к Redis
	redisStatus := database.HealthCheckRedis()

	// Определяем общий статус на основе зависимостей
	status := "healthy"
	if dbStatus == "disconnected" || kafkaStatus == "disconnected" {
		status = "degraded"
	}

	return &helpers.HealthCheckResponse{
		Status:    status,
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   s.config.Version,
		Dependencies: helpers.HealthCheckDependencies{
			Kafka:        kafkaStatus,
			Database:     dbStatus,
			Redis:        redisStatus,
			ConfigLoaded: s.config.ConfigLoaded,
			Service:      s.config.ServiceName,
		},
	}, nil
}

// GenerateStatement генерирует банковскую выписку
// ПОЛНЫЙ WORKFLOW С KAFKA:
//  1. Валидация входных данных
//  2. Создание Statement ID
//  3. Симуляция расчетов (для примера)
//  4. Публикация в Kafka
//  5. Возврат ответа клиенту
func (s *calculationService) GenerateStatementToKafka(ctx context.Context, req *helpers.GenerateStatementRequest) (*helpers.GenerateStatementResponse, error) {
	// ШАГ 1: Генерируем уникальный ID для выписки
	statementID := "stmt_" + req.Month + "_" + req.AccountID

	log.Println("========================================")
	log.Printf("📥 ПОЛУЧЕН ЗАПРОС на генерацию выписки")
	log.Printf("   AccountID: %s", req.AccountID)
	log.Printf("   Month: %s", req.Month)
	log.Printf("   BusinessType: %s", req.BusinessType)
	log.Printf("   InitialBalance: %.2f", req.InitialBalance)
	log.Printf("   StatementID: %s", statementID)
	log.Println("========================================")

	// ШАГ 2: Симулируем расчеты (в реальности здесь будут сложные вычисления)
	log.Println("⚙️  Выполняем расчеты...")
	time.Sleep(500 * time.Millisecond) // Имитация работы

	// Создаем результаты расчетов
	calculationData := map[string]interface{}{
		"statementId":    statementID,
		"accountId":      req.AccountID,
		"month":          req.Month,
		"initialBalance": req.InitialBalance,
		"finalBalance":   req.InitialBalance + 5000.00, // Простой пример
		"totalRevenue":   10000.00,
		"totalExpenses":  -5000.00,
		"netProfit":      5000.00,
		"transactions":   []string{"transaction1", "transaction2"}, // Заглушка
	}

	log.Println("✓ Расчеты завершены")

	// ШАГ 3: Публикуем результаты в Kafka
	if s.kafkaProducer != nil {
		log.Println("📤 Отправляем результаты в Kafka...")

		// Создаем сообщение для Kafka
		kafkaMsg := &kafka.CalculationCompletedMessage{
			StatementID:   statementID,
			AccountID:     req.AccountID,
			Month:         req.Month,
			Status:        kafka.StatusCompleted,
			Data:          calculationData,
			CorrelationID: statementID, // Используем statementID как correlation ID
			Timestamp:     time.Now(),
		}

		// Публикуем в Kafka
		if err := s.kafkaProducer.PublishCalculationCompleted(ctx, kafkaMsg); err != nil {
			log.Printf("❌ Ошибка публикации в Kafka: %v", err)
			return nil, fmt.Errorf("failed to publish to Kafka: %w", err)
		}

		log.Println("✓ Сообщение успешно отправлено в Kafka!")
		log.Printf("   Topic: %s", kafka.TopicCalculationCompleted)
		log.Printf("   StatementID: %s", statementID)
	} else {
		log.Println("⚠️  Kafka producer не инициализирован (работаем без Kafka)")
	}

	// ШАГ 4: Возвращаем ответ клиенту
	return &helpers.GenerateStatementResponse{
		StatementID: statementID,
		Status:      helpers.StatusProcessing.String(),
		Message:     "Statement generation started and sent to Kafka",
	}, nil
}

// StartConsumer запускает Kafka consumer для чтения сообщений
// ДЕМО МЕТОД: Показывает как читать сообщения из Kafka
func (s *calculationService) StartConsumer(ctx context.Context) error {
	log.Println("========================================")
	log.Println("🎧 ЗАПУСК KAFKA CONSUMER")
	log.Println("========================================")

	// Получаем брокеры из переменных окружения (те же что и для Producer)
	kafkaBrokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	if len(kafkaBrokers) == 0 || kafkaBrokers[0] == "" {
		kafkaBrokers = []string{"kafka1:9092", "kafka2:9093"} // Fallback на кластер
	}

	log.Printf("📡 Connecting to Kafka brokers: %v", kafkaBrokers)

	// Создаем конфигурацию consumer
	consumerConfig := kafka.DefaultConsumerConfig(
		kafkaBrokers,                              // Kafka брокеры из env (кластер)
		kafka.ConsumerGroupMatematikaService,      // Consumer group ID
		[]string{kafka.TopicCalculationCompleted}, // Топики для подписки
	)

	// Создаем consumer (возвращает *KafkaConsumer, а не интерфейс)
	kafkaConsumer, err := kafka.NewConsumer(consumerConfig, log.Default())
	if err != nil {
		return fmt.Errorf("failed to create consumer: %w", err)
	}

	// Приводим к конкретному типу для доступа к RegisterHandler
	concreteConsumer, ok := kafkaConsumer.(*kafka.KafkaConsumer)
	if !ok {
		return fmt.Errorf("unexpected consumer type")
	}

	// Регистрируем handler для топика
	concreteConsumer.RegisterHandler(kafka.TopicCalculationCompleted, func(ctx context.Context, message *sarama.ConsumerMessage) error {
		log.Println("========================================")
		log.Println("📨 ПОЛУЧЕНО СООБЩЕНИЕ ИЗ KAFKA")
		log.Printf("   Topic: %s", message.Topic)
		log.Printf("   Partition: %d", message.Partition)
		log.Printf("   Offset: %d", message.Offset)
		log.Printf("   Key: %s", string(message.Key))
		log.Println("----------------------------------------")
		log.Printf("   Message: %s", string(message.Value))
		log.Println("========================================")

		// Десериализуем сообщение
		var msg kafka.CalculationCompletedMessage
		if err := kafka.UnmarshalMessage(message, &msg); err != nil {
			log.Printf("❌ Ошибка десериализации: %v", err)
			return err
		}

		log.Println("📊 РАСПАРСЕННЫЕ ДАННЫЕ:")
		log.Printf("   StatementID: %s", msg.StatementID)
		log.Printf("   AccountID: %s", msg.AccountID)
		log.Printf("   Month: %s", msg.Month)
		log.Printf("   Status: %s", msg.Status)
		log.Printf("   CorrelationID: %s", msg.CorrelationID)
		log.Println("========================================")

		return nil // Успешная обработка
	})

	// Запускаем consumer
	if err := concreteConsumer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start consumer: %w", err)
	}

	log.Println("✓ Kafka consumer запущен и слушает сообщения...")
	return nil
}

// GetStatementStatusByID получает статус выписки
// Параметры:
//   - ctx: Контекст запроса
//   - id: UUID выписки
//
// Возвращает error если выписка не найдена
func (s *calculationService) GetStatementStatusByID(ctx context.Context, id string) (interface{}, error) {
	// TODO: Получить статус из БД
	return map[string]string{
		"statementId": id,
		"status":      helpers.StatusCompleted.String(),
	}, nil
}

// GetStatementResultByID получает результаты расчетов
// Параметры:
//   - ctx: Контекст запроса
//   - id: UUID выписки
//
// Возвращает error если выписка не найдена или еще не готова
func (s *calculationService) GetStatementResultByID(ctx context.Context, id string) (interface{}, error) {
	// TODO: Получить результаты из БД
	return map[string]string{
		"statementId": id,
		"result":      "calculation data here",
	}, nil
}

// ================================================
// GET TRANSACTIONS !!!! ТЕСТОВЫЕ МЕТОДЫ !!!!
// ================================================

func (s *calculationService) GetTransactions(ctx context.Context) ([]Transaction, error) {
	return s.calcRepo.GetTransactions(ctx, "2025-01-01", "2025-01-03")
}

// ================================================
// GET DAILY BALANCES !!!! ТЕСТОВЫЕ МЕТОДЫ !!!!
// ================================================

func (s *calculationService) GetDailyBalances(ctx context.Context) ([]DailyBalance, error) {
	return s.calcRepo.GetDailyBalances(ctx, "2025-01-01", "2025-01-03")
}

// ================================================
// GET STATEMENTS !!!! ТЕСТОВЫЕ МЕТОДЫ !!!!
// ================================================

func (s *calculationService) GetStatements(ctx context.Context) ([]Statement, error) {
	return s.calcRepo.GetStatements(ctx)
}

// ================================================
// GENERATE STATEMENT
// ================================================

type GenerateStatementResponse struct {
	Transactions         []Transaction
	DailyClosingBalances []DailyBalance
}

// func (s *calculationService) GenerateStatement(ctx context.Context, req *helpers.StatementStateRequest) (*helpers.GenerateStatementResponse, error) {
func (s *calculationService) GenerateStatement(ctx context.Context, req *helpers.StatementStateRequest) (*GenerateStatementResponse, error) {
	// получить транзакции из БД в пределах указанного месяца
	manualIncomes := req.CustomData.ManualIncomes
	manualExpenses := req.CustomData.ManualExpenses
	transactions, err := s.calcRepo.GetTransactions(ctx, manualIncomes[0].Date, manualExpenses[0].Date)
	if err != nil {
		return nil, err
	}

	fmt.Println("TRANSACTIONS: ", transactions)

	// получить forwarding_info

	// получить daily_closing_balances
	dailyClosingBalances, err := s.calcRepo.GetDailyBalances(ctx, manualIncomes[0].Date, manualExpenses[0].Date)
	if err != nil {
		return nil, err
	}

	fmt.Println("DAILY CLOSING BALANCES: ", dailyClosingBalances)

	// получить financial_summary

	return &GenerateStatementResponse{
		Transactions:         transactions,
		DailyClosingBalances: dailyClosingBalances,
	}, nil
}

func (s *calculationService) GetBusinessRules(ctx context.Context) ([]BusinessRule, error) {
	return s.calcRepo.GetBusinessRules(ctx)
}

func (s *calculationService) GetAdminConfig(ctx context.Context) (*helpers.AdminConfigResponse, error) {
	return &helpers.AdminConfigResponse{
		ExpenseCategories: []helpers.ExpenseCategoryResponse{
			{
				ID:                uuid.New(),
				Name:              "Marketing",
				DefaultPercentMin: 0.01,
				DefaultPercentMax: 0.05,
			},
			{
				ID:                uuid.New(),
				Name:              "Cleaning",
				DefaultPercentMin: 0.01,
				DefaultPercentMax: 0.05,
			},
		},
		Schedules: []helpers.SchedulesResponse{
			{
				ID:           uuid.New(),
				CategoryID:   uuid.New(),
				Frequency:    "monthly",
				PreferredDay: "1",
				WeekOfMonth:  1,
				NTimes:       1,
				TimeWindow:   "10:00-12:00",
			},
		},
		IncomeTemplates: []helpers.IncomeTemplateResponse{
			{
				ID:                       uuid.New(),
				BusinessModel:            "B2C",
				Category:                 "Marketing",
				CountMin:                 1,
				CountMax:                 10,
				PercentPerTransactionMin: 0.01,
				PercentPerTransactionMax: 0.05,
				DefaultMethods:           []string{"ACH", "WIRE", "ZELLE", "GATEWAY", "OTHER"},
			},
		},
	}, nil
}
