package seeds

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/internal/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedDatabase - наполняет БД моковыми данными
func SeedDatabase(db *gorm.DB) error {
	// Заполняем таблицы
	if err := SeedV2Tables(db); err != nil {
		return err
	}
	
	return nil
}

// ============================================================================
// SEED TABLES
// ============================================================================

// SeedV2Tables - заполняет таблицы моковыми данными
func SeedV2Tables(db *gorm.DB) error {
	log.Println("🌱 Seeding database tables...")

	return db.Transaction(func(tx *gorm.DB) error {
		// 1. Seed Users (моковые пользователи)
		if err := seedUsers(tx); err != nil {
			return err
		}

		// 2. Seed Holidays (из config/holidays.json)
		if err := seedHolidays(tx); err != nil {
			return err
		}

		// 3. Seed Transaction Templates (из config/templates.json)
		if err := seedTransactionTemplates(tx); err != nil {
			return err
		}

		// 4. Seed Default Customers (из config/customers.json)
		if err := seedDefaultCustomers(tx); err != nil {
			return err
		}

		// 5. Seed Generation Requests
		if err := seedGenerationRequests(tx); err != nil {
			return err
		}

		// 6. Seed Generated Transactions
		if err := seedGeneratedTransactions(tx); err != nil {
			return err
		}

		// 7. Seed Financial Summaries
		if err := seedFinancialSummaries(tx); err != nil {
			return err
		}

		// 8. Seed Daily Balances (v2)
		if err := seedDailyBalancesV2(tx); err != nil {
			return err
		}

		log.Println("✅ All tables seeded successfully")
		return nil
	})
}

// seedUsers - заполняет таблицу users моковыми данными
func seedUsers(db *gorm.DB) error {
	log.Println("📦 Seeding: Users")

	// Очищаем существующих пользователей
	if err := db.Where("1 = 1").Delete(&models.User{}).Error; err != nil {
		log.Printf("  ⚠️  Warning: Could not clear existing users: %v", err)
	}

	// Хешируем пароли (все пароли: "password123")
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	passwordHash := string(hashedPassword)

	// Создаем моковых пользователей
	users := []models.User{
		{
			ID:           uuid.MustParse("00000000-0000-0000-0000-000000000001"),
			Email:        "admin@example.com",
			PasswordHash: passwordHash,
			Role:         models.RoleAdmin,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           uuid.MustParse("00000000-0000-0000-0000-000000000002"),
			Email:        "user1@example.com",
			PasswordHash: passwordHash,
			Role:         models.RoleUser,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           uuid.MustParse("00000000-0000-0000-0000-000000000003"),
			Email:        "user2@example.com",
			PasswordHash: passwordHash,
			Role:         models.RoleUser,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           uuid.MustParse("00000000-0000-0000-0000-000000000004"),
			Email:        "test@example.com",
			PasswordHash: passwordHash,
			Role:         models.RoleUser,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}

	if err := db.Create(&users).Error; err != nil {
		log.Printf("  ⚠️  Failed to create users: %v", err)
		return err
	}

	log.Printf("  ✓ Seeded %d users", len(users))
	log.Printf("    - Admin: admin@example.com (password: password123)")
	log.Printf("    - Users: user1@example.com, user2@example.com, test@example.com (password: password123)")
	return nil
}

// seedHolidays - заполняет таблицу holidays из config/holidays.json
func seedHolidays(db *gorm.DB) error {
	log.Println("📦 Seeding: Holidays")

	// Очищаем существующие праздники
	if err := db.Where("1 = 1").Delete(&models.HolidayDB{}).Error; err != nil {
		log.Printf("  ⚠️  Warning: Could not clear existing holidays: %v", err)
	}

	// Загружаем из config/holidays.json
	configPath := "./config"
	filePath := filepath.Join(configPath, "holidays.json")
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("  ⚠️  Warning: Could not open holidays.json: %v", err)
		// Создаем дефолтные праздники
		return seedDefaultHolidays(db)
	}
	defer file.Close()

	var holidayModels []models.Holiday
	if err := json.NewDecoder(file).Decode(&holidayModels); err != nil {
		log.Printf("  ⚠️  Warning: Could not decode holidays.json: %v", err)
		return seedDefaultHolidays(db)
	}

	// Конвертируем в ORM модели
	for _, hm := range holidayModels {
		date, err := time.Parse("2006-01-02", hm.Date)
		if err != nil {
			log.Printf("  ⚠️  Warning: Could not parse holiday date %s: %v", hm.Date, err)
			continue
		}

		country := hm.Country
		if country == "" {
			country = "RU" // Дефолтное значение
		}

		holiday := models.HolidayDB{
			ID:          uuid.New(),
			HolidayDate: date,
			Name:        hm.Name,
			Country:     country,
		}

		if err := db.Create(&holiday).Error; err != nil {
			log.Printf("  ⚠️  Failed to create holiday %s: %v", hm.Name, err)
			continue
		}
	}

	log.Printf("  ✓ Seeded %d holidays", len(holidayModels))
	return nil
}

// seedDefaultHolidays - создает дефолтные праздники если файл не найден
func seedDefaultHolidays(db *gorm.DB) error {
	// Дефолтные праздники для 2025 года
	defaultHolidays := []struct {
		date    string
		name    string
		country string
	}{
		{"2025-01-01", "New Year's Day", "US"},
		{"2025-01-07", "Orthodox Christmas", "RU"},
		{"2025-02-23", "Defender of the Fatherland Day", "RU"},
		{"2025-03-08", "International Women's Day", "RU"},
		{"2025-05-01", "Labor Day", "RU"},
		{"2025-05-09", "Victory Day", "RU"},
		{"2025-06-12", "Russia Day", "RU"},
		{"2025-11-04", "Unity Day", "RU"},
		{"2025-12-25", "Christmas", "US"},
	}

	for _, dh := range defaultHolidays {
		date, err := time.Parse("2006-01-02", dh.date)
		if err != nil {
			continue
		}

		holiday := models.HolidayDB{
			ID:          uuid.New(),
			HolidayDate: date,
			Name:        dh.name,
			Country:     dh.country,
		}

		if err := db.Create(&holiday).Error; err != nil {
			log.Printf("  ⚠️  Failed to create default holiday %s: %v", dh.name, err)
			continue
		}
	}

	log.Println("  ✓ Seeded default holidays")
	return nil
}

// seedTransactionTemplates - заполняет таблицу transaction_templates из config/templates.json
func seedTransactionTemplates(db *gorm.DB) error {
	log.Println("📦 Seeding: Transaction Templates")

	// Очищаем существующие шаблоны
	if err := db.Where("1 = 1").Delete(&models.TransactionTemplateDB{}).Error; err != nil {
		log.Printf("  ⚠️  Warning: Could not clear existing templates: %v", err)
	}

	// Загружаем из config/templates.json
	configPath := "./config"
	filePath := filepath.Join(configPath, "templates.json")
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("  ⚠️  Warning: Could not open templates.json: %v", err)
		// Создаем дефолтные шаблоны
		return seedDefaultTemplates(db)
	}
	defer file.Close()

	var templateModels []models.TransactionTemplate
	if err := json.NewDecoder(file).Decode(&templateModels); err != nil {
		log.Printf("  ⚠️  Warning: Could not decode templates.json: %v", err)
		return seedDefaultTemplates(db)
	}

	// Конвертируем в ORM модели
	for i, tm := range templateModels {
		templateKey := fmt.Sprintf("tm_%s_%d", tm.Category, i+1)
		
		// Конвертируем BusinessHours в JSONB
		businessHoursJSON := models.JSONB{
			"start": tm.BusinessHours.Start,
			"end":   tm.BusinessHours.End,
		}

		// Конвертируем WeekOfMonth в IntArray
		weekOfMonth := models.IntArray(tm.WeekOfMonth)
		if weekOfMonth == nil {
			weekOfMonth = models.IntArray{}
		}

		template := models.TransactionTemplateDB{
			ID:              uuid.New(),
			TemplateKey:     templateKey,
			Category:        tm.Category,
			Type:            tm.Type,
			IsPercentage:    tm.IsPercentage,
			Frequency:       stringPtr(tm.Frequency),
			PreferredDay:    stringPtr(tm.PreferredDay),
			WeekOfMonth:     weekOfMonth,
			BusinessHours:   businessHoursJSON,
			IsOptional:      tm.IsOptional,
			Method:          stringPtr(tm.Method),
			MinTransactions: tm.MinTransactions,
			MaxTransactions: tm.MaxTransactions,
		}

		if tm.IsPercentage {
			template.PercentageMin = &tm.PercentageMin
			template.PercentageMax = &tm.PercentageMax
		} else {
			template.FixedAmount = &tm.FixedAmount
		}

		if err := db.Create(&template).Error; err != nil {
			log.Printf("  ⚠️  Failed to create template %s: %v", templateKey, err)
			continue
		}
	}

	log.Printf("  ✓ Seeded %d transaction templates", len(templateModels))
	return nil
}

// seedDefaultTemplates - создает дефолтные шаблоны если файл не найден
func seedDefaultTemplates(db *gorm.DB) error {
	templates := []models.TransactionTemplateDB{
		{
			ID:              uuid.New(),
			TemplateKey:     "tm_payroll_1",
			Category:        "Payroll",
			Type:            "expense",
			IsPercentage:    true,
			PercentageMin:   floatPtr(0.27),
			PercentageMax:   floatPtr(0.275),
			Frequency:       stringPtr("biweekly"),
			PreferredDay:    stringPtr("Friday"),
			WeekOfMonth:     models.IntArray{2, 4},
			BusinessHours:   models.JSONB{"start": "08:00", "end": "18:00"},
			IsOptional:      false,
			Method:          stringPtr("ACH_DEBIT"),
			MinTransactions: 2,
			MaxTransactions: 2,
		},
	}

	for _, tm := range templates {
		if err := db.Create(&tm).Error; err != nil {
			return err
		}
	}

	log.Println("  ✓ Seeded default transaction templates")
	return nil
}

// seedDefaultCustomers - заполняет таблицу default_customers из config/customers.json
func seedDefaultCustomers(db *gorm.DB) error {
	log.Println("📦 Seeding: Default Customers")

	// Очищаем существующих клиентов
	if err := db.Where("1 = 1").Delete(&models.DefaultCustomerDB{}).Error; err != nil {
		log.Printf("  ⚠️  Warning: Could not clear existing customers: %v", err)
	}

	// Загружаем из config/customers.json
	configPath := "./config"
	filePath := filepath.Join(configPath, "customers.json")
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("  ⚠️  Warning: Could not open customers.json: %v", err)
		return seedDefaultCustomersData(db)
	}
	defer file.Close()

	var customerModels []models.DefaultCustomer
	if err := json.NewDecoder(file).Decode(&customerModels); err != nil {
		log.Printf("  ⚠️  Warning: Could not decode customers.json: %v", err)
		return seedDefaultCustomersData(db)
	}

	// Конвертируем в ORM модели
	for _, cm := range customerModels {
		customer := models.DefaultCustomerDB{
			ID:              uuid.New(),
			Name:            cm.Name,
			Category:        cm.Category,
			MinPercent:      cm.MinPercent,
			MaxPercent:      cm.MaxPercent,
			MinTransactions: cm.MinTransactions,
			MaxTransactions: cm.MaxTransactions,
		}

		if err := db.Create(&customer).Error; err != nil {
			log.Printf("  ⚠️  Failed to create customer %s: %v", cm.Name, err)
			continue
		}
	}

	log.Printf("  ✓ Seeded %d default customers", len(customerModels))
	return nil
}

// seedDefaultCustomersData - создает дефолтных клиентов если файл не найден
func seedDefaultCustomersData(db *gorm.DB) error {
	customers := []models.DefaultCustomerDB{
		{
			ID:              uuid.New(),
			Name:            "GlobalTech Solutions",
			Category:        "IT Services",
			MinPercent:      0.15,
			MaxPercent:      0.25,
			MinTransactions: 2,
			MaxTransactions: 5,
		},
		{
			ID:              uuid.New(),
			Name:            "DataStream Corp",
			Category:        "Data Analytics",
			MinPercent:      0.1,
			MaxPercent:      0.2,
			MinTransactions: 1,
			MaxTransactions: 3,
		},
	}

	for _, c := range customers {
		if err := db.Create(&c).Error; err != nil {
			return err
		}
	}

	log.Println("  ✓ Seeded default customers")
	return nil
}

// seedGenerationRequests - заполняет таблицу generation_requests
func seedGenerationRequests(db *gorm.DB) error {
	log.Println("📦 Seeding: Generation Requests")

	// Очищаем зависимые таблицы ПЕРЕД удалением generation_requests (из-за foreign key constraints)
	if err := db.Where("1 = 1").Delete(&models.FinancialSummaryDB{}).Error; err != nil {
		log.Printf("  ⚠️  Warning: Could not clear existing financial summaries: %v", err)
	}
	if err := db.Where("1 = 1").Delete(&models.DailyBalanceV2{}).Error; err != nil {
		log.Printf("  ⚠️  Warning: Could not clear existing daily balances: %v", err)
	}
	if err := db.Where("1 = 1").Delete(&models.GeneratedTransaction{}).Error; err != nil {
		log.Printf("  ⚠️  Warning: Could not clear existing transactions: %v", err)
	}

	// Теперь очищаем существующие запросы
	if err := db.Where("1 = 1").Delete(&models.GenerationRequest{}).Error; err != nil {
		log.Printf("  ⚠️  Warning: Could not clear existing requests: %v", err)
	}

	now := time.Now()
	completedAt := now.Add(5 * time.Minute)

	requests := []models.GenerationRequest{
		{
			ID:                 uuid.New(),
			UserID:             nil,
			Month:              "2025-01",
			Year:               2025,
			Turnover:           100000.00,
			DesiredProfitPercent: 15.0,
			Model:              "B2C",
			InitialBalance:     50000.00,
			ScaleFactor:        1,
			CustomData:         models.JSONB{},
			Status:             "completed",
			ErrorMessage:       nil,
			CreatedAt:          now.Add(-24 * time.Hour),
			CompletedAt:        &completedAt,
			UpdatedAt:          completedAt,
		},
		{
			ID:                 uuid.New(),
			UserID:             nil,
			Month:              "2025-01",
			Year:               2025,
			Turnover:           200000.00,
			DesiredProfitPercent: 20.0,
			Model:              "B2B",
			InitialBalance:     100000.00,
			ScaleFactor:        1,
			CustomData:         models.JSONB{"customCustomers": []string{"GlobalTech Solutions", "DataStream Corp"}},
			Status:             "completed",
			ErrorMessage:       nil,
			CreatedAt:          now.Add(-12 * time.Hour),
			CompletedAt:        &completedAt,
			UpdatedAt:          completedAt,
		},
		{
			ID:                 uuid.New(),
			UserID:             nil,
			Month:              "2025-02",
			Year:               2025,
			Turnover:           150000.00,
			DesiredProfitPercent: 18.0,
			Model:              "B2C",
			InitialBalance:     75000.00,
			ScaleFactor:        2,
			CustomData:         models.JSONB{},
			Status:             "processing",
			ErrorMessage:       nil,
			CreatedAt:          now.Add(-1 * time.Hour),
			CompletedAt:        nil,
			UpdatedAt:          now.Add(-1 * time.Hour),
		},
	}

	for _, req := range requests {
		if err := db.Create(&req).Error; err != nil {
			log.Printf("  ⚠️  Failed to create generation request: %v", err)
			continue
		}
	}

	log.Printf("  ✓ Seeded %d generation requests", len(requests))
	return nil
}

// seedGeneratedTransactions - заполняет таблицу generated_transactions
func seedGeneratedTransactions(db *gorm.DB) error {
	log.Println("📦 Seeding: Generated Transactions")

	// Очищаем существующие транзакции (если есть)
	if err := db.Where("1 = 1").Delete(&models.GeneratedTransaction{}).Error; err != nil {
		log.Printf("  ⚠️  Warning: Could not clear existing transactions: %v", err)
	}

	// Получаем generation requests
	var requests []models.GenerationRequest
	if err := db.Find(&requests).Error; err != nil {
		return fmt.Errorf("failed to get generation requests: %w", err)
	}

	if len(requests) == 0 {
		log.Println("  ⚠️  No generation requests found, skipping transactions")
		return nil
	}

	var transactions []models.GeneratedTransaction
	balanceAfter := 50000.00

	// Для каждого запроса создаем транзакции
	for i, req := range requests {
		if req.Status != "completed" {
			continue
		}

		// Доходы
		for j := 0; j < 4; j++ {
			transactionDate := time.Date(req.Year, time.Month(1), 3+j*7, 10+j, 0, 0, 0, time.UTC)
			amount := req.Turnover * 0.25
			balanceAfter += amount

			transactions = append(transactions, models.GeneratedTransaction{
				ID:                 uuid.New(),
				RequestID:          req.ID,
				TransactionID:      fmt.Sprintf("inc_%d_%d", i+1, j+1),
				TransactionDate:    transactionDate,
				PostingDate:        transactionDate,
				Type:               "income",
				Category:           "Stripe",
				Method:             "ACH_CREDIT",
				Amount:             amount,
				BalanceAfter:       &balanceAfter,
				IsManual:           false,
				CalculationDetails: models.JSONB{"source": "B2C", "gateway": "Stripe"},
				SortOrder:          intPtr(j + 1),
			})
		}

		// Расходы
		expenseCategories := []string{"Payroll", "Топливо", "Подписка ПО"}
		for j, category := range expenseCategories {
			transactionDate := time.Date(req.Year, time.Month(1), 5+j*10, 14, 0, 0, 0, time.UTC)
			amount := -req.Turnover * 0.1
			balanceAfter += amount

			transactions = append(transactions, models.GeneratedTransaction{
				ID:                 uuid.New(),
				RequestID:          req.ID,
				TransactionID:      fmt.Sprintf("exp_%d_%d", i+1, j+1),
				TransactionDate:    transactionDate,
				PostingDate:        transactionDate,
				Type:               "expense",
				Category:           category,
				Method:             "card",
				Amount:             amount,
				BalanceAfter:       &balanceAfter,
				IsManual:           false,
				CalculationDetails: models.JSONB{"template": category},
				SortOrder:          intPtr(10 + j + 1),
			})
		}
	}

	for _, tx := range transactions {
		if err := db.Create(&tx).Error; err != nil {
			log.Printf("  ⚠️  Failed to create transaction: %v", err)
			continue
		}
	}

	log.Printf("  ✓ Seeded %d generated transactions", len(transactions))
	return nil
}

// seedFinancialSummaries - заполняет таблицу financial_summaries
func seedFinancialSummaries(db *gorm.DB) error {
	log.Println("📦 Seeding: Financial Summaries")

	// Очищаем существующие финансовые сводки (если есть)
	if err := db.Where("1 = 1").Delete(&models.FinancialSummaryDB{}).Error; err != nil {
		log.Printf("  ⚠️  Warning: Could not clear existing financial summaries: %v", err)
	}

	// Получаем completed generation requests
	var requests []models.GenerationRequest
	if err := db.Where("status = ?", "completed").Find(&requests).Error; err != nil {
		return fmt.Errorf("failed to get generation requests: %w", err)
	}

	if len(requests) == 0 {
		log.Println("  ⚠️  No completed generation requests found, skipping financial summaries")
		return nil
	}

	var summaries []models.FinancialSummaryDB

	for _, req := range requests {
		// Получаем транзакции для расчета
		var transactions []models.GeneratedTransaction
		if err := db.Where("request_id = ?", req.ID).Find(&transactions).Error; err != nil {
			continue
		}

		var totalRevenue, totalExpenses float64
		var finalBalance float64

		for _, tx := range transactions {
			if tx.Type == "income" {
				totalRevenue += tx.Amount
			} else {
				totalExpenses += tx.Amount
			}
			if tx.BalanceAfter != nil {
				finalBalance = *tx.BalanceAfter
			}
		}

		netProfit := totalRevenue + totalExpenses // totalExpenses уже отрицательное

		summaries = append(summaries, models.FinancialSummaryDB{
			ID:             uuid.New(),
			RequestID:      req.ID,
			InitialBalance: req.InitialBalance,
			FinalBalance:   finalBalance,
			TotalRevenue:   totalRevenue,
			TotalExpenses:  -totalExpenses, // Делаем положительным для хранения
			NetProfit:      netProfit,
		})
	}

	for _, summary := range summaries {
		if err := db.Create(&summary).Error; err != nil {
			log.Printf("  ⚠️  Failed to create financial summary: %v", err)
			continue
		}
	}

	log.Printf("  ✓ Seeded %d financial summaries", len(summaries))
	return nil
}

// seedDailyBalancesV2 - заполняет таблицу daily_balances
func seedDailyBalancesV2(db *gorm.DB) error {
	log.Println("📦 Seeding: Daily Balances")

	// Очищаем существующие ежедневные балансы (если есть)
	if err := db.Where("1 = 1").Delete(&models.DailyBalanceV2{}).Error; err != nil {
		log.Printf("  ⚠️  Warning: Could not clear existing daily balances: %v", err)
	}

	// Получаем completed generation requests
	var requests []models.GenerationRequest
	if err := db.Where("status = ?", "completed").Find(&requests).Error; err != nil {
		return fmt.Errorf("failed to get generation requests: %w", err)
	}

	if len(requests) == 0 {
		log.Println("  ⚠️  No completed generation requests found, skipping daily balances")
		return nil
	}

	var dailyBalances []models.DailyBalanceV2

	for _, req := range requests {
		// Получаем financial summary
		var summary models.FinancialSummaryDB
		if err := db.Where("request_id = ?", req.ID).First(&summary).Error; err != nil {
			log.Printf("  ⚠️  No financial summary found for request %s, skipping", req.ID)
			continue
		}

		// Создаем балансы для первых 7 дней месяца
		initialBalance := summary.InitialBalance
		balanceStep := (summary.FinalBalance - summary.InitialBalance) / 7.0

		for day := 1; day <= 7; day++ {
			balanceDate := time.Date(req.Year, time.Month(1), day, 0, 0, 0, 0, time.UTC)
			balance := initialBalance + (balanceStep * float64(day))

			dailyBalances = append(dailyBalances, models.DailyBalanceV2{
				ID:          uuid.New(),
				RequestID:   req.ID,
				BalanceDate: balanceDate,
				Balance:     balance,
			})
		}
	}

	if len(dailyBalances) > 0 {
		if err := db.Create(&dailyBalances).Error; err != nil {
			log.Printf("  ⚠️  Failed to create daily balances: %v", err)
			return err
		}
	}

	log.Printf("  ✓ Seeded %d daily balances", len(dailyBalances))
	return nil
}

// ============================================================================
// ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ
// ============================================================================

// stringPtr - создание указателя на string
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// floatPtr - создание указателя на float64
func floatPtr(f float64) *float64 {
	return &f
}

// intPtr - создание указателя на int
func intPtr(i int) *int {
	return &i
}
