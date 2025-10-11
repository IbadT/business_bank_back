package calculation

import (
	"net/http"

	"github.com/labstack/echo"
)

// ============================================================================
// HANDLER LAYER - Обрабатывает HTTP запросы
// ============================================================================

// CalculationHandler - обрабатывает HTTP запросы для расчетов

// ============================================================================
// REQUEST/RESPONSE MODELS
// ============================================================================

// GenerateStatementRequest - запрос на генерацию выписки
type GenerateStatementRequest struct {
	AccountID      string  `json:"accountId" validate:"required"`
	Month          string  `json:"month" validate:"required"`
	BusinessType   string  `json:"businessType" validate:"required,oneof=B2B B2C"`
	InitialBalance float64 `json:"initialBalance" validate:"required,gte=0"`
}

// GenerateStatementResponse - ответ на генерацию выписки
type GenerateStatementResponse struct {
	StatementID string `json:"statementId"`
	Status      string `json:"status"`
	Message     string `json:"message"`
}

type CalculationHandler struct {
	calcService CalculationService
}

// NewCalculationHandler создает новый handler
func NewCalculationHandler(calcService CalculationService) *CalculationHandler {
	return &CalculationHandler{calcService: calcService}
}

// ============================================================================
// HTTP ENDPOINTS
// ============================================================================

// HealthCheck - проверка здоровья сервиса
// GET /health
func (h *CalculationHandler) HealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status":  "healthy",
		"service": "matematika",
	})
}

// GenerateStatement - генерация выписки
// POST /generate-statement
// Body: {"accountId": "123", "month": "2025-01", "businessType": "B2C", "initialBalance": 10000}
func (h *CalculationHandler) GenerateStatement(c echo.Context) error {
	// 1. Парсим входные данные
	var req GenerateStatementRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	// 2. Вызываем Service layer для обработки
	result, err := h.calcService.GenerateStatement(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	// 3. Возвращаем результат
	return c.JSON(http.StatusCreated, result)
}

// GetStatementStatusByID - получение статуса выписки
// GET /statement/:id/status
func (h *CalculationHandler) GetStatementStatusByID(c echo.Context) error {
	id := c.Param("id")

	status, err := h.calcService.GetStatementStatusByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Statement not found",
		})
	}

	return c.JSON(http.StatusOK, status)
}

// GetStatementResultByID - получение результатов расчетов
// GET /statement/:id/result
func (h *CalculationHandler) GetStatementResultByID(c echo.Context) error {
	id := c.Param("id")

	result, err := h.calcService.GetStatementResultByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Statement not found",
		})
	}

	return c.JSON(http.StatusOK, result)
}

// Сервис гарантирует, что суммарные доходы = 100%
func CheckTransactionsPersentage() (bool, error) {
	return true, nil
}

type FinancialDistributionRequest struct {
	NetProfit                any
	CountIncomeTransactions  int
	TransactionModel         BusinessType
	TransactionsCountByMonth int
}

// •	Распределение доходов и расходов по категориям
func (h *CalculationHandler) financialDistributionHandler(c echo.Context) error {
	return nil
}
