package service

import (
	balanceservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/balance"
	baseamountservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/base"
	breakdownservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/breakdown"
	gatewayservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/gateway"
	generatorservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/generator"
	holidayservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/holiday"
	seedservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/seed"
	transactionservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/transaction"
	userservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/user"
)

type Services struct {
	UserService              userservice.UserService
	GeneratorService         generatorservice.GeneratorService
	HolidayService           holidayservice.HolidayService
	TransactionService       transactionservice.TransactionService
	GatewayService           gatewayservice.GatewayService
	BaseAmountService        baseamountservice.BaseAmountService
	BreakdownService         breakdownservice.BreakdownService
	BalanceAdjustmentService balanceservice.BalanceAdjustmentService
	SeedService              seedservice.SeedService
}

func NewServices(userService userservice.UserService, generatorService generatorservice.GeneratorService, holidayService holidayservice.HolidayService, transactionService transactionservice.TransactionService, gatewayService gatewayservice.GatewayService, baseAmountService baseamountservice.BaseAmountService, breakdownService breakdownservice.BreakdownService, balanceAdjustmentService balanceservice.BalanceAdjustmentService, seedService seedservice.SeedService) *Services {
	return &Services{
		UserService:              userService,
		GeneratorService:         generatorService,
		HolidayService:           holidayService,
		TransactionService:       transactionService,
		GatewayService:           gatewayService,
		BaseAmountService:        baseAmountService,
		BreakdownService:         breakdownService,
		BalanceAdjustmentService: balanceAdjustmentService,
		SeedService:              seedService,
	}
}
