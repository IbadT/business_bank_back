package domain

import (
	"time"

	"github.com/google/uuid"
)

type GeneratedTransaction struct {
	ID                 uuid.UUID
	RequestID          uuid.UUID
	TransactionID      string
	TransactionDate    time.Time
	PostingDate        time.Time
	Type               string
	Category           string
	Method             string
	Amount             float64
	BalanceAfter       *float64
	IsManual           bool
	SortOrder          *int
	CalculationDetails map[string]interface{} // Добавлено для сохранения calculation_details
}

// NewGeneratedTransaction создает новую транзакцию с базовыми полями
func NewGeneratedTransaction(
	requestID uuid.UUID,
	transactionID string,
	transactionDate time.Time,
	postingDate time.Time,
	transactionType string,
	category string,
	method string,
	amount float64,
) *GeneratedTransaction {
	// Если postingDate не указана, используем transactionDate
	if postingDate.IsZero() {
		postingDate = transactionDate
	}

	return &GeneratedTransaction{
		ID:                 uuid.New(),
		RequestID:          requestID,
		TransactionID:      transactionID,
		TransactionDate:    transactionDate,
		PostingDate:        postingDate,
		Type:               transactionType,
		Category:           category,
		Method:             method,
		Amount:             amount,
		IsManual:           false, // По умолчанию не ручная
		CalculationDetails: make(map[string]interface{}),
	}
}

// SetBalanceAfter устанавливает баланс после транзакции
func (gt *GeneratedTransaction) SetBalanceAfter(balance float64) *GeneratedTransaction {
	gt.BalanceAfter = &balance
	return gt
}

// SetIsManual помечает транзакцию как ручную
func (gt *GeneratedTransaction) SetIsManual(isManual bool) *GeneratedTransaction {
	gt.IsManual = isManual
	return gt
}

// SetSortOrder устанавливает порядок сортировки
func (gt *GeneratedTransaction) SetSortOrder(order int) *GeneratedTransaction {
	gt.SortOrder = &order
	return gt
}

// SetCalculationDetails устанавливает детали расчета
func (gt *GeneratedTransaction) SetCalculationDetails(details map[string]interface{}) *GeneratedTransaction {
	if details == nil {
		gt.CalculationDetails = make(map[string]interface{})
	} else {
		gt.CalculationDetails = details
	}
	return gt
}

// AddCalculationDetail добавляет одну деталь расчета
func (gt *GeneratedTransaction) AddCalculationDetail(key string, value interface{}) *GeneratedTransaction {
	if gt.CalculationDetails == nil {
		gt.CalculationDetails = make(map[string]interface{})
	}
	gt.CalculationDetails[key] = value
	return gt
}

// TransactionEntity интерфейс для работы с entities.Transaction
type TransactionEntity interface {
	GetID() string
	GetTransactionDate() time.Time
	GetPostingDate() time.Time
	GetType() string
	GetCategory() string
	GetMethod() string
	GetAmount() float64
	GetBalanceAfter() float64
	IsManualTransaction() bool
	GetCalculationDetails() map[string]interface{}
}

// FromEntityTransaction создает GeneratedTransaction из entities.Transaction
func FromEntityTransaction(
	requestID uuid.UUID,
	entityTx TransactionEntity,
	sortOrder int,
) *GeneratedTransaction {
	balanceAfter := entityTx.GetBalanceAfter()
	gt := NewGeneratedTransaction(
		requestID,
		entityTx.GetID(),
		entityTx.GetTransactionDate(),
		entityTx.GetPostingDate(),
		entityTx.GetType(),
		entityTx.GetCategory(),
		entityTx.GetMethod(),
		entityTx.GetAmount(),
	).
		SetBalanceAfter(balanceAfter).
		SetIsManual(entityTx.IsManualTransaction()).
		SetSortOrder(sortOrder).
		SetCalculationDetails(entityTx.GetCalculationDetails())

	return gt
}
