package calculation

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/IbadT/business_bank_back/services/matematika/internal/helpers"
	"github.com/labstack/echo/v4"
)

// ============================================================================
// HANDLER LAYER - Обрабатывает HTTP запросы
// ============================================================================

// CalculationHandler - обрабатывает HTTP запросы для расчетов

type CalculationHandler struct {
	calcService CalculationService
}

// NewCalculationHandler создает новый handler
func NewCalculationHandler(calcService CalculationService) *CalculationHandler {
	return &CalculationHandler{
		calcService: calcService,
	}
}

// ============================================================================
// HTTP ENDPOINTS
// ============================================================================

// HealthCheck - проверка здоровья сервиса
// @Summary Health check
// @Description Проверяет состояние сервиса и его зависимостей
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} helpers.HealthCheckResponse
// @Failure 500 {object} helpers.ErrorResponse
// @Router /health [get]
func (h *CalculationHandler) HealthCheck(c echo.Context) error {
	ctx := c.Request().Context()
	response, err := h.calcService.HealthCheck(ctx)
	if err != nil {
		// Error - тип ошибки, Message - общее сообщение, Details - детали из err
		errorResponse := helpers.GenerateErrorResponse(helpers.ErrInternalServerError, helpers.ErrFailedToCheckHealth, err.Error())
		return c.JSON(http.StatusInternalServerError, errorResponse)
	}
	return c.JSON(http.StatusOK, response)
}

// ================================================
// POST /generate-statement-kafka
// ================================================

// GenerateStatement - генерация выписки
// POST /generate-statement
// Body: {"accountId": "123", "month": "2025-01", "businessType": "B2C", "initialBalance": 10000}
func (h *CalculationHandler) GenerateStatementToKafka(c echo.Context) error {
	// 1. Парсим входные данные
	var req helpers.GenerateStatementRequest
	if err := c.Bind(&req); err != nil {
		// Error - тип ошибки, Message - общее сообщение, Details - детали из err
		errorResponse := helpers.GenerateErrorResponse(helpers.ErrInvalidRequest, helpers.ErrInvalidRequestBody, err.Error())
		return c.JSON(http.StatusBadRequest, errorResponse)
	}

	// 2. Вызываем Service layer для обработки
	result, err := h.calcService.GenerateStatementToKafka(c.Request().Context(), &req)
	if err != nil {
		// Error - тип ошибки, Message - общее сообщение, Details - детали из err
		errorResponse := helpers.GenerateErrorResponse(helpers.ErrInternalServerError, helpers.ErrFailedToKafkaStatement, err.Error())
		return c.JSON(http.StatusInternalServerError, errorResponse)
	}

	fmt.Println("RESULT: ", result)

	// 3. Возвращаем результат
	return c.JSON(http.StatusCreated, result)
}

// ================================================
// GET /statement/:id/status
// ================================================

// GetStatementStatusByID - получение статуса выписки
// GET /statement/:id/status
func (h *CalculationHandler) GetStatementStatusByID(c echo.Context) error {
	id := c.Param("id")

	status, err := h.calcService.GetStatementStatusByID(c.Request().Context(), id)
	if err != nil {
		// Error - тип ошибки, Message - общее сообщение, Details - детали из err
		errorResponse := helpers.GenerateErrorResponse(helpers.ErrNotFound, helpers.ErrStatementNotFound, err.Error())
		return c.JSON(http.StatusNotFound, errorResponse)
	}

	return c.JSON(http.StatusOK, status)
}

// ================================================
// GET /statement/:id/result
// ================================================

// GetStatementResultByID - получение результатов расчетов
// GET /statement/:id/result
func (h *CalculationHandler) GetStatementResultByID(c echo.Context) error {
	id := c.Param("id")

	result, err := h.calcService.GetStatementResultByID(c.Request().Context(), id)
	if err != nil {
		// Error - тип ошибки, Message - общее сообщение, Details - детали из err
		errorResponse := helpers.GenerateErrorResponse(helpers.ErrNotFound, helpers.ErrStatementNotFound, err.Error())
		return c.JSON(http.StatusNotFound, errorResponse)
	}

	return c.JSON(http.StatusOK, result)
}

// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!

// ================================================
// POST /generate-statement
// ================================================

// GenerateStatement - генерация финансовой выписки
// @Summary Generate financial statement
// @Description Генерирует финансовую выписку на основе переданных данных
// @Tags statements
// @Accept json
// @Produce json
// @Param request body helpers.StatementStateRequest true "Statement generation request"
// @Success 202 {object} helpers.GenerateStatementResponse
// @Failure 400 {object} helpers.ErrorResponse "Invalid request"
// @Failure 422 {object} helpers.ErrorResponse "Validation error"
// @Failure 500 {object} helpers.ErrorResponse "Internal server error"
// @Router /generate-statement [post]
func (h *CalculationHandler) GenerateStatement(c echo.Context) error {
	var req helpers.StatementStateRequest

	// 1. Парсим входные данные
	if err := c.Bind(&req); err != nil {
		// Error - тип ошибки, Message - общее сообщение, Details - детали из err
		errorResponse := helpers.GenerateErrorResponse(helpers.ErrInvalidRequest, helpers.ErrInvalidRequestBody, err.Error())
		return c.JSON(http.StatusBadRequest, errorResponse)
	}

	// 2. Валидация с использованием validator
	if err := helpers.NewRequestValidator().ValidateRequest(&req); err != nil {
		if errors.Is(err, helpers.ErrFutureMonth) {
			// Error - тип ошибки, Message - общее сообщение, Details - детали из err
			errorResponse := helpers.GenerateErrorResponse(helpers.ErrFutureMonth, helpers.ErrFutureMonth, err.Error())
			return c.JSON(http.StatusUnprocessableEntity, errorResponse)
		} else {
			// Error - тип ошибки, Message - общее сообщение, Details - детали из err
			errorResponse := helpers.GenerateErrorResponse(helpers.ErrInvalidRequest, helpers.ErrInvalidRequestBody, err.Error())
			return c.JSON(http.StatusBadRequest, errorResponse)
		}
	}

	// 3. Call service
	ctx := c.Request().Context()
	result, err := h.calcService.GenerateStatement(ctx, &req)
	if err != nil {
		// Обработка специфичных ошибок валидации
		if errors.Is(err, helpers.ErrInsufficientBalance) {
			// Error - тип ошибки, Message - общее сообщение, Details - детали из err
			errorResponse := helpers.GenerateErrorResponse(helpers.ErrInsufficientBalance, helpers.ErrInsufficientBalance, err.Error())
			return c.JSON(http.StatusUnprocessableEntity, errorResponse)
		}
		if errors.Is(err, helpers.ErrFutureMonth) {
			// Error - тип ошибки, Message - общее сообщение, Details - детали из err
			errorResponse := helpers.GenerateErrorResponse(helpers.ErrFutureMonth, helpers.ErrFutureMonth, err.Error())
			return c.JSON(http.StatusUnprocessableEntity, errorResponse)
		}
		// Общая ошибка сервера
		// Error - тип ошибки, Message - общее сообщение, Details - детали из err
		errorResponse := helpers.GenerateErrorResponse(helpers.ErrInternalServerError, helpers.ErrFailedToGenerateStatement, err.Error())
		return c.JSON(http.StatusInternalServerError, errorResponse)
	}

	return c.JSON(202, result) // 202 Accepted для async операции
}

// ================================================
// GET /config
// ================================================

// AdminConfig - получение административной конфигурации
// @Summary Get admin configuration
// @Description Возвращает административную конфигурацию (категории расходов, расписания, шаблоны доходов)
// @Tags admin
// @Accept json
// @Produce json
// @Success 200 {object} helpers.AdminConfigResponse
// @Failure 500 {object} helpers.ErrorResponse
// @Router /admin/config [get]
func (h *CalculationHandler) AdminConfig(c echo.Context) error {
	ctx := c.Request().Context()
	response, err := h.calcService.GetAdminConfig(ctx)
	if err != nil {
		// Error - тип ошибки, Message - общее сообщение, Details - детали из err
		errorResponse := helpers.GenerateErrorResponse(helpers.ErrInternalServerError, helpers.ErrFailedToLoadConfiguration, err.Error())
		return c.JSON(http.StatusInternalServerError, errorResponse)
	}

	return c.JSON(http.StatusOK, response)
}

// ================================================
// GET /transactions
// ================================================
// GetTransactions - получение списка транзакций
// GET /transactions
func (h *CalculationHandler) GetTransactions(c echo.Context) error {
	ctx := c.Request().Context()
	transactions, err := h.calcService.GetTransactions(ctx)
	if err != nil {
		// Error - тип ошибки, Message - общее сообщение, Details - детали из err
		errorResponse := helpers.GenerateErrorResponse(helpers.ErrInternalServerError, helpers.ErrFailedToGetTransactions, err.Error())
		return c.JSON(http.StatusInternalServerError, errorResponse)
	}
	return c.JSON(http.StatusOK, transactions)
}

// ================================================
// GET /business-rules
// ================================================
// GetBusinessRules - получение список бизнес-правил
// GET /business-rules
func (h *CalculationHandler) GetBusinessRules(c echo.Context) error {
	ctx := c.Request().Context()
	businessRules, err := h.calcService.GetBusinessRules(ctx)
	if err != nil {
		// Error - тип ошибки, Message - общее сообщение, Details - детали из err
		errorResponse := helpers.GenerateErrorResponse(helpers.ErrInternalServerError, helpers.ErrFailedToGetBusinessRules, err.Error())
		return c.JSON(http.StatusInternalServerError, errorResponse)
	}
	return c.JSON(http.StatusOK, businessRules)
}

// ================================================
// GET /daily-balances
// ================================================
// GetDailyBalances - получение список дневных балансов
// GET /daily-balances
func (h *CalculationHandler) GetDailyBalances(c echo.Context) error {
	ctx := c.Request().Context()
	dailyBalances, err := h.calcService.GetDailyBalances(ctx)
	if err != nil {
		// Error - тип ошибки, Message - общее сообщение, Details - детали из err
		errorResponse := helpers.GenerateErrorResponse(helpers.ErrInternalServerError, helpers.ErrFailedToGetDailyBalances, err.Error())
		return c.JSON(http.StatusInternalServerError, errorResponse)
	}
	return c.JSON(http.StatusOK, dailyBalances)
}

// ================================================
// GET /statements
// ================================================
// GetStatements - получение список выписок
// GET /statements
func (h *CalculationHandler) GetStatements(c echo.Context) error {
	ctx := c.Request().Context()
	statements, err := h.calcService.GetStatements(ctx)
	if err != nil {
		// Error - тип ошибки, Message - общее сообщение, Details - детали из err
		errorResponse := helpers.GenerateErrorResponse(helpers.ErrInternalServerError, helpers.ErrFailedToGetStatements, err.Error())
		return c.JSON(http.StatusInternalServerError, errorResponse)
	}
	return c.JSON(http.StatusOK, statements)
}
