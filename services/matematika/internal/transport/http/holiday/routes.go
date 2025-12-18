package holiday

import "github.com/labstack/echo/v4"

func RegisterHolidayRoutes(e *echo.Group, h *Handler) {
	e.POST("", h.AddHoliday)
	e.GET("", h.GetHolidays)
	e.GET("/is-holiday", h.IsHoliday)
	e.PUT("/:id", h.UpdateHoliday)
	e.DELETE("/:id", h.DeleteHoliday)
}
