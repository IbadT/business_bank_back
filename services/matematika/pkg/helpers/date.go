package helpers

import "time"

// FindWeekdaysInMonth находит все дни указанного дня недели в месяце
// Возвращает слайс дат, отсортированный по возрастанию
func FindWeekdaysInMonth(year, month int, weekday time.Weekday) []time.Time {
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	var weekdays []time.Time
	current := firstDay
	
	for current.Month() == time.Month(month) {
		if current.Weekday() == weekday {
			weekdays = append(weekdays, current)
		}
		current = current.AddDate(0, 0, 1)
	}
	
	return weekdays
}

// GetNthWeekdayInMonth возвращает N-й день указанного дня недели в месяце (1-based)
// Если такого дня нет, возвращает последний найденный или первый день месяца
func GetNthWeekdayInMonth(year, month int, weekday time.Weekday, n int) time.Time {
	weekdays := FindWeekdaysInMonth(year, month, weekday)
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	
	if len(weekdays) >= n {
		return weekdays[n-1] // n-1 потому что 1-based
	} else if len(weekdays) > 0 {
		return weekdays[len(weekdays)-1]
	}
	return firstDay
}
