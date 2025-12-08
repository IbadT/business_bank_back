// internal/transport/http/middleware/auth.go
package middleware

import (
	"fmt"
	"strings"

	"github.com/IbadT/business_bank_back/services/matematika/internal/database"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

const (
	UserIDKey = "user_id" // Ключ для хранения userID в контексте Echo
	UserRoleKey = "user_role" // Ключ для хранения роли пользователя в контексте Echo
)

// JWTConfig - конфигурация для JWT middleware
type JWTConfig struct {
	SecretKey string
	Skipper   func(c echo.Context) bool
}

// DefaultJWTConfig возвращает конфигурацию по умолчанию
func DefaultJWTConfig() JWTConfig {
	return JWTConfig{
		SecretKey: getJWTSecret(),
		Skipper: func(c echo.Context) bool {
			// Пропускаем Swagger и публичные эндпоинты
			path := c.Request().URL.Path
			return strings.HasPrefix(path, "/swagger") || 
				   path == "/health" || 
				   path == "/api/health" ||
				   path == "/api/login" ||
				   path == "/api/register"
		},
	}
}

// getJWTSecret получает секретный ключ из переменных окружения
func getJWTSecret() string {
	// secret := os.Getenv("JWT_SECRET")
	secret := database.GetEnv("JWT_SECRET", "super-secret-word")
	if secret == "" {
		// В продакшене это должно быть обязательно установлено
		return "super-secret-word"
	}
	return secret
}

// JWTAuthMiddleware - middleware для проверки JWT токена
func JWTAuthMiddleware(config JWTConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Пропускаем если Skipper возвращает true
			if config.Skipper != nil && config.Skipper(c) {
				return next(c)
			}

			// Извлекаем токен из заголовка Authorization
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(401, map[string]string{
					"error": "Authorization header is required",
				})
			}

			// Обрабатываем формат "Bearer <token>" или просто "<token>"
			var tokenString string
			authHeader = strings.TrimSpace(authHeader)
			
			if strings.HasPrefix(authHeader, "Bearer ") {
				// Формат "Bearer <token>"
				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
				tokenString = strings.TrimSpace(tokenString)
			} else {
				// Просто токен без префикса (для Swagger UI)
				tokenString = authHeader
			}
			
			if tokenString == "" {
				return c.JSON(401, map[string]string{
					"error": "Token is required",
				})
			}

			// Парсим и валидируем токен
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				// Проверяем алгоритм подписи
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, echo.ErrUnauthorized
				}
				return []byte(config.SecretKey), nil
			})

			if err != nil {
				// Логируем ошибку для отладки
				c.Logger().Debugf("JWT parse error: %v", err)
				return c.JSON(401, map[string]string{
					"error": "Invalid or expired token",
					"details": err.Error(),
				})
			}
			
			if !token.Valid {
				return c.JSON(401, map[string]string{
					"error": "Invalid or expired token",
				})
			}

			// Извлекаем claims
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return c.JSON(401, map[string]string{
					"error": "Invalid token claims",
				})
			}

			// Извлекаем userID из claims (обычно это "sub" или "user_id")
			// UUID может быть сохранен как строка или как интерфейс
			var userIDStr string
			subValue, exists := claims["sub"]
			if !exists {
				// Пробуем альтернативный ключ
				subValue, exists = claims["user_id"]
				if !exists {
					return c.JSON(401, map[string]string{
						"error": "User ID not found in token",
					})
				}
			}
			
			// Преобразуем в строку (может быть string или другой тип)
			switch v := subValue.(type) {
			case string:
				userIDStr = v
			case fmt.Stringer:
				userIDStr = v.String()
			default:
				// Пробуем преобразовать через fmt.Sprintf
				userIDStr = fmt.Sprintf("%v", v)
			}
			
			if userIDStr == "" {
				return c.JSON(401, map[string]string{
					"error": "User ID is empty in token",
				})
			}

			// Извлекаем роль из claims (если есть)
			roleStr, _ := claims["role"].(string)
			if roleStr == "" {
				// По умолчанию роль "user"
				roleStr = "user"
			}

			// Сохраняем userID и роль в контексте Echo для использования в handlers
			c.Set(UserIDKey, userIDStr)
			c.Set(UserRoleKey, roleStr)

			return next(c)
		}
	}
}

// GetUserID извлекает userID из контекста Echo
func GetUserID(c echo.Context) *string {
	userID, ok := c.Get(UserIDKey).(string)
	if !ok || userID == "" {
		return nil
	}
	return &userID
}

// GetUserRole извлекает роль пользователя из контекста Echo
func GetUserRole(c echo.Context) string {
	role, ok := c.Get(UserRoleKey).(string)
	if !ok || role == "" {
		return "user" // По умолчанию роль "user"
	}
	return role
}

// RequireRole создает middleware для проверки роли пользователя
func RequireRole(allowedRoles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userRole := GetUserRole(c)
			
			// Проверяем, есть ли роль пользователя в списке разрешенных
			allowed := false
			for _, role := range allowedRoles {
				if userRole == role {
					allowed = true
					break
				}
			}
			
			if !allowed {
				return c.JSON(403, map[string]string{
					"error": "Insufficient permissions. Required roles: " + strings.Join(allowedRoles, ", "),
				})
			}
			
			return next(c)
		}
	}
}

// RequireAdmin создает middleware для проверки, что пользователь является администратором
func RequireAdmin() echo.MiddlewareFunc {
	return RequireRole("admin")
}
