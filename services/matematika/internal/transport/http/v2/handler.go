// internal/transport/http/v2/handler.go
package v2

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/internal/domain"
	"github.com/IbadT/business_bank_back/services/matematika/internal/service"
	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/dto"
	authMiddleware "github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/middleware"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// TODO: переделать на уже существующие функции валидации и использовать strings.Contains
// contains проверяет, содержит ли строка подстроку
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// Handler - HTTP handlers для API v2
type Handler struct {
	generatorService   service.GeneratorService
	userService        service.UserService
	holidayService     service.HolidayService
	transactionService service.TransactionService
	gatewayService     service.GatewayService
	breakdownService   service.BreakdownService
}

// NewHandler создает новый v2 handler
func NewHandler(generatorService service.GeneratorService,
	userService service.UserService,
	holidayService service.HolidayService,
	transactionService service.TransactionService,
	gatewayService service.GatewayService,
	breakdownService service.BreakdownService,
) *Handler {
	return &Handler{
		generatorService:   generatorService,
		userService:        userService,
		holidayService:     holidayService,
		transactionService: transactionService,
		gatewayService:     gatewayService,
		breakdownService:   breakdownService,
	}
}

// Init регистрирует все роуты для API
func (h *Handler) Init(api *echo.Group) {
	// STATEMENTS
	api.POST("/generate", h.Generate)

	// AUTH
	api.POST("/login", h.Login)
	api.POST("/register", h.Register)

	// HOLIDAYS
	api.POST("/holidays", h.AddHoliday)
	api.GET("/holidays", h.GetHolidays)
	api.GET("/holidays/is-holiday", h.IsHoliday)
	api.PUT("/holidays/:id", h.UpdateHoliday)
	api.DELETE("/holidays/:id", h.DeleteHoliday)

	// TRANSACTIONS
	api.POST("/transactions", h.CreateTransaction)
	api.POST("/transactions/batch", h.CreateBatchTransactions)
	api.GET("/transactions/count/:request_id", h.GetTransactionsCount)
	api.GET("/transactions/type/:type/:request_id", h.GetTransactionsByTypeAndRequestID)
	api.GET("/transactions/method/:method/:request_id", h.GetTransactionsByMethodAndRequestID)
	api.GET("/transactions/:request_id", h.GetTransactionsByRequestID)

	// GATEWAY (шлюзы) - пользовательские роуты
	api.GET("/gateway/b2c", h.GetB2CGateways)
	api.PUT("/gateway/b2c", h.UpdateB2CGateways)
	api.DELETE("/gateway/b2c", h.DeleteB2CGateways)

	// GATEWAY (шлюзы) - администраторские роуты
	// TODO: Реализовать администраторские роуты для управления шлюзами:
	// TODO: Все администраторские роуты должны проверять роль пользователя (только admin)
	// - GET /api/admin/gateways - получить список всех доступных шлюзов из gateways.csv
	// - POST /api/admin/gateways - добавить новый шлюз (требуется обновление gateways.csv)
	// - PUT /api/admin/gateways/:id - обновить шлюз (требуется обновление gateways.csv)
	// - DELETE /api/admin/gateways/:id - удалить шлюз (требуется обновление gateways.csv)
	// - GET /api/admin/gateways/users - получить список всех пользователей с их выбранными шлюзами
	// - GET /api/admin/gateways/users/:user_id - получить выбранный шлюз конкретного пользователя
	// - PUT /api/admin/gateways/users/:user_id - установить шлюз для конкретного пользователя (принудительно)
	// - DELETE /api/admin/gateways/users/:user_id - удалить выбранный шлюз для конкретного пользователя

	api.GET("/breakdown/revenue/:request_id", h.CalculateRevenueBreakdown)
	api.GET("/breakdown/expenses/:request_id", h.CalculateExpensesBreakdown)
}

// Generate - генерация финансовой выписки
// @Summary      Генерация финансовой выписки
// @Description  Генерирует финансовую выписку с транзакциями на основе переданных параметров. Поддерживает модели B2C и B2B, позволяет задавать желаемый процент прибыли, начальный баланс и дополнительные кастомные данные.
// @Tags         statements
// @Accept       json
// @Produce      json
// @security     BearerAuth
// @Param        request  body      dto.GenerateRequest  true  "Параметры генерации выписки"
// @Success      200      {object}  dto.GenerateResponse  "Успешная генерация выписки"
// @Failure      400      {object}  map[string]string     "Некорректный запрос - ошибки валидации входных параметров"
// @Failure      401      {object}  map[string]string     "Требуется авторизация"
// @Failure      422      {object}  map[string]string     "Ошибка валидации - транзакция приведет к отрицательному балансу"
// @Failure      500      {object}  map[string]string     "Внутренняя ошибка сервера"
// @Router       /api/generate [post]
func (h *Handler) Generate(c echo.Context) error {
	var req dto.GenerateRequest

	// 1. Парсим входные данные
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	// 2. Валидация базовых полей
	if req.Turnover <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "turnover must be greater than 0",
		})
	}
	if req.DesiredProfitPercent < 0 || req.DesiredProfitPercent > 100 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "desiredProfitPercent must be between 0 and 100",
		})
	}
	if req.Model != "B2C" && req.Model != "B2B" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "model must be either B2C or B2B",
		})
	}
	if req.InitialBalance < 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "initialBalance cannot be negative",
		})
	}

	// 3. Извлекаем userID из контекста (установлен JWT middleware)
	userID := authMiddleware.GetUserID(c)

	// 4. Call service
	result, err := h.generatorService.GenerateTransactions(&req, userID)
	if err != nil {
		// Логируем ошибку для отладки
		c.Logger().Errorf("GenerateTransactions error: %v", err)

		// Обработка специфичных ошибок
		if errors.Is(err, service.ErrNegativeBalance) {
			return c.JSON(http.StatusUnprocessableEntity, map[string]string{
				"error": err.Error(),
			})
		}
		// Проверка на ошибку недостаточного баланса (может быть в тексте ошибки)
		errMsg := err.Error()
		if errMsg != "" &&
			(contains(errMsg, "insufficient balance") || contains(errMsg, "negative balance")) {
			return c.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
				"error": errMsg,
				"code":  http.StatusUnprocessableEntity,
			})
		}
		if errors.Is(err, service.ErrInvalidModel) {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"error": err.Error(),
				"code":  http.StatusBadRequest,
			})
		}
		// Общая ошибка сервера
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to generate statement",
			"details": err.Error(),
			"code":    http.StatusInternalServerError,
		})
	}

	return c.JSON(http.StatusOK, result)
}

// Login - авторизация пользователя
// @Summary      Авторизация пользователя
// @Description  Авторизует пользователя по email и паролю. Возвращает токены доступа и обновления.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.LoginRequest  true  "Данные для авторизации"
// @Success      200      {object}  dto.TokenResponse  "Успешная авторизация"
// @Failure      400      {object}  map[string]string     "Некорректный запрос - ошибки валидации входных параметров"
// @Failure      401      {object}  map[string]string     "Неверный email или пароль"
// @Failure      500      {object}  map[string]string     "Внутренняя ошибка сервера"
// @Router       /api/login [post]
func (h *Handler) Login(c echo.Context) error {
	var req dto.LoginRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	token, err := h.userService.Login(req.Email, req.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error":   "Invalid email or password",
			"details": err.Error(),
			"code":    http.StatusUnauthorized,
		})
	}

	return c.JSON(http.StatusOK, dto.TokenResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	})
}

// Register - регистрация нового пользователя
// @Summary      Регистрация пользователя
// @Description  Регистрирует нового пользователя по email и паролю. Возвращает токены доступа и обновления.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.RegisterRequest  true  "Данные для регистрации"
// @Success      200      {object}  dto.TokenResponse  "Успешная регистрация"
// @Failure      400      {object}  map[string]string     "Некорректный запрос - ошибки валидации входных параметров"
// @Failure      409      {object}  map[string]string     "Пользователь с таким email уже существует"
// @Failure      500      {object}  map[string]string     "Внутренняя ошибка сервера"
// @Router       /api/register [post]
func (h *Handler) Register(c echo.Context) error {
	var req dto.RegisterRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid request body",
			"details": err.Error(),
			"code":    http.StatusBadRequest,
		})
	}

	token, err := h.userService.Register(req.Email, req.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error":   "Invalid email or password",
			"details": err.Error(),
			"code":    http.StatusBadRequest,
		})
	}

	return c.JSON(http.StatusOK, dto.TokenResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	})
}

// AddHoliday - добавление праздника
// @Summary      Добавление праздника
// @Description  Добавляет новый праздник в базу данных. Требуется авторизация. Поддерживаемые страны: RU, US.
// @Tags         holidays
// @Accept       json
// @Produce      json
// @security     BearerAuth
// @Param        request  body      dto.HolidayRequest  true  "Данные для добавления праздника"
// @Success      200      {object}  dto.MessageResponse  "Успешное добавление праздника"
// @Failure      400      {object}  map[string]interface{}  "Некорректный запрос - ошибки валидации входных параметров"
// @Failure      401      {object}  map[string]string     "Требуется авторизация"
// @Failure      500      {object}  map[string]interface{}  "Внутренняя ошибка сервера"
// @Router       /api/holidays [post]
func (h *Handler) AddHoliday(c echo.Context) error {
	var req dto.HolidayRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid request body",
			"details": err.Error(),
			"code":    http.StatusBadRequest,
		})
	}

	// Парсим дату из формата YYYY-MM-DD
	holidayDate, err := time.Parse("2006-01-02", req.HolidayDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid date format. Expected YYYY-MM-DD (e.g., 2025-12-25)",
			"details": err.Error(),
			"code":    http.StatusBadRequest,
		})
	}

	if err := h.holidayService.AddHoliday(holidayDate, req.Name, req.Country); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to add holiday",
			"details": err.Error(),
			"code":    http.StatusInternalServerError,
		})
	}
	return c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Holiday added successfully",
		Code:    http.StatusOK,
	})
}

// IsHoliday - проверка является ли дата праздником
// @Summary      Проверка является ли дата праздником
// @Description  Проверяет, является ли указанная дата праздником. Формат даты: YYYY-MM-DD (например, 2024-12-15). Требуется авторизация.
// @Tags         holidays
// @Accept       json
// @Produce      json
// @security     BearerAuth
// @Param        date  query      string  true  "Дата для проверки в формате YYYY-MM-DD" example:"2024-12-15"
// @Success      200      {object}  dto.IsHolidayResponse  "Результат проверки даты"
// @Failure      400      {object}  map[string]interface{}  "Некорректный запрос - ошибки валидации входных параметров"
// @Failure      401      {object}  map[string]string     "Требуется авторизация"
// @Failure      500      {object}  map[string]interface{}  "Внутренняя ошибка сервера"
// @Router       /api/holidays/is-holiday [get]
func (h *Handler) IsHoliday(c echo.Context) error {
	reqDate := c.QueryParam("date")
	if reqDate == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "date parameter is required",
			"code":  http.StatusBadRequest,
		})
	}

	date, err := time.Parse("2006-01-02", reqDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid date format",
			"details": err.Error(),
			"code":    http.StatusBadRequest,
		})
	}
	isHoliday := h.holidayService.IsHoliday(date)
	return c.JSON(http.StatusOK, dto.IsHolidayResponse{
		IsHoliday: isHoliday,
		Date:      reqDate,
	})
}

// GetHolidays - получить список всех праздников для указанного года
// @Summary      Получить список всех праздников для указанного года
// @Description  Получает список всех праздников для указанного года. Формат года: YYYY (например, 2024). Требуется авторизация.
// @Tags         holidays
// @Accept       json
// @Produce      json
// @security     BearerAuth
// @Param        year  query      string  true  "Год для получения праздников в формате YYYY" example:"2024"
// @Success      200      {object}  dto.GetHolidaysResponse  "Успешное получение списка праздников"
// @Failure      400      {object}  map[string]interface{}  "Некорректный запрос - ошибки валидации входных параметров"
// @Failure      401      {object}  map[string]string     "Требуется авторизация"
// @Failure      500      {object}  map[string]interface{}  "Внутренняя ошибка сервера"
// @Router       /api/holidays [get]
func (h *Handler) GetHolidays(c echo.Context) error {
	year := c.QueryParam("year")
	if year == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "year parameter is required",
			"code":  http.StatusBadRequest,
		})
	}
	yearTime, err := time.Parse("2006", year)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid year format",
			"details": err.Error(),
			"code":    http.StatusBadRequest,
		})
	}
	holidays, err := h.holidayService.GetHolidays(yearTime)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to get holidays",
			"details": err.Error(),
			"code":    http.StatusInternalServerError,
		})
	}

	// Конвертируем domain.Holiday в dto.HolidayResponse
	holidayResponses := make([]dto.HolidayResponse, len(holidays))
	for i, holiday := range holidays {
		holidayResponses[i] = dto.HolidayResponse{
			HolidayDate: holiday.HolidayDate,
			Name:        holiday.Name,
			Country:     holiday.Country,
		}
	}

	return c.JSON(http.StatusOK, dto.GetHolidaysResponse{
		Holidays: holidayResponses,
		Year:     year,
	})
}

// UpdateHoliday - обновление праздника
// @Summary      Обновление праздника
// @Description  Обновляет существующий праздник в базе данных по его ID. Требуется авторизация.
// @Tags         holidays
// @Accept       json
// @Produce      json
// @security     BearerAuth
// @Param        id  path      string  true  "UUID праздника" example:"550e8400-e29b-41d4-a716-446655440000"
// @Param        request  body      dto.HolidayRequest  true  "Данные для обновления праздника"
// @Success      200      {object}  dto.MessageResponse  "Успешное обновление праздника"
// @Failure      400      {object}  map[string]interface{}  "Некорректный запрос - ошибки валидации входных параметров"
// @Failure      401      {object}  map[string]string     "Требуется авторизация"
// @Failure      404      {object}  map[string]string     "Праздник не найден"
// @Failure      500      {object}  map[string]interface{}  "Внутренняя ошибка сервера"
// @Router       /api/holidays/{id} [put]
func (h *Handler) UpdateHoliday(c echo.Context) error {
	holidayID := c.Param("id")
	if holidayID == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "id parameter is required",
			"code":  http.StatusBadRequest,
		})
	}
	id, err := uuid.Parse(holidayID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid id format",
			"details": err.Error(),
			"code":    http.StatusBadRequest,
		})
	}
	var req dto.HolidayRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid request body",
			"details": err.Error(),
			"code":    http.StatusBadRequest,
		})
	}

	// Парсим дату из формата YYYY-MM-DD
	holidayDate, err := time.Parse("2006-01-02", req.HolidayDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid date format. Expected YYYY-MM-DD (e.g., 2025-12-25)",
			"details": err.Error(),
			"code":    http.StatusBadRequest,
		})
	}

	if err := h.holidayService.UpdateHoliday(id, holidayDate, req.Name, req.Country); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to update holiday",
			"details": err.Error(),
			"code":    http.StatusInternalServerError,
		})
	}
	return c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Holiday updated successfully",
		Code:    http.StatusOK,
	})
}

// DeleteHoliday - удаление праздника
// @Summary      Удаление праздника
// @Description  Удаляет праздник из базы данных по его ID. Требуется авторизация.
// @Tags         holidays
// @Accept       json
// @Produce      json
// @security     BearerAuth
// @Param        id  path      string  true  "UUID праздника" example:"550e8400-e29b-41d4-a716-446655440000"
// @Success      200      {object}  dto.MessageResponse  "Успешное удаление праздника"
// @Failure      400      {object}  map[string]interface{}  "Некорректный запрос - ошибки валидации входных параметров"
// @Failure      401      {object}  map[string]string     "Требуется авторизация"
// @Failure      404      {object}  map[string]string     "Праздник не найден"
// @Failure      500      {object}  map[string]interface{}  "Внутренняя ошибка сервера"
// @Router       /api/holidays/{id} [delete]
func (h *Handler) DeleteHoliday(c echo.Context) error {
	holidayID := c.Param("id")
	if holidayID == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "id parameter is required",
			"code":  http.StatusBadRequest,
		})
	}

	id, err := uuid.Parse(holidayID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid id format",
			"details": err.Error(),
			"code":    http.StatusBadRequest,
		})
	}

	if err := h.holidayService.DeleteHoliday(id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to delete holiday",
			"details": err.Error(),
			"code":    http.StatusInternalServerError,
		})
	}
	return c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Holiday deleted successfully",
		Code:    http.StatusOK,
	})
}

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

	if err := h.transactionService.CreateTransaction(&req); err != nil {
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

	transactions, err := h.transactionService.GetByRequestID(requestID)
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

	count, err := h.transactionService.GetCountByRequestID(requestID)
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

	if err := h.transactionService.CreateBatchTransactions(&req); err != nil {
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

	transactions, err := h.transactionService.GetByTypeAndRequestID(transactionType, requestID)
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

	transactions, err := h.transactionService.GetByMethodAndRequestID(transactionMethod, requestID)
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

// GetB2CGateways - получение списка шлюзов для B2C
// @Summary      Получение списка шлюзов для B2C
// @Description  Получает список шлюзов для B2C. Требуется авторизация.
// @Tags         gateway
// @Accept       json
// @Produce      json
// @security     BearerAuth
// @Success      200      {object}  dto.B2CGatewayResponse  "Успешное получение списка шлюзов для B2C"
// @Failure      400      {object}  map[string]interface{}  "Некорректный запрос - ошибки валидации входных параметров"
// @Failure      401      {object}  map[string]string     "Требуется авторизация"
// @Failure      500      {object}  map[string]interface{}  "Внутренняя ошибка сервера"
// @Router       /api/gateway/b2c [get]
func (h *Handler) GetB2CGateways(c echo.Context) error {
	userIDStr := authMiddleware.GetUserID(c)

	userID, err := uuid.Parse(*userIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid userID format",
			"details": err.Error(),
			"code":    http.StatusBadRequest,
		})
	}

	gateway, err := h.gatewayService.GetB2CGateways(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to get B2C gateway",
			"details": err.Error(),
			"code":    http.StatusInternalServerError,
		})
	}

	// Если шлюз не найден - возвращаем 404
	if gateway == nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"error":   "B2C gateway not found",
			"message": "No gateway has been saved for this user. A gateway will be automatically selected during the first B2C generation.",
			"code":    http.StatusNotFound,
		})
	}

	return c.JSON(http.StatusOK, dto.B2CGatewayResponse{
		Gateway: domain.Gateway{
			ID:   gateway.ID,
			Name: gateway.Name,
		},
		Code: http.StatusOK,
	})
}

// UpdateB2CGateways - обновление списка шлюзов для B2C
// @Summary      Обновление шлюза для B2C
// @Description  Обновляет выбранный шлюз для B2C. Если gateway_id не указан, выбирается случайный шлюз. Требуется авторизация.
// @Tags         gateway
// @Accept       json
// @Produce      json
// @security     BearerAuth
// @Param        request  body      dto.UpdateB2CGatewayRequest  true  "Данные для обновления шлюза"
// @Success      200      {object}  dto.MessageResponse          "Успешное обновление шлюза"
// @Failure      400      {object}  map[string]interface{}      "Некорректный запрос"
// @Failure      401      {object}  map[string]string           "Требуется авторизация"
// @Failure      500      {object}  map[string]interface{}     "Внутренняя ошибка сервера"
// @security     BearerAuth
// @Router       /api/gateway/b2c [put]
func (h *Handler) UpdateB2CGateways(c echo.Context) error {
	userIDStr := authMiddleware.GetUserID(c)

	userID, err := uuid.Parse(*userIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid userID format",
			"details": err.Error(),
			"code":    http.StatusBadRequest,
		})
	}

	var req dto.UpdateB2CGatewayRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid request body",
			"details": err.Error(),
			"code":    http.StatusBadRequest,
		})
	}

	if err := h.gatewayService.SaveB2CGateways(userID, req.GatewayID); err != nil {
		// Проверяем, является ли ошибка "gateway not found"
		if err.Error() == "gateway not found" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"error":   "Invalid gateway ID",
				"details": "The specified gateway ID does not exist in the available gateways list",
				"code":    http.StatusBadRequest,
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to update B2C gateway",
			"details": err.Error(),
			"code":    http.StatusInternalServerError,
		})
	}

	return c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "B2C gateways updated successfully",
		Code:    http.StatusOK,
	})
}

// DeleteB2CGateways - удаление списка шлюзов для B2C
// @Summary      Удаление шлюза для B2C
// @Description  Удаляет сохраненный шлюз для B2C. При следующей генерации будет выбран новый случайный шлюз. Требуется авторизация.
// @Tags         gateway
// @Accept       json
// @Produce      json
// @security     BearerAuth
// @Success      200      {object}  dto.MessageResponse          "Успешное удаление шлюза"
// @Failure      400      {object}  map[string]interface{}        "Некорректный запрос"
// @Failure      401      {object}  map[string]string           "Требуется авторизация"
// @Failure      500      {object}  map[string]interface{}     "Внутренняя ошибка сервера"
// @Router       /api/gateway/b2c [delete]
func (h *Handler) DeleteB2CGateways(c echo.Context) error {
	userIDStr := authMiddleware.GetUserID(c)

	userID, err := uuid.Parse(*userIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid userID format",
			"details": err.Error(),
			"code":    http.StatusBadRequest,
		})
	}

	if err := h.gatewayService.DeleteB2CGateways(userID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to delete B2C gateways",
			"details": err.Error(),
			"code":    http.StatusInternalServerError,
		})
	}

	return c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "B2C gateways deleted successfully",
		Code:    http.StatusOK,
	})
}

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
// @Router       /api/breakdown/revenue/{request_id} [get]
func (h *Handler) CalculateRevenueBreakdown(c echo.Context) error {
	requestIDStr := c.Param("request_id")
	if requestIDStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "request_id parameter is required",
			"code":  http.StatusBadRequest,
		})
	}

	result, err := h.breakdownService.GetRevenueBreakdown(requestIDStr)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorMessage := "Failed to get revenue breakdown"

		if errors.Is(err, service.ErrInvalidRequestID) {
			statusCode = http.StatusBadRequest
			errorMessage = "Invalid request_id format. Expected UUID format (e.g., 550e8400-e29b-41d4-a716-446655440000)"
		}

		return c.JSON(statusCode, map[string]interface{}{
			"error":   errorMessage,
			"details": err.Error(),
			"code":    statusCode,
		})
	}

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
// @Router       /api/breakdown/expenses/{request_id} [get]
func (h *Handler) CalculateExpensesBreakdown(c echo.Context) error {
	requestIDStr := c.Param("request_id")
	if requestIDStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "request_id parameter is required",
			"code":  http.StatusBadRequest,
		})
	}

	result, err := h.breakdownService.GetExpensesBreakdown(requestIDStr)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorMessage := "Failed to get expenses breakdown"

		if errors.Is(err, service.ErrInvalidRequestID) {
			statusCode = http.StatusBadRequest
			errorMessage = "Invalid request_id format. Expected UUID format (e.g., 550e8400-e29b-41d4-a716-446655440000)"
		}

		return c.JSON(statusCode, map[string]interface{}{
			"error":   errorMessage,
			"details": err.Error(),
			"code":    statusCode,
		})
	}

	return c.JSON(http.StatusOK, dto.CalculateExpensesBreakdownResponse{
		RequestID: requestIDStr,
		ExpensesBreakdown: dto.ExpensesBreakdown{
			ByCard:    result.ByCard,
			ByAccount: result.ByAccount,
		},
		Code: http.StatusOK,
	})
}
