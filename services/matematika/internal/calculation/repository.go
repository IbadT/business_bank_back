package calculation

import (
	"context"

	"gorm.io/gorm"
)

type CalculationRepository interface {
	SaveStatement(ctx context.Context, id string, statement MatematikaResponse) error
	GetStatementByID(ctx context.Context, id string) (*MatematikaResponse, error)
	UpdateStatus(ctx context.Context, id string, status StatementStatus) error
	GetStatus(ctx context.Context, id string) (StatementStatus, error)
	Exists(ctx context.Context, id string) (bool, error)
	// SaveState(ctx context.Context, state StatementState) error
	// LoadState(ctx context.Context, id string) (*StatementState, error)
	// GetLastBalance(ctx context.Context, id string) (float64, error)
}

type calculationRepository struct {
	db *gorm.DB
}

func NewCalculationRepository(db *gorm.DB) CalculationRepository {
	return &calculationRepository{db: db}
}

func (r *calculationRepository) SaveStatement(ctx context.Context, id string, statement MatematikaResponse) error {
	return nil
}

func (r *calculationRepository) GetStatementByID(ctx context.Context, id string) (*MatematikaResponse, error) {
	return &MatematikaResponse{}, nil
}

func (r *calculationRepository) UpdateStatus(ctx context.Context, id string, status StatementStatus) error {
	return nil
}

func (r *calculationRepository) GetStatus(ctx context.Context, id string) (StatementStatus, error) {
	return StatementStatus(""), nil
}

func (r *calculationRepository) Exists(ctx context.Context, id string) (bool, error) {
	return false, nil
}
