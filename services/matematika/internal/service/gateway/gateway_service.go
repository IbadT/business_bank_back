package gatewayservice

import (
	"context"
	"errors"
	"math/rand"

	"github.com/IbadT/business_bank_back/services/matematika/internal/cache"
	"github.com/IbadT/business_bank_back/services/matematika/internal/domain/entities"
	"github.com/IbadT/business_bank_back/services/matematika/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrGatewayNotFound = errors.New("gateway not found")
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
	return s.gatewayRepository.GetB2CGateways(userID)
}

func (s *gatewayService) SaveB2CGateways(userID uuid.UUID, gatewayID string) error {
	var gateway *entities.Gateway

	ctxBg := context.Background()

	if gatewayID == "" {
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
				return err
			}
			if len(allGateways) == 0 {
				return ErrGatewayNotFound
			}
			idx := rand.Intn(len(allGateways))
			gateway = allGateways[idx]
		}
	} else {
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
				return err
			}

			for _, g := range allGateways {
				if g.ID == gatewayID {
					gateway = g
					break
				}
			}

			if gateway == nil {
				return ErrGatewayNotFound
			}
		}
	}

	return s.gatewayRepository.SaveB2CGateways(userID, gateway.ID, gateway.Name)
}

func (s *gatewayService) DeleteB2CGateways(userID uuid.UUID) error {
	return s.gatewayRepository.DeleteB2CGateways(userID)
}


// ================================ ADMIN GATEWAY SERVICES ================================
func (s *gatewayService) GetAdminGateways () error {
	return nil
}
func (s *gatewayService) GetAdminUsersGateways () error {
	return nil
}
func (s *gatewayService) GetAdminUserGateway () error {
	return nil
}
func (s *gatewayService) CreateAdminGateway () error {
	return nil
}
func (s *gatewayService) UpdateAdminGateway () error {
	return nil
}
func (s *gatewayService) UpdateAdminUserGateway () error {
	return nil
}
func (s *gatewayService) DeleteAdminGateway () error {
	return nil
}
func (s *gatewayService) DeleteAdminUserGateway () error {
	return nil
}