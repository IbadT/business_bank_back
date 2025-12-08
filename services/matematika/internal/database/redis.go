package database

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Host string
	Port int
}

func NewRedisConfig() *RedisConfig {
	return &RedisConfig{
		Host: GetEnv("REDIS_HOST", "redis"),
		Port: GetEnvInt("REDIS_PORT", 6379),
	}
}

func InitRedis() *redis.Client {
	config := NewRedisConfig()
	return redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", config.Host, config.Port),
	})
}

func HealthCheckRedis() string {
	client := InitRedis()
	defer client.Close()

	if err := client.Ping(context.Background()).Err(); err != nil {
		return "disconnected"
	}

	return "connected"
}
