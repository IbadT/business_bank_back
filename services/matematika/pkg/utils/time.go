package utils

import (
	"strconv"
	"strings"
)

// ParseTimeFromString парсит время из строки "HH:MM"
// Возвращает (hour, minute). Если формат неверный, возвращает (8, 0) по умолчанию
func ParseTimeFromString(timeStr string) (int, int) {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return 8, 0 // По умолчанию
	}
	hour, _ := strconv.Atoi(parts[0])
	minute, _ := strconv.Atoi(parts[1])
	return hour, minute
}
