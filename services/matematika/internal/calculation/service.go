package calculation

import "context"

type CalculationService interface {
	GenerateStatement(ctx context.Context) error
	GetStatementStatusByID(ctx context.Context, id string) error
	GetStatementResultByID(ctx context.Context, id string) error
}

type calculationService struct {
	calcRepo CalculationRepository
}

func NewCalculationService(calcRepo CalculationRepository) CalculationService {
	return &calculationService{calcRepo: calcRepo}
}

// - POST /generate-statement
// - GET /statement/{id}/status
// - GET /statement/{id}/result

func (s *calculationService) GenerateStatement(ctx context.Context) error {
	return nil
}

func (s *calculationService) GetStatementStatusByID(ctx context.Context, id string) error {
	return nil
}

func (s *calculationService) GetStatementResultByID(ctx context.Context, id string) error {
	return nil
}
