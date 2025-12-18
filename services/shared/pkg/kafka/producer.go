package kafka

import (
	"context"
	"time"

	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"
)

type Producer interface {
	PublishMessage(ctx context.Context, msg interface{}) error
	Close() error
}

type KafkaProducer struct {
	producer sarama.SyncProducer
	config   *ProducerConfig
	logger   *logrus.Logger
}

type ProducerConfig struct {
	Brokers          []string
	RequiredAcks     sarama.RequiredAcks
	Compression      sarama.CompressionCodec
	MaxRetry         int
	RetryBackoff     time.Duration
	IdempotentWrites bool
}
