package jwt_pkg

import (
	"errors"
	"fmt"
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/internal/database"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func GenerateTokens(userID uuid.UUID) (string, string, error) {
	// Сохраняем UUID как строку для корректной работы с JWT
	userIDStr := userID.String()
	
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userIDStr,
		"exp": time.Now().Add(time.Hour * 4).Unix(), // 4 часа
	})

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userIDStr,
		"exp": time.Now().Add(time.Hour * 24 * 2).Unix(), // 2 дня
	})

	// Используем тот же дефолтный ключ, что и в middleware
	secretKey := database.GetEnv("JWT_SECRET", "super-secret-word")
	if secretKey == "" {
		secretKey = "super-secret-word"
	}

	accessTokenStr, err := accessToken.SignedString([]byte(secretKey))
	if err != nil {
		return "", "", err
	}

	refreshTokenStr, err := refreshToken.SignedString([]byte(secretKey))
	if err != nil {
		return "", "", err
	}

	return accessTokenStr, refreshTokenStr, nil
}

func VerifyToken(tokenString string) error {
	secretKey := database.GetEnv("JWT_SECRET", "super-secret-word")
	if secretKey == "" {
		secretKey = "super-secret-word"
	}
	
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверяем алгоритм подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return err
	}

	if !token.Valid {
		return errors.New("invalid token")
	}

	return nil
}

func ExtractClaims(tokenString string) (jwt.MapClaims, error) {
	secretKey := database.GetEnv("JWT_SECRET", "super-secret-word")
	if secretKey == "" {
		secretKey = "super-secret-word"
	}
	
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверяем алгоритм подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// UserData содержит данные пользователя из JWT токена
type UserData struct {
	UserID string
	Role   string
}

// GetDataFromClaims извлекает userID и роль из JWT claims
func GetDataFromClaims(claim jwt.MapClaims) (*UserData, error) {
	// Извлекаем userID из claims (обычно это "sub" или "user_id")
	var userIDStr string
	subValue, exists := claim["sub"]
	if !exists {
		// Пробуем альтернативный ключ
		subValue, exists = claim["user_id"]
		if !exists {
			return nil, errors.New("user ID not found in claims")
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
		return nil, errors.New("user ID is empty in claims")
	}

	// Извлекаем роль из claims (если есть)
	roleStr, _ := claim["role"].(string)
	if roleStr == "" {
		// По умолчанию роль "user"
		roleStr = "user"
	}

	return &UserData{
		UserID: userIDStr,
		Role:   roleStr,
	}, nil
}