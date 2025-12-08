package helpers

import (
	"fmt"
	"regexp"
	"time"
)

type RequestValidator struct {
}

// NewRequestValidator - создает новый валидатор
func NewRequestValidator() *RequestValidator {
	return &RequestValidator{}
}

// ValidateRequest - полная валидация входного JSON
func (v *RequestValidator) ValidateRequest(req *StatementStateRequest) error {
	// 1. Валидация companyInfo
	if err := v.validateCompanyInfo(req.CompanyInfo); err != nil {
		return fmt.Errorf("company_info validation failed: %w", err)
	}

	// 2. Валидация financials
	if err := v.validateFinancials(req.Financials); err != nil {
		return fmt.Errorf("financials validation failed: %w", err)
	}

	// 3. Валидация customData
	if err := v.validateCustomData(req.CustomData); err != nil {
		return fmt.Errorf("custom_data validation failed: %w", err)
	}

	return nil
}

func (v *RequestValidator) validateCustomData(customData CustomData) error {
	// 1. Валидация manualIncomes
	for i, income := range customData.ManualIncomes {
		if err := v.validateManualIncome(income); err != nil {
			return fmt.Errorf("manualIncomes[%d]: %w", i, err)
		}
	}

	// 2. Валидация manualExpenses
	for i, expense := range customData.ManualExpenses {
		if err := v.validateManualExpense(expense); err != nil {
			return fmt.Errorf("manualExpenses[%d]: %w", i, err)
		}
	}

	// 3. Валидация customCustomers (максимум 20 элементов)
	if len(customData.CustomCustomers) > 20 {
		return fmt.Errorf("%w: customCustomers cannot exceed 20 elements, got %d", ErrInvalidLength, len(customData.CustomCustomers))
	}

	// 4. Валидация customContractors
	for i, contractor := range customData.CustomContractors {
		if contractor.TransactionType == "" {
			return fmt.Errorf("customContractors[%d]: transactionType %w", i, ErrRequired)
		}
		if contractor.Name == "" {
			return fmt.Errorf("customContractors[%d]: name %w", i, ErrRequired)
		}
	}

	// 5. Валидация disableCategories (опционально, проверка будет в бизнес-логике)
	// Здесь просто проверяем что это массив строк
	_ = customData.DisableCategories

	return nil
}

// validateManualIncome проверяет валидность ManualIncome
func (v *RequestValidator) validateManualIncome(income ManualIncome) error {
	// Валидация даты
	if err := v.validateDate(income.Date); err != nil {
		return fmt.Errorf("date: %w", err)
	}

	// Валидация суммы
	if income.Amount < 0 {
		return fmt.Errorf("%w: amount must be >= 0, got %.2f", ErrInvalidValue, income.Amount)
	}

	// Category опционально, но если указан, не должен быть пустым
	if income.Category != "" && len(income.Category) == 0 {
		return fmt.Errorf("%w: category cannot be empty if provided", ErrInvalidValue)
	}

	return nil
}

// validateManualExpense проверяет валидность ManualExpense
func (v *RequestValidator) validateManualExpense(expense ManualExpense) error {
	// Валидация даты
	if err := v.validateDate(expense.Date); err != nil {
		return fmt.Errorf("date: %w", err)
	}

	// Валидация суммы
	if expense.Amount < 0 {
		return fmt.Errorf("%w: amount must be >= 0, got %.2f", ErrInvalidValue, expense.Amount)
	}

	// Category обязателен для expense
	if expense.Category == "" {
		return fmt.Errorf("%w: category is required for manualExpense", ErrRequired)
	}

	return nil
}

// validateDate проверяет формат даты YYYY-MM-DD
func (v *RequestValidator) validateDate(date string) error {
	if date == "" {
		return fmt.Errorf("%w: date cannot be empty", ErrRequired)
	}

	// Проверка формата YYYY-MM-DD
	dateRegex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	if !dateRegex.MatchString(date) {
		return fmt.Errorf("%w: date must be in format YYYY-MM-DD, got %s", ErrInvalidFormat, date)
	}

	// Парсинг даты
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return fmt.Errorf("%w: invalid date format %s: %v", ErrInvalidFormat, date, err)
	}

	// Проверка что дата не в будущем (можно разрешить текущую дату)
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	requestedDate := time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, time.UTC)

	if requestedDate.After(today) {
		return fmt.Errorf("%w: date cannot be in the future, got %s", ErrInvalidDate, date)
	}

	return nil
}

func (v *RequestValidator) validateFinancials(financials Financials) error {
	fmt.Println("--------------------------------")
	fmt.Println("--------------------------------")
	fmt.Println("financials: ", financials)
	fmt.Println("--------------------------------")
	fmt.Println("--------------------------------")
	// 1. Валидация startBalance
	if financials.StartBalance < 0 {
		return fmt.Errorf("%w: startBalance must be >= 0, got %.2f", ErrInvalidValue, financials.StartBalance)
	}

	// 2. Валидация turnover
	if financials.Turnover <= 0 {
		return fmt.Errorf("%w: turnover must be > 0, got %.2f", ErrInvalidValue, financials.Turnover)
	}

	// 3. Валидация profitPercent (0-50%)
	if financials.ProfitPercent < 0 || financials.ProfitPercent > 50 {
		return fmt.Errorf("%w: profitPercent must be between 0 and 50, got %.2f", ErrInvalidValue, financials.ProfitPercent)
	}

	// 4. Валидация targetProfit
	if financials.TargetProfit < 0 {
		return fmt.Errorf("%w: targetProfit must be >= 0, got %.2f", ErrInvalidValue, financials.TargetProfit)
	}

	// 5. Проверка что targetProfit не превышает turnover
	if financials.TargetProfit > 0 && financials.TargetProfit > financials.Turnover {
		return fmt.Errorf("%w: targetProfit (%.2f) cannot exceed turnover (%.2f)", ErrInvalidValue, financials.TargetProfit, financials.Turnover)
	}

	// 6. Проверка что указан либо periods, либо months + startMonth
	hasPeriods := len(financials.Periods) > 0
	hasMonthsAndStartMonth := financials.Months > 0 && financials.StartMonth != ""

	if !hasPeriods && !hasMonthsAndStartMonth {
		return fmt.Errorf("%w: either 'periods' or 'months + startMonth' must be specified", ErrRequired)
	}

	if hasPeriods && hasMonthsAndStartMonth {
		// Если указаны оба варианта, приоритет у periods - игнорируем months и startMonth
		// Не возвращаем ошибку, просто используем periods
	}

	// 7. Валидация months (1-36) - только если указан months
	if hasMonthsAndStartMonth {
		if financials.Months < 1 || financials.Months > 36 {
			return fmt.Errorf("%w: months must be between 1 and 36, got %d", ErrInvalidValue, financials.Months)
		}
	}

	// 8. Валидация startMonth (формат YYYY-MM) - только если указан startMonth
	if hasMonthsAndStartMonth {
		if err := v.validateMonth(financials.StartMonth); err != nil {
			return fmt.Errorf("startMonth: %w", err)
		}
	}

	// 9. Валидация periods (если указан)
	if hasPeriods {
		for i, period := range financials.Periods {
			if err := v.validateMonth(period); err != nil {
				return fmt.Errorf("periods[%d]: %w", i, err)
			}
		}
	}

	// 10. Валидация operationMultiplier (0-5, но обычно > 0)
	if financials.OperationMultiplier <= 0 || financials.OperationMultiplier > 5 {
		return fmt.Errorf("%w: operationMultiplier must be between 0 and 5, got %.2f", ErrInvalidValue, financials.OperationMultiplier)
	}

	// 11. Дополнительная бизнес-валидация: проверка достаточности баланса
	// Если указаны и profitPercent и turnover, проверяем что баланс достаточен
	if financials.ProfitPercent > 0 && financials.Turnover > 0 {
		targetProfit := financials.Turnover * (financials.ProfitPercent / 100)
		targetExpenses := financials.Turnover - targetProfit
		minBalance := targetExpenses * 0.5 // Минимум 50% от расходов

		if financials.StartBalance < minBalance {
			return fmt.Errorf("%w: startBalance (%.2f) too low for estimated expenses (%.2f), minimum %.2f recommended",
				ErrInsufficientBalance, financials.StartBalance, targetExpenses, minBalance)
		}
	}

	return nil
}

// validateMonth проверяет формат месяца YYYY-MM и что месяц не в будущем
func (v *RequestValidator) validateMonth(month string) error {
	if month == "" {
		return fmt.Errorf("%w: month cannot be empty", ErrRequired)
	}

	// Проверка формата YYYY-MM
	monthRegex := regexp.MustCompile(`^\d{4}-\d{2}$`)
	if !monthRegex.MatchString(month) {
		return fmt.Errorf("%w: month must be in format YYYY-MM, got %s", ErrInvalidFormat, month)
	}

	// Парсинг месяца
	parsedMonth, err := time.Parse("2006-01", month)
	if err != nil {
		return fmt.Errorf("%w: invalid month format %s: %v", ErrInvalidFormat, month, err)
	}

	// Проверка что месяц не в будущем
	now := time.Now()
	currentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	requestedMonth := time.Date(parsedMonth.Year(), parsedMonth.Month(), 1, 0, 0, 0, 0, time.UTC)

	if requestedMonth.After(currentMonth) {
		return fmt.Errorf("%w: cannot generate statement for future month %s", ErrFutureMonth, month)
	}

	return nil
}

func (v *RequestValidator) validateCompanyInfo(companyInfo CompanyInfo) error {
	var companyFields = []string{
		companyInfo.CompanyName,
		companyInfo.OwnerName,
		companyInfo.AccountNumber,
		companyInfo.AssociatedCard,
		companyInfo.Model,
		companyInfo.State,
		companyInfo.Industry,
	}
	for _, field := range companyFields {
		if field == "" {
			return ErrRequired
		}
	}
	return nil
}

// Custom errors

// formatValidationErrors - форматирование ошибок валидации
func (v *RequestValidator) FormatValidationErrors(err error) interface{} {
	return err.Error()
}

func GenerateErrorResponse(errorType error, message error, details string) ErrorResponse {
	return ErrorResponse{
		Error:   errorType,
		Message: message, // Общее сообщение (константа)
		Details: details, // Детали из err.Error() - без дублирования
	}
}
