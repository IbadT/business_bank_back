package utils

import (
	"github.com/google/uuid"
)

// ParseUserID конвертирует userID из строки в UUID
// Возвращает nil, если userIDStr пустой или невалидный
func ParseUserID(userIDStr *string) *uuid.UUID {
	if userIDStr == nil || *userIDStr == "" {
		return nil
	}

	userID, err := uuid.Parse(*userIDStr)
	if err != nil {
		return nil
	}

	return &userID
}
