package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"
)

type Consumer interface {
	Start(ctx context.Context) error
	Close() error
}

type MessageHandler func(ctx context.Context, message *sarama.ConsumerMessage) error

type KafkaConsumer struct {
	consumerGroup sarama.ConsumerGroup
	topics        []string
	handlers      map[string]MessageHandler
	config        *ConsumerConfig
	logger        *logrus.Logger
	wg            sync.WaitGroup
}

type ConsumerConfig struct {
	Brokers        []string
	GroupID        string
	Topics         []string
	StartOffset    int64
	MaxRetry       int
	RetryBackoff   time.Duration
	SessionTimeout time.Duration
}

type ConsumerGroupHandler struct {
	consumer *KafkaConsumer
}

func (h *ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	h.consumer.logger.Info("Consumer group session started")
	return nil
}

func (h *ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	h.consumer.logger.Info("Consumer group session ended")
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
				h.consumer.logger.Errorf("Error processing message: %v", err)
			}

			session.MarkMessage(message, "")
		case <-session.Context().Done():
			return nil
		}
	}
}

func (h *ConsumerGroupHandler) processMessage(ctx context.Context, message *sarama.ConsumerMessage) error {
	startTime := time.Now()

	h.consumer.logger.Infof("Received message from topic=%s partition=%d offset=%d key=%s",
		message.Topic, message.Partition, message.Offset, string(message.Key))

	correlationID := extractHeaderValue(message.Headers, "correlation-id")
	if correlationID != "" {
		h.consumer.logger.Infof("Processing message with correlation-id=%s", correlationID)
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
			h.consumer.logger.Infof("Retry attempt %d/%d for message from topic %s",
				attempt, h.consumer.config.MaxRetry, message.Topic)
			// Exponential backoff: каждый повтор ждет дольше
			time.Sleep(h.consumer.config.RetryBackoff * time.Duration(attempt))
		}

		// Вызываем handler для обработки сообщения
		if err := handler(ctx, message); err == nil {
			// Успешная обработка!
			duration := time.Since(startTime)
			// Логируем для метрик производительности
			h.consumer.logger.Infof("Message processed successfully in %v (topic=%s, offset=%d)",
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

func NewConsumer(config *ConsumerConfig, logger *logrus.Logger) (Consumer, error) {
	// Если логгер не передан, используем стандартный
	if logger == nil {
		logger = logrus.New()
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
	logrus.Infof("Kafka consumer group %s created successfully", config.GroupID)

	// Возвращаем инициализированный consumer
	return &KafkaConsumer{
		consumerGroup: consumerGroup,
		topics:        config.Topics,
		handlers:      make(map[string]MessageHandler), // Инициализируем пустую map для handlers
		config:        config,
		logger:        logrus.New(),
	}, nil
}

func (c *KafkaConsumer) RegisterHandler(topic string, handler MessageHandler) {
	c.handlers[topic] = handler
	c.logger.Infof("Handler registered for topic: %s", topic)
}

func (c *KafkaConsumer) Start(ctx context.Context) error {
	c.logger.Infof("Starting Kafka consumer for topics: %v", c.topics)

	handler := &ConsumerGroupHandler{consumer: c}
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for err := range c.consumerGroup.Errors() {
			c.logger.Errorf("ERROR: Consumer group error: %v", err)
		}
	}()

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			if err := c.consumerGroup.Consume(ctx, c.topics, handler); err != nil {
				c.logger.Errorf("ERROR: Consumer group consume error: %v", err)
			}

			if ctx.Err() != nil {
				c.logger.Info("Context cancelled, stopping consumer")
				return
			}
		}
	}()

	c.logger.Info("Kafka consumer started successfully")
	return nil
}

func (c *KafkaConsumer) Close() error {
	c.logger.Info("Closing Kafka consumer...")

	if err := c.consumerGroup.Close(); err != nil {
		c.logger.Errorf("ERROR: Failed to close consumer group: %v", err)
		return err
	}

	c.wg.Wait()

	c.logger.Info("Kafka consumer closed successfully")
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
		Brokers:        brokers,
		GroupID:        groupID,                // Consumer group ID для распределенной обработки
		Topics:         topics,                 // Топики для чтения
		StartOffset:    sarama.OffsetNewest,    // Начинать с НОВЫХ сообщений (не читаем историю)
		MaxRetry:       3,                      // 3 повтора обработки при ошибке
		RetryBackoff:   100 * time.Millisecond, // 100ms базовая задержка между повторами
		SessionTimeout: 10 * time.Second,       // 10s timeout - если consumer не отвечает, его исключат из группы
	}
}
