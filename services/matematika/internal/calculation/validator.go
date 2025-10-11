package calculation

import (
	"errors"
)

type RequestValidator struct {
}

// NewRequestValidator - создает новый валидатор
func NewRequestValidator() *RequestValidator {
	return &RequestValidator{}
}

// ValidateRequest - полная валидация входного JSON
func (v *RequestValidator) ValidateRequest(req *GenerateStatementRequest) error {
	// TODO: реализовать валидацию согласно README.md
	return nil
}

// Custom errors
var (
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrFutureMonth         = errors.New("cannot generate statement for future month")
)

// formatValidationErrors - форматирование ошибок валидации
func (v *RequestValidator) FormatValidationErrors(err error) interface{} {
	return err.Error()
}
