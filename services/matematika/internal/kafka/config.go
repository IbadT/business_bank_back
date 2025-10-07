package kafka

import "github.com/IBM/sarama"

func NewKafkaConfig() *sarama.Config {
	config := sarama.NewConfig()
	config.Version = sarama.V3_0_0_0
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true
	return config
}

func NewKafkaProducer(brokers []string) (sarama.SyncProducer, error) {
	config := NewKafkaConfig()
	return sarama.NewSyncProducer(brokers, config)
}
