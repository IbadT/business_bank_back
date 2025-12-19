package utils

import (
	"fmt"
	"time"
)

// FormatMonth форматирует год и месяц в строку формата "YYYY-MM"
func FormatMonth(year, month int) string {
	return fmt.Sprintf("%d-%02d", year, month)
}

// FirstDayOfMonth возвращает первый день указанного месяца
func FirstDayOfMonth(year, month int) time.Time {
	return time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
}
