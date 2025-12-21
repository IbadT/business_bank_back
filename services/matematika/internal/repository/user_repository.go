package repository

import (
	"errors"
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/internal/domain"
	"github.com/IbadT/business_bank_back/services/matematika/internal/models"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/helpers"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/logger"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	GetByEmail(email string) (*domain.User, error)
	Create(user *domain.User) error
	GetByID(id uuid.UUID) (*models.User, error)
	UpdateAssociatedCard(user *domain.User) error
}

type userRepository struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		DB: db,
	}
}

func (r *userRepository) GetByEmail(email string) (*domain.User, error) {
	op := "repository.user.getByEmail"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"email": email})
	log.Info("Getting user by email")
	
	var user domain.User
	if err := r.DB.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Warn("User not found")
			return nil, helpers.ErrUserNotFound
		}
		log.Error(err, "Failed to get user by email")
		return nil, err
	}
	
	log.Success("User retrieved by email")
	return &user, nil
}

func (r *userRepository) Create(user *domain.User) error {
	op := "repository.user.create"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"email": user.Email,
		"user_id": user.ID,
	})
	log.Info("Creating user")

	// Создаем пользователя в БД
	if err := r.DB.Create(user).Error; err != nil {
		log.Error(err, "Failed to create user")
		return err
	}

	log.Success("User created successfully")
	return nil
}

func (r *userRepository) GetByID(id uuid.UUID) (*models.User, error) {
	op := "repository.user.getByID"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"user_id": id})
	log.Info("Getting user by ID")
	
	var user models.User
	if err := r.DB.Model(&models.User{}).Where("id = ?", id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Warn("User not found")
			return nil, helpers.ErrUserNotFound
		}
		log.Error(err, "Failed to get user by ID")
		return nil, err
	}
	
	log.Success("User retrieved by ID")
	return &user, nil
}

func (r *userRepository) UpdateAssociatedCard(user *domain.User) error {
	op := "repository.user.updateAssociatedCard"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"user_id": user.ID,
		"card":    user.AssociatedCard,
	})
	log.Info("Updating associated card")

	// Обновляем только поле associated_card в таблице users
	updateData := map[string]interface{}{
		"associated_card": user.AssociatedCard,
		"updated_at":      time.Now(),
	}

	result := r.DB.Model(&models.User{}).
		Where("id = ?", user.ID).
		Updates(updateData)

	if result.Error != nil {
		log.Error(result.Error, "Failed to update associated card")
		return result.Error
	}

	if result.RowsAffected == 0 {
		log.Warn("User not found or no changes made")
		return helpers.ErrUserNotFoundOrNoChanges
	}

	log.Success("Associated card updated successfully")
	return nil
}
