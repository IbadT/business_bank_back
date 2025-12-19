package baseamount

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	authMiddleware "github.com/IbadT/business_bank_back/services/matematika/internal/middleware"
	baseamountservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/base"
	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/dto"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	s baseamountservice.BaseAmountService
}

func NewHandler(s baseamountservice.BaseAmountService) *Handler {
	return &Handler{s}
}

// ========================= BASE AMOUNTS =========================
// TODO: Добавить админку

// GetBaseAmount - получение базовых сумм
// @Summary      Получение базовых сумм
// @Description  Получает базовые суммы для мобильной связи, коммунальных и лизинга. Требуется авторизация.
// @Tags         base-amounts
// @Accept       json
// @Produce      json
// @security     BearerAuth
// @Success      200      {object}  dto.BaseAmountsResponse  "Успешное получение базовых сумм"
// @Failure      400      {object}  map[string]interface{}  "Некорректный запрос - неверный формат UUID"
// @Failure      401      {object}  map[string]string     "Требуется авторизация"
// @Failure      500      {object}  map[string]interface{}  "Внутренняя ошибка сервера"
// @Router       /api/base-amounts [get]
func (h *Handler) GetBaseAmount(c echo.Context) error {
	userIDStr := authMiddleware.GetUserID(c)
	if userIDStr == nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": "Unauthorized",
			"code":  http.StatusUnauthorized,
		})
	}
	baseAmounts, err := h.s.GetBaseAmount(*userIDStr)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "invalid userID") {
			statusCode = http.StatusBadRequest
		}

		return c.JSON(statusCode, map[string]interface{}{
			"error":   "Failed to get base amounts",
			"details": err.Error(),
			"code":    statusCode,
		})
	}
	return c.JSON(http.StatusOK, dto.BaseAmountsResponse{
		UserID:              *userIDStr,
		MobileBaseAmount:    baseAmounts.MobileBaseAmount,
		UtilitiesBaseAmount: baseAmounts.UtilitiesBaseAmount,
		LeasingBaseAmount:   baseAmounts.LeasingBaseAmount,
		Code:                http.StatusOK,
	})
}

// CalculateMobileAmount - расчет суммы мобильной связи
// @Summary      Расчет суммы мобильной связи
// @Description  Рассчитывает сумму мобильной связи. Первый месяц: $200-500 (фиксируется). Последующие месяцы: ±15% от базовой суммы. Требуется авторизация.
// @Tags         base-amounts
// @Accept       json
// @Produce      json
// @security     BearerAuth
// @Param        is_first_month  query      bool  false  "Является ли это первым месяцем (по умолчанию false)" example:"true"
// @Success      200      {object}  dto.CalculateMobileAmountResponse  "Успешное получение рассчитанной суммы мобильной связи"
// @Failure      400      {object}  map[string]interface{}  "Некорректный запрос"
// @Failure      401      {object}  map[string]string     "Требуется авторизация"
// @Failure      404      {object}  map[string]interface{}  "Базовая сумма не найдена (для последующих месяцев)"
// @Failure      500      {object}  map[string]interface{}  "Внутренняя ошибка сервера"
// @Router       /api/base-amounts/mobile/calculate [get]
func (h *Handler) CalculateMobileAmount(c echo.Context) error {
	userIDStr := authMiddleware.GetUserID(c)
	if userIDStr == nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": "Unauthorized",
			"code":  http.StatusUnauthorized,
		})
	}

	// Получаем параметр is_first_month из query (по умолчанию false)
	isFirstMonth := c.QueryParam("is_first_month") == "true"

	// Получаем месяц из query параметров или используем текущий месяц
	monthStr := c.QueryParam("month")
	if monthStr == "" {
		monthStr = time.Now().Format("2006-01")
	}

	amount, err := h.s.CalculateMobileAmount(*userIDStr, isFirstMonth, monthStr)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		}

		return c.JSON(statusCode, map[string]interface{}{
			"error":   "Failed to calculate mobile amount",
			"details": err.Error(),
			"code":    statusCode,
		})
	}

	return c.JSON(http.StatusOK, dto.CalculateMobileAmountResponse{
		UserID:       *userIDStr,
		Amount:       amount,
		IsFirstMonth: isFirstMonth,
		Code:         http.StatusOK,
	})
}

// CalculateUtilitiesAmount - расчет суммы коммунальных
// @Summary      Расчет суммы коммунальных
// @Description  Рассчитывает сумму коммунальных. Первый месяц: $200-500 (фиксируется). Последующие месяцы: ±15% от базовой суммы. Требуется авторизация.
// @Tags         base-amounts
// @Accept       json
// @Produce      json
// @security     BearerAuth
// @Param        is_first_month  query      bool  false  "Является ли это первым месяцем (по умолчанию false)" example:"true"
// @Success      200      {object}  dto.CalculateUtilitiesAmountResponse  "Успешное получение рассчитанной суммы коммунальных"
// @Failure      400      {object}  map[string]interface{}  "Некорректный запрос"
// @Failure      401      {object}  map[string]string     "Требуется авторизация"
// @Failure      404      {object}  map[string]interface{}  "Базовая сумма не найдена (для последующих месяцев)"
// @Failure      500      {object}  map[string]interface{}  "Внутренняя ошибка сервера"
// @Router       /api/base-amounts/utilities/calculate [get]
func (h *Handler) CalculateUtilitiesAmount(c echo.Context) error {
	userIDStr := authMiddleware.GetUserID(c)
	if userIDStr == nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": "Unauthorized",
			"code":  http.StatusUnauthorized,
		})
	}

	// Получаем параметр is_first_month из query (по умолчанию false)
	isFirstMonth := c.QueryParam("is_first_month") == "true"

	// Получаем месяц из query параметров или используем текущий месяц
	monthStr := c.QueryParam("month")
	if monthStr == "" {
		monthStr = time.Now().Format("2006-01")
	}

	amount, err := h.s.CalculateUtilitiesAmount(*userIDStr, isFirstMonth, monthStr)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		}

		return c.JSON(statusCode, map[string]interface{}{
			"error":   "Failed to calculate utilities amount",
			"details": err.Error(),
			"code":    statusCode,
		})
	}

	return c.JSON(http.StatusOK, dto.CalculateUtilitiesAmountResponse{
		UserID:       *userIDStr,
		Amount:       amount,
		IsFirstMonth: isFirstMonth,
		Code:         http.StatusOK,
	})
}

// CalculateLeasingAmount - расчет суммы лизинга
// @Summary      Расчет суммы лизинга
// @Description  Рассчитывает сумму лизинга. Первый месяц: 11.5-12% оборота (фиксируется). Последующие месяцы: повторяется 1:1. Требуется авторизация. Для первого месяца параметр turnover обязателен.
// @Tags         base-amounts
// @Accept       json
// @Produce      json
// @security     BearerAuth
// @Param        turnover  query      float64  false  "Оборот для расчета (обязателен только для первого месяца)" example:"100000.00"
// @Param        is_first_month  query      bool  false  "Является ли это первым месяцем (по умолчанию false)" example:"true"
// @Success      200      {object}  dto.CalculateLeasingAmountResponse  "Успешное получение рассчитанной суммы лизинга"
// @Failure      400      {object}  map[string]interface{}  "Некорректный запрос - turnover обязателен для первого месяца или должен быть положительным числом"
// @Failure      401      {object}  map[string]string     "Требуется авторизация"
// @Failure      404      {object}  map[string]interface{}  "Базовая сумма не найдена (для последующих месяцев)"
// @Failure      500      {object}  map[string]interface{}  "Внутренняя ошибка сервера"
// @Router       /api/base-amounts/leasing/calculate [get]
func (h *Handler) CalculateLeasingAmount(c echo.Context) error {
	userIDStr := authMiddleware.GetUserID(c)
	if userIDStr == nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": "Unauthorized",
			"code":  http.StatusUnauthorized,
		})
	}

	// Получаем параметры из query
	isFirstMonth := c.QueryParam("is_first_month") == "true"
	turnoverStr := c.QueryParam("turnover")

	var turnover float64
	if isFirstMonth {
		// Для первого месяца turnover обязателен
		if turnoverStr == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"error": "turnover parameter is required for first month",
				"code":  http.StatusBadRequest,
			})
		}

		var err error
		turnover, err = strconv.ParseFloat(turnoverStr, 64)
		if err != nil || turnover <= 0 {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"error": "turnover must be a positive number",
				"code":  http.StatusBadRequest,
			})
		}
	} else {
		// Для последующих месяцев turnover не нужен, но можно передать для информации
		if turnoverStr != "" {
			turnover, _ = strconv.ParseFloat(turnoverStr, 64)
		}
	}

	// Получаем месяц из query параметров или используем текущий месяц
	monthStr := c.QueryParam("month")
	if monthStr == "" {
		monthStr = time.Now().Format("2006-01")
	}

	amount, err := h.s.CalculateLeasingAmount(*userIDStr, turnover, isFirstMonth, monthStr)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		} else if strings.Contains(err.Error(), "turnover must be greater than 0") {
			statusCode = http.StatusBadRequest
		}

		return c.JSON(statusCode, map[string]interface{}{
			"error":   "Failed to calculate leasing amount",
			"details": err.Error(),
			"code":    statusCode,
		})
	}

	return c.JSON(http.StatusOK, dto.CalculateLeasingAmountResponse{
		UserID:       *userIDStr,
		Amount:       amount,
		Turnover:     turnover,
		IsFirstMonth: isFirstMonth,
		Code:         http.StatusOK,
	})
}

// ResetMobileBaseAmount - сброс суммы мобильной связи
// @Summary      Сброс суммы мобильной связи
// @Description  Удаляет сохраненную базовую сумму мобильной связи. Требуется авторизация.
// @Tags         base-amounts
// @Accept       json
// @Produce      json
// @security     BearerAuth
// @Success      200      {object}  dto.MessageResponse  "Успешный сброс суммы мобильной связи"
// @Failure      401      {object}  map[string]string     "Требуется авторизация"
// @Failure      500      {object}  map[string]interface{}  "Внутренняя ошибка сервера"
// @Router       /api/base-amounts/mobile [delete]
func (h *Handler) ResetMobileBaseAmount(c echo.Context) error {
	userIDStr := authMiddleware.GetUserID(c)
	if userIDStr == nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": "Unauthorized",
			"code":  http.StatusUnauthorized,
		})
	}

	if err := h.s.DeleteMobileBaseAmount(*userIDStr); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to reset mobile base amount",
			"details": err.Error(),
			"code":    http.StatusInternalServerError,
		})
	}

	return c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Mobile base amount reset successfully",
		Code:    http.StatusOK,
	})
}

// ResetUtilitiesBaseAmount - сброс суммы коммунальных
// @Summary      Сброс суммы коммунальных
// @Description  Удаляет сохраненную базовую сумму коммунальных. Требуется авторизация.
// @Tags         base-amounts
// @Accept       json
// @Produce      json
// @security     BearerAuth
// @Success      200      {object}  dto.MessageResponse  "Успешный сброс суммы коммунальных"
// @Failure      401      {object}  map[string]string     "Требуется авторизация"
// @Failure      500      {object}  map[string]interface{}  "Внутренняя ошибка сервера"
// @Router       /api/base-amounts/utilities [delete]
func (h *Handler) ResetUtilitiesBaseAmount(c echo.Context) error {
	userIDStr := authMiddleware.GetUserID(c)
	if userIDStr == nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": "Unauthorized",
			"code":  http.StatusUnauthorized,
		})
	}

	if err := h.s.DeleteUtilitiesBaseAmount(*userIDStr); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to reset utilities base amount",
			"details": err.Error(),
			"code":    http.StatusInternalServerError,
		})
	}

	return c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Utilities base amount reset successfully",
		Code:    http.StatusOK,
	})
}

// ResetLeasingBaseAmount - сброс суммы лизинга
// @Summary      Сброс суммы лизинга
// @Description  Удаляет сохраненную базовую сумму лизинга. Требуется авторизация.
// @Tags         base-amounts
// @Accept       json
// @Produce      json
// @security     BearerAuth
// @Success      200      {object}  dto.MessageResponse  "Успешный сброс суммы лизинга"
// @Failure      401      {object}  map[string]string     "Требуется авторизация"
// @Failure      500      {object}  map[string]interface{}  "Внутренняя ошибка сервера"
// @Router       /api/base-amounts/leasing [delete]
func (h *Handler) ResetLeasingBaseAmount(c echo.Context) error {
	userIDStr := authMiddleware.GetUserID(c)
	if userIDStr == nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": "Unauthorized",
			"code":  http.StatusUnauthorized,
		})
	}

	if err := h.s.DeleteLeasingBaseAmount(*userIDStr); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to reset leasing base amount",
			"details": err.Error(),
			"code":    http.StatusInternalServerError,
		})
	}

	return c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Leasing base amount reset successfully",
		Code:    http.StatusOK,
	})
}
