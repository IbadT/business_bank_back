// internal/service/holiday_service.go
package holidayservice

import (
	"context"
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/internal/cache"
	"github.com/IbadT/business_bank_back/services/matematika/internal/domain"
	"github.com/IbadT/business_bank_back/services/matematika/internal/repository"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/helpers"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/logger"
	"github.com/google/uuid"
)

type HolidayService interface {
	IsHoliday(date time.Time) bool
	GetNextBusinessDay(date time.Time) time.Time
	AddHoliday(date time.Time, name string, country string) error
	GetHolidays(year time.Time) ([]domain.Holiday, error)
	UpdateHoliday(id uuid.UUID, holiday_date time.Time, name string, country string) error
	DeleteHoliday(id uuid.UUID) error
}

type holidayService struct {
	holidayRepo repository.HolidayRepository
	cache       *cache.CacheService
}

func NewHolidayService(holidayRepo repository.HolidayRepository, cache *cache.CacheService) HolidayService {
	return &holidayService{
		holidayRepo: holidayRepo,
		cache:       cache,
	}
}

func (hs *holidayService) IsHoliday(date time.Time) bool {
	op := "service.holiday.isHoliday"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"date": date.Format("2006-01-02")})
	log.Debug("Checking if date is holiday")

	dateStr := date.Format("2006-01-02")
	exists, hasData := hs.cache.IsHoliday(context.Background(), dateStr)

	// Если нашли в Redis - возвращаем результат
	if hasData {
		log.Debug("Holiday check result from cache: %v", exists)
		return exists
	}

	// Если данных в Redis нет - идем в БД
	_, err := hs.holidayRepo.GetByDate(dateStr)
	if err != nil {
		log.Debug("Date is not a holiday")
		return false
	}

	// Если holiday найден в БД - это выходной
	log.Debug("Date is a holiday")
	return true
}

func (hs *holidayService) GetNextBusinessDay(date time.Time) time.Time {
	op := "service.holiday.getNextBusinessDay"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"date": date.Format("2006-01-02")})
	log.Debug("Getting next business day")

	nextDay := date.AddDate(0, 0, 1)

	for nextDay.Weekday() == time.Saturday || nextDay.Weekday() == time.Sunday || hs.IsHoliday(nextDay) {
		nextDay = nextDay.AddDate(0, 0, 1)
	}

	log.WithFields(logger.Fields{"next_business_day": nextDay.Format("2006-01-02")}).Debug("Next business day calculated")
	return nextDay
}

func (hs *holidayService) AddHoliday(date time.Time, name string, country string) error {
	op := "service.holiday.addHoliday"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"name":    name,
		"date":    date.Format("2006-01-02"),
		"country": country,
	})
	log.Info("Adding holiday")

	holiday, err := domain.NewHoliday(date, name, country)
	if err != nil {
		log.Error(err, "Failed to create holiday domain object")
		return err
	}

	// проверить, существует ли праздник с такой датой
	existingHoliday, err := hs.holidayRepo.GetByDate(holiday.HolidayDate)
	if err == nil && existingHoliday.HolidayDate == holiday.HolidayDate {
		log.Warn("Holiday already exists for date: %s", holiday.HolidayDate)
		return helpers.ErrHolidayAlreadyExists
	}

	if err := hs.holidayRepo.Create(holiday); err != nil {
		log.Error(err, "Failed to create holiday in repository")
		return err
	}

	// удаляем кэш праздников
	hs.cache.DelHolidays(context.Background())

	log.Success("Holiday added successfully")
	return nil
}

// TODO: добавить redis для кэша
func (hs *holidayService) GetHolidays(year time.Time) ([]domain.Holiday, error) {
	op := "service.holiday.getHolidays"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"year": year.Year()})
	log.Info("Getting holidays for year")

	holidays, err := hs.cache.GetHolidays(context.Background())
	if err == nil && len(holidays) > 0 {
		log.Debug("Holidays retrieved from cache")
		return holidays, nil
	}

	holidays, err = hs.holidayRepo.GetHolidays(year)
	if err != nil {
		log.Error(err, "Failed to get holidays from repository")
		return nil, err
	}
	hs.cache.SetHolidays(context.Background(), holidays)
	
	log.WithFields(logger.Fields{"count": len(holidays)}).Success("Holidays retrieved for year")
	return holidays, nil
}

func (hs *holidayService) UpdateHoliday(id uuid.UUID, holiday_date time.Time, name, country string) error {
	op := "service.holiday.updateHoliday"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"holiday_id": id,
		"name":       name,
		"date":       holiday_date.Format("2006-01-02"),
		"country":    country,
	})
	log.Info("Updating holiday")

	holiday, err := domain.NewHoliday(holiday_date, name, country)
	if err != nil {
		log.Error(err, "Failed to create holiday domain object")
		return err
	}
	if err := hs.holidayRepo.Update(id, holiday); err != nil {
		log.Error(err, "Failed to update holiday in repository")
		return err
	}
	// удаляем кэш праздников
	hs.cache.DelHolidays(context.Background())
	
	log.Success("Holiday updated successfully")
	return nil
}

func (hs *holidayService) DeleteHoliday(id uuid.UUID) error {
	op := "service.holiday.deleteHoliday"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"holiday_id": id})
	log.Info("Deleting holiday")

	if err := hs.holidayRepo.Delete(id); err != nil {
		log.Error(err, "Failed to delete holiday from repository")
		return err
	}
	// удаляем кэш праздников
	hs.cache.DelHolidays(context.Background())
	
	log.Success("Holiday deleted successfully")
	return nil
}
