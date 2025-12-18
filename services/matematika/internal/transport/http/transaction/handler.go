package transactions

import (
	"errors"
	"net/http"

	"github.com/IbadT/business_bank_back/services/matematika/internal/service"
	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/dto"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	s service.TransactionService
}

func NewHandler(s service.TransactionService) *Handler {
	return &Handler{s}
}

// ========================= TRANSACTIONS =========================
// CreateTransaction - создание транзакции
// @Summary      Создание транзакции
// @Description  Создает новую транзакцию на основе переданных параметров. Требуется авторизация.
// @Tags         transactions
// @Accept       json
// @Produce      json
// @security     BearerAuth
// @Param        request  body      dto.CreateTransactionRequest  true  "Данные для создания транзакции"
// @Success      200      {object}  dto.MessageResponse  "Успешное создание транзакции"
// @Failure      400      {object}  map[string]interface{}  "Некорректный запрос - ошибки валидации входных параметров"
// @Failure      401      {object}  map[string]string     "Требуется авторизация"
// @Failure      500      {object}  map[string]interface{}  "Внутренняя ошибка сервера"
// @Router       /api/transactions [post]
func (h *Handler) CreateTransaction(c echo.Context) error {
	var req dto.CreateTransactionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid request body",
			"details": err.Error(),
			"code":    http.StatusBadRequest,
		})
	}

	if err := h.s.CreateTransaction(&req); err != nil {
		// Определяем статус код на основе типа ошибки
		statusCode := http.StatusInternalServerError
		if errors.Is(err, service.ErrInvalidRequestID) ||
			errors.Is(err, service.ErrInvalidDate) ||
			errors.Is(err, service.ErrInvalidTransactionType) ||
			errors.Is(err, service.ErrInvalidAmount) ||
			errors.Is(err, service.ErrEmptyCategory) ||
			errors.Is(err, service.ErrEmptyMethod) {
			statusCode = http.StatusBadRequest
		}

		return c.JSON(statusCode, map[string]interface{}{
			"error":   "Failed to create transaction",
			"details": err.Error(),
			"code":    statusCode,
		})
	}

	return c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Transaction created successfully",
		Code:    http.StatusOK,
	})
}

// GetTransactionsByRequestID - получение списка транзакций по request_id
// @Summary      Получение списка транзакций по request_id
// @Description  Получает список транзакций по request_id. Требуется авторизация.
// @Tags         transactions
// @Accept       json
// @Produce      json
// @security     BearerAuth
// @Param        request_id  path      string  true  "UUID запроса" example:"550e8400-e29b-41d4-a716-446655440000"
// @Success      200      {object}  dto.GetTransactionsResponse  "Успешное получение списка транзакций"
// @Failure      400      {object}  map[string]interface{}  "Некорректный запрос - ошибки валидации входных параметров"
// @Failure      401      {object}  map[string]string     "Требуется авторизация"
// @Failure      500      {object}  map[string]interface{}  "Внутренняя ошибка сервера"
// @Router       /api/transactions/{request_id} [get]
func (h *Handler) GetTransactionsByRequestID(c echo.Context) error {
	requestID := c.Param("request_id")
	if requestID == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "request_id parameter is required",
			"code":  http.StatusBadRequest,
		})
	}

	transactions, err := h.s.GetByRequestID(requestID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, service.ErrInvalidRequestID) {
			statusCode = http.StatusBadRequest
		}

		return c.JSON(statusCode, map[string]interface{}{
			"error":   "Failed to get transactions",
			"details": err.Error(),
			"code":    statusCode,
		})
	}

	return c.JSON(http.StatusOK, dto.GetTransactionsResponse{
		Transactions: transactions,
		Code:         http.StatusOK,
	})
}

// GetTransactionsCount - получение количества транзакций по request_id
// @Summary      Получение количества транзакций по request_id
// @Description  Получает количество транзакций по request_id. Требуется авторизация.
// @Tags         transactions
// @Accept       json
// @Produce      json
// @security     BearerAuth
// @Param        request_id  path      string  true  "UUID запроса" example:"550e8400-e29b-41d4-a716-446655440000"
// @Success      200      {object}  dto.GetTransactionsCountResponse  "Успешное получение количества транзакций"
// @Failure      400      {object}  map[string]interface{}  "Некорректный запрос - ошибки валидации входных параметров"
// @Failure      401      {object}  map[string]string     "Требуется авторизация"
// @Failure      500      {object}  map[string]interface{}  "Внутренняя ошибка сервера"
// @Router       /api/transactions/count/{request_id} [get]
func (h *Handler) GetTransactionsCount(c echo.Context) error {
	requestID := c.Param("request_id")
	if requestID == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "request_id parameter is required",
			"code":  http.StatusBadRequest,
		})
	}

	count, err := h.s.GetCountByRequestID(requestID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, service.ErrInvalidRequestID) {
			statusCode = http.StatusBadRequest
		}

		return c.JSON(statusCode, map[string]interface{}{
			"error":   "Failed to get transactions count",
			"details": err.Error(),
			"code":    statusCode,
		})
	}

	return c.JSON(http.StatusOK, dto.GetTransactionsCountResponse{
		Count: count,
		Code:  http.StatusOK,
	})
}

// CreateBatchTransactions - создание пачки транзакций
// @Summary      Создание пачки транзакций
// @Description  Создает пачку транзакций на основе переданных параметров. Требуется авторизация.
// @Tags         transactions
// @Accept       json
// @Produce      json
// @security     BearerAuth
// @Param        request  body      dto.CreateBatchTransactionsRequest  true  "Данные для создания пачки транзакций"
// @Success      200      {object}  dto.MessageResponse  "Успешное создание пачки транзакций"
// @Failure      400      {object}  map[string]interface{}  "Некорректный запрос - ошибки валидации входных параметров"
// @Failure      401      {object}  map[string]string     "Требуется авторизация"
// @Failure      500      {object}  map[string]interface{}  "Внутренняя ошибка сервера"
// @Router       /api/transactions/batch [post]
func (h *Handler) CreateBatchTransactions(c echo.Context) error {
	var req dto.CreateBatchTransactionsRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid request body",
			"details": err.Error(),
			"code":    http.StatusBadRequest,
		})
	}

	if err := h.s.CreateBatchTransactions(&req); err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, service.ErrInvalidRequestID) ||
			errors.Is(err, service.ErrInvalidDate) ||
			errors.Is(err, service.ErrInvalidTransactionType) ||
			errors.Is(err, service.ErrInvalidAmount) ||
			errors.Is(err, service.ErrEmptyCategory) ||
			errors.Is(err, service.ErrEmptyMethod) {
			statusCode = http.StatusBadRequest
		}

		return c.JSON(statusCode, map[string]interface{}{
			"error":   "Failed to create batch transactions",
			"details": err.Error(),
			"code":    statusCode,
		})
	}

	return c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Batch transactions created successfully",
		Code:    http.StatusOK,
	})
}

// GetTransactionsByTypeAndRequestID - получение списка транзакций по типу и request_id
// @Summary      Получение списка транзакций по типу и request_id
// @Description  Получает список транзакций по типу и request_id. Требуется авторизация.
// @Tags         transactions
// @Accept       json
// @Produce      json
// @security     BearerAuth
// @Param        type  path      string  true  "Тип транзакции" example:"income" enums:"income,expense"
// @Param        request_id  path      string  true  "UUID запроса" example:"550e8400-e29b-41d4-a716-446655440000"
// @Success      200      {object}  dto.GetTransactionsResponse  "Успешное получение списка транзакций"
// @Failure      400      {object}  map[string]interface{}  "Некорректный запрос - ошибки валидации входных параметров"
// @Failure      401      {object}  map[string]string     "Требуется авторизация"
// @Failure      500      {object}  map[string]interface{}  "Внутренняя ошибка сервера"
// @Router       /api/transactions/type/{type}/{request_id} [get]
func (h *Handler) GetTransactionsByTypeAndRequestID(c echo.Context) error {
	transactionType := c.Param("type")
	if transactionType == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "type parameter is required",
			"code":  http.StatusBadRequest,
		})
	}

	requestID := c.Param("request_id")
	if requestID == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "request_id parameter is required",
			"code":  http.StatusBadRequest,
		})
	}

	transactions, err := h.s.GetByTypeAndRequestID(transactionType, requestID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, service.ErrInvalidRequestID) || errors.Is(err, service.ErrInvalidTransactionType) {
			statusCode = http.StatusBadRequest
		}

		return c.JSON(statusCode, map[string]interface{}{
			"error":   "Failed to get transactions",
			"details": err.Error(),
			"code":    statusCode,
		})
	}

	return c.JSON(http.StatusOK, dto.GetTransactionsResponse{
		Transactions: transactions,
		Code:         http.StatusOK,
	})
}

// GetTransactionsByMethodAndRequestID - получение списка транзакций по методу и request_id
// @Summary      Получение списка транзакций по методу и request_id
// @Description  Получает список транзакций по методу и request_id. Требуется авторизация.
// @Tags         transactions
// @Accept       json
// @Produce      json
// @security     BearerAuth
// @Param        method  path      string  true  "Метод транзакции" example:"card"
// @Param        request_id  path      string  true  "UUID запроса" example:"550e8400-e29b-41d4-a716-446655440000"
// @Success      200      {object}  dto.GetTransactionsResponse  "Успешное получение списка транзакций"
// @Failure      400      {object}  map[string]interface{}  "Некорректный запрос - ошибки валидации входных параметров"
// @Failure      401      {object}  map[string]string     "Требуется авторизация"
// @Failure      500      {object}  map[string]interface{}  "Внутренняя ошибка сервера"
// @Router       /api/transactions/method/{method}/{request_id} [get]
func (h *Handler) GetTransactionsByMethodAndRequestID(c echo.Context) error {
	transactionMethod := c.Param("method")
	if transactionMethod == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "method parameter is required",
			"code":  http.StatusBadRequest,
		})
	}

	requestID := c.Param("request_id")
	if requestID == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "request_id parameter is required",
			"code":  http.StatusBadRequest,
		})
	}

	transactions, err := h.s.GetByMethodAndRequestID(transactionMethod, requestID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, service.ErrInvalidRequestID) || errors.Is(err, service.ErrEmptyMethod) {
			statusCode = http.StatusBadRequest
		}

		return c.JSON(statusCode, map[string]interface{}{
			"error":   "Failed to get transactions",
			"details": err.Error(),
			"code":    statusCode,
		})
	}

	return c.JSON(http.StatusOK, dto.GetTransactionsResponse{
		Transactions: transactions,
		Code:         http.StatusOK,
	})
}
