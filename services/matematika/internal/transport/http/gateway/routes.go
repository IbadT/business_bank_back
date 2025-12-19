package gateway

import "github.com/labstack/echo/v4"

func RegisterGatewayRoutes(e *echo.Group, h *Handler) {
	e.GET("/b2c", h.GetB2CGateways)
	e.PUT("/b2c", h.UpdateB2CGateways)
	e.DELETE("/b2c", h.DeleteB2CGateways)

	// GATEWAY (шлюзы) - администраторские роуты
	// TODO: Реализовать администраторские роуты для управления шлюзами:
	// TODO: Все администраторские роуты должны проверять роль пользователя (только admin)
	
	// - GET /api/admin/gateways - получить список всех доступных шлюзов из gateways.csv
	e.GET("/admin", h.GetAdminGateways)
	// - GET /api/admin/gateways/users - получить список всех пользователей с их выбранными шлюзами
	e.GET("/admin/users", h.GetAdminUsersGateways)
	// - GET /api/admin/gateways/users/:user_id - получить выбранный шлюз конкретного пользователя
	e.GET("/admin/users/:user_id", h.GetAdminUserGateway)

	// - POST /api/admin/gateways - добавить новый шлюз (требуется обновление gateways.csv)
	e.POST("/admin", h.CreateAdminGateway)
	// - PUT /api/admin/gateways/:id - обновить шлюз (требуется обновление gateways.csv)
	e.PUT("/admin/:id", h.UpdateAdminGateway)
	// - PUT /api/admin/gateways/users/:user_id - установить шлюз для конкретного пользователя (принудительно)
	e.PUT("/admin/users/:user_id", h.UpdateAdminUserGateway)

	// - DELETE /api/admin/gateways/:id - удалить шлюз (требуется обновление gateways.csv)
	e.DELETE("/admin/:id", h.DeleteAdminGateway)
	// - DELETE /api/admin/gateways/users/:user_id - удалить выбранный шлюз для конкретного пользователя
	e.DELETE("/admin/users/:user_id", h.DeleteAdminUserGateway)
}
