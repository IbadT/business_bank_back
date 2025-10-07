package main

import (
	"log"

	"github.com/IbadT/business_bank_back/services/matematika/internal/calculation"
	"github.com/IbadT/business_bank_back/services/matematika/internal/database"
	"github.com/joho/godotenv"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	e := echo.New()

	port := database.GetEnv("PORT", "8080")
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	calcRepo := calculation.NewCalculationRepository(db)
	calcService := calculation.NewCalculationService(calcRepo)
	_ = calculation.NewCalculationHandler(calcService)

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	// e.Use(middleware.CORS())

	log.Fatal(e.Start(":" + port))
}
