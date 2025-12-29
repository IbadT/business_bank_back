package baseamount

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	authMiddleware "github.com/IbadT/business_bank_back/services/matematika/internal/middleware"
	baseamountservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/base"
	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/dto"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/helpers"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/logger"
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
// @Failure      400      {object}  dto.ErrorResponse  "Некорректный запрос - неверный формат UUID"
// @Failure      401      {object}  dto.ErrorResponse     "Требуется авторизация"
// @Failure      500      {object}  dto.ErrorResponse  "Внутренняя ошибка сервера"
// @Router       /api/base-amounts [get]
func (h *Handler) GetBaseAmount(c echo.Context) error {
	op := "http.handler.baseAmount.getBaseAmount"
	log := logger.GetLogger().WithOperation(op)
	
	userIDStr := authMiddleware.GetUserID(c)
	if userIDStr == nil {
		log.Warn("User ID not found in context")
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: helpers.ErrMsgUnauthorized,
			Code:  http.StatusUnauthorized,
		})
	}
	
	log = log.WithFields(logger.Fields{"user_id": *userIDStr})
	log.Info("Getting base amounts")
	
	baseAmounts, err := h.s.GetBaseAmount(*userIDStr)
	if err != nil {
		log.Error(err, "Failed to get base amounts")
		statusCode := http.StatusInternalServerError
		if errors.Is(err, helpers.ErrInvalidUserID) || errors.Is(err, helpers.ErrUserIDRequired) {
			statusCode = http.StatusBadRequest
		}

		return c.JSON(statusCode, dto.ErrorResponse{
			Error:   helpers.ErrMsgFailedToGetBaseAmounts,
			Details: err.Error(),
			Code:    statusCode,
		})
	}
	
	log.WithFields(logger.Fields{
		"mobile":    baseAmounts.MobileBaseAmount.Amount,
		"utilities": baseAmounts.UtilitiesBaseAmount.Amount,
		"leasing":   baseAmounts.LeasingBaseAmount.Amount,
	}).Success("Base amounts retrieved")
	
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
// @Failure      400      {object}  dto.ErrorResponse  "Некорректный запрос"
// @Failure      401      {object}  dto.ErrorResponse     "Требуется авторизация"
// @Failure      404      {object}  dto.ErrorResponse  "Базовая сумма не найдена (для последующих месяцев)"
// @Failure      500      {object}  dto.ErrorResponse  "Внутренняя ошибка сервера"
// @Router       /api/base-amounts/mobile/calculate [get]
func (h *Handler) CalculateMobileAmount(c echo.Context) error {
	op := "http.handler.baseAmount.calculateMobileAmount"
	log := logger.GetLogger().WithOperation(op)
	
	userIDStr := authMiddleware.GetUserID(c)
	if userIDStr == nil {
		log.Warn("User ID not found in context")
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: helpers.ErrMsgUnauthorized,
			Code:  http.StatusUnauthorized,
		})
	}

	// Получаем параметр is_first_month из query (по умолчанию false)
	isFirstMonth := c.QueryParam("is_first_month") == "true"

	// Получаем месяц из query параметров или используем текущий месяц
	monthStr := c.QueryParam("month")
	if monthStr == "" {
		monthStr = time.Now().Format("2006-01")
	}

	log = log.WithFields(logger.Fields{
		"user_id":      *userIDStr,
		"is_first_month": isFirstMonth,
		"month":        monthStr,
	})
	log.Info("Calculating mobile amount")

	amount, err := h.s.CalculateMobileAmount(*userIDStr, isFirstMonth, monthStr)
	if err != nil {
		log.Error(err, "Failed to calculate mobile amount")
		statusCode := http.StatusInternalServerError
		if errors.Is(err, helpers.ErrMobileBaseAmountNotFound) {
			statusCode = http.StatusNotFound
		}

		return c.JSON(statusCode, dto.ErrorResponse{
			Error:   helpers.ErrMsgFailedToCalculateMobileAmount,
			Details: err.Error(),
			Code:    statusCode,
		})
	}

	log.WithFields(logger.Fields{"amount": amount}).Success("Mobile amount calculated")

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
// @Failure      400      {object}  dto.ErrorResponse  "Некорректный запрос"
// @Failure      401      {object}  dto.ErrorResponse     "Требуется авторизация"
// @Failure      404      {object}  dto.ErrorResponse  "Базовая сумма не найдена (для последующих месяцев)"
// @Failure      500      {object}  dto.ErrorResponse  "Внутренняя ошибка сервера"
// @Router       /api/base-amounts/utilities/calculate [get]
func (h *Handler) CalculateUtilitiesAmount(c echo.Context) error {
	op := "http.handler.baseAmount.calculateUtilitiesAmount"
	log := logger.GetLogger().WithOperation(op)
	
	userIDStr := authMiddleware.GetUserID(c)
	if userIDStr == nil {
		log.Warn("User ID not found in context")
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: helpers.ErrMsgUnauthorized,
			Code:  http.StatusUnauthorized,
		})
	}

	// Получаем параметр is_first_month из query (по умолчанию false)
	isFirstMonth := c.QueryParam("is_first_month") == "true"

	// Получаем месяц из query параметров или используем текущий месяц
	monthStr := c.QueryParam("month")
	if monthStr == "" {
		monthStr = time.Now().Format("2006-01")
	}

	log = log.WithFields(logger.Fields{
		"user_id":      *userIDStr,
		"is_first_month": isFirstMonth,
		"month":        monthStr,
	})
	log.Info("Calculating utilities amount")

	amount, err := h.s.CalculateUtilitiesAmount(*userIDStr, isFirstMonth, monthStr)
	if err != nil {
		log.Error(err, "Failed to calculate utilities amount")
		statusCode := http.StatusInternalServerError
		if errors.Is(err, helpers.ErrUtilitiesBaseAmountNotFound) {
			statusCode = http.StatusNotFound
		}

		return c.JSON(statusCode, dto.ErrorResponse{
			Error:   helpers.ErrMsgFailedToCalculateUtilitiesAmount,
			Details: err.Error(),
			Code:    statusCode,
		})
	}

	log.WithFields(logger.Fields{"amount": amount}).Success("Utilities amount calculated")

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
// @Failure      400      {object}  dto.ErrorResponse  "Некорректный запрос - turnover обязателен для первого месяца или должен быть положительным числом"
// @Failure      401      {object}  dto.ErrorResponse     "Требуется авторизация"
// @Failure      404      {object}  dto.ErrorResponse  "Базовая сумма не найдена (для последующих месяцев)"
// @Failure      500      {object}  dto.ErrorResponse  "Внутренняя ошибка сервера"
// @Router       /api/base-amounts/leasing/calculate [get]
func (h *Handler) CalculateLeasingAmount(c echo.Context) error {
	op := "http.handler.baseAmount.calculateLeasingAmount"
	log := logger.GetLogger().WithOperation(op)
	
	userIDStr := authMiddleware.GetUserID(c)
	if userIDStr == nil {
		log.Warn("User ID not found in context")
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: helpers.ErrMsgUnauthorized,
			Code:  http.StatusUnauthorized,
		})
	}

	// Получаем параметры из query
	isFirstMonth := c.QueryParam("is_first_month") == "true"
	turnoverStr := c.QueryParam("turnover")

	var turnover float64
	if isFirstMonth {
		// Для первого месяца turnover обязателен
		if turnoverStr == "" {
			log.Warn("turnover parameter is required for first month")
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: helpers.ErrMsgTurnoverParameterRequiredForFirstMonth,
			Code:  http.StatusBadRequest,
		})
		}

		var err error
		turnover, err = strconv.ParseFloat(turnoverStr, 64)
		if err != nil || turnover <= 0 {
			log.Warn("Invalid turnover: %s", turnoverStr)
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: helpers.ErrMsgTurnoverMustBePositiveNumber,
			Code:  http.StatusBadRequest,
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

	log = log.WithFields(logger.Fields{
		"user_id":      *userIDStr,
		"is_first_month": isFirstMonth,
		"month":        monthStr,
		"turnover":     turnover,
	})
	log.Info("Calculating leasing amount")

	amount, err := h.s.CalculateLeasingAmount(*userIDStr, turnover, isFirstMonth, monthStr)
	if err != nil {
		log.Error(err, "Failed to calculate leasing amount")
		statusCode := http.StatusInternalServerError
		if errors.Is(err, helpers.ErrLeasingBaseAmountNotFound) {
			statusCode = http.StatusNotFound
		} else if errors.Is(err, helpers.ErrTurnoverMustBeGreaterThanZeroForLeasing) {
			statusCode = http.StatusBadRequest
		}

		return c.JSON(statusCode, dto.ErrorResponse{
			Error:   helpers.ErrMsgFailedToCalculateLeasingAmount,
			Details: err.Error(),
			Code:    statusCode,
		})
	}

	log.WithFields(logger.Fields{"amount": amount}).Success("Leasing amount calculated")

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
// @Failure      401      {object}  dto.ErrorResponse     "Требуется авторизация"
// @Failure      500      {object}  dto.ErrorResponse  "Внутренняя ошибка сервера"
// @Router       /api/base-amounts/mobile [delete]
func (h *Handler) ResetMobileBaseAmount(c echo.Context) error {
	op := "http.handler.baseAmount.resetMobileBaseAmount"
	log := logger.GetLogger().WithOperation(op)
	
	userIDStr := authMiddleware.GetUserID(c)
	if userIDStr == nil {
		log.Warn("User ID not found in context")
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: helpers.ErrMsgUnauthorized,
			Code:  http.StatusUnauthorized,
		})
	}

	log = log.WithFields(logger.Fields{"user_id": *userIDStr})
	log.Info("Resetting mobile base amount")

	if err := h.s.DeleteMobileBaseAmount(*userIDStr); err != nil {
		log.Error(err, "Failed to reset mobile base amount")
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   helpers.ErrMsgFailedToResetMobileBaseAmount,
			Details: err.Error(),
			Code:    http.StatusInternalServerError,
		})
	}

	log.Success("Mobile base amount reset successfully")

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
// @Failure      401      {object}  dto.ErrorResponse     "Требуется авторизация"
// @Failure      500      {object}  dto.ErrorResponse  "Внутренняя ошибка сервера"
// @Router       /api/base-amounts/utilities [delete]
func (h *Handler) ResetUtilitiesBaseAmount(c echo.Context) error {
	op := "http.handler.baseAmount.resetUtilitiesBaseAmount"
	log := logger.GetLogger().WithOperation(op)
	
	userIDStr := authMiddleware.GetUserID(c)
	if userIDStr == nil {
		log.Warn("User ID not found in context")
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: helpers.ErrMsgUnauthorized,
			Code:  http.StatusUnauthorized,
		})
	}

	log = log.WithFields(logger.Fields{"user_id": *userIDStr})
	log.Info("Resetting utilities base amount")

	if err := h.s.DeleteUtilitiesBaseAmount(*userIDStr); err != nil {
		log.Error(err, "Failed to reset utilities base amount")
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   helpers.ErrMsgFailedToResetUtilitiesBaseAmount,
			Details: err.Error(),
			Code:    http.StatusInternalServerError,
		})
	}

	log.Success("Utilities base amount reset successfully")

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
// @Failure      401      {object}  dto.ErrorResponse     "Требуется авторизация"
// @Failure      500      {object}  dto.ErrorResponse  "Внутренняя ошибка сервера"
// @Router       /api/base-amounts/leasing [delete]
func (h *Handler) ResetLeasingBaseAmount(c echo.Context) error {
	op := "http.handler.baseAmount.resetLeasingBaseAmount"
	log := logger.GetLogger().WithOperation(op)
	
	userIDStr := authMiddleware.GetUserID(c)
	if userIDStr == nil {
		log.Warn("User ID not found in context")
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: helpers.ErrMsgUnauthorized,
			Code:  http.StatusUnauthorized,
		})
	}

	log = log.WithFields(logger.Fields{"user_id": *userIDStr})
	log.Info("Resetting leasing base amount")

	if err := h.s.DeleteLeasingBaseAmount(*userIDStr); err != nil {
		log.Error(err, "Failed to reset leasing base amount")
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   helpers.ErrMsgFailedToResetLeasingBaseAmount,
			Details: err.Error(),
			Code:    http.StatusInternalServerError,
		})
	}

	log.Success("Leasing base amount reset successfully")

	return c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Leasing base amount reset successfully",
		Code:    http.StatusOK,
	})
}
