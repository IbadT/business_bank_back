package utils

import "fmt"

// FormatPercentage форматирует процент в строку с 4 знаками после запятой
func FormatPercentage(percentage float64) string {
	return fmt.Sprintf("%.4f", percentage)
}

// FormatPercentagePercent форматирует процент в строку с символом "%" и 2 знаками после запятой
func FormatPercentagePercent(percentage float64) string {
	return fmt.Sprintf("%.2f%%", percentage*100)
}
