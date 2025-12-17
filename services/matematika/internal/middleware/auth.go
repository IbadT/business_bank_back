package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/internal/database"
	jwt_pkg "github.com/IbadT/business_bank_back/services/matematika/pkg/jwt"
	"github.com/labstack/echo/v4"
)

const (
	UserIDKey   = "user_id"   // Ключ для хранения userID в контексте Echo
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
			// Пропускаем Swagger, публичные эндпоинты и pprof
			path := c.Request().URL.Path
			return strings.HasPrefix(path, "/swagger") ||
				strings.HasPrefix(path, "/debug/pprof") ||
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

			// Извлекаем access_token из cookie
			accessCookie, err := c.Cookie("access_token")
			var tokenString string
			var needRefresh bool

			if err != nil || accessCookie == nil || accessCookie.Value == "" {
				// Access token отсутствует, пробуем refresh
				needRefresh = true
			} else {
				tokenString = strings.TrimSpace(accessCookie.Value)
				if tokenString == "" {
					needRefresh = true
				} else {
					// Пробуем валидировать access_token
					if err := jwt_pkg.VerifyToken(tokenString); err != nil {
						// Токен истек или невалидный, пробуем refresh
						needRefresh = true
					}
				}
			}

			// Если access_token истек или отсутствует, пробуем refresh
			if needRefresh {
				// Извлекаем refresh_token из cookie
				refreshCookie, err := c.Cookie("refresh_token")
				if err != nil || refreshCookie == nil || refreshCookie.Value == "" {
					return c.JSON(401, map[string]string{
						"error": "Необходимо авторизоваться. Токены не найдены в cookies.",
					})
				}

				refreshTokenString := strings.TrimSpace(refreshCookie.Value)
				if refreshTokenString == "" {
					return c.JSON(401, map[string]string{
						"error": "Необходимо авторизоваться. Refresh token пустой.",
					})
				}

				// Генерируем новые токены из refresh_token
				newAccessToken, newRefreshToken, err := jwt_pkg.RefreshToken(refreshTokenString)
				if err != nil {
					return c.JSON(401, map[string]string{
						"error":   "Необходимо авторизоваться. Не удалось обновить токен.",
						"details": err.Error(),
					})
				}

				// Обновляем access_token в cookie
				accessCookie = &http.Cookie{
					Name:     "access_token",
					Value:    newAccessToken,
					Path:     "/",
					HttpOnly: true,
					Secure:   true,
					SameSite: http.SameSiteStrictMode,
					Expires:  time.Now().Add(time.Hour * 24 * 2), // 2 дня
				}
				c.SetCookie(accessCookie)

				// Обновляем refresh_token в cookie
				refreshCookie = &http.Cookie{
					Name:     "refresh_token",
					Value:    newRefreshToken,
					Path:     "/",
					HttpOnly: true,
					Secure:   true,
					SameSite: http.SameSiteStrictMode,
					Expires:  time.Now().Add(time.Hour * 24 * 2), // 2 дня
				}
				c.SetCookie(refreshCookie)

				// Используем новый access_token
				tokenString = newAccessToken
			}

			// Извлекаем claims из токена
			claims, err := jwt_pkg.ExtractClaims(tokenString)
			if err != nil {
				return c.JSON(401, map[string]string{
					"error":   "Invalid token",
					"details": err.Error(),
				})
			}

			// Извлекаем userID из claims
			userData, err := jwt_pkg.GetDataFromClaims(claims)
			if err != nil {
				return c.JSON(401, map[string]string{
					"error":   "Invalid token",
					"details": err.Error(),
				})
			}

			// Сохраняем userID и роль в контексте Echo для использования в handlers
			c.Set(UserIDKey, userData.UserID)
			c.Set(UserRoleKey, userData.Role)

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
