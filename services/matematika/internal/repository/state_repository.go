// internal/repository/state_repository.go
package repository

import (
	"log"

	"github.com/IbadT/business_bank_back/services/matematika/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StateRepository - интерфейс для работы с состоянием генерации
type StateRepository interface {
	GetState(userID *uuid.UUID, stateKey string) (*models.GenerationState, error)
	SaveState(userID *uuid.UUID, stateKey string, stateValue models.JSONB) error
	GetSoftwareSubscriptionWeekday(userID *uuid.UUID) (int, error) // 0-6 (Sunday-Saturday)
	SaveSoftwareSubscriptionWeekday(userID *uuid.UUID, weekday int) error
}

type stateRepository struct {
	db *gorm.DB
}

// NewStateRepository создает новый StateRepository
func NewStateRepository(db *gorm.DB) StateRepository {
	return &stateRepository{db: db}
}

// GetState получает состояние по ключу
func (r *stateRepository) GetState(userID *uuid.UUID, stateKey string) (*models.GenerationState, error) {
	var state models.GenerationState
	query := r.db.Where("state_key = ?", stateKey)
	
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	} else {
		query = query.Where("user_id IS NULL")
	}
	
	if err := query.First(&state).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Состояние не найдено - это нормально
		}
		return nil, err
	}
	
	return &state, nil
}

// SaveState сохраняет состояние
func (r *stateRepository) SaveState(userID *uuid.UUID, stateKey string, stateValue models.JSONB) error {
	log.Printf("[DEBUG] SaveState called: userID=%v, stateKey=%s, stateValue=%+v", userID, stateKey, stateValue)
	
	state := models.GenerationState{
		UserID:    userID,
		StateKey:  stateKey,
		StateValue: stateValue,
	}
	
	// Используем FirstOrCreate с обновлением StateValue если запись существует
	var existing models.GenerationState
	query := r.db.Where("state_key = ?", stateKey)
	
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
		log.Printf("[DEBUG] Searching for state with userID=%v", *userID)
	} else {
		query = query.Where("user_id IS NULL")
		log.Printf("[DEBUG] Searching for state with userID IS NULL")
	}
	
	if err := query.First(&existing).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Создаем новую запись
			log.Printf("[DEBUG] State not found, creating new: stateKey=%s, userID=%v", stateKey, userID)
			if err := r.db.Create(&state).Error; err != nil {
				log.Printf("[ERROR] Failed to create state: %v", err)
				return err
			}
			log.Printf("[DEBUG] Successfully created state")
			return nil
		}
		log.Printf("[ERROR] Failed to query state: %v", err)
		return err
	}
	
	// Обновляем существующую запись
	log.Printf("[DEBUG] State found, updating: id=%v", existing.ID)
	existing.StateValue = stateValue
	if err := r.db.Save(&existing).Error; err != nil {
		log.Printf("[ERROR] Failed to update state: %v", err)
		return err
	}
	log.Printf("[DEBUG] Successfully updated state")
	return nil
}

// GetSoftwareSubscriptionWeekday получает сохраненный день недели для подписки ПО
func (r *stateRepository) GetSoftwareSubscriptionWeekday(userID *uuid.UUID) (int, error) {
	state, err := r.GetState(userID, "software_subscription_weekday")
	if err != nil {
		return -1, err
	}
	
	if state == nil {
		return -1, nil // Не найдено
	}
	
	weekday, ok := state.StateValue["weekday"].(float64)
	if !ok {
		return -1, nil
	}
	
	return int(weekday), nil
}

// SaveSoftwareSubscriptionWeekday сохраняет день недели для подписки ПО
func (r *stateRepository) SaveSoftwareSubscriptionWeekday(userID *uuid.UUID, weekday int) error {
	stateValue := models.JSONB{
		"weekday": weekday,
	}
	return r.SaveState(userID, "software_subscription_weekday", stateValue)
}
