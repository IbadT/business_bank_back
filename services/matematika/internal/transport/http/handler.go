package http

import (
	"net/http"

	authMiddleware "github.com/IbadT/business_bank_back/services/matematika/internal/middleware"
	"github.com/IbadT/business_bank_back/services/matematika/internal/service"
	balanceservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/balance"
	baseamountservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/base"
	breakdownservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/breakdown"
	gatewayservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/gateway"
	generatorservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/generator"
	holidayservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/holiday"
	seedservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/seed"
	transactionservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/transaction"
	userservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/user"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// Handler - основной HTTP handler для роутинга
type Handler struct {
	services    *service.Services
}

// NewHandler создает новый HTTP handler
func NewHandler(
	generatorService generatorservice.GeneratorService,
	userService userservice.UserService,
	holidayService holidayservice.HolidayService,
	transactionService transactionservice.TransactionService,
	gatewayService gatewayservice.GatewayService,
	breakdownService breakdownservice.BreakdownService,
	baseAmountService baseamountservice.BaseAmountService,
	balanceAdjustmentService balanceservice.BalanceAdjustmentService,
	seedService seedservice.SeedService,
) *Handler {
	services := service.NewServices(
		userService,
		generatorService,
		holidayService,
		transactionService,
		gatewayService,
		baseAmountService,
		breakdownService,
		balanceAdjustmentService,
		seedService,
	)
	return &Handler{
		services: services,
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

	// pprof endpoints - требуют авторизации и роль admin
	// pprof уже подключен в main.go через _ "net/http/pprof" и работает через http.DefaultServeMux
	// Проксируем запросы к DefaultServeMux с проверкой admin роли
	pprofGroup := router.Group("/debug/pprof")
	pprofGroup.Use(authMiddleware.RequireAdmin())
	pprofGroup.Any("/*", func(c echo.Context) error {
		// Проксируем запрос к стандартному pprof handler
		http.DefaultServeMux.ServeHTTP(c.Response(), c.Request())
		return nil
	})

	// API routes
	h.initAPI(router)

	return router
}

// initAPI регистрирует API роуты
func (h *Handler) initAPI(router *echo.Echo) {
	api := router.Group("/api")
	RegisterRoutes(api, h.services)
}
