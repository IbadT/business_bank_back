package http

import (
	"net/http"
	"net/http/pprof"

	authMiddleware "github.com/IbadT/business_bank_back/services/matematika/internal/middleware"
	"github.com/IbadT/business_bank_back/services/matematika/internal/service"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// Handler - основной HTTP handler для роутинга
type Handler struct {
	seedService service.SeedService
	services    *service.Services
}

// NewHandler создает новый HTTP handler
func NewHandler(seedService service.SeedService,
	generatorService service.GeneratorService,
	userService service.UserService,
	holidayService service.HolidayService,
	transactionService service.TransactionService,
	gatewayService service.GatewayService,
	breakdownService service.BreakdownService,
	baseAmountService service.BaseAmountService,
	balanceAdjustmentService service.BalanceAdjustmentService,
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
		seedService: seedService,
		services:    services,
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

	// Seed endpoint - должен быть ДО JWT middleware (создает пользователей)
	router.POST("/seed", h.Seed)

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

// Seed выполняет заполнение базы данных seed данными
// @Summary      Заполнить базу данных seed данными
// @Description  Заполняет базу данных тестовыми данными (пользователи, праздники, транзакции и т.д.)
// @Tags         seed
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "База данных успешно заполнена"
// @Failure      500  {object}  map[string]interface{}  "Ошибка при заполнении базы данных"
// @Router       /seed [post]
func (h *Handler) Seed(c echo.Context) error {
	if h.seedService == nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Seed service not available",
			"code":  http.StatusInternalServerError,
		})
	}

	if err := h.seedService.SeedDatabase(); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to seed database",
			"details": err.Error(),
			"code":    http.StatusInternalServerError,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Database seeded successfully",
		"code":    http.StatusOK,
	})
}
