package gatewayservice

import (
	"context"
	"math/rand"

	"github.com/IbadT/business_bank_back/services/matematika/internal/cache"
	"github.com/IbadT/business_bank_back/services/matematika/internal/domain/entities"
	"github.com/IbadT/business_bank_back/services/matematika/internal/repository"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/helpers"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/logger"
	"github.com/google/uuid"
)

type GatewayService interface {
	GetB2CGateways(userID uuid.UUID) (*entities.Gateway, error)
	SaveB2CGateways(userID uuid.UUID, gatewayID string) error
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

type gatewayService struct {
	gatewayRepository repository.GatewayRepository
	// ДЛЯ СЛУЧАЙНОГО ВЫБОРА ШЛЮЗА из CSV ФАЙЛА
	configRepo repository.ConfigRepository
	cache      *cache.CacheService
}

func NewGatewayService(gatewayRepository repository.GatewayRepository, configRepo repository.ConfigRepository, cache *cache.CacheService) GatewayService {
	return &gatewayService{
		gatewayRepository: gatewayRepository,
		configRepo:        configRepo,
		cache:             cache,
	}
}

func (s *gatewayService) GetB2CGateways(userID uuid.UUID) (*entities.Gateway, error) {
	op := "service.gateway.getB2CGateways"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"user_id": userID})
	log.Info("Getting B2C gateways")

	gateway, err := s.gatewayRepository.GetB2CGateways(userID)
	if err != nil {
		log.Error(err, "Failed to get B2C gateways from repository")
		return nil, err
	}

	if gateway == nil {
		log.Info("B2C gateway not found for user")
		return nil, nil
	}

	log.WithFields(logger.Fields{
		"gateway_id":   gateway.ID,
		"gateway_name": gateway.Name,
	}).Success("B2C gateway retrieved")
	return gateway, nil
}

func (s *gatewayService) SaveB2CGateways(userID uuid.UUID, gatewayID string) error {
	op := "service.gateway.saveB2CGateways"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"user_id":    userID,
		"gateway_id": gatewayID,
	})
	log.Info("Saving B2C gateway")

	var gateway *entities.Gateway

	ctxBg := context.Background()

	if gatewayID == "" {
		log.Info("Gateway ID not provided, selecting random gateway")
		// Если gatewayID не указан - выбираем случайный из доступных
		gatewaysFromCache, err := s.cache.GetGateways(ctxBg)
		if err == nil && len(gatewaysFromCache) > 0 {
			// Конвертируем domain.Gateway в entities.Gateway
			idx := rand.Intn(len(gatewaysFromCache))
			gateway = entities.NewGateway(gatewaysFromCache[idx].ID, gatewaysFromCache[idx].Name)
		} else {
			// Если кеш пуст или ошибка - берем из CSV
			allGateways, err := s.configRepo.GetGateways()
			if err != nil {
				log.Error(err, "Failed to get gateways from config")
				return err
			}
			if len(allGateways) == 0 {
				log.Warn("No gateways found in config")
				return helpers.ErrGatewayNotFound
			}
			idx := rand.Intn(len(allGateways))
			gateway = allGateways[idx]
			log.WithFields(logger.Fields{"selected_gateway": gateway.ID}).Debug("Random gateway selected from config")
		}
	} else {
		log.Info("Gateway ID provided, searching for gateway")
		// Ищем шлюз по ID
		gatewaysFromCache, err := s.cache.GetGateways(ctxBg)
		if err == nil && len(gatewaysFromCache) > 0 {
			// Ищем в кеше по ID
			for _, g := range gatewaysFromCache {
				if g.ID == gatewayID {
					gateway = entities.NewGateway(g.ID, g.Name)
					break
				}
			}
		}

		// Если не нашли в кеше или кеш пуст - ищем в CSV
		if gateway == nil {
			allGateways, err := s.configRepo.GetGateways()
			if err != nil {
				log.Error(err, "Failed to get gateways from config")
				return err
			}

			for _, g := range allGateways {
				if g.ID == gatewayID {
					gateway = g
					break
				}
			}

			if gateway == nil {
				log.Warn("Gateway not found: %s", gatewayID)
				return helpers.ErrGatewayNotFound
			}
			log.Debug("Gateway found in config: %s", gatewayID)
		}
	}

	if err := s.gatewayRepository.SaveB2CGateways(userID, gateway.ID, gateway.Name); err != nil {
		log.Error(err, "Failed to save B2C gateway in repository")
		return err
	}

	log.Success("B2C gateway saved successfully")
	return nil
}

func (s *gatewayService) DeleteB2CGateways(userID uuid.UUID) error {
	op := "service.gateway.deleteB2CGateways"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"user_id": userID})
	log.Info("Deleting B2C gateway")

	if err := s.gatewayRepository.DeleteB2CGateways(userID); err != nil {
		log.Error(err, "Failed to delete B2C gateway from repository")
		return err
	}

	log.Success("B2C gateway deleted successfully")
	return nil
}


// ================================ ADMIN GATEWAY SERVICES ================================
func (s *gatewayService) GetAdminGateways() error {
	op := "service.gateway.getAdminGateways"
	log := logger.GetLogger().WithOperation(op)
	log.Info("Getting admin gateways (not implemented)")
	return nil
}
func (s *gatewayService) GetAdminUsersGateways() error {
	op := "service.gateway.getAdminUsersGateways"
	log := logger.GetLogger().WithOperation(op)
	log.Info("Getting admin users gateways (not implemented)")
	return nil
}
func (s *gatewayService) GetAdminUserGateway() error {
	op := "service.gateway.getAdminUserGateway"
	log := logger.GetLogger().WithOperation(op)
	log.Info("Getting admin user gateway (not implemented)")
	return nil
}
func (s *gatewayService) CreateAdminGateway() error {
	op := "service.gateway.createAdminGateway"
	log := logger.GetLogger().WithOperation(op)
	log.Info("Creating admin gateway (not implemented)")
	return nil
}
func (s *gatewayService) UpdateAdminGateway() error {
	op := "service.gateway.updateAdminGateway"
	log := logger.GetLogger().WithOperation(op)
	log.Info("Updating admin gateway (not implemented)")
	return nil
}
func (s *gatewayService) UpdateAdminUserGateway() error {
	op := "service.gateway.updateAdminUserGateway"
	log := logger.GetLogger().WithOperation(op)
	log.Info("Updating admin user gateway (not implemented)")
	return nil
}
func (s *gatewayService) DeleteAdminGateway() error {
	op := "service.gateway.deleteAdminGateway"
	log := logger.GetLogger().WithOperation(op)
	log.Info("Deleting admin gateway (not implemented)")
	return nil
}
func (s *gatewayService) DeleteAdminUserGateway() error {
	op := "service.gateway.deleteAdminUserGateway"
	log := logger.GetLogger().WithOperation(op)
	log.Info("Deleting admin user gateway (not implemented)")
	return nil
}