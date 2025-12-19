// internal/repository/state_repository.go
package repository

import (
	"fmt"
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/internal/models"
	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/dto"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/logger"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StateRepository - интерфейс для работы с состоянием генерации
type StateRepository interface {
	GetState(userID uuid.UUID, stateKey string) (*models.GenerationState, error)
	SaveState(userID uuid.UUID, stateKey string, stateValue models.JSONB) error
	GetSoftwareSubscriptionWeekday(userID uuid.UUID) (int, error) // 0-6 (Sunday-Saturday)
	SaveSoftwareSubscriptionWeekday(userID uuid.UUID, weekday int) error

	GetMobileBaseAmount(userID uuid.UUID) (float64, error)
	SaveMobileBaseAmount(userID uuid.UUID, amount float64, firstMonth string) error

	GetUtilitiesBaseAmount(userID uuid.UUID) (float64, error)
	SaveUtilitiesBaseAmount(userID uuid.UUID, amount float64, firstMonth string) error

	GetLeasingBaseAmount(userID uuid.UUID) (float64, error)
	SaveLeasingBaseAmount(userID uuid.UUID, amount float64, firstMonth string, turnover float64) error

	// Методы для работы с associatedCard
	GetAssociatedCard(userID uuid.UUID) (string, error)
	SaveAssociatedCard(userID uuid.UUID, cardNumber string) error

	// Универсальный метод для получения всех базовых сумм
	GetBaseAmount(userID uuid.UUID) (*dto.BaseAmountsResponse, error)

	// Методы для удаления (reset)
	DeleteMobileBaseAmount(userID uuid.UUID) error
	DeleteUtilitiesBaseAmount(userID uuid.UUID) error
	DeleteLeasingBaseAmount(userID uuid.UUID) error
}

type stateRepository struct {
	db *gorm.DB
}

// NewStateRepository создает новый StateRepository
func NewStateRepository(db *gorm.DB) StateRepository {
	return &stateRepository{db: db}
}

// GetState получает состояние по ключу
func (r *stateRepository) GetState(userID uuid.UUID, stateKey string) (*models.GenerationState, error) {
	op := "repository.state.getState"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"state_key": stateKey,
		"user_id":   userID,
	})
	log.Debug("Getting state")

	var state models.GenerationState
	query := r.db.Where("state_key = ?", stateKey)

	if userID != uuid.Nil {
		query = query.Where("user_id = ?", userID)
	} else {
		query = query.Where("user_id IS NULL")
	}

	if err := query.First(&state).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Debug("State not found")
			return nil, nil // Состояние не найдено - это нормально
		}
		log.Error(err, "Failed to get state")
		return nil, err
	}

	log.Debug("State retrieved")
	return &state, nil
}

// SaveState сохраняет состояние
func (r *stateRepository) SaveState(userID uuid.UUID, stateKey string, stateValue models.JSONB) error {
	op := "repository.state.saveState"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"state_key": stateKey,
		"user_id":   userID,
	})
	log.Debug("Saving state")

	// Правильно обрабатываем uuid.Nil - сохраняем nil указатель вместо указателя на uuid.Nil
	var userIDPtr *uuid.UUID
	if userID != uuid.Nil {
		userIDPtr = &userID
	}

	state := models.GenerationState{
		UserID:     userIDPtr,
		StateKey:   stateKey,
		StateValue: stateValue,
	}

	// Используем FirstOrCreate с обновлением StateValue если запись существует
	var existing models.GenerationState
	query := r.db.Where("state_key = ?", stateKey)

	if userID != uuid.Nil {
		query = query.Where("user_id = ?", userID)
	} else {
		query = query.Where("user_id IS NULL")
	}

	if err := query.First(&existing).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Создаем новую запись
			if err := r.db.Create(&state).Error; err != nil {
				log.Error(err, "Failed to create state")
				return err
			}
			log.Debug("State created successfully")
			return nil
		}
		log.Error(err, "Failed to query state")
		return err
	}

	// Обновляем существующую запись
	existing.StateValue = stateValue
	if err := r.db.Save(&existing).Error; err != nil {
		log.Error(err, "Failed to update state")
		return err
	}
	log.Debug("State updated successfully")
	return nil
}

// GetSoftwareSubscriptionWeekday получает сохраненный день недели для подписки ПО
func (r *stateRepository) GetSoftwareSubscriptionWeekday(userID uuid.UUID) (int, error) {
	op := "repository.state.getSoftwareSubscriptionWeekday"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"user_id": userID})
	log.Debug("Getting software subscription weekday")

	state, err := r.GetState(userID, "software_subscription_weekday")
	if err != nil {
		log.Error(err, "Failed to get software subscription weekday")
		return -1, err
	}

	if state == nil {
		log.Debug("Software subscription weekday not found")
		return -1, nil // Не найдено
	}

	weekday, ok := state.StateValue["weekday"].(float64)
	if !ok {
		log.Debug("Weekday value not found in state")
		return -1, nil
	}

	log.WithFields(logger.Fields{"weekday": int(weekday)}).Debug("Software subscription weekday retrieved")
	return int(weekday), nil
}

// SaveSoftwareSubscriptionWeekday сохраняет день недели для подписки ПО
func (r *stateRepository) SaveSoftwareSubscriptionWeekday(userID uuid.UUID, weekday int) error {
	op := "repository.state.saveSoftwareSubscriptionWeekday"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"user_id": userID,
		"weekday": weekday,
	})
	log.Info("Saving software subscription weekday")

	stateValue := models.JSONB{
		"weekday": weekday,
	}
	if err := r.SaveState(userID, "software_subscription_weekday", stateValue); err != nil {
		log.Error(err, "Failed to save software subscription weekday")
		return err
	}

	log.Success("Software subscription weekday saved successfully")
	return nil
}

func (r *stateRepository) GetMobileBaseAmount(userID uuid.UUID) (float64, error) {
	op := "repository.state.getMobileBaseAmount"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"user_id": userID})
	log.Debug("Getting mobile base amount")

	var state models.GenerationState
	if err := r.db.
		Model(&models.GenerationState{}).
		Where("user_id = ?", userID).
		Where("state_key = ?", "mobile_base_amount").
		First(&state).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Debug("Mobile base amount not found")
			return 0, nil
		}
		log.Error(err, "Failed to get mobile base amount")
		return 0, err
	}

	baseAmount, ok := state.StateValue["base_amount"].(float64)
	if !ok {
		log.Debug("Mobile base amount value not found in state")
		return 0, nil
	}
	
	log.WithFields(logger.Fields{"amount": baseAmount}).Debug("Mobile base amount retrieved")
	return baseAmount, nil
}

func (r *stateRepository) SaveMobileBaseAmount(userID uuid.UUID, amount float64, firstMonth string) error {
	op := "repository.state.saveMobileBaseAmount"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"user_id":     userID,
		"amount":      amount,
		"first_month": firstMonth,
	})
	log.Info("Saving mobile base amount")

	stateValue := models.JSONB{
		"base_amount": amount,
		"first_month": firstMonth,
	}
	if err := r.SaveState(userID, "mobile_base_amount", stateValue); err != nil {
		log.Error(err, "Failed to save mobile base amount")
		return err
	}

	log.Success("Mobile base amount saved successfully")
	return nil
}

func (r *stateRepository) GetUtilitiesBaseAmount(userID uuid.UUID) (float64, error) {
	op := "repository.state.getUtilitiesBaseAmount"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"user_id": userID})
	log.Debug("Getting utilities base amount")

	var state models.GenerationState
	if err := r.db.
		Model(&models.GenerationState{}).
		Where("user_id = ?", userID).
		Where("state_key = ?", "utilities_base_amount").
		First(&state).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Debug("Utilities base amount not found")
			return 0, nil
		}
		log.Error(err, "Failed to get utilities base amount")
		return 0, err
	}

	baseAmount, ok := state.StateValue["base_amount"].(float64)
	if !ok {
		log.Debug("Utilities base amount value not found in state")
		return 0, nil
	}
	
	log.WithFields(logger.Fields{"amount": baseAmount}).Debug("Utilities base amount retrieved")
	return baseAmount, nil
}

func (r *stateRepository) SaveUtilitiesBaseAmount(userID uuid.UUID, amount float64, firstMonth string) error {
	op := "repository.state.saveUtilitiesBaseAmount"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"user_id":     userID,
		"amount":      amount,
		"first_month": firstMonth,
	})
	log.Info("Saving utilities base amount")

	stateValue := models.JSONB{
		"base_amount": amount,
		"first_month": firstMonth,
	}
	if err := r.SaveState(userID, "utilities_base_amount", stateValue); err != nil {
		log.Error(err, "Failed to save utilities base amount")
		return err
	}

	log.Success("Utilities base amount saved successfully")
	return nil
}

func (r *stateRepository) GetLeasingBaseAmount(userID uuid.UUID) (float64, error) {
	op := "repository.state.getLeasingBaseAmount"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"user_id": userID})
	log.Debug("Getting leasing base amount")

	var state models.GenerationState
	if err := r.db.
		Model(&models.GenerationState{}).
		Where("user_id = ?", userID).
		Where("state_key = ?", "leasing_base_amount").
		First(&state).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Debug("Leasing base amount not found")
			return 0, nil
		}
		log.Error(err, "Failed to get leasing base amount")
		return 0, err
	}

	baseAmount, ok := state.StateValue["base_amount"].(float64)
	if !ok {
		log.Debug("Leasing base amount value not found in state")
		return 0, nil
	}
	
	log.WithFields(logger.Fields{"amount": baseAmount}).Debug("Leasing base amount retrieved")
	return baseAmount, nil
}

func (r *stateRepository) SaveLeasingBaseAmount(userID uuid.UUID, amount float64, firstMonth string, turnover float64) error {
	op := "repository.state.saveLeasingBaseAmount"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"user_id":     userID,
		"amount":      amount,
		"first_month": firstMonth,
		"turnover":    turnover,
	})
	log.Info("Saving leasing base amount")

	stateValue := models.JSONB{
		"base_amount":          amount,
		"first_month":          firstMonth,
		"first_month_turnover": turnover,
	}
	if err := r.SaveState(userID, "leasing_base_amount", stateValue); err != nil {
		log.Error(err, "Failed to save leasing base amount")
		return err
	}

	log.Success("Leasing base amount saved successfully")
	return nil
}

func (r *stateRepository) GetBaseAmount(userID uuid.UUID) (*dto.BaseAmountsResponse, error) {
	op := "repository.state.getBaseAmount"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"user_id": userID})
	log.Info("Getting base amounts")

	response := &dto.BaseAmountsResponse{
		UserID: userID.String(),
		Code:   200,
	}

	// Получаем мобильную связь
	mobileState, err := r.GetState(userID, "mobile_base_amount")
	if err != nil {
		log.Error(err, "Failed to get mobile base amount")
		return nil, fmt.Errorf("failed to get mobile base amount: %w", err)
	}
	if mobileState != nil {
		baseAmount, _ := mobileState.StateValue["base_amount"].(float64)
		firstMonth, _ := mobileState.StateValue["first_month"].(string)
		response.MobileBaseAmount = &dto.BaseAmountInfo{
			Amount:     baseAmount,
			FirstMonth: firstMonth,
			CreatedAt:  mobileState.CreatedAt.Format(time.RFC3339),
			UpdatedAt:  mobileState.UpdatedAt.Format(time.RFC3339),
		}
	}

	// Получаем коммунальные
	utilitiesState, err := r.GetState(userID, "utilities_base_amount")
	if err != nil {
		return nil, fmt.Errorf("failed to get utilities base amount: %w", err)
	}
	if utilitiesState != nil {
		baseAmount, _ := utilitiesState.StateValue["base_amount"].(float64)
		firstMonth, _ := utilitiesState.StateValue["first_month"].(string)
		response.UtilitiesBaseAmount = &dto.BaseAmountInfo{
			Amount:     baseAmount,
			FirstMonth: firstMonth,
			CreatedAt:  utilitiesState.CreatedAt.Format(time.RFC3339),
			UpdatedAt:  utilitiesState.UpdatedAt.Format(time.RFC3339),
		}
	}

	// Получаем лизинг
	leasingState, err := r.GetState(userID, "leasing_base_amount")
	if err != nil {
		log.Error(err, "Failed to get leasing base amount")
		return nil, fmt.Errorf("failed to get leasing base amount: %w", err)
	}
	if leasingState != nil {
		baseAmount, _ := leasingState.StateValue["base_amount"].(float64)
		firstMonth, _ := leasingState.StateValue["first_month"].(string)
		turnover, _ := leasingState.StateValue["first_month_turnover"].(float64)
		response.LeasingBaseAmount = &dto.BaseAmountInfo{
			Amount:             baseAmount,
			FirstMonth:         firstMonth,
			FirstMonthTurnover: turnover,
			CreatedAt:          leasingState.CreatedAt.Format(time.RFC3339),
			UpdatedAt:          leasingState.UpdatedAt.Format(time.RFC3339),
		}
	}

	log.Success("Base amounts retrieved")
	return response, nil
}

func (r *stateRepository) DeleteMobileBaseAmount(userID uuid.UUID) error {
	op := "repository.state.deleteMobileBaseAmount"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"user_id": userID})
	log.Info("Deleting mobile base amount")

	if err := r.db.
		Where("user_id = ?", userID).
		Where("state_key = ?", "mobile_base_amount").
		Delete(&models.GenerationState{}).Error; err != nil {
		log.Error(err, "Failed to delete mobile base amount")
		return err
	}

	log.Success("Mobile base amount deleted successfully")
	return nil
}

func (r *stateRepository) DeleteUtilitiesBaseAmount(userID uuid.UUID) error {
	op := "repository.state.deleteUtilitiesBaseAmount"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"user_id": userID})
	log.Info("Deleting utilities base amount")

	if err := r.db.
		Where("user_id = ?", userID).
		Where("state_key = ?", "utilities_base_amount").
		Delete(&models.GenerationState{}).Error; err != nil {
		log.Error(err, "Failed to delete utilities base amount")
		return err
	}

	log.Success("Utilities base amount deleted successfully")
	return nil
}

func (r *stateRepository) DeleteLeasingBaseAmount(userID uuid.UUID) error {
	op := "repository.state.deleteLeasingBaseAmount"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"user_id": userID})
	log.Info("Deleting leasing base amount")

	if err := r.db.
		Where("user_id = ?", userID).
		Where("state_key = ?", "leasing_base_amount").
		Delete(&models.GenerationState{}).Error; err != nil {
		log.Error(err, "Failed to delete leasing base amount")
		return err
	}

	log.Success("Leasing base amount deleted successfully")
	return nil
}

// GetAssociatedCard получает сохраненный номер карты для пользователя
func (r *stateRepository) GetAssociatedCard(userID uuid.UUID) (string, error) {
	op := "repository.state.getAssociatedCard"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"user_id": userID})
	log.Debug("Getting associated card")

	state, err := r.GetState(userID, "associated_card")
	if err != nil {
		log.Error(err, "Failed to get associated card")
		return "", err
	}

	if state == nil {
		log.Debug("Associated card not found")
		return "", nil // Не найдено
	}

	cardNumber, ok := state.StateValue["card_number"].(string)
	if !ok {
		log.Debug("Card number not found in state")
		return "", nil
	}

	log.Debug("Associated card retrieved")
	return cardNumber, nil
}

// SaveAssociatedCard сохраняет номер карты для пользователя
func (r *stateRepository) SaveAssociatedCard(userID uuid.UUID, cardNumber string) error {
	op := "repository.state.saveAssociatedCard"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"user_id": userID,
		"card":    cardNumber,
	})
	log.Info("Saving associated card")

	stateValue := models.JSONB{
		"card_number": cardNumber,
	}
	if err := r.SaveState(userID, "associated_card", stateValue); err != nil {
		log.Error(err, "Failed to save associated card")
		return err
	}

	log.Success("Associated card saved successfully")
	return nil
}
