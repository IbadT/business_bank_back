package transportgrpc

import (
	"context"
	"strings"

	jwt_pkg "github.com/IbadT/business_bank_back/services/matematika/pkg/jwt"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	// MetadataKeyAuthorization - ключ для авторизации в metadata
	MetadataKeyAuthorization = "authorization"
	// MetadataKeyUserID - ключ для user_id в metadata (если передается напрямую)
	MetadataKeyUserID = "user_id"
)

// getUserIDFromMetadata извлекает user_id из gRPC metadata
// Сначала проверяет заголовок "user_id", если его нет - извлекает из JWT токена в заголовке "authorization"
func getUserIDFromMetadata(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	// Сначала пробуем получить user_id напрямую из metadata
	userIDValues := md.Get(MetadataKeyUserID)
	if len(userIDValues) > 0 && userIDValues[0] != "" {
		return strings.TrimSpace(userIDValues[0]), nil
	}

	// Если user_id нет, пробуем извлечь из JWT токена
	authValues := md.Get(MetadataKeyAuthorization)
	if len(authValues) == 0 || authValues[0] == "" {
		return "", status.Errorf(codes.Unauthenticated, "authorization header is not provided")
	}

	// Убираем префикс "Bearer " если есть
	tokenString := strings.TrimSpace(authValues[0])
	if strings.HasPrefix(tokenString, "Bearer ") {
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		tokenString = strings.TrimSpace(tokenString)
	}

	if tokenString == "" {
		return "", status.Errorf(codes.Unauthenticated, "authorization token is empty")
	}

	// Извлекаем claims из JWT токена
	claims, err := jwt_pkg.ExtractClaims(tokenString)
	if err != nil {
		return "", status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	// Извлекаем userID из claims
	userData, err := jwt_pkg.GetDataFromClaims(claims)
	if err != nil {
		return "", status.Errorf(codes.Unauthenticated, "failed to extract user_id from token: %v", err)
	}

	return userData.UserID, nil
}

// getUserIDFromMetadataSafe безопасно извлекает user_id из metadata
// Возвращает пустую строку и nil error, если user_id не найден (для опциональных случаев)
func getUserIDFromMetadataSafe(ctx context.Context, log *logger.Logger) *string {
	userID, err := getUserIDFromMetadata(ctx)
	if err != nil {
		if log != nil {
			log.Debug("user_id not found in metadata: %v", err)
		}
		return nil
	}
	return &userID
}
