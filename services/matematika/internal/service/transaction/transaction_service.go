package transactionservice

import (
	"errors"
	"fmt"
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/internal/domain"
	"github.com/IbadT/business_bank_back/services/matematika/internal/domain/value_objects"
	"github.com/IbadT/business_bank_back/services/matematika/internal/repository"
	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/dto"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/helpers"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/logger"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TransactionService interface {
	CreateTransaction(req *dto.CreateTransactionRequest) error
	CreateBatchTransactions(req *dto.CreateBatchTransactionsRequest) error
	GetByRequestID(requestIDStr string) ([]domain.GeneratedTransaction, error)
	GetCountByRequestID(requestIDStr string) (int64, error)
	GetByTypeAndRequestID(transactionType string, requestIDStr string) ([]domain.GeneratedTransaction, error)
	GetByMethodAndRequestID(transactionMethod string, requestIDStr string) ([]domain.GeneratedTransaction, error)
}

type transactionService struct {
	transactionRepo      repository.TransactionRepository
	generationRequestRepo repository.GenerationRequestRepository
}

func NewTransactionService(transactionRepo repository.TransactionRepository, generationRequestRepo repository.GenerationRequestRepository) TransactionService {
	return &transactionService{
		transactionRepo:      transactionRepo,
		generationRequestRepo: generationRequestRepo,
	}
}

func (s *transactionService) CreateTransaction(req *dto.CreateTransactionRequest) error {
	op := "service.transaction.createTransaction"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"request_id": req.RequestID,
		"type":       req.Type,
		"method":     req.Method,
		"amount":     req.Amount,
	})
	log.Info("Creating transaction")

	// Валидация и парсинг requestID
	requestID, err := helpers.ParseUUID(req.RequestID)
	if err != nil {
		log.Error(err, "Invalid requestID format")
		return err
	}

	// Проверка существования request_id в базе данных
	_, err = s.generationRequestRepo.GetByID(requestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Warn("RequestID not found in database: %s", requestID)
			return helpers.ErrRequestIDNotFound
		}
		log.Error(err, "Failed to check requestID existence")
		return fmt.Errorf("failed to check requestID existence: %w", err)
	}

	// Валидация и парсинг transactionDate (дата + время в формате ISO8601)
	var transactionDate time.Time

	// Пробуем ISO8601 формат (2006-01-02T15:04:05Z07:00)
	transactionDate, err = time.Parse(time.RFC3339, req.TransactionDate)
	if err != nil {
		// Пробуем упрощенный ISO8601 без timezone (2006-01-02T15:04:05)
		transactionDate, err = time.Parse("2006-01-02T15:04:05", req.TransactionDate)
		if err != nil {
			log.Error(err, "Invalid transaction date format: %s", req.TransactionDate)
			return fmt.Errorf("%w: %w", helpers.ErrInvalidDateFormat, err)
		}
		// Если время указано без timezone, используем UTC
		transactionDate = time.Date(transactionDate.Year(), transactionDate.Month(), transactionDate.Day(),
			transactionDate.Hour(), transactionDate.Minute(), transactionDate.Second(), 0, time.UTC)
	}

	// Парсинг postingDate (если не указана, используем transactionDate)
	postingDate := transactionDate
	if req.PostingDate != "" {
		postingDate, err = time.Parse("2006-01-02", req.PostingDate)
		if err != nil {
			return fmt.Errorf("%w: %w", helpers.ErrInvalidDateFormat, err)
		}
		// PostingDate - это только дата (без времени)
		postingDate = time.Date(postingDate.Year(), postingDate.Month(), postingDate.Day(), 0, 0, 0, 0, time.UTC)
	} else {
		// Если postingDate не указана, используем дату из transactionDate (без времени)
		postingDate = time.Date(transactionDate.Year(), transactionDate.Month(), transactionDate.Day(), 0, 0, 0, 0, time.UTC)
	}

	// Валидация типа транзакции
	if req.Type != helpers.TransactionTypeIncomeStr && req.Type != helpers.TransactionTypeExpenseStr {
		log.Warn("Invalid transaction type: %s", req.Type)
		return helpers.ErrInvalidTransactionType
	}

	// Валидация категории
	if req.Category == "" {
		log.Warn("Category is required")
		return helpers.ErrEmptyCategory
	}

	// Валидация метода
	if req.Method == "" {
		log.Warn("Method is required")
		return helpers.ErrEmptyMethod
	}
	// Проверка валидности метода платежа
	if _, err := value_objects.NewPaymentMethod(req.Method); err != nil {
		log.Error(err, "Invalid payment method: %s", req.Method)
		return fmt.Errorf("invalid payment method: %w", helpers.ErrInvalidPaymentMethod)
	}

	// Валидация суммы
	if req.Amount <= 0 {
		log.Warn("Invalid amount: %f", req.Amount)
		return helpers.ErrInvalidAmount
	}

	// Генерация transactionID
	transactionID := utils.GenerateTransactionID("manual", 1)

	// Создание доменной сущности
	transaction := domain.NewGeneratedTransaction(requestID, transactionID, transactionDate, postingDate, req.Type, req.Category, req.Method, req.Amount)

	// Сохранение в БД
	if err := s.transactionRepo.Create(transaction); err != nil {
		log.Error(err, "Failed to create transaction in repository")
		return err
	}

	log.Success("Transaction created successfully")
	return nil
}

func (s *transactionService) CreateBatchTransactions(req *dto.CreateBatchTransactionsRequest) error {
	op := "service.transaction.createBatchTransactions"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"count": len(req.Transactions)})
	log.Info("Creating batch transactions")

	if len(req.Transactions) == 0 {
		log.Warn("Transactions array is empty")
		return helpers.ErrTransactionsArrayEmpty
	}

	transactions := make([]*domain.GeneratedTransaction, len(req.Transactions))

	// Проверяем request_id один раз для первой транзакции (все должны иметь одинаковый request_id)
	var firstRequestID uuid.UUID
	if len(req.Transactions) > 0 {
		var err error
		firstRequestID, err = helpers.ParseUUID(req.Transactions[0].RequestID)
		if err != nil {
			log.Error(err, "Invalid requestID format for first transaction")
			return fmt.Errorf("invalid requestId format for first transaction: %w", err)
		}

		// Проверка существования request_id в базе данных
		_, err = s.generationRequestRepo.GetByID(firstRequestID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Warn("RequestID not found in database: %s", firstRequestID)
				return helpers.ErrRequestIDNotFound
			}
			log.Error(err, "Failed to check requestID existence")
			return fmt.Errorf("failed to check requestID existence: %w", err)
		}
	}

	for i, tx := range req.Transactions {
		// Валидация и парсинг requestID
		requestID, err := helpers.ParseUUID(tx.RequestID)
		if err != nil {
			log.Error(err, "Invalid requestID format at index %d", i)
			return fmt.Errorf("invalid requestId format at index %d: %w", i, err)
		}

		// Проверка, что все транзакции имеют одинаковый requestID
		// TODO: могут ли они иметь разные requestID?
		if i > 0 {
			if requestID != firstRequestID {
				return fmt.Errorf("%w at index %d", helpers.ErrAllTransactionsSameRequestID, i)
			}
		}

		// Валидация и парсинг transactionDate (дата + время в формате ISO8601)
		var transactionDate time.Time

		// Пробуем ISO8601 формат (2006-01-02T15:04:05Z07:00)
		// TODO: вынести логику проверки в отдельный метод в helper или validation(используется часто)
		transactionDate, err = time.Parse(time.RFC3339, tx.TransactionDate)
		if err != nil {
			// Пробуем упрощенный ISO8601 без timezone (2006-01-02T15:04:05)
			transactionDate, err = time.Parse("2006-01-02T15:04:05", tx.TransactionDate)
			if err != nil {
				log.Error(err, "Invalid transaction date format at index %d: %s", i, tx.TransactionDate)
				return fmt.Errorf("%w at index %d: %w", helpers.ErrInvalidTransactionDateFormat, i, err)
			}
			// Если время указано без timezone, используем UTC
			transactionDate = time.Date(transactionDate.Year(), transactionDate.Month(), transactionDate.Day(),
				transactionDate.Hour(), transactionDate.Minute(), transactionDate.Second(), 0, time.UTC)
		}

		// Парсинг postingDate (если не указана, используем transactionDate)
		postingDate := transactionDate
		if tx.PostingDate != "" {
			postingDate, err = time.Parse("2006-01-02", tx.PostingDate)
			if err != nil {
				return fmt.Errorf("%w at index %d: %w", helpers.ErrInvalidPostingDateFormat, i, err)
			}
			// PostingDate - это только дата (без времени)
			postingDate = time.Date(postingDate.Year(), postingDate.Month(), postingDate.Day(), 0, 0, 0, 0, time.UTC)
		} else {
			// Если postingDate не указана, используем дату из transactionDate (без времени)
			postingDate = time.Date(transactionDate.Year(), transactionDate.Month(), transactionDate.Day(), 0, 0, 0, 0, time.UTC)
		}

		// Валидация типа транзакции
		if tx.Type != helpers.TransactionTypeIncomeStr && tx.Type != helpers.TransactionTypeExpenseStr {
			log.Warn("Invalid transaction type at index %d: %s", i, tx.Type)
			return fmt.Errorf("%w at index %d", helpers.ErrInvalidTransactionType, i)
		}

		// Валидация категории
		if tx.Category == "" {
			log.Warn("Category is required at index %d", i)
			return fmt.Errorf("%w at index %d", helpers.ErrEmptyCategory, i)
		}

		// Валидация метода
		if tx.Method == "" {
			log.Warn("Method is required at index %d", i)
			return fmt.Errorf("%w at index %d", helpers.ErrEmptyMethod, i)
		}
		// Проверка валидности метода платежа
		if _, err := value_objects.NewPaymentMethod(tx.Method); err != nil {
			log.Error(err, "Invalid payment method at index %d: %s", i, tx.Method)
			return fmt.Errorf("%w at index %d: %w", helpers.ErrInvalidPaymentMethod, i, err)
		}

		// Валидация суммы
		if tx.Amount <= 0 {
			log.Warn("Invalid amount at index %d: %f", i, tx.Amount)
			return fmt.Errorf("%w at index %d", helpers.ErrInvalidAmount, i)
		}

		// Генерация transactionID для каждой транзакции
		transactionID := utils.GenerateTransactionID("manual", i+1)

		// Создание доменной сущности
		transactions[i] = domain.NewGeneratedTransaction(requestID, transactionID, transactionDate, postingDate, tx.Type, tx.Category, tx.Method, tx.Amount)
	}

	// Сохранение в БД
	if err := s.transactionRepo.CreateBatch(transactions); err != nil {
		log.Error(err, "Failed to create batch transactions in repository")
		return err
	}

	log.Success("Batch transactions created successfully")
	return nil
}

func (s *transactionService) GetByRequestID(requestIDStr string) ([]domain.GeneratedTransaction, error) {
	op := "service.transaction.getByRequestID"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"request_id": requestIDStr})
	log.Info("Getting transactions by request ID")

	requestID, err := helpers.ParseUUID(requestIDStr)
	if err != nil {
		log.Error(err, "Invalid requestID format")
		return nil, err
	}

	transactions, err := s.transactionRepo.GetByRequestID(requestID)
	if err != nil {
		log.Error(err, "Failed to get transactions from repository")
		return nil, err
	}

	log.WithFields(logger.Fields{"count": len(transactions)}).Success("Transactions retrieved by request ID")
	return transactions, nil
}

func (s *transactionService) GetCountByRequestID(requestIDStr string) (int64, error) {
	op := "service.transaction.getCountByRequestID"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"request_id": requestIDStr})
	log.Info("Getting transactions count by request ID")

	requestID, err := helpers.ParseUUID(requestIDStr)
	if err != nil {
		log.Error(err, "Invalid requestID format")
		return 0, err
	}

	count, err := s.transactionRepo.GetCountByRequestID(requestID)
	if err != nil {
		log.Error(err, "Failed to get transactions count from repository")
		return 0, err
	}

	log.WithFields(logger.Fields{"count": count}).Success("Transactions count retrieved")
	return count, nil
}

func (s *transactionService) GetByTypeAndRequestID(transactionType string, requestIDStr string) ([]domain.GeneratedTransaction, error) {
	op := "service.transaction.getByTypeAndRequestID"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"request_id": requestIDStr,
		"type":       transactionType,
	})
	log.Info("Getting transactions by type and request ID")

	// Валидация типа транзакции
	if transactionType != helpers.TransactionTypeIncomeStr && transactionType != helpers.TransactionTypeExpenseStr {
		log.Warn("Invalid transaction type: %s", transactionType)
		return nil, helpers.ErrInvalidTransactionType
	}

	requestID, err := helpers.ParseUUID(requestIDStr)
	if err != nil {
		log.Error(err, "Invalid requestID format")
		return nil, err
	}

	transactions, err := s.transactionRepo.GetByTypeAndRequestID(transactionType, requestID)
	if err != nil {
		log.Error(err, "Failed to get transactions from repository")
		return nil, err
	}

	log.WithFields(logger.Fields{"count": len(transactions)}).Success("Transactions retrieved by type and request ID")
	return transactions, nil
}

func (s *transactionService) GetByMethodAndRequestID(transactionMethod string, requestIDStr string) ([]domain.GeneratedTransaction, error) {
	op := "service.transaction.getByMethodAndRequestID"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"request_id": requestIDStr,
		"method":     transactionMethod,
	})
	log.Info("Getting transactions by method and request ID")

	if transactionMethod == "" {
		log.Warn("Method is required")
		return nil, helpers.ErrEmptyMethod
	}

	requestID, err := helpers.ParseUUID(requestIDStr)
	if err != nil {
		log.Error(err, "Invalid requestID format")
		return nil, err
	}

	transactions, err := s.transactionRepo.GetByMethodAndRequestID(transactionMethod, requestID)
	if err != nil {
		log.Error(err, "Failed to get transactions from repository")
		return nil, err
	}

	log.WithFields(logger.Fields{"count": len(transactions)}).Success("Transactions retrieved by method and request ID")
	return transactions, nil
}
