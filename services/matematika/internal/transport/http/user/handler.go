package user

import (
	"errors"
	"net/http"
	"time"

	authMiddleware "github.com/IbadT/business_bank_back/services/matematika/internal/middleware"
	userservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/user"
	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/dto"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/helpers"
	jwt_pkg "github.com/IbadT/business_bank_back/services/matematika/pkg/jwt"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/logger"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	userService userservice.UserService
}

func NewHandler(s userservice.UserService) *Handler {
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
	op := "http.handler.user.saveAssociatedCard"
	log := logger.GetLogger().WithOperation(op)
	
	var req dto.SaveAssociatedCardRequest

	// 1. Парсим входные данные
	if err := c.Bind(&req); err != nil {
		log.Error(err, "Invalid request body")
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid request body",
			"details": err.Error(),
			"code":    http.StatusBadRequest,
		})
	}

	// 2. Извлекаем userID из контекста (установлен JWT middleware)
	userIDStr := authMiddleware.GetUserID(c)
	if userIDStr == nil {
		log.Warn("User ID not found in context")
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error":   "Unauthorized",
			"details": "User ID is required",
			"code":    http.StatusUnauthorized,
		})
	}

	log = log.WithFields(logger.Fields{
		"user_id": *userIDStr,
		"card":    req.AssociatedCard,
	})
	log.Info("Saving associated card")

	// 3. Вызываем сервис для сохранения номера карты
	if err := h.userService.SaveAssociatedCard(*userIDStr, req.AssociatedCard); err != nil {
		log.Error(err, "Failed to save associated card", logger.Fields{"user_id": *userIDStr})
		statusCode := http.StatusInternalServerError
		if errors.Is(err, helpers.ErrUserIDRequired) || errors.Is(err, helpers.ErrInvalidUserID) {
			statusCode = http.StatusBadRequest
		} else if errors.Is(err, helpers.ErrUserNotFound) || errors.Is(err, helpers.ErrUserNotFoundOrNoChanges) {
			statusCode = http.StatusNotFound
		} else if errors.Is(err, helpers.ErrAssociatedCardRequired) || errors.Is(err, helpers.ErrAssociatedCardInvalidLength) || errors.Is(err, helpers.ErrAssociatedCardInvalidFormat) {
			statusCode = http.StatusBadRequest
		}
		return c.JSON(statusCode, map[string]interface{}{
			"error":   "Failed to save associated card",
			"details": err.Error(),
			"code":    statusCode,
		})
	}

	log.Success("Associated card saved successfully")

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
	op := "http.handler.user.login"
	log := logger.GetLogger().WithOperation(op)
	
	var req dto.LoginRequest

	if err := c.Bind(&req); err != nil {
		log.Error(err, "Invalid request body")
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	log.Info("Login attempt for email: %s", req.Email)

	token, err := h.userService.Login(req.Email, req.Password)
	if err != nil {
		log.Error(err, "Login failed for email: %s", req.Email)
		statusCode := http.StatusUnauthorized
		if errors.Is(err, helpers.ErrUserNotFound) || errors.Is(err, helpers.ErrInvalidPassword) {
			statusCode = http.StatusUnauthorized
		}
		return c.JSON(statusCode, map[string]interface{}{
			"error":   "Invalid email or password",
			"details": err.Error(),
			"code":    statusCode,
		})
	}

	// Устанавливаем access_token в cookie
	accessCookie := jwt_pkg.SetCookies(token.AccessToken, "access_token", time.Hour*4) // 4 часа
	c.SetCookie(accessCookie)

	// Устанавливаем refresh_token в cookie
	refreshCookie := jwt_pkg.SetCookies(token.RefreshToken, "refresh_token", time.Hour*24*2) // 2 дня
	c.SetCookie(refreshCookie)

	log.WithFields(logger.Fields{"email": req.Email}).Success("User logged in successfully")

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
	op := "http.handler.user.register"
	log := logger.GetLogger().WithOperation(op)
	
	var req dto.RegisterRequest

	if err := c.Bind(&req); err != nil {
		log.Error(err, "Invalid request body")
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid request body",
			"details": err.Error(),
			"code":    http.StatusBadRequest,
		})
	}

	log.Info("Registration attempt for email: %s", req.Email)

	token, err := h.userService.Register(req.Email, req.Password)
	if err != nil {
		log.Error(err, "Registration failed for email: %s", req.Email)
		statusCode := http.StatusBadRequest
		if errors.Is(err, helpers.ErrUserAlreadyExists) {
			statusCode = http.StatusConflict
		} else if errors.Is(err, helpers.ErrInvalidEmail) || errors.Is(err, helpers.ErrPasswordRequired) || errors.Is(err, helpers.ErrPasswordTooShort) {
			statusCode = http.StatusBadRequest
		}
		return c.JSON(statusCode, map[string]interface{}{
			"error":   "Registration failed",
			"details": err.Error(),
			"code":    statusCode,
		})
	}

	// Устанавливаем access_token в cookie
	accessCookie := jwt_pkg.SetCookies(token.AccessToken, "access_token", time.Hour*4) // 4 часа
	c.SetCookie(accessCookie)

	// Устанавливаем refresh_token в cookie
	refreshCookie := jwt_pkg.SetCookies(token.RefreshToken, "refresh_token", time.Hour*24*2) // 2 дня
	c.SetCookie(refreshCookie)

	log.WithFields(logger.Fields{"email": req.Email}).Success("User registered successfully")

	return c.JSON(http.StatusOK, dto.TokenResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	})
}
