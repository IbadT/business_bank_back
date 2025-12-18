package service

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/internal/repository"
	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/dto"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/utils"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type BaseAmountService interface {
	// Получение базовых сумм
	GetMobileBaseAmount(userIDStr string) (float64, error)
	GetUtilitiesBaseAmount(userIDStr string) (float64, error)
	GetLeasingBaseAmount(userIDStr string) (float64, error)
	GetBaseAmount(userIDStr string) (*dto.BaseAmountsResponse, error)

	// Получение первого месяца из БД
	GetMobileFirstMonth(userIDStr string) (string, error)
	GetUtilitiesFirstMonth(userIDStr string) (string, error)
	GetLeasingFirstMonth(userIDStr string) (string, error)

	// Расчет суммы с учетом ±15% для мобильной/коммунальных
	CalculateMobileAmount(userIDStr string, isFirstMonth bool, monthStr string) (float64, error)
	CalculateUtilitiesAmount(userIDStr string, isFirstMonth bool, monthStr string) (float64, error)

	// Расчет суммы лизинга
	CalculateLeasingAmount(userIDStr string, turnover float64, isFirstMonth bool, monthStr string) (float64, error)

	// Сохранение базовых сумм (вызывается из generator)
	SaveMobileBaseAmount(userIDStr string, amount float64, firstMonth string) error
	SaveUtilitiesBaseAmount(userIDStr string, amount float64, firstMonth string) error
	SaveLeasingBaseAmount(userIDStr string, amount float64, firstMonth string, turnover float64) error

	// Удаление базовых сумм (для тестирования)
	DeleteMobileBaseAmount(userIDStr string) error
	DeleteUtilitiesBaseAmount(userIDStr string) error
	DeleteLeasingBaseAmount(userIDStr string) error
}

type baseAmountService struct {
	stateRepo repository.StateRepository
}

func NewBaseAmountService(stateRepo repository.StateRepository) BaseAmountService {
	return &baseAmountService{
		stateRepo: stateRepo,
	}
}

func (s *baseAmountService) GetMobileBaseAmount(userIDStr string) (float64, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return 0, err
	}
	return s.stateRepo.GetMobileBaseAmount(userID)
}

func (s *baseAmountService) GetUtilitiesBaseAmount(userIDStr string) (float64, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return 0, err
	}
	return s.stateRepo.GetUtilitiesBaseAmount(userID)
}

func (s *baseAmountService) GetLeasingBaseAmount(userIDStr string) (float64, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return 0, err
	}
	return s.stateRepo.GetLeasingBaseAmount(userID)
}

func (s *baseAmountService) GetBaseAmount(userIDStr string) (*dto.BaseAmountsResponse, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, err
	}
	return s.stateRepo.GetBaseAmount(userID)
}

func (s *baseAmountService) GetMobileFirstMonth(userIDStr string) (string, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return "", fmt.Errorf("invalid userID format: %w", err)
	}
	state, err := s.stateRepo.GetState(userID, "mobile_base_amount")
	if err != nil {
		// Реальная ошибка БД/сети - логируем, но возвращаем как "запись не найдена"
		// чтобы не блокировать генерацию. В будущем нужно добавить проверку истории генераций.
		logrus.WithError(err).Warnf("[WARN] GetState error for mobile_base_amount (userID=%s): %v, treating as not found", userIDStr, err)
		return "", nil // Возвращаем как "запись не найдена"
	}
	if state == nil {
		return "", nil
	}
	firstMonth, _ := state.StateValue["first_month"].(string)
	return firstMonth, nil
}

func (s *baseAmountService) GetUtilitiesFirstMonth(userIDStr string) (string, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return "", fmt.Errorf("invalid userID format: %w", err)
	}
	state, err := s.stateRepo.GetState(userID, "utilities_base_amount")
	if err != nil {
		// Реальная ошибка БД/сети - логируем, но возвращаем как "запись не найдена"
		// чтобы не блокировать генерацию. В будущем нужно добавить проверку истории генераций.
		logrus.WithError(err).Warnf("[WARN] GetState error for utilities_base_amount (userID=%s): %v, treating as not found", userIDStr, err)
		return "", nil // Возвращаем как "запись не найдена"
	}
	if state == nil {
		return "", nil
	}
	firstMonth, _ := state.StateValue["first_month"].(string)
	return firstMonth, nil
}

func (s *baseAmountService) GetLeasingFirstMonth(userIDStr string) (string, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return "", fmt.Errorf("invalid userID format: %w", err)
	}
	state, err := s.stateRepo.GetState(userID, "leasing_base_amount")
	if err != nil {
		// Реальная ошибка БД/сети - логируем, но возвращаем как "запись не найдена"
		// чтобы не блокировать генерацию. В будущем нужно добавить проверку истории генераций.
		logrus.WithError(err).Warnf("[WARN] GetState error for leasing_base_amount (userID=%s): %v, treating as not found", userIDStr, err)
		return "", nil // Возвращаем как "запись не найдена"
	}
	if state == nil {
		return "", nil
	}
	firstMonth, _ := state.StateValue["first_month"].(string)
	return firstMonth, nil
}

func (s *baseAmountService) CalculateMobileAmount(userIDStr string, isFirstMonth bool, monthStr string) (float64, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return 0, fmt.Errorf("invalid userID format: %w", err)
	}

	if isFirstMonth {
		// Проверяем, нужно ли сохранять новую базовую сумму
		// Сохраняем только если записи нет или запрашиваемый месяц раньше сохраненного
		savedFirstMonth, _ := s.GetMobileFirstMonth(userIDStr)
		shouldSave := (savedFirstMonth == "" || monthStr < savedFirstMonth)

		if shouldSave {
			// [15][16] Первый месяц: фиксируется в диапазоне $200–500
			// TODO: мне кажется, рандомное значение не должно быть !!!!!!!!!!!
			amount := 200.0 + rand.Float64()*(500.0-200.0)
			amount = utils.RoundToCents(amount)

			// Сохраняем базовую сумму с месяцем из запроса
			if monthStr == "" {
				monthStr = time.Now().Format("2006-01")
			}
			if err := s.stateRepo.SaveMobileBaseAmount(userID, amount, monthStr); err != nil {
				return 0, fmt.Errorf("failed to save mobile base amount: %w", err)
			}

			return amount, nil
		} else {
			// Используем сохраненную базовую сумму (для повторной генерации того же месяца)
			baseAmount, err := s.stateRepo.GetMobileBaseAmount(userID)
			if err != nil {
				return 0, fmt.Errorf("failed to get mobile base amount: %w", err)
			}
			return baseAmount, nil
		}
	}

	// Последующие месяцы: ±15% от базовой суммы
	baseAmount, err := s.stateRepo.GetMobileBaseAmount(userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get mobile base amount: %w", err)
	}

	if baseAmount == 0 {
		return 0, errors.New("mobile base amount not found. Generate first month first")
	}

	// Применяем ±15% вариацию
	deviation := (rand.Float64()*2 - 1) * 0.15 // от -0.15 до +0.15
	// TODO: мне кажется, рандомное значение не должно быть !!!!!!!!!!!
	amount := baseAmount * (1.0 + deviation)
	amount = utils.RoundToCents(amount)

	return amount, nil
}

func (s *baseAmountService) CalculateUtilitiesAmount(userIDStr string, isFirstMonth bool, monthStr string) (float64, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return 0, fmt.Errorf("invalid userID format: %w", err)
	}

	if isFirstMonth {
		// Проверяем, нужно ли сохранять новую базовую сумму
		// Сохраняем только если записи нет или запрашиваемый месяц раньше сохраненного
		savedFirstMonth, _ := s.GetUtilitiesFirstMonth(userIDStr)
		shouldSave := (savedFirstMonth == "" || monthStr < savedFirstMonth)

		if shouldSave {
			// [15][16] Первый месяц: фиксируется в диапазоне $200–500
			amount := 200.0 + rand.Float64()*(500.0-200.0)
			amount = utils.RoundToCents(amount)

			// Сохраняем базовую сумму с месяцем из запроса
			if monthStr == "" {
				monthStr = time.Now().Format("2006-01")
			}
			if err := s.stateRepo.SaveUtilitiesBaseAmount(userID, amount, monthStr); err != nil {
				return 0, fmt.Errorf("failed to save utilities base amount: %w", err)
			}

			return amount, nil
		} else {
			// Используем сохраненную базовую сумму (для повторной генерации того же месяца)
			baseAmount, err := s.stateRepo.GetUtilitiesBaseAmount(userID)
			if err != nil {
				return 0, fmt.Errorf("failed to get utilities base amount: %w", err)
			}
			return baseAmount, nil
		}
	}

	// Последующие месяцы: ±15% от базовой суммы
	baseAmount, err := s.stateRepo.GetUtilitiesBaseAmount(userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get utilities base amount: %w", err)
	}

	if baseAmount == 0 {
		return 0, errors.New("utilities base amount not found. Generate first month first")
	}

	// Применяем ±15% вариацию
	deviation := (rand.Float64()*2 - 1) * 0.15 // от -0.15 до +0.15
	amount := baseAmount * (1.0 + deviation)
	// TODO: мне кажется, рандомное значение не должно быть !!!!!!!!!!!
	amount = utils.RoundToCents(amount)

	return amount, nil
}

func (s *baseAmountService) CalculateLeasingAmount(userIDStr string, turnover float64, isFirstMonth bool, monthStr string) (float64, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return 0, fmt.Errorf("invalid userID format: %w", err)
	}

	if isFirstMonth {
		// Проверяем, нужно ли сохранять новую базовую сумму
		// Сохраняем только если записи нет или запрашиваемый месяц раньше сохраненного
		savedFirstMonth, _ := s.GetLeasingFirstMonth(userIDStr)
		shouldSave := (savedFirstMonth == "" || monthStr < savedFirstMonth)

		if shouldSave {
			if turnover <= 0 {
				return 0, errors.New("turnover must be greater than 0 for first month leasing calculation")
			}

			// [19] Первый месяц: ~11.5–12% оборота
			percentage := 0.115 + rand.Float64()*(0.12-0.115) // 11.5% - 12%
			amount := turnover * percentage
			amount = utils.RoundToCents(amount)

			// Сохраняем базовую сумму с turnover и месяцем из запроса
			if monthStr == "" {
				monthStr = time.Now().Format("2006-01")
			}
			if err := s.stateRepo.SaveLeasingBaseAmount(userID, amount, monthStr, turnover); err != nil {
				return 0, fmt.Errorf("failed to save leasing base amount: %w", err)
			}

			return amount, nil
		} else {
			// Используем сохраненную базовую сумму (для повторной генерации того же месяца)
			baseAmount, err := s.stateRepo.GetLeasingBaseAmount(userID)
			if err != nil {
				return 0, fmt.Errorf("failed to get leasing base amount: %w", err)
			}
			return baseAmount, nil
		}
	}

	// [19] Последующие месяцы: повторяется 1:1 (без изменений)
	baseAmount, err := s.stateRepo.GetLeasingBaseAmount(userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get leasing base amount: %w", err)
	}

	if baseAmount == 0 {
		return 0, errors.New("leasing base amount not found. Generate first month first")
	}

	// Возвращаем без изменений (1:1)
	return baseAmount, nil
}

func (s *baseAmountService) SaveMobileBaseAmount(userIDStr string, amount float64, firstMonth string) error {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return err
	}
	return s.stateRepo.SaveMobileBaseAmount(userID, amount, firstMonth)
}

func (s *baseAmountService) SaveUtilitiesBaseAmount(userIDStr string, amount float64, firstMonth string) error {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return err
	}
	return s.stateRepo.SaveUtilitiesBaseAmount(userID, amount, firstMonth)
}

func (s *baseAmountService) SaveLeasingBaseAmount(userIDStr string, amount float64, firstMonth string, turnover float64) error {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fmt.Errorf("invalid userID format: %w", err)
	}
	return s.stateRepo.SaveLeasingBaseAmount(userID, amount, firstMonth, turnover)
}

func (s *baseAmountService) DeleteMobileBaseAmount(userIDStr string) error {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fmt.Errorf("invalid userID format: %w", err)
	}
	return s.stateRepo.DeleteMobileBaseAmount(userID)
}

func (s *baseAmountService) DeleteUtilitiesBaseAmount(userIDStr string) error {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fmt.Errorf("invalid userID format: %w", err)
	}
	return s.stateRepo.DeleteUtilitiesBaseAmount(userID)
}

func (s *baseAmountService) DeleteLeasingBaseAmount(userIDStr string) error {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fmt.Errorf("invalid userID format: %w", err)
	}
	return s.stateRepo.DeleteLeasingBaseAmount(userID)
}
