// internal/domain/entities/template.go
package entities

import (
	"math/rand"

	"github.com/IbadT/business_bank_back/services/matematika/internal/domain/value_objects"
)

// TransactionTemplate - доменная сущность шаблона транзакции [7-31]
type TransactionTemplate struct {
	ID              string
	Category        string
	Type            value_objects.TransactionType
	IsPercentage    bool
	PercentageRange PercentageRange
	FixedAmount     float64
	Schedule        Schedule
	BusinessHours   BusinessHours
	IsOptional      bool
	Priority        int
	PaymentMethod   value_objects.PaymentMethod
	TransactionRange TransactionRange
}

// PercentageRange - диапазон процентов [7-12]
type PercentageRange struct {
	Min float64
	Max float64
}

// Schedule - расписание транзакции [22-31]
type Schedule struct {
	Frequency     string   // monthly, biweekly, weekly, once
	PreferredDay  string   // Monday, Friday, 15
	WeekOfMonth   []int    // [2,4]
	MinOccurrences int
	MaxOccurrences int
}

// BusinessHours - рабочие часы [33]
type BusinessHours struct {
	Start string // "08:00"
	End   string // "18:00"
}

// TransactionRange - диапазон количества транзакций [2-4]
type TransactionRange struct {
	Min int
	Max int
}

// NewPercentageTemplate создает шаблон процентной операции [7-12]
func NewPercentageTemplate(
	category string,
	transactionType value_objects.TransactionType,
	paymentMethod value_objects.PaymentMethod,
	minPercent float64,
	maxPercent float64,
	schedule Schedule,
) *TransactionTemplate {
	return &TransactionTemplate{
		Category: category,
		Type:     transactionType,
		IsPercentage: true,
		PercentageRange: PercentageRange{
			Min: minPercent,
			Max: maxPercent,
		},
		Schedule:      schedule,
		PaymentMethod: paymentMethod,
		IsOptional:    false,
		Priority:      100,
	}
}

// NewFixedTemplate создает шаблон фиксированной операции [13-19]
func NewFixedTemplate(
	category string,
	transactionType value_objects.TransactionType,
	paymentMethod value_objects.PaymentMethod,
	fixedAmount float64,
	schedule Schedule,
) *TransactionTemplate {
	return &TransactionTemplate{
		Category:      category,
		Type:          transactionType,
		IsPercentage:  false,
		FixedAmount:   fixedAmount,
		Schedule:      schedule,
		PaymentMethod: paymentMethod,
		IsOptional:    false,
		Priority:      100,
	}
}

// CalculateAmount рассчитывает сумму на основе оборота
func (t *TransactionTemplate) CalculateAmount(turnover float64) float64 {
	if t.IsPercentage {
		percentage := t.PercentageRange.Min + 
			(rand.Float64() * (t.PercentageRange.Max - t.PercentageRange.Min))
		return turnover * percentage
	}
	return t.FixedAmount
}

// GetOccurrences возвращает количество вхождений в месяц
func (t *TransactionTemplate) GetOccurrences() int {
	if t.Schedule.MinOccurrences == t.Schedule.MaxOccurrences {
		return t.Schedule.MinOccurrences
	}
	return t.Schedule.MinOccurrences + 
		rand.Intn(t.Schedule.MaxOccurrences-t.Schedule.MinOccurrences+1)
}