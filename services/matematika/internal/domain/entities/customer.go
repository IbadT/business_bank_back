// internal/domain/entities/customer.go
package entities

import "math/rand"

// Customer - доменная сущность клиента [36]
type Customer struct {
	ID        string
	Name      string
	Category  string
	PercentRange PercentRange
	TransactionRange TransactionRange
}

// PercentRange - диапазон процентов [37]
type PercentRange struct {
	Min float64
	Max float64
}

// NewCustomer создает нового клиента
func NewCustomer(id, name, category string, minPercent, maxPercent float64, minTx, maxTx int) *Customer {
	return &Customer{
		ID:       id,
		Name:     name,
		Category: category,
		PercentRange: PercentRange{
			Min: minPercent,
			Max: maxPercent,
		},
		TransactionRange: TransactionRange{
			Min: minTx,
			Max: maxTx,
		},
	}
}

// CalculateAmount рассчитывает сумму транзакции клиента
func (c *Customer) CalculateAmount(turnover float64) float64 {
	percentage := c.PercentRange.Min + 
		(rand.Float64() * (c.PercentRange.Max - c.PercentRange.Min))
	return turnover * percentage
}

// GetTransactionCount возвращает количество транзакций [36]
func (c *Customer) GetTransactionCount() int {
	if c.TransactionRange.Min == c.TransactionRange.Max {
		return c.TransactionRange.Min
	}
	return c.TransactionRange.Min + 
		rand.Intn(c.TransactionRange.Max-c.TransactionRange.Min+1)
}