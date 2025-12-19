package helpers

import (
	"github.com/google/uuid"
)

// ParseUUID парсит строку в UUID и возвращает ошибку из error_constants при неудаче
// Используется для requestID и других UUID
func ParseUUID(s string) (uuid.UUID, error) {
	if s == "" {
		return uuid.Nil, ErrRequestIDEmpty
	}
	
	parsed, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, ErrInvalidRequestID
	}
	
	return parsed, nil
}

// ParseUserID парсит строку в UUID для userID и возвращает ошибку из error_constants при неудаче
func ParseUserID(s string) (uuid.UUID, error) {
	if s == "" {
		return uuid.Nil, ErrUserIDRequired
	}
	
	parsed, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, ErrInvalidUserID
	}
	
	return parsed, nil
}

// ParseUUIDOrNil парсит строку в UUID, возвращает nil если строка пустая или невалидная
// Используется для опциональных UUID (например, когда userID может быть nil)
func ParseUUIDOrNil(s *string) *uuid.UUID {
	if s == nil || *s == "" {
		return nil
	}
	
	parsed, err := uuid.Parse(*s)
	if err != nil {
		return nil
	}
	
	return &parsed
}

// ParseHolidayID парсит строку в UUID для holiday ID и возвращает ошибку из error_constants при неудаче
func ParseHolidayID(s string) (uuid.UUID, error) {
	if s == "" {
		return uuid.Nil, ErrRequired
	}
	
	parsed, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, ErrInvalidFormat
	}
	
	return parsed, nil
}
