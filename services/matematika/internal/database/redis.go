package database

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Host string
	Port int
}

type RateLimiter struct {
	limiter  *redis_rate.Limiter
	maxRate  int
	interval time.Duration
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

func NewRateLimiter(client *redis.Client, maxRate int, interval time.Duration) *RateLimiter {
	limiter := redis_rate.NewLimiter(client)
	return &RateLimiter{
		limiter:  limiter,
		maxRate:  maxRate,
		interval: interval,
	}
}

// extractClientIP извлекает реальный IP клиента с учетом proxy/load balancer
func extractClientIP(r *http.Request) string {
	// Проверяем X-Forwarded-For (может содержать несколько IP через запятую)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Проверяем X-Real-IP (устанавливается Nginx)
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fallback на RemoteAddr (убираем порт если есть)
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// Limit возвращает http.Handler middleware для rate limiting
func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := extractClientIP(r)

		// Определяем rate limit в зависимости от интервала
		var rate redis_rate.Limit
		switch rl.interval {
		case time.Second:
			rate = redis_rate.PerSecond(rl.maxRate)
		case time.Minute:
			rate = redis_rate.PerMinute(rl.maxRate)
		case time.Hour:
			rate = redis_rate.PerHour(rl.maxRate)
		default:
			rate = redis_rate.PerMinute(rl.maxRate)
		}

		// Проверяем лимит через Redis
		result, err := rl.limiter.Allow(context.Background(), clientIP, rate)

		// Устанавливаем headers (всегда)
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.maxRate))
		if err == nil {
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(int(result.Remaining)))
			w.Header().Set("X-RateLimit-Reset", strconv.Itoa(int(result.ResetAfter/time.Second)))
		}

		// Обработка ошибок Redis - fail-open (пропускаем запрос если Redis недоступен)
		if err != nil {
			// Логируем ошибку, но пропускаем запрос для отказоустойчивости
			next.ServeHTTP(w, r)
			return
		}

		// Проверяем, разрешен ли запрос
		if result.Allowed == 0 {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}

		// Запрос разрешен, передаем дальше
		next.ServeHTTP(w, r)
	})
}