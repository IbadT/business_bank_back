package repository

import (
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GenerationRequestRepository interface {
	Create(request *models.GenerationRequest) (*models.GenerationRequest, error)
	GetByID(id uuid.UUID) (*models.GenerationRequest, error)
	GetByUserID(userID uuid.UUID) ([]*models.GenerationRequest, error)
	GetCompletedByUserID(userID uuid.UUID) ([]*models.GenerationRequest, error)
	UpdateStatus(id uuid.UUID, status string, errorMessage *string) error
	UpdateCompletedAt(id uuid.UUID, completedAt time.Time) error
}

type generationRequestRepository struct {
	db *gorm.DB
}

func NewGenerationRequestRepository(db *gorm.DB) GenerationRequestRepository {
	return &generationRequestRepository{
		db: db,
	}
}

func (r *generationRequestRepository) Create(request *models.GenerationRequest) (*models.GenerationRequest, error) {
	if err := r.db.Create(request).Error; err != nil {
		return nil, err
	}
	return request, nil
}

func (r *generationRequestRepository) GetByID(id uuid.UUID) (*models.GenerationRequest, error) {
	var request models.GenerationRequest
	if err := r.db.Where("id = ?", id).First(&request).Error; err != nil {
		return nil, err
	}
	return &request, nil
}

// GetByUserID получает все запросы генерации для пользователя
func (r *generationRequestRepository) GetByUserID(userID uuid.UUID) ([]*models.GenerationRequest, error) {
	var requests []models.GenerationRequest
	if err := r.db.Where("user_id = ?", userID).
		Order("created_at ASC").
		Find(&requests).Error; err != nil {
		return nil, err
	}
	
	result := make([]*models.GenerationRequest, len(requests))
	for i := range requests {
		result[i] = &requests[i]
	}
	return result, nil
}

// GetCompletedByUserID получает только завершенные запросы генерации для пользователя
func (r *generationRequestRepository) GetCompletedByUserID(userID uuid.UUID) ([]*models.GenerationRequest, error) {
	var requests []models.GenerationRequest
	if err := r.db.Where("user_id = ? AND status = ?", userID, "completed").
		Order("created_at ASC").
		Find(&requests).Error; err != nil {
		return nil, err
	}
	
	result := make([]*models.GenerationRequest, len(requests))
	for i := range requests {
		result[i] = &requests[i]
	}
	return result, nil
}

func (r *generationRequestRepository) UpdateStatus(id uuid.UUID, status string, errorMessage *string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}
	if errorMessage != nil {
		updates["error_message"] = *errorMessage
	}
	return r.db.Model(&models.GenerationRequest{}).Where("id = ?", id).Updates(updates).Error
}

func (r *generationRequestRepository) UpdateCompletedAt(id uuid.UUID, completedAt time.Time) error {
	return r.db.Model(&models.GenerationRequest{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"completed_at": completedAt,
			"status":       "completed",
			"updated_at":   time.Now(),
		}).Error
}
