package seeds

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/internal/calculation"
	"github.com/IbadT/business_bank_back/services/matematika/internal/helpers"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SeedDatabase - наполняет БД моковыми данными
func SeedDatabase(db *gorm.DB) error {
	// Используем новую функцию для заполнения таблиц
	return SeedTables(db)
}

// ============================================================================
// ГЕНЕРАТОРЫ ТРАНЗАКЦИЙ
// ============================================================================

func generateRetailTransactions(month string, initialBalance, revenue float64) []calculation.TransactionResponse {
	return []calculation.TransactionResponse{
		{
			TransactionID:   "t_retail_inc_001",
			TransactionDate: month + "-03T15:00:00",
			PostingDate:     month + "-03",
			Type:            helpers.TransactionTypeIncome,
			Category:        "Пополнение шлюз",
			Method:          helpers.TransactionMethodBankTransfer,
			Amount:          30000.00,
			BalanceAfter:    initialBalance + 30000.00,
		},
		{
			TransactionID:   "t_retail_exp_001",
			TransactionDate: month + "-05T10:00:00",
			PostingDate:     month + "-05",
			Type:            helpers.TransactionTypeExpense,
			Category:        "Inventory Purchase",
			Method:          "card",
			Amount:          -45000.00,
			BalanceAfter:    initialBalance - 15000.00,
		},
		{
			TransactionID:   "t_retail_exp_002",
			TransactionDate: month + "-14T17:00:00",
			PostingDate:     month + "-14",
			Type:            helpers.TransactionTypeExpense,
			Category:        "Payroll ADP",
			Method:          helpers.TransactionMethodBankTransfer,
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
// SEED BUSINESS RULES
// ============================================================================

// seedBusinessRules - заполнение дефолтными бизнес-правилами
func seedBusinessRules(db *gorm.DB) error {
	log.Println("📦 Seeding: Business Rules")

	// Очищаем существующие правила (опционально)
	if err := db.Where("1 = 1").Delete(&calculation.BusinessRule{}).Error; err != nil {
		log.Printf("  ⚠️  Warning: Could not clear existing business rules: %v", err)
	}

	rules := []calculation.BusinessRule{
		{
			ID:           uuid.New(),
			BusinessType: "B2B",
			Description:  "Default profit margin rule for B2B businesses: 15-25%",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           uuid.New(),
			BusinessType: "B2C",
			Description:  "Default profit margin rule for B2C businesses: 10-20%",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           uuid.New(),
			BusinessType: "B2B",
			Description:  "Transaction frequency rule: B2B typically has 5-15 transactions per month",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           uuid.New(),
			BusinessType: "B2C",
			Description:  "Transaction frequency rule: B2C typically has 20-50 transactions per month",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}

	for _, rule := range rules {
		if err := db.Create(&rule).Error; err != nil {
			log.Printf("  ❌ Failed to create business rule: %v", err)
			return err
		}
	}

	log.Println("  ✓ Seeded 4 business rules")
	return nil
}

// SeedTables - заполняет таблицы Statement, Transaction, DailyBalance, BusinessRule базовыми данными
func SeedTables(db *gorm.DB) error {
	log.Println("🌱 Seeding database tables...")

	// Используем транзакцию для атомарности
	return db.Transaction(func(tx *gorm.DB) error {
		// 1. Seed Statements
		if err := seedStatements(tx); err != nil {
			return err
		}

		// 2. Seed Transactions
		if err := seedTransactions(tx); err != nil {
			return err
		}

		// 3. Seed DailyBalances
		if err := seedDailyBalances(tx); err != nil {
			return err
		}

		// 4. Seed BusinessRules
		if err := seedBusinessRules(tx); err != nil {
			return err
		}

		log.Println("✅ All tables seeded successfully")
		return nil
	})
}

// seedStatements - заполняет таблицу statements
func seedStatements(db *gorm.DB) error {
	log.Println("📦 Seeding: Statements")

	statements := []calculation.Statement{
		{
			ID:             uuid.New(),
			AccountID:      "201290125551",
			Month:          "2025-01",
			Status:         "completed",
			BusinessType:   "B2C",
			InitialBalance: 100000.00,
			FinalBalance:   163149.16,
			TotalRevenue:   100000.00,
			TotalExpenses:  -36850.84,
			NetProfit:      63149.16,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             uuid.New(),
			AccountID:      "201290125551",
			Month:          "2025-02",
			Status:         "completed",
			BusinessType:   "B2C",
			InitialBalance: 163149.16,
			FinalBalance:   119569.16,
			TotalRevenue:   100000.00,
			TotalExpenses:  -58762.25,
			NetProfit:      41237.75,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             uuid.New(),
			AccountID:      "301892345678",
			Month:          "2025-01",
			Status:         "completed",
			BusinessType:   "B2B",
			InitialBalance: 250000.00,
			FinalBalance:   318750.50,
			TotalRevenue:   200000.00,
			TotalExpenses:  -131249.50,
			NetProfit:      68750.50,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
	}

	for i, stmt := range statements {
		result := db.Create(&stmt)
		if result.Error != nil {
			log.Printf("  ❌ Failed to create statement %d: %v", i+1, result.Error)
			return result.Error
		}
		if result.RowsAffected == 0 {
			log.Printf("  ⚠️  Warning: Statement %d created but no rows affected", i+1)
		}
		log.Printf("  ✓ Created statement %d: ID=%s, Account=%s, Month=%s, RowsAffected=%d",
			i+1, stmt.ID, stmt.AccountID, stmt.Month, result.RowsAffected)
	}

	// Проверяем что данные действительно записались
	var count int64
	if err := db.Model(&calculation.Statement{}).Count(&count).Error; err != nil {
		log.Printf("  ⚠️  Warning: Could not verify statements count: %v", err)
	} else {
		log.Printf("  ✓ Verified: %d statements in database", count)
	}

	log.Printf("  ✓ Seeded %d statements", len(statements))
	return nil
}

// seedTransactions - заполняет таблицу transactions
func seedTransactions(db *gorm.DB) error {
	log.Println("📦 Seeding: Transactions")

	// Получаем statement IDs
	var statements []calculation.Statement
	if err := db.Find(&statements).Error; err != nil {
		return fmt.Errorf("failed to get statements: %w", err)
	}

	if len(statements) == 0 {
		return fmt.Errorf("no statements found, seed statements first")
	}

	transactions := []calculation.Transaction{
		{
			ID:              uuid.New(),
			StatementID:     statements[0].ID,
			TransactionDate: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC),
			TransactionType: "income",
			Category:        "Пополнение шлюз",
			Method:          helpers.TransactionMethodBankTransfer,
			Amount:          5000.00,
			BalanceAfter:    105000.00,
			IsManual:        true,
			UserNotes:       "Initial deposit",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              uuid.New(),
			StatementID:     statements[0].ID,
			TransactionDate: time.Date(2025, 1, 2, 10, 11, 0, 0, time.UTC),
			TransactionType: "income",
			Category:        "retails_ca.csv",
			Method:          helpers.TransactionMethodACHCredit,
			Amount:          8500.00,
			BalanceAfter:    113500.00,
			IsManual:        false,
			UserNotes:       "",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              uuid.New(),
			StatementID:     statements[0].ID,
			TransactionDate: time.Date(2025, 1, 4, 13, 0, 0, 0, time.UTC),
			TransactionType: "expense",
			Category:        "Оплата платной дороги",
			Method:          helpers.TransactionMethodBankTransfer,
			Amount:          -35.00,
			BalanceAfter:    122365.00,
			IsManual:        false,
			UserNotes:       "",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              uuid.New(),
			StatementID:     statements[1].ID,
			TransactionDate: time.Date(2025, 2, 3, 10, 25, 0, 0, time.UTC),
			TransactionType: "income",
			Category:        "retails_ca.csv",
			Method:          helpers.TransactionMethodACHCredit,
			Amount:          8350.00,
			BalanceAfter:    171499.16,
			IsManual:        false,
			UserNotes:       "",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              uuid.New(),
			StatementID:     statements[2].ID,
			TransactionDate: time.Date(2025, 1, 3, 9, 15, 0, 0, time.UTC),
			TransactionType: "income",
			Category:        "retails_ca.csv",
			Method:          helpers.TransactionMethodACHCredit,
			Amount:          15000.00,
			BalanceAfter:    265000.00,
			IsManual:        false,
			UserNotes:       "",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
	}

	for _, tx := range transactions {
		if err := db.Create(&tx).Error; err != nil {
			log.Printf("  ⚠️  Failed to create transaction: %v", err)
			continue
		}
	}

	log.Printf("  ✓ Seeded %d transactions", len(transactions))
	return nil
}

// seedDailyBalances - заполняет таблицу daily_balances
func seedDailyBalances(db *gorm.DB) error {
	log.Println("📦 Seeding: Daily Balances")

	// Получаем statement IDs
	var statements []calculation.Statement
	if err := db.Find(&statements).Error; err != nil {
		return fmt.Errorf("failed to get statements: %w", err)
	}

	if len(statements) == 0 {
		return fmt.Errorf("no statements found, seed statements first")
	}

	var dailyBalances []calculation.DailyBalance

	// Для каждого statement создаем daily balances за месяц
	for _, stmt := range statements {
		// Парсим месяц
		year := 2025
		month := 1
		if stmt.Month == "2025-02" {
			month = 2
		}

		// Создаем балансы для первых 5 дней месяца
		initialBalance := stmt.InitialBalance
		balanceStep := (stmt.FinalBalance - stmt.InitialBalance) / 5.0

		for day := 1; day <= 5; day++ {
			dailyBalances = append(dailyBalances, calculation.DailyBalance{
				ID:          uuid.New(),
				StatementID: stmt.ID,
				Date:        time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC),
				Balance:     initialBalance + (balanceStep * float64(day)),
			})
		}
	}

	for _, dbRecord := range dailyBalances {
		if err := db.Create(&dbRecord).Error; err != nil {
			log.Printf("  ⚠️  Failed to create daily balance: %v", err)
			continue
		}
	}

	log.Printf("  ✓ Seeded %d daily balances", len(dailyBalances))
	return nil
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
			Type:            helpers.TransactionTypeIncome,
			Category:        "Project Payment",
			Method:          helpers.TransactionMethodWireCredit,
			Amount:          45000.00,
			BalanceAfter:    initialBalance + 45000.00,
		},
		{
			TransactionID:   "t_const_exp_001",
			TransactionDate: month + "-08T09:00:00",
			PostingDate:     month + "-08",
			Type:            helpers.TransactionTypeExpense,
			Category:        "Materials",
			Method:          helpers.TransactionMethodBankTransfer,
			Amount:          -35000.00,
			BalanceAfter:    initialBalance + 10000.00,
		},
		{
			TransactionID:   "t_const_exp_002",
			TransactionDate: month + "-14T17:00:00",
			PostingDate:     month + "-14",
			Type:            helpers.TransactionTypeExpense,
			Category:        "Payroll ADP",
			Method:          helpers.TransactionMethodBankTransfer,
			Amount:          -48000.00,
			BalanceAfter:    initialBalance - 38000.00,
		},
	}
}
