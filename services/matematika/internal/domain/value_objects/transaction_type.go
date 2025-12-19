// internal/domain/value_objects/transaction_type.go
package value_objects

import (
	"github.com/IbadT/business_bank_back/services/matematika/pkg/helpers"
)

// TransactionType - Value Object типа транзакции [44]
type TransactionType string

const (
	Income  TransactionType = "income"
	Expense TransactionType = "expense"
)

// NewTransactionType создает новый тип транзакции с валидацией
func NewTransactionType(t string) (TransactionType, error) {
	switch t {
	case string(Income), string(Expense):
		return TransactionType(t), nil
	default:
		return "", helpers.ErrInvalidTransactionType
	}
}

// String возвращает строковое представление
func (tt TransactionType) String() string {
	return string(tt)
}

// IsValid проверяет валидность типа
func (tt TransactionType) IsValid() bool {
	return tt == Income || tt == Expense
}
