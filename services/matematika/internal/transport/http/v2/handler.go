// internal/transport/http/v2/handler.go
package v2

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/IbadT/business_bank_back/services/matematika/internal/service"
	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/dto"
	authMiddleware "github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/middleware"
	"github.com/labstack/echo/v4"
)

// contains проверяет, содержит ли строка подстроку
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// Handler - HTTP handlers для API v2
type Handler struct {
	generatorService service.GeneratorService
	userService service.UserService
}

// NewHandler создает новый v2 handler
func NewHandler(generatorService service.GeneratorService, userService service.UserService) *Handler {
	return &Handler{
		generatorService: generatorService,
		userService: userService,
	}
}

// Init регистрирует все роуты для API
func (h *Handler) Init(api *echo.Group) {
	// Statement generation endpoint
	api.POST("/generate", h.Generate)
	api.POST("/login", h.Login)
	api.POST("/register", h.Register)
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
				"code": "INSUFFICIENT_BALANCE",
			})
		}
		if errors.Is(err, service.ErrInvalidModel) {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"error": err.Error(),
				"code": "INVALID_MODEL",
			})
		}
		// Общая ошибка сервера
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to generate statement",
			"details": err.Error(),
			"code": "INTERNAL_ERROR",
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

	// if err := h.validate.Validate(req); err != nil {
	// 	return c.JSON(http.StatusBadRequest, map[string]string{
	// 		"error": err.Error(),
	// 		"code": "VALIDATION_ERROR",
	// 	})
	// }

	token, err := h.userService.Login(req.Email, req.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid email or password",
			"details": err.Error(),
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
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
			"details": err.Error(),
		})
	}
	fmt.Println("REQ: ", req)

	token, err := h.userService.Register(req.Email, req.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid email or password",
			"details": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, dto.TokenResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	})
}