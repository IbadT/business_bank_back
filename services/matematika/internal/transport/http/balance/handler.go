package balance

import (
	"errors"
	"net/http"

	balanceservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/balance"
	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/dto"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/helpers"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/logger"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	s balanceservice.BalanceAdjustmentService
}

func NewHandler(s balanceservice.BalanceAdjustmentService) *Handler {
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
// @Failure      400      {object}  dto.ErrorResponse  "Некорректный запрос - ошибки валидации входных параметров"
// @Failure      401      {object}  dto.ErrorResponse     "Требуется авторизация"
// @Failure      404      {object}  dto.ErrorResponse  "Транзакции не найдены"
// @Failure      500      {object}  dto.ErrorResponse  "Внутренняя ошибка сервера"
// @Router       /api/balances/validate-balance [post]
func (h *Handler) ValidateBalance(c echo.Context) error {
	op := "http.handler.balance.validateBalance"
	log := logger.GetLogger().WithOperation(op)
	
	var req dto.ValidateBalanceRequest

	// Парсим входные данные
	if err := c.Bind(&req); err != nil {
		log.Error(err, "Invalid request body")
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   helpers.ErrMsgInvalidRequestBody,
			Details: err.Error(),
			Code:    http.StatusBadRequest,
		})
	}

	// Валидация request_id
	if req.RequestID == "" {
		log.Warn("requestId is required")
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: helpers.ErrMsgRequestIDRequiredAlt,
			Code:  http.StatusBadRequest,
		})
	}

	log = log.WithFields(logger.Fields{"request_id": req.RequestID})
	log.Info("Validating balance")

	// Вызываем сервис для валидации баланса
	result, err := h.s.ValidateBalance(req.RequestID)
	if err != nil {
		log.Error(err, "Failed to validate balance")
		statusCode := http.StatusInternalServerError
		if errors.Is(err, helpers.ErrInvalidRequestID) || errors.Is(err, helpers.ErrRequestIDEmpty) {
			statusCode = http.StatusBadRequest
		} else if errors.Is(err, helpers.ErrGenerationRequestNotFound) || errors.Is(err, helpers.ErrNoTransactionsFound) {
			statusCode = http.StatusNotFound
		}

		return c.JSON(statusCode, dto.ErrorResponse{
			Error:   helpers.ErrMsgFailedToValidateBalance,
			Details: err.Error(),
			Code:    statusCode,
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

	log.WithFields(logger.Fields{
		"is_valid":    result.IsValid,
		"issues_count": len(issues),
	}).Success("Balance validation completed")

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
// @Failure      400      {object}  dto.ErrorResponse  "Некорректный запрос - ошибки валидации входных параметров"
// @Failure      401      {object}  dto.ErrorResponse     "Требуется авторизация"
// @Failure      404      {object}  dto.ErrorResponse  "Корректировки не найдены"
// @Failure      500      {object}  dto.ErrorResponse  "Внутренняя ошибка сервера"
// @Router       /api/balances/{request_id}/balance-adjustment [get]
func (h *Handler) GetBalanceAdjustment(c echo.Context) error {
	op := "http.handler.balance.getBalanceAdjustment"
	log := logger.GetLogger().WithOperation(op)
	
	requestIDStr := c.Param("request_id")

	// Валидация request_id
	if requestIDStr == "" {
		log.Warn("request_id parameter is required")
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: helpers.ErrMsgRequestIDRequired,
			Code:  http.StatusBadRequest,
		})
	}

	log = log.WithFields(logger.Fields{"request_id": requestIDStr})
	log.Info("Getting balance adjustment")

	// Вызываем сервис для получения скорректированных транзакций
	transactions, err := h.s.GetAdjustedTransactions(requestIDStr)
	if err != nil {
		log.Error(err, "Failed to get balance adjustment")
		statusCode := http.StatusInternalServerError
		if errors.Is(err, helpers.ErrInvalidRequestID) || errors.Is(err, helpers.ErrRequestIDEmpty) {
			statusCode = http.StatusBadRequest
		} else if errors.Is(err, helpers.ErrGenerationRequestNotFound) || errors.Is(err, helpers.ErrNoTransactionsFound) {
			statusCode = http.StatusNotFound
		}

		return c.JSON(statusCode, dto.ErrorResponse{
			Error:   helpers.ErrMsgFailedToGetBalanceAdjustment,
			Details: err.Error(),
			Code:    statusCode,
		})
	}

	if len(transactions) == 0 {
		log.Warn("No balance adjustments found")
		return c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   helpers.ErrMsgNoBalanceAdjustmentsFound,
			Details: "requestId: " + requestIDStr,
			Code:    http.StatusNotFound,
		})
	}

	log.WithFields(logger.Fields{"transactions_count": len(transactions)}).Success("Balance adjustment retrieved")

	return c.JSON(http.StatusOK, dto.GetBalanceAdjustmentResponse{
		RequestID:    requestIDStr,
		Transactions: transactions,
		Code:         http.StatusOK,
	})
}
