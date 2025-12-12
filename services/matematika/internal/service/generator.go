// internal/service/generator.go
package service

import (
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/internal/domain"
	"github.com/IbadT/business_bank_back/services/matematika/internal/domain/entities"
	"github.com/IbadT/business_bank_back/services/matematika/internal/domain/value_objects"
	"github.com/IbadT/business_bank_back/services/matematika/internal/models"
	"github.com/IbadT/business_bank_back/services/matematika/internal/repository"
	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/dto"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/utils"
	"github.com/google/uuid"
)

var (
	ErrNegativeBalance = errors.New("transaction would result in negative balance") // [43]
	ErrInvalidModel    = errors.New("invalid business model")
	ErrInvalidRequest  = errors.New("invalid request parameters")
	ErrUnauthorized    = errors.New("user authentication required") // Требуется авторизация
)

// GeneratorService - Use Case генерации транзакций
type GeneratorService interface {
	GenerateTransactions(req *dto.GenerateRequest, userID *string) (*dto.GenerateResponse, error)
}

type generatorService struct {
	configRepo               repository.ConfigRepository
	stateRepo                repository.StateRepository
	userRepo                 repository.UserRepository
	holidayRepo              repository.HolidayRepository
	gatewayRepo              repository.GatewayRepository
	generationRequestRepo    repository.GenerationRequestRepository
	transactionRepo          repository.TransactionRepository
	dateCalculator           *dateCalculator
	amountCalculator         *amountCalculator
	holidayService           HolidayService
	gatewayService           GatewayService
	baseAmountService        BaseAmountService
	breakdownService         BreakdownService
	balanceAdjustmentService BalanceAdjustmentService
	templates                []*entities.TransactionTemplate
	gateways                 []*entities.Gateway
	customers                []*entities.Customer
}

func NewGeneratorService(configRepo repository.ConfigRepository, stateRepo repository.StateRepository, userRepo repository.UserRepository, holidayRepo repository.HolidayRepository, gatewayRepo repository.GatewayRepository, holidayService HolidayService, gatewayService GatewayService, baseAmountService BaseAmountService, breakdownService BreakdownService, balanceAdjustmentService BalanceAdjustmentService, generationRequestRepo repository.GenerationRequestRepository, transactionRepo repository.TransactionRepository) (GeneratorService, error) {
	// Загружаем доменные сущности
	holidays, err := configRepo.GetHolidays()
	if err != nil {
		return nil, err
	}

	templates, err := configRepo.GetTransactionTemplates()
	if err != nil {
		return nil, err
	}

	gateways, err := configRepo.GetGateways()
	if err != nil {
		return nil, err
	}

	customers, err := configRepo.GetCustomers()
	if err != nil {
		return nil, err
	}

	return &generatorService{
		configRepo:               configRepo,
		stateRepo:                stateRepo,
		userRepo:                 userRepo,
		holidayRepo:              holidayRepo,
		gatewayRepo:              gatewayRepo,
		generationRequestRepo:    generationRequestRepo,
		transactionRepo:          transactionRepo,
		dateCalculator:           newDateCalculator(holidays, stateRepo, holidayService),
		amountCalculator:         newAmountCalculator(),
		holidayService:           holidayService,
		gatewayService:           gatewayService,
		baseAmountService:        baseAmountService,
		breakdownService:         breakdownService,
		balanceAdjustmentService: balanceAdjustmentService,
		templates:                templates,
		gateways:                 gateways,
		customers:                customers,
	}, nil
}

func (s *generatorService) GenerateTransactions(req *dto.GenerateRequest, userID *string) (*dto.GenerateResponse, error) {
	// [1] Валидация запроса
	if err := s.validateRequest(req); err != nil {
		return nil, err
	}

	// Проверка авторизации - userID обязателен
	if userID == nil || *userID == "" {
		return nil, ErrUnauthorized
	}

	// Создаем GenerationRequest в БД перед генерацией
	var userIDUUID *uuid.UUID
	parsed, err := uuid.Parse(*userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}
	userIDUUID = &parsed

	monthStr := fmt.Sprintf("%d-%02d", req.Year, req.Month)
	var customData models.JSONB
	if req.CustomData != nil {
		// Конвертируем CustomData в JSONB (упрощенная версия)
		customData = make(models.JSONB)
		if len(req.CustomData.ManualTransactions) > 0 {
			customData["manualTransactions"] = req.CustomData.ManualTransactions
		}
		if req.CustomData.CompanyInfo.OwnerName != "" || req.CustomData.CompanyInfo.CompanyName != "" {
			customData["companyInfo"] = req.CustomData.CompanyInfo
		}
		if len(req.CustomData.CustomCustomers) > 0 {
			customData["customCustomers"] = req.CustomData.CustomCustomers
		}
		if len(req.CustomData.CustomContractors) > 0 {
			customData["customContractors"] = req.CustomData.CustomContractors
		}
	}

	generationRequest := &models.GenerationRequest{
		UserID:               userIDUUID,
		Month:                monthStr,
		Year:                 req.Year,
		Turnover:             req.Turnover,
		DesiredProfitPercent: req.DesiredProfitPercent,
		Model:                req.Model,
		InitialBalance:       req.InitialBalance,
		ScaleFactor:          req.ScaleFactor,
		CustomData:           customData,
		Status:               "processing",
	}

	createdRequest, err := s.generationRequestRepo.Create(generationRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to create generation request: %w", err)
	}

	requestID := createdRequest.ID

	// [1][5] Рассчет целевой прибыли
	targetProfit := req.Turnover * (req.DesiredProfitPercent / 100)
	totalExpensesTarget := req.Turnover - targetProfit

	// [2][35][36] Генерация доходов в зависимости от модели
	var incomeTransactions []*entities.Transaction

	switch req.Model {
	case "B2C":
		incomeTransactions, err = s.generateB2CIncomes(req, userID)
		if err != nil {
			errorMsg := err.Error()
			s.generationRequestRepo.UpdateStatus(requestID, "failed", &errorMsg)
			return nil, err
		}
	case "B2B":
		incomeTransactions = s.generateB2BIncomes(req)
	default:
		errorMsg := ErrInvalidModel.Error()
		s.generationRequestRepo.UpdateStatus(requestID, "failed", &errorMsg)
		return nil, ErrInvalidModel
	}

	// [7-21] Генерация расходов
	expenseTransactions, err := s.generateExpenses(req, totalExpensesTarget, userID)
	if err != nil {
		errorMsg := err.Error()
		s.generationRequestRepo.UpdateStatus(requestID, "failed", &errorMsg)
		return nil, err
	}

	// [38] Объединение транзакций
	allTransactions := append(incomeTransactions, expenseTransactions...)

	// [38] Добавление ручных транзакций
	var manualIncomeAmount, manualExpenseAmount float64
	if req.CustomData != nil && len(req.CustomData.ManualTransactions) > 0 {
		manualTransactions := s.convertManualTransactions(req.CustomData.ManualTransactions)
		allTransactions = append(allTransactions, manualTransactions...)

		// [38] Учитываем ручные транзакции при расчете целевой прибыли
		for _, tx := range manualTransactions {
			if tx.IsIncome() {
				manualIncomeAmount += tx.Amount
			} else {
				manualExpenseAmount += tx.Amount // расходы отрицательные
			}
		}
	}

	// [38] Масштабирование
	if req.ScaleFactor > 1 {
		allTransactions = s.scaleTransactions(allTransactions, req.ScaleFactor)
	}

	// Сортировка по дате
	allTransactions = s.sortTransactionsByDate(allTransactions)

	// [42] Балансировка и нормализация
	// Скорректируем targetProfit с учетом ручных транзакций
	// Ручные доходы увеличивают оборот, ручные расходы уменьшают прибыль
	adjustedTurnover := req.Turnover + manualIncomeAmount
	adjustedTargetProfit := targetProfit + manualIncomeAmount + manualExpenseAmount // manualExpenseAmount отрицательный

	balancedTransactions, err := s.balanceAndNormalize(allTransactions, adjustedTurnover, adjustedTargetProfit)
	if err != nil {
		errorMsg := err.Error()
		s.generationRequestRepo.UpdateStatus(requestID, "failed", &errorMsg)
		return nil, err
	}

	// Расчет балансов
	transactionsWithBalance, err := s.calculateBalances(balancedTransactions, req.InitialBalance)
	if err != nil {
		// Если есть ошибка недостатка баланса, пытаемся скорректировать
		if strings.Contains(err.Error(), "insufficient balance") {
			// Используем стратегию "postpone" по умолчанию (можно добавить в запрос позже)
			strategy := StrategyPostpone
			adjustedTransactions, adjustments, adjustErr := s.balanceAdjustmentService.AdjustTransactionsForBalance(
				balancedTransactions,
				req.InitialBalance,
				strategy,
				s.dateCalculator,
				req.Year,
				req.Month,
			)
			if adjustErr != nil {
				errorMsg := fmt.Sprintf("failed to adjust transactions: %v (original error: %v)", adjustErr, err)
				s.generationRequestRepo.UpdateStatus(requestID, "failed", &errorMsg)
				return nil, fmt.Errorf("failed to adjust transactions: %w (original error: %v)", adjustErr, err)
			}

			// Пересчитываем балансы после корректировки
			transactionsWithBalance, err = s.recalculateBalances(adjustedTransactions, req.InitialBalance)
			if err != nil {
				errorMsg := fmt.Sprintf("failed to recalculate balances after adjustment: %v", err)
				s.generationRequestRepo.UpdateStatus(requestID, "failed", &errorMsg)
				return nil, fmt.Errorf("failed to recalculate balances after adjustment: %w", err)
			}

			// Логируем корректировки
			if len(adjustments) > 0 {
				log.Printf("[INFO] Applied %d balance adjustments", len(adjustments))
			}
		} else {
			return nil, err
		}
	}

	// [43] Проверка отрицательного баланса (после корректировки)
	if err := s.checkNegativeBalance(transactionsWithBalance); err != nil {
		// Если после корректировки все еще есть отрицательный баланс, пробуем уменьшить суммы
		strategy := StrategyReduce
		adjustedTransactions, adjustments, adjustErr := s.balanceAdjustmentService.AdjustTransactionsForBalance(
			transactionsWithBalance,
			req.InitialBalance,
			strategy,
			s.dateCalculator,
			req.Year,
			req.Month,
		)
		if adjustErr != nil {
			errorMsg := fmt.Sprintf("failed to adjust transactions by reducing amounts: %v (original error: %v)", adjustErr, err)
			s.generationRequestRepo.UpdateStatus(requestID, "failed", &errorMsg)
			return nil, fmt.Errorf("failed to adjust transactions by reducing amounts: %w (original error: %v)", adjustErr, err)
		}

		// Пересчитываем балансы после корректировки
		transactionsWithBalance, err = s.recalculateBalances(adjustedTransactions, req.InitialBalance)
		if err != nil {
			errorMsg := fmt.Sprintf("failed to recalculate balances after reduction: %v", err)
			s.generationRequestRepo.UpdateStatus(requestID, "failed", &errorMsg)
			return nil, fmt.Errorf("failed to recalculate balances after reduction: %w", err)
		}

		// Логируем корректировки
		if len(adjustments) > 0 {
			log.Printf("[INFO] Applied %d balance adjustments by reducing amounts", len(adjustments))
		}

		// Финальная проверка
		if err := s.checkNegativeBalance(transactionsWithBalance); err != nil {
			// Обновляем статус на failed
			errorMsg := err.Error()
			s.generationRequestRepo.UpdateStatus(requestID, "failed", &errorMsg)
			return nil, fmt.Errorf("negative balance still exists after all adjustments: %w", err)
		}
	}

	// TODO: нужно ли это, если в balance_adjustment_service уже есть сортировка ?????
	// Финальная сортировка перед сохранением и формированием ответа
	// Гарантируем, что транзакции отсортированы по transactionDate
	transactionsWithBalance = s.sortTransactionsByDate(transactionsWithBalance)

	// [168] Помечаем первые транзакции каждой категории флагом FixAsFirst
	s.markFirstTransactionsByCategory(transactionsWithBalance)

	// Конвертируем entities.Transaction в domain.GeneratedTransaction для сохранения в БД
	domainTransactions := s.convertToDomainTransactions(transactionsWithBalance, requestID)

	// Сохраняем транзакции в БД
	if err := s.transactionRepo.CreateBatch(domainTransactions); err != nil {
		log.Printf("[ERROR] Failed to save transactions to database: %v", err)
		errorMsg := fmt.Sprintf("failed to save transactions to database: %v", err)
		s.generationRequestRepo.UpdateStatus(requestID, "failed", &errorMsg)
		// Не прерываем выполнение, но обновляем статус и логируем ошибку
	} else {
		log.Printf("[INFO] Saved %d transactions to database for request_id: %s", len(domainTransactions), requestID)
	}

	// Обновляем статус GenerationRequest на "completed"
	completedAt := time.Now()
	if err := s.generationRequestRepo.UpdateCompletedAt(requestID, completedAt); err != nil {
		log.Printf("[ERROR] Failed to update generation request status: %v", err)
		// Не прерываем выполнение, но логируем ошибку
	}

	// Формирование ответа
	// userIDUUID гарантированно не nil, так как проверено выше
	return s.buildResponse(transactionsWithBalance, req, requestID, *userIDUUID), nil
}

func (s *generatorService) validateRequest(req *dto.GenerateRequest) error {
	if req.Turnover <= 0 {
		return errors.New("turnover must be greater than 0")
	}
	if req.DesiredProfitPercent < 0 || req.DesiredProfitPercent > 100 {
		return errors.New("desired profit percent must be between 0 and 100")
	}
	if req.Model != "B2C" && req.Model != "B2B" {
		return errors.New("model must be either B2C or B2B")
	}
	if req.InitialBalance < 0 {
		return errors.New("initial balance cannot be negative")
	}
	return nil
}

func (s *generatorService) generateB2CIncomes(req *dto.GenerateRequest, userID *string) ([]*entities.Transaction, error) {
	// [35] Получаем пятницы в месяце
	fridays := s.dateCalculator.getFridaysInMonth(req.Year, req.Month)

	// [35] Получаем или выбираем шлюз (сохраняем для всех месяцев)
	var gateway *entities.Gateway
	var err error

	if userID != nil {
		// Парсим userID из string в UUID
		userUUID, parseErr := uuid.Parse(*userID)
		if parseErr == nil {
			// Используем GatewayService для получения сохраненного шлюза
			gateway, err = s.gatewayService.GetB2CGateways(userUUID)
			if err != nil {
				return nil, err
			}
		}
	}

	// Если шлюз не найден - выбираем случайный и сохраняем
	if gateway == nil {
		gateway = s.selectGateway()

		// [35] Сохраняем выбранный шлюз для пользователя через GatewayService
		if userID != nil {
			userUUID, parseErr := uuid.Parse(*userID)
			if parseErr == nil {
				// Сохраняем через сервис (передаем ID выбранного шлюза)
				if err := s.gatewayService.SaveB2CGateways(userUUID, gateway.ID); err != nil {
					// Логируем ошибку, но не прерываем генерацию
					// Шлюз уже выбран, генерация может продолжиться
					log.Printf("[WARN] Failed to save gateway via GatewayService: %v", err)
				}
			}
		}
	}

	var transactions []*entities.Transaction
	totalGenerated := 0.0

	// [34] Распределение в зависимости от количества пятниц
	basePercentage := 1.0 / float64(len(fridays))

	for i, friday := range fridays {
		// [35] ±4.5% отклонение
		deviation := s.amountCalculator.generateDeviation(0.045)
		percentage := basePercentage + deviation

		var amount float64
		if i == len(fridays)-1 {
			// Корректировка последней транзакции
			amount = req.Turnover - totalGenerated
		} else {
			amount = req.Turnover * percentage
			totalGenerated += amount
		}

		// [33] Генерация времени в рабочие часы (08:00-18:00)
		transactionTime := s.dateCalculator.generateBusinessTime(friday, 8, 18)

		// Создание доменной сущности
		transaction := entities.NewTransaction(
			generateTransactionID("inc", i+1),
			transactionTime,
			friday,
			value_objects.Income,
			gateway.Name,
			value_objects.ACHCredit,
			utils.RoundToCents(amount),
		)
		transaction.SetBalanceAfter(0) // Будет рассчитано позже

		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

func (s *generatorService) generateB2BIncomes(req *dto.GenerateRequest) []*entities.Transaction {
	var transactions []*entities.Transaction
	transactionCounter := 1
	totalGenerated := 0.0

	// Используем кастомных клиентов если есть [6]
	customers := s.customers
	if req.CustomData != nil && len(req.CustomData.CustomCustomers) > 0 {
		customers = s.createCustomCustomers(req.CustomData.CustomCustomers)
	}

	for _, customer := range customers {
		// [36] Количество транзакций для клиента
		numTransactions := customer.GetTransactionCount()

		for i := 0; i < numTransactions; i++ {
			// [37] Расчет суммы
			amount := customer.CalculateAmount(req.Turnover)

			// [6] Распределение методов платежа
			paymentMethod := s.selectPaymentMethod()

			// Генерация даты (будний день, не праздник)
			transactionDate := s.dateCalculator.generateRandomBusinessDate(req.Year, req.Month)

			// [33] Генерация времени в рабочие часы (08:00-18:00)
			transactionTime := s.dateCalculator.generateBusinessTime(transactionDate, 8, 18)

			// [32] Проверка праздников для B2B-пополнений (ACH, Wire)
			if paymentMethod == value_objects.ACHCredit || paymentMethod == value_objects.Wire {
				if s.holidayService.IsHoliday(transactionDate) {
					transactionDate = s.holidayService.GetNextBusinessDay(transactionDate)
					transactionTime = s.dateCalculator.generateBusinessTime(transactionDate, 8, 18)
				}
			}

			// Создание транзакции
			transaction := entities.NewTransaction(
				generateTransactionID("inc", transactionCounter),
				transactionTime,
				transactionDate,
				value_objects.Income,
				customer.Name,
				paymentMethod,
				utils.RoundToCents(amount),
			)
			transaction.SetBalanceAfter(0) // Будет рассчитано позже

			transactions = append(transactions, transaction)
			totalGenerated += amount
			transactionCounter++

			// [6] Ограничение на количество транзакций
			if transactionCounter > 20 {
				break
			}
		}

		if transactionCounter > 20 {
			break
		}
	}

	// Корректировка последней транзакции
	if len(transactions) > 0 {
		correction := req.Turnover - totalGenerated
		lastTransaction := transactions[len(transactions)-1]
		lastTransaction.Amount = utils.RoundToCents(lastTransaction.Amount + correction)
	}

	return transactions
}

func (s *generatorService) generateExpenses(req *dto.GenerateRequest, totalExpensesTarget float64, userID *string) ([]*entities.Transaction, error) {
	var transactions []*entities.Transaction
	totalGenerated := 0.0

	// Разделяем шаблоны на обязательные и опциональные [39-41]
	mandatoryTemplates, optionalTemplates := s.separateTemplates()
	log.Printf("[DEBUG] separateTemplates: mandatory=%d, optional=%d", len(mandatoryTemplates), len(optionalTemplates))

	// Генерация обязательных расходов
	for _, template := range mandatoryTemplates {
		templateTransactions, amount := s.generateTransactionsFromTemplate(req, template, userID)
		transactions = append(transactions, templateTransactions...)
		totalGenerated += amount
	}

	// [12][40] Генерация опциональных расходов в пределах бюджета
	remainingBudget := totalExpensesTarget - totalGenerated
	log.Printf("[DEBUG] generateExpenses: totalExpensesTarget=%.2f, totalGenerated=%.2f, remainingBudget=%.2f",
		totalExpensesTarget, totalGenerated, remainingBudget)

	// Проверка: если обязательные расходы превышают бюджет, логируем предупреждение
	// Нормализация позже скорректирует суммы для достижения целевой прибыли
	if remainingBudget < 0 {
		log.Printf("[WARN] Mandatory expenses (%.2f) exceed totalExpensesTarget (%.2f) by %.2f. Normalization will adjust.",
			totalGenerated, totalExpensesTarget, -remainingBudget)
	}
	if remainingBudget > 0 {
		// Перемешиваем для случайного выбора [41]
		shuffledOptional := s.shuffleTemplates(optionalTemplates)

		// [25][14] Приоритет для "Подписка ПО" - всегда генерируем если есть бюджет
		var softwareSubscriptionTemplate *entities.TransactionTemplate
		var otherOptionalTemplates []*entities.TransactionTemplate
		log.Printf("[DEBUG] Processing %d optional templates", len(shuffledOptional))
		for _, template := range shuffledOptional {
			log.Printf("[DEBUG] Optional template: category=%s, isOptional=%v", template.Category, template.IsOptional)
			if template.Category == "Подписка ПО" {
				softwareSubscriptionTemplate = template
				log.Printf("[DEBUG] Found Подписка ПО template!")
			} else {
				otherOptionalTemplates = append(otherOptionalTemplates, template)
			}
		}

		// Сначала генерируем "Подписка ПО" если есть
		if softwareSubscriptionTemplate != nil && remainingBudget > 0 {
			log.Printf("[DEBUG] Generating Подписка ПО: remainingBudget=%.2f", remainingBudget)
			templateTransactions, amount := s.generateTransactionsFromTemplate(req, softwareSubscriptionTemplate, userID)
			log.Printf("[DEBUG] Подписка ПО generated: amount=%.2f, transactions=%d", amount, len(templateTransactions))
			if amount <= remainingBudget {
				transactions = append(transactions, templateTransactions...)
				totalGenerated += amount
				remainingBudget -= amount
				log.Printf("[DEBUG] Подписка ПО added to transactions, new remainingBudget=%.2f", remainingBudget)
			} else {
				log.Printf("[DEBUG] Подписка ПО amount (%.2f) exceeds remainingBudget (%.2f)", amount, remainingBudget)
			}
		} else {
			if softwareSubscriptionTemplate == nil {
				log.Printf("[DEBUG] softwareSubscriptionTemplate is nil - Подписка ПО not found in optional templates")
			} else {
				log.Printf("[DEBUG] remainingBudget <= 0: %.2f", remainingBudget)
			}
		}

		// Затем остальные опциональные
		for _, template := range otherOptionalTemplates {
			if remainingBudget <= 0 {
				break
			}

			templateTransactions, amount := s.generateTransactionsFromTemplate(req, template, userID)
			if amount <= remainingBudget {
				transactions = append(transactions, templateTransactions...)
				totalGenerated += amount
				remainingBudget -= amount
			}
		}
	}

	return transactions, nil
}

func (s *generatorService) generateTransactionsFromTemplate(
	req *dto.GenerateRequest,
	template *entities.TransactionTemplate,
	userID *string,
) ([]*entities.Transaction, float64) {
	var transactions []*entities.Transaction
	totalAmount := 0.0

	// Количество транзакций по шаблону
	// numTransactions := template.GetOccurrences()
	var numTransactions int
	if template.Category == "IRS налоги" || template.Category == "IRS" {
		numTransactions = s.getIRSOccurrences(template, req.Year, req.Month)
	} else {
		numTransactions = template.GetOccurrences()
	}

	// Для процентных транзакций: сначала рассчитываем общую сумму, затем делим на количество
	// Это исправляет баг, когда каждая транзакция получала полный процент от оборота
	var totalCategoryAmount float64
	var calculationDetails map[string]interface{}

	if template.IsPercentage {
		// Рассчитываем общую сумму для всей категории
		percentage := template.PercentageRange.Min +
			(rand.Float64() * (template.PercentageRange.Max - template.PercentageRange.Min))
		totalCategoryAmount = req.Turnover * percentage
		// Округляем общую сумму для категории, чтобы избежать ошибок округления при распределении
		totalCategoryAmount = utils.RoundToCents(totalCategoryAmount)
	} else {
		// Для фиксированных операций с расчетами [20][21]
		fixedAmount, details := s.calculateFixedAmount(template.Category, template.FixedAmount, req, userID)
		totalCategoryAmount = fixedAmount
		calculationDetails = details
	}

	for i := 0; i < numTransactions; i++ {
		// Расчет суммы для каждой транзакции
		var amount float64

		if template.IsPercentage {
			// Для процентных: делим общую сумму на количество транзакций
			if i == numTransactions-1 {
				// Последняя транзакция: корректируем для точного соответствия
				// totalAmount накапливает округленные положительные значения до применения знака
				// Используем округленный totalCategoryAmount и округленный totalAmount
				amount = totalCategoryAmount - totalAmount
				// Округляем последнюю транзакцию
				amount = utils.RoundToCents(amount)
				// Накапливаем для последней транзакции (до применения знака)
				totalAmount += amount
			} else {
				// Остальные: равномерное распределение
				amount = totalCategoryAmount / float64(numTransactions)
				// Округляем каждую транзакцию
				amount = utils.RoundToCents(amount)
				// Накапливаем округленное положительное значение до применения знака
				totalAmount += amount
			}
		} else {
			// Для фиксированных: используем уже рассчитанную сумму
			amount = totalCategoryAmount
			// Округляем фиксированную сумму
			amount = utils.RoundToCents(amount)
			// Для фиксированных накапливаем только один раз (если несколько транзакций - это ошибка конфигурации)
			if i == 0 {
				totalAmount += amount
			}
		}

		// Расходы должны быть отрицательными
		if template.Type == value_objects.Expense {
			amount = -amount
		}

		// Генерация даты транзакции
		var transactionDate time.Time
		var postingDate time.Time

		// [25][14] Для подписки ПО используем сохраненный день недели
		if template.Category == "Подписка ПО" {
			userIDUUID := s.getUserID(userID)
			log.Printf("[DEBUG] generateTransactionsFromTemplate: Подписка ПО detected, userID=%v, userIDUUID=%v", userID, userIDUUID)
			baseDate := time.Date(req.Year, time.Month(req.Month), 1, 0, 0, 0, 0, time.UTC)
			log.Printf("[DEBUG] Calling calculateSoftwareSubscriptionDate: baseDate=%v, userIDUUID=%v", baseDate.Format("2006-01-02"), userIDUUID)
			transactionDate = s.dateCalculator.calculateSoftwareSubscriptionDate(baseDate, userIDUUID)
			postingDate = transactionDate
			log.Printf("[DEBUG] Подписка ПО date generated: %v (weekday=%d)", transactionDate.Format("2006-01-02"), int(transactionDate.Weekday()))
		} else if template.Category == "IRS налоги" || template.Category == "IRS" {
			// [23][24] Для IRS налогов - всегда 15-е число (или следующий рабочий день)
			transactionDate = s.dateCalculator.calculateIRSDate(req.Year, req.Month, i+1)
			postingDate = transactionDate
		} else if template.Category == "Перевод владельцу" || template.Category == "Owner Transfer" {
			// [22] Для "Перевод владельцу" - 1 раз в месяц, в будний день (не праздничный)
			// generateRandomBusinessDate уже гарантирует будний день (не праздничный)
			transactionDate = s.dateCalculator.generateRandomBusinessDate(req.Year, req.Month)
			// postingDate должен быть таким же, как transactionDate, или скорректированным, если это праздник
			// Но generateRandomBusinessDate уже гарантирует будний день, поэтому postingDate = transactionDate
			postingDate = transactionDate
		} else {
			transactionDate = s.dateCalculator.calculateTransactionDate(template, req.Year, req.Month, i+1)
			postingDate = s.dateCalculator.calculatePostingDate(template, req.Year, req.Month, i+1)
		}

		// [32] Корректировка даты если праздник (для операций по счету)
		// Для "Перевод владельцу" дата уже гарантированно будний день (не праздничный), пропускаем
		// Для IRS налогов дата уже скорректирована в calculateIRSDate, пропускаем
		isOwnerTransfer := template.Category == "Перевод владельцу" || template.Category == "Owner Transfer"
		if template.PaymentMethod.IsAccountTransfer() && !isOwnerTransfer && template.Category != "IRS налоги" && template.Category != "IRS" {
			if s.holidayService.IsHoliday(transactionDate) {
				transactionDate = s.holidayService.GetNextBusinessDay(transactionDate)
			}
			if s.holidayService.IsHoliday(postingDate) {
				postingDate = s.holidayService.GetNextBusinessDay(postingDate)
			}
		}

		// [33] Генерация времени согласно BusinessHours
		var transactionTime time.Time
		if template.Category == "Подписка ПО" {
			// [33] 00:01 для подписок
			transactionTime = time.Date(
				transactionDate.Year(), transactionDate.Month(), transactionDate.Day(),
				0, 1, 0, 0, time.UTC)
		} else if template.PaymentMethod.IsCardOperation() {
			// [33] 09:00-20:00 для карт
			transactionTime = s.dateCalculator.generateBusinessTime(transactionDate, 9, 20)
		} else {
			// [33] 08:00-18:00 для операций по счету
			// Используем BusinessHours из шаблона если есть
			startHour, endHour := 8, 18
			if template.BusinessHours.Start != "" && template.BusinessHours.End != "" {
				startHour, _ = parseTimeFromString(template.BusinessHours.Start)
				endHour, _ = parseTimeFromString(template.BusinessHours.End)
			}
			transactionTime = s.dateCalculator.generateBusinessTime(transactionDate, startHour, endHour)
		}

		// Создание транзакции
		// amount уже округлен выше для процентных транзакций
		transaction := entities.NewTransaction(
			generateTemplateTransactionID(template.Category, i+1),
			transactionTime,
			postingDate,
			template.Type,
			template.Category,
			template.PaymentMethod,
			amount, // Уже округлен для процентных, для фиксированных тоже округлен
		)

		if calculationDetails != nil {
			transaction.SetCalculationDetails(calculationDetails)
		}

		transactions = append(transactions, transaction)
		// totalAmount уже накоплен выше для всех транзакций (до применения знака)
	}

	// Возвращаем общую сумму категории (для процентных это totalCategoryAmount, для фиксированных - сумма всех транзакций)
	return transactions, totalAmount
}

// isFirstMonthForCategory проверяет, является ли месяц первым для категории
// Использует проверку истории генераций для более надежного определения
func (s *generatorService) isFirstMonthForCategory(userID *string, categoryKey string, monthStr string) bool {
	if userID == nil || *userID == "" {
		// Если userID нет, используем fallback логику
		return true
	}

	userUUID, err := uuid.Parse(*userID)
	if err != nil {
		log.Printf("[WARN] Invalid userID in isFirstMonthForCategory: %v", err)
		return true
	}

	// 1. Проверяем сохраненный first_month из state
	savedFirstMonth := ""
	switch categoryKey {
	case "leasing":
		savedFirstMonth, _ = s.baseAmountService.GetLeasingFirstMonth(*userID)
	case "mobile":
		savedFirstMonth, _ = s.baseAmountService.GetMobileFirstMonth(*userID)
	case "utilities":
		savedFirstMonth, _ = s.baseAmountService.GetUtilitiesFirstMonth(*userID)
	}

	// Если есть сохраненный first_month и запрашиваемый месяц <= сохраненному, это первый месяц
	if savedFirstMonth != "" {
		return monthStr <= savedFirstMonth
	}

	// 2. Если сохраненного first_month нет, проверяем историю генераций
	// Если у пользователя есть завершенные генерации, это не первый месяц
	completedRequests, err := s.generationRequestRepo.GetCompletedByUserID(userUUID)
	if err != nil {
		// Если ошибка при проверке истории, логируем и считаем первым месяцем (fallback)
		log.Printf("[WARN] Failed to check generation history for userID=%s: %v, treating as first month", *userID, err)
		return true
	}

	// Если есть хотя бы одна завершенная генерация, это не первый месяц
	if len(completedRequests) > 0 {
		return false
	}

	// Если нет завершенных генераций и нет сохраненного first_month, это первый месяц
	return true
}

func (s *generatorService) calculateFixedAmount(category string, baseAmount float64, req *dto.GenerateRequest, userID *string) (float64, map[string]interface{}) {
	switch category {
	case "Перегруз":
		// [20][21] вес (200–1000 lb) * ставку ($0.011–$0.039)
		weight := 200 + rand.Intn(801) // 200-1000
		rate := 0.011 + rand.Float64()*(0.039-0.011)
		amount := float64(weight) * rate
		details := map[string]interface{}{
			"weight_lb":   weight,
			"rate_per_lb": fmt.Sprintf("%.3f", rate),
			"formula":     "weight * rate",
		}
		return amount, details

	case "Лизинг", "Leasing":
		// [19] Используем BaseAmountService для расчета лизинга
		if userID == nil || *userID == "" {
			// Fallback на старую логику если userID не указан
			firstMonth := s.amountCalculator.isFirstMonth(req)
			if firstMonth {
				percentage := 0.115 + rand.Float64()*(0.12-0.115)
				amount := req.Turnover * percentage
				details := map[string]interface{}{
					"type":                    "first_month_lease",
					"percentage_of_turnover":  percentage,
					"fixed_for_future_months": true,
				}
				s.amountCalculator.saveLeaseAmount(amount)
				s.amountCalculator.firstMonthTurnover = req.Turnover
				s.amountCalculator.isFirstMonthFlag = false
				return amount, details
			} else {
				amount := s.amountCalculator.getSavedLeaseAmount()
				details := map[string]interface{}{
					"type": "recurring_lease",
				}
				return amount, details
			}
		}

		// Проверяем, является ли это первым месяцем
		// Используем проверку истории генераций для более надежного определения
		monthStr := fmt.Sprintf("%d-%02d", req.Year, req.Month)
		isFirstMonth := s.isFirstMonthForCategory(userID, "leasing", monthStr)

		amount, err := s.baseAmountService.CalculateLeasingAmount(*userID, req.Turnover, isFirstMonth, monthStr)
		if err != nil {
			// Fallback на старую логику при ошибке
			log.Printf("[WARN] Failed to calculate leasing amount via BaseAmountService: %v, using fallback", err)
			firstMonth := s.amountCalculator.isFirstMonth(req)
			if firstMonth {
				// TODO: мне кажется, рандомное значение не должно быть !!!!!!!!!!!
				percentage := 0.115 + rand.Float64()*(0.12-0.115)
				amount = req.Turnover * percentage
			} else {
				amount = s.amountCalculator.getSavedLeaseAmount()
			}
		}

		details := map[string]interface{}{
			"type": "lease",
		}
		if isFirstMonth {
			details["is_first_month"] = true
		} else {
			details["is_first_month"] = false
		}
		return amount, details

	case "Мобильная связь", "Mobile":
		// [15][16] Используем BaseAmountService для расчета мобильной связи
		if userID == nil || *userID == "" {
			return baseAmount, nil
		}

		// Проверяем, является ли это первым месяцем
		// Используем проверку истории генераций для более надежного определения
		monthStr := fmt.Sprintf("%d-%02d", req.Year, req.Month)
		isFirstMonth := s.isFirstMonthForCategory(userID, "mobile", monthStr)

		amount, err := s.baseAmountService.CalculateMobileAmount(*userID, isFirstMonth, monthStr)
		if err != nil {
			log.Printf("[WARN] Failed to calculate mobile amount via BaseAmountService: %v, using baseAmount", err)
			return baseAmount, nil
		}

		details := map[string]interface{}{
			"type": "mobile",
		}
		if isFirstMonth {
			details["is_first_month"] = true
		} else {
			details["is_first_month"] = false
		}
		return amount, details

	case "Коммунальные", "Utilities":
		// [15][16] Используем BaseAmountService для расчета коммунальных
		if userID == nil || *userID == "" {
			return baseAmount, nil
		}

		// Проверяем, является ли это первым месяцем
		// Используем проверку истории генераций для более надежного определения
		monthStr := fmt.Sprintf("%d-%02d", req.Year, req.Month)
		isFirstMonth := s.isFirstMonthForCategory(userID, "utilities", monthStr)

		amount, err := s.baseAmountService.CalculateUtilitiesAmount(*userID, isFirstMonth, monthStr)
		if err != nil {
			log.Printf("[WARN] Failed to calculate utilities amount via BaseAmountService: %v, using baseAmount", err)
			return baseAmount, nil
		}

		details := map[string]interface{}{
			"type": "utilities",
		}
		if isFirstMonth {
			details["is_first_month"] = true
		} else {
			details["is_first_month"] = false
		}
		return amount, details

	default:
		return baseAmount, nil
	}
}

// parseTimeFromString парсит время из строки "HH:MM"
func parseTimeFromString(timeStr string) (int, int) {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return 8, 0 // По умолчанию
	}
	hour, _ := strconv.Atoi(parts[0])
	minute, _ := strconv.Atoi(parts[1])
	return hour, minute
}

// convertManualTransactions конвертирует ручные транзакции из DTO в entities
func (s *generatorService) convertManualTransactions(manual []dto.ManualTransaction) []*entities.Transaction {
	var transactions []*entities.Transaction
	for i, manualTx := range manual {
		transactionType, _ := value_objects.NewTransactionType(manualTx.Type)
		paymentMethod, _ := value_objects.NewPaymentMethod(manualTx.Method)

		transaction := entities.NewTransaction(
			generateTransactionID("manual", i+1),
			manualTx.TransactionDate,
			manualTx.TransactionDate,
			transactionType,
			manualTx.Category,
			paymentMethod,
			utils.RoundToCents(manualTx.Amount),
		)
		transaction.SetManual(true)
		transactions = append(transactions, transaction)
	}
	return transactions
}

// selectPaymentMethod выбирает случайный метод платежа для B2B [6]
func (s *generatorService) selectPaymentMethod() value_objects.PaymentMethod {
	// 70% ACH Credit, 30% Electronic Payment
	if rand.Float64() < 0.7 {
		return value_objects.ACHCredit
	}
	return value_objects.ElectronicPayment
}

// getUserID конвертирует userID из строки в UUID
func (s *generatorService) getUserID(userIDStr *string) *uuid.UUID {
	if userIDStr == nil || *userIDStr == "" {
		return nil
	}

	userID, err := uuid.Parse(*userIDStr)
	if err != nil {
		return nil
	}

	return &userID
}

// selectGateway выбирает случайный шлюз [35]
func (s *generatorService) selectGateway() *entities.Gateway {
	if len(s.gateways) == 0 {
		return entities.NewGateway("gw_default", "Stripe Gateway")
	}
	return s.gateways[rand.Intn(len(s.gateways))]
}

// getIRSOccurrences возвращает количество транзакций IRS налогов [23][24]
// 1 транзакция в обычные месяцы, 2 транзакции в квартальные месяцы (январь, апрель, июнь, сентябрь)
func (s *generatorService) getIRSOccurrences(template *entities.TransactionTemplate, year, month int) int {
	// Проверяем, является ли месяц квартальным
	// Квартальные месяцы: январь (1), апрель (4), июнь (6), сентябрь (9)
	if s.dateCalculator.isQuarterlyMonth(month) {
		return 2 // Квартальный месяц - 2 транзакции
	}
	return 1 // Обычный месяц - 1 транзакция
}

// separateTemplates разделяет шаблоны на обязательные и опциональные [39-41]
func (s *generatorService) separateTemplates() ([]*entities.TransactionTemplate, []*entities.TransactionTemplate) {
	var mandatory []*entities.TransactionTemplate
	var optional []*entities.TransactionTemplate

	for _, template := range s.templates {
		// [25][14] "Подписка ПО" всегда обязательная, даже если помечена как optional
		if template.Category == "Подписка ПО" {
			mandatory = append(mandatory, template)
		} else if template.IsOptional {
			optional = append(optional, template)
		} else {
			mandatory = append(mandatory, template)
		}
	}

	return mandatory, optional
}

// shuffleTemplates перемешивает шаблоны для случайного выбора [41]
func (s *generatorService) shuffleTemplates(templates []*entities.TransactionTemplate) []*entities.TransactionTemplate {
	shuffled := make([]*entities.TransactionTemplate, len(templates))
	copy(shuffled, templates)

	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	return shuffled
}

// createCustomCustomers создает кастомных клиентов из списка имен [6]
func (s *generatorService) createCustomCustomers(names []string) []*entities.Customer {
	var customers []*entities.Customer
	for i, name := range names {
		customers = append(customers, entities.NewCustomer(
			fmt.Sprintf("custom_%d", i+1),
			name,
			"custom",
			0.055, // 5.5%
			0.085, // 8.5%
			2,     // min transactions
			8,     // max transactions
		))
	}
	return customers
}

// isFirstMonth проверяет, является ли это первым месяцем
func (s *generatorService) isFirstMonth() bool {
	// Упрощенная реализация - можно доработать с учетом истории
	return s.amountCalculator.isFirstMonthFlag
}

// getCurrentTurnover возвращает текущий оборот
func (s *generatorService) getCurrentTurnover() float64 {
	return s.amountCalculator.firstMonthTurnover
}

// generateTransactionID генерирует ID транзакции
func generateTransactionID(prefix string, num int) string {
	return fmt.Sprintf("t_%s_%03d", prefix, num)
}

// generateTemplateTransactionID генерирует ID транзакции из шаблона
func generateTemplateTransactionID(category string, num int) string {
	return fmt.Sprintf("t_exp_%s_%03d", category, num)
}

// sortTransactionsByDate сортирует транзакции по дате
func (s *generatorService) sortTransactionsByDate(transactions []*entities.Transaction) []*entities.Transaction {
	// Используем sort.Slice для эффективной сортировки O(n log n)
	sort.Slice(transactions, func(i, j int) bool {
		// Сортируем по TransactionDate (время совершения транзакции)
		// Если даты одинаковые, используем ID для стабильной сортировки
		if transactions[i].TransactionDate.Equal(transactions[j].TransactionDate) {
			return transactions[i].ID < transactions[j].ID
		}
		return transactions[i].TransactionDate.Before(transactions[j].TransactionDate)
	})
	return transactions
}

// balanceAndNormalize балансирует и нормализует суммы [42]
func (s *generatorService) balanceAndNormalize(transactions []*entities.Transaction, turnover, targetProfit float64) ([]*entities.Transaction, error) {
	// Округляем все суммы
	for _, tx := range transactions {
		tx.Amount = utils.RoundToCents(tx.Amount)
	}

	// Рассчитываем текущие итоги
	currentIncome := 0.0
	currentExpense := 0.0

	for _, tx := range transactions {
		if tx.IsIncome() {
			currentIncome += tx.Amount
		} else {
			currentExpense += tx.Amount
		}
	}

	currentProfit := currentIncome + currentExpense // расходы отрицательные

	// Выравниваем итоги
	incomeDiff := turnover - currentIncome
	profitDiff := targetProfit - currentProfit

	// [42] Корректируем доходы: сумма всех доходов должна равняться turnover
	if math.Abs(incomeDiff) > 0.01 {
		// Находим последнюю транзакцию дохода (не ручную)
		for i := len(transactions) - 1; i >= 0; i-- {
			if transactions[i].IsIncome() && !transactions[i].IsManual {
				transactions[i].Amount += incomeDiff
				transactions[i].Amount = utils.RoundToCents(transactions[i].Amount)
				break
			}
		}
	}

	// [42] Корректируем расходы: прибыль должна равняться targetProfit
	// profitDiff = targetProfit - currentProfit
	// Если profitDiff > 0, нужно увеличить прибыль (уменьшить расходы)
	// Если profitDiff < 0, нужно уменьшить прибыль (увеличить расходы)
	if math.Abs(profitDiff) > 0.01 {
		// Если разница большая (> 1% от оборота), распределяем корректировку по нескольким транзакциям
		// Иначе корректируем только последнюю транзакцию (для небольших погрешностей округления)
		if math.Abs(profitDiff) > turnover*0.01 {
			// Большая разница - распределяем по нескольким транзакциям расходов (не ручным)
			expenseTransactions := make([]*entities.Transaction, 0)
			for _, tx := range transactions {
				if tx.IsExpense() && !tx.IsManual {
					expenseTransactions = append(expenseTransactions, tx)
				}
			}

			if len(expenseTransactions) > 0 {
				// Распределяем корректировку пропорционально абсолютным значениям расходов
				totalExpenseAbs := 0.0
				for _, tx := range expenseTransactions {
					totalExpenseAbs += math.Abs(tx.Amount)
				}

				if totalExpenseAbs > 0 {
					// Распределяем profitDiff пропорционально по транзакциям
					// profitDiff > 0 означает, что нужно увеличить прибыль (уменьшить расходы)
					// profitDiff < 0 означает, что нужно уменьшить прибыль (увеличить расходы)
					remainingDiff := profitDiff

					// Распределяем по всем транзакциям, начиная с последней
					for i := len(expenseTransactions) - 1; i >= 0 && math.Abs(remainingDiff) > 0.01; i-- {
						proportion := math.Abs(expenseTransactions[i].Amount) / totalExpenseAbs
						adjustment := remainingDiff * proportion

						// Ограничиваем корректировку, чтобы не сделать транзакцию положительной
						// Если adjustment > 0 (уменьшаем расход), проверяем, что транзакция останется отрицательной
						// Если adjustment < 0 (увеличиваем расход), это всегда безопасно
						if adjustment > 0 && expenseTransactions[i].Amount-adjustment > 0 {
							// Ограничиваем adjustment, чтобы транзакция осталась отрицательной или нулем
							adjustment = expenseTransactions[i].Amount
						}

						// Корректируем: amount -= adjustment
						// Если adjustment > 0, уменьшаем расход (делаем менее отрицательным)
						// Если adjustment < 0, увеличиваем расход (делаем более отрицательным)
						expenseTransactions[i].Amount -= adjustment
						expenseTransactions[i].Amount = utils.RoundToCents(expenseTransactions[i].Amount)
						remainingDiff -= adjustment
					}

					// Если осталась небольшая разница, корректируем последнюю транзакцию
					if math.Abs(remainingDiff) > 0.01 && len(expenseTransactions) > 0 {
						lastTx := expenseTransactions[len(expenseTransactions)-1]
						// Ограничиваем, чтобы не сделать транзакцию положительной
						// Если remainingDiff > 0 (уменьшаем расход), проверяем, что транзакция останется отрицательной
						if remainingDiff > 0 && lastTx.Amount-remainingDiff > 0 {
							remainingDiff = lastTx.Amount
						}
						lastTx.Amount -= remainingDiff
						lastTx.Amount = utils.RoundToCents(lastTx.Amount)
					}
				}
			}
		} else {
			// Небольшая разница - корректируем только последнюю транзакцию (для погрешностей округления)
			for i := len(transactions) - 1; i >= 0; i-- {
				if transactions[i].IsExpense() && !transactions[i].IsManual {
					// Корректируем расход: amount -= profitDiff
					// Если profitDiff > 0 (нужно увеличить прибыль), уменьшаем расход (делаем менее отрицательным)
					// Если profitDiff < 0 (нужно уменьшить прибыль), увеличиваем расход (делаем более отрицательным)
					transactions[i].Amount -= profitDiff
					transactions[i].Amount = utils.RoundToCents(transactions[i].Amount)
					break
				}
			}
		}
	}

	// Финальная проверка и корректировка: пересчитываем итоги после корректировки
	// [42] Гарантируем точное соответствие целевым значениям согласно ТЗ
	maxIterations := 5
	for iteration := 0; iteration < maxIterations; iteration++ {
		finalIncome := 0.0
		finalExpense := 0.0
		for _, tx := range transactions {
			if tx.IsIncome() {
				finalIncome += tx.Amount
			} else {
				finalExpense += tx.Amount
			}
		}
		finalProfit := finalIncome + finalExpense

		// Проверяем, что итоги соответствуют целевым значениям (с допустимой погрешностью округления)
		incomeError := math.Abs(turnover - finalIncome)
		profitError := math.Abs(targetProfit - finalProfit)

		// Если погрешности в пределах допустимого (0.02 для округления), выходим
		if incomeError <= 0.02 && profitError <= 0.02 {
			break
		}

		// [42] Корректируем доходы: сумма всех доходов должна равняться turnover
		if incomeError > 0.02 {
			log.Printf("[DEBUG] Income normalization iteration %d: target=%.2f, actual=%.2f, diff=%.2f", iteration+1, turnover, finalIncome, incomeError)
			// Находим последнюю транзакцию дохода (не ручную)
			for i := len(transactions) - 1; i >= 0; i-- {
				if transactions[i].IsIncome() && !transactions[i].IsManual {
					incomeDiff := turnover - finalIncome
					transactions[i].Amount += incomeDiff
					transactions[i].Amount = utils.RoundToCents(transactions[i].Amount)
					break
				}
			}
		}

		// [42] Корректируем расходы: прибыль должна равняться targetProfit
		if profitError > 0.02 {
			log.Printf("[DEBUG] Profit normalization iteration %d: target=%.2f, actual=%.2f, diff=%.2f", iteration+1, targetProfit, finalProfit, profitError)
			// Находим последнюю транзакцию расхода (не ручную)
			for i := len(transactions) - 1; i >= 0; i-- {
				if transactions[i].IsExpense() && !transactions[i].IsManual {
					profitDiff := targetProfit - finalProfit
					// Ограничиваем корректировку, чтобы не сделать транзакцию положительной
					if profitDiff > 0 && transactions[i].Amount-profitDiff > 0 {
						profitDiff = transactions[i].Amount
					}
					transactions[i].Amount -= profitDiff
					transactions[i].Amount = utils.RoundToCents(transactions[i].Amount)
					break
				}
			}
		}
	}

	// Финальная проверка после всех итераций
	finalIncome := 0.0
	finalExpense := 0.0
	for _, tx := range transactions {
		if tx.IsIncome() {
			finalIncome += tx.Amount
		} else {
			finalExpense += tx.Amount
		}
	}
	finalProfit := finalIncome + finalExpense

	incomeError := math.Abs(turnover - finalIncome)
	profitError := math.Abs(targetProfit - finalProfit)

	if incomeError > 0.05 {
		log.Printf("[WARN] Income normalization final error: target=%.2f, actual=%.2f, diff=%.2f", turnover, finalIncome, incomeError)
	}
	if profitError > 0.05 {
		log.Printf("[WARN] Profit normalization final error: target=%.2f, actual=%.2f, diff=%.2f", targetProfit, finalProfit, profitError)
	}

	return transactions, nil
}

// calculateBalances рассчитывает балансы после каждой транзакции
func (s *generatorService) calculateBalances(transactions []*entities.Transaction, initialBalance float64) ([]*entities.Transaction, error) {
	currentBalance := initialBalance

	for _, tx := range transactions {
		// Проверяем, достаточно ли средств
		if tx.IsExpense() && currentBalance+tx.Amount < 0 {
			return nil, fmt.Errorf("insufficient balance on %s: required %.2f, available %.2f",
				tx.TransactionDate.Format("2006-01-02"),
				-tx.Amount, currentBalance)
		}

		currentBalance += tx.Amount
		tx.SetBalanceAfter(utils.RoundToCents(currentBalance))
	}

	return transactions, nil
}

// recalculateBalances - пересчет балансов после корректировки транзакций
func (s *generatorService) recalculateBalances(transactions []*entities.Transaction, initialBalance float64) ([]*entities.Transaction, error) {
	currentBalance := initialBalance

	for _, tx := range transactions {
		currentBalance += tx.Amount
		tx.SetBalanceAfter(utils.RoundToCents(currentBalance))
	}

	return transactions, nil
}

// checkNegativeBalance проверяет отрицательный баланс [43]
func (s *generatorService) checkNegativeBalance(transactions []*entities.Transaction) error {
	for _, tx := range transactions {
		if tx.BalanceAfter < 0 {
			return ErrNegativeBalance
		}
	}
	return nil
}

// buildResponse формирует ответ с транзакциями [44-54]
func (s *generatorService) buildResponse(transactions []*entities.Transaction, req *dto.GenerateRequest, requestID uuid.UUID, userID uuid.UUID) *dto.GenerateResponse {
	// Рассчитываем финансовую сводку
	finalBalance := req.InitialBalance
	totalRevenue := 0.0
	totalExpenses := 0.0

	for _, tx := range transactions {
		if tx.IsIncome() {
			totalRevenue += tx.Amount
		} else {
			totalExpenses += tx.Amount
		}
	}

	if len(transactions) > 0 {
		finalBalance = transactions[len(transactions)-1].BalanceAfter
	}

	// [42][45] Рассчитываем netProfit согласно ТЗ: сумма всех доходов минус сумма всех расходов
	// totalExpenses отрицательное, поэтому netProfit = totalRevenue + totalExpenses
	netProfit := totalRevenue + totalExpenses

	// [5][6] Проверка соответствия ТЗ: суммарные доходы должны быть = 100% оборота
	// Учитываем ручные доходы, если они есть
	expectedRevenue := req.Turnover
	manualIncomeAmount := 0.0
	for _, tx := range transactions {
		if tx.IsIncome() && tx.IsManual {
			manualIncomeAmount += tx.Amount
		}
	}
	if manualIncomeAmount > 0 {
		expectedRevenue = req.Turnover + manualIncomeAmount
	}

	// [1][5] Проверка соответствия ТЗ: прибыль должна соответствовать desiredProfitPercent
	// Учитываем ручные транзакции
	expectedProfit := req.Turnover * (req.DesiredProfitPercent / 100)
	manualExpenseAmount := 0.0
	for _, tx := range transactions {
		if tx.IsExpense() && tx.IsManual {
			manualExpenseAmount += tx.Amount // отрицательное
		}
	}
	if manualIncomeAmount > 0 || manualExpenseAmount < 0 {
		expectedProfit = expectedProfit + manualIncomeAmount + manualExpenseAmount
	}

	// Проверка и логирование отклонений (допустимая погрешность округления 0.05)
	revenueError := math.Abs(expectedRevenue - totalRevenue)
	profitError := math.Abs(expectedProfit - netProfit)

	if revenueError > 0.05 {
		log.Printf("[WARN] Revenue mismatch: expected=%.2f (turnover=%.2f + manualIncome=%.2f), actual=%.2f, diff=%.2f",
			expectedRevenue, req.Turnover, manualIncomeAmount, totalRevenue, revenueError)
	}
	if profitError > 0.05 {
		log.Printf("[WARN] Profit mismatch: expected=%.2f (%.2f%% of %.2f + manual adjustments), actual=%.2f, diff=%.2f",
			expectedProfit, req.DesiredProfitPercent, req.Turnover, netProfit, profitError)
	}

	summary := dto.FinancialSummary{
		InitialBalance: utils.RoundToCents(req.InitialBalance),
		FinalBalance:   utils.RoundToCents(finalBalance),
		TotalRevenue:   utils.RoundToCents(totalRevenue),
		TotalExpenses:  utils.RoundToCents(totalExpenses),
		NetProfit:      utils.RoundToCents(netProfit),
	}

	// Рассчитываем ежедневные балансы
	dailyBalances := s.calculateDailyBalances(transactions, req.InitialBalance, req.Year, req.Month)

	// Конвертируем entities.Transaction в dto.Transaction
	dtoTransactions := make([]dto.Transaction, len(transactions))
	for i, tx := range transactions {
		dtoTransactions[i] = dto.Transaction{
			TransactionID:      tx.ID,
			TransactionDate:    tx.TransactionDate,
			PostingDate:        tx.PostingDate.Format("2006-01-02"),
			Type:               tx.Type.String(),
			Category:           tx.Category,
			Method:             tx.Method.String(),
			Amount:             tx.Amount,
			BalanceAfter:       tx.BalanceAfter,
			IsManual:           tx.IsManual,
			FixAsFirst:         tx.FixAsFirst,
			CalculationDetails: tx.CalculationDetails,
		}
	}

	// Получаем associatedCard из таблицы users [51][52]
	// Номер карты должен быть задан пользователем через API /api/user/associated-card
	associatedCard := s.getAssociatedCard(userID)

	// Формируем forwardingInfo
	forwardingInfo := dto.ForwardingInfo{
		AssociatedCard:    associatedCard,
		OwnerName:         "",
		CompanyName:       "",
		CustomCustomers:   []string{},
		CustomContractors: []dto.CustomContractor{},
	}

	if req.CustomData != nil {
		forwardingInfo.OwnerName = req.CustomData.CompanyInfo.OwnerName
		forwardingInfo.CompanyName = req.CustomData.CompanyInfo.CompanyName
		forwardingInfo.CustomCustomers = req.CustomData.CustomCustomers
		forwardingInfo.CustomContractors = req.CustomData.CustomContractors
	}

	// Рассчитываем разбивку по методам платежа
	revenueBreakdown := s.breakdownService.CalculateRevenueBreakdown(transactions)
	expensesBreakdown := s.breakdownService.CalculateExpensesBreakdown(transactions)
	transactionCounts := s.breakdownService.CalculateTransactionCounts(transactions)

	return &dto.GenerateResponse{
		RequestID:            requestID.String(),
		Transactions:         dtoTransactions,
		FinancialSummary:     summary,
		DailyClosingBalances: dailyBalances,
		ForwardingInfo:       forwardingInfo,
		RevenueBreakdown:     revenueBreakdown,
		ExpensesBreakdown:    expensesBreakdown,
		TransactionCounts:    transactionCounts,
	}
}

// calculateDailyBalances рассчитывает ежедневные балансы [50]
// Генерирует балансы для всех дней месяца, а не только для дней с транзакциями
func (s *generatorService) calculateDailyBalances(transactions []*entities.Transaction, initialBalance float64, year, month int) []dto.DailyBalance {
	// Собираем балансы по датам из транзакций
	// Для каждой даты берем баланс после последней транзакции этого дня
	balancesByDate := make(map[string]float64)

	// Начальный баланс на первый день месяца
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	balancesByDate[firstDay.Format("2006-01-02")] = initialBalance

	// Обновляем балансы после каждой транзакции
	// Если в один день несколько транзакций, последняя перезапишет баланс
	for _, tx := range transactions {
		dateKey := tx.PostingDate.Format("2006-01-02")
		balancesByDate[dateKey] = tx.BalanceAfter
	}

	// Получаем последний день месяца
	lastDayOfMonth := time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.UTC)
	daysInMonth := lastDayOfMonth.Day()

	// Генерируем балансы для всех дней месяца
	var dailyBalances []dto.DailyBalance
	currentBalance := initialBalance

	for day := 1; day <= daysInMonth; day++ {
		currentDate := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		dateKey := currentDate.Format("2006-01-02")

		// Если есть транзакции в этот день, используем баланс после последней транзакции
		if balance, exists := balancesByDate[dateKey]; exists {
			currentBalance = balance
		}
		// Если транзакций нет, используем баланс предыдущего дня (currentBalance уже содержит его)

		dailyBalances = append(dailyBalances, dto.DailyBalance{
			Date:    dateKey,
			Balance: utils.RoundToCents(currentBalance),
		})
	}

	return dailyBalances
}

// markFirstTransactionsByCategory помечает первые транзакции каждой категории флагом FixAsFirst [168]
// Транзакции должны быть отсортированы по transactionDate перед вызовом этой функции
func (s *generatorService) markFirstTransactionsByCategory(transactions []*entities.Transaction) {
	// Используем map для отслеживания первой транзакции каждой категории
	firstSeenCategory := make(map[string]bool)

	for _, tx := range transactions {
		// Если это первая транзакция данной категории, помечаем её
		if !firstSeenCategory[tx.Category] {
			tx.SetFixAsFirst(true)
			firstSeenCategory[tx.Category] = true
		} else {
			tx.SetFixAsFirst(false)
		}
	}
}

// getAssociatedCard получает сохраненный номер карты из таблицы users [51][52]
// Номер карты должен быть задан пользователем через API /api/user/associated-card
// userID должен быть валидным (не uuid.Nil), проверка выполняется в GenerateTransactions
func (s *generatorService) getAssociatedCard(userID uuid.UUID) string {
	// Получаем пользователя из таблицы users
	userModel, err := s.userRepo.GetByID(userID)
	if err != nil {
		log.Printf("[WARN] Failed to get user %v: %v, associated card not available", userID, err)
		return ""
	}

	// Если у пользователя есть сохраненный номер карты, возвращаем его
	if userModel.AssociatedCard != nil && *userModel.AssociatedCard != "" {
		return *userModel.AssociatedCard
	}

	// Если номер карты не найден, возвращаем пустую строку
	// Пользователь должен задать номер карты через API /api/user/associated-card
	log.Printf("[WARN] User %v does not have associated card set. Please set it via /api/user/associated-card endpoint", userID)
	return ""
}

// scaleTransactions масштабирует транзакции [38]
func (s *generatorService) scaleTransactions(transactions []*entities.Transaction, scaleFactor int) []*entities.Transaction {
	var scaled []*entities.Transaction

	for _, tx := range transactions {
		// Не масштабируем Payroll ADP и фиксированные 1 раз в месяц
		if tx.Category == "Payroll" || tx.IsManual {
			scaled = append(scaled, tx)
			continue
		}

		// Масштабируем количество, уменьшая сумму
		for i := 0; i < scaleFactor; i++ {
			scaledTx := *tx
			scaledTx.Amount = utils.RoundToCents(tx.Amount / float64(scaleFactor))
			scaledTx.ID = fmt.Sprintf("%s_%d", tx.ID, i+1)
			scaled = append(scaled, &scaledTx)
		}
	}

	return scaled
}

// convertToDomainTransactions конвертирует entities.Transaction в domain.GeneratedTransaction для сохранения в БД
func (s *generatorService) convertToDomainTransactions(transactions []*entities.Transaction, requestID uuid.UUID) []*domain.GeneratedTransaction {
	domainTransactions := make([]*domain.GeneratedTransaction, len(transactions))

	for i, tx := range transactions {
		balanceAfter := tx.BalanceAfter

		// Определяем sort_order на основе индекса
		sortOrder := i + 1

		domainTx := &domain.GeneratedTransaction{
			ID:                 uuid.New(),
			RequestID:          requestID,
			TransactionID:      tx.ID,
			TransactionDate:    tx.TransactionDate,
			PostingDate:        tx.PostingDate,
			Type:               tx.Type.String(),
			Category:           tx.Category,
			Method:             tx.Method.String(),
			Amount:             tx.Amount,
			BalanceAfter:       &balanceAfter,
			IsManual:           tx.IsManual,
			SortOrder:          &sortOrder,
			CalculationDetails: tx.CalculationDetails, // Сохраняем calculation_details
		}

		domainTransactions[i] = domainTx
	}

	return domainTransactions
}
