package seedservice

import (
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
	return seeds.SeedDatabase(s.db)
}
