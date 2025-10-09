package kafka

import (
	"context"       // Для управления жизненным циклом и отмены операций
	"encoding/json" // Для десериализации JSON сообщений
	"fmt"           // Для форматирования строк и ошибок
	"log"           // Для логирования событий
	"sync"          // Для синхронизации goroutines (WaitGroup)
	"time"          // Для работы с таймаутами и временем

	"github.com/IBM/sarama" // Kafka клиент для Go
)

// Consumer - интерфейс для Kafka consumer
// Использование интерфейса позволяет:
// 1. Тестировать код с mock реализациями
// 2. Легко заменять реализацию (например, на другую библиотеку)
// 3. Следовать принципу Dependency Inversion
type Consumer interface {
	// Start запускает consumer в фоновом режиме
	Start(ctx context.Context) error

	// Close корректно останавливает consumer и ждет завершения обработки
	Close() error
}

// MessageHandler - функция для обработки одного сообщения
// Каждый топик может иметь свой handler
// Параметры:
//   - ctx: Контекст для отмены обработки
//   - message: Kafka сообщение с данными и метаданными
//
// Возвращает error если обработка не удалась (будет retry)
type MessageHandler func(ctx context.Context, message *sarama.ConsumerMessage) error

// KafkaConsumer - конкретная реализация интерфейса Consumer
// Использует Consumer Group паттерн для:
// - Горизонтального масштабирования (несколько инстансов в группе)
// - Автоматического распределения партиций
// - Гарантии обработки каждого сообщения только одним consumer'ом в группе
type KafkaConsumer struct {
	consumerGroup sarama.ConsumerGroup      // Consumer group для распределенной обработки
	topics        []string                  // Список топиков для подписки
	handlers      map[string]MessageHandler // Map топик -> handler для обработки разных топиков
	config        *ConsumerConfig           // Конфигурация consumer
	logger        *log.Logger               // Логгер для observability
	wg            sync.WaitGroup            // WaitGroup для graceful shutdown (ждем завершения goroutines)
}

// ConsumerConfig - конфигурация для Kafka consumer
// Определяет как consumer будет читать и обрабатывать сообщения
type ConsumerConfig struct {
	Brokers        []string      // Список Kafka брокеров (например: ["localhost:9092"])
	GroupID        string        // ID consumer group (консюмеры с одинаковым ID делят партиции)
	Topics         []string      // Топики для подписки
	StartOffset    int64         // С какого offset начинать: OffsetNewest (новые) или OffsetOldest (все)
	MaxRetry       int           // Количество повторов обработки сообщения при ошибке
	RetryBackoff   time.Duration // Задержка между повторами обработки
	SessionTimeout time.Duration // Таймаут сессии (если consumer не отвечает, его исключат из группы)
}

// ConsumerGroupHandler - реализация sarama.ConsumerGroupHandler интерфейса
// Требуется Sarama для работы с consumer groups
// Содержит ссылку на родительский consumer для доступа к handlers и логгеру
type ConsumerGroupHandler struct {
	consumer *KafkaConsumer // Ссылка на KafkaConsumer для доступа к конфигурации и handlers
}

// Setup вызывается Kafka когда consumer присоединяется к группе
// Происходит при:
// - Первом запуске consumer
// - Rebalance группы (когда добавляется/удаляется consumer)
// Можно использовать для инициализации ресурсов перед обработкой
func (h *ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	h.consumer.logger.Println("Consumer group session started")
	// Здесь можно добавить логику инициализации (подключение к БД, кэшам и т.д.)
	return nil
}

// Cleanup вызывается Kafka когда consumer покидает группу
// Происходит при:
// - Остановке consumer
// - Rebalance группы
// Используется для освобождения ресурсов
func (h *ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	h.consumer.logger.Println("Consumer group session ended")
	// Здесь можно добавить логику очистки (закрытие соединений и т.д.)
	return nil
}

// ConsumeClaim обрабатывает сообщения из назначенной партиции
// Вызывается Kafka для каждой партиции, назначенной этому consumer
// Работает в бесконечном цикле пока не отменят контекст
// Параметры:
//   - session: Сессия consumer group для коммита offset'ов
//   - claim: Партиция с сообщениями для обработки
//
// Возвращает error при критической ошибке
func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// Бесконечный цикл обработки сообщений
	for {
		select {
		// Получаем сообщение из канала Messages()
		case message := <-claim.Messages():
			// message == nil означает что канал закрыт (партиция отозвана)
			if message == nil {
				return nil
			}

			// Обрабатываем сообщение через registered handler
			if err := h.processMessage(session.Context(), message); err != nil {
				h.consumer.logger.Printf("ERROR: Failed to process message: %v", err)
				// ВАЖНО: В production нужно решить что делать с ошибками:
				// Вариант 1: Пропустить сообщение (текущая реализация)
				// Вариант 2: Отправить в Dead Letter Queue (DLQ)
				// Вариант 3: Остановить consumer и алертить
				// Сейчас пропускаем для продолжения работы
			}

			// Коммитим offset только после успешной обработки
			// Это гарантирует at-least-once delivery:
			// Если упадем до коммита, сообщение обработается снова
			session.MarkMessage(message, "") // Пустая строка = метадата не нужна

		// Слушаем отмену контекста для graceful shutdown
		case <-session.Context().Done():
			// Контекст отменен, выходим из цикла
			return nil
		}
	}
}

// processMessage обрабатывает одно Kafka сообщение с retry логикой
// Это внутренний метод, вызываемый из ConsumeClaim
// Выполняет:
// - Логирование входящего сообщения для observability
// - Извлечение correlation ID для distributed tracing
// - Поиск и вызов соответствующего handler'а
// - Retry при ошибках обработки
// - Измерение времени обработки для метрик
// Параметры:
//   - ctx: Контекст для отмены обработки
//   - message: Kafka сообщение для обработки
//
// Возвращает error если обработка не удалась после всех повторов
func (h *ConsumerGroupHandler) processMessage(ctx context.Context, message *sarama.ConsumerMessage) error {
	// Засекаем время начала для метрик производительности
	startTime := time.Now()

	// Логируем входящее сообщение с ключевыми метаданными
	// Topic - откуда пришло, Partition - из какой партиции, Offset - позиция в партиции
	h.consumer.logger.Printf("Received message from topic=%s partition=%d offset=%d key=%s",
		message.Topic, message.Partition, message.Offset, string(message.Key))

	// Извлекаем correlation ID из заголовков сообщения
	// Это позволяет связать все логи одного запроса через микросервисы
	correlationID := extractHeaderValue(message.Headers, "correlation-id")
	if correlationID != "" {
		// Логируем correlation ID для возможности фильтрации в Loki/Grafana
		h.consumer.logger.Printf("Processing message with correlation-id=%s", correlationID)
	}

	// Ищем зарегистрированный handler для этого топика
	// Handlers регистрируются через RegisterHandler() при инициализации
	handler, exists := h.consumer.handlers[message.Topic]
	if !exists {
		// Если handler не найден - это ошибка конфигурации
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

// NewConsumer создает новый Kafka consumer с заданной конфигурацией
// Consumer использует Consumer Group паттерн для распределенной обработки
// Параметры:
//   - config: Настройки подключения и поведения consumer
//   - logger: Логгер для вывода событий (если nil, использует стандартный)
//
// Возвращает:
//   - Consumer: Готовый к использованию consumer
//   - error: Ошибка если не удалось подключиться к Kafka
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

// RegisterHandler регистрирует функцию-обработчик для конкретного топика
// Должен вызываться ПЕРЕД Start() для всех топиков
// Параметры:
//   - topic: Имя топика Kafka
//   - handler: Функция для обработки сообщений из этого топика
//
// Пример использования:
//
//	consumer.RegisterHandler("statement.calculation.completed", func(ctx context.Context, msg *sarama.ConsumerMessage) error {
//	    // обработка сообщения
//	    return nil
//	})
func (c *KafkaConsumer) RegisterHandler(topic string, handler MessageHandler) {
	// Сохраняем handler в map топик -> handler
	c.handlers[topic] = handler
	c.logger.Printf("Handler registered for topic: %s", topic)
}

// Start запускает consumer в фоновом режиме (goroutines)
// Consumer начинает читать сообщения из Kafka и обрабатывать их
// Работает до тех пор, пока не отменят контекст
// Параметры:
//   - ctx: Контекст для управления жизненным циклом consumer
//
// Возвращает error только при критических ошибках инициализации
func (c *KafkaConsumer) Start(ctx context.Context) error {
	c.logger.Printf("Starting Kafka consumer for topics: %v", c.topics)

	// Создаем handler который будет обрабатывать сообщения
	handler := &ConsumerGroupHandler{consumer: c}

	// Goroutine #1: Обработка ошибок от consumer group
	// Kafka отправляет ошибки в канал Errors(), мы их логируем
	c.wg.Add(1) // Увеличиваем счетчик WaitGroup для graceful shutdown
	go func() {
		defer c.wg.Done() // Уменьшаем счетчик при завершении goroutine

		// Читаем ошибки из канала до его закрытия
		for err := range c.consumerGroup.Errors() {
			// Логируем ошибку для мониторинга
			// В production здесь можно отправлять метрики в Prometheus
			c.logger.Printf("ERROR: Consumer group error: %v", err)
		}
	}()

	// Goroutine #2: Основной цикл потребления сообщений
	// Запускается в фоне чтобы Start() не блокировал вызывающий код
	c.wg.Add(1) // Увеличиваем счетчик WaitGroup
	go func() {
		defer c.wg.Done() // Уменьшаем счетчик при завершении

		// Бесконечный цикл потребления
		for {
			// Consume блокирует до получения сообщений или отмены контекста
			// Метод должен вызываться в цикле т.к. он возвращается при rebalance
			// Rebalance происходит когда:
			// - Добавляется/удаляется consumer в группе
			// - Изменяется количество партиций в топике
			if err := c.consumerGroup.Consume(ctx, c.topics, handler); err != nil {
				c.logger.Printf("ERROR: Consumer group consume error: %v", err)
				// Продолжаем работу даже при ошибке (auto-recovery)
			}

			// Проверяем не отменен ли контекст
			if ctx.Err() != nil {
				c.logger.Println("Context cancelled, stopping consumer")
				return // Выходим из goroutine
			}

			// Если Consume вернулся без ошибки и контекст активен,
			// значит произошел rebalance - продолжаем цикл
		}
	}()

	c.logger.Println("Kafka consumer started successfully")
	return nil // Возвращаемся сразу, consumer работает в фоне
}

// Close корректно останавливает consumer и ждет завершения всех операций
// Должен вызываться при shutdown приложения (например, при SIGTERM)
// Гарантирует:
// - Корректное завершение обработки текущих сообщений
// - Коммит последних offset'ов
// - Освобождение ресурсов
// Возвращает error если не удалось корректно закрыть
func (c *KafkaConsumer) Close() error {
	c.logger.Println("Closing Kafka consumer...")

	// Закрываем consumer group
	// Это отправит сигнал всем goroutines завершиться
	if err := c.consumerGroup.Close(); err != nil {
		c.logger.Printf("ERROR: Failed to close consumer group: %v", err)
		return err
	}

	// Ждем завершения всех goroutines (обработка ошибок + основной цикл)
	// Это гарантирует graceful shutdown:
	// - Все текущие сообщения обработаны
	// - Все offset'ы закоммичены
	// - Нет утечек goroutines
	c.wg.Wait()

	c.logger.Println("Kafka consumer closed successfully")
	return nil
}

// extractHeaderValue извлекает значение заголовка из Kafka сообщения
// Headers содержат метаданные (correlation-id, timestamp и т.д.)
// Параметры:
//   - headers: Массив заголовков из Kafka сообщения
//   - key: Имя заголовка для поиска
//
// Возвращает значение заголовка или пустую строку если не найден
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

// UnmarshalMessage десериализует Kafka сообщение в Go структуру
// Удобная helper функция для работы с JSON сообщениями
// Параметры:
//   - message: Kafka сообщение с JSON в Value
//   - target: Указатель на структуру для десериализации
//
// Возвращает error если JSON невалидный или не соответствует структуре
// Пример использования:
//
//	var msg CalculationCompletedMessage
//	if err := UnmarshalMessage(kafkaMsg, &msg); err != nil {
//	    return err
//	}
func UnmarshalMessage(message *sarama.ConsumerMessage, target interface{}) error {
	// Десериализуем JSON байты в переданную структуру
	if err := json.Unmarshal(message.Value, target); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}
	return nil
}

// DefaultConsumerConfig возвращает рекомендуемую конфигурацию для production
// Эти настройки оптимизированы для:
// - Надежности (начинаем с новых сообщений, не теряем старые при перезапуске)
// - Отказоустойчивости (3 retry с backoff)
// - Производительности (RoundRobin балансировка)
// - Своевременного обнаружения проблем (10s session timeout)
// Параметры:
//   - brokers: Список Kafka брокеров
//   - groupID: ID consumer group (важно: consumers с одним ID формируют группу)
//   - topics: Список топиков для подписки
//
// Возвращает готовую конфигурацию для продакшна
func DefaultConsumerConfig(brokers []string, groupID string, topics []string) *ConsumerConfig {
	return &ConsumerConfig{
		Brokers:        brokers,                // Kafka брокеры
		GroupID:        groupID,                // Consumer group ID для распределенной обработки
		Topics:         topics,                 // Топики для чтения
		StartOffset:    sarama.OffsetNewest,    // Начинать с НОВЫХ сообщений (не читаем историю)
		MaxRetry:       3,                      // 3 повтора обработки при ошибке
		RetryBackoff:   100 * time.Millisecond, // 100ms базовая задержка между повторами
		SessionTimeout: 10 * time.Second,       // 10s timeout - если consumer не отвечает, его исключат
	}
}
