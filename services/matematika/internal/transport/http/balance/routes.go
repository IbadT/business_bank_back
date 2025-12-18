package balance

import "github.com/labstack/echo/v4"

func RegisterBalanceRoutes(e *echo.Group, h *Handler) {
	e.POST("/validate-balance", h.ValidateBalance)
	e.GET("/:request_id/balance-adjustment", h.GetBalanceAdjustment)
}
