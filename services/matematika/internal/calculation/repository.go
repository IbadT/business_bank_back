package calculation

import "gorm.io/gorm"

type CalculationRepository interface {
	GenerateStatement() error
	GetStatementStatusByID(id string) error
	GetStatementResultByID(id string) error
}

type calculationRepository struct {
	db *gorm.DB
}

func NewCalculationRepository(db *gorm.DB) CalculationRepository {
	return &calculationRepository{db: db}
}

// - POST /generate-statement
// - GET /statement/{id}/status
// - GET /statement/{id}/result

func (r *calculationRepository) GenerateStatement() error {
	return nil
}

func (r *calculationRepository) GetStatementStatusByID(id string) error {
	return nil
}

func (r *calculationRepository) GetStatementResultByID(id string) error {
	return nil
}
