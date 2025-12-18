package gateway

import (
	"net/http"

	"github.com/IbadT/business_bank_back/services/matematika/internal/domain"
	authMiddleware "github.com/IbadT/business_bank_back/services/matematika/internal/middleware"
	"github.com/IbadT/business_bank_back/services/matematika/internal/service"
	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/dto"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	s service.GatewayService
}

func NewHandler(s service.GatewayService) *Handler {
	return &Handler{s}
}

// ========================= GATEWAY =========================
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
// @Router       /api/gateways/b2c [get]
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

	gateway, err := h.s.GetB2CGateways(userID)
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
// @Router       /api/gateways/b2c [put]
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

	if err := h.s.SaveB2CGateways(userID, req.GatewayID); err != nil {
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
// @Router       /api/gateways/b2c [delete]
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

	if err := h.s.DeleteB2CGateways(userID); err != nil {
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
