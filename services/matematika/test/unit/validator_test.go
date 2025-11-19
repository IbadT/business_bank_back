package calculation_test

import (
	"testing"

	"github.com/IbadT/business_bank_back/services/matematika/internal/helpers"
	"github.com/IbadT/business_bank_back/services/matematika/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ================================================
// TestValidateRequest проверяет работу валидатора напрямую
// ================================================
func TestValidateRequest(t *testing.T) {
	tests := []struct {
		name          string
		request       helpers.StatementStateRequest
		shouldFail    bool
		expectedError string
		description   string
	}{
		// Успешный случай
		{
			name:        "Valid request - should pass validation",
			request:     mocks.GenerateStatementMocks(),
			shouldFail:  false,
			description: "Валидный запрос должен пройти валидацию",
		},
		// Валидация CompanyInfo
		{
			name:          "Missing company name - should fail",
			request:       mocks.MockMissingCompanyName(),
			shouldFail:    true,
			expectedError: "company_info validation failed",
			description:   "Отсутствие companyName должно вызвать ошибку валидации",
		},
		{
			name:          "Missing owner name - should fail",
			request:       mocks.MockMissingOwnerName(),
			shouldFail:    true,
			expectedError: "company_info validation failed",
			description:   "Отсутствие ownerName должно вызвать ошибку валидации",
		},
		{
			name:          "Missing account number - should fail",
			request:       mocks.MockMissingAccountNumber(),
			shouldFail:    true,
			expectedError: "company_info validation failed",
			description:   "Отсутствие accountNumber должно вызвать ошибку валидации",
		},
		{
			name:          "Missing associated card - should fail",
			request:       mocks.MockMissingAssociatedCard(),
			shouldFail:    true,
			expectedError: "company_info validation failed",
			description:   "Отсутствие associatedCard должно вызвать ошибку валидации",
		},
		{
			name:          "Missing model - should fail",
			request:       mocks.MockMissingModel(),
			shouldFail:    true,
			expectedError: "company_info validation failed",
			description:   "Отсутствие model должно вызвать ошибку валидации",
		},
		{
			name:          "Missing state - should fail",
			request:       mocks.MockMissingState(),
			shouldFail:    true,
			expectedError: "company_info validation failed",
			description:   "Отсутствие state должно вызвать ошибку валидации",
		},
		{
			name:          "Missing industry - should fail",
			request:       mocks.MockMissingIndustry(),
			shouldFail:    true,
			expectedError: "company_info validation failed",
			description:   "Отсутствие industry должно вызвать ошибку валидации",
		},
		// Валидация Financials
		{
			name:          "Negative start balance - should fail",
			request:       mocks.MockNegativeStartBalance(),
			shouldFail:    true,
			expectedError: "financials validation failed",
			description:   "Отрицательный startBalance должен вызвать ошибку валидации",
		},
		{
			name:          "Zero turnover - should fail",
			request:       mocks.MockZeroTurnover(),
			shouldFail:    true,
			expectedError: "financials validation failed",
			description:   "Нулевой turnover должен вызвать ошибку валидации",
		},
		{
			name:          "Negative turnover - should fail",
			request:       mocks.MockNegativeTurnover(),
			shouldFail:    true,
			expectedError: "financials validation failed",
			description:   "Отрицательный turnover должен вызвать ошибку валидации",
		},
		{
			name:          "Invalid profit percent - should fail",
			request:       mocks.MockInvalidProfitPercent(),
			shouldFail:    true,
			expectedError: "financials validation failed",
			description:   "profitPercent > 50% должен вызвать ошибку валидации",
		},
		{
			name:          "Negative profit percent - should fail",
			request:       mocks.MockNegativeProfitPercent(),
			shouldFail:    true,
			expectedError: "financials validation failed",
			description:   "Отрицательный profitPercent должен вызвать ошибку валидации",
		},
		{
			name:          "Target profit exceeds turnover - should fail",
			request:       mocks.MockTargetProfitExceedsTurnover(),
			shouldFail:    true,
			expectedError: "financials validation failed",
			description:   "targetProfit > turnover должен вызвать ошибку валидации",
		},
		{
			name:          "Negative target profit - should fail",
			request:       mocks.MockNegativeTargetProfit(),
			shouldFail:    true,
			expectedError: "financials validation failed",
			description:   "Отрицательный targetProfit должен вызвать ошибку валидации",
		},
		{
			name:          "Invalid months - should fail",
			request:       mocks.MockInvalidMonths(),
			shouldFail:    true,
			expectedError: "financials validation failed",
			description:   "months > 36 должен вызвать ошибку валидации",
		},
		{
			name:          "Zero months - should fail",
			request:       mocks.MockZeroMonths(),
			shouldFail:    true,
			expectedError: "financials validation failed",
			description:   "months = 0 должен вызвать ошибку валидации",
		},
		{
			name:          "Future month - should fail",
			request:       mocks.MockFutureMonth(),
			shouldFail:    true,
			expectedError: "financials validation failed",
			description:   "Месяц в будущем должен вызвать ошибку валидации",
		},
		{
			name:          "Invalid month format - should fail",
			request:       mocks.MockInvalidMonthFormat(),
			shouldFail:    true,
			expectedError: "financials validation failed",
			description:   "Неправильный формат месяца должен вызвать ошибку валидации",
		},
		{
			name:          "Empty month - should fail",
			request:       mocks.MockEmptyMonth(),
			shouldFail:    true,
			expectedError: "financials validation failed",
			description:   "Пустой месяц должен вызвать ошибку валидации",
		},
		{
			name:          "Invalid operation multiplier - should fail",
			request:       mocks.MockInvalidOperationMultiplier(),
			shouldFail:    true,
			expectedError: "financials validation failed",
			description:   "operationMultiplier > 5 должен вызвать ошибку валидации",
		},
		{
			name:          "Zero operation multiplier - should fail",
			request:       mocks.MockZeroOperationMultiplier(),
			shouldFail:    true,
			expectedError: "financials validation failed",
			description:   "operationMultiplier = 0 должен вызвать ошибку валидации",
		},
		{
			name:          "Negative operation multiplier - should fail",
			request:       mocks.MockNegativeOperationMultiplier(),
			shouldFail:    true,
			expectedError: "financials validation failed",
			description:   "Отрицательный operationMultiplier должен вызвать ошибку валидации",
		},
		{
			name:          "Both periods and months specified - should fail",
			request:       mocks.MockBothPeriodsAndMonths(),
			shouldFail:    true,
			expectedError: "financials validation failed",
			description:   "Указание и periods и months одновременно должно вызвать ошибку",
		},
		{
			name:          "Neither periods nor months specified - should fail",
			request:       mocks.MockNeitherPeriodsNorMonths(),
			shouldFail:    true,
			expectedError: "financials validation failed",
			description:   "Отсутствие и periods и months должно вызвать ошибку",
		},
		{
			name:          "Insufficient balance - should fail",
			request:       mocks.MockInsufficientBalance(),
			shouldFail:    true,
			expectedError: "financials validation failed",
			description:   "Недостаточный баланс должен вызвать ошибку валидации",
		},
		{
			name:          "Invalid periods format - should fail",
			request:       mocks.MockInvalidPeriodsFormat(),
			shouldFail:    true,
			expectedError: "financials validation failed",
			description:   "Неправильный формат в periods должен вызвать ошибку валидации",
		},
		{
			name:          "Future periods - should fail",
			request:       mocks.MockFuturePeriods(),
			shouldFail:    true,
			expectedError: "financials validation failed",
			description:   "Будущие месяцы в periods должны вызвать ошибку валидации",
		},
		// Валидация CustomData
		{
			name:          "Invalid manual income date - should fail",
			request:       mocks.MockInvalidManualIncomeDate(),
			shouldFail:    true,
			expectedError: "custom_data validation failed",
			description:   "Неправильная дата в manualIncome должна вызвать ошибку валидации",
		},
		{
			name:          "Negative manual income amount - should fail",
			request:       mocks.MockNegativeManualIncomeAmount(),
			shouldFail:    true,
			expectedError: "custom_data validation failed",
			description:   "Отрицательная сумма в manualIncome должна вызвать ошибку валидации",
		},
		{
			name:          "Invalid manual expense date - should fail",
			request:       mocks.MockInvalidManualExpenseDate(),
			shouldFail:    true,
			expectedError: "custom_data validation failed",
			description:   "Неправильная дата в manualExpense должна вызвать ошибку валидации",
		},
		{
			name:          "Negative manual expense amount - should fail",
			request:       mocks.MockNegativeManualExpenseAmount(),
			shouldFail:    true,
			expectedError: "custom_data validation failed",
			description:   "Отрицательная сумма в manualExpense должна вызвать ошибку валидации",
		},
		{
			name:          "Missing manual expense category - should fail",
			request:       mocks.MockMissingManualExpenseCategory(),
			shouldFail:    true,
			expectedError: "custom_data validation failed",
			description:   "Отсутствие category в manualExpense должно вызвать ошибку валидации",
		},
		{
			name:          "Too many custom customers - should fail",
			request:       mocks.MockTooManyCustomCustomers(),
			shouldFail:    true,
			expectedError: "custom_data validation failed",
			description:   "Слишком много customCustomers (> 20) должно вызвать ошибку валидации",
		},
		{
			name:          "Empty custom contractor name - should fail",
			request:       mocks.MockEmptyCustomContractorName(),
			shouldFail:    true,
			expectedError: "custom_data validation failed",
			description:   "Пустое имя в customContractor должно вызвать ошибку валидации",
		},
		{
			name:          "Empty custom contractor type - should fail",
			request:       mocks.MockEmptyCustomContractorType(),
			shouldFail:    true,
			expectedError: "custom_data validation failed",
			description:   "Пустой тип в customContractor должен вызвать ошибку валидации",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			validator := helpers.NewRequestValidator()

			// Act
			err := validator.ValidateRequest(&tt.request)

			// Assert
			if tt.shouldFail {
				// Должна быть ошибка валидации
				require.Error(t, err, tt.description)
				assert.Contains(t, err.Error(), tt.expectedError, tt.description)
			} else {
				// Не должно быть ошибки
				require.NoError(t, err, tt.description)
			}
		})
	}
}
