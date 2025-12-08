// internal/domain/value_objects/date.go
package value_objects

import "time"

// BusinessDate - Value Object бизнес-даты с учетом праздников [32][33]
type BusinessDate struct {
	Date      time.Time
	IsHoliday bool
	IsWeekend bool
}

// NewBusinessDate создает новую бизнес-дату
func NewBusinessDate(date time.Time, holidays []time.Time) *BusinessDate {
	isWeekend := date.Weekday() == time.Saturday || date.Weekday() == time.Sunday
	
	isHoliday := false
	for _, holiday := range holidays {
		if date.Year() == holiday.Year() &&
		   date.Month() == holiday.Month() &&
		   date.Day() == holiday.Day() {
			isHoliday = true
			break
		}
	}
	
	return &BusinessDate{
		Date:      date,
		IsHoliday: isHoliday,
		IsWeekend: isWeekend,
	}
}

// IsBusinessDay проверяет, является ли день рабочим
func (bd *BusinessDate) IsBusinessDay() bool {
	return !bd.IsHoliday && !bd.IsWeekend
}

// GetNextBusinessDay возвращает следующий рабочий день [32]
func (bd *BusinessDate) GetNextBusinessDay(holidays []time.Time) *BusinessDate {
	nextDay := bd.Date.AddDate(0, 0, 1)
	nextBusinessDate := NewBusinessDate(nextDay, holidays)
	
	for !nextBusinessDate.IsBusinessDay() {
		nextDay = nextDay.AddDate(0, 0, 1)
		nextBusinessDate = NewBusinessDate(nextDay, holidays)
	}
	
	return nextBusinessDate
}

// BusinessTime - Value Object бизнес-времени [33]
type BusinessTime struct {
	Time       time.Time
	IsBusiness bool
}

// NewBusinessTime создает новое бизнес-время
func NewBusinessTime(date time.Time, startHour, endHour int) *BusinessTime {
	hour := date.Hour()
	isBusiness := hour >= startHour && hour <= endHour
	
	return &BusinessTime{
		Time:       date,
		IsBusiness: isBusiness,
	}
}