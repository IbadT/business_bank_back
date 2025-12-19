package gateway

import (
	"net/http"

	"github.com/IbadT/business_bank_back/services/matematika/internal/domain"
	authMiddleware "github.com/IbadT/business_bank_back/services/matematika/internal/middleware"
	"github.com/IbadT/business_bank_back/services/matematika/internal/models"
	gatewayservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/gateway"
	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/dto"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/helpers"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/logger"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	s gatewayservice.GatewayService
}

func NewHandler(s gatewayservice.GatewayService) *Handler {
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
	op := "http.handler.gateway.getB2CGateways"
	log := logger.GetLogger().WithOperation(op)
	
	userIDStr := authMiddleware.GetUserID(c)

	userID, err := helpers.ParseUserID(*userIDStr)
	if err != nil {
		log.Error(err, "Invalid userID format: %s", *userIDStr)
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   err.Error(),
			"details": err.Error(),
			"code":    http.StatusBadRequest,
		})
	}

	log = log.WithFields(logger.Fields{"user_id": *userIDStr})
	log.Info("Getting B2C gateways")

	gateway, err := h.s.GetB2CGateways(userID)
	if err != nil {
		log.Error(err, "Failed to get B2C gateway")
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to get B2C gateway",
			"details": err.Error(),
			"code":    http.StatusInternalServerError,
		})
	}

	// Если шлюз не найден - возвращаем 404
	if gateway == nil {
		log.Warn("B2C gateway not found for user")
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"error":   "B2C gateway not found",
			"message": "No gateway has been saved for this user. A gateway will be automatically selected during the first B2C generation.",
			"code":    http.StatusNotFound,
		})
	}

	log.WithFields(logger.Fields{
		"gateway_id":   gateway.ID,
		"gateway_name": gateway.Name,
	}).Success("B2C gateway retrieved")

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
	op := "http.handler.gateway.updateB2CGateways"
	log := logger.GetLogger().WithOperation(op)
	
	userIDStr := authMiddleware.GetUserID(c)

	userID, err := helpers.ParseUserID(*userIDStr)
	if err != nil {
		log.Error(err, "Invalid userID format: %s", *userIDStr)
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   err.Error(),
			"details": err.Error(),
			"code":    http.StatusBadRequest,
		})
	}

	var req dto.UpdateB2CGatewayRequest
	if err := c.Bind(&req); err != nil {
		log.Error(err, "Invalid request body")
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid request body",
			"details": err.Error(),
			"code":    http.StatusBadRequest,
		})
	}

	log = log.WithFields(logger.Fields{
		"user_id":   *userIDStr,
		"gateway_id": req.GatewayID,
	})
	log.Info("Updating B2C gateway")

	if err := h.s.SaveB2CGateways(userID, req.GatewayID); err != nil {
		// Проверяем, является ли ошибка "gateway not found"
		if err.Error() == "gateway not found" {
			log.Warn("Gateway not found: %s", req.GatewayID)
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"error":   "Invalid gateway ID",
				"details": "The specified gateway ID does not exist in the available gateways list",
				"code":    http.StatusBadRequest,
			})
		}
		log.Error(err, "Failed to update B2C gateway")
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to update B2C gateway",
			"details": err.Error(),
			"code":    http.StatusInternalServerError,
		})
	}

	log.Success("B2C gateway updated successfully")

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
	op := "http.handler.gateway.deleteB2CGateways"
	log := logger.GetLogger().WithOperation(op)
	
	userIDStr := authMiddleware.GetUserID(c)

	userID, err := helpers.ParseUserID(*userIDStr)
	if err != nil {
		log.Error(err, "Invalid userID format: %s", *userIDStr)
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   err.Error(),
			"details": err.Error(),
			"code":    http.StatusBadRequest,
		})
	}

	log = log.WithFields(logger.Fields{"user_id": *userIDStr})
	log.Info("Deleting B2C gateway")

	if err := h.s.DeleteB2CGateways(userID); err != nil {
		log.Error(err, "Failed to delete B2C gateway")
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to delete B2C gateways",
			"details": err.Error(),
			"code":    http.StatusInternalServerError,
		})
	}

	log.Success("B2C gateway deleted successfully")

	return c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "B2C gateways deleted successfully",
		Code:    http.StatusOK,
	})
}



func (h *Handler) GetAdminGateways(c echo.Context) error {
	userRole := authMiddleware.GetUserRole(c)
	if userRole != models.RoleAdmin {
		return c.JSON(http.StatusForbidden, map[string]interface{}{
			"error":   "Insufficient permissions. Required role: admin",
			"details": "Only administrators can access this resource",
			"code":    http.StatusForbidden,
		})
	}
	return h.s.GetAdminGateways()
}

func (h *Handler) GetAdminUsersGateways(c echo.Context) error {
	userRole := authMiddleware.GetUserRole(c)
	if userRole != models.RoleAdmin {
		return c.JSON(http.StatusForbidden, map[string]interface{}{
			"error":   "Insufficient permissions. Required role: admin",
			"details": "Only administrators can access this resource",
			"code":    http.StatusForbidden,
		})
	}
	return h.s.GetAdminUsersGateways()
}

func (h *Handler) GetAdminUserGateway(c echo.Context) error {
	userRole := authMiddleware.GetUserRole(c)
	if userRole != models.RoleAdmin {
		return c.JSON(http.StatusForbidden, map[string]interface{}{
			"error":   "Insufficient permissions. Required role: admin",
			"details": "Only administrators can access this resource",
			"code":    http.StatusForbidden,
		})
	}
	return h.s.GetAdminUserGateway()
}

func (h *Handler) CreateAdminGateway(c echo.Context) error {
	userRole := authMiddleware.GetUserRole(c)
	if userRole != models.RoleAdmin {
		return c.JSON(http.StatusForbidden, map[string]interface{}{
			"error":   "Insufficient permissions. Required role: admin",
			"details": "Only administrators can access this resource",
			"code":    http.StatusForbidden,
		})
	}
	return h.s.CreateAdminGateway()
}

func (h *Handler) UpdateAdminGateway(c echo.Context) error {
	userRole := authMiddleware.GetUserRole(c)
	if userRole != models.RoleAdmin {
		return c.JSON(http.StatusForbidden, map[string]interface{}{
			"error":   "Insufficient permissions. Required role: admin",
			"details": "Only administrators can access this resource",
			"code":    http.StatusForbidden,
		})
	}
	return h.s.UpdateAdminGateway()
}

func (h *Handler) UpdateAdminUserGateway(c echo.Context) error {
	userRole := authMiddleware.GetUserRole(c)
	if userRole != models.RoleAdmin {
		return c.JSON(http.StatusForbidden, map[string]interface{}{
			"error":   "Insufficient permissions. Required role: admin",
			"details": "Only administrators can access this resource",
			"code":    http.StatusForbidden,
		})
	}
	return h.s.UpdateAdminUserGateway()
}

func (h *Handler) DeleteAdminGateway(c echo.Context) error {
	userRole := authMiddleware.GetUserRole(c)
	if userRole != models.RoleAdmin {
		return c.JSON(http.StatusForbidden, map[string]interface{}{
			"error":   "Insufficient permissions. Required role: admin",
			"details": "Only administrators can access this resource",
			"code":    http.StatusForbidden,
		})
	}
	return h.s.DeleteAdminGateway()
}

func (h *Handler) DeleteAdminUserGateway(c echo.Context) error {
	userRole := authMiddleware.GetUserRole(c)
	if userRole != models.RoleAdmin {
		return c.JSON(http.StatusForbidden, map[string]interface{}{
			"error":   "Insufficient permissions. Required role: admin",
			"details": "Only administrators can access this resource",
			"code":    http.StatusForbidden,
		})
	}
	return h.s.DeleteAdminUserGateway()
}
