package seedservice

import (
	"github.com/IbadT/business_bank_back/services/matematika/pkg/logger"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/seeds"
	"gorm.io/gorm"
)

type SeedService interface {
	SeedDatabase() error
}

type seedService struct {
	db *gorm.DB
}

func NewSeedService(db *gorm.DB) SeedService {
	return &seedService{
		db: db,
	}
}

func (s *seedService) SeedDatabase() error {
	op := "service.seed.seedDatabase"
	log := logger.GetLogger().WithOperation(op)
	log.Info("Starting database seeding")

	if err := seeds.SeedDatabase(s.db); err != nil {
		log.Error(err, "Failed to seed database")
		return err
	}

	log.Success("Database seeded successfully")
	return nil
}
