// internal/service/generator.go
package generatorservice

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/internal/domain"
	"github.com/IbadT/business_bank_back/services/matematika/internal/domain/entities"
	"github.com/IbadT/business_bank_back/services/matematika/internal/domain/value_objects"
	"github.com/IbadT/business_bank_back/services/matematika/internal/repository"
	balanceservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/balance"
	baseamountservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/base"
	breakdownservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/breakdown"
	dateservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/date"
	gatewayservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/gateway"
	holidayservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/holiday"
	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/dto"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/helpers"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/utils"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/validator"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

var (
	ErrInvalidModel   = errors.New("invalid business model")
	ErrInvalidRequest = errors.New("invalid request parameters")
	ErrUnauthorized   = errors.New("user authentication required") // Требуется авторизация
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
	dateCalculator           *dateservice.DateCalculator
	holidayService           holidayservice.HolidayService
	gatewayService           gatewayservice.GatewayService
	baseAmountService        baseamountservice.BaseAmountService
	breakdownService         breakdownservice.BreakdownService
	balanceAdjustmentService balanceservice.BalanceAdjustmentService
	expenseGenerator         *expenseGenerator
	fixedAmountCalculator    *fixedAmountCalculator
	templates                []*entities.TransactionTemplate
	gateways                 []*entities.Gateway
	customers                []*entities.Customer
}

func NewGeneratorService(configRepo repository.ConfigRepository, stateRepo repository.StateRepository, userRepo repository.UserRepository, holidayRepo repository.HolidayRepository, gatewayRepo repository.GatewayRepository, holidayService holidayservice.HolidayService, gatewayService gatewayservice.GatewayService, baseAmountService baseamountservice.BaseAmountService, breakdownService breakdownservice.BreakdownService, balanceAdjustmentService balanceservice.BalanceAdjustmentService, generationRequestRepo repository.GenerationRequestRepository, transactionRepo repository.TransactionRepository) (GeneratorService, error) {
	// Загружаем доменные сущности параллельно
	var holidays []*domain.Holiday
	var templates []*entities.TransactionTemplate
	var gateways []*entities.Gateway
	var customers []*entities.Customer

	g, _ := errgroup.WithContext(context.Background())

	g.Go(func() error {
		var err error
		holidays, err = configRepo.GetHolidays()
		return err
	})

	g.Go(func() error {
		var err error
		templates, err = configRepo.GetTransactionTemplates()
		return err
	})

	g.Go(func() error {
		var err error
		gateways, err = configRepo.GetGateways()
		return err
	})

	g.Go(func() error {
		var err error
		customers, err = configRepo.GetCustomers()
		return err
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	dateCalc := dateservice.NewDateCalculator(holidays, stateRepo, holidayService, baseAmountService, generationRequestRepo)

	fixedAmountCalc := newFixedAmountCalculator(baseAmountService, dateCalc)
	
	service := &generatorService{
		configRepo:               configRepo,
		stateRepo:                stateRepo,
		userRepo:                 userRepo,
		holidayRepo:              holidayRepo,
		gatewayRepo:              gatewayRepo,
		generationRequestRepo:    generationRequestRepo,
		transactionRepo:          transactionRepo,
		dateCalculator:           dateCalc,
		holidayService:           holidayService,
		gatewayService:           gatewayService,
		baseAmountService:        baseAmountService,
		breakdownService:         breakdownService,
		balanceAdjustmentService: balanceAdjustmentService,
		fixedAmountCalculator:     fixedAmountCalc,
		expenseGenerator:         newExpenseGenerator(dateCalc, holidayService, fixedAmountCalc),
		templates:                templates,
		gateways:                 gateways,
		customers:                customers,
	}
	
	return service, nil
}

func (s *generatorService) GenerateTransactions(req *dto.GenerateRequest, userID *string) (*dto.GenerateResponse, error) {
	// [1] Валидация запроса
	if err := validator.ValidateGenerateRequest(req); err != nil {
		return nil, err
	}

	// Проверка авторизации - userID обязателен
	if userID == nil || *userID == "" {
		return nil, ErrUnauthorized
	}

	// Создаем GenerationRequest в БД перед генерацией
	userIDUUID := utils.ParseUserID(userID)
	if userIDUUID == nil {
		return nil, fmt.Errorf("invalid user ID")
	}

	monthStr := utils.FormatMonth(req.Year, req.Month)
	customData := s.convertCustomDataToJSONB(req.CustomData)

	// Используем фабрику для создания GenerationRequest
	requestFactory := domain.NewGenerationRequestFactory()
	generationRequest := requestFactory.Create(
		userIDUUID,
		monthStr,
		req.Year,
		req.Turnover,
		req.DesiredProfitPercent,
		req.Model,
		req.InitialBalance,
		req.ScaleFactor,
		customData,
	)

	createdRequest, err := s.generationRequestRepo.Create(generationRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to create generation request: %w", err)
	}

	requestID := createdRequest.ID

	// [1][5] Рассчет целевой прибыли используя CalculateNumericalCore из patterns.go
	numericalCore := helpers.CalculateNumericalCore(req.Turnover, req.DesiredProfitPercent)
	targetProfit := numericalCore.TargetProfit
	totalExpensesTarget := numericalCore.TotalExpensesTarget

	// Получаем правила для последовательных выписок из patterns.go (README строка 74-78)
	// Эти правила используются для:
	// - 60% ± 10% клиентов повторяются между месяцами (50-70%)
	// - 70% ± 10% подрядчиков повторяются между месяцами (60-80%)
	// - Фиксированные контрагенты (Warehouse rent, ERP, Insurance) неизменны
	// - Даты смещаются, шаблоны дней недели сохраняются
	// - При ручных операциях доли пересчитываются для чистой прибыли 4-8%
	// Правила используются при генерации B2B клиентов и расходов для обеспечения повторяемости
	// между месяцами согласно требованиям README
	_ = helpers.GetSequentialStatementsRules() // Получаем правила для будущего использования при реализации логики повторяемости

	// [2][35][36] Генерация доходов в зависимости от модели
	var incomeTransactions []*entities.Transaction

	incomeTransactions, err = s.generateIncomes(req, userID)
	if err != nil {
		s.updateRequestStatusOnError(requestID, err)
		return nil, err
	}

	// [7-21] Генерация расходов
	expenseTransactions, err := s.generateExpenses(req, totalExpensesTarget, userID)
	if err != nil {
		s.updateRequestStatusOnError(requestID, err)
		return nil, err
	}

	// [38] Объединение транзакций
	allTransactions := append(incomeTransactions, expenseTransactions...)

	// [38] Добавление ручных транзакций
	allTransactions, manualIncomeAmount, manualExpenseAmount := s.addManualTransactions(allTransactions, req.CustomData)

	// [38] При добавлении ручных операций удаляем мелкие опциональные транзакции, чтобы не выйти за пределы 39-75
	// [12][40] Удаляем мелкие транзакции из опциональных категорий при необходимости
	if req.CustomData != nil && len(req.CustomData.ManualTransactions) > 0 {
		optionalCategories := s.getOptionalCategoriesMap()
		allTransactions = s.removeSmallOptionalTransactions(allTransactions, optionalCategories, req.Model)
	}

	// [2][3][4] Валидация количества транзакций согласно требованиям README
	// Валидация выполняется ДО масштабирования, так как "Эти диапазоны могут масштабироваться при необходимости"
	if req.ScaleFactor <= 1 {
		// Валидируем только если нет масштабирования
		if err := validator.ValidateTransactionCounts(allTransactions, req.Model); err != nil {
			s.updateRequestStatusOnError(requestID, err)
			return nil, err
		}
	} else {
		// При масштабировании только логируем исходное количество
		var incomeCount, expenseCount int
		for _, tx := range allTransactions {
			if tx.IsIncome() {
				incomeCount++
			} else {
				expenseCount++
			}
		}
		logrus.Infof("[INFO] ScaleFactor=%d, skipping transaction count validation (ranges can be scaled). Original counts: income=%d, expense=%d, total=%d",
			req.ScaleFactor, incomeCount, expenseCount, len(allTransactions))
	}

	// [38] Масштабирование
	if req.ScaleFactor > 1 {
		allTransactions = s.scaleTransactions(allTransactions, req.ScaleFactor)
	}

	// Сортировка по дате
	allTransactions = helpers.SortTransactionsByDate(allTransactions)

	// [42] Балансировка и нормализация
	// Скорректируем targetProfit с учетом ручных транзакций
	// Ручные доходы увеличивают оборот, ручные расходы уменьшают прибыль
	adjustedTurnover := req.Turnover + manualIncomeAmount
	adjustedTargetProfit := targetProfit + manualIncomeAmount + manualExpenseAmount // manualExpenseAmount отрицательный

	// [12][40] Получаем список опциональных категорий для приоритетного удаления
	optionalCategories := s.getOptionalCategoriesMap()
	
	balancedTransactions, err := s.balanceAdjustmentService.BalanceAndNormalize(allTransactions, adjustedTurnover, adjustedTargetProfit, optionalCategories)
	if err != nil {
		s.updateRequestStatusOnError(requestID, err)
		return nil, err
	}

	// Расчет и корректировка балансов
	transactionsWithBalance, err := s.calculateAndAdjustBalances(balancedTransactions, req.InitialBalance, req.Year, req.Month, requestID)
	if err != nil {
		return nil, err
	}

	// TODO: нужно ли это, если в balance_adjustment_service уже есть сортировка ?????
	// Финальная сортировка перед сохранением и формированием ответа
	// Гарантируем, что транзакции отсортированы по transactionDate
	transactionsWithBalance = helpers.SortTransactionsByDate(transactionsWithBalance)

	// [168] Помечаем первые транзакции каждой категории флагом FixAsFirst
	s.markFirstTransactionsByCategory(transactionsWithBalance)

	// Конвертируем entities.Transaction в domain.GeneratedTransaction для сохранения в БД
	domainTransactions := s.convertToDomainTransactions(transactionsWithBalance, requestID)

	// Сохраняем транзакции и обновляем статус
	s.saveTransactionsAndUpdateStatus(domainTransactions, requestID)

	// Формирование ответа
	// userIDUUID гарантированно не nil, так как проверено выше
	return s.buildResponse(transactionsWithBalance, req, requestID, *userIDUUID), nil
}


func (s *generatorService) generateB2CIncomes(req *dto.GenerateRequest, userID *string) ([]*entities.Transaction, error) {
	// Используем CalculateB2CReplenishment из patterns.go для расчета B2C пополнений
	b2cData := helpers.CalculateB2CReplenishment(req.Turnover, req.Year, req.Month)

	// [35] Получаем или выбираем шлюз (сохраняем для всех месяцев)
	var gateway *entities.Gateway
	var err error

	if userID != nil {
		// Парсим userID из string в UUID
		userUUID := utils.ParseUserID(userID)
		if userUUID != nil {
			// Используем GatewayService для получения сохраненного шлюза
			gateway, err = s.gatewayService.GetB2CGateways(*userUUID)
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
			userUUID := utils.ParseUserID(userID)
			if userUUID != nil {
				// Сохраняем через сервис (передаем ID выбранного шлюза)
				if err := s.gatewayService.SaveB2CGateways(*userUUID, gateway.ID); err != nil {
					// Логируем ошибку, но не прерываем генерацию
					// Шлюз уже выбран, генерация может продолжиться
					logrus.Infof("[WARN] Failed to save gateway via GatewayService: %v", err)
				}
			}
		}
	}

	var transactions []*entities.Transaction

	// Конвертируем B2CTransactionData в entities.Transaction
	for i, data := range b2cData {
		// [32] Проверка праздников для B2C пополнений (ACH Credit - операция по счету)
		transactionDate := data.Date
		if s.holidayService.IsHoliday(transactionDate) {
			transactionDate = s.holidayService.GetNextBusinessDay(transactionDate)
		}

		// [33] Генерация времени в рабочие часы (08:00-18:00)
		transactionTime := s.dateCalculator.GenerateBusinessTime(transactionDate, 8, 18)

		// Создание доменной сущности
		transaction := entities.NewTransaction(
			utils.GenerateTransactionID("inc", i+1),
			transactionTime,
			transactionDate,
			value_objects.Income,
			gateway.Name,
			value_objects.ACHCredit,
			data.Amount,
		)
		transaction.SetBalanceAfter(0) // Будет рассчитано позже

		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

func (s *generatorService) generateB2BIncomes(req *dto.GenerateRequest) []*entities.Transaction {
	// Получаем правила для последовательных выписок (README строка 74-78)
	// Используется для обеспечения 60% ± 10% повторяемости клиентов между месяцами
	sequentialRules := helpers.GetSequentialStatementsRules()
	_ = sequentialRules // TODO: Реализовать логику повторяемости клиентов между месяцами

	// Используем кастомных клиентов если есть [6]
	customers := s.customers
	if req.CustomData != nil && len(req.CustomData.CustomCustomers) > 0 {
		customers = s.createCustomCustomers(req.CustomData.CustomCustomers)
	}

	// Группируем customers по категориям для создания B2BCategoryConfig
	customersByCategory := make(map[string][]*entities.Customer)
	for _, customer := range customers {
		category := customer.Category
		if category == "" {
			category = "additional" // Категория по умолчанию
		}
		customersByCategory[category] = append(customersByCategory[category], customer)
	}

	// Создаем B2BCategoryConfig для каждой категории
	var categories []helpers.B2BCategoryConfig
	for categoryName, categoryCustomers := range customersByCategory {
		// Определяем диапазоны на основе первого клиента категории (все клиенты в категории должны иметь одинаковые диапазоны)
		if len(categoryCustomers) == 0 {
			continue
		}
		firstCustomer := categoryCustomers[0]

		// Преобразуем имена клиентов в строки
		customerNames := make([]string, len(categoryCustomers))
		for i, c := range categoryCustomers {
			customerNames[i] = c.Name
		}

		// Определяем min/max платежей на основе TransactionRange клиентов
		// Для категории используем диапазон из первого клиента (категория должна иметь свой диапазон)
		// Если в категории несколько клиентов, берем максимальные значения для категории
		minPayments := firstCustomer.TransactionRange.Min
		maxPayments := firstCustomer.TransactionRange.Max
		for i := 1; i < len(categoryCustomers); i++ {
			if categoryCustomers[i].TransactionRange.Max > maxPayments {
				maxPayments = categoryCustomers[i].TransactionRange.Max
			}
			if categoryCustomers[i].TransactionRange.Min < minPayments {
				minPayments = categoryCustomers[i].TransactionRange.Min
			}
		}

		categories = append(categories, helpers.B2BCategoryConfig{
			CategoryName:  categoryName,
			MinPayments:   minPayments,
			MaxPayments:   maxPayments,
			MinPercentage: firstCustomer.PercentRange.Min,
			MaxPercentage: firstCustomer.PercentRange.Max,
			Customers:     customerNames,
		})
	}

	// Используем CalculateB2BReplenishment из patterns.go
	b2bData := helpers.CalculateB2BReplenishment(req.Turnover, categories, req.Year, req.Month)

	// Конвертируем B2BTransactionData в entities.Transaction
	var transactions []*entities.Transaction
	transactionCounter := 1

	for _, data := range b2bData {
		// Преобразуем строковый метод платежа в value_objects.PaymentMethod
		var paymentMethod value_objects.PaymentMethod
		switch data.PaymentMethod {
		case "ACH-credit":
			paymentMethod = value_objects.ACHCredit
		case "Wire":
			paymentMethod = value_objects.Wire
		case "Zelle":
			paymentMethod = value_objects.Zelle
		default:
			paymentMethod = value_objects.ElectronicPayment
		}

		// Генерация даты (будний день, не праздник)
		transactionDate := data.Date
		if transactionDate.IsZero() {
			transactionDate = s.dateCalculator.GenerateRandomBusinessDate(req.Year, req.Month)
		}

		// [33] Генерация времени в рабочие часы (08:00-18:00)
		transactionTime := s.dateCalculator.GenerateBusinessTime(transactionDate, 8, 18)

		// [32] Проверка праздников для B2B-пополнений (ACH, Wire, internal transfers)
		// Операции по счету (ACH, Wire, Account/internal transfers) не проводятся в праздничные дни
		if paymentMethod.IsAccountTransfer() {
			if s.holidayService.IsHoliday(transactionDate) {
				transactionDate = s.holidayService.GetNextBusinessDay(transactionDate)
				transactionTime = s.dateCalculator.GenerateBusinessTime(transactionDate, 8, 18)
			}
		}

		// Создание транзакции
		transaction := entities.NewTransaction(
			utils.GenerateTransactionID("inc", transactionCounter),
			transactionTime,
			transactionDate,
			value_objects.Income,
			data.CustomerName,
			paymentMethod,
			data.Amount,
		)
		transaction.SetBalanceAfter(0) // Будет рассчитано позже

		transactions = append(transactions, transaction)
		transactionCounter++
	}

	return transactions
}

func (s *generatorService) generateExpenses(req *dto.GenerateRequest, totalExpensesTarget float64, userID *string) ([]*entities.Transaction, error) {
	// Получаем правила для последовательных выписок (README строка 74-78)
	// Используется для обеспечения 70% ± 10% повторяемости подрядчиков между месяцами
	// и проверки фиксированных контрагентов (Warehouse rent, ERP, Insurance)
	sequentialRules := helpers.GetSequentialStatementsRules()
	_ = sequentialRules // TODO: Реализовать логику повторяемости подрядчиков и фиксированных контрагентов

	var transactions []*entities.Transaction
	totalGenerated := 0.0

	// Разделяем шаблоны на обязательные и опциональные [39-41]
	mandatoryTemplates, optionalTemplates := s.separateTemplates()
	logrus.Infof("[DEBUG] separateTemplates: mandatory=%d, optional=%d", len(mandatoryTemplates), len(optionalTemplates))

	// Генерация обязательных расходов параллельно
	if len(mandatoryTemplates) > 0 {
		type templateResult struct {
			transactions []*entities.Transaction
			amount       float64
		}

		results := make([]templateResult, len(mandatoryTemplates))
		var wg sync.WaitGroup

		for i, template := range mandatoryTemplates {
			wg.Add(1)
			go func(idx int, tmpl *entities.TransactionTemplate) {
				defer wg.Done()
				templateTransactions, amount := s.generateTransactionsFromTemplate(req, tmpl, userID)
				results[idx] = templateResult{
					transactions: templateTransactions,
					amount:       amount,
				}
			}(i, template)
		}

		wg.Wait()

		// Собираем результаты
		for _, result := range results {
			transactions = append(transactions, result.transactions...)
			totalGenerated += result.amount
		}
	}

	// [12][40] Генерация опциональных расходов в пределах бюджета
	remainingBudget := totalExpensesTarget - totalGenerated
	logrus.Infof("[DEBUG] generateExpenses: totalExpensesTarget=%.2f, totalGenerated=%.2f, remainingBudget=%.2f",
		totalExpensesTarget, totalGenerated, remainingBudget)

	// Проверка: если обязательные расходы превышают бюджет, логируем предупреждение
	// Нормализация позже скорректирует суммы для достижения целевой прибыли
	if remainingBudget < 0 {
		logrus.Infof("[WARN] Mandatory expenses (%.2f) exceed totalExpensesTarget (%.2f) by %.2f. Normalization will adjust.",
			totalGenerated, totalExpensesTarget, -remainingBudget)
	}
	if remainingBudget > 0 {
		// Перемешиваем для случайного выбора [41]
		shuffledOptional := s.shuffleTemplates(optionalTemplates)

		// [25][14] Приоритет для "Подписка ПО" - всегда генерируем если есть бюджет
		var softwareSubscriptionTemplate *entities.TransactionTemplate
		var otherOptionalTemplates []*entities.TransactionTemplate
		logrus.Infof("[DEBUG] Processing %d optional templates", len(shuffledOptional))
		for _, template := range shuffledOptional {
			logrus.Infof("[DEBUG] Optional template: category=%s, isOptional=%v", template.Category, template.IsOptional)
			if value_objects.IsSoftwareSubscription(template.Category) {
				softwareSubscriptionTemplate = template
				logrus.Infof("[DEBUG] Found Подписка ПО template!")
			} else {
				otherOptionalTemplates = append(otherOptionalTemplates, template)
			}
		}

		// Сначала генерируем "Подписка ПО" если есть
		if softwareSubscriptionTemplate != nil && remainingBudget > 0 {
			logrus.Infof("[DEBUG] Generating Подписка ПО: remainingBudget=%.2f", remainingBudget)
			templateTransactions, amount := s.generateTransactionsFromTemplate(req, softwareSubscriptionTemplate, userID)
			logrus.Infof("[DEBUG] Подписка ПО generated: amount=%.2f, transactions=%d", amount, len(templateTransactions))
			if amount <= remainingBudget {
				transactions = append(transactions, templateTransactions...)
				totalGenerated += amount
				remainingBudget -= amount
				logrus.Infof("[DEBUG] Подписка ПО added to transactions, new remainingBudget=%.2f", remainingBudget)
			} else {
				logrus.Infof("[DEBUG] Подписка ПО amount (%.2f) exceeds remainingBudget (%.2f)", amount, remainingBudget)
			}
		} else {
			if softwareSubscriptionTemplate == nil {
				logrus.Infof("[DEBUG] softwareSubscriptionTemplate is nil - Подписка ПО not found in optional templates")
			} else {
				logrus.Infof("[DEBUG] remainingBudget <= 0: %.2f", remainingBudget)
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
	// Проверяем, можно ли использовать функции из patterns.go для этой категории
	if transactions, totalAmount, ok := s.expenseGenerator.GenerateFromPatterns(req, template, userID); ok {
		return transactions, totalAmount
	}

	// Если не подходит под patterns.go, используем старую логику
	var transactions []*entities.Transaction
	totalAmount := 0.0

	// Количество транзакций по шаблону
	// numTransactions := template.GetOccurrences()
	var numTransactions int
	if value_objects.IsIRSTaxes(template.Category) {
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
		percentage := utils.RandomPercentage(template.PercentageRange.Min, template.PercentageRange.Max)
		totalCategoryAmount = req.Turnover * percentage
		// Округляем общую сумму для категории, чтобы избежать ошибок округления при распределении
		totalCategoryAmount = utils.RoundToCents(totalCategoryAmount)
		
		// [205-207] Сохраняем детали расчёта для процентных операций в calculationDetails
		calculationDetails = map[string]interface{}{
			"type":              "percentage_expense",
			"percentage":        utils.FormatPercentage(percentage),
			"percentage_percent": utils.FormatPercentagePercent(percentage),
			"turnover":          req.Turnover,
			"total_amount":      totalCategoryAmount,
			"transaction_count": numTransactions,
			"formula":           "turnover * percentage / transaction_count",
		}
		} else {
			// Для фиксированных операций с расчетами [20][21]
			// [17][18] Специальная обработка для "Платные дороги": каждая транзакция должна иметь случайное значение $20/$35/$50
			isTollRoads := value_objects.IsTollRoads(template.Category)
			// [20][21] Специальная обработка для "Перегруз": каждая транзакция должна рассчитываться отдельно (вес * ставка)
			isOverload := value_objects.IsOverload(template.Category)
			
			if isTollRoads {
				// Для платных дорог не рассчитываем общую сумму заранее - каждая транзакция получит случайное значение
				totalCategoryAmount = 0
				calculationDetails = map[string]interface{}{
					"type":         "fixed_expense",
					"category":     "toll_roads",
					"description":  "Each transaction has random value: $20, $35, or $50",
					"values":        []float64{20.0, 35.0, 50.0},
				}
			} else if isOverload {
				// Для перегрузки не рассчитываем общую сумму заранее - каждая транзакция получит свой вес и ставку
				totalCategoryAmount = 0
				calculationDetails = map[string]interface{}{
					"type":         "fixed_expense",
					"category":     "overload",
					"description":  "Each transaction calculated as weight (200-1000 lb) * rate ($0.011-$0.039)",
					"formula":      "weight * rate",
				}
			} else {
				fixedAmount, details := s.fixedAmountCalculator.CalculateFixedAmount(template.Category, template.FixedAmount, req, userID)
				totalCategoryAmount = fixedAmount
				calculationDetails = details
				// [205-207] Если details == nil (для категорий без специальных расчетов), создаем базовые детали
				if calculationDetails == nil {
					calculationDetails = map[string]interface{}{
						"type":         "fixed_expense",
						"fixed_amount": totalCategoryAmount,
						"source":       "template",
						"category":     template.Category,
						"transaction_count": numTransactions,
					}
				} else {
					// Добавляем информацию о категории и количестве транзакций, если их нет в деталях
					if _, ok := calculationDetails["category"]; !ok {
						calculationDetails["category"] = template.Category
					}
					if _, ok := calculationDetails["transaction_count"]; !ok && numTransactions > 1 {
						calculationDetails["transaction_count"] = numTransactions
					}
				}
			}
		}

	for i := 0; i < numTransactions; i++ {
		// Расчет суммы для каждой транзакции
		var amount float64
		var transactionCalculationDetails map[string]interface{}

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
			// Для процентных используем общие детали расчёта
			transactionCalculationDetails = calculationDetails
		} else {
			// Для фиксированных операций
			isTollRoads := value_objects.IsTollRoads(template.Category)
			// [20][21] Для "Перегруз": каждая транзакция рассчитывается отдельно (вес * ставка)
			isOverload := value_objects.IsOverload(template.Category)
			
			if isTollRoads {
				// [17][18] Для платных дорог: каждая транзакция получает случайное значение $20/$35/$50
				tollAmount, tollDetails := s.fixedAmountCalculator.CalculateFixedAmount(template.Category, template.FixedAmount, req, userID)
				amount = tollAmount
				amount = utils.RoundToCents(amount)
				// Используем детали расчёта для этой конкретной транзакции
				// Добавляем номер транзакции в детали
				transactionCalculationDetails = make(map[string]interface{})
				for k, v := range tollDetails {
					transactionCalculationDetails[k] = v
				}
				transactionCalculationDetails["transaction_number"] = i + 1
				transactionCalculationDetails["total_transactions"] = numTransactions
				// Накапливаем сумму для каждой транзакции
				totalAmount += amount
			} else if isOverload {
				// [20][21] Для перегрузки: каждая транзакция получает свой вес и ставку
				overloadAmount, overloadDetails := s.fixedAmountCalculator.CalculateFixedAmount(template.Category, template.FixedAmount, req, userID)
				amount = overloadAmount
				amount = utils.RoundToCents(amount)
				// Используем детали расчёта для этой конкретной транзакции
				// Добавляем номер транзакции в детали
				transactionCalculationDetails = make(map[string]interface{})
				for k, v := range overloadDetails {
					transactionCalculationDetails[k] = v
				}
				transactionCalculationDetails["transaction_number"] = i + 1
				transactionCalculationDetails["total_transactions"] = numTransactions
				// Накапливаем сумму для каждой транзакции
				totalAmount += amount
			} else {
				// Для остальных фиксированных: используем уже рассчитанную сумму
				amount = totalCategoryAmount
				// Округляем фиксированную сумму
				amount = utils.RoundToCents(amount)
				// Для фиксированных накапливаем только один раз (если несколько транзакций - это ошибка конфигурации)
				if i == 0 {
					totalAmount += amount
				}
				// Используем общие детали расчёта, но добавляем информацию о транзакции
				if calculationDetails != nil {
					transactionCalculationDetails = make(map[string]interface{})
					for k, v := range calculationDetails {
						transactionCalculationDetails[k] = v
					}
					if numTransactions > 1 {
						transactionCalculationDetails["transaction_number"] = i + 1
						transactionCalculationDetails["total_transactions"] = numTransactions
					}
				} else {
					transactionCalculationDetails = nil
				}
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
		if value_objects.IsSoftwareSubscription(template.Category) {
			userIDUUID := utils.ParseUserID(userID)
			logrus.Infof("[DEBUG] generateTransactionsFromTemplate: Подписка ПО detected, userID=%v, userIDUUID=%v", userID, userIDUUID)
			baseDate := utils.FirstDayOfMonth(req.Year, req.Month)
			logrus.Infof("[DEBUG] Calling calculateSoftwareSubscriptionDate: baseDate=%v, userIDUUID=%v", baseDate.Format("2006-01-02"), userIDUUID)
			transactionDate = s.dateCalculator.CalculateSoftwareSubscriptionDate(baseDate, userIDUUID)
			postingDate = transactionDate
			logrus.Infof("[DEBUG] Подписка ПО date generated: %v (weekday=%d)", transactionDate.Format("2006-01-02"), int(transactionDate.Weekday()))
		} else if value_objects.IsIRSTaxes(template.Category) {
			// [23][24] Для IRS налогов - всегда 15-е число (или следующий рабочий день)
			transactionDate = s.dateCalculator.CalculateIRSDate(req.Year, req.Month, i+1)
			postingDate = transactionDate
		} else if value_objects.IsOwnerTransfer(template.Category) {
			// [22] Для "Перевод владельцу" - 1 раз в месяц, в будний день (не праздничный)
			// generateRandomBusinessDate уже гарантирует будний день (не праздничный)
			transactionDate = s.dateCalculator.GenerateRandomBusinessDate(req.Year, req.Month)
			// postingDate должен быть таким же, как transactionDate, или скорректированным, если это праздник
			// Но generateRandomBusinessDate уже гарантирует будний день, поэтому postingDate = transactionDate
			postingDate = transactionDate
		} else if value_objects.IsMobile(template.Category) {
			// [26][27] Для мобильной связи - 2-я пятница месяца
			baseDate := utils.FirstDayOfMonth(req.Year, req.Month)
			transactionDate = s.dateCalculator.FindNthWeekdayInMonth(baseDate, "Friday", 2)
			// [32] Проверяем праздники только для операций по счету (мобильная связь обычно по карте)
			if template.PaymentMethod.IsAccountTransfer() && s.holidayService.IsHoliday(transactionDate) {
				transactionDate = s.holidayService.GetNextBusinessDay(transactionDate)
			}
			postingDate = transactionDate
		} else {
			transactionDate = s.dateCalculator.CalculateTransactionDate(template, req.Year, req.Month, i+1)
			postingDate = s.dateCalculator.CalculatePostingDate(template, req.Year, req.Month, i+1)
		}

		// [32] Корректировка даты если праздник (для операций по счету)
		// Для "Перевод владельцу" дата уже гарантированно будний день (не праздничный), пропускаем
		// Для IRS налогов дата уже скорректирована в calculateIRSDate, пропускаем
		isOwnerTransfer := value_objects.IsOwnerTransfer(template.Category)
		if template.PaymentMethod.IsAccountTransfer() && !isOwnerTransfer && !value_objects.IsIRSTaxes(template.Category) {
			if s.holidayService.IsHoliday(transactionDate) {
				transactionDate = s.holidayService.GetNextBusinessDay(transactionDate)
			}
			if s.holidayService.IsHoliday(postingDate) {
				postingDate = s.holidayService.GetNextBusinessDay(postingDate)
			}
		}

		// [33] Генерация времени согласно BusinessHours
		var transactionTime time.Time
		if value_objects.IsSoftwareSubscription(template.Category) {
			// [33] 00:01 для подписок
			transactionTime = time.Date(
				transactionDate.Year(), transactionDate.Month(), transactionDate.Day(),
				0, 1, 0, 0, time.UTC)
		} else if template.PaymentMethod.IsCardOperation() {
			// [33] 09:00-20:00 для карт
			transactionTime = s.dateCalculator.GenerateBusinessTime(transactionDate, 9, 20)
		} else {
			// [33] 08:00-18:00 для операций по счету
			// Используем BusinessHours из шаблона если есть
			startHour, endHour := 8, 18
			if template.BusinessHours.Start != "" && template.BusinessHours.End != "" {
				startHour, _ = utils.ParseTimeFromString(template.BusinessHours.Start)
				endHour, _ = utils.ParseTimeFromString(template.BusinessHours.End)
			}
			transactionTime = s.dateCalculator.GenerateBusinessTime(transactionDate, startHour, endHour)
		}

		// Создание транзакции
		// amount уже округлен выше для процентных транзакций
		transaction := entities.NewTransaction(
			utils.GenerateTemplateTransactionID(template.Category, i+1),
			transactionTime,
			postingDate,
			template.Type,
			template.Category,
			template.PaymentMethod,
			amount, // Уже округлен для процентных, для фиксированных тоже округлен
		)

		// [205-207] Сохраняем детали расчёта для каждой транзакции
		// Для платных дорог каждая транзакция имеет свои детали (случайное значение)
		// Для остальных категорий используются общие детали расчёта
		if transactionCalculationDetails != nil {
			transaction.SetCalculationDetails(transactionCalculationDetails)
		} else if calculationDetails != nil {
			// Fallback на общие детали, если для транзакции не заданы отдельные
			transaction.SetCalculationDetails(calculationDetails)
		}

		transactions = append(transactions, transaction)
		// totalAmount уже накоплен выше для всех транзакций (до применения знака)
	}

	// Возвращаем общую сумму категории (для процентных это totalCategoryAmount, для фиксированных - сумма всех транзакций)
	return transactions, totalAmount
}

// convertManualTransactions конвертирует ручные транзакции из DTO в entities
func (s *generatorService) convertManualTransactions(manual []dto.ManualTransaction) []*entities.Transaction {
	var transactions []*entities.Transaction
	for i, manualTx := range manual {
		transactionType, _ := value_objects.NewTransactionType(manualTx.Type)
		paymentMethod, _ := value_objects.NewPaymentMethod(manualTx.Method)

		transaction := entities.NewTransaction(
			utils.GenerateTransactionID("manual", i+1),
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
	if s.dateCalculator.IsQuarterlyMonth(month) {
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
		if value_objects.IsSoftwareSubscription(template.Category) {
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

// getOptionalCategoriesMap возвращает карту опциональных категорий для быстрой проверки [12][40]
func (s *generatorService) getOptionalCategoriesMap() map[string]bool {
	optionalMap := make(map[string]bool)
	for _, template := range s.templates {
		// [25][14] "Подписка ПО" не является опциональной для удаления, даже если помечена как optional
		if template.IsOptional && !value_objects.IsSoftwareSubscription(template.Category) {
			optionalMap[template.Category] = true
		}
	}
	return optionalMap
}

// removeSmallOptionalTransactions удаляет мелкие опциональные транзакции при добавлении ручных операций [38]
// [12][40] Удаляет мелкие транзакции из опциональных категорий, чтобы не выйти за пределы 39-75 транзакций
func (s *generatorService) removeSmallOptionalTransactions(transactions []*entities.Transaction, optionalCategories map[string]bool, model string) []*entities.Transaction {
	if len(optionalCategories) == 0 {
		return transactions
	}

	// Определяем максимальное количество транзакций (75 для B2C/B2B)
	maxTransactions := 75
	if model == "B2C" {
		maxTransactions = 75
	} else if model == "B2B" {
		maxTransactions = 75
	}

	// Если количество транзакций не превышает лимит, ничего не делаем
	if len(transactions) <= maxTransactions {
		return transactions
	}

	// Собираем опциональные транзакции расходов (не ручные)
	optionalExpenses := make([]*entities.Transaction, 0)
	otherTransactions := make([]*entities.Transaction, 0)

	for _, tx := range transactions {
		if tx.IsExpense() && !tx.IsManualTransaction() && optionalCategories[tx.Category] {
			optionalExpenses = append(optionalExpenses, tx)
		} else {
			otherTransactions = append(otherTransactions, tx)
		}
	}

	// Сортируем опциональные транзакции по абсолютной сумме (от меньших к большим)
	for i := 0; i < len(optionalExpenses)-1; i++ {
		for j := i + 1; j < len(optionalExpenses); j++ {
			if math.Abs(optionalExpenses[i].Amount) > math.Abs(optionalExpenses[j].Amount) {
				optionalExpenses[i], optionalExpenses[j] = optionalExpenses[j], optionalExpenses[i]
			}
		}
	}

	// Удаляем мелкие опциональные транзакции до тех пор, пока не достигнем лимита
	excessCount := len(transactions) - maxTransactions
	remainingOptional := make([]*entities.Transaction, 0)

	for i, tx := range optionalExpenses {
		if i < excessCount {
			// Удаляем эту транзакцию
			logrus.Infof("[INFO] Removing small optional transaction after manual transactions: %s, amount=%.2f",
				tx.Category, tx.Amount)
		} else {
			// Оставляем эту транзакцию
			remainingOptional = append(remainingOptional, tx)
		}
	}

	// Объединяем обязательные и оставшиеся опциональные транзакции
	result := otherTransactions
	result = append(result, remainingOptional...)

	return result
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
		logrus.Infof("[WARN] Failed to get user %v: %v, associated card not available", userID, err)
		return ""
	}

	// Если у пользователя есть сохраненный номер карты, возвращаем его
	if userModel.AssociatedCard != nil && *userModel.AssociatedCard != "" {
		return *userModel.AssociatedCard
	}

	// Если номер карты не найден, возвращаем пустую строку
	// Пользователь должен задать номер карты через API /api/user/associated-card
	logrus.Infof("[WARN] User %v does not have associated card set. Please set it via /api/user/associated-card endpoint", userID)
	return ""
}

// scaleTransactions масштабирует транзакции [38]
// [38] Не масштабирует Payroll ADP и фиксированные 1 раз в месяц (Подписка ПО, Мобильная связь, Коммунальные, Лизинг)
func (s *generatorService) scaleTransactions(transactions []*entities.Transaction, scaleFactor int) []*entities.Transaction {
	var scaled []*entities.Transaction

	// [38] Категории, которые не масштабируются (фиксированные 1 раз в месяц)
	fixedMonthlyCategories := value_objects.GetFixedMonthlyCategories()

	for _, tx := range transactions {
		// [38] Не масштабируем Payroll ADP, фиксированные 1 раз в месяц и ручные транзакции
		if fixedMonthlyCategories[tx.Category] || tx.IsManualTransaction() {
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

// TODO: вынести в domain
// convertToDomainTransactions конвертирует entities.Transaction в domain.GeneratedTransaction для сохранения в БД
// Использует фабрику domain.FromEntityTransaction для создания сущностей
func (s *generatorService) convertToDomainTransactions(transactions []*entities.Transaction, requestID uuid.UUID) []*domain.GeneratedTransaction {
	domainTransactions := make([]*domain.GeneratedTransaction, len(transactions))

	if len(transactions) == 0 {
		return domainTransactions
	}

	// Параллельная конвертация для больших объемов
	if len(transactions) > 100 {
		var wg sync.WaitGroup
		batchSize := len(transactions) / 4 // 4 горутины
		if batchSize < 1 {
			batchSize = 1
		}

		for batch := 0; batch < 4; batch++ {
			start := batch * batchSize
			end := start + batchSize
			if batch == 3 || end > len(transactions) {
				end = len(transactions)
			}

			if start >= end {
				break
			}

			wg.Add(1)
			go func(startIdx, endIdx int) {
				defer wg.Done()
				for i := startIdx; i < endIdx; i++ {
					sortOrder := i + 1
					domainTransactions[i] = domain.FromEntityTransaction(requestID, transactions[i], sortOrder)
				}
			}(start, end)
		}

		wg.Wait()
	} else {
		// Для малых объемов выполняем последовательно
		for i, tx := range transactions {
			sortOrder := i + 1
			domainTransactions[i] = domain.FromEntityTransaction(requestID, tx, sortOrder)
		}
	}

	return domainTransactions
}
// 1591
// 1878
// 2416
// 2514