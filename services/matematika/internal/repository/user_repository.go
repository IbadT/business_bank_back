package repository

import (
	"errors"
	"fmt"

	"github.com/IbadT/business_bank_back/services/matematika/internal/domain"
	"github.com/IbadT/business_bank_back/services/matematika/internal/models"
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
	// Устанавливаем роль: если указана валидная роль - используем её, иначе "user"
	role := models.RoleUser
	if user.Role != "" && models.IsValidRole(user.Role) {
		role = user.Role
	}
	
	// Конвертируем domain.User в models.User для работы с GORM
	dbUser := models.User{
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		Role:         role, // Всегда явно устанавливаем валидную роль
	}
	
	// Дополнительная проверка перед созданием - гарантируем валидную роль
	if !models.IsValidRole(dbUser.Role) {
		dbUser.Role = models.RoleUser
	}

	fmt.Println("DBUSER: ", dbUser)
	
	// Создаем пользователя в БД
	if err := r.db.Create(&dbUser).Error; err != nil {
		return err
	}
	
	// Обновляем domain.User с данными из БД (ID, timestamps, role)
	user.ID = dbUser.ID
	user.Role = dbUser.Role
	user.CreatedAt = dbUser.CreatedAt
	user.UpdatedAt = dbUser.UpdatedAt
	
	return nil
}