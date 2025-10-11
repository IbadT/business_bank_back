package domain

import (
	"errors"

	calc "github.com/IbadT/business_bank_back/services/matematika/internal/calculation"
)

// Проверяет:
// ✅ Бизнес-инварианты
// ✅ Правила предметной области
// ✅ Ограничения бизнес-процесса
// ✅ Консистентность состояния объекта
// Где: Domain слой (независимо от источника данных)

// пример валидации для создания пользователя
func NewStatement(accountID string, businessType string, initialBalance float64) (*calc.Statement, error) {
	if accountID == "" {
		return nil, errors.New("accountID is required")
	}
	if businessType == "" {
		return nil, errors.New("businessType is required")
	}
	if initialBalance <= 0 {
		return nil, errors.New("initialBalance must be greater than 0")
	}
	return &calc.Statement{AccountID: accountID, BusinessType: businessType, InitialBalance: initialBalance}, nil
}
