package utils

import "fmt"

// GenerateTransactionID генерирует ID транзакции
// Формат: t_{prefix}_{num:03d}
func GenerateTransactionID(prefix string, num int) string {
	return fmt.Sprintf("t_%s_%03d", prefix, num)
}

// GenerateTemplateTransactionID генерирует ID транзакции из шаблона категории
// Формат: t_exp_{category}_{num:03d}
func GenerateTemplateTransactionID(category string, num int) string {
	return fmt.Sprintf("t_exp_%s_%03d", category, num)
}
