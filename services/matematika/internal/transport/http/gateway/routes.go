package gateway

import "github.com/labstack/echo/v4"

func RegisterGatewayRoutes(e *echo.Group, h *Handler) {
	e.GET("/b2c", h.GetB2CGateways)
	e.PUT("/b2c", h.UpdateB2CGateways)
	e.DELETE("/b2c", h.DeleteB2CGateways)
}
