package main

import (
	"log"

	"github.com/IbadT/business_bank_back/services/matematika/internal/database"
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

	// Запускаем seeding
	if err := database.SeedDatabase(db); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	// Опционально: расширенные данные
	log.Println("")
	log.Println("========================================")
	log.Println("🌱 EXTENDED SEEDING (optional)")
	log.Println("========================================")

	if err := database.SeedExtendedData(db); err != nil {
		log.Fatalf("Failed to seed extended data: %v", err)
	}

	log.Println("")
	log.Println("========================================")
	log.Println("✅ ALL SEEDING COMPLETED")
	log.Println("========================================")
	log.Println("")
	log.Println("📊 Seeded data summary:")
	log.Println("  • Srb Autos LLC (B2C):        2 months")
	log.Println("  • TechCorp Industries (B2B):  1 month")
	log.Println("  • FastFood LLC (B2C):         3 months")
	log.Println("  • Construction LLC (B2B):     1 month")
	log.Println("  • RetailStore Inc (B2C):      1 month")
	log.Println("")
	log.Println("  TOTAL: 5 companies, 8 statements")
	log.Println("")
}
