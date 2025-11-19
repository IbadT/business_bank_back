package calculation

import (
	"context"
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/internal/helpers"
	"gorm.io/gorm"
)

type CalculationRepository interface {
	SaveStatement(ctx context.Context, id string, statement MatematikaResponse) error
	// GetStatementByID(ctx context.Context, id string) (*MatematikaResponse, error)
	GetStatementByID(ctx context.Context, id string) (*Statement, error)
	UpdateStatus(ctx context.Context, id string, status helpers.StatementStatus) error
	GetStatus(ctx context.Context, id string) (helpers.StatementStatus, error)
	Exists(ctx context.Context, id string) (bool, error)

	GetTransactions(ctx context.Context, startDate, endDate string) ([]Transaction, error)

	GetDailyBalances(ctx context.Context, startDate, endDate string) ([]DailyBalance, error)

	GetBusinessRules(ctx context.Context) ([]BusinessRule, error)
	GetStatements(ctx context.Context) ([]Statement, error)
	GetLastBalance(ctx context.Context, id string) (float64, error)

	// STATE !!!! ПРОВЕРИТЬ !!!!
	SaveState(ctx context.Context, state Statement) error
	LoadState(ctx context.Context, id string) (*Statement, error)
}

type calculationRepository struct {
	db *gorm.DB
}

func NewCalculationRepository(db *gorm.DB) CalculationRepository {
	return &calculationRepository{db: db}
}

// ================================================
// GET TRANSACTIONS
// ================================================
// Получает транзакции из БД в пределах указанного периода
// Параметры:
//   - ctx: Контекст запроса
//   - startDate: Дата начала периода
//   - endDate: Дата окончания периода
//
// Возвращает список транзакций или ошибку
// Если дата начала или окончания не указаны, то возвращает все транзакции
func (r *calculationRepository) GetTransactions(ctx context.Context, startDate string, endDate string) ([]Transaction, error) {
	var transactions []Transaction
	newStartDate, newEndDate, err := helpers.ParseDates(startDate, endDate)
	if err != nil {
		return nil, err
	}

	// проверить, мозможно возвращаются не все данные
	if err := r.db.WithContext(ctx).
		Where("transaction_date BETWEEN ? AND ?", newStartDate, newEndDate).
		Order("transaction_date ASC").
		Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

func (r *calculationRepository) SaveStatement(ctx context.Context, id string, statement MatematikaResponse) error {
	return nil
}

// func (r *calculationRepository) GetStatementByID(ctx context.Context, id string) (*MatematikaResponse, error) {
func (r *calculationRepository) GetStatementByID(ctx context.Context, id string) (*Statement, error) {
	var statement Statement
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&statement).Error; err != nil {
		return nil, err
	}
	return &statement, nil
}

func (r *calculationRepository) UpdateStatus(ctx context.Context, id string, status helpers.StatementStatus) error {
	if err := r.db.WithContext(ctx).Model(&Statement{}).Where("id = ?", id).Update("status", status).Error; err != nil {
		return err
	}
	return nil
}

func (r *calculationRepository) GetStatus(ctx context.Context, id string) (helpers.StatementStatus, error) {
	var statement Statement
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&statement).Error; err != nil {
		return "", err
	}
	return helpers.StatementStatus(statement.Status), nil
}

func (r *calculationRepository) Exists(ctx context.Context, id string) (bool, error) {
	var statement Statement
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&statement).Error; err != nil {
		return false, err
	}
	return true, nil
}

func (r *calculationRepository) GetBusinessRules(ctx context.Context) ([]BusinessRule, error) {
	var businessRules []BusinessRule
	if err := r.db.WithContext(ctx).Find(&businessRules).Error; err != nil {
		return nil, err
	}
	return businessRules, nil
}

type DailyBalanceResponse struct {
	Date    time.Time `json:"date"`
	Balance float64   `json:"balance"`
}

// func (r *calculationRepository) GetDailyBalances(ctx context.Context, startDate, endDate string) ([]DailyBalanceResponse, error) {
func (r *calculationRepository) GetDailyBalances(ctx context.Context, startDate, endDate string) ([]DailyBalance, error) {
	var dailyBalancesResponse []DailyBalance
	// var dailyBalancesResponse []DailyBalanceResponse
	startDate, endDate, err := helpers.ParseDates(startDate, endDate)
	if err != nil {
		return nil, err
	}

	if err := r.db.WithContext(ctx).
		Where("date BETWEEN ? AND ?", startDate, endDate).
		Order("date ASC").
		Find(&dailyBalancesResponse).Error; err != nil {
		return nil, err
	}
	return dailyBalancesResponse, nil
}

func (r *calculationRepository) GetStatements(ctx context.Context) ([]Statement, error) {
	var statements []Statement
	if err := r.db.WithContext(ctx).Find(&statements).Error; err != nil {
		return nil, err
	}
	return statements, nil
}

func (r *calculationRepository) GetLastBalance(ctx context.Context, id string) (float64, error) {
	var dailyBalances []DailyBalance
	if err := r.db.WithContext(ctx).Model(&DailyBalance{}).Where("statement_id = ?", id).Order("date DESC").First(&dailyBalances).Error; err != nil {
		return 0, err
	}
	return dailyBalances[0].Balance, nil
}

// ================================================
// STATE !!!! ПРОВЕРИТЬ
// ================================================

func (r *calculationRepository) SaveState(ctx context.Context, state Statement) error {
	if err := r.db.WithContext(ctx).Model(&Statement{}).Create(&state).Error; err != nil {
		return err
	}
	return nil
}

func (r *calculationRepository) LoadState(ctx context.Context, id string) (*Statement, error) {
	var state Statement
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&state).Error; err != nil {
		return nil, err
	}
	return &state, nil
}
