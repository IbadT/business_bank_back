package kafka

import "github.com/IBM/sarama" // Kafka клиент для Go

// NewKafkaConfig создает базовую конфигурацию для Kafka
// Используется как основа для producer и consumer конфигураций
// Возвращает sarama.Config с оптимальными настройками для production
func NewKafkaConfig() *sarama.Config {
	// Создаем новую конфигурацию с дефолтными значениями
	config := sarama.NewConfig()

	// Указываем версию Kafka протокола
	// V3_0_0_0 - стабильная версия с поддержкой всех современных фич
	config.Version = sarama.V3_0_0_0

	// Producer настройки (будут переопределены в NewProducer):
	config.Producer.RequiredAcks = sarama.WaitForAll // Максимальная надежность
	config.Producer.Retry.Max = 5                    // 5 повторов при ошибке
	config.Producer.Return.Successes = true          // Возвращать успешные отправки
	config.Producer.Idempotent = true

	return config
}

// NewKafkaProducer создает простой Kafka producer (deprecated)
// DEPRECATED: Используйте NewProducer() из producer.go для production
// Эта функция оставлена для обратной совместимости
// Параметры:
//   - brokers: Список Kafka брокеров
//
// Возвращает SyncProducer или error
func NewKafkaProducer(brokers []string) (*sarama.SyncProducer, error) {
	config := NewKafkaConfig()
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}
	return &producer, nil
}

func HealthCheckKafka(kafkaBrokers []string) string {
	kafkaStatus := "disconnected"
	if len(kafkaBrokers) > 0 {
		config := NewKafkaConfig()
		client, err := sarama.NewClient(kafkaBrokers, config)
		if err == nil {
			brokers := client.Brokers()
			if len(brokers) > 0 {
				kafkaStatus = "connected"
			}
			client.Close()
		}
	}
	return kafkaStatus
}
