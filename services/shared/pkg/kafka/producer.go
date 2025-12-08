package kafka

import (
	"context"
	"log"
	"time"

	"github.com/IBM/sarama"
)

type Producer interface {
	PublishMessage(ctx context.Context, msg interface{}) error
	Close() error
}

type KafkaProducer struct {
	producer sarama.SyncProducer
	config *ProducerConfig
	logger *log.Logger
}

type ProducerConfig struct {
	Brokers []string
	RequiredAcks sarama.RequiredAcks
	Compression sarama.CompressionCodec
	MaxRetry int
	RetryBackoff time.Duration
	IdempotentWrites bool
}