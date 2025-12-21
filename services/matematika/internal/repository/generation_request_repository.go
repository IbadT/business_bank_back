package repository

import (
	"errors"
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/internal/models"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/logger"
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
	op := "repository.generationRequest.create"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"request_id": request.ID,
		"user_id":    request.UserID,
		"status":     request.Status,
	})
	log.Info("Creating generation request")

	if err := r.db.Create(request).Error; err != nil {
		log.Error(err, "Failed to create generation request")
		return nil, err
	}
	
	log.Success("Generation request created successfully")
	return request, nil
}

func (r *generationRequestRepository) GetByID(id uuid.UUID) (*models.GenerationRequest, error) {
	op := "repository.generationRequest.getByID"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"request_id": id})
	log.Info("Getting generation request by ID")

	var request models.GenerationRequest
	if err := r.db.Where("id = ?", id).First(&request).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Warn("Generation request not found")
		} else {
			log.Error(err, "Failed to get generation request by ID")
		}
		return nil, err
	}
	
	log.Success("Generation request retrieved by ID")
	return &request, nil
}

// GetByUserID получает все запросы генерации для пользователя
func (r *generationRequestRepository) GetByUserID(userID uuid.UUID) ([]*models.GenerationRequest, error) {
	op := "repository.generationRequest.getByUserID"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"user_id": userID})
	log.Info("Getting generation requests by user ID")

	var requests []models.GenerationRequest
	if err := r.db.Where("user_id = ?", userID).
		Order("created_at ASC").
		Find(&requests).Error; err != nil {
		log.Error(err, "Failed to get generation requests by user ID")
		return nil, err
	}
	
	result := make([]*models.GenerationRequest, len(requests))
	for i := range requests {
		result[i] = &requests[i]
	}
	
	log.WithFields(logger.Fields{"count": len(result)}).Success("Generation requests retrieved by user ID")
	return result, nil
}

// GetCompletedByUserID получает только завершенные запросы генерации для пользователя
func (r *generationRequestRepository) GetCompletedByUserID(userID uuid.UUID) ([]*models.GenerationRequest, error) {
	op := "repository.generationRequest.getCompletedByUserID"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"user_id": userID})
	log.Info("Getting completed generation requests by user ID")

	var requests []models.GenerationRequest
	if err := r.db.Where("user_id = ? AND status = ?", userID, "completed").
		Order("created_at ASC").
		Find(&requests).Error; err != nil {
		log.Error(err, "Failed to get completed generation requests")
		return nil, err
	}
	
	result := make([]*models.GenerationRequest, len(requests))
	for i := range requests {
		result[i] = &requests[i]
	}
	
	log.WithFields(logger.Fields{"count": len(result)}).Success("Completed generation requests retrieved")
	return result, nil
}

func (r *generationRequestRepository) UpdateStatus(id uuid.UUID, status string, errorMessage *string) error {
	op := "repository.generationRequest.updateStatus"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"request_id": id,
		"status":     status,
	})
	log.Info("Updating generation request status")

	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}
	if errorMessage != nil {
		updates["error_message"] = *errorMessage
		log = log.WithFields(logger.Fields{"error_message": *errorMessage})
	}
	
	if err := r.db.Model(&models.GenerationRequest{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		log.Error(err, "Failed to update generation request status")
		return err
	}

	log.Success("Generation request status updated successfully")
	return nil
}

func (r *generationRequestRepository) UpdateCompletedAt(id uuid.UUID, completedAt time.Time) error {
	op := "repository.generationRequest.updateCompletedAt"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"request_id":  id,
		"completed_at": completedAt,
	})
	log.Info("Updating generation request completed at")

	if err := r.db.Model(&models.GenerationRequest{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"completed_at": completedAt,
			"status":       "completed",
			"updated_at":   time.Now(),
		}).Error; err != nil {
		log.Error(err, "Failed to update generation request completed at")
		return err
	}

	log.Success("Generation request completed at updated successfully")
	return nil
}
