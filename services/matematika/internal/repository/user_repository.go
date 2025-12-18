package repository

import (
	"errors"
	"fmt"
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/internal/domain"
	"github.com/IbadT/business_bank_back/services/matematika/internal/models"
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
	var user domain.User
	if err := r.DB.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	fmt.Println("USER GET BY EMAIL: ", user)
	return &user, nil
}

func (r *userRepository) Create(user *domain.User) error {
	fmt.Println("USER CREATE: ", user)

	// Создаем пользователя в БД
	if err := r.DB.Create(user).Error; err != nil {
		return err
	}

	return nil
}

func (r *userRepository) GetByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.DB.Model(&models.User{}).Where("id = ?", id).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) UpdateAssociatedCard(user *domain.User) error {
	// Обновляем только поле associated_card в таблице users
	updateData := map[string]interface{}{
		"associated_card": user.AssociatedCard,
		"updated_at":      time.Now(),
	}

	result := r.DB.Model(&models.User{}).
		Where("id = ?", user.ID).
		Updates(updateData)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("user not found or no changes made")
	}

	return nil
}
