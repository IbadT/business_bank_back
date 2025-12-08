// internal/service/holiday_service.go
package service

import (
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/internal/domain/entities"
)

type holidayService struct {
	holidays   []*entities.Holiday
	holidayMap map[string]bool
}

func newHolidayService(holidays []*entities.Holiday) *holidayService {
	holidayMap := make(map[string]bool)
	for _, holiday := range holidays {
		dateStr := holiday.Date.Format("2006-01-02")
		holidayMap[dateStr] = true
	}

	return &holidayService{
		holidays:   holidays,
		holidayMap: holidayMap,
	}
}

func (hs *holidayService) IsHoliday(date time.Time) bool {
	dateStr := date.Format("2006-01-02")
	return hs.holidayMap[dateStr]
}

func (hs *holidayService) GetNextBusinessDay(date time.Time) time.Time {
	nextDay := date.AddDate(0, 0, 1)

	for nextDay.Weekday() == time.Saturday || nextDay.Weekday() == time.Sunday || hs.IsHoliday(nextDay) {
		nextDay = nextDay.AddDate(0, 0, 1)
	}

	return nextDay
}

