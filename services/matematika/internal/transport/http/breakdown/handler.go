package breakdown

import (
	"errors"
	"net/http"

	breakdownservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/breakdown"
	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/dto"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/helpers"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/logger"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	s breakdownservice.BreakdownService
}

func NewHandler(s breakdownservice.BreakdownService) *Handler {
	return &Handler{s}
}

// ========================= BREAKDOWN =========================
// CalculateRevenueBreakdown - расчет разбивки доходов
// @Summary      Расчет разбивки доходов
// @Description  Рассчитывает разбивку доходов по методам платежа (ACH, Wire, Zelle, Gateway, Other) для указанного request_id. Требуется авторизация.
// @Tags         breakdown
// @Accept       json
// @Produce      json
// @security     BearerAuth
// @Param        request_id  path      string  true  "UUID запроса генерации" example:"550e8400-e29b-41d4-a716-446655440000"
// @Success      200      {object}  dto.CalculateRevenueBreakdownResponse  "Успешное получение разбивки доходов"
// @Failure      400      {object}  map[string]interface{}  "Некорректный запрос - неверный формат UUID"
// @Failure      401      {object}  map[string]string     "Требуется авторизация"
// @Failure      404      {object}  map[string]interface{}  "Транзакции не найдены"
// @Failure      500      {object}  map[string]interface{}  "Внутренняя ошибка сервера"
// @Router       /api/breakdowns/revenue/{request_id} [get]
func (h *Handler) CalculateRevenueBreakdown(c echo.Context) error {
	op := "http.handler.breakdown.calculateRevenueBreakdown"
	log := logger.GetLogger().WithOperation(op)
	
	requestIDStr := c.Param("request_id")
	if requestIDStr == "" {
		log.Warn("request_id parameter is required")
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "request_id parameter is required",
			"code":  http.StatusBadRequest,
		})
	}

	log = log.WithFields(logger.Fields{"request_id": requestIDStr})
	log.Info("Calculating revenue breakdown")

	result, err := h.s.GetRevenueBreakdown(requestIDStr)
	if err != nil {
		log.Error(err, "Failed to calculate revenue breakdown")
		statusCode := http.StatusInternalServerError
		errorMessage := "Failed to get revenue breakdown"

		if errors.Is(err, helpers.ErrInvalidRequestID) {
			statusCode = http.StatusBadRequest
			errorMessage = "Invalid request_id format. Expected UUID format (e.g., 550e8400-e29b-41d4-a716-446655440000)"
		}

		return c.JSON(statusCode, map[string]interface{}{
			"error":   errorMessage,
			"details": err.Error(),
			"code":    statusCode,
		})
	}

	log.WithFields(logger.Fields{
		"total_ach":     result.TotalAch,
		"total_wire":    result.TotalWire,
		"total_zelle":   result.TotalZelle,
		"total_gateway": result.TotalGateway,
	}).Success("Revenue breakdown calculated")

	return c.JSON(http.StatusOK, dto.CalculateRevenueBreakdownResponse{
		RequestID: requestIDStr,
		RevenueBreakdown: dto.RevenueBreakdown{
			TotalAch:     result.TotalAch,
			TotalWire:    result.TotalWire,
			TotalZelle:   result.TotalZelle,
			TotalGateway: result.TotalGateway,
			TotalOther:   result.TotalOther,
		},
		Code: http.StatusOK,
	})
}

// CalculateExpensesBreakdown - расчет разбивки расходов
// @Summary      Расчет разбивки расходов
// @Description  Рассчитывает разбивку расходов по методам платежа (card vs account) для указанного request_id. Требуется авторизация.
// @Tags         breakdown
// @Accept       json
// @Produce      json
// @security     BearerAuth
// @Param        request_id  path      string  true  "UUID запроса генерации" example:"550e8400-e29b-41d4-a716-446655440000"
// @Success      200      {object}  dto.CalculateExpensesBreakdownResponse  "Успешное получение разбивки расходов"
// @Failure      400      {object}  map[string]interface{}  "Некорректный запрос - неверный формат UUID"
// @Failure      401      {object}  map[string]string     "Требуется авторизация"
// @Failure      404      {object}  map[string]interface{}  "Транзакции не найдены"
// @Failure      500      {object}  map[string]interface{}  "Внутренняя ошибка сервера"
// @Router       /api/breakdowns/expenses/{request_id} [get]
func (h *Handler) CalculateExpensesBreakdown(c echo.Context) error {
	op := "http.handler.breakdown.calculateExpensesBreakdown"
	log := logger.GetLogger().WithOperation(op)
	
	requestIDStr := c.Param("request_id")
	if requestIDStr == "" {
		log.Warn("request_id parameter is required")
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "request_id parameter is required",
			"code":  http.StatusBadRequest,
		})
	}

	log = log.WithFields(logger.Fields{"request_id": requestIDStr})
	log.Info("Calculating expenses breakdown")

	result, err := h.s.GetExpensesBreakdown(requestIDStr)
	if err != nil {
		log.Error(err, "Failed to calculate expenses breakdown")
		statusCode := http.StatusInternalServerError
		errorMessage := "Failed to get expenses breakdown"

		if errors.Is(err, helpers.ErrInvalidRequestID) {
			statusCode = http.StatusBadRequest
			errorMessage = "Invalid request_id format. Expected UUID format (e.g., 550e8400-e29b-41d4-a716-446655440000)"
		}

		return c.JSON(statusCode, map[string]interface{}{
			"error":   errorMessage,
			"details": err.Error(),
			"code":    statusCode,
		})
	}

	log.WithFields(logger.Fields{
		"by_card":    result.ByCard,
		"by_account": result.ByAccount,
	}).Success("Expenses breakdown calculated")

	return c.JSON(http.StatusOK, dto.CalculateExpensesBreakdownResponse{
		RequestID: requestIDStr,
		ExpensesBreakdown: dto.ExpensesBreakdown{
			ByCard:    result.ByCard,
			ByAccount: result.ByAccount,
		},
		Code: http.StatusOK,
	})
}
