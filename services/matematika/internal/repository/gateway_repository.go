package repository

import (
	"github.com/IbadT/business_bank_back/services/matematika/internal/domain/entities"
	"github.com/IbadT/business_bank_back/services/matematika/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GatewayRepository interface {
	GetB2CGateways(userID uuid.UUID) (*entities.Gateway, error)
	SaveB2CGateways(userID uuid.UUID, gatewayID string, gatewayName string) error
	DeleteB2CGateways(userID uuid.UUID) error
}

type gatewayRepository struct {
	DB *gorm.DB
}

func NewGatewayRepository(db *gorm.DB) GatewayRepository {
	return &gatewayRepository{
		DB: db,
	}
}

func (r *gatewayRepository) GetB2CGateways(userID uuid.UUID) (*entities.Gateway, error) {
	var userGateway models.UserGateway
	if err := r.DB.Where("user_id = ?", userID).First(&userGateway).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Шлюз не найден - это нормально
		}
		return nil, err
	}

	return entities.NewGateway(userGateway.GatewayID, userGateway.GatewayName), nil
}

func (r *gatewayRepository) SaveB2CGateways(userID uuid.UUID, gatewayID string, gatewayName string) error {
	userGateway := models.UserGateway{
		UserID:      userID,
		GatewayID:   gatewayID,
		GatewayName: gatewayName,
	}

	// Используем FirstOrCreate с обновлением если запись существует
	var existing models.UserGateway
	if err := r.DB.Where("user_id = ?", userID).First(&existing).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Создаем новую запись
			return r.DB.Create(&userGateway).Error
		}
		return err
	}

	// Обновляем существующую запись
	existing.GatewayID = gatewayID
	existing.GatewayName = gatewayName
	return r.DB.Save(&existing).Error
}

func (r *gatewayRepository) DeleteB2CGateways(userID uuid.UUID) error {
	return r.DB.Where("user_id = ?", userID).Delete(&models.UserGateway{}).Error
}
