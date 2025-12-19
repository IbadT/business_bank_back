package seed

import "github.com/labstack/echo/v4"

func RegisterSeedRoutes(e *echo.Group, h *Handler) {
	e.POST("", h.Seed)
}