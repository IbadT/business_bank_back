// internal/service/generator.go
package service

import (
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/internal/domain/entities"
	"github.com/IbadT/business_bank_back/services/matematika/internal/domain/value_objects"
	"github.com/IbadT/business_bank_back/services/matematika/internal/repository"
	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/dto"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/utils"
	"github.com/google/uuid"
)

var (
	ErrNegativeBalance = errors.New("transaction would result in negative balance") // [43]
	ErrInvalidModel    = errors.New("invalid business model")
	ErrInvalidRequest  = errors.New("invalid request parameters")
)

// GeneratorService - Use Case генерации транзакций
type GeneratorService interface {
	GenerateTransactions(req *dto.GenerateRequest, userID *string) (*dto.GenerateResponse, error)
}

type generatorService struct {
	configRepo      repository.ConfigRepository
	stateRepo       repository.StateRepository
	holidayRepo     repository.HolidayRepository
	dateCalculator  *dateCalculator
	amountCalculator *amountCalculator
	holidayService  HolidayService
	templates       []*entities.TransactionTemplate
	gateways        []*entities.Gateway
	customers       []*entities.Customer
}

func NewGeneratorService(configRepo repository.ConfigRepository, stateRepo repository.StateRepository, holidayRepo repository.HolidayRepository) (GeneratorService, error) {
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
		configRepo:      configRepo,
		stateRepo:       stateRepo,
		holidayRepo:     holidayRepo,
		dateCalculator:  newDateCalculator(holidays, stateRepo),
		amountCalculator: newAmountCalculator(),
		holidayService:  NewHolidayService(holidayRepo),
		templates:       templates,
		gateways:        gateways,
		customers:       customers,
	}, nil
}

func (s *generatorService) GenerateTransactions(req *dto.GenerateRequest, userID *string) (*dto.GenerateResponse, error) {
	// [1] Валидация запроса
	if err := s.validateRequest(req); err != nil {
		return nil, err
	}
	
	// [1][5] Рассчет целевой прибыли
	targetProfit := req.Turnover * (req.DesiredProfitPercent / 100)
	totalExpensesTarget := req.Turnover - targetProfit
	
	// [2][35][36] Генерация доходов в зависимости от модели
	var incomeTransactions []*entities.Transaction
	switch req.Model {
	case "B2C":
		incomeTransactions = s.generateB2CIncomes(req)
	case "B2B":
		incomeTransactions = s.generateB2BIncomes(req)
	default:
		return nil, ErrInvalidModel
	}
	
	// [7-21] Генерация расходов
	expenseTransactions, err := s.generateExpenses(req, totalExpensesTarget, userID)
	if err != nil {
		return nil, err
	}
	
	// [38] Объединение транзакций
	allTransactions := append(incomeTransactions, expenseTransactions...)
	
	// [38] Добавление ручных транзакций
	if req.CustomData != nil && len(req.CustomData.ManualTransactions) > 0 {
		manualTransactions := s.convertManualTransactions(req.CustomData.ManualTransactions)
		allTransactions = append(allTransactions, manualTransactions...)
	}
	
	// [38] Масштабирование
	if req.ScaleFactor > 1 {
		allTransactions = s.scaleTransactions(allTransactions, req.ScaleFactor)
	}
	
	// Сортировка по дате
	allTransactions = s.sortTransactionsByDate(allTransactions)
	
	// [42] Балансировка и нормализация
	balancedTransactions, err := s.balanceAndNormalize(allTransactions, req.Turnover, targetProfit)
	if err != nil {
		return nil, err
	}
	
	// Расчет балансов
	transactionsWithBalance, err := s.calculateBalances(balancedTransactions, req.InitialBalance)
	if err != nil {
		return nil, err
	}
	
	// [43] Проверка отрицательного баланса
	if err := s.checkNegativeBalance(transactionsWithBalance); err != nil {
		return nil, err
	}
	
	// Формирование ответа
	return s.buildResponse(transactionsWithBalance, req), nil
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

func (s *generatorService) generateB2CIncomes(req *dto.GenerateRequest) []*entities.Transaction {
	// [35] Получаем пятницы в месяце
	fridays := s.dateCalculator.getFridaysInMonth(req.Year, req.Month)
	
	// [35] Выбираем шлюз (сохраняем для всех месяцев)
	gateway := s.selectGateway()
	
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
	
	return transactions
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
	numTransactions := template.GetOccurrences()
	
	for i := 0; i < numTransactions; i++ {
		// Расчет суммы
		amount := template.CalculateAmount(req.Turnover)
		var calculationDetails map[string]interface{}
		
		// Для фиксированных операций с расчетами [20][21]
		if !template.IsPercentage {
			amount, calculationDetails = s.calculateFixedAmount(template.Category, amount, req)
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
		} else {
			transactionDate = s.dateCalculator.calculateTransactionDate(template, req.Year, req.Month, i+1)
			postingDate = s.dateCalculator.calculatePostingDate(template, req.Year, req.Month, i+1)
		}
		
		// [32] Корректировка даты если праздник (для операций по счету)
		if template.PaymentMethod.IsAccountTransfer() {
			transactionDate = s.holidayService.GetNextBusinessDay(transactionDate)
			postingDate = s.holidayService.GetNextBusinessDay(postingDate)
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
		transaction := entities.NewTransaction(
			generateTemplateTransactionID(template.Category, i+1),
			transactionTime,
			postingDate,
			template.Type,
			template.Category,
			template.PaymentMethod,
			utils.RoundToCents(amount),
		)
		
		if calculationDetails != nil {
			transaction.SetCalculationDetails(calculationDetails)
		}
		
		transactions = append(transactions, transaction)
		totalAmount += math.Abs(amount) // Для подсчета используем абсолютное значение
	}
	
	return transactions, totalAmount
}

func (s *generatorService) calculateFixedAmount(category string, baseAmount float64, req *dto.GenerateRequest) (float64, map[string]interface{}) {
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
		
	case "Лизинг":
		// [19] Логика лизинга: первый месяц 11.5-12% оборота, затем фиксируется
		firstMonth := s.amountCalculator.isFirstMonth(req)
		if firstMonth {
			// [19] 11.5-12% оборота для первого месяца
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
			// [19] Повторяется 1:1 в последующих месяцах
			amount := s.amountCalculator.getSavedLeaseAmount()
			details := map[string]interface{}{
				"type": "recurring_lease",
			}
			return amount, details
		}
		
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
	// Используем встроенную сортировку
	for i := 0; i < len(transactions)-1; i++ {
		for j := i + 1; j < len(transactions); j++ {
			if transactions[i].TransactionDate.After(transactions[j].TransactionDate) {
				transactions[i], transactions[j] = transactions[j], transactions[i]
			}
		}
	}
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
	
	// Корректируем последнюю транзакцию каждого типа
	if len(transactions) > 0 {
		for i := len(transactions) - 1; i >= 0; i-- {
			if transactions[i].IsIncome() && incomeDiff != 0 {
				transactions[i].Amount += incomeDiff
				transactions[i].Amount = utils.RoundToCents(transactions[i].Amount)
				incomeDiff = 0
			}
			if transactions[i].IsExpense() && profitDiff != 0 {
				transactions[i].Amount -= profitDiff
				transactions[i].Amount = utils.RoundToCents(transactions[i].Amount)
				profitDiff = 0
			}
		}
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
func (s *generatorService) buildResponse(transactions []*entities.Transaction, req *dto.GenerateRequest) *dto.GenerateResponse {
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
	
	summary := dto.FinancialSummary{
		InitialBalance: utils.RoundToCents(req.InitialBalance),
		FinalBalance:   utils.RoundToCents(finalBalance),
		TotalRevenue:   utils.RoundToCents(totalRevenue),
		TotalExpenses:  utils.RoundToCents(totalExpenses),
		NetProfit:      utils.RoundToCents(totalRevenue + totalExpenses),
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
			CalculationDetails: tx.CalculationDetails,
		}
	}
	
	// Формируем forwardingInfo
	forwardingInfo := dto.ForwardingInfo{
		AssociatedCard:  s.generateCardNumber(),
		OwnerName:       "",
		CompanyName:     "",
		CustomCustomers: []string{},
		CustomContractors: []dto.CustomContractor{},
	}
	
	if req.CustomData != nil {
		forwardingInfo.OwnerName = req.CustomData.CompanyInfo.OwnerName
		forwardingInfo.CompanyName = req.CustomData.CompanyInfo.CompanyName
		forwardingInfo.CustomCustomers = req.CustomData.CustomCustomers
		forwardingInfo.CustomContractors = req.CustomData.CustomContractors
	}
	
	return &dto.GenerateResponse{
		Transactions:         dtoTransactions,
		FinancialSummary:     summary,
		DailyClosingBalances: dailyBalances,
		ForwardingInfo:       forwardingInfo,
	}
}

// calculateDailyBalances рассчитывает ежедневные балансы [50]
func (s *generatorService) calculateDailyBalances(transactions []*entities.Transaction, initialBalance float64, year, month int) []dto.DailyBalance {
	balancesByDate := make(map[string]float64)
	
	// Начальный баланс на первый день месяца
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	balancesByDate[firstDay.Format("2006-01-02")] = initialBalance
	
	// Обновляем балансы после каждой транзакции
	for _, tx := range transactions {
		dateKey := tx.PostingDate.Format("2006-01-02")
		balancesByDate[dateKey] = tx.BalanceAfter
	}
	
	// Преобразуем в срез
	var dailyBalances []dto.DailyBalance
	for dateStr, balance := range balancesByDate {
		dailyBalances = append(dailyBalances, dto.DailyBalance{
			Date:    dateStr,
			Balance: utils.RoundToCents(balance),
		})
	}
	
	// Сортируем по дате
	for i := 0; i < len(dailyBalances)-1; i++ {
		for j := i + 1; j < len(dailyBalances); j++ {
			if dailyBalances[i].Date > dailyBalances[j].Date {
				dailyBalances[i], dailyBalances[j] = dailyBalances[j], dailyBalances[i]
			}
		}
	}
	
	return dailyBalances
}

// generateCardNumber генерирует номер карты [51][52]
func (s *generatorService) generateCardNumber() string {
	return fmt.Sprintf("%016d", rand.Int63n(10000000000000000))
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











// // internal/service/generator.go
// package service

// import (
// 	"errors"
// 	"fmt"
// 	"time"
// 	"math"
// 	"math/rand"

// 	"matematika-service/models"
// 	"matematika-service/internal/repository"
// 	"matematika-service/pkg/utils"
// )

// var (
// 	ErrNegativeBalance = errors.New("транзакция приведет к отрицательному балансу") // [43]
// 	ErrInvalidModel    = errors.New("неверная бизнес-модель")
// )

// type GeneratorService interface {
// 	GenerateTransactions(req *models.GenerateRequest) (*models.GenerateResponse, error)
// }

// type generatorService struct {
// 	configRepo repository.ConfigRepository
// 	calculator *dateCalculator
// 	holidays   []models.Holiday
// 	templates  []models.TransactionTemplate
// 	gateways   []models.Gateway
// 	customers  []models.DefaultCustomer
// }

// func NewGeneratorService(configRepo repository.ConfigRepository) (GeneratorService, error) {
// 	// Загружаем конфигурации при инициализации
// 	holidays, err := configRepo.GetHolidays()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to load holidays: %w", err)
// 	}

// 	templates, err := configRepo.GetTransactionTemplates()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to load templates: %w", err)
// 	}

// 	gateways, err := configRepo.GetGateways()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to load gateways: %w", err)
// 	}

// 	customers, err := configRepo.GetDefaultCustomers()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to load customers: %w", err)
// 	}

// 	return &generatorService{
// 		configRepo: configRepo,
// 		calculator: newDateCalculator(holidays),
// 		holidays:   holidays,
// 		templates:  templates,
// 		gateways:   gateways,
// 		customers:  customers,
// 	}, nil
// }

// func (s *generatorService) GenerateTransactions(req *models.GenerateRequest) (*models.GenerateResponse, error) {
// 	// [1] Рассчитываем целевую прибыль
// 	targetProfit := req.Turnover * (req.DesiredProfitPercent / 100)
// 	totalExpensesTarget := req.Turnover - targetProfit

// 	// [2] Генерируем доходы в зависимости от модели
// 	var incomeTransactions []models.Transaction
// 	switch req.Model {
// 	case "B2C":
// 		incomeTransactions = s.generateB2CIncomes(req)
// 	case "B2B":
// 		incomeTransactions = s.generateB2BIncomes(req)
// 	default:
// 		return nil, ErrInvalidModel
// 	}

// 	// [3] Генерируем расходы
// 	expenseTransactions, err := s.generateExpenses(req, totalExpensesTarget)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// [4] Объединяем транзакции
// 	allTransactions := append(incomeTransactions, expenseTransactions...)

// 	// [5] Добавляем ручные транзакции если есть
// 	if req.CustomData != nil && len(req.CustomData.ManualTransactions) > 0 {
// 		allTransactions = s.mergeManualTransactions(allTransactions, req.CustomData.ManualTransactions)
// 	}

// 	// [6] Масштабируем если нужно
// 	if req.ScaleFactor > 1 {
// 		allTransactions = s.scaleTransactions(allTransactions, req.ScaleFactor)
// 	}

// 	// [7] Сортируем по дате
// 	allTransactions = s.sortTransactionsByDate(allTransactions)

// 	// [8] Балансируем и нормализуем суммы
// 	balancedTransactions, err := s.balanceAndNormalize(allTransactions, req.Turnover, targetProfit)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// [9] Рассчитываем балансы
// 	transactionsWithBalance, err := s.calculateBalances(balancedTransactions, req.InitialBalance)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// [10] Проверяем отрицательный баланс
// 	if err := s.checkNegativeBalance(transactionsWithBalance); err != nil {
// 		return nil, err
// 	}

// 	// [11] Формируем ответ
// 	response := s.buildResponse(transactionsWithBalance, req)

// 	return response, nil
// }