package transactions

import "github.com/labstack/echo/v4"

func RegisterTransactionsRoutes(e *echo.Group, h *Handler) {
	e.POST("", h.CreateTransaction)
	e.POST("/batch", h.CreateBatchTransactions)
	e.GET("/count/:request_id", h.GetTransactionsCount)
	e.GET("/type/:type/:request_id", h.GetTransactionsByTypeAndRequestID)
	e.GET("/method/:method/:request_id", h.GetTransactionsByMethodAndRequestID)
	e.GET("/:request_id", h.GetTransactionsByRequestID)
}
