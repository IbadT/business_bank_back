package service

type Services struct {
	UserService              UserService
	GeneratorService         GeneratorService
	HolidayService           HolidayService
	TransactionService       TransactionService
	GatewayService           GatewayService
	BaseAmountService        BaseAmountService
	BreakdownService         BreakdownService
	BalanceAdjustmentService BalanceAdjustmentService
	SeedService              SeedService
}

func NewServices(userService UserService, generatorService GeneratorService, holidayService HolidayService, transactionService TransactionService, gatewayService GatewayService, baseAmountService BaseAmountService, breakdownService BreakdownService, balanceAdjustmentService BalanceAdjustmentService, seedService SeedService) *Services {
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
