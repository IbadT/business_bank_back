package balance

import (
	"net/http"
	"strings"

	"github.com/IbadT/business_bank_back/services/matematika/internal/service"
	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/dto"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	s service.BalanceAdjustmentService
}

func NewHandler(s service.BalanceAdjustmentService) *Handler {
	return &Handler{s}
}

// ========================= BALANCE =========================
// ValidateBalance - валидация баланса
// @Summary      Валидация баланса транзакций
// @Description  Проверяет баланс транзакций по request_id. Возвращает информацию о проблемах с балансом, если они есть. Требуется авторизация.
// @Tags         balance
// @Accept       json
// @Produce      json
// @security     BearerAuth
// @Param        request  body      dto.ValidateBalanceRequest  true  "Параметры валидации баланса"
// @Success      200      {object}  dto.ValidateBalanceResponse  "Успешная валидация баланса"
// @Failure      400      {object}  map[string]interface{}  "Некорректный запрос - ошибки валидации входных параметров"
// @Failure      401      {object}  map[string]string     "Требуется авторизация"
// @Failure      404      {object}  map[string]interface{}  "Транзакции не найдены"
// @Failure      500      {object}  map[string]interface{}  "Внутренняя ошибка сервера"
// @Router       /api/balances/validate-balance [post]
func (h *Handler) ValidateBalance(c echo.Context) error {
	var req dto.ValidateBalanceRequest

	// Парсим входные данные
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid request body",
			"details": err.Error(),
			"code":    http.StatusBadRequest,
		})
	}

	// Валидация request_id
	if req.RequestID == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "requestId is required",
			"code":  http.StatusBadRequest,
		})
	}

	// Вызываем сервис для валидации баланса
	result, err := h.s.ValidateBalance(req.RequestID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "invalid requestID") || strings.Contains(err.Error(), "empty") {
			statusCode = http.StatusBadRequest
		} else if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "no transactions found") {
			statusCode = http.StatusNotFound
		}

		return c.JSON(statusCode, map[string]interface{}{
			"error":   "Failed to validate balance",
			"details": err.Error(),
			"code":    statusCode,
		})
	}

	// Конвертируем результат сервиса в DTO
	issues := make([]dto.BalanceIssue, len(result.Issues))
	for i, issue := range result.Issues {
		issues[i] = dto.BalanceIssue{
			TransactionID:    issue.TransactionID,
			Date:             issue.Date,
			RequiredBalance:  issue.RequiredBalance,
			AvailableBalance: issue.AvailableBalance,
			Shortage:         issue.Shortage,
			ActionTaken:      issue.ActionTaken,
			NewDate:          issue.NewDate,
			OriginalAmount:   issue.OriginalAmount,
			AdjustedAmount:   issue.AdjustedAmount,
		}
	}

	response := dto.ValidateBalanceResponse{
		RequestID: result.RequestID,
		IsValid:   result.IsValid,
		Issues:    issues,
		Code:      http.StatusOK,
	}

	return c.JSON(http.StatusOK, response)
}

// GetBalanceAdjustment - получение корректировки баланса
// @Summary      Получение скорректированных транзакций
// @Description  Получает список транзакций, которые были скорректированы из-за недостатка баланса (перенесены или уменьшены). Требуется авторизация.
// @Tags         balance
// @Accept       json
// @Produce      json
// @security     BearerAuth
// @Param        request_id  path      string  true  "UUID запроса генерации" example:"550e8400-e29b-41d4-a716-446655440000"
// @Success      200      {object}  dto.GetBalanceAdjustmentResponse  "Успешное получение корректировок"
// @Failure      400      {object}  map[string]interface{}  "Некорректный запрос - ошибки валидации входных параметров"
// @Failure      401      {object}  map[string]string     "Требуется авторизация"
// @Failure      404      {object}  map[string]interface{}  "Корректировки не найдены"
// @Failure      500      {object}  map[string]interface{}  "Внутренняя ошибка сервера"
// @Router       /api/balances/{request_id}/balance-adjustment [get]
func (h *Handler) GetBalanceAdjustment(c echo.Context) error {
	requestIDStr := c.Param("request_id")

	// Валидация request_id
	if requestIDStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "request_id parameter is required",
			"code":  http.StatusBadRequest,
		})
	}

	// Вызываем сервис для получения скорректированных транзакций
	transactions, err := h.s.GetAdjustedTransactions(requestIDStr)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "invalid requestID") || strings.Contains(err.Error(), "empty") {
			statusCode = http.StatusBadRequest
		} else if strings.Contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		}

		return c.JSON(statusCode, map[string]interface{}{
			"error":   "Failed to get balance adjustment",
			"details": err.Error(),
			"code":    statusCode,
		})
	}

	if len(transactions) == 0 {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"error":     "No balance adjustments found for the given request_id",
			"requestId": requestIDStr,
			"code":      http.StatusNotFound,
		})
	}

	return c.JSON(http.StatusOK, dto.GetBalanceAdjustmentResponse{
		RequestID:    requestIDStr,
		Transactions: transactions,
		Code:         http.StatusOK,
	})
}
