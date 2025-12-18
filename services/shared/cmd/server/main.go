package main

import (
	"github.com/IbadT/business_bank_back/services/shared/internal/app"
	"github.com/IbadT/business_bank_back/services/shared/internal/database"
	"github.com/sirupsen/logrus"
)

func main() {
	cfg := &app.Config{
		Port:     database.GetEnv("PORT", "8083"),
		GRPCPort: database.GetEnv("GRPC_PORT", "9093"),
	}

	application := app.NewApp(cfg)
	if err := application.Run(); err != nil {
		logrus.Fatalf("Failed to run application: %v", err)
	}
}
