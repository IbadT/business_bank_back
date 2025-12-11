package repository

import (
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/internal/domain"
	"github.com/IbadT/business_bank_back/services/matematika/internal/models"
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
	// Парсим дату из строки для поиска в БД
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return domain.Holiday{}, err
	}
	
	var holidayModel models.Holiday
	if err := r.DB.Where("holiday_date = ?", date).First(&holidayModel).Error; err != nil {
		return domain.Holiday{}, err
	}
	
	// Конвертируем models.Holiday в domain.Holiday
	return domain.Holiday{
		HolidayDate: holidayModel.HolidayDate.Format("2006-01-02"),
		Name:        holidayModel.Name,
		Country:     holidayModel.Country,
	}, nil
}

func (r *holidayRepository) Create(holiday *domain.Holiday) error {
	// Парсим дату из строки
	holidayDate, err := time.Parse("2006-01-02", holiday.HolidayDate)
	if err != nil {
		return err
	}
	
	// Конвертируем domain.Holiday в models.Holiday для сохранения в БД
	holidayModel := models.Holiday{
		HolidayDate: holidayDate,
		Name:        holiday.Name,
		Country:     holiday.Country,
	}
	
	return r.DB.Create(&holidayModel).Error
}

func (r *holidayRepository) GetHolidays(year time.Time) ([]domain.Holiday, error) {
	var holidayModels []models.Holiday
	startDate := time.Date(year.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year.Year(), 12, 31, 23, 59, 59, 999999999, time.UTC)
	
	if err := r.DB.
		Where("holiday_date BETWEEN ? AND ?", startDate, endDate).
		Find(&holidayModels).
		Error; err != nil {
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
	
	return holidays, nil
}

func (r *holidayRepository) Update(id uuid.UUID, holiday *domain.Holiday) error {
	// Парсим дату из строки
	holidayDate, err := time.Parse("2006-01-02", holiday.HolidayDate)
	if err != nil {
		return err
	}
	
	// Конвертируем domain.Holiday в models.Holiday для обновления в БД
	holidayModel := models.Holiday{
		HolidayDate: holidayDate,
		Name:        holiday.Name,
		Country:     holiday.Country,
	}
	
	return r.DB.
		Model(&models.Holiday{}).
		Where("id = ?", id).
		Updates(&holidayModel).
		Error
}

func (r *holidayRepository) Delete(id uuid.UUID) error {
	return r.DB.Where("id = ?", id).Delete(&models.Holiday{}).Error
}