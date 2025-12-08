// internal/domain/entities/holiday.go
package entities

import "time"

// Holiday - доменная сущность праздника [32]
type Holiday struct {
	Date    time.Time
	Name    string
	Country string
}

// NewHoliday создает новый праздник
func NewHoliday(date time.Time, name string, country string) *Holiday {
	return &Holiday{
		Date:    date,
		Name:    name,
		Country: country,
	}
}

// IsHoliday проверяет, является ли дата праздником
func (h *Holiday) IsHoliday(date time.Time) bool {
	return h.Date.Year() == date.Year() &&
		h.Date.Month() == date.Month() &&
		h.Date.Day() == date.Day()
}