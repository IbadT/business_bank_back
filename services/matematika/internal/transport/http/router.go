package http

import (
	"github.com/IbadT/business_bank_back/services/matematika/internal/service"
	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/balance"
	baseamount "github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/base_amount"
	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/breakdown"
	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/gateway"
	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/generate"
	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/holiday"
	transactions "github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/transaction"
	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/user"
	"github.com/labstack/echo/v4"
)

func RegisterRoutes(api *echo.Group, services *service.Services) {
	// Auth routes (без префикса /users)
	userHandler := user.NewHandler(services.UserService)
	user.RegisterAuthRoutes(api, userHandler)

	// User routes
	userGroup := api.Group("/users")
	user.RegisterUserRoutes(userGroup, userHandler)

	holidayHandler := holiday.NewHandler(services.HolidayService)
	holidayGroup := api.Group("/holidays")
	holiday.RegisterHolidayRoutes(holidayGroup, holidayHandler)

	generateHandler := generate.NewHandler(services.GeneratorService)
	generateGroup := api.Group("/generates")
	generate.RegisterGenerateRoutes(generateGroup, generateHandler)

	transactionHandler := transactions.NewHandler(services.TransactionService)
	transactionGroup := api.Group("/transactions")
	transactions.RegisterTransactionsRoutes(transactionGroup, transactionHandler)

	gatewayHandler := gateway.NewHandler(services.GatewayService)
	gatewayGroup := api.Group("/gateways")
	gateway.RegisterGatewayRoutes(gatewayGroup, gatewayHandler)

	breakdownHandler := breakdown.NewHandler(services.BreakdownService)
	breakdownGroup := api.Group("/breakdowns")
	breakdown.RegisterBreakdownRoutes(breakdownGroup, breakdownHandler)

	baseAmountHandler := baseamount.NewHandler(services.BaseAmountService)
	baseAmountGroup := api.Group("/base-amounts")
	baseamount.RegisterBaseAmountRoutes(baseAmountGroup, baseAmountHandler)

	balanceHandler := balance.NewHandler(services.BalanceAdjustmentService)
	balanceGroup := api.Group("/balances")
	balance.RegisterBalanceRoutes(balanceGroup, balanceHandler)
}
