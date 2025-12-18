// internal/repository/state_repository.go
package repository

import (
	"fmt"
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/internal/models"
	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/dto"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
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
	var state models.GenerationState
	query := r.db.Where("state_key = ?", stateKey)

	if userID != uuid.Nil {
		query = query.Where("user_id = ?", userID)
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
func (r *stateRepository) SaveState(userID uuid.UUID, stateKey string, stateValue models.JSONB) error {
	logrus.Debugf("[DEBUG] SaveState called: userID=%v, stateKey=%s, stateValue=%+v", userID, stateKey, stateValue)

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
		logrus.Debugf("[DEBUG] Searching for state with userID=%v", userID)
	} else {
		query = query.Where("user_id IS NULL")
		logrus.Debugf("[DEBUG] Searching for state with userID IS NULL")
	}

	if err := query.First(&existing).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Создаем новую запись
			logrus.Debugf("[DEBUG] State not found, creating new: stateKey=%s, userID=%v", stateKey, userID)
			if err := r.db.Create(&state).Error; err != nil {
				logrus.Debugf("[ERROR] Failed to create state: %v", err)
				return err
			}
			logrus.Debugf("[DEBUG] Successfully created state")
			return nil
		}
		logrus.Debugf("[ERROR] Failed to query state: %v", err)
		return err
	}

	// Обновляем существующую запись
	logrus.Debugf("[DEBUG] State found, updating: id=%v", existing.ID)
	existing.StateValue = stateValue
	if err := r.db.Save(&existing).Error; err != nil {
		logrus.Debugf("[ERROR] Failed to update state: %v", err)
		return err
	}
	logrus.Debugf("[DEBUG] Successfully updated state")
	return nil
}

// GetSoftwareSubscriptionWeekday получает сохраненный день недели для подписки ПО
func (r *stateRepository) GetSoftwareSubscriptionWeekday(userID uuid.UUID) (int, error) {
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
func (r *stateRepository) SaveSoftwareSubscriptionWeekday(userID uuid.UUID, weekday int) error {
	stateValue := models.JSONB{
		"weekday": weekday,
	}
	return r.SaveState(userID, "software_subscription_weekday", stateValue)
}

func (r *stateRepository) GetMobileBaseAmount(userID uuid.UUID) (float64, error) {
	var state models.GenerationState
	if err := r.db.
		Model(&models.GenerationState{}).
		Where("user_id = ?", userID).
		Where("state_key = ?", "mobile_base_amount").
		First(&state).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, err
	}

	baseAmount, ok := state.StateValue["base_amount"].(float64)
	if !ok {
		return 0, nil
	}
	return baseAmount, nil
}

func (r *stateRepository) SaveMobileBaseAmount(userID uuid.UUID, amount float64, firstMonth string) error {
	stateValue := models.JSONB{
		"base_amount": amount,
		"first_month": firstMonth,
	}
	return r.SaveState(userID, "mobile_base_amount", stateValue)
}

func (r *stateRepository) GetUtilitiesBaseAmount(userID uuid.UUID) (float64, error) {
	var state models.GenerationState
	if err := r.db.
		Model(&models.GenerationState{}).
		Where("user_id = ?", userID).
		Where("state_key = ?", "utilities_base_amount").
		First(&state).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, err
	}

	baseAmount, ok := state.StateValue["base_amount"].(float64)
	if !ok {
		return 0, nil
	}
	return baseAmount, nil
}

func (r *stateRepository) SaveUtilitiesBaseAmount(userID uuid.UUID, amount float64, firstMonth string) error {
	stateValue := models.JSONB{
		"base_amount": amount,
		"first_month": firstMonth,
	}
	return r.SaveState(userID, "utilities_base_amount", stateValue)
}

func (r *stateRepository) GetLeasingBaseAmount(userID uuid.UUID) (float64, error) {
	var state models.GenerationState
	if err := r.db.
		Model(&models.GenerationState{}).
		Where("user_id = ?", userID).
		Where("state_key = ?", "leasing_base_amount").
		First(&state).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, err
	}

	baseAmount, ok := state.StateValue["base_amount"].(float64)
	if !ok {
		return 0, nil
	}
	return baseAmount, nil
}

func (r *stateRepository) SaveLeasingBaseAmount(userID uuid.UUID, amount float64, firstMonth string, turnover float64) error {
	stateValue := models.JSONB{
		"base_amount":          amount,
		"first_month":          firstMonth,
		"first_month_turnover": turnover,
	}
	return r.SaveState(userID, "leasing_base_amount", stateValue)
}

func (r *stateRepository) GetBaseAmount(userID uuid.UUID) (*dto.BaseAmountsResponse, error) {
	response := &dto.BaseAmountsResponse{
		UserID: userID.String(),
		Code:   200,
	}

	// Получаем мобильную связь
	mobileState, err := r.GetState(userID, "mobile_base_amount")
	if err != nil {
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

	return response, nil
}

func (r *stateRepository) DeleteMobileBaseAmount(userID uuid.UUID) error {
	return r.db.
		Where("user_id = ?", userID).
		Where("state_key = ?", "mobile_base_amount").
		Delete(&models.GenerationState{}).Error
}

func (r *stateRepository) DeleteUtilitiesBaseAmount(userID uuid.UUID) error {
	return r.db.
		Where("user_id = ?", userID).
		Where("state_key = ?", "utilities_base_amount").
		Delete(&models.GenerationState{}).Error
}

func (r *stateRepository) DeleteLeasingBaseAmount(userID uuid.UUID) error {
	return r.db.
		Where("user_id = ?", userID).
		Where("state_key = ?", "leasing_base_amount").
		Delete(&models.GenerationState{}).Error
}

// GetAssociatedCard получает сохраненный номер карты для пользователя
func (r *stateRepository) GetAssociatedCard(userID uuid.UUID) (string, error) {
	state, err := r.GetState(userID, "associated_card")
	if err != nil {
		return "", err
	}

	if state == nil {
		return "", nil // Не найдено
	}

	cardNumber, ok := state.StateValue["card_number"].(string)
	if !ok {
		return "", nil
	}

	return cardNumber, nil
}

// SaveAssociatedCard сохраняет номер карты для пользователя
func (r *stateRepository) SaveAssociatedCard(userID uuid.UUID, cardNumber string) error {
	stateValue := models.JSONB{
		"card_number": cardNumber,
	}
	return r.SaveState(userID, "associated_card", stateValue)
}
