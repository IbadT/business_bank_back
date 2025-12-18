package breakdown

import "github.com/labstack/echo/v4"

func RegisterBreakdownRoutes(e *echo.Group, h *Handler) {
	e.GET("/revenue/:request_id", h.CalculateRevenueBreakdown)
	e.GET("/expenses/:request_id", h.CalculateExpensesBreakdown)
}
