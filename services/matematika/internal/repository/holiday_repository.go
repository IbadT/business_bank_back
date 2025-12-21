package repository

import (
	"errors"
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/internal/domain"
	"github.com/IbadT/business_bank_back/services/matematika/internal/models"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/logger"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type HolidayRepository interface {
	GetByDate(dateStr string) (domain.Holiday, error)
	Create(holiday *domain.Holiday) error
	GetHolidays(year time.Time) ([]domain.Holiday, error)
	Update(id uuid.UUID, holiday *domain.Holiday) error
	Delete(id uuid.UUID) error
}

type holidayRepository struct {
	DB *gorm.DB

}

func NewHolidayRepository(db *gorm.DB) HolidayRepository {
	return &holidayRepository{
		DB: db,
	}
}

func (r *holidayRepository) GetByDate(dateStr string) (domain.Holiday, error) {
	op := "repository.holiday.getByDate"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"date": dateStr})
	log.Info("Getting holiday by date")

	// Парсим дату из строки для поиска в БД
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		log.Error(err, "Invalid date format")
		return domain.Holiday{}, err
	}
	
	var holidayModel models.Holiday
	if err := r.DB.Where("holiday_date = ?", date).First(&holidayModel).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Warn("Holiday not found for date")
		} else {
			log.Error(err, "Failed to get holiday by date")
		}
		return domain.Holiday{}, err
	}
	
	log.Success("Holiday retrieved by date")
	// Конвертируем models.Holiday в domain.Holiday
	return domain.Holiday{
		HolidayDate: holidayModel.HolidayDate.Format("2006-01-02"),
		Name:        holidayModel.Name,
		Country:     holidayModel.Country,
	}, nil
}

func (r *holidayRepository) Create(holiday *domain.Holiday) error {
	op := "repository.holiday.create"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"name":    holiday.Name,
		"date":    holiday.HolidayDate,
		"country": holiday.Country,
	})
	log.Info("Creating holiday")

	// Парсим дату из строки
	holidayDate, err := time.Parse("2006-01-02", holiday.HolidayDate)
	if err != nil {
		log.Error(err, "Invalid date format")
		return err
	}
	
	// Конвертируем domain.Holiday в models.Holiday для сохранения в БД
	holidayModel := models.Holiday{
		HolidayDate: holidayDate,
		Name:        holiday.Name,
		Country:     holiday.Country,
	}
	
	if err := r.DB.Create(&holidayModel).Error; err != nil {
		log.Error(err, "Failed to create holiday")
		return err
	}

	log.Success("Holiday created successfully")
	return nil
}

func (r *holidayRepository) GetHolidays(year time.Time) ([]domain.Holiday, error) {
	op := "repository.holiday.getHolidays"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"year": year.Year()})
	log.Info("Getting holidays for year")

	var holidayModels []models.Holiday
	startDate := time.Date(year.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year.Year(), 12, 31, 23, 59, 59, 999999999, time.UTC)
	
	if err := r.DB.
		Where("holiday_date BETWEEN ? AND ?", startDate, endDate).
		Find(&holidayModels).
		Error; err != nil {
		log.Error(err, "Failed to get holidays")
		return []domain.Holiday{}, err
	}
	
	// Конвертируем models.Holiday в domain.Holiday
	holidays := make([]domain.Holiday, len(holidayModels))
	for i, hm := range holidayModels {
		holidays[i] = domain.Holiday{
			HolidayDate: hm.HolidayDate.Format("2006-01-02"),
			Name:        hm.Name,
			Country:     hm.Country,
		}
	}
	
	log.WithFields(logger.Fields{"count": len(holidays)}).Success("Holidays retrieved for year")
	return holidays, nil
}

func (r *holidayRepository) Update(id uuid.UUID, holiday *domain.Holiday) error {
	op := "repository.holiday.update"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"holiday_id": id})
	log.Info("Updating holiday")

	// Парсим дату из строки
	holidayDate, err := time.Parse("2006-01-02", holiday.HolidayDate)
	if err != nil {
		log.Error(err, "Invalid date format")
		return err
	}
	
	// Конвертируем domain.Holiday в models.Holiday для обновления в БД
	holidayModel := models.Holiday{
		HolidayDate: holidayDate,
		Name:        holiday.Name,
		Country:     holiday.Country,
	}
	
	if err := r.DB.
		Model(&models.Holiday{}).
		Where("id = ?", id).
		Updates(&holidayModel).
		Error; err != nil {
		log.Error(err, "Failed to update holiday")
		return err
	}

	log.Success("Holiday updated successfully")
	return nil
}

func (r *holidayRepository) Delete(id uuid.UUID) error {
	op := "repository.holiday.delete"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"holiday_id": id})
	log.Info("Deleting holiday")

	if err := r.DB.Where("id = ?", id).Delete(&models.Holiday{}).Error; err != nil {
		log.Error(err, "Failed to delete holiday")
		return err
	}

	log.Success("Holiday deleted successfully")
	return nil
}