// internal/domain/entities/transaction.go
package entities

import (
	"errors"
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/internal/domain/value_objects"
)

// Transaction - доменная сущность транзакции [44]
type Transaction struct {
	ID                string
	TransactionDate   time.Time
	PostingDate       time.Time
	Type              value_objects.TransactionType
	Category          string
	Method            value_objects.PaymentMethod
	Amount            float64
	BalanceAfter      float64
	IsManual          bool
	CalculationDetails map[string]interface{}
}

// NewTransaction - фабричный метод создания транзакции
func NewTransaction(
	id string,
	transactionDate time.Time,
	postingDate time.Time,
	transactionType value_objects.TransactionType,
	category string,
	method value_objects.PaymentMethod,
	amount float64,
) *Transaction {
	return &Transaction{
		ID:              id,
		TransactionDate: transactionDate,
		PostingDate:     postingDate,
		Type:            transactionType,
		Category:        category,
		Method:          method,
		Amount:          amount,
		IsManual:        false,
	}
}

// SetBalanceAfter устанавливает баланс после транзакции [44]
func (t *Transaction) SetBalanceAfter(balance float64) {
	t.BalanceAfter = balance
}

// SetCalculationDetails устанавливает детали расчета [20][21]
func (t *Transaction) SetCalculationDetails(details map[string]interface{}) {
	t.CalculationDetails = details
}

// SetManual устанавливает флаг ручной транзакции [44]
func (t *Transaction) SetManual(isManual bool) {
	t.IsManual = isManual
}

// IsIncome проверяет, является ли транзакция доходом
func (t *Transaction) IsIncome() bool {
	return t.Type == value_objects.Income
}

// IsExpense проверяет, является ли транзакция расходом
func (t *Transaction) IsExpense() bool {
	return t.Type == value_objects.Expense
}

// IsValid проверяет валидность транзакции [43]
func (t *Transaction) IsValid() error {
	if t.Amount == 0 {
		return ErrInvalidAmount
	}
	if t.Category == "" {
		return ErrInvalidCategory
	}
	return nil
}

var (
	ErrInvalidAmount = errors.New("invalid transaction amount")
	ErrInvalidCategory = errors.New("invalid transaction category")
)