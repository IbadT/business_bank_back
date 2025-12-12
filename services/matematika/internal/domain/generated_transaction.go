package domain

import (
	"fmt"
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

func NewGeneratedTransaction(requestID uuid.UUID, transactionID string, transactionDate time.Time, postingDate time.Time, transactionType string, category string, method string, amount float64) *GeneratedTransaction {
	// Если postingDate не указана, используем transactionDate
	if postingDate.IsZero() {
		postingDate = transactionDate
	}

	return &GeneratedTransaction{
		ID:              uuid.New(),
		RequestID:       requestID,
		TransactionID:   transactionID,
		TransactionDate: transactionDate,
		PostingDate:     postingDate,
		Type:            transactionType,
		Category:        category,
		Method:          method,
		Amount:          amount,
		// TODO: проверить
		IsManual: true, // Ручные транзакции помечаются как manual
	}
}

// GenerateTransactionID генерирует ID транзакции
func GenerateTransactionID(prefix string, num int) string {
	return fmt.Sprintf("t_%s_%03d", prefix, num)
}
