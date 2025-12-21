package balanceservice

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/internal/domain"
	"github.com/IbadT/business_bank_back/services/matematika/internal/domain/entities"
	"github.com/IbadT/business_bank_back/services/matematika/internal/repository"
	dateservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/date"
	generatorservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/date"
	transactionservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/transaction"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/helpers"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/logger"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/utils"
	"github.com/google/uuid"
)

// BalanceHandlingStrategy - стратегия обработки недостатка баланса
type BalanceHandlingStrategy string

const (
	StrategyPostpone BalanceHandlingStrategy = "postpone" // перенос на день позже
	StrategyReduce   BalanceHandlingStrategy = "reduce"   // уменьшение суммы
	StrategyHybrid   BalanceHandlingStrategy = "hybrid"   // комбинированная
)

// BalanceAdjustment - информация о корректировке транзакции
type BalanceAdjustment struct {
	TransactionID                  string
	WasAdjusted                    bool
	AdjustmentType                 string
	OriginalDate                   *time.Time
	OriginalAmount                 *float64
	AdjustmentReason               string
	AvailableBalanceAtOriginalDate float64
	RequiredBalance                float64
	Shortage                       float64
}

// BalanceIssue - информация о проблеме с балансом
type BalanceIssue struct {
	TransactionID    string
	Date             string
	RequiredBalance  float64
	AvailableBalance float64
	Shortage         float64
	ActionTaken      string
	NewDate          string
	OriginalAmount   float64
	AdjustedAmount   float64
}

// ValidateBalanceResult - результат валидации баланса
type ValidateBalanceResult struct {
	IsValid   bool
	Issues    []BalanceIssue
	RequestID string
}

type BalanceAdjustmentService interface {
	GetAdjustedTransactions(requestIDStr string) ([]domain.GeneratedTransaction, error)
	ValidateBalance(requestIDStr string) (*ValidateBalanceResult, error)
	AdjustTransactionsForBalance(
		transactions []*entities.Transaction,
		initialBalance float64,
		strategy BalanceHandlingStrategy,
		dateCalculator *dateservice.DateCalculator,
		year, month int,
	) ([]*entities.Transaction, []*BalanceAdjustment, error)
	// BalanceAndNormalize балансирует и нормализует суммы транзакций
	// optionalCategories - карта опциональных категорий для приоритетного удаления
	BalanceAndNormalize(transactions []*entities.Transaction, turnover, targetProfit float64, optionalCategories map[string]bool) ([]*entities.Transaction, error)
	// CalculateBalances рассчитывает балансы после каждой транзакции
	CalculateBalances(transactions []*entities.Transaction, initialBalance float64) ([]*entities.Transaction, error)
	// RecalculateBalances пересчитывает балансы после корректировки транзакций
	RecalculateBalances(transactions []*entities.Transaction, initialBalance float64) ([]*entities.Transaction, error)
}

type balanceAdjustmentService struct {
	transactionRepo       repository.TransactionRepository
	transactionService    transactionservice.TransactionService
	generationRequestRepo repository.GenerationRequestRepository
}

func NewBalanceAdjustmentService(transactionRepo repository.TransactionRepository, transactionService transactionservice.TransactionService, generationRequestRepo repository.GenerationRequestRepository) BalanceAdjustmentService {
	return &balanceAdjustmentService{
		transactionRepo:       transactionRepo,
		transactionService:    transactionService,
		generationRequestRepo: generationRequestRepo,
	}
}

func (s *balanceAdjustmentService) GetAdjustedTransactions(requestIDStr string) ([]domain.GeneratedTransaction, error) {
	op := "service.balance.getAdjustedTransactions"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"request_id": requestIDStr})
	log.Info("Getting adjusted transactions")

	requestID, err := helpers.ParseUUID(requestIDStr)
	if err != nil {
		log.Error(err, "Invalid requestID format")
		return nil, err
	}

	if requestID == uuid.Nil {
		log.Warn("requestID cannot be empty")
		return nil, helpers.ErrRequestIDEmpty
	}

	// Получаем модели из БД
	ormTransactions, err := s.transactionRepo.GetAdjustedTransactionsByRequestID(requestID)
	if err != nil {
		log.Error(err, "Failed to get adjusted transactions from repository")
		return nil, fmt.Errorf("%w: %w", helpers.ErrFailedToGetAdjustedTransactions, err)
	}

	// Конвертируем в domain
	domainTransactions := make([]domain.GeneratedTransaction, len(ormTransactions))
	for i, tx := range ormTransactions {
		domainTransactions[i] = domain.GeneratedTransaction{
			ID:              tx.ID,
			RequestID:       tx.RequestID,
			TransactionID:   tx.TransactionID,
			TransactionDate: tx.TransactionDate,
			PostingDate:     tx.PostingDate,
			Type:            tx.Type,
			Category:        tx.Category,
			Method:          tx.Method,
			Amount:          tx.Amount,
			BalanceAfter:    tx.BalanceAfter,
			IsManual:        tx.IsManual,
			SortOrder:       tx.SortOrder,
		}
	}

	log.WithFields(logger.Fields{"count": len(domainTransactions)}).Success("Adjusted transactions retrieved")
	return domainTransactions, nil
}

// ValidateBalance - валидация баланса транзакций
func (s *balanceAdjustmentService) ValidateBalance(requestIDStr string) (*ValidateBalanceResult, error) {
	op := "service.balance.validateBalance"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"request_id": requestIDStr})
	log.Info("Validating balance")

	// Парсим request_id
	requestID, err := helpers.ParseUUID(requestIDStr)
	if err != nil {
		log.Error(err, "Invalid requestID format")
		return nil, err
	}

	// Получаем GenerationRequest для получения initial_balance
	genRequest, err := s.generationRequestRepo.GetByID(requestID)
	if err != nil {
		log.Error(err, "Generation request not found")
		return nil, fmt.Errorf("%w: %w", helpers.ErrGenerationRequestNotFound, err)
	}

	// Получаем транзакции по request_id (уже отсортированы по дате в БД)
	transactions, err := s.transactionService.GetByRequestID(requestIDStr)
	if err != nil {
		log.Error(err, "Failed to get transactions")
		return nil, fmt.Errorf("%w: %w", helpers.ErrFailedToGetTransactions, err)
	}

	if len(transactions) == 0 {
		log.Warn("No transactions found for request_id")
		return nil, fmt.Errorf("%w: %s", helpers.ErrNoTransactionsFound, requestIDStr)
	}

	// Проверяем балансы транзакций
	issues := []BalanceIssue{}
	isValid := true

	// Используем initial_balance из GenerationRequest
	initialBalance := genRequest.InitialBalance
	currentBalance := initialBalance

	for _, tx := range transactions {
		// Рассчитываем баланс до этой транзакции
		balanceBefore := currentBalance

		// Рассчитываем баланс после этой транзакции
		calculatedBalanceAfter := balanceBefore + tx.Amount

		// Проверяем, был ли отрицательный баланс ДО корректировки
		// или есть ли проблемы сейчас
		// Используем рассчитанный баланс, а не сохраненный (который может быть неправильным)
		hasNegativeBalance := calculatedBalanceAfter < 0

		// Проверяем, была ли транзакция скорректирована
		wasAdjusted := false
		adjustmentType := ""
		originalDate := ""
		originalAmount := 0.0

		if tx.CalculationDetails != nil {
			if wasAdj, ok := tx.CalculationDetails["was_adjusted"].(bool); ok && wasAdj {
				wasAdjusted = true

				// Получаем тип корректировки
				if adjType, ok := tx.CalculationDetails["adjustment_type"].(string); ok {
					adjustmentType = adjType
				}

				// Получаем информацию о корректировке
				if origDate, ok := tx.CalculationDetails["original_date"].(string); ok && origDate != "" {
					originalDate = origDate
				}
				if origAmount, ok := tx.CalculationDetails["original_amount"].(float64); ok {
					originalAmount = origAmount
				}
			}
		}

		// Если есть отрицательный баланс ИЛИ была корректировка - добавляем в issues
		if hasNegativeBalance || wasAdjusted {
			if hasNegativeBalance {
				isValid = false
			}

			issue := BalanceIssue{
				TransactionID:    tx.TransactionID,
				Date:             tx.TransactionDate.Format("2006-01-02"),
				RequiredBalance:  balanceBefore + tx.Amount, // Баланс, который нужен для транзакции
				AvailableBalance: balanceBefore,
				Shortage:         0,
				ActionTaken:      "none",
			}

			// Если есть отрицательный баланс, рассчитываем недостачу
			if hasNegativeBalance {
				issue.Shortage = -calculatedBalanceAfter
			} else if wasAdjusted {
				// Если была корректировка, рассчитываем недостачу на основе оригинальной суммы
				if originalAmount != 0 {
					// Рассчитываем, сколько не хватало для оригинальной суммы
					requiredForOriginal := balanceBefore + originalAmount
					if requiredForOriginal < 0 {
						issue.Shortage = -requiredForOriginal
					} else if originalAmount != tx.Amount {
						// Если сумма была уменьшена, показываем разницу
						issue.Shortage = originalAmount - tx.Amount
					}
				}
			}

			// Устанавливаем информацию о корректировке
			if wasAdjusted {
				if adjustmentType != "" {
					issue.ActionTaken = adjustmentType
				} else {
					issue.ActionTaken = "adjusted"
				}

				if originalDate != "" {
					issue.NewDate = originalDate
				}
				if originalAmount != 0 {
					issue.OriginalAmount = originalAmount
					issue.AdjustedAmount = tx.Amount
				}
			}

			issues = append(issues, issue)
		}

		// Обновляем текущий баланс на основе рассчитанного значения
		// Используем рассчитанный баланс, а не сохраненный (который может быть неправильным)
		currentBalance = calculatedBalanceAfter
	}

	log.WithFields(logger.Fields{
		"is_valid":    isValid,
		"issues_count": len(issues),
		"transactions_count": len(transactions),
	}).Success("Balance validation completed")

	return &ValidateBalanceResult{
		IsValid:   isValid,
		Issues:    issues,
		RequestID: requestIDStr,
	}, nil
}

// AdjustTransactionsForBalance - основная функция корректировки транзакций при недостатке баланса
func (s *balanceAdjustmentService) AdjustTransactionsForBalance(
	transactions []*entities.Transaction,
	initialBalance float64,
	strategy BalanceHandlingStrategy,
	dateCalculator *generatorservice.DateCalculator,
	year, month int,
) ([]*entities.Transaction, []*BalanceAdjustment, error) {
	op := "service.balance.adjustTransactionsForBalance"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"strategy":        strategy,
		"initial_balance": initialBalance,
		"transactions_count": len(transactions),
	})
	log.Info("Adjusting transactions for balance")

	if len(transactions) == 0 {
		log.Debug("No transactions to adjust")
		return transactions, []*BalanceAdjustment{}, nil
	}

	// Сортируем транзакции по дате
	sorted := utils.SortTransactionsByDate(transactions)
	adjustments := []*BalanceAdjustment{}
	maxIterations := 10 // защита от бесконечного цикла
	iteration := 0

	for iteration < maxIterations {
		iteration++
		hasAdjustments := false
		currentBalance := initialBalance
		monthEnd := time.Date(year, time.Month(month+1), 0, 23, 59, 59, 0, time.UTC)

		// Проходим по транзакциям и проверяем баланс
		for i, tx := range sorted {
			if tx.IsExpense() && currentBalance+tx.Amount < 0 {
				// Недостаток баланса - корректируем
				requiredAmount := -tx.Amount
				shortage := -(currentBalance + tx.Amount)

				var adjustment *BalanceAdjustment
				var err error

				switch strategy {
				case StrategyPostpone:
					adjustment, err = s.adjustByPostponing(tx, sorted, currentBalance, i, dateCalculator, monthEnd)
				case StrategyReduce:
					adjustment, err = s.adjustByReducing(tx, currentBalance, requiredAmount, shortage)
				case StrategyHybrid:
					// Сначала пробуем перенос, если не получается - уменьшаем
					adjustment, err = s.adjustByPostponing(tx, sorted, currentBalance, i, dateCalculator, monthEnd)
					if err != nil {
						adjustment, err = s.adjustByReducing(tx, currentBalance, requiredAmount, shortage)
					}
				default:
					adjustment, err = s.adjustByPostponing(tx, sorted, currentBalance, i, dateCalculator, monthEnd)
				}

				if err != nil {
					return nil, nil, fmt.Errorf("%w %s: %w", helpers.ErrFailedToAdjustTransactions, tx.ID, err)
				}

				if adjustment != nil && adjustment.WasAdjusted {
					adjustments = append(adjustments, adjustment)
					hasAdjustments = true
					// Сохраняем информацию о корректировке в calculation_details
					s.saveAdjustmentToTransaction(tx, adjustment)
				}

				// Обновляем баланс с учетом скорректированной транзакции
				currentBalance += tx.Amount
				tx.SetBalanceAfter(utils.RoundToCents(currentBalance))
			} else {
				// Нормальная транзакция - просто обновляем баланс
				currentBalance += tx.Amount
				tx.SetBalanceAfter(utils.RoundToCents(currentBalance))
			}
		}

		// Если были корректировки, пересортировываем и пересчитываем балансы
		if hasAdjustments {
			sorted = utils.SortTransactionsByDate(sorted)
			// Пересчитываем балансы заново
			currentBalance = initialBalance
			for _, tx := range sorted {
				currentBalance += tx.Amount
				tx.SetBalanceAfter(utils.RoundToCents(currentBalance))
			}
		} else {
			// Нет корректировок - выходим из цикла
			break
		}
	}

	if iteration >= maxIterations {
		log.Warn("Balance adjustment reached max iterations: %d", maxIterations)
	}

	log.WithFields(logger.Fields{
		"adjustments_count": len(adjustments),
		"iterations":        iteration,
	}).Success("Transactions adjusted for balance")

	return sorted, adjustments, nil
}

// adjustByPostponing - корректировка путем переноса на следующий день
func (s *balanceAdjustmentService) adjustByPostponing(
	tx *entities.Transaction,
	allTransactions []*entities.Transaction,
	currentBalance float64,
	currentIndex int,
	dateCalculator *generatorservice.DateCalculator,
	monthEnd time.Time,
) (*BalanceAdjustment, error) {
	requiredAmount := -tx.Amount
	originalDate := tx.TransactionDate

	// Ищем следующую дату с достаточным балансом
	newDate, _, err := s.findNextAvailableDate(
		tx,
		allTransactions,
		currentBalance,
		currentIndex,
		dateCalculator,
		monthEnd,
		requiredAmount,
	)

	if err != nil {
		return nil, fmt.Errorf("%w: %w", helpers.ErrCannotPostponeTransaction, err)
	}

	// Переносим транзакцию на новую дату
	tx.TransactionDate = newDate
	tx.PostingDate = newDate

	originalAmount := tx.Amount
	adjustment := &BalanceAdjustment{
		TransactionID:                  tx.ID,
		WasAdjusted:                    true,
		AdjustmentType:                 "postponed",
		OriginalDate:                   &originalDate,
		OriginalAmount:                 &originalAmount, // Сохраняем оригинальную сумму для отслеживания
		AdjustmentReason:               fmt.Sprintf("Insufficient balance on %s. Postponed to %s after income transaction", originalDate.Format("2006-01-02"), newDate.Format("2006-01-02")),
		AvailableBalanceAtOriginalDate: currentBalance,
		RequiredBalance:                requiredAmount,
		Shortage:                       -(currentBalance + originalAmount),
	}

	return adjustment, nil
}

// adjustByReducing - корректировка путем уменьшения суммы
func (s *balanceAdjustmentService) adjustByReducing(
	tx *entities.Transaction,
	currentBalance float64,
	requiredAmount float64,
	shortage float64,
) (*BalanceAdjustment, error) {
	originalAmount := tx.Amount

	// Уменьшаем сумму до доступного баланса (оставляем 1% запас)
	newAmount := -currentBalance * 0.99
	newAmount = utils.RoundToCents(newAmount)

	// Проверяем, что новая сумма не слишком мала (минимум 1 цент)
	if newAmount > -0.01 {
		return nil, fmt.Errorf("%w (%.2f)", helpers.ErrCannotReduceTransactionAmount, currentBalance)
	}

	// Обновляем сумму транзакции
	tx.Amount = newAmount

	adjustment := &BalanceAdjustment{
		TransactionID:                  tx.ID,
		WasAdjusted:                    true,
		AdjustmentType:                 "reduced",
		OriginalDate:                   nil, // не переносили дату
		OriginalAmount:                 &originalAmount,
		AdjustmentReason:               fmt.Sprintf("Insufficient balance. Reduced amount from %.2f to %.2f", originalAmount, newAmount),
		AvailableBalanceAtOriginalDate: currentBalance,
		RequiredBalance:                requiredAmount,
		Shortage:                       shortage,
	}

	return adjustment, nil
}

// findNextAvailableDate - поиск следующей даты с достаточным балансом
func (s *balanceAdjustmentService) findNextAvailableDate(
	expenseTx *entities.Transaction,
	allTransactions []*entities.Transaction,
	currentBalance float64,
	currentIndex int,
	dateCalculator *dateservice.DateCalculator,
	monthEnd time.Time,
	requiredAmount float64,
) (time.Time, float64, error) {
	startDate := expenseTx.TransactionDate

	// Итерация по дням вперед (максимум до конца месяца)
	for date := startDate.AddDate(0, 0, 1); !date.After(monthEnd); date = date.AddDate(0, 0, 1) {
		// Пропускаем выходные и праздники
		if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
			continue
		}
		if dateCalculator.IsHoliday(date) {
			continue
		}

		// Рассчитываем баланс на эту дату
		balanceOnDate, err := s.calculateBalanceUpToDate(
			allTransactions,
			currentBalance,
			date,
			currentIndex, // не учитываем текущую транзакцию
		)
		if err != nil {
			continue
		}

		// Проверяем, достаточно ли средств
		if balanceOnDate >= requiredAmount {
			return date, balanceOnDate, nil
		}
	}

	return time.Time{}, 0, helpers.ErrNoAvailableDate
}

// calculateBalanceUpToDate - расчет баланса на определенную дату
func (s *balanceAdjustmentService) calculateBalanceUpToDate(
	allTransactions []*entities.Transaction,
	initialBalance float64,
	targetDate time.Time,
	excludeIndex int, // индекс транзакции, которую нужно исключить из расчета
) (float64, error) {
	balance := initialBalance

	for i, tx := range allTransactions {
		// Пропускаем транзакцию, которую корректируем
		if i == excludeIndex {
			continue
		}

		// Учитываем только транзакции до целевой даты (включительно)
		if tx.TransactionDate.After(targetDate) {
			break
		}

		balance += tx.Amount
	}

	return utils.RoundToCents(balance), nil
}

// saveAdjustmentToTransaction - сохранение информации о корректировке в calculation_details
func (s *balanceAdjustmentService) saveAdjustmentToTransaction(
	tx *entities.Transaction,
	adjustment *BalanceAdjustment,
) {
	if tx.CalculationDetails == nil {
		tx.CalculationDetails = make(map[string]interface{})
	}

	// Сохраняем информацию о корректировке
	tx.CalculationDetails["was_adjusted"] = adjustment.WasAdjusted
	tx.CalculationDetails["adjustment_type"] = adjustment.AdjustmentType
	if adjustment.OriginalDate != nil {
		tx.CalculationDetails["original_date"] = adjustment.OriginalDate.Format(time.RFC3339)
	}
	if adjustment.OriginalAmount != nil {
		tx.CalculationDetails["original_amount"] = *adjustment.OriginalAmount
	}
	tx.CalculationDetails["adjustment_reason"] = adjustment.AdjustmentReason
	tx.CalculationDetails["available_balance_at_original_date"] = adjustment.AvailableBalanceAtOriginalDate
	tx.CalculationDetails["required_balance"] = adjustment.RequiredBalance
	tx.CalculationDetails["shortage"] = adjustment.Shortage
}

// BalanceAndNormalize балансирует и нормализует суммы [42]
// [12][40] При необходимости сначала удаляет опциональные категории до нуля, прежде чем уменьшать основные расходы
func (s *balanceAdjustmentService) BalanceAndNormalize(transactions []*entities.Transaction, turnover, targetProfit float64, optionalCategories map[string]bool) ([]*entities.Transaction, error) {
	op := "service.balance.balanceAndNormalize"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"transactions_count": len(transactions),
		"turnover":          turnover,
		"target_profit":     targetProfit,
	})
	log.Info("Balancing and normalizing transactions")

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

	// [12][40] Если нужно увеличить прибыль (profitDiff > 0), сначала удаляем опциональные категории
	// [41] При генерации разных месяцев набор вырезаемых категорий может варьироваться для реалистичности (случайным образом)
	if profitDiff > 0.01 && optionalCategories != nil && len(optionalCategories) > 0 {
		// Группируем опциональные транзакции по категориям для случайного выбора категорий [41]
		optionalByCategory := make(map[string][]*entities.Transaction)
		mandatoryTransactions := make([]*entities.Transaction, 0)
		
		for _, tx := range transactions {
			if tx.IsExpense() && !tx.IsManualTransaction() {
				if optionalCategories[tx.Category] {
					optionalByCategory[tx.Category] = append(optionalByCategory[tx.Category], tx)
				} else {
					mandatoryTransactions = append(mandatoryTransactions, tx)
				}
			} else {
				mandatoryTransactions = append(mandatoryTransactions, tx)
			}
		}

		// [41] Случайно выбираем категории для удаления (для реалистичности между месяцами)
		categoriesToRemove := make(map[string]bool)
		categoryList := make([]string, 0, len(optionalByCategory))
		for cat := range optionalByCategory {
			categoryList = append(categoryList, cat)
		}
		
		// Перемешиваем категории для случайного выбора
		rand.Shuffle(len(categoryList), func(i, j int) {
			categoryList[i], categoryList[j] = categoryList[j], categoryList[i]
		})

		// Собираем все опциональные транзакции, сортируем по сумме (от меньших к большим)
		allOptionalTransactions := make([]*entities.Transaction, 0)
		for _, txs := range optionalByCategory {
			allOptionalTransactions = append(allOptionalTransactions, txs...)
		}

		// Сортируем опциональные транзакции по абсолютной сумме (от меньших к большим)
		for i := 0; i < len(allOptionalTransactions)-1; i++ {
			for j := i + 1; j < len(allOptionalTransactions); j++ {
				if math.Abs(allOptionalTransactions[i].Amount) > math.Abs(allOptionalTransactions[j].Amount) {
					allOptionalTransactions[i], allOptionalTransactions[j] = allOptionalTransactions[j], allOptionalTransactions[i]
				}
			}
		}

		// Удаляем опциональные транзакции до тех пор, пока не достигнем целевой прибыли или не закончатся опциональные
		removedAmount := 0.0
		remainingOptional := make([]*entities.Transaction, 0)
		
		for _, tx := range allOptionalTransactions {
			if removedAmount < profitDiff {
				// Удаляем эту транзакцию (не добавляем в remainingOptional)
				removedAmount += math.Abs(tx.Amount)
				categoriesToRemove[tx.Category] = true
				op := "service.balance.balanceAndNormalize"
				log := logger.GetLogger().WithOperation(op)
				log.Info("Removing optional category transaction: %s, amount=%.2f, removedAmount=%.2f, profitDiff=%.2f",
					tx.Category, tx.Amount, removedAmount, profitDiff)
			} else {
				// Оставляем эту транзакцию
				remainingOptional = append(remainingOptional, tx)
			}
		}

		// Обновляем список транзакций: оставляем только обязательные и не удаленные опциональные
		transactions = mandatoryTransactions
		transactions = append(transactions, remainingOptional...)

		// Пересчитываем итоги после удаления опциональных
		currentIncome = 0.0
		currentExpense = 0.0
		for _, tx := range transactions {
			if tx.IsIncome() {
				currentIncome += tx.Amount
			} else {
				currentExpense += tx.Amount
			}
		}
		currentProfit = currentIncome + currentExpense
		profitDiff = targetProfit - currentProfit
	}

	// [42] Корректируем доходы: сумма всех доходов должна равняться turnover
	if math.Abs(incomeDiff) > 0.01 {
		// Находим последнюю транзакцию дохода (не ручную)
		for i := len(transactions) - 1; i >= 0; i-- {
			if transactions[i].IsIncome() && !transactions[i].IsManualTransaction() {
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
				if tx.IsExpense() && !tx.IsManualTransaction() {
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
				if transactions[i].IsExpense() && !transactions[i].IsManualTransaction() {
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
			log.Debug("Income normalization iteration %d: target=%.2f, actual=%.2f, diff=%.2f", iteration+1, turnover, finalIncome, incomeError)
			// Находим последнюю транзакцию дохода (не ручную)
			for i := len(transactions) - 1; i >= 0; i-- {
				if transactions[i].IsIncome() && !transactions[i].IsManualTransaction() {
					incomeDiff := turnover - finalIncome
					transactions[i].Amount += incomeDiff
					transactions[i].Amount = utils.RoundToCents(transactions[i].Amount)
					break
				}
			}
		}

		// [42] Корректируем расходы: прибыль должна равняться targetProfit
		if profitError > 0.02 {
			log.Debug("Profit normalization iteration %d: target=%.2f, actual=%.2f, diff=%.2f", iteration+1, targetProfit, finalProfit, profitError)
			// Находим последнюю транзакцию расхода (не ручную)
			for i := len(transactions) - 1; i >= 0; i-- {
				if transactions[i].IsExpense() && !transactions[i].IsManualTransaction() {
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
		log.Warn("Income normalization final error: target=%.2f, actual=%.2f, diff=%.2f", turnover, finalIncome, incomeError)
	}
	if profitError > 0.05 {
		log.Warn("Profit normalization final error: target=%.2f, actual=%.2f, diff=%.2f", targetProfit, finalProfit, profitError)
	}

	log.WithFields(logger.Fields{
		"final_income":  finalIncome,
		"final_expense": finalExpense,
		"final_profit":  finalProfit,
		"income_error":  incomeError,
		"profit_error":  profitError,
	}).Success("Transactions balanced and normalized")

	return transactions, nil
}

// CalculateBalances рассчитывает балансы после каждой транзакции
func (s *balanceAdjustmentService) CalculateBalances(transactions []*entities.Transaction, initialBalance float64) ([]*entities.Transaction, error) {
	op := "service.balance.calculateBalances"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"transactions_count": len(transactions),
		"initial_balance":    initialBalance,
	})
	log.Debug("Calculating balances for transactions")

	currentBalance := initialBalance

	for _, tx := range transactions {
		// Проверяем, достаточно ли средств
		if tx.IsExpense() && currentBalance+tx.Amount < 0 {
			log.Error(nil, "Insufficient balance on %s: required %.2f, available %.2f",
				tx.TransactionDate.Format("2006-01-02"),
				-tx.Amount, currentBalance)
			return nil, fmt.Errorf("%w on %s: required %.2f, available %.2f",
				helpers.ErrInsufficientBalance,
				tx.TransactionDate.Format("2006-01-02"),
				-tx.Amount, currentBalance)
		}

		currentBalance += tx.Amount
		tx.SetBalanceAfter(utils.RoundToCents(currentBalance))
	}

	log.WithFields(logger.Fields{"final_balance": currentBalance}).Debug("Balances calculated successfully")
	return transactions, nil
}

// RecalculateBalances - пересчет балансов после корректировки транзакций
func (s *balanceAdjustmentService) RecalculateBalances(transactions []*entities.Transaction, initialBalance float64) ([]*entities.Transaction, error) {
	op := "service.balance.recalculateBalances"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"transactions_count": len(transactions),
		"initial_balance":    initialBalance,
	})
	log.Debug("Recalculating balances for transactions")

	currentBalance := initialBalance

	for _, tx := range transactions {
		currentBalance += tx.Amount
		tx.SetBalanceAfter(utils.RoundToCents(currentBalance))
	}

	log.WithFields(logger.Fields{"final_balance": currentBalance}).Debug("Balances recalculated successfully")
	return transactions, nil
}

