package http

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/internal/database"
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
	"github.com/IbadT/business_bank_back/services/matematika/pkg/helpers"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	redisclient "github.com/redis/go-redis/v9"
	echoSwagger "github.com/swaggo/echo-swagger"
	"gorm.io/gorm"
)

// Handler - основной HTTP handler для роутинга
type Handler struct {
	services    *service.Services
	db          *gorm.DB
	redisClient *redisclient.Client
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
	db *gorm.DB,
	redisClient *redisclient.Client,
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
		services:    services,
		db:          db,
		redisClient: redisClient,
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
	router.GET("/health", h.HealthCheck)
	router.GET("/api/health", h.HealthCheck)

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

// HealthCheck проверяет статус сервиса и всех зависимостей
// @Summary      Health check
// @Description  Проверяет доступность сервиса и всех зависимостей (база данных, Redis)
// @Tags         health
// @Accept       json
// @Produce      json
// @Success      200  {object}  helpers.HealthCheckResponse  "Успешная проверка"
// @Failure      503  {object}  helpers.HealthCheckResponse  "Сервис недоступен"
// @Router       /health [get]
func (h *Handler) HealthCheck(c echo.Context) error {
	status := "ok"
	httpStatus := http.StatusOK

	// Проверка базы данных
	dbStatus := database.HealthCheckDB(h.db)
	if dbStatus == "disconnected" {
		status = "degraded"
		httpStatus = http.StatusServiceUnavailable
	}

	// Проверка Redis
	redisStatus := "disconnected"
	if h.redisClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := h.redisClient.Ping(ctx).Err(); err == nil {
			redisStatus = "connected"
		}
	} else {
		// Redis не инициализирован - это не критично, но отмечаем как disconnected
		redisStatus = "not_initialized"
	}

	// Если Redis недоступен, но это не критично для работы сервиса (опциональная зависимость)
	// Проверяем config_loaded
	configLoaded := true
	if os.Getenv("CONFIG_LOADED") == "" {
		configLoaded = false
	}

	response := helpers.HealthCheckResponse{
		Status:    status,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   "1.0",
		Dependencies: helpers.HealthCheckDependencies{
			Database:     dbStatus,
			Redis:        redisStatus,
			// Kafka:        "not_configured", // Kafka не используется в matematika service
			ConfigLoaded: configLoaded,
			Service:      "matematika",
		},
	}

	return c.JSON(httpStatus, response)
}
