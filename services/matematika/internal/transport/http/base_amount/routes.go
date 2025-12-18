package baseamount

import "github.com/labstack/echo/v4"

func RegisterBaseAmountRoutes(e *echo.Group, h *Handler) {
	e.GET("", h.GetBaseAmount)
	e.GET("/mobile/calculate", h.CalculateMobileAmount)
	e.GET("/utilities/calculate", h.CalculateUtilitiesAmount)
	e.GET("/leasing/calculate", h.CalculateLeasingAmount)
	e.DELETE("/mobile", h.ResetMobileBaseAmount)
	e.DELETE("/utilities", h.ResetUtilitiesBaseAmount)
	e.DELETE("/leasing", h.ResetLeasingBaseAmount)
}
