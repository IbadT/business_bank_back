package generate

import "github.com/labstack/echo/v4"

func RegisterGenerateRoutes(e *echo.Group, h *Handler) {
	e.POST("/generate", h.Generate)
}
