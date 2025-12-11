package main

import (
	"log"

	"github.com/IbadT/business_bank_back/services/matematika/internal/database"
	"github.com/IbadT/business_bank_back/services/matematika/internal/models"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/seeds"
	"github.com/joho/godotenv"
)

func main() {
	log.Println("========================================")
	log.Println("🌱 DATABASE SEEDING")
	log.Println("========================================")

	// Загружаем .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Подключаемся к БД
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("✓ Database connected")

	// Автоматическая миграция моделей
	log.Println("🔄 Running database migrations...")
	if err := db.AutoMigrate(
		&models.GenerationRequest{},
		&models.GeneratedTransaction{},
		&models.FinancialSummaryDB{},
		&models.DailyBalanceV2{},
		&models.TransactionTemplateDB{},
		&models.DefaultCustomerDB{},
		&models.Holiday{},
	); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}
	log.Println("✓ Database migrations completed")

	// Запускаем seeding
	if err := seeds.SeedDatabase(db); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	// Опционально: расширенные данные
	log.Println("")
	log.Println("========================================")
	log.Println("🌱 EXTENDED SEEDING (optional)")
	log.Println("========================================")

	// if err := seeds.SeedExtendedData(db); err != nil {
	// 	log.Fatalf("Failed to seed extended data: %v", err)
	// }

	log.Println("")
	log.Println("========================================")
	log.Println("✅ ALL SEEDING COMPLETED")
	log.Println("========================================")
	log.Println("")
	log.Println("📊 Seeded data summary:")
	log.Println("  • Holidays:                  from config/holidays.json")
	log.Println("  • Transaction Templates:     from config/templates.json")
	log.Println("  • Default Customers:         from config/customers.json")
	log.Println("  • Generation Requests:      3 requests (2 completed, 1 processing)")
	log.Println("  • Generated Transactions:   multiple transactions per request")
	log.Println("  • Financial Summaries:      for completed requests")
	log.Println("  • Daily Balances:           for completed requests")
	log.Println("")
}
