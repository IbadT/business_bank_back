package kafka

import "github.com/IBM/sarama"

func NewKafkaConfig() *sarama.Config {
	config := sarama.NewConfig()

	config.Version = sarama.V3_0_0_0

	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true
	config.Producer.Idempotent = true

	return config
}

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
		}
		client.Close()
	}
	return kafkaStatus
}