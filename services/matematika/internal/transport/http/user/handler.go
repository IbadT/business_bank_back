package user

import (
	"net/http"
	"time"

	authMiddleware "github.com/IbadT/business_bank_back/services/matematika/internal/middleware"
	"github.com/IbadT/business_bank_back/services/matematika/internal/service"
	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/dto"
	jwt_pkg "github.com/IbadT/business_bank_back/services/matematika/pkg/jwt"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	userService service.UserService
}

func NewHandler(s service.UserService) *Handler {
	return &Handler{s}
}

// ========================= USER =========================
// SaveAssociatedCard - сохранение номера карты
// @Summary      Сохранение номера карты
// @Description  Сохраняет номер карты для пользователя. Требуется авторизация.
// @Tags         user
// @Accept       json
// @Produce      json
// @security     BearerAuth
// @Param        request  body      dto.SaveAssociatedCardRequest  true  "Данные для сохранения номера карты"
// @Success      200      {object}  dto.SaveAssociatedCardResponse  "Успешное сохранение номера карты"
// @Failure      400      {object}  map[string]interface{}  "Некорректный запрос - ошибки валидации входных параметров"
// @Failure      401      {object}  map[string]string     "Требуется авторизация"
// @Failure      500      {object}  map[string]interface{}  "Внутренняя ошибка сервера"
// @Router       /api/user/associated-card [put]
func (h *Handler) SaveAssociatedCard(c echo.Context) error {
	var req dto.SaveAssociatedCardRequest

	// 1. Парсим входные данные
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid request body",
			"details": err.Error(),
			"code":    http.StatusBadRequest,
		})
	}

	// 2. Извлекаем userID из контекста (установлен JWT middleware)
	userIDStr := authMiddleware.GetUserID(c)
	if userIDStr == nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error":   "Unauthorized",
			"details": "User ID is required",
			"code":    http.StatusUnauthorized,
		})
	}

	// 3. Вызываем сервис для сохранения номера карты
	if err := h.userService.SaveAssociatedCard(*userIDStr, req.AssociatedCard); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to save associated card",
			"details": err.Error(),
			"code":    http.StatusInternalServerError,
		})
	}

	return c.JSON(http.StatusOK, dto.SaveAssociatedCardResponse{
		Message: "Associated card saved successfully",
		Code:    http.StatusOK,
	})
}

// ========================= AUTH =========================
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

	// Устанавливаем access_token в cookie
	accessCookie := jwt_pkg.SetCookies(token.AccessToken, "access_token", time.Hour*4) // 4 часа
	c.SetCookie(accessCookie)

	// Устанавливаем refresh_token в cookie
	refreshCookie := jwt_pkg.SetCookies(token.RefreshToken, "refresh_token", time.Hour*24*2) // 2 дня
	c.SetCookie(refreshCookie)

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

	// Устанавливаем access_token в cookie
	accessCookie := jwt_pkg.SetCookies(token.AccessToken, "access_token", time.Hour*4) // 4 часа
	c.SetCookie(accessCookie)

	// Устанавливаем refresh_token в cookie
	refreshCookie := jwt_pkg.SetCookies(token.RefreshToken, "refresh_token", time.Hour*24*2) // 2 дня
	c.SetCookie(refreshCookie)

	return c.JSON(http.StatusOK, dto.TokenResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	})
}
