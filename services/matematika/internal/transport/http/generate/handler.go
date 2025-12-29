package generate

import (
	"errors"
	"fmt"
	"net/http"

	authMiddleware "github.com/IbadT/business_bank_back/services/matematika/internal/middleware"
	generatorservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/generator"
	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/dto"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/helpers"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/logger"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	s generatorservice.GeneratorService
}

func NewHandler(s generatorservice.GeneratorService) *Handler {
	return &Handler{s}
}

// ========================= GENERATE =========================

// Generate - генерация финансовой выписки
// @Summary      Генерация финансовой выписки
// @Description  Генерирует финансовую выписку с транзакциями на основе переданных параметров. Поддерживает модели B2C и B2B, позволяет задавать желаемый процент прибыли, начальный баланс и дополнительные кастомные данные.
// @Tags         generator
// @Accept       json
// @Produce      json
// @security     BearerAuth
// @Param        request  body      dto.GenerateRequest  true  "Параметры генерации выписки"
// @Success      200      {object}  dto.GenerateResponse  "Успешная генерация выписки"
// @Failure      400      {object}  dto.ErrorResponse     "Некорректный запрос - ошибки валидации входных параметров"
// @Failure      401      {object}  dto.ErrorResponse     "Требуется авторизация"
// @Failure      422      {object}  dto.ErrorResponse     "Ошибка валидации - транзакция приведет к отрицательному балансу"
// @Failure      500      {object}  dto.ErrorResponse     "Внутренняя ошибка сервера"
// @Router       /api/generate [post]
func (h *Handler) Generate(c echo.Context) error {
	op := "http.handler.generate.generate"
	log := logger.GetLogger().WithOperation(op)
	
	var req dto.GenerateRequest

	// 1. Парсим входные данные
	if err := c.Bind(&req); err != nil {
		log.Error(err, "Invalid request body")
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   helpers.ErrMsgInvalidRequestBody,
			Details: err.Error(),
			Code:    http.StatusBadRequest,
		})
	}

	log = log.WithFields(logger.Fields{
		"month":                req.Month,
		"year":                 req.Year,
		"model":                req.Model,
		"turnover":             req.Turnover,
		"desired_profit_percent": req.DesiredProfitPercent,
	})
	log.Info("Generating transactions")

	// 2. Валидация базовых полей
	if req.Turnover <= 0 {
		log.Warn("Invalid turnover: %f", req.Turnover)
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   helpers.ErrMsgTurnoverMustBeGreaterThanZero,
			Details: fmt.Sprintf("Received value: %f", req.Turnover),
			Code:    http.StatusBadRequest,
		})
	}
	if req.DesiredProfitPercent < 0 || req.DesiredProfitPercent > 100 {
		log.Warn("Invalid desired profit percent: %f", req.DesiredProfitPercent)
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   helpers.ErrMsgDesiredProfitPercentInvalid,
			Details: fmt.Sprintf("Received value: %f (must be between 0 and 100)", req.DesiredProfitPercent),
			Code:    http.StatusBadRequest,
		})
	}
	if req.Model != "B2C" && req.Model != "B2B" {
		log.Warn("Invalid model: %s", req.Model)
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   helpers.ErrMsgModelMustBeB2COrB2B,
			Details: fmt.Sprintf("Received value: %s (must be either 'B2C' or 'B2B')", req.Model),
			Code:    http.StatusBadRequest,
		})
	}
	if req.InitialBalance < 0 {
		log.Warn("Invalid initial balance: %f", req.InitialBalance)
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   helpers.ErrMsgInitialBalanceCannotBeNegative,
			Details: fmt.Sprintf("Received value: %f", req.InitialBalance),
			Code:    http.StatusBadRequest,
		})
	}

	// 3. Извлекаем userID из контекста (установлен JWT middleware)
	userID := authMiddleware.GetUserID(c)
	if userID != nil {
		log = log.WithFields(logger.Fields{"user_id": *userID})
	}

	// 4. Call service
	result, err := h.s.GenerateTransactions(&req, userID)
	if err != nil {
		log.Error(err, "Failed to generate transactions")

		// Обработка специфичных ошибок
		if errors.Is(err, helpers.ErrUnauthorized) {
			return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error:   helpers.ErrMsgUserAuthenticationRequired,
				Details: err.Error(),
				Code:    http.StatusUnauthorized,
			})
		}
		if errors.Is(err, helpers.ErrNegativeBalance) {
			return c.JSON(http.StatusUnprocessableEntity, dto.ErrorResponse{
				Error:   helpers.ErrMsgTransactionWouldResultInNegativeBalance,
				Details: err.Error(),
				Code:    http.StatusUnprocessableEntity,
			})
		}
		// Проверка на ошибку недостаточного баланса
		if errors.Is(err, helpers.ErrInsufficientBalance) {
			return c.JSON(http.StatusUnprocessableEntity, dto.ErrorResponse{
				Error:   helpers.ErrMsgInsufficientBalance,
				Details: err.Error(),
				Code:    http.StatusUnprocessableEntity,
			})
		}
		if errors.Is(err, helpers.ErrInvalidModel) {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   helpers.ErrMsgInvalidBusinessModel,
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
		}
		// Общая ошибка сервера
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   helpers.ErrMsgFailedToGenerateStatement,
			Details: err.Error(),
			Code:    http.StatusInternalServerError,
		})
	}

	log.WithFields(logger.Fields{
		"request_id": result.RequestID,
		"total_transactions": result.TransactionCounts.Total,
	}).Success("Transactions generated successfully")

	return c.JSON(http.StatusOK, result)
}
