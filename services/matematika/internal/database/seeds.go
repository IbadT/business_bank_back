package database

import (
	"context"
	"log"

	"github.com/IbadT/business_bank_back/services/matematika/internal/calculation"
	"gorm.io/gorm"
)

// SeedDatabase - наполняет БД моковыми данными
func SeedDatabase(db *gorm.DB) error {
	ctx := context.Background()

	log.Println("🌱 Starting database seeding...")

	// Создаем repository для работы с данными
	repo := calculation.NewCalculationRepository(db)

	// Очищаем существующие данные (опционально)
	if err := clearExistingData(db); err != nil {
		return err
	}

	// Seed компания 1: B2C - Srb Autos LLC
	if err := seedSrbAutos(ctx, repo); err != nil {
		return err
	}

	// Seed компания 2: B2B - TechCorp Industries
	if err := seedTechCorp(ctx, repo); err != nil {
		return err
	}

	// Seed компания 3: B2C - FastFood LLC
	if err := seedFastFood(ctx, repo); err != nil {
		return err
	}

	log.Println("✅ Database seeding completed successfully")
	return nil
}

// clearExistingData - очистка существующих mock данных
func clearExistingData(db *gorm.DB) error {
	// Удаляем только mock данные (не все!)
	if err := db.Where("account_id LIKE 'MOCK_%'").Delete(&calculation.Statement{}).Error; err != nil {
		return err
	}
	log.Println("🗑️  Cleared existing mock data")
	return nil
}

// ============================================================================
// КОМПАНИЯ 1: Srb Autos LLC (B2C - Automotive)
// ============================================================================

func seedSrbAutos(ctx context.Context, repo calculation.CalculationRepository) error {
	log.Println("📦 Seeding: Srb Autos LLC (B2C)")

	accountNumber := "201290125551"
	associatedCard := "2091222000102910"

	// JANUARY 2025
	januaryData := calculation.MatematikaResponse{
		"JANUARY 2025": calculation.MonthlyStatement{
			FinancialSummary: calculation.FinancialSummary{
				CompanyName:    "Srb Autos LLC.",
				AccountNumber:  accountNumber,
				Period:         "2025-01-01 - 2025-01-31",
				InitialBalance: 100000.00,
				FinalBalance:   163149.16,
				TotalRevenue:   100000.00,
				TotalExpenses:  -36850.84,
				NetProfit:      63149.16,
			},
			Transactions: []calculation.TransactionResponse{
				{
					TransactionID:   "t_man_001",
					TransactionDate: "2025-01-01T10:00:00",
					PostingDate:     "2025-01-01",
					Type:            calculation.TransactionTypeIncome,
					Category:        "Пополнение шлюз",
					Method:          calculation.TransactionMethodBankTransfer,
					Amount:          5000.00,
					IsManual:        true,
					BalanceAfter:    105000.00,
					FixAsFirst:      false,
				},
				{
					TransactionID:   "t_inc_001",
					TransactionDate: "2025-01-02T10:11:00",
					PostingDate:     "2025-01-02",
					Type:            calculation.TransactionTypeIncome,
					Category:        "retails_ca.csv",
					Method:          calculation.TransactionMethodACHCredit,
					Amount:          8500.00,
					IsManual:        false,
					BalanceAfter:    113500.00,
					FixAsFirst:      false,
				},
				{
					TransactionID:   "t_inc_002",
					TransactionDate: "2025-01-03T14:20:00",
					PostingDate:     "2025-01-03",
					Type:            calculation.TransactionTypeIncome,
					Category:        "wholesale_ca.csv",
					Method:          calculation.TransactionMethodACHCredit,
					Amount:          7400.00,
					IsManual:        false,
					BalanceAfter:    122400.00,
					FixAsFirst:      false,
				},
				{
					TransactionID:   "t_exp_001",
					TransactionDate: "2025-01-04T13:00:00",
					PostingDate:     "2025-01-04",
					Type:            calculation.TransactionTypeExpense,
					Category:        "Оплата платной дороги",
					Method:          "card",
					Amount:          -35.00,
					IsManual:        false,
					BalanceAfter:    122365.00,
					FixAsFirst:      false,
				},
				{
					TransactionID:   "t_exp_002",
					TransactionDate: "2025-01-05T14:00:00",
					PostingDate:     "2025-01-05",
					Type:            calculation.TransactionTypeExpense,
					Category:        "Маркетинг",
					Method:          calculation.TransactionMethodBankTransfer,
					Amount:          -1200.00,
					IsManual:        true,
					BalanceAfter:    121165.00,
					FixAsFirst:      false,
				},
				{
					TransactionID:   "t_exp_003",
					TransactionDate: "2025-01-06T09:30:00",
					PostingDate:     "2025-01-06",
					Type:            calculation.TransactionTypeExpense,
					Category:        "Топливо / Fleet",
					Method:          "card",
					Amount:          -315.88,
					IsManual:        false,
					BalanceAfter:    120849.12,
					FixAsFirst:      false,
				},
				{
					TransactionID:   "t_exp_004",
					TransactionDate: "2025-01-10T18:01:05",
					PostingDate:     "2025-01-10",
					Type:            calculation.TransactionTypeExpense,
					Category:        "Мобильная связь",
					Method:          "card",
					Amount:          -350.00,
					IsManual:        false,
					BalanceAfter:    140748.26,
					FixAsFirst:      true,
				},
				{
					TransactionID:   "t_exp_005",
					TransactionDate: "2025-01-11T00:01:00",
					PostingDate:     "2025-01-11",
					Type:            calculation.TransactionTypeExpense,
					Category:        "Подписка ПО",
					Method:          "card",
					Amount:          -150.00,
					IsManual:        false,
					BalanceAfter:    140598.26,
					FixAsFirst:      true,
					FixAsFirstCount: intPtr(1),
				},
				{
					TransactionID:   "t_exp_006",
					TransactionDate: "2025-01-24T17:00:00",
					PostingDate:     "2025-01-24",
					Type:            calculation.TransactionTypeExpense,
					Category:        "Payroll ADP",
					Method:          calculation.TransactionMethodBankTransfer,
					Amount:          -13750.00,
					IsManual:        false,
					BalanceAfter:    167399.16,
					FixAsFirst:      true,
				},
				{
					TransactionID:   "t_exp_007",
					TransactionDate: "2025-01-31T17:30:00",
					PostingDate:     "2025-01-31",
					Type:            calculation.TransactionTypeExpense,
					Category:        "Leasing",
					Method:          calculation.TransactionMethodBankTransfer,
					Amount:          -11800.00,
					IsManual:        false,
					BalanceAfter:    163149.16,
					FixAsFirst:      true,
				},
			},
			ForwardingInfo: calculation.ForwardingInfo{
				AssociatedCard: associatedCard,
				OwnerName:      "John Doe",
				CustomCustomers: []string{
					"Super LLC",
					"Lulu Inc.",
				},
				CustomContractors: []calculation.CustomContractor{
					{TransactionType: "Бухгалтер", Name: "Jakson Sam CPA"},
					{TransactionType: "Топливо", Name: "LumNuft Inc"},
				},
			},
			DailyClosingBalances: generateDailyBalances("2025-01", 100000.00, 163149.16, 31),
			Totals: &calculation.Totals{
				TotalRevenue:  100000.00,
				TotalExpenses: -36850.84,
				NetProfit:     63149.16,
			},
			RevenueBreakdown: &calculation.RevenueBreakdown{
				TotalAch:     72050.00,
				TotalWire:    6000.00,
				TotalZelle:   1500.00,
				TotalGateway: 5000.00,
				TotalOther:   15450.00,
			},
			ExpensesBreakdown: &calculation.ExpensesBreakdown{
				ByCard:    -820.88,
				ByAccount: -36029.96,
			},
			TransactionCounts: &calculation.TransactionCounts{
				Total: 34,
				Deposits: calculation.DepositCounts{
					Total: 15,
					Ach:   9,
					Wire:  1,
					Zelle: 1,
				},
				Withdrawals: calculation.WithdrawalCounts{
					Total:       19,
					FromAccount: 11,
					ByCard:      8,
				},
			},
		},
	}

	if err := saveStatement(ctx, repo, "stmt_2025-01_"+accountNumber, januaryData); err != nil {
		return err
	}

	// FEBRUARY 2025
	februaryData := calculation.MatematikaResponse{
		"FEBRUARY 2025": calculation.MonthlyStatement{
			FinancialSummary: calculation.FinancialSummary{
				CompanyName:    "Srb Autos LLC.",
				AccountNumber:  accountNumber,
				Period:         "2025-02-01 - 2025-02-28",
				InitialBalance: 163149.16,
				FinalBalance:   119569.16,
				TotalRevenue:   100000.00,
				TotalExpenses:  -58762.25,
				NetProfit:      41237.75,
			},
			Transactions: []calculation.TransactionResponse{
				{
					TransactionID:   "t_feb_inc_001",
					TransactionDate: "2025-02-03T10:25:00",
					PostingDate:     "2025-02-03",
					Type:            calculation.TransactionTypeIncome,
					Category:        "retails_ca.csv",
					Method:          calculation.TransactionMethodACHCredit,
					Amount:          8350.00,
					BalanceAfter:    171499.16,
				},
				{
					TransactionID:   "t_feb_exp_001",
					TransactionDate: "2025-02-11T00:01:00",
					PostingDate:     "2025-02-11",
					Type:            calculation.TransactionTypeExpense,
					Category:        "Подписка ПО",
					Method:          "card",
					Amount:          -150.00,
					BalanceAfter:    171349.16,
					FixAsFirst:      true,
				},
				{
					TransactionID:   "t_feb_exp_002",
					TransactionDate: "2025-02-14T17:00:00",
					PostingDate:     "2025-02-14",
					Type:            calculation.TransactionTypeExpense,
					Category:        "Payroll ADP",
					Method:          calculation.TransactionMethodBankTransfer,
					Amount:          -13600.00,
					BalanceAfter:    154819.16,
					FixAsFirst:      true,
				},
				{
					TransactionID:   "t_feb_exp_003",
					TransactionDate: "2025-02-28T17:00:00",
					PostingDate:     "2025-02-28",
					Type:            calculation.TransactionTypeExpense,
					Category:        "Payroll ADP",
					Method:          calculation.TransactionMethodBankTransfer,
					Amount:          -13600.00,
					BalanceAfter:    119569.16,
					FixAsFirst:      true,
				},
			},
			ForwardingInfo: calculation.ForwardingInfo{
				AssociatedCard:  associatedCard,
				OwnerName:       "John Doe",
				CustomCustomers: []string{"Super LLC", "Lulu Inc."},
			},
			DailyClosingBalances: generateDailyBalances("2025-02", 163149.16, 119569.16, 28),
		},
	}

	if err := saveStatement(ctx, repo, "stmt_2025-02_"+accountNumber, februaryData); err != nil {
		return err
	}

	log.Println("  ✓ Seeded 2 months for Srb Autos LLC")
	return nil
}

// ============================================================================
// КОМПАНИЯ 2: TechCorp Industries (B2B - Technology)
// ============================================================================

func seedTechCorp(ctx context.Context, repo calculation.CalculationRepository) error {
	log.Println("📦 Seeding: TechCorp Industries (B2B)")

	accountNumber := "301892345678"
	associatedCard := "4532123456789012"

	januaryData := calculation.MatematikaResponse{
		"JANUARY 2025": calculation.MonthlyStatement{
			FinancialSummary: calculation.FinancialSummary{
				CompanyName:    "TechCorp Industries Inc.",
				AccountNumber:  accountNumber,
				Period:         "2025-01-01 - 2025-01-31",
				InitialBalance: 250000.00,
				FinalBalance:   318750.50,
				TotalRevenue:   200000.00,
				TotalExpenses:  -131249.50,
				NetProfit:      68750.50,
			},
			Transactions: generateB2BTransactions("2025-01", 250000.00, 200000.00),
			ForwardingInfo: calculation.ForwardingInfo{
				AssociatedCard: associatedCard,
				OwnerName:      "Sarah Johnson",
				CustomCustomers: []string{
					"GlobalTech Solutions",
					"DataStream Corp",
					"CloudNine Systems",
				},
				CustomContractors: []calculation.CustomContractor{
					{TransactionType: "IT-dev", Name: "DevSquad LLC"},
					{TransactionType: "Бухгалтер", Name: "FinPro Accounting"},
				},
			},
			DailyClosingBalances: generateDailyBalances("2025-01", 250000.00, 318750.50, 31),
		},
	}

	if err := saveStatement(ctx, repo, "stmt_2025-01_"+accountNumber, januaryData); err != nil {
		return err
	}

	log.Println("  ✓ Seeded 1 month for TechCorp Industries")
	return nil
}

// ============================================================================
// КОМПАНИЯ 3: FastFood LLC (B2C - Restaurant)
// ============================================================================

func seedFastFood(ctx context.Context, repo calculation.CalculationRepository) error {
	log.Println("📦 Seeding: FastFood LLC (B2C)")

	accountNumber := "402156789012"
	associatedCard := "5412987654321098"

	januaryData := calculation.MatematikaResponse{
		"JANUARY 2025": calculation.MonthlyStatement{
			FinancialSummary: calculation.FinancialSummary{
				CompanyName:    "FastFood LLC",
				AccountNumber:  accountNumber,
				Period:         "2025-01-01 - 2025-01-31",
				InitialBalance: 50000.00,
				FinalBalance:   58250.75,
				TotalRevenue:   80000.00,
				TotalExpenses:  -71749.25,
				NetProfit:      8250.75,
			},
			Transactions: generateB2CRestaurantTransactions("2025-01", 50000.00, 80000.00),
			ForwardingInfo: calculation.ForwardingInfo{
				AssociatedCard:  associatedCard,
				OwnerName:       "Michael Brown",
				CustomCustomers: []string{},
			},
			DailyClosingBalances: generateDailyBalances("2025-01", 50000.00, 58250.75, 31),
		},
	}

	if err := saveStatement(ctx, repo, "stmt_2025-01_"+accountNumber, januaryData); err != nil {
		return err
	}

	// FEBRUARY 2025
	februaryData := calculation.MatematikaResponse{
		"FEBRUARY 2025": calculation.MonthlyStatement{
			FinancialSummary: calculation.FinancialSummary{
				CompanyName:    "FastFood LLC",
				AccountNumber:  accountNumber,
				Period:         "2025-02-01 - 2025-02-28",
				InitialBalance: 58250.75,
				FinalBalance:   65100.20,
				TotalRevenue:   75000.00,
				TotalExpenses:  -68150.55,
				NetProfit:      6849.45,
			},
			Transactions: generateB2CRestaurantTransactions("2025-02", 58250.75, 75000.00),
			ForwardingInfo: calculation.ForwardingInfo{
				AssociatedCard: associatedCard,
				OwnerName:      "Michael Brown",
			},
			DailyClosingBalances: generateDailyBalances("2025-02", 58250.75, 65100.20, 28),
		},
	}

	if err := saveStatement(ctx, repo, "stmt_2025-02_"+accountNumber, februaryData); err != nil {
		return err
	}

	// MARCH 2025
	marchData := calculation.MatematikaResponse{
		"MARCH 2025": calculation.MonthlyStatement{
			FinancialSummary: calculation.FinancialSummary{
				CompanyName:    "FastFood LLC",
				AccountNumber:  accountNumber,
				Period:         "2025-03-01 - 2025-03-31",
				InitialBalance: 65100.20,
				FinalBalance:   72550.80,
				TotalRevenue:   82000.00,
				TotalExpenses:  -74549.40,
				NetProfit:      7450.60,
			},
			Transactions: generateB2CRestaurantTransactions("2025-03", 65100.20, 82000.00),
			ForwardingInfo: calculation.ForwardingInfo{
				AssociatedCard: associatedCard,
				OwnerName:      "Michael Brown",
			},
			DailyClosingBalances: generateDailyBalances("2025-03", 65100.20, 72550.80, 31),
		},
	}

	if err := saveStatement(ctx, repo, "stmt_2025-03_"+accountNumber, marchData); err != nil {
		return err
	}

	log.Println("  ✓ Seeded 3 months for FastFood LLC")
	return nil
}

// ============================================================================
// ГЕНЕРАТОРЫ ТРАНЗАКЦИЙ
// ============================================================================

// generateB2BTransactions - генерация B2B транзакций
func generateB2BTransactions(month string, initialBalance, revenue float64) []calculation.TransactionResponse {
	transactions := []calculation.TransactionResponse{
		// Доходы - B2B клиенты (10-20 транзакций)
		{
			TransactionID:   "t_b2b_inc_001",
			TransactionDate: month + "-03T09:15:00",
			PostingDate:     month + "-03",
			Type:            calculation.TransactionTypeIncome,
			Category:        "retails_ca.csv",
			Method:          calculation.TransactionMethodACHCredit,
			Amount:          15000.00,
			BalanceAfter:    initialBalance + 15000.00,
		},
		{
			TransactionID:   "t_b2b_inc_002",
			TransactionDate: month + "-05T11:20:00",
			PostingDate:     month + "-05",
			Type:            calculation.TransactionTypeIncome,
			Category:        "wholesale_ca.csv",
			Method:          calculation.TransactionMethodACHCredit,
			Amount:          18500.00,
			BalanceAfter:    initialBalance + 33500.00,
		},
		{
			TransactionID:   "t_b2b_inc_003",
			TransactionDate: month + "-07T14:30:00",
			PostingDate:     month + "-07",
			Type:            calculation.TransactionTypeIncome,
			Category:        "agriculture_ca.csv",
			Method:          "Electronic Payment",
			Amount:          12000.00,
			BalanceAfter:    initialBalance + 45500.00,
		},
		{
			TransactionID:   "t_b2b_inc_004",
			TransactionDate: month + "-10T10:00:00",
			PostingDate:     month + "-10",
			Type:            calculation.TransactionTypeIncome,
			Category:        "factoring_avance_ca.csv",
			Method:          calculation.TransactionMethodACHCredit,
			Amount:          16000.00,
			BalanceAfter:    initialBalance + 61500.00,
		},

		// Расходы
		{
			TransactionID:   "t_b2b_exp_001",
			TransactionDate: month + "-08T12:00:00",
			PostingDate:     month + "-08",
			Type:            calculation.TransactionTypeExpense,
			Category:        "Топливо / Fleet",
			Method:          "card",
			Amount:          -2850.50,
			BalanceAfter:    initialBalance + 58649.50,
		},
		{
			TransactionID:   "t_b2b_exp_002",
			TransactionDate: month + "-10T17:00:00",
			PostingDate:     month + "-10",
			Type:            calculation.TransactionTypeExpense,
			Category:        "Payroll ADP",
			Method:          calculation.TransactionMethodBankTransfer,
			Amount:          -27000.00,
			BalanceAfter:    initialBalance + 31649.50,
		},
		{
			TransactionID:   "t_b2b_exp_003",
			TransactionDate: month + "-15T11:00:00",
			PostingDate:     month + "-15",
			Type:            calculation.TransactionTypeExpense,
			Category:        "IRS-налоги",
			Method:          calculation.TransactionMethodBankTransfer,
			Amount:          -3500.00,
			BalanceAfter:    initialBalance + 28149.50,
		},
	}

	return transactions
}

// generateB2CRestaurantTransactions - генерация транзакций для ресторана (B2C)
func generateB2CRestaurantTransactions(month string, initialBalance, revenue float64) []calculation.TransactionResponse {
	// Упрощенный набор транзакций для ресторана
	revenuePerWeek := revenue / 4.0

	transactions := []calculation.TransactionResponse{
		// Доходы - пятничные поступления от платежного шлюза
		{
			TransactionID:   "t_rest_inc_001",
			TransactionDate: month + "-03T16:00:00",
			PostingDate:     month + "-03",
			Type:            calculation.TransactionTypeIncome,
			Category:        "Пополнение шлюз",
			Method:          calculation.TransactionMethodBankTransfer,
			Amount:          revenuePerWeek,
			BalanceAfter:    initialBalance + revenuePerWeek,
		},
		{
			TransactionID:   "t_rest_inc_002",
			TransactionDate: month + "-10T16:00:00",
			PostingDate:     month + "-10",
			Type:            calculation.TransactionTypeIncome,
			Category:        "Пополнение шлюз",
			Method:          calculation.TransactionMethodBankTransfer,
			Amount:          revenuePerWeek,
			BalanceAfter:    initialBalance + revenuePerWeek*2,
		},
		{
			TransactionID:   "t_rest_inc_003",
			TransactionDate: month + "-17T16:00:00",
			PostingDate:     month + "-17",
			Type:            calculation.TransactionTypeIncome,
			Category:        "Пополнение шлюз",
			Method:          calculation.TransactionMethodBankTransfer,
			Amount:          revenuePerWeek,
			BalanceAfter:    initialBalance + revenuePerWeek*3,
		},
		{
			TransactionID:   "t_rest_inc_004",
			TransactionDate: month + "-24T16:00:00",
			PostingDate:     month + "-24",
			Type:            calculation.TransactionTypeIncome,
			Category:        "Пополнение шлюз",
			Method:          calculation.TransactionMethodBankTransfer,
			Amount:          revenuePerWeek,
			BalanceAfter:    initialBalance + revenue,
		},

		// Расходы - типичные для ресторана
		{
			TransactionID:   "t_rest_exp_001",
			TransactionDate: month + "-05T10:00:00",
			PostingDate:     month + "-05",
			Type:            calculation.TransactionTypeExpense,
			Category:        "Продукты и поставки",
			Method:          "card",
			Amount:          -28000.00,
			BalanceAfter:    initialBalance + revenue - 28000.00,
		},
		{
			TransactionID:   "t_rest_exp_002",
			TransactionDate: month + "-10T00:01:00",
			PostingDate:     month + "-10",
			Type:            calculation.TransactionTypeExpense,
			Category:        "Подписка ПО",
			Method:          "card",
			Amount:          -150.00,
			BalanceAfter:    initialBalance + revenue - 28150.00,
			FixAsFirst:      true,
		},
		{
			TransactionID:   "t_rest_exp_003",
			TransactionDate: month + "-14T17:00:00",
			PostingDate:     month + "-14",
			Type:            calculation.TransactionTypeExpense,
			Category:        "Payroll ADP",
			Method:          calculation.TransactionMethodBankTransfer,
			Amount:          -22000.00,
			BalanceAfter:    initialBalance + revenue - 50150.00,
			FixAsFirst:      true,
		},
		{
			TransactionID:   "t_rest_exp_004",
			TransactionDate: month + "-15T11:00:00",
			PostingDate:     month + "-15",
			Type:            calculation.TransactionTypeExpense,
			Category:        "Аренда помещения",
			Method:          calculation.TransactionMethodBankTransfer,
			Amount:          -8500.00,
			BalanceAfter:    initialBalance + revenue - 58650.00,
		},
		{
			TransactionID:   "t_rest_exp_005",
			TransactionDate: month + "-21T10:30:00",
			PostingDate:     month + "-21",
			Type:            calculation.TransactionTypeExpense,
			Category:        "Коммунальные",
			Method:          "card",
			Amount:          -1250.00,
			BalanceAfter:    initialBalance + revenue - 59900.00,
			FixAsFirst:      true,
		},
		{
			TransactionID:   "t_rest_exp_006",
			TransactionDate: month + "-28T17:00:00",
			PostingDate:     month + "-28",
			Type:            calculation.TransactionTypeExpense,
			Category:        "Payroll ADP",
			Method:          calculation.TransactionMethodBankTransfer,
			Amount:          -22000.00,
			BalanceAfter:    initialBalance + revenue - 81900.00,
			FixAsFirst:      true,
		},
	}

	return transactions
}

func generateRetailTransactions(month string, initialBalance, revenue float64) []calculation.TransactionResponse {
	return []calculation.TransactionResponse{
		{
			TransactionID:   "t_retail_inc_001",
			TransactionDate: month + "-03T15:00:00",
			PostingDate:     month + "-03",
			Type:            calculation.TransactionTypeIncome,
			Category:        "Пополнение шлюз",
			Method:          calculation.TransactionMethodBankTransfer,
			Amount:          30000.00,
			BalanceAfter:    initialBalance + 30000.00,
		},
		{
			TransactionID:   "t_retail_exp_001",
			TransactionDate: month + "-05T10:00:00",
			PostingDate:     month + "-05",
			Type:            calculation.TransactionTypeExpense,
			Category:        "Inventory Purchase",
			Method:          "card",
			Amount:          -45000.00,
			BalanceAfter:    initialBalance - 15000.00,
		},
		{
			TransactionID:   "t_retail_exp_002",
			TransactionDate: month + "-14T17:00:00",
			PostingDate:     month + "-14",
			Type:            calculation.TransactionTypeExpense,
			Category:        "Payroll ADP",
			Method:          calculation.TransactionMethodBankTransfer,
			Amount:          -33000.00,
			BalanceAfter:    initialBalance - 48000.00,
		},
	}
}

// ============================================================================
// ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ
// ============================================================================

// saveStatement - сохранение statement в БД
func saveStatement(ctx context.Context, repo calculation.CalculationRepository, id string, data calculation.MatematikaResponse) error {
	if err := repo.SaveStatement(ctx, id, data); err != nil {
		log.Printf("  ❌ Failed to save statement %s: %v", id, err)
		return err
	}
	return nil
}

// generateDailyBalances - генерация daily closing balances
func generateDailyBalances(month string, initialBalance, finalBalance float64, days int) []calculation.DailyClosingBalance {
	balances := make([]calculation.DailyClosingBalance, days)

	// Линейное распределение баланса по дням (упрощенно)
	step := (finalBalance - initialBalance) / float64(days)

	for i := 0; i < days; i++ {
		day := i + 1
		date := formatDate(month, day)
		balance := initialBalance + (step * float64(i+1))

		balances[i] = calculation.DailyClosingBalance{
			Date:    date,
			Balance: roundToCents(balance),
		}
	}

	return balances
}

// formatDate - форматирование даты "2025-01" + day → "2025-01-05"
func formatDate(month string, day int) string {
	return month + formatDay(day)
}

// formatDay - форматирование дня: 5 → "-05"
func formatDay(day int) string {
	if day < 10 {
		return "-0" + string(rune('0'+day))
	}
	return "-" + string(rune('0'+day/10)) + string(rune('0'+day%10))
}

// roundToCents - округление до центов
func roundToCents(amount float64) float64 {
	return float64(int(amount*100+0.5)) / 100
}

// intPtr - создание указателя на int
func intPtr(i int) *int {
	return &i
}

// ============================================================================
// РАСШИРЕННЫЕ SEED ДАННЫЕ
// ============================================================================

// SeedExtendedData - расширенные mock данные с большим количеством транзакций
func SeedExtendedData(db *gorm.DB) error {
	log.Println("🌱 Starting extended database seeding...")

	ctx := context.Background()
	repo := calculation.NewCalculationRepository(db)

	// Компания 4: Construction LLC (B2B - Construction)
	if err := seedConstruction(ctx, repo); err != nil {
		return err
	}

	// Компания 5: RetailStore Inc (B2C - Retail)
	if err := seedRetailStore(ctx, repo); err != nil {
		return err
	}

	log.Println("✅ Extended database seeding completed")
	return nil
}

func seedConstruction(ctx context.Context, repo calculation.CalculationRepository) error {
	log.Println("📦 Seeding: Construction LLC (B2B)")

	accountNumber := "503789456123"

	januaryData := calculation.MatematikaResponse{
		"JANUARY 2025": calculation.MonthlyStatement{
			FinancialSummary: calculation.FinancialSummary{
				CompanyName:    "Construction LLC",
				AccountNumber:  accountNumber,
				Period:         "2025-01-01 - 2025-01-31",
				InitialBalance: 150000.00,
				FinalBalance:   178900.00,
				TotalRevenue:   180000.00,
				TotalExpenses:  -151100.00,
				NetProfit:      28900.00,
			},
			Transactions: generateConstructionTransactions("2025-01", 150000.00, 180000.00),
			ForwardingInfo: calculation.ForwardingInfo{
				AssociatedCard: "6011123456789012",
				OwnerName:      "David Smith",
				CustomCustomers: []string{
					"BuildRight Corp",
					"HomeConstruct Inc",
				},
			},
			DailyClosingBalances: generateDailyBalances("2025-01", 150000.00, 178900.00, 31),
		},
	}

	return saveStatement(ctx, repo, "stmt_2025-01_"+accountNumber, januaryData)
}

func seedRetailStore(ctx context.Context, repo calculation.CalculationRepository) error {
	log.Println("📦 Seeding: RetailStore Inc (B2C)")

	accountNumber := "604567890123"

	januaryData := calculation.MatematikaResponse{
		"JANUARY 2025": calculation.MonthlyStatement{
			FinancialSummary: calculation.FinancialSummary{
				CompanyName:    "RetailStore Inc",
				AccountNumber:  accountNumber,
				Period:         "2025-01-01 - 2025-01-31",
				InitialBalance: 75000.00,
				FinalBalance:   83250.00,
				TotalRevenue:   120000.00,
				TotalExpenses:  -111750.00,
				NetProfit:      8250.00,
			},
			Transactions: generateRetailTransactions("2025-01", 75000.00, 120000.00),
			ForwardingInfo: calculation.ForwardingInfo{
				AssociatedCard: "3782123456789012",
				OwnerName:      "Lisa Anderson",
			},
			DailyClosingBalances: generateDailyBalances("2025-01", 75000.00, 83250.00, 31),
		},
	}

	return saveStatement(ctx, repo, "stmt_2025-01_"+accountNumber, januaryData)
}

func generateConstructionTransactions(month string, initialBalance, revenue float64) []calculation.TransactionResponse {
	return []calculation.TransactionResponse{
		{
			TransactionID:   "t_const_inc_001",
			TransactionDate: month + "-05T10:00:00",
			PostingDate:     month + "-05",
			Type:            calculation.TransactionTypeIncome,
			Category:        "Project Payment",
			Method:          calculation.TransactionMethodWireCredit,
			Amount:          45000.00,
			BalanceAfter:    initialBalance + 45000.00,
		},
		{
			TransactionID:   "t_const_exp_001",
			TransactionDate: month + "-08T09:00:00",
			PostingDate:     month + "-08",
			Type:            calculation.TransactionTypeExpense,
			Category:        "Materials",
			Method:          calculation.TransactionMethodBankTransfer,
			Amount:          -35000.00,
			BalanceAfter:    initialBalance + 10000.00,
		},
		{
			TransactionID:   "t_const_exp_002",
			TransactionDate: month + "-14T17:00:00",
			PostingDate:     month + "-14",
			Type:            calculation.TransactionTypeExpense,
			Category:        "Payroll ADP",
			Method:          calculation.TransactionMethodBankTransfer,
			Amount:          -48000.00,
			BalanceAfter:    initialBalance - 38000.00,
		},
	}
}
