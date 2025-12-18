package user

import (
	"github.com/labstack/echo/v4"
)

func RegisterUserRoutes(e *echo.Group, h *Handler) {
	e.PUT("/associated-card", h.SaveAssociatedCard)
}

func RegisterAuthRoutes(e *echo.Group, h *Handler) {
	e.POST("/login", h.Login)
	e.POST("/register", h.Register)
}
