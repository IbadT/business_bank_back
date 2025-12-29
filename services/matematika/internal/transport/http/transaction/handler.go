package transactions

import (
	"errors"
	"net/http"

	transactionservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/transaction"
	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/dto"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/helpers"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/logger"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	s transactionservice.TransactionService
}

func NewHandler(s transactionservice.TransactionService) *Handler {
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
// @Failure      400      {object}  dto.ErrorResponse  "Некорректный запрос - ошибки валидации входных параметров"
// @Failure      401      {object}  dto.ErrorResponse     "Требуется авторизация"
// @Failure      500      {object}  dto.ErrorResponse  "Внутренняя ошибка сервера"
// @Router       /api/transactions [post]
func (h *Handler) CreateTransaction(c echo.Context) error {
	op := "http.handler.transaction.createTransaction"
	log := logger.GetLogger().WithOperation(op)
	
	var req dto.CreateTransactionRequest
	if err := c.Bind(&req); err != nil {
		log.Error(err, "Invalid request body")
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   helpers.ErrMsgInvalidRequestBody,
			Details: err.Error(),
			Code:    http.StatusBadRequest,
		})
	}

	log = log.WithFields(logger.Fields{
		"request_id": req.RequestID,
		"type":       req.Type,
		"method":     req.Method,
		"amount":     req.Amount,
	})
	log.Info("Creating transaction")

	if err := h.s.CreateTransaction(&req); err != nil {
		log.Error(err, "Failed to create transaction")
		// Определяем статус код на основе типа ошибки
		statusCode := http.StatusInternalServerError
		if errors.Is(err, helpers.ErrInvalidRequestID) ||
			errors.Is(err, helpers.ErrInvalidDate) ||
			errors.Is(err, helpers.ErrInvalidTransactionType) ||
			errors.Is(err, helpers.ErrInvalidAmount) ||
			errors.Is(err, helpers.ErrEmptyCategory) ||
			errors.Is(err, helpers.ErrEmptyMethod) {
			statusCode = http.StatusBadRequest
		}

		return c.JSON(statusCode, dto.ErrorResponse{
			Error:   helpers.ErrMsgFailedToCreateTransaction,
			Details: err.Error(),
			Code:    statusCode,
		})
	}

	log.Success("Transaction created successfully")

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
// @Failure      400      {object}  dto.ErrorResponse  "Некорректный запрос - ошибки валидации входных параметров"
// @Failure      401      {object}  dto.ErrorResponse     "Требуется авторизация"
// @Failure      500      {object}  dto.ErrorResponse  "Внутренняя ошибка сервера"
// @Router       /api/transactions/{request_id} [get]
func (h *Handler) GetTransactionsByRequestID(c echo.Context) error {
	op := "http.handler.transaction.getTransactionsByRequestID"
	log := logger.GetLogger().WithOperation(op)
	
	requestID := c.Param("request_id")
	if requestID == "" {
		log.Warn("request_id parameter is required")
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: helpers.ErrMsgRequestIDRequired,
			Code:  http.StatusBadRequest,
		})
	}

	log = log.WithFields(logger.Fields{"request_id": requestID})
	log.Info("Getting transactions by request ID")

	transactions, err := h.s.GetByRequestID(requestID)
	if err != nil {
		log.Error(err, "Failed to get transactions by request ID")
		statusCode := http.StatusInternalServerError
		if errors.Is(err, helpers.ErrInvalidRequestID) {
			statusCode = http.StatusBadRequest
		}

		return c.JSON(statusCode, dto.ErrorResponse{
			Error:   helpers.ErrMsgFailedToGetTransactions,
			Details: err.Error(),
			Code:    statusCode,
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
// @Failure      400      {object}  dto.ErrorResponse  "Некорректный запрос - ошибки валидации входных параметров"
// @Failure      401      {object}  dto.ErrorResponse     "Требуется авторизация"
// @Failure      500      {object}  dto.ErrorResponse  "Внутренняя ошибка сервера"
// @Router       /api/transactions/count/{request_id} [get]
func (h *Handler) GetTransactionsCount(c echo.Context) error {
	op := "http.handler.transaction.getTransactionsCount"
	log := logger.GetLogger().WithOperation(op)
	
	requestID := c.Param("request_id")
	if requestID == "" {
		log.Warn("request_id parameter is required")
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: helpers.ErrMsgRequestIDRequired,
			Code:  http.StatusBadRequest,
		})
	}

	log = log.WithFields(logger.Fields{"request_id": requestID})
	log.Info("Getting transactions count")

	count, err := h.s.GetCountByRequestID(requestID)
	if err != nil {
		log.Error(err, "Failed to get transactions count")
		statusCode := http.StatusInternalServerError
		if errors.Is(err, helpers.ErrInvalidRequestID) {
			statusCode = http.StatusBadRequest
		}

		return c.JSON(statusCode, dto.ErrorResponse{
			Error:   helpers.ErrMsgFailedToGetTransactionsCount,
			Details: err.Error(),
			Code:    statusCode,
		})
	}

	log.WithFields(logger.Fields{"count": count}).Success("Transactions count retrieved")

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
// @Failure      400      {object}  dto.ErrorResponse  "Некорректный запрос - ошибки валидации входных параметров"
// @Failure      401      {object}  dto.ErrorResponse     "Требуется авторизация"
// @Failure      500      {object}  dto.ErrorResponse  "Внутренняя ошибка сервера"
// @Router       /api/transactions/batch [post]
func (h *Handler) CreateBatchTransactions(c echo.Context) error {
	op := "http.handler.transaction.createBatchTransactions"
	log := logger.GetLogger().WithOperation(op)
	
	var req dto.CreateBatchTransactionsRequest
	if err := c.Bind(&req); err != nil {
		log.Error(err, "Invalid request body")
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   helpers.ErrMsgInvalidRequestBody,
			Details: err.Error(),
			Code:    http.StatusBadRequest,
		})
	}

	log = log.WithFields(logger.Fields{
		"count": len(req.Transactions),
	})
	log.Info("Creating batch transactions")

	if err := h.s.CreateBatchTransactions(&req); err != nil {
		log.Error(err, "Failed to create batch transactions")
		statusCode := http.StatusInternalServerError
		if errors.Is(err, helpers.ErrInvalidRequestID) ||
			errors.Is(err, helpers.ErrInvalidDate) ||
			errors.Is(err, helpers.ErrInvalidTransactionType) ||
			errors.Is(err, helpers.ErrInvalidAmount) ||
			errors.Is(err, helpers.ErrEmptyCategory) ||
			errors.Is(err, helpers.ErrEmptyMethod) {
			statusCode = http.StatusBadRequest
		}

		return c.JSON(statusCode, dto.ErrorResponse{
			Error:   helpers.ErrMsgFailedToCreateBatchTransactions,
			Details: err.Error(),
			Code:    statusCode,
		})
	}

	log.Success("Batch transactions created successfully")

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
// @Failure      400      {object}  dto.ErrorResponse  "Некорректный запрос - ошибки валидации входных параметров"
// @Failure      401      {object}  dto.ErrorResponse     "Требуется авторизация"
// @Failure      500      {object}  dto.ErrorResponse  "Внутренняя ошибка сервера"
// @Router       /api/transactions/type/{type}/{request_id} [get]
func (h *Handler) GetTransactionsByTypeAndRequestID(c echo.Context) error {
	op := "http.handler.transaction.getTransactionsByTypeAndRequestID"
	log := logger.GetLogger().WithOperation(op)
	
	transactionType := c.Param("type")
	if transactionType == "" {
		log.Warn("type parameter is required")
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: helpers.ErrMsgTypeParameterRequired,
			Code:  http.StatusBadRequest,
		})
	}

	requestID := c.Param("request_id")
	if requestID == "" {
		log.Warn("request_id parameter is required")
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: helpers.ErrMsgRequestIDRequired,
			Code:  http.StatusBadRequest,
		})
	}

	log = log.WithFields(logger.Fields{
		"request_id": requestID,
		"type":       transactionType,
	})
	log.Info("Getting transactions by type and request ID")

	transactions, err := h.s.GetByTypeAndRequestID(transactionType, requestID)
	if err != nil {
		log.Error(err, "Failed to get transactions by type and request ID")
		statusCode := http.StatusInternalServerError
		if errors.Is(err, helpers.ErrInvalidRequestID) || errors.Is(err, helpers.ErrInvalidTransactionType) {
			statusCode = http.StatusBadRequest
		}

		return c.JSON(statusCode, dto.ErrorResponse{
			Error:   helpers.ErrMsgFailedToGetTransactions,
			Details: err.Error(),
			Code:    statusCode,
		})
	}

	log.WithFields(logger.Fields{"count": len(transactions)}).Success("Transactions retrieved by type and request ID")

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
// @Failure      400      {object}  dto.ErrorResponse  "Некорректный запрос - ошибки валидации входных параметров"
// @Failure      401      {object}  dto.ErrorResponse     "Требуется авторизация"
// @Failure      500      {object}  dto.ErrorResponse  "Внутренняя ошибка сервера"
// @Router       /api/transactions/method/{method}/{request_id} [get]
func (h *Handler) GetTransactionsByMethodAndRequestID(c echo.Context) error {
	op := "http.handler.transaction.getTransactionsByMethodAndRequestID"
	log := logger.GetLogger().WithOperation(op)
	
	transactionMethod := c.Param("method")
	if transactionMethod == "" {
		log.Warn("method parameter is required")
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: helpers.ErrMsgMethodParameterRequired,
			Code:  http.StatusBadRequest,
		})
	}

	requestID := c.Param("request_id")
	if requestID == "" {
		log.Warn("request_id parameter is required")
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: helpers.ErrMsgRequestIDRequired,
			Code:  http.StatusBadRequest,
		})
	}

	log = log.WithFields(logger.Fields{
		"request_id": requestID,
		"method":     transactionMethod,
	})
	log.Info("Getting transactions by method and request ID")

	transactions, err := h.s.GetByMethodAndRequestID(transactionMethod, requestID)
	if err != nil {
		log.Error(err, "Failed to get transactions by method and request ID")
		statusCode := http.StatusInternalServerError
		if errors.Is(err, helpers.ErrInvalidRequestID) || errors.Is(err, helpers.ErrEmptyMethod) {
			statusCode = http.StatusBadRequest
		}

		return c.JSON(statusCode, dto.ErrorResponse{
			Error:   helpers.ErrMsgFailedToGetTransactions,
			Details: err.Error(),
			Code:    statusCode,
		})
	}

	log.WithFields(logger.Fields{"count": len(transactions)}).Success("Transactions retrieved by method and request ID")

	return c.JSON(http.StatusOK, dto.GetTransactionsResponse{
		Transactions: transactions,
		Code:         http.StatusOK,
	})
}
