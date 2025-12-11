package service

import (
	"errors"
	"math/rand"

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
}

type gatewayService struct {
	gatewayRepository repository.GatewayRepository
	// ДЛЯ СЛУЧАЙНОГО ВЫБОРА ШЛЮЗА из CSV ФАЙЛА
	configRepo repository.ConfigRepository
}

func NewGatewayService(gatewayRepository repository.GatewayRepository, configRepo repository.ConfigRepository) GatewayService {
	return &gatewayService{
		gatewayRepository: gatewayRepository,
		configRepo:        configRepo,
	}
}

func (s *gatewayService) GetB2CGateways(userID uuid.UUID) (*entities.Gateway, error) {
	return s.gatewayRepository.GetB2CGateways(userID)
}

func (s *gatewayService) SaveB2CGateways(userID uuid.UUID, gatewayID string) error {
	var gateway *entities.Gateway

	if gatewayID == "" {
		// Если gatewayID не указан - выбираем случайный из доступных
		// TODO: проверить, что за случайный csv файл ????
		allGateways, err := s.configRepo.GetGateways()
		if err != nil {
			return err
		}
		if len(allGateways) == 0 {
			// Fallback на дефолтный шлюз
			gateway = entities.NewGateway("gw_default", "Stripe Gateway")
		} else {
			gateway = allGateways[rand.Intn(len(allGateways))]
		}
	} else {
		// Ищем шлюз по ID
		allGateways, err := s.configRepo.GetGateways()
		if err != nil {
			return err
		}

		gateway = nil
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

	return s.gatewayRepository.SaveB2CGateways(userID, gateway.ID, gateway.Name)
}

func (s *gatewayService) DeleteB2CGateways(userID uuid.UUID) error {
	return s.gatewayRepository.DeleteB2CGateways(userID)
}
