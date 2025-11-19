package calculation_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/internal/calculation"
	"github.com/IbadT/business_bank_back/services/matematika/internal/helpers"
	"github.com/IbadT/business_bank_back/services/matematika/test/mocks"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockCalculationService - mock для CalculationService
type mockCalculationService struct{}

func (m *mockCalculationService) HealthCheck(ctx context.Context) (*helpers.HealthCheckResponse, error) {
	return &helpers.HealthCheckResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   "1.0.0",
		Dependencies: helpers.HealthCheckDependencies{
			Kafka:        "connected",
			Database:     "connected",
			ConfigLoaded: true,
			Service:      "matematika",
		},
	}, nil
}

func (m *mockCalculationService) GetTransactions(ctx context.Context) ([]calculation.Transaction, error) {
	return nil, nil
}

func (m *mockCalculationService) GetBusinessRules(ctx context.Context) ([]calculation.BusinessRule, error) {
	return nil, nil
}

func (m *mockCalculationService) GetDailyBalances(ctx context.Context) ([]calculation.DailyBalance, error) {
	return nil, nil
}

func (m *mockCalculationService) GetStatements(ctx context.Context) ([]calculation.Statement, error) {
	return nil, nil
}

// ================================================
// TestGenerateStatement проверяет что generate-statement endpoint возвращает корректный ответ
// ================================================
func TestGenerateStatement(t *testing.T) {
	// Arrange
	e := echo.New()
	mockRightRequest := mocks.GenerateStatementMocks()
	jsonData, err := json.Marshal(mockRightRequest)
	require.NoError(t, err)
	rec := httptest.NewRecorder()

	req := httptest.NewRequest(http.MethodPost, "/generate-statement", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	mockService := &mockCalculationService{}
	c := e.NewContext(req, rec)
	handler := calculation.NewCalculationHandler(mockService)

	// Act
	err = handler.GenerateStatement(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, rec.Code)

	var response calculation.GenerateStatementResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.NotNil(t, response.Transactions)
	assert.NotNil(t, response.DailyClosingBalances)
}

// ================================================
// TestGenerateStatementValidation проверяет валидацию входных данных
// ================================================
func TestGenerateStatementValidation(t *testing.T) {
	tests := []struct {
		name           string
		request        helpers.StatementStateRequest
		expectedStatus int
		expectedError  string
		description    string
	}{
		// Успешный случай
		{
			name:           "Valid request - should pass validation",
			request:        mocks.GenerateStatementMocks(),
			expectedStatus: http.StatusAccepted,
			description:    "Валидный запрос должен пройти валидацию",
		},
		// Валидация CompanyInfo
		{
			name:           "Missing company name - should fail",
			request:        mocks.MockMissingCompanyName(),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "company_info validation failed",
			description:    "Отсутствие companyName должно вызвать ошибку валидации",
		},
		{
			name:           "Missing owner name - should fail",
			request:        mocks.MockMissingOwnerName(),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "company_info validation failed",
			description:    "Отсутствие ownerName должно вызвать ошибку валидации",
		},
		{
			name:           "Missing account number - should fail",
			request:        mocks.MockMissingAccountNumber(),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "company_info validation failed",
			description:    "Отсутствие accountNumber должно вызвать ошибку валидации",
		},
		{
			name:           "Missing associated card - should fail",
			request:        mocks.MockMissingAssociatedCard(),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "company_info validation failed",
			description:    "Отсутствие associatedCard должно вызвать ошибку валидации",
		},
		{
			name:           "Missing model - should fail",
			request:        mocks.MockMissingModel(),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "company_info validation failed",
			description:    "Отсутствие model должно вызвать ошибку валидации",
		},
		{
			name:           "Missing state - should fail",
			request:        mocks.MockMissingState(),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "company_info validation failed",
			description:    "Отсутствие state должно вызвать ошибку валидации",
		},
		{
			name:           "Missing industry - should fail",
			request:        mocks.MockMissingIndustry(),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "company_info validation failed",
			description:    "Отсутствие industry должно вызвать ошибку валидации",
		},
		// Валидация Financials
		{
			name:           "Negative start balance - should fail",
			request:        mocks.MockNegativeStartBalance(),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "financials validation failed",
			description:    "Отрицательный startBalance должен вызвать ошибку валидации",
		},
		{
			name:           "Zero turnover - should fail",
			request:        mocks.MockZeroTurnover(),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "financials validation failed",
			description:    "Нулевой turnover должен вызвать ошибку валидации",
		},
		{
			name:           "Negative turnover - should fail",
			request:        mocks.MockNegativeTurnover(),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "financials validation failed",
			description:    "Отрицательный turnover должен вызвать ошибку валидации",
		},
		{
			name:           "Invalid profit percent - should fail",
			request:        mocks.MockInvalidProfitPercent(),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "financials validation failed",
			description:    "profitPercent > 50% должен вызвать ошибку валидации",
		},
		{
			name:           "Negative profit percent - should fail",
			request:        mocks.MockNegativeProfitPercent(),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "financials validation failed",
			description:    "Отрицательный profitPercent должен вызвать ошибку валидации",
		},
		{
			name:           "Target profit exceeds turnover - should fail",
			request:        mocks.MockTargetProfitExceedsTurnover(),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "financials validation failed",
			description:    "targetProfit > turnover должен вызвать ошибку валидации",
		},
		{
			name:           "Negative target profit - should fail",
			request:        mocks.MockNegativeTargetProfit(),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "financials validation failed",
			description:    "Отрицательный targetProfit должен вызвать ошибку валидации",
		},
		{
			name:           "Invalid months - should fail",
			request:        mocks.MockInvalidMonths(),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "financials validation failed",
			description:    "months > 36 должен вызвать ошибку валидации",
		},
		{
			name:           "Zero months - should fail",
			request:        mocks.MockZeroMonths(),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "financials validation failed",
			description:    "months = 0 должен вызвать ошибку валидации",
		},
		{
			name:           "Future month - should fail",
			request:        mocks.MockFutureMonth(),
			expectedStatus: http.StatusUnprocessableEntity, // 422 - специальная ошибка валидации
			expectedError:  "cannot generate statement for future month",
			description:    "Месяц в будущем должен вызвать ошибку валидации",
		},
		{
			name:           "Invalid month format - should fail",
			request:        mocks.MockInvalidMonthFormat(),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "financials validation failed",
			description:    "Неправильный формат месяца должен вызвать ошибку валидации",
		},
		{
			name:           "Empty month - should fail",
			request:        mocks.MockEmptyMonth(),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "financials validation failed",
			description:    "Пустой месяц должен вызвать ошибку валидации",
		},
		{
			name:           "Invalid operation multiplier - should fail",
			request:        mocks.MockInvalidOperationMultiplier(),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "financials validation failed",
			description:    "operationMultiplier > 5 должен вызвать ошибку валидации",
		},
		{
			name:           "Zero operation multiplier - should fail",
			request:        mocks.MockZeroOperationMultiplier(),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "financials validation failed",
			description:    "operationMultiplier = 0 должен вызвать ошибку валидации",
		},
		{
			name:           "Negative operation multiplier - should fail",
			request:        mocks.MockNegativeOperationMultiplier(),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "financials validation failed",
			description:    "Отрицательный operationMultiplier должен вызвать ошибку валидации",
		},
		{
			name:           "Both periods and months specified - should fail",
			request:        mocks.MockBothPeriodsAndMonths(),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "financials validation failed",
			description:    "Указание и periods и months одновременно должно вызвать ошибку",
		},
		{
			name:           "Neither periods nor months specified - should fail",
			request:        mocks.MockNeitherPeriodsNorMonths(),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "financials validation failed",
			description:    "Отсутствие и periods и months должно вызвать ошибку",
		},
		{
			name:           "Insufficient balance - should fail",
			request:        mocks.MockInsufficientBalance(),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "financials validation failed",
			description:    "Недостаточный баланс должен вызвать ошибку валидации",
		},
		{
			name:           "Invalid periods format - should fail",
			request:        mocks.MockInvalidPeriodsFormat(),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "financials validation failed",
			description:    "Неправильный формат в periods должен вызвать ошибку валидации",
		},
		{
			name:           "Future periods - should fail",
			request:        mocks.MockFuturePeriods(),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "financials validation failed",
			description:    "Будущие месяцы в periods должны вызвать ошибку валидации",
		},
		// Валидация CustomData
		{
			name:           "Invalid manual income date - should fail",
			request:        mocks.MockInvalidManualIncomeDate(),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "custom_data validation failed",
			description:    "Неправильная дата в manualIncome должна вызвать ошибку валидации",
		},
		{
			name:           "Negative manual income amount - should fail",
			request:        mocks.MockNegativeManualIncomeAmount(),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "custom_data validation failed",
			description:    "Отрицательная сумма в manualIncome должна вызвать ошибку валидации",
		},
		{
			name:           "Invalid manual expense date - should fail",
			request:        mocks.MockInvalidManualExpenseDate(),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "custom_data validation failed",
			description:    "Неправильная дата в manualExpense должна вызвать ошибку валидации",
		},
		{
			name:           "Negative manual expense amount - should fail",
			request:        mocks.MockNegativeManualExpenseAmount(),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "custom_data validation failed",
			description:    "Отрицательная сумма в manualExpense должна вызвать ошибку валидации",
		},
		{
			name:           "Missing manual expense category - should fail",
			request:        mocks.MockMissingManualExpenseCategory(),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "custom_data validation failed",
			description:    "Отсутствие category в manualExpense должно вызвать ошибку валидации",
		},
		{
			name:           "Too many custom customers - should fail",
			request:        mocks.MockTooManyCustomCustomers(),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "custom_data validation failed",
			description:    "Слишком много customCustomers (> 20) должно вызвать ошибку валидации",
		},
		{
			name:           "Empty custom contractor name - should fail",
			request:        mocks.MockEmptyCustomContractorName(),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "custom_data validation failed",
			description:    "Пустое имя в customContractor должно вызвать ошибку валидации",
		},
		{
			name:           "Empty custom contractor type - should fail",
			request:        mocks.MockEmptyCustomContractorType(),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "custom_data validation failed",
			description:    "Пустой тип в customContractor должен вызвать ошибку валидации",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			e := echo.New()
			jsonData, err := json.Marshal(tt.request)
			require.NoError(t, err)
			rec := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodPost, "/generate-statement", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			mockService := &mockCalculationService{}
			c := e.NewContext(req, rec)
			handler := calculation.NewCalculationHandler(mockService)

			// Act
			err = handler.GenerateStatement(c)

			// Assert
			if tt.expectedStatus == http.StatusAccepted {
				// Успешный случай
				require.NoError(t, err, tt.description)
				assert.Equal(t, tt.expectedStatus, rec.Code, tt.description)

				var response helpers.GenerateStatementResponse
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotEmpty(t, response.StatementID, tt.description)
			} else {
				// Ошибка валидации
				require.NoError(t, err, "Handler should not return error, error should be in response")
				assert.Equal(t, tt.expectedStatus, rec.Code, tt.description)

				// Проверяем что в ответе есть информация об ошибке
				var errorResponse helpers.ErrorResponse
				err = json.Unmarshal(rec.Body.Bytes(), &errorResponse)
				if err == nil {
					// Если удалось распарсить ErrorResponse, проверяем что ошибка связана с валидацией
					errorStr := fmt.Sprintf("%v", errorResponse.Message)
					assert.Contains(t, errorStr, tt.expectedError, tt.description)
				} else {
					// Если не ErrorResponse, проверяем что тело ответа содержит информацию об ошибке
					bodyStr := rec.Body.String()
					assert.Contains(t, bodyStr, tt.expectedError, tt.description)
				}
			}
		})
	}
}

// ================================================
// TestHealthCheck проверяет что health endpoint возвращает корректный ответ
// ================================================
func TestHealthCheck(t *testing.T) {
	// Arrange
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Создаем handler с mock service
	mockService := &mockCalculationService{}
	handler := calculation.NewCalculationHandler(mockService)

	// Act
	err := handler.HealthCheck(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Проверяем структуру ответа
	var response helpers.HealthCheckResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Проверяем поля ответа
	assert.Equal(t, "healthy", response.Status)
	assert.Equal(t, "1.0.0", response.Version)
	// Database может быть "connected" или "disconnected" в зависимости от наличия БД
	assert.Contains(t, []string{"connected", "disconnected"}, response.Dependencies.Database)
	// Kafka может быть "connected" или "disconnected" в зависимости от наличия Kafka
	assert.Contains(t, []string{"connected", "disconnected"}, response.Dependencies.Kafka)
	assert.True(t, response.Dependencies.ConfigLoaded)
	assert.Equal(t, "matematika", response.Dependencies.Service)

	// Проверяем что timestamp валидный RFC3339 формат
	_, err = time.Parse(time.RFC3339, response.Timestamp)
	assert.NoError(t, err, "Timestamp should be in RFC3339 format")
}

// ================================================
// GenerateStatementToKafka проверяет что kafka statement endpoint возвращает корректный ответ
// ================================================
func (m *mockCalculationService) GenerateStatementToKafka(ctx context.Context, req *helpers.GenerateStatementRequest) (*helpers.GenerateStatementResponse, error) {
	return nil, nil
}

// ================================================
// GenerateStatement проверяет что statement endpoint возвращает корректный ответ
// ================================================
func (m *mockCalculationService) GenerateStatement(ctx context.Context, req *helpers.StatementStateRequest) (*calculation.GenerateStatementResponse, error) {
	// Возвращаем валидный response для теста
	return &calculation.GenerateStatementResponse{
		Transactions:         []calculation.Transaction{},
		DailyClosingBalances: []calculation.DailyBalance{},
	}, nil
}

// ================================================
// GetStatementStatusByID проверяет что status endpoint возвращает корректный ответ
// ================================================
func (m *mockCalculationService) GetStatementStatusByID(ctx context.Context, id string) (interface{}, error) {
	return nil, nil
}

// ================================================
// GetStatementResultByID проверяет что result endpoint возвращает корректный ответ
// ================================================
func (m *mockCalculationService) GetStatementResultByID(ctx context.Context, id string) (interface{}, error) {
	return nil, nil
}

// ================================================
// StartConsumer проверяет что consumer запускается
// ================================================
func (m *mockCalculationService) StartConsumer(ctx context.Context) error {
	return nil
}

// ================================================
// GetAdminConfig проверяет что admin config endpoint возвращает корректный ответ
// ================================================
func (m *mockCalculationService) GetAdminConfig(ctx context.Context) (*helpers.AdminConfigResponse, error) {
	return nil, nil
}
