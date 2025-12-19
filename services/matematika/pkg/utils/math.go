// pkg/utils/math.go
package utils

import (
	"math"
	"math/rand"
)

func RoundToCents(amount float64) float64 {
	return math.Round(amount*100) / 100
}

// RandomPercentage генерирует случайный процент из диапазона [min, max]
func RandomPercentage(min, max float64) float64 {
	if min >= max {
		return min
	}
	return min + rand.Float64()*(max-min)
}

// DistributeAmount распределяет сумму между транзакциями
// Возвращает сумму для транзакции с индексом i из общего количества count
// Последняя транзакция получает остаток для точного соответствия общей сумме
func DistributeAmount(totalAmount float64, i, count int) float64 {
	if count <= 0 {
		return 0
	}
	if i == count-1 {
		// Последняя транзакция: корректируем для точного соответствия
		// Вычитаем уже распределенные суммы
		distributed := (totalAmount / float64(count)) * float64(i)
		return totalAmount - distributed
	}
	// Остальные: равномерное распределение
	return totalAmount / float64(count)
}
