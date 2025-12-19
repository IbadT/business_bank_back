package helpers

import (
	"sort"

	"github.com/IbadT/business_bank_back/services/matematika/internal/domain/entities"
)

// SortTransactionsByDate сортирует транзакции по дате
// Использует sort.Slice для эффективной сортировки O(n log n)
// Сортирует по TransactionDate (время совершения транзакции)
// Если даты одинаковые, использует ID для стабильной сортировки
func SortTransactionsByDate(transactions []*entities.Transaction) []*entities.Transaction {
	// Используем sort.Slice для эффективной сортировки O(n log n)
	sort.Slice(transactions, func(i, j int) bool {
		// Сортируем по TransactionDate (время совершения транзакции)
		// Если даты одинаковые, используем ID для стабильной сортировки
		if transactions[i].TransactionDate.Equal(transactions[j].TransactionDate) {
			return transactions[i].ID < transactions[j].ID
		}
		return transactions[i].TransactionDate.Before(transactions[j].TransactionDate)
	})
	return transactions
}
