package kafka

import (
	"context"       // Для управления жизненным циклом операций и отмены
	"encoding/json" // Для сериализации сообщений в JSON
	"fmt"           // Для форматирования строк и ошибок
	"log"           // Для логирования событий
	"time"          // Для работы с временными метками

	"github.com/IBM/sarama"  // Kafka клиент для Go
	"github.com/google/uuid" // Для генерации уникальных идентификаторов
)

// Producer - интерфейс для Kafka producer
// Использование интерфейса позволяет:
// 1. Легко тестировать код с mock реализациями
// 2. Заменять реализацию без изменения кода
// 3. Следовать принципу Dependency Inversion (SOLID)
type Producer interface {
	// PublishStatement публикует запрос на генерацию выписки
	PublishStatement(ctx context.Context, msg *StatementMessage) error

	// PublishCalculationCompleted публикует результат расчетов
	PublishCalculationCompleted(ctx context.Context, msg *CalculationCompletedMessage) error

	// Close корректно закрывает соединение с Kafka
	Close() error
}

// KafkaProducer - конкретная реализация интерфейса Producer
// Хранит:
// - producer: Sarama SyncProducer для синхронной отправки (гарантия доставки)
// - config: Настройки producer (brokers, retry, compression)
// - logger: Логгер для отслеживания операций
type KafkaProducer struct {
	producer sarama.SyncProducer // Синхронный producer - ждет подтверждения от Kafka
	config   *ProducerConfig     // Конфигурация подключения и поведения
	logger   *log.Logger         // Логгер для observability
}

// ProducerConfig - конфигурация для Kafka producer
// Определяет как producer будет работать с Kafka
type ProducerConfig struct {
	Brokers          []string                // Список Kafka брокеров (например: ["localhost:9092"])
	RequiredAcks     sarama.RequiredAcks     // Уровень подтверждения (WaitForAll = максимальная надежность)
	Compression      sarama.CompressionCodec // Сжатие сообщений (Snappy - баланс скорость/размер)
	MaxRetry         int                     // Количество повторов при ошибке отправки
	RetryBackoff     time.Duration           // Задержка между повторами (exponential backoff)
	IdempotentWrites bool                    // Гарантия, что дублирующие отправки не создадут дубли
}

// StatementMessage - сообщение для запроса генерации выписки
// Это сообщение отправляется когда клиент запрашивает создание выписки
// Содержит все необходимые параметры для генерации
type StatementMessage struct {
	StatementID    string  `json:"statementId"`    // Уникальный ID выписки (UUID)
	AccountID      string  `json:"accountId"`      // ID аккаунта клиента
	Month          string  `json:"month"`          // Месяц для выписки (формат: YYYY-MM)
	BusinessType   string  `json:"businessType"`   // Тип бизнеса (B2B или B2C)
	InitialBalance float64 `json:"initialBalance"` // Начальный баланс на начало периода
}

// CalculationCompletedMessage - сообщение о завершении расчетов
// Отправляется Matematika Service после успешного завершения расчетов
// Содержит все рассчитанные данные для дальнейшего форматирования в Maska Service
type CalculationCompletedMessage struct {
	StatementID   string                 `json:"statementId"`   // ID выписки для связи
	AccountID     string                 `json:"accountId"`     // ID аккаунта
	Month         string                 `json:"month"`         // Месяц выписки
	Status        string                 `json:"status"`        // Статус: completed/failed
	Data          map[string]interface{} `json:"data"`          // Рассчитанные данные (транзакции, балансы)
	CorrelationID string                 `json:"correlationId"` // ID для трассировки через сервисы
	Timestamp     time.Time              `json:"timestamp"`     // Время завершения расчетов
}

// NewProducer создает новый Kafka producer с заданной конфигурацией
// Параметры:
//   - config: Настройки подключения и поведения producer
//   - logger: Логгер для вывода событий (если nil, использует стандартный)
//
// Возвращает:
//   - Producer: Готовый к использованию producer
//   - error: Ошибка если не удалось подключиться к Kafka
func NewProducer(config *ProducerConfig, logger *log.Logger) (Producer, error) {
	// Если логгер не передан, используем стандартный
	if logger == nil {
		logger = log.Default()
	}

	// Создаем базовую Kafka конфигурацию
	saramaConfig := NewKafkaConfig()

	// Настраиваем параметры producer для надежности и производительности:
	saramaConfig.Producer.RequiredAcks = config.RequiredAcks   // WaitForAll = ждем подтверждения от всех реплик
	saramaConfig.Producer.Retry.Max = config.MaxRetry          // Количество повторов при ошибке
	saramaConfig.Producer.Retry.Backoff = config.RetryBackoff  // Задержка между повторами
	saramaConfig.Producer.Compression = config.Compression     // Сжатие для уменьшения трафика
	saramaConfig.Producer.Idempotent = config.IdempotentWrites // Идемпотентность - защита от дублей
	saramaConfig.Producer.Return.Successes = true              // Возвращать успешные отправки
	saramaConfig.Producer.Return.Errors = true                 // Возвращать ошибки для обработки

	// ВАЖНО: Для идемпотентного producer требуется MaxOpenRequests = 1
	// Это гарантирует порядок сообщений и предотвращает дубликаты
	if config.IdempotentWrites {
		saramaConfig.Net.MaxOpenRequests = 1
	}

	// Создаем синхронный producer (блокирует до получения подтверждения)
	// Синхронный выбран для гарантии доставки критичных сообщений
	producer, err := sarama.NewSyncProducer(config.Brokers, saramaConfig)
	if err != nil {
		// Используем %w для wrapping ошибки - позволяет использовать errors.Is/As
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}

	// Логируем успешное подключение для observability
	logger.Printf("Kafka producer connected to brokers: %v", config.Brokers)

	// Возвращаем инициализированный producer
	return &KafkaProducer{
		producer: producer,
		config:   config,
		logger:   logger,
	}, nil
}

// PublishStatement публикует сообщение о запросе генерации выписки
// Используется когда нужно инициировать генерацию новой выписки
// Параметры:
//   - ctx: Контекст для отмены операции
//   - msg: Данные запроса на генерацию
//
// Возвращает error если не удалось опубликовать
func (p *KafkaProducer) PublishStatement(ctx context.Context, msg *StatementMessage) error {
	// Делегируем публикацию общему методу publish
	// Используем StatementID как ключ для партиционирования (все сообщения одной выписки в одну партицию)
	return p.publish(ctx, TopicStatementGenerationRequest, msg.StatementID, msg)
}

// PublishCalculationCompleted публикует сообщение о завершении расчетов
// Вызывается Matematika Service после успешного расчета выписки
// Сообщение получит Maska Service для форматирования
// Параметры:
//   - ctx: Контекст для отмены операции
//   - msg: Результаты расчетов
//
// Возвращает error если не удалось опубликовать
func (p *KafkaProducer) PublishCalculationCompleted(ctx context.Context, msg *CalculationCompletedMessage) error {
	// Добавляем timestamp если не указан (важно для audit trail)
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}

	// Добавляем correlationId если не указан (для distributed tracing)
	// UUID гарантирует уникальность и позволяет отследить запрос через все сервисы
	if msg.CorrelationID == "" {
		msg.CorrelationID = uuid.New().String()
	}

	// Делегируем публикацию общему методу
	return p.publish(ctx, TopicCalculationCompleted, msg.StatementID, msg)
}

// publish - внутренний метод для публикации сообщений в Kafka
// Общий метод используется всеми публичными методами для DRY принципа
// Параметры:
//   - ctx: Контекст для возможности отмены операции
//   - topic: Имя Kafka топика
//   - key: Ключ для партиционирования (сообщения с одинаковым ключом идут в одну партицию)
//   - payload: Данные для отправки (будут сериализованы в JSON)
//
// Возвращает error если не удалось опубликовать после всех повторов
func (p *KafkaProducer) publish(ctx context.Context, topic, key string, payload interface{}) error {
	// Проверяем не отменен ли контекст перед началом работы
	// Если ctx.Done() закрыт, операция была отменена извне
	select {
	case <-ctx.Done():
		return ctx.Err() // Возвращаем ошибку отмены
	default:
		// Контекст активен, продолжаем
	}

	// Сериализуем payload в JSON байты
	// JSON выбран как универсальный формат, понятный всем сервисам
	message, err := json.Marshal(payload)
	if err != nil {
		p.logger.Printf("ERROR: Failed to marshal message: %v", err)
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Создаем Kafka сообщение с метаданными
	kafkaMsg := &sarama.ProducerMessage{
		Topic: topic,                       // Топик назначения
		Key:   sarama.StringEncoder(key),   // Ключ для партиционирования (важно для порядка!)
		Value: sarama.ByteEncoder(message), // Тело сообщения (JSON байты)

		// Headers - метаданные для трассировки и отладки
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("correlation-id"),              // Correlation ID для distributed tracing
				Value: []byte(extractCorrelationID(payload)), // Извлекаем из payload
			},
			{
				Key:   []byte("timestamp"),                     // Временная метка создания сообщения
				Value: []byte(time.Now().Format(time.RFC3339)), // RFC3339 - стандартный формат
			},
		},
		Timestamp: time.Now(), // Kafka timestamp для упорядочивания
	}

	// Retry логика с exponential backoff
	// Пытаемся отправить MaxRetry+1 раз (первая попытка + повторы)
	var lastErr error
	for attempt := 0; attempt <= p.config.MaxRetry; attempt++ {
		// Если это повтор (не первая попытка), делаем задержку
		if attempt > 0 {
			p.logger.Printf("Retry attempt %d/%d for topic %s", attempt, p.config.MaxRetry, topic)
			// Exponential backoff: каждый следующий повтор ждет дольше
			// attempt=1: RetryBackoff*1, attempt=2: RetryBackoff*2, и т.д.
			time.Sleep(p.config.RetryBackoff * time.Duration(attempt))
		}

		// Отправляем сообщение синхронно
		// partition - номер партиции куда попало сообщение
		// offset - позиция сообщения в партиции
		partition, offset, err := p.producer.SendMessage(kafkaMsg)
		if err == nil {
			// Успешная отправка! Логируем для мониторинга
			p.logger.Printf("Message published successfully to topic=%s partition=%d offset=%d key=%s",
				topic, partition, offset, key)
			return nil // Выходим из функции с успехом
		}

		// Сохраняем ошибку для потенциального возврата
		lastErr = err
		p.logger.Printf("ERROR: Failed to publish message (attempt %d/%d): %v", attempt+1, p.config.MaxRetry+1, err)
	}

	// Все попытки исчерпаны, возвращаем последнюю ошибку
	return fmt.Errorf("failed to publish message after %d attempts: %w", p.config.MaxRetry+1, lastErr)
}

// Close корректно закрывает producer и освобождает ресурсы
// Должен вызываться при завершении работы приложения (defer producer.Close())
// Возвращает error если не удалось корректно закрыть соединение
func (p *KafkaProducer) Close() error {
	// Пытаемся закрыть producer
	if err := p.producer.Close(); err != nil {
		// Логируем ошибку для диагностики
		p.logger.Printf("ERROR: Failed to close kafka producer: %v", err)
		return err
	}
	// Успешное закрытие
	p.logger.Println("Kafka producer closed successfully")
	return nil
}

// extractCorrelationID извлекает correlation ID из разных типов сообщений
// Correlation ID используется для трассировки запроса через все микросервисы
// Позволяет в логах и метриках связать все события одного запроса
// Параметры:
//   - payload: Сообщение любого типа
//
// Возвращает correlation ID или генерирует новый UUID если не найден
func extractCorrelationID(payload interface{}) string {
	// Type switch - проверяем конкретный тип payload
	switch v := payload.(type) {
	case *CalculationCompletedMessage:
		// Для сообщения о завершении используем CorrelationID
		return v.CorrelationID
	case *StatementMessage:
		// Для запроса на генерацию используем StatementID как correlation ID
		return v.StatementID
	default:
		// Для неизвестных типов генерируем новый UUID
		// Это fallback для расширяемости
		return uuid.New().String()
	}
}

// DefaultProducerConfig возвращает рекомендуемую конфигурацию для production
// Эти настройки оптимизированы для:
// - Максимальной надежности (WaitForAll)
// - Производительности (Snappy compression)
// - Отказоустойчивости (3 retry с backoff)
// - Идемпотентности (защита от дублей)
// Параметры:
//   - brokers: Список Kafka брокеров
//
// Возвращает готовую конфигурацию для продакшна
func DefaultProducerConfig(brokers []string) *ProducerConfig {
	return &ProducerConfig{
		Brokers:          brokers,                  // Kafka брокеры
		RequiredAcks:     sarama.WaitForAll,        // Ждем подтверждения от ВСЕХ реплик (самая надежная настройка)
		Compression:      sarama.CompressionSnappy, // Snappy - быстрое сжатие, уменьшает сетевой трафик
		MaxRetry:         3,                        // 3 повтора при ошибке (баланс надежность/время)
		RetryBackoff:     100 * time.Millisecond,   // 100ms базовая задержка (exp backoff: 100ms, 200ms, 300ms)
		IdempotentWrites: true,                     // Идемпотентность - Kafka не создаст дубли при повторной отправке
	}
}
