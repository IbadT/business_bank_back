package http

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

type APIHandler interface {
	Init(api *echo.Group)
}

type Handler struct {
	// service
	apiHandlerv1 APIHandler
}

func NewHandler(apiHandlerv1 APIHandler) *Handler {
	return &Handler{
		apiHandlerv1: apiHandlerv1,
	}
}

func (h *Handler) Init() *echo.Echo {
	router := echo.New()

	router.Use(middleware.Recover())
	router.Use(middleware.Logger())
	router.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE, echo.OPTIONS},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	router.GET("/swagger/*", echoSwagger.WrapHandler)

	h.initAPI(router)

	return router
}

func (h *Handler) initAPI(router *echo.Echo) {
	api := router.Group("/api")
	h.apiHandlerv1.Init(api)
}

func (h *Handler) SetAPIHandler(handler APIHandler) {
	h.apiHandlerv1 = handler
}