package repository

import (
	"errors"

	"github.com/IbadT/business_bank_back/services/matematika/internal/domain/entities"
	"github.com/IbadT/business_bank_back/services/matematika/internal/models"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/logger"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GatewayRepository interface {
	GetB2CGateways(userID uuid.UUID) (*entities.Gateway, error)
	SaveB2CGateways(userID uuid.UUID, gatewayID string, gatewayName string) error
	DeleteB2CGateways(userID uuid.UUID) error

	GetAdminGateways() error
	GetAdminUsersGateways() error
	GetAdminUserGateway() error
	CreateAdminGateway() error
	UpdateAdminGateway() error
	UpdateAdminUserGateway() error
	DeleteAdminGateway() error
	DeleteAdminUserGateway() error
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
	op := "repository.gateway.getB2CGateways"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"user_id": userID})
	log.Info("Getting B2C gateways")

	var userGateway models.UserGateway
	if err := r.DB.Where("user_id = ?", userID).First(&userGateway).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Info("B2C gateway not found for user")
			return nil, nil // Шлюз не найден - это нормально
		}
		log.Error(err, "Failed to get B2C gateways")
		return nil, err
	}

	log.WithFields(logger.Fields{
		"gateway_id":   userGateway.GatewayID,
		"gateway_name": userGateway.GatewayName,
	}).Success("B2C gateway retrieved")

	return entities.NewGateway(userGateway.GatewayID, userGateway.GatewayName), nil
}

func (r *gatewayRepository) SaveB2CGateways(userID uuid.UUID, gatewayID string, gatewayName string) error {
	op := "repository.gateway.saveB2CGateways"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"user_id":     userID,
		"gateway_id":  gatewayID,
		"gateway_name": gatewayName,
	})
	log.Info("Saving B2C gateway")

	userGateway := models.UserGateway{
		UserID:      userID,
		GatewayID:   gatewayID,
		GatewayName: gatewayName,
	}

	// Используем FirstOrCreate с обновлением если запись существует
	var existing models.UserGateway
	if err := r.DB.Where("user_id = ?", userID).First(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Создаем новую запись
			if err := r.DB.Create(&userGateway).Error; err != nil {
				log.Error(err, "Failed to create B2C gateway")
				return err
			}
			log.Success("B2C gateway created successfully")
			return nil
		}
		log.Error(err, "Failed to check existing B2C gateway")
		return err
	}

	// Обновляем существующую запись
	existing.GatewayID = gatewayID
	existing.GatewayName = gatewayName
	if err := r.DB.Save(&existing).Error; err != nil {
		log.Error(err, "Failed to update B2C gateway")
		return err
	}

	log.Success("B2C gateway updated successfully")
	return nil
}

func (r *gatewayRepository) DeleteB2CGateways(userID uuid.UUID) error {
	op := "repository.gateway.deleteB2CGateways"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"user_id": userID})
	log.Info("Deleting B2C gateway")

	if err := r.DB.Where("user_id = ?", userID).Delete(&models.UserGateway{}).Error; err != nil {
		log.Error(err, "Failed to delete B2C gateway")
		return err
	}

	log.Success("B2C gateway deleted successfully")
	return nil
}





// ================================ ADMIN GATEWAY REPOSITORY ================================
func (r *gatewayRepository) GetAdminGateways() error {
	return nil
}
func (r *gatewayRepository) GetAdminUsersGateways() error {
	return nil
}
func (r *gatewayRepository) GetAdminUserGateway() error {
	return nil
}
func (r *gatewayRepository) CreateAdminGateway() error {
	return nil
}
func (r *gatewayRepository) UpdateAdminGateway() error {
	return nil
}
func (r *gatewayRepository) UpdateAdminUserGateway() error {
	return nil
}
func (r *gatewayRepository) DeleteAdminGateway() error {
		return nil
}
func (r *gatewayRepository) DeleteAdminUserGateway() error {
	return nil
}