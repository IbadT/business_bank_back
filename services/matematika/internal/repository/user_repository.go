package repository

import (
	"errors"
	"fmt"

	"github.com/IbadT/business_bank_back/services/matematika/internal/domain"
	"gorm.io/gorm"
)

type UserRepository interface {
	GetByEmail(email string) (*domain.User, error)
	Create(user *domain.User) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) GetByEmail(email string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
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
	if err := r.db.Create(user).Error; err != nil {
		return err
	}
	
	return nil
}