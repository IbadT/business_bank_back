package cache

import "github.com/IbadT/business_bank_back/services/matematika/pkg/redis"

type Repository struct {
	rds *redis.RDS
}

func NewRepository(rds *redis.RDS) *Repository {
	return &Repository{
		rds: rds,
	}
}
