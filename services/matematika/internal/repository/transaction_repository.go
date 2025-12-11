package repository

import (
	"github.com/IbadT/business_bank_back/services/matematika/internal/domain"
	"github.com/IbadT/business_bank_back/services/matematika/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TransactionRepository interface {
	Create(transaction *domain.GeneratedTransaction) error
	CreateBatch(transactions []*domain.GeneratedTransaction) error
	GetByRequestID(requestID uuid.UUID) ([]domain.GeneratedTransaction, error)
	GetCountByRequestID(requestID uuid.UUID) (int64, error)
	GetByTypeAndRequestID(transactionType string, requestID uuid.UUID) ([]domain.GeneratedTransaction, error)
	GetByMethodAndRequestID(transactionMethod string, requestID uuid.UUID) ([]domain.GeneratedTransaction, error)
}

type transactionRepository struct {
	DB *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepository{
		DB: db,
	}
}

func (r *transactionRepository) Create(transaction *domain.GeneratedTransaction) error {
	// TODO: вынести конвертацию в сервис и использовать только готовую модель, а не domain
	ormTransaction := r.domainToORM(transaction)
	return r.DB.Create(&ormTransaction).Error
}

func (r *transactionRepository) CreateBatch(transactions []*domain.GeneratedTransaction) error {
	ormTransactions := make([]*models.GeneratedTransaction, len(transactions))
	for i, tx := range transactions {
		// TODO: вынести конвертацию в сервис и использовать только готовую модель, а не domain
		ormTx := r.domainToORM(tx)
		ormTransactions[i] = &ormTx
	}
	return r.DB.Create(ormTransactions).Error
}

// domainToORM конвертирует domain.GeneratedTransaction в models.GeneratedTransaction
func (r *transactionRepository) domainToORM(tx *domain.GeneratedTransaction) models.GeneratedTransaction {
	return models.GeneratedTransaction{
		ID:                 tx.ID,
		RequestID:          tx.RequestID,
		TransactionID:      tx.TransactionID,
		TransactionDate:    tx.TransactionDate,
		PostingDate:        tx.PostingDate,
		Type:               tx.Type,
		Category:           tx.Category,
		Method:             tx.Method,
		Amount:             tx.Amount,
		BalanceAfter:       tx.BalanceAfter,
		IsManual:           tx.IsManual,
		CalculationDetails: nil, // Можно добавить позже если нужно
		SortOrder:          tx.SortOrder,
	}
}

func (r *transactionRepository) GetByRequestID(requestID uuid.UUID) ([]domain.GeneratedTransaction, error) {
	var ormTransactions []models.GeneratedTransaction
	// Используем Model() для явного указания таблицы
	if err := r.DB.Model(&models.GeneratedTransaction{}).
		Where("request_id = ?", requestID).
		Find(&ormTransactions).Error; err != nil {
		return []domain.GeneratedTransaction{}, err
	}

	domainTransactions := make([]domain.GeneratedTransaction, len(ormTransactions))
	for i, tx := range ormTransactions {
		domainTransactions[i] = r.ormToDomain(tx)
	}
	// TODO: вернуть модель, а в сервисе конвертировать в domain
	return domainTransactions, nil
}

func (r *transactionRepository) GetCountByRequestID(requestID uuid.UUID) (int64, error) {
	var count int64
	if err := r.DB.Model(&models.GeneratedTransaction{}).Where("request_id = ?", requestID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *transactionRepository) GetByTypeAndRequestID(transactionType string, requestID uuid.UUID) ([]domain.GeneratedTransaction, error) {
	var ormTransactions []models.GeneratedTransaction
	// Используем Model() для явного указания таблицы
	// В GORM поле Type маппится в колонку "type" (без column тега используется snake_case)
	// Но "type" - зарезервированное слово в PostgreSQL, поэтому используем экранирование
	if err := r.DB.Table("generated_transactions").
		Where("request_id = ?", requestID).
		Where("\"type\" = ?", transactionType).
		Find(&ormTransactions).Error; err != nil {
		return []domain.GeneratedTransaction{}, err
	}

	domainTransactions := make([]domain.GeneratedTransaction, len(ormTransactions))
	for i, tx := range ormTransactions {
		domainTransactions[i] = r.ormToDomain(tx)
	}
	// TODO: вернуть модель, а в сервисе конвертировать в domain
	return domainTransactions, nil
}

func (r *transactionRepository) GetByMethodAndRequestID(transactionMethod string, requestID uuid.UUID) ([]domain.GeneratedTransaction, error) {
	var ormTransactions []models.GeneratedTransaction
	// Используем Model() для явного указания таблицы
	if err := r.DB.Model(&models.GeneratedTransaction{}).
		Where("request_id = ? AND method = ?", requestID, transactionMethod).
		Find(&ormTransactions).Error; err != nil {
		return []domain.GeneratedTransaction{}, err
	}

	domainTransactions := make([]domain.GeneratedTransaction, len(ormTransactions))
	for i, tx := range ormTransactions {
		domainTransactions[i] = r.ormToDomain(tx)
	}
	return domainTransactions, nil
}

// ormToDomain конвертирует models.GeneratedTransaction в domain.GeneratedTransaction
func (r *transactionRepository) ormToDomain(tx models.GeneratedTransaction) domain.GeneratedTransaction {
	return domain.GeneratedTransaction{
		ID:              tx.ID,
		RequestID:       tx.RequestID,
		TransactionID:   tx.TransactionID,
		TransactionDate: tx.TransactionDate,
		PostingDate:     tx.PostingDate,
		Type:            tx.Type,
		Category:        tx.Category,
		Method:          tx.Method,
		Amount:          tx.Amount,
		BalanceAfter:    tx.BalanceAfter,
		IsManual:        tx.IsManual,
		SortOrder:       tx.SortOrder,
	}
}
