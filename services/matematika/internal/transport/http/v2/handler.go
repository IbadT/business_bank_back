// internal/transport/http/v2/handler.go
package v2

import (
	"errors"
	"net/http"
	"strings"
	"time"

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
	generatorService service.GeneratorService
	userService service.UserService
	holidayService service.HolidayService
}

// NewHandler создает новый v2 handler
func NewHandler(generatorService service.GeneratorService, userService service.UserService, holidayService service.HolidayService) *Handler {
	return &Handler{
		generatorService: generatorService,
		userService: userService,
		holidayService: holidayService,
	}
}

// Init регистрирует все роуты для API
func (h *Handler) Init(api *echo.Group) {
	// Statement generation endpoint
	api.POST("/generate", h.Generate)
	api.POST("/login", h.Login)
	api.POST("/register", h.Register)

	api.POST("/holidays", h.AddHoliday)
	api.GET("/holidays", h.GetHolidays)
	api.GET("/holidays/is-holiday", h.IsHoliday)
	api.PUT("/holidays/:id", h.UpdateHoliday)
	api.DELETE("/holidays/:id", h.DeleteHoliday)

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
			"error": "Invalid request body",
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
				"code": http.StatusUnprocessableEntity,
			})
		}
		if errors.Is(err, service.ErrInvalidModel) {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"error": err.Error(),
				"code": http.StatusBadRequest,
			})
		}
		// Общая ошибка сервера
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to generate statement",
			"details": err.Error(),
			"code": http.StatusInternalServerError,
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
			"error": "Invalid request body",
			"details": err.Error(),
		})
	}

	token, err := h.userService.Login(req.Email, req.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": "Invalid email or password",
			"details": err.Error(),
			"code": http.StatusUnauthorized,
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
			"error": "Invalid request body",
			"details": err.Error(),
			"code": http.StatusBadRequest,
		})
	}

	token, err := h.userService.Register(req.Email, req.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": "Invalid email or password",
			"details": err.Error(),
			"code": http.StatusBadRequest,
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
			"error": "Invalid request body",
			"details": err.Error(),
			"code": http.StatusBadRequest,
		})
	}

	// Парсим дату из формата YYYY-MM-DD
	holidayDate, err := time.Parse("2006-01-02", req.HolidayDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid date format. Expected YYYY-MM-DD (e.g., 2025-12-25)",
			"details": err.Error(),
			"code": http.StatusBadRequest,
		})
	}

	if err := h.holidayService.AddHoliday(holidayDate, req.Name, req.Country); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to add holiday",
			"details": err.Error(),
			"code": http.StatusInternalServerError,
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
			"code": http.StatusBadRequest,
		})
	}

	date, err := time.Parse("2006-01-02", reqDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid date format",
			"details": err.Error(),
			"code": http.StatusBadRequest,
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
			"code": http.StatusBadRequest,
		})
	}
	yearTime, err := time.Parse("2006", year)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid year format",
			"details": err.Error(),
			"code": http.StatusBadRequest,
		})
	}
	holidays, err := h.holidayService.GetHolidays(yearTime)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to get holidays",
			"details": err.Error(),
			"code": http.StatusInternalServerError,
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
			"code": http.StatusBadRequest,
		})
	}
	id, err := uuid.Parse(holidayID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid id format",
			"details": err.Error(),
			"code": http.StatusBadRequest,
		})
	}
	var req dto.HolidayRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid request body",
			"details": err.Error(),
			"code": http.StatusBadRequest,
		})
	}

	// Парсим дату из формата YYYY-MM-DD
	holidayDate, err := time.Parse("2006-01-02", req.HolidayDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid date format. Expected YYYY-MM-DD (e.g., 2025-12-25)",
			"details": err.Error(),
			"code": http.StatusBadRequest,
		})
	}

	if err := h.holidayService.UpdateHoliday(id, holidayDate, req.Name, req.Country); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to update holiday",
			"details": err.Error(),
			"code": http.StatusInternalServerError,
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
			"code": http.StatusBadRequest,
		})
	}

	id, err := uuid.Parse(holidayID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid id format",
			"details": err.Error(),
			"code": http.StatusBadRequest,
		})
	}
	
	if err := h.holidayService.DeleteHoliday(id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to delete holiday",
			"details": err.Error(),
			"code": http.StatusInternalServerError,
		})
	}
	return c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Holiday deleted successfully",
		Code:    http.StatusOK,
	})
}