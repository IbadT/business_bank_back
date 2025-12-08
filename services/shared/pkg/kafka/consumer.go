package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/IBM/sarama"
)

type Consumer interface {
	Start(ctx context.Context) error
	Close() error
}

type MessageHandler func(ctx context.Context, message *sarama.ConsumerMessage) error

type KafkaConsumer struct {
	consumerGroup sarama.ConsumerGroup
	topics []string
	handlers map[string]MessageHandler
	config *ConsumerConfig
	logger *log.Logger
	wg sync.WaitGroup
}

type ConsumerConfig struct {
	Brokers []string
	GroupID string
	Topics []string
	StartOffset int64
	MaxRetry int
	RetryBackoff time.Duration
	SessionTimeout time.Duration
}

type ConsumerGroupHandler struct {
	consumer *KafkaConsumer
}

func (h *ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	h.consumer.logger.Println("Consumer group session started")
	return nil
}

func (h *ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	h.consumer.logger.Println("Consumer group session ended")
	return nil
}

func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			if err := h.processMessage(session.Context(), message); err != nil {
				h.consumer.logger.Printf("Error processing message: %v", err)
			}

			session.MarkMessage(message, "")
		case <-session.Context().Done():
			return nil
		}
	}
}

func (h *ConsumerGroupHandler) processMessage(ctx context.Context, message *sarama.ConsumerMessage) error {
	startTime := time.Now()

	h.consumer.logger.Printf("Received message from topic=%s partition=%d offset=%d key=%s",
		message.Topic, message.Partition, message.Offset, string(message.Key))

	correlationID := extractHeaderValue(message.Headers, "correlation-id")
	if correlationID != "" {
		h.consumer.logger.Printf("Processing message with correlation-id=%s", correlationID)
	}

	handler, exists := h.consumer.handlers[message.Topic]
	if !exists {
		return fmt.Errorf("no handler found for topic: %s", message.Topic)
	}

		// Retry логика с exponential backoff для обработки сообщения
	// Пытаемся обработать MaxRetry+1 раз (первая попытка + повторы)
	var lastErr error
	for attempt := 0; attempt <= h.consumer.config.MaxRetry; attempt++ {
		// Если это повтор (не первая попытка), делаем задержку
		if attempt > 0 {
			h.consumer.logger.Printf("Retry attempt %d/%d for message from topic %s",
				attempt, h.consumer.config.MaxRetry, message.Topic)
			// Exponential backoff: каждый повтор ждет дольше
			time.Sleep(h.consumer.config.RetryBackoff * time.Duration(attempt))
		}

		// Вызываем handler для обработки сообщения
		if err := handler(ctx, message); err == nil {
			// Успешная обработка!
			duration := time.Since(startTime)
			// Логируем для метрик производительности
			h.consumer.logger.Printf("Message processed successfully in %v (topic=%s, offset=%d)",
				duration, message.Topic, message.Offset)
			return nil // Выходим с успехом
		} else {
			// Сохраняем ошибку для потенциального возврата
			lastErr = err
		}
	}

	// Все попытки исчерпаны, возвращаем ошибку
	// В production это сообщение должно пойти в Dead Letter Queue
	return fmt.Errorf("failed to process message after %d attempts: %w",
		h.consumer.config.MaxRetry+1, lastErr)
}

func NewConsumer(config *ConsumerConfig, logger *log.Logger) (Consumer, error) {
	// Если логгер не передан, используем стандартный
	if logger == nil {
		logger = log.Default()
	}

	// Создаем базовую Kafka конфигурацию
	saramaConfig := NewKafkaConfig()

	// Настраиваем стратегию rebalance:
	// RoundRobin - равномерно распределяет партиции между consumers в группе
	// Альтернативы: Range, Sticky (сохраняет назначение партиций при rebalance)
	saramaConfig.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()

	// Начальный offset для новых consumer groups:
	// OffsetNewest - начинать с новых сообщений (пропустить старые)
	// OffsetOldest - читать все сообщения с начала
	saramaConfig.Consumer.Offsets.Initial = config.StartOffset

	// Таймаут сессии - если consumer не отправляет heartbeat, его исключат из группы
	// Это защищает от "зависших" consumers
	saramaConfig.Consumer.Group.Session.Timeout = config.SessionTimeout

	// Включаем возврат ошибок для обработки
	saramaConfig.Consumer.Return.Errors = true

	// Создаем consumer group с указанным GroupID
	// Все consumers с одинаковым GroupID формируют группу и делят партиции
	consumerGroup, err := sarama.NewConsumerGroup(config.Brokers, config.GroupID, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	// Логируем успешное создание для observability
	logger.Printf("Kafka consumer group %s created successfully", config.GroupID)

	// Возвращаем инициализированный consumer
	return &KafkaConsumer{
		consumerGroup: consumerGroup,
		topics:        config.Topics,
		handlers:      make(map[string]MessageHandler), // Инициализируем пустую map для handlers
		config:        config,
		logger:        logger,
	}, nil
}

func (c *KafkaConsumer) RegisterHandler(topic string, handler MessageHandler) {
	c.handlers[topic] = handler
	c.logger.Printf("Handler registered for topic: %s", topic)
}

func (c *KafkaConsumer) Start(ctx context.Context) error {
	c.logger.Println("Starting Kafka consumer for topics: %v", c.topics)

	handler := &ConsumerGroupHandler{consumer: c}
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for err := range c.consumerGroup.Errors() {
			c.logger.Printf("ERROR: Consumer group error: %v", err)
		}
	}()

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			if err := c.consumerGroup.Consume(ctx, c.topics, handler); err != nil {
				c.logger.Printf("ERROR: Consumer group consume error: %v", err)
			}

			if ctx.Err() != nil {
				c.logger.Println("Context cancelled, stopping consumer")
				return
			}
		}
	}()

	c.logger.Println("Kafka consumer started successfully")
	return nil
}

func (c *KafkaConsumer) Close() error {
	c.logger.Println("Closing Kafka consumer...")

	if err := c.consumerGroup.Close(); err != nil {
		c.logger.Printf("ERROR: Failed to close consumer group: %v", err)
		return err
	}

	c.wg.Wait()

	c.logger.Println("Kafka consumer closed successfully")
	return nil
}

func extractHeaderValue(headers []*sarama.RecordHeader, key string) string {
	// Перебираем все заголовки
	for _, header := range headers {
		// Сравниваем ключ заголовка с искомым
		if string(header.Key) == key {
			// Нашли! Возвращаем значение
			return string(header.Value)
		}
	}
	// Заголовок не найден
	return ""
}

func UnmarshalMessage(message *sarama.ConsumerMessage, target interface{}) error {
	if err := json.Unmarshal(message.Value, target); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}
	return nil
}

func DefaultConsumerConfig(brokers []string, groupID string, topics []string) *ConsumerConfig {
	return &ConsumerConfig{
		Brokers: brokers,
		GroupID:        groupID,                // Consumer group ID для распределенной обработки
		Topics:         topics,                 // Топики для чтения
		StartOffset:    sarama.OffsetNewest,    // Начинать с НОВЫХ сообщений (не читаем историю)
		MaxRetry:       3,                      // 3 повтора обработки при ошибке
		RetryBackoff:   100 * time.Millisecond, // 100ms базовая задержка между повторами
		SessionTimeout: 10 * time.Second,       // 10s timeout - если consumer не отвечает, его исключат из группы
	}
}