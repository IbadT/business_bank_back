package calculation

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/IbadT/business_bank_back/services/matematika/internal/kafka"
)

// ============================================================================
// ИНТЕРФЕЙС СЕРВИСА
// ============================================================================

// CalculationService - интерфейс бизнес-логики расчетов
// Определяет контракт для работы с выписками
type CalculationService interface {
	// GenerateStatement генерирует новую выписку и публикует результат в Kafka
	GenerateStatement(ctx context.Context, req *GenerateStatementRequest) (*GenerateStatementResponse, error)

	// GetStatementStatusByID получает статус выписки по ID
	GetStatementStatusByID(ctx context.Context, id string) (interface{}, error)

	// GetStatementResultByID получает результаты расчетов по ID
	GetStatementResultByID(ctx context.Context, id string) (interface{}, error)

	// StartConsumer запускает Kafka consumer для чтения сообщений
	StartConsumer(ctx context.Context) error
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
}

// NewCalculationService создает новый сервис БЕЗ Kafka (для обратной совместимости)
// DEPRECATED: Используйте NewCalculationServiceWithKafka для production
func NewCalculationService(calcRepo CalculationRepository, kafkaProducer kafka.Producer) CalculationService {
	return &calculationService{
		calcRepo:      calcRepo,
		kafkaProducer: kafkaProducer,
	}
}

// NewCalculationServiceWithKafka создает новый сервис С Kafka producer
// Рекомендуемый способ для production
// Параметры:
//   - calcRepo: Repository для работы с БД
//   - kafkaProducer: Producer для публикации событий в Kafka
//
// Возвращает готовый к использованию сервис
func NewCalculationServiceWithKafka(calcRepo CalculationRepository, kafkaProducer kafka.Producer) CalculationService {
	return &calculationService{
		calcRepo:      calcRepo,
		kafkaProducer: kafkaProducer, // Внедряем Kafka через Dependency Injection
	}
}

// ============================================================================
// МЕТОДЫ СЕРВИСА
// ============================================================================

// GenerateStatement генерирует банковскую выписку
// ПОЛНЫЙ WORKFLOW С KAFKA:
//  1. Валидация входных данных
//  2. Создание Statement ID
//  3. Симуляция расчетов (для примера)
//  4. Публикация в Kafka
//  5. Возврат ответа клиенту
func (s *calculationService) GenerateStatement(ctx context.Context, req *GenerateStatementRequest) (*GenerateStatementResponse, error) {
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
	return &GenerateStatementResponse{
		StatementID: statementID,
		Status:      "processing",
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
		"status":      "completed",
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
