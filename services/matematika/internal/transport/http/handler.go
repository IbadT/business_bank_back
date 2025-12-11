package http

import (
	"github.com/IbadT/business_bank_back/services/matematika/internal/service"
	authMiddleware "github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/middleware"
	v2 "github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/v2"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// Handler - основной HTTP handler для роутинга
type Handler struct {
	generatorService service.GeneratorService
	apiHandler       *v2.Handler
}

// NewHandler создает новый HTTP handler
func NewHandler(generatorService service.GeneratorService, userService service.UserService, holidayService service.HolidayService) *Handler {
	return &Handler{
		generatorService: generatorService,
		apiHandler:       v2.NewHandler(generatorService, userService, holidayService),
	}
}

// Init инициализирует Echo router со всеми middleware и роутами
func (h *Handler) Init() *echo.Echo {
	router := echo.New()

	// Middleware
	router.Use(middleware.Recover())
	router.Use(middleware.Logger())
	router.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE, echo.OPTIONS},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	// Swagger documentation - должен быть зарегистрирован ДО JWT middleware
	// Явно указываем имя инстанса "swagger" из docs.SwaggerInfo
	router.GET("/swagger/*", echoSwagger.EchoWrapHandler(echoSwagger.InstanceName("swagger")))

	// Health check endpoints - должны быть ДО JWT middleware
	router.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})
	router.GET("/api/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	// JWT Authentication Middleware (пропускает Swagger и публичные эндпоинты)
	router.Use(authMiddleware.JWTAuthMiddleware(authMiddleware.DefaultJWTConfig()))

	// API routes
	h.initAPI(router)

	return router
}

// initAPI регистрирует API роуты
func (h *Handler) initAPI(router *echo.Echo) {
	api := router.Group("/api")
	h.apiHandler.Init(api)
}
