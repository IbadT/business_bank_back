package repository

import (
	"github.com/IbadT/business_bank_back/services/matematika/internal/domain"
	"github.com/IbadT/business_bank_back/services/matematika/internal/models"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/helpers"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/logger"
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

	GetIncomeTransactionsByRequestID(requestID uuid.UUID) ([]models.GeneratedTransaction, error)
	GetExpenseTransactionsByRequestID(requestID uuid.UUID) ([]models.GeneratedTransaction, error)

	GetAdjustedTransactionsByRequestID(requestID uuid.UUID) ([]models.GeneratedTransaction, error)
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
	op := "repository.transaction.create"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"transaction_id": transaction.TransactionID,
		"request_id":     transaction.RequestID,
		"type":           transaction.Type,
		"amount":         transaction.Amount,
	})
	log.Info("Creating transaction")

	// TODO: вынести конвертацию в сервис и использовать только готовую модель, а не domain
	ormTransaction := r.domainToORM(transaction)
	if err := r.DB.Create(&ormTransaction).Error; err != nil {
		log.Error(err, "Failed to create transaction")
		return err
	}

	log.Success("Transaction created successfully")
	return nil
}

func (r *transactionRepository) CreateBatch(transactions []*domain.GeneratedTransaction) error {
	op := "repository.transaction.createBatch"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"count": len(transactions)})
	log.Info("Creating batch transactions")

	ormTransactions := make([]*models.GeneratedTransaction, len(transactions))
	for i, tx := range transactions {
		// TODO: вынести конвертацию в сервис и использовать только готовую модель, а не domain
		ormTx := r.domainToORM(tx)
		ormTransactions[i] = &ormTx
	}
	
	if err := r.DB.Create(ormTransactions).Error; err != nil {
		log.Error(err, "Failed to create batch transactions")
		return err
	}

	log.Success("Batch transactions created successfully")
	return nil
}

// domainToORM конвертирует domain.GeneratedTransaction в models.GeneratedTransaction
func (r *transactionRepository) domainToORM(tx *domain.GeneratedTransaction) models.GeneratedTransaction {
	var calculationDetails models.JSONB
	if tx.CalculationDetails != nil {
		calculationDetails = models.JSONB(tx.CalculationDetails)
	}

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
		CalculationDetails: calculationDetails,
		SortOrder:          tx.SortOrder,
	}
}

func (r *transactionRepository) GetByRequestID(requestID uuid.UUID) ([]domain.GeneratedTransaction, error) {
	op := "repository.transaction.getByRequestID"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"request_id": requestID})
	log.Info("Getting transactions by request ID")

	var ormTransactions []models.GeneratedTransaction
	// Используем Model() для явного указания таблицы
	if err := r.DB.Model(&models.GeneratedTransaction{}).
		Where("request_id = ?", requestID).
		Order("transaction_date ASC").
		Find(&ormTransactions).Error; err != nil {
		log.Error(err, "Failed to get transactions by request ID")
		return []domain.GeneratedTransaction{}, err
	}

	domainTransactions := make([]domain.GeneratedTransaction, len(ormTransactions))
	for i, tx := range ormTransactions {
		domainTransactions[i] = r.ormToDomain(tx)
	}
	
	log.WithFields(logger.Fields{"count": len(domainTransactions)}).Success("Transactions retrieved by request ID")
	// TODO: вернуть модель, а в сервисе конвертировать в domain
	return domainTransactions, nil
}

func (r *transactionRepository) GetCountByRequestID(requestID uuid.UUID) (int64, error) {
	op := "repository.transaction.getCountByRequestID"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"request_id": requestID})
	log.Info("Getting transactions count by request ID")

	var count int64
	if err := r.DB.Model(&models.GeneratedTransaction{}).Where("request_id = ?", requestID).Count(&count).Error; err != nil {
		log.Error(err, "Failed to get transactions count")
		return 0, err
	}
	
	log.WithFields(logger.Fields{"count": count}).Success("Transactions count retrieved")
	return count, nil
}

func (r *transactionRepository) GetByTypeAndRequestID(transactionType string, requestID uuid.UUID) ([]domain.GeneratedTransaction, error) {
	op := "repository.transaction.getByTypeAndRequestID"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"request_id": requestID,
		"type":       transactionType,
	})
	log.Info("Getting transactions by type and request ID")

	var ormTransactions []models.GeneratedTransaction
	// Используем Model() для явного указания таблицы
	// В GORM поле Type маппится в колонку "type" (без column тега используется snake_case)
	// Но "type" - зарезервированное слово в PostgreSQL, поэтому используем экранирование
	if err := r.DB.Table("generated_transactions").
		Where("request_id = ?", requestID).
		Where("\"type\" = ?", transactionType).
		Find(&ormTransactions).Error; err != nil {
		log.Error(err, "Failed to get transactions by type and request ID")
		return []domain.GeneratedTransaction{}, err
	}

	domainTransactions := make([]domain.GeneratedTransaction, len(ormTransactions))
	for i, tx := range ormTransactions {
		domainTransactions[i] = r.ormToDomain(tx)
	}
	
	log.WithFields(logger.Fields{"count": len(domainTransactions)}).Success("Transactions retrieved by type and request ID")
	// TODO: вернуть модель, а в сервисе конвертировать в domain
	return domainTransactions, nil
}

func (r *transactionRepository) GetByMethodAndRequestID(transactionMethod string, requestID uuid.UUID) ([]domain.GeneratedTransaction, error) {
	op := "repository.transaction.getByMethodAndRequestID"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"request_id": requestID,
		"method":     transactionMethod,
	})
	log.Info("Getting transactions by method and request ID")

	var ormTransactions []models.GeneratedTransaction
	// Используем Model() для явного указания таблицы
	if err := r.DB.Model(&models.GeneratedTransaction{}).
		Where("request_id = ? AND method = ?", requestID, transactionMethod).
		Find(&ormTransactions).Error; err != nil {
		log.Error(err, "Failed to get transactions by method and request ID")
		return []domain.GeneratedTransaction{}, err
	}

	domainTransactions := make([]domain.GeneratedTransaction, len(ormTransactions))
	for i, tx := range ormTransactions {
		domainTransactions[i] = r.ormToDomain(tx)
	}
	
	log.WithFields(logger.Fields{"count": len(domainTransactions)}).Success("Transactions retrieved by method and request ID")
	return domainTransactions, nil
}

// ormToDomain конвертирует models.GeneratedTransaction в domain.GeneratedTransaction
func (r *transactionRepository) ormToDomain(tx models.GeneratedTransaction) domain.GeneratedTransaction {
	var calculationDetails map[string]interface{}
	if tx.CalculationDetails != nil {
		calculationDetails = map[string]interface{}(tx.CalculationDetails)
	}

	return domain.GeneratedTransaction{
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
		SortOrder:          tx.SortOrder,
		CalculationDetails: calculationDetails,
	}
}

func (r *transactionRepository) GetIncomeTransactionsByRequestID(requestID uuid.UUID) ([]models.GeneratedTransaction, error) {
	op := "repository.transaction.getIncomeTransactionsByRequestID"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"request_id": requestID})
	log.Info("Getting income transactions by request ID")

	var transactions []models.GeneratedTransaction
	if err := r.DB.Model(&models.GeneratedTransaction{}).
		Where("request_id = ?", requestID).
		Where("\"type\" = ?", helpers.TransactionTypeIncomeStr).
		Order("transaction_date ASC").
		Find(&transactions).Error; err != nil {
		log.Error(err, "Failed to get income transactions")
		return []models.GeneratedTransaction{}, err
	}
	
	log.WithFields(logger.Fields{"count": len(transactions)}).Success("Income transactions retrieved")
	return transactions, nil
}
func (r *transactionRepository) GetExpenseTransactionsByRequestID(requestID uuid.UUID) ([]models.GeneratedTransaction, error) {
	op := "repository.transaction.getExpenseTransactionsByRequestID"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"request_id": requestID})
	log.Info("Getting expense transactions by request ID")

	var transactions []models.GeneratedTransaction
	if err := r.DB.Model(&models.GeneratedTransaction{}).
		Where("request_id = ?", requestID).
		Where("\"type\" = ?", helpers.TransactionTypeExpenseStr).
		Order("transaction_date ASC").
		Find(&transactions).Error; err != nil {
		log.Error(err, "Failed to get expense transactions")
		return []models.GeneratedTransaction{}, err
	}
	
	log.WithFields(logger.Fields{"count": len(transactions)}).Success("Expense transactions retrieved")
	return transactions, nil
}

func (r *transactionRepository) GetAdjustedTransactionsByRequestID(requestID uuid.UUID) ([]models.GeneratedTransaction, error) {
	op := "repository.transaction.getAdjustedTransactionsByRequestID"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"request_id": requestID})
	log.Info("Getting adjusted transactions by request ID")

	var transactions []models.GeneratedTransaction
	if err := r.DB.Model(&models.GeneratedTransaction{}).
		Where("request_id = ?", requestID).Where("calculation_details->>'was_adjusted' = 'true'").
		Order("transaction_date ASC").
		Find(&transactions).Error; err != nil {
		log.Error(err, "Failed to get adjusted transactions")
		return []models.GeneratedTransaction{}, err
	}

	result := make([]models.GeneratedTransaction, len(transactions))
	for i, tx := range transactions {
		result[i] = tx
	}

	log.WithFields(logger.Fields{"count": len(result)}).Success("Adjusted transactions retrieved")
	return transactions, nil
}
