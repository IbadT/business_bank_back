package middleware

import (
	"net/http"

	"github.com/IbadT/business_bank_back/services/matematika/internal/database"
	"github.com/labstack/echo/v4"
)

// RateLimitMiddleware создает Echo middleware из rate limiter
func RateLimitMiddleware(rateLimiter *database.RateLimiter) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Создаем адаптер для Echo
			adapter := &echoAdapter{
				context: c,
				next:    next,
			}

			// Создаем http.Handler из rate limiter
			// Если rate limit прошел, вызываем next handler
			rateLimitHandler := rateLimiter.Limit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Это будет вызвано только если rate limit прошел
				// Вызываем next handler Echo через адаптер
				adapter.err = next(c)
			}))

			// Вызываем rate limit handler
			rateLimitHandler.ServeHTTP(adapter, c.Request())

			// Если rate limit не был превышен, next уже был вызван в адаптере
			// Если был превышен, ответ уже отправлен
			return adapter.err
		}
	}
}

// echoAdapter адаптирует Echo context к http.ResponseWriter
type echoAdapter struct {
	context echo.Context
	next    echo.HandlerFunc
	err     error
	written bool
}

func (a *echoAdapter) Header() http.Header {
	return a.context.Response().Header()
}

func (a *echoAdapter) Write(b []byte) (int, error) {
	if !a.written {
		a.written = true
		// Если rate limit прошел, вызываем next handler
		if a.next != nil {
			a.err = a.next(a.context)
		}
	}
	return a.context.Response().Write(b)
}

func (a *echoAdapter) WriteHeader(statusCode int) {
	if !a.written {
		a.written = true
		a.context.Response().Status = statusCode
		// Если это не ошибка rate limit, вызываем next
		if statusCode != http.StatusTooManyRequests && a.next != nil {
			a.err = a.next(a.context)
		}
	}
}

