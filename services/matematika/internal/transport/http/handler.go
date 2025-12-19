package http

import (
	"net/http"
	"net/http/pprof"

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

	// pprof endpoints - должны быть ДО JWT middleware
	pprofGroup := router.Group("/debug/pprof")
	pprofGroup.GET("", echo.WrapHandler(http.HandlerFunc(pprof.Index)))
	pprofGroup.GET("/", echo.WrapHandler(http.HandlerFunc(pprof.Index)))
	pprofGroup.GET("/cmdline", echo.WrapHandler(http.HandlerFunc(pprof.Cmdline)))
	pprofGroup.GET("/profile", echo.WrapHandler(http.HandlerFunc(pprof.Profile)))
	pprofGroup.POST("/profile", echo.WrapHandler(http.HandlerFunc(pprof.Profile)))
	pprofGroup.GET("/symbol", echo.WrapHandler(http.HandlerFunc(pprof.Symbol)))
	pprofGroup.POST("/symbol", echo.WrapHandler(http.HandlerFunc(pprof.Symbol)))
	pprofGroup.GET("/trace", echo.WrapHandler(http.HandlerFunc(pprof.Trace)))
	pprofGroup.POST("/trace", echo.WrapHandler(http.HandlerFunc(pprof.Trace)))
	pprofGroup.GET("/allocs", echo.WrapHandler(pprof.Handler("allocs")))
	pprofGroup.GET("/block", echo.WrapHandler(pprof.Handler("block")))
	pprofGroup.GET("/goroutine", echo.WrapHandler(pprof.Handler("goroutine")))
	pprofGroup.GET("/heap", echo.WrapHandler(pprof.Handler("heap")))
	pprofGroup.GET("/mutex", echo.WrapHandler(pprof.Handler("mutex")))
	pprofGroup.GET("/threadcreate", echo.WrapHandler(pprof.Handler("threadcreate")))

	// JWT Authentication Middleware (пропускает Swagger и публичные эндпоинты)
	router.Use(authMiddleware.JWTAuthMiddleware(authMiddleware.DefaultJWTConfig()))

	// API routes
	h.initAPI(router)

	return router
}

// initAPI регистрирует API роуты
func (h *Handler) initAPI(router *echo.Echo) {
	api := router.Group("/api")
	RegisterRoutes(api, h.services)
}
