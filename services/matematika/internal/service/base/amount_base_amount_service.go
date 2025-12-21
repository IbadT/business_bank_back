package baseamountservice

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/internal/repository"
	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/dto"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/helpers"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/logger"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/utils"
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
	op := "service.baseAmount.getMobileBaseAmount"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"user_id": userIDStr})
	log.Debug("Getting mobile base amount")

	userID, err := helpers.ParseUserID(userIDStr)
	if err != nil {
		log.Error(err, "Invalid userID format")
		return 0, err
	}
	amount, err := s.stateRepo.GetMobileBaseAmount(userID)
	if err != nil {
		log.Error(err, "Failed to get mobile base amount from repository")
		return 0, err
	}
	log.WithFields(logger.Fields{"amount": amount}).Debug("Mobile base amount retrieved")
	return amount, nil
}

func (s *baseAmountService) GetUtilitiesBaseAmount(userIDStr string) (float64, error) {
	op := "service.baseAmount.getUtilitiesBaseAmount"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"user_id": userIDStr})
	log.Debug("Getting utilities base amount")

	userID, err := helpers.ParseUserID(userIDStr)
	if err != nil {
		log.Error(err, "Invalid userID format")
		return 0, err
	}
	amount, err := s.stateRepo.GetUtilitiesBaseAmount(userID)
	if err != nil {
		log.Error(err, "Failed to get utilities base amount from repository")
		return 0, err
	}
	log.WithFields(logger.Fields{"amount": amount}).Debug("Utilities base amount retrieved")
	return amount, nil
}

func (s *baseAmountService) GetLeasingBaseAmount(userIDStr string) (float64, error) {
	op := "service.baseAmount.getLeasingBaseAmount"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"user_id": userIDStr})
	log.Debug("Getting leasing base amount")

	userID, err := helpers.ParseUserID(userIDStr)
	if err != nil {
		log.Error(err, "Invalid userID format")
		return 0, err
	}
	amount, err := s.stateRepo.GetLeasingBaseAmount(userID)
	if err != nil {
		log.Error(err, "Failed to get leasing base amount from repository")
		return 0, err
	}
	log.WithFields(logger.Fields{"amount": amount}).Debug("Leasing base amount retrieved")
	return amount, nil
}

func (s *baseAmountService) GetBaseAmount(userIDStr string) (*dto.BaseAmountsResponse, error) {
	op := "service.baseAmount.getBaseAmount"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"user_id": userIDStr})
	log.Info("Getting base amounts")

	userID, err := helpers.ParseUserID(userIDStr)
	if err != nil {
		log.Error(err, "Invalid userID format")
		return nil, err
	}
	baseAmounts, err := s.stateRepo.GetBaseAmount(userID)
	if err != nil {
		log.Error(err, "Failed to get base amounts from repository")
		return nil, err
	}
	log.Success("Base amounts retrieved")
	return baseAmounts, nil
}

func (s *baseAmountService) GetMobileFirstMonth(userIDStr string) (string, error) {
	op := "service.baseAmount.getMobileFirstMonth"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"user_id": userIDStr})
	log.Debug("Getting mobile first month")

	userID, err := helpers.ParseUserID(userIDStr)
	if err != nil {
		log.Error(err, "Invalid userID format")
		return "", err
	}
	state, err := s.stateRepo.GetState(userID, "mobile_base_amount")
	if err != nil {
		// Реальная ошибка БД/сети - логируем, но возвращаем как "запись не найдена"
		// чтобы не блокировать генерацию. В будущем нужно добавить проверку истории генераций.
		log.Warn("GetState error for mobile_base_amount, treating as not found: %v", err)
		return "", nil // Возвращаем как "запись не найдена"
	}
	if state == nil {
		log.Debug("Mobile first month not found")
		return "", nil
	}
	firstMonth, _ := state.StateValue["first_month"].(string)
	log.WithFields(logger.Fields{"first_month": firstMonth}).Debug("Mobile first month retrieved")
	return firstMonth, nil
}

func (s *baseAmountService) GetUtilitiesFirstMonth(userIDStr string) (string, error) {
	op := "service.baseAmount.getUtilitiesFirstMonth"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"user_id": userIDStr})
	log.Debug("Getting utilities first month")

	userID, err := helpers.ParseUserID(userIDStr)
	if err != nil {
		log.Error(err, "Invalid userID format")
		return "", err
	}
	state, err := s.stateRepo.GetState(userID, "utilities_base_amount")
	if err != nil {
		// Реальная ошибка БД/сети - логируем, но возвращаем как "запись не найдена"
		// чтобы не блокировать генерацию. В будущем нужно добавить проверку истории генераций.
		log.Warn("GetState error for utilities_base_amount, treating as not found: %v", err)
		return "", nil // Возвращаем как "запись не найдена"
	}
	if state == nil {
		log.Debug("Utilities first month not found")
		return "", nil
	}
	firstMonth, _ := state.StateValue["first_month"].(string)
	log.WithFields(logger.Fields{"first_month": firstMonth}).Debug("Utilities first month retrieved")
	return firstMonth, nil
}

func (s *baseAmountService) GetLeasingFirstMonth(userIDStr string) (string, error) {
	op := "service.baseAmount.getLeasingFirstMonth"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"user_id": userIDStr})
	log.Debug("Getting leasing first month")

	userID, err := helpers.ParseUserID(userIDStr)
	if err != nil {
		log.Error(err, "Invalid userID format")
		return "", err
	}
	state, err := s.stateRepo.GetState(userID, "leasing_base_amount")
	if err != nil {
		// Реальная ошибка БД/сети - логируем, но возвращаем как "запись не найдена"
		// чтобы не блокировать генерацию. В будущем нужно добавить проверку истории генераций.
		log.Warn("GetState error for leasing_base_amount, treating as not found: %v", err)
		return "", nil // Возвращаем как "запись не найдена"
	}
	if state == nil {
		log.Debug("Leasing first month not found")
		return "", nil
	}
	firstMonth, _ := state.StateValue["first_month"].(string)
	log.WithFields(logger.Fields{"first_month": firstMonth}).Debug("Leasing first month retrieved")
	return firstMonth, nil
}

func (s *baseAmountService) CalculateMobileAmount(userIDStr string, isFirstMonth bool, monthStr string) (float64, error) {
	op := "service.baseAmount.calculateMobileAmount"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"user_id":      userIDStr,
		"is_first_month": isFirstMonth,
		"month":        monthStr,
	})
	log.Info("Calculating mobile amount")

	userID, err := helpers.ParseUserID(userIDStr)
	if err != nil {
		log.Error(err, "Invalid userID format")
		return 0, err
	}

	if isFirstMonth {
		log.Debug("First month calculation")
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
				log.Error(err, "Failed to save mobile base amount")
				return 0, fmt.Errorf("%w: %w", helpers.ErrFailedToSaveMobileBaseAmount, err)
			}

			log.WithFields(logger.Fields{"amount": amount, "first_month": monthStr}).Success("Mobile amount calculated for first month")
			return amount, nil
		} else {
			// Используем сохраненную базовую сумму (для повторной генерации того же месяца)
			baseAmount, err := s.stateRepo.GetMobileBaseAmount(userID)
			if err != nil {
				log.Error(err, "Failed to get mobile base amount")
				return 0, fmt.Errorf("%w: %w", helpers.ErrFailedToGetMobileBaseAmount, err)
			}
			log.WithFields(logger.Fields{"amount": baseAmount}).Debug("Using saved mobile base amount")
			return baseAmount, nil
		}
	}

	// Последующие месяцы: ±15% от базовой суммы
	baseAmount, err := s.stateRepo.GetMobileBaseAmount(userID)
	if err != nil {
		log.Error(err, "Failed to get mobile base amount")
		return 0, fmt.Errorf("failed to get mobile base amount: %w", err)
	}

	if baseAmount == 0 {
		log.Warn("Mobile base amount not found")
		return 0, helpers.ErrMobileBaseAmountNotFound
	}

	// Применяем ±15% вариацию
	deviation := (rand.Float64()*2 - 1) * 0.15 // от -0.15 до +0.15
	// TODO: мне кажется, рандомное значение не должно быть !!!!!!!!!!!
	amount := baseAmount * (1.0 + deviation)
	amount = utils.RoundToCents(amount)

	log.WithFields(logger.Fields{"amount": amount, "base_amount": baseAmount, "deviation": deviation}).Success("Mobile amount calculated")
	return amount, nil
}

func (s *baseAmountService) CalculateUtilitiesAmount(userIDStr string, isFirstMonth bool, monthStr string) (float64, error) {
	op := "service.baseAmount.calculateUtilitiesAmount"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"user_id":      userIDStr,
		"is_first_month": isFirstMonth,
		"month":        monthStr,
	})
	log.Info("Calculating utilities amount")

	userID, err := helpers.ParseUserID(userIDStr)
	if err != nil {
		log.Error(err, "Invalid userID format")
		return 0, err
	}

	if isFirstMonth {
		log.Debug("First month calculation")
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
				log.Error(err, "Failed to save utilities base amount")
				return 0, fmt.Errorf("%w: %w", helpers.ErrFailedToSaveUtilitiesBaseAmount, err)
			}

			log.WithFields(logger.Fields{"amount": amount, "first_month": monthStr}).Success("Utilities amount calculated for first month")
			return amount, nil
		} else {
			// Используем сохраненную базовую сумму (для повторной генерации того же месяца)
			baseAmount, err := s.stateRepo.GetUtilitiesBaseAmount(userID)
			if err != nil {
				log.Error(err, "Failed to get utilities base amount")
				return 0, fmt.Errorf("%w: %w", helpers.ErrFailedToGetUtilitiesBaseAmount, err)
			}
			log.WithFields(logger.Fields{"amount": baseAmount}).Debug("Using saved utilities base amount")
			return baseAmount, nil
		}
	}

	// Последующие месяцы: ±15% от базовой суммы
	baseAmount, err := s.stateRepo.GetUtilitiesBaseAmount(userID)
	if err != nil {
		log.Error(err, "Failed to get utilities base amount")
		return 0, fmt.Errorf("failed to get utilities base amount: %w", err)
	}

	if baseAmount == 0 {
		log.Warn("Utilities base amount not found")
		return 0, helpers.ErrUtilitiesBaseAmountNotFound
	}

	// Применяем ±15% вариацию
	deviation := (rand.Float64()*2 - 1) * 0.15 // от -0.15 до +0.15
	amount := baseAmount * (1.0 + deviation)
	// TODO: мне кажется, рандомное значение не должно быть !!!!!!!!!!!
	amount = utils.RoundToCents(amount)

	log.WithFields(logger.Fields{"amount": amount, "base_amount": baseAmount, "deviation": deviation}).Success("Utilities amount calculated")
	return amount, nil
}

func (s *baseAmountService) CalculateLeasingAmount(userIDStr string, turnover float64, isFirstMonth bool, monthStr string) (float64, error) {
	op := "service.baseAmount.calculateLeasingAmount"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"user_id":      userIDStr,
		"is_first_month": isFirstMonth,
		"month":        monthStr,
		"turnover":     turnover,
	})
	log.Info("Calculating leasing amount")

	userID, err := helpers.ParseUserID(userIDStr)
	if err != nil {
		log.Error(err, "Invalid userID format")
		return 0, err
	}

	if isFirstMonth {
		log.Debug("First month calculation")
		// Проверяем, нужно ли сохранять новую базовую сумму
		// Сохраняем только если записи нет или запрашиваемый месяц раньше сохраненного
		savedFirstMonth, _ := s.GetLeasingFirstMonth(userIDStr)
		shouldSave := (savedFirstMonth == "" || monthStr < savedFirstMonth)

		if shouldSave {
			if turnover <= 0 {
				log.Warn("Turnover must be greater than 0 for first month")
				return 0, helpers.ErrTurnoverMustBeGreaterThanZeroForLeasing
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
				log.Error(err, "Failed to save leasing base amount")
				return 0, fmt.Errorf("%w: %w", helpers.ErrFailedToSaveLeasingBaseAmount, err)
			}

			log.WithFields(logger.Fields{"amount": amount, "first_month": monthStr, "percentage": percentage}).Success("Leasing amount calculated for first month")
			return amount, nil
		} else {
			// Используем сохраненную базовую сумму (для повторной генерации того же месяца)
			baseAmount, err := s.stateRepo.GetLeasingBaseAmount(userID)
			if err != nil {
				log.Error(err, "Failed to get leasing base amount")
				return 0, fmt.Errorf("%w: %w", helpers.ErrFailedToGetLeasingBaseAmount, err)
			}
			log.WithFields(logger.Fields{"amount": baseAmount}).Debug("Using saved leasing base amount")
			return baseAmount, nil
		}
	}

	// [19] Последующие месяцы: повторяется 1:1 (без изменений)
	baseAmount, err := s.stateRepo.GetLeasingBaseAmount(userID)
	if err != nil {
		log.Error(err, "Failed to get leasing base amount")
		return 0, fmt.Errorf("failed to get leasing base amount: %w", err)
	}

	if baseAmount == 0 {
		log.Warn("Leasing base amount not found")
		return 0, helpers.ErrLeasingBaseAmountNotFound
	}

	log.WithFields(logger.Fields{"amount": baseAmount}).Success("Leasing amount calculated (1:1)")
	// Возвращаем без изменений (1:1)
	return baseAmount, nil
}

func (s *baseAmountService) SaveMobileBaseAmount(userIDStr string, amount float64, firstMonth string) error {
	op := "service.baseAmount.saveMobileBaseAmount"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"user_id":     userIDStr,
		"amount":      amount,
		"first_month": firstMonth,
	})
	log.Info("Saving mobile base amount")

	userID, err := helpers.ParseUserID(userIDStr)
	if err != nil {
		log.Error(err, "Invalid userID format")
		return err
	}
	if err := s.stateRepo.SaveMobileBaseAmount(userID, amount, firstMonth); err != nil {
		log.Error(err, "Failed to save mobile base amount to repository")
		return err
	}
	log.Success("Mobile base amount saved successfully")
	return nil
}

func (s *baseAmountService) SaveUtilitiesBaseAmount(userIDStr string, amount float64, firstMonth string) error {
	op := "service.baseAmount.saveUtilitiesBaseAmount"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"user_id":     userIDStr,
		"amount":      amount,
		"first_month": firstMonth,
	})
	log.Info("Saving utilities base amount")

	userID, err := helpers.ParseUserID(userIDStr)
	if err != nil {
		log.Error(err, "Invalid userID format")
		return err
	}
	if err := s.stateRepo.SaveUtilitiesBaseAmount(userID, amount, firstMonth); err != nil {
		log.Error(err, "Failed to save utilities base amount to repository")
		return err
	}
	log.Success("Utilities base amount saved successfully")
	return nil
}

func (s *baseAmountService) SaveLeasingBaseAmount(userIDStr string, amount float64, firstMonth string, turnover float64) error {
	op := "service.baseAmount.saveLeasingBaseAmount"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"user_id":     userIDStr,
		"amount":      amount,
		"first_month": firstMonth,
		"turnover":    turnover,
	})
	log.Info("Saving leasing base amount")

	userID, err := helpers.ParseUserID(userIDStr)
	if err != nil {
		log.Error(err, "Invalid userID format")
		return fmt.Errorf("invalid userID format: %w", err)
	}
	if err := s.stateRepo.SaveLeasingBaseAmount(userID, amount, firstMonth, turnover); err != nil {
		log.Error(err, "Failed to save leasing base amount to repository")
		return err
	}
	log.Success("Leasing base amount saved successfully")
	return nil
}

func (s *baseAmountService) DeleteMobileBaseAmount(userIDStr string) error {
	op := "service.baseAmount.deleteMobileBaseAmount"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"user_id": userIDStr})
	log.Info("Deleting mobile base amount")

	userID, err := helpers.ParseUserID(userIDStr)
	if err != nil {
		log.Error(err, "Invalid userID format")
		return fmt.Errorf("invalid userID format: %w", err)
	}
	if err := s.stateRepo.DeleteMobileBaseAmount(userID); err != nil {
		log.Error(err, "Failed to delete mobile base amount")
		return err
	}

	log.Success("Mobile base amount deleted successfully")
	return nil
}

func (s *baseAmountService) DeleteUtilitiesBaseAmount(userIDStr string) error {
	op := "service.baseAmount.deleteUtilitiesBaseAmount"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"user_id": userIDStr})
	log.Info("Deleting utilities base amount")

	userID, err := helpers.ParseUserID(userIDStr)
	if err != nil {
		log.Error(err, "Invalid userID format")
		return fmt.Errorf("invalid userID format: %w", err)
	}
	if err := s.stateRepo.DeleteUtilitiesBaseAmount(userID); err != nil {
		log.Error(err, "Failed to delete utilities base amount")
		return err
	}

	log.Success("Utilities base amount deleted successfully")
	return nil
}

func (s *baseAmountService) DeleteLeasingBaseAmount(userIDStr string) error {
	op := "service.baseAmount.deleteLeasingBaseAmount"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"user_id": userIDStr})
	log.Info("Deleting leasing base amount")

	userID, err := helpers.ParseUserID(userIDStr)
	if err != nil {
		log.Error(err, "Invalid userID format")
		return fmt.Errorf("invalid userID format: %w", err)
	}
	if err := s.stateRepo.DeleteLeasingBaseAmount(userID); err != nil {
		log.Error(err, "Failed to delete leasing base amount")
		return err
	}

	log.Success("Leasing base amount deleted successfully")
	return nil
}
