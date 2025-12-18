package service

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/internal/domain"
	"github.com/IbadT/business_bank_back/services/matematika/internal/domain/entities"
	"github.com/IbadT/business_bank_back/services/matematika/internal/repository"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/utils"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
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
		dateCalculator *dateCalculator,
		year, month int,
	) ([]*entities.Transaction, []*BalanceAdjustment, error)
}

type balanceAdjustmentService struct {
	transactionRepo       repository.TransactionRepository
	transactionService    TransactionService
	generationRequestRepo repository.GenerationRequestRepository
}

func NewBalanceAdjustmentService(transactionRepo repository.TransactionRepository, transactionService TransactionService, generationRequestRepo repository.GenerationRequestRepository) BalanceAdjustmentService {
	return &balanceAdjustmentService{
		transactionRepo:       transactionRepo,
		transactionService:    transactionService,
		generationRequestRepo: generationRequestRepo,
	}
}

func (s *balanceAdjustmentService) GetAdjustedTransactions(requestIDStr string) ([]domain.GeneratedTransaction, error) {
	requestID, err := uuid.Parse(requestIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid requestID format: %w", err)
	}

	if requestID == uuid.Nil {
		return nil, errors.New("requestID cannot be empty")
	}

	// Получаем модели из БД
	ormTransactions, err := s.transactionRepo.GetAdjustedTransactionsByRequestID(requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to get adjusted transactions: %w", err)
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

	return domainTransactions, nil
}

// ValidateBalance - валидация баланса транзакций
func (s *balanceAdjustmentService) ValidateBalance(requestIDStr string) (*ValidateBalanceResult, error) {
	// Парсим request_id
	requestID, err := uuid.Parse(requestIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid requestID format: %w", err)
	}

	// Получаем GenerationRequest для получения initial_balance
	genRequest, err := s.generationRequestRepo.GetByID(requestID)
	if err != nil {
		return nil, fmt.Errorf("generation request not found: %w", err)
	}

	// Получаем транзакции по request_id (уже отсортированы по дате в БД)
	transactions, err := s.transactionService.GetByRequestID(requestIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}

	if len(transactions) == 0 {
		return nil, fmt.Errorf("no transactions found for request_id: %s", requestIDStr)
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
	dateCalculator *dateCalculator,
	year, month int,
) ([]*entities.Transaction, []*BalanceAdjustment, error) {
	if len(transactions) == 0 {
		return transactions, []*BalanceAdjustment{}, nil
	}

	// Сортируем транзакции по дате
	sorted := s.sortTransactionsByDate(transactions)
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
					return nil, nil, fmt.Errorf("failed to adjust transaction %s: %w", tx.ID, err)
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
			sorted = s.sortTransactionsByDate(sorted)
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
		logrus.Warnf("[WARN] Balance adjustment reached max iterations (%d)", maxIterations)
	}

	return sorted, adjustments, nil
}

// adjustByPostponing - корректировка путем переноса на следующий день
func (s *balanceAdjustmentService) adjustByPostponing(
	tx *entities.Transaction,
	allTransactions []*entities.Transaction,
	currentBalance float64,
	currentIndex int,
	dateCalculator *dateCalculator,
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
		return nil, fmt.Errorf("cannot postpone transaction: %w", err)
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
		return nil, fmt.Errorf("cannot reduce transaction amount: available balance too low (%.2f)", currentBalance)
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
	dateCalculator *dateCalculator,
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
		if dateCalculator.isHoliday(date) {
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

	return time.Time{}, 0, errors.New("no available date found within month")
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

// sortTransactionsByDate - сортировка транзакций по дате
func (s *balanceAdjustmentService) sortTransactionsByDate(transactions []*entities.Transaction) []*entities.Transaction {
	// Создаем копию для сортировки
	sorted := make([]*entities.Transaction, len(transactions))
	copy(sorted, transactions)

	// Используем sort.Slice для эффективной сортировки O(n log n)
	sort.Slice(sorted, func(i, j int) bool {
		// Сортируем по TransactionDate (время совершения транзакции)
		// Если даты одинаковые, используем ID для стабильной сортировки
		if sorted[i].TransactionDate.Equal(sorted[j].TransactionDate) {
			return sorted[i].ID < sorted[j].ID
		}
		return sorted[i].TransactionDate.Before(sorted[j].TransactionDate)
	})

	return sorted
}
