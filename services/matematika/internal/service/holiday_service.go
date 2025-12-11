// internal/service/holiday_service.go
package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/internal/domain"
	"github.com/IbadT/business_bank_back/services/matematika/internal/repository"
	"github.com/google/uuid"
)

type holidayService struct {
	holidayRepo repository.HolidayRepository
}

type HolidayService interface {
	IsHoliday(date time.Time) bool
	GetNextBusinessDay(date time.Time) time.Time
	AddHoliday(date time.Time, name string, country string) error
	GetHolidays(year time.Time) ([]domain.Holiday, error)
	UpdateHoliday(id uuid.UUID, holiday_date time.Time, name string, country string) error
	DeleteHoliday(id uuid.UUID) error
}

func NewHolidayService(holidayRepo repository.HolidayRepository) HolidayService {
	return &holidayService{
		holidayRepo: holidayRepo,
	}
}

// TODO: добавить redis для кэша
func (hs *holidayService) IsHoliday(date time.Time) bool {
	dateStr := date.Format("2006-01-02")
	holiday, err := hs.holidayRepo.GetByDate(dateStr);
	if err != nil {
		return false
	}
	fmt.Println("HOLIDAY: ", holiday)
	return holiday.HolidayDate == dateStr
}

func (hs *holidayService) GetNextBusinessDay(date time.Time) time.Time {
	nextDay := date.AddDate(0, 0, 1)

	for nextDay.Weekday() == time.Saturday || nextDay.Weekday() == time.Sunday || hs.IsHoliday(nextDay) {
		nextDay = nextDay.AddDate(0, 0, 1)
	}

	return nextDay
}

func (hs *holidayService) AddHoliday(date time.Time, name string, country string) error {
	holiday, err := domain.NewHoliday(date, name, country)
	if err != nil  {
		return err
	}

	// проверить, существует ли праздник с такой датой
	existingHoliday, err := hs.holidayRepo.GetByDate(holiday.HolidayDate)
	if err == nil && existingHoliday.HolidayDate == holiday.HolidayDate {
		return errors.New("holiday already exists")
	}

	if err := hs.holidayRepo.Create(holiday); err != nil {
		return err
	}

	return nil
}

func (hs *holidayService) GetHolidays(year time.Time) ([]domain.Holiday, error) {
	holidays, err := hs.holidayRepo.GetHolidays(year)
	if err != nil {
		return nil, err
	}
	return holidays, nil
}

func (hs *holidayService) UpdateHoliday(id uuid.UUID, holiday_date time.Time, name, country string) error {
	holiday, err := domain.NewHoliday(holiday_date, name, country)
	if err != nil {
		return err
	}
	if err := hs.holidayRepo.Update(id, holiday); err != nil {
		return err
	}
	return nil
}

func (hs *holidayService) DeleteHoliday(id uuid.UUID) error {
	if err := hs.holidayRepo.Delete(id); err != nil {
		return err
	}
	return nil
}