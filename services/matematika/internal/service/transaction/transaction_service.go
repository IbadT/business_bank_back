package transactionservice

import (
	"errors"
	"fmt"
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/internal/domain"
	"github.com/IbadT/business_bank_back/services/matematika/internal/domain/value_objects"
	"github.com/IbadT/business_bank_back/services/matematika/internal/repository"
	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/dto"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/utils"
	"github.com/google/uuid"
)

var (
	ErrInvalidRequestID       = errors.New("invalid requestId format")
	ErrInvalidDate            = errors.New("invalid date format. Expected ISO8601 (YYYY-MM-DDTHH:MM:SSZ) or YYYY-MM-DD")
	ErrInvalidTransactionType = errors.New("invalid transaction type. Must be 'income' or 'expense'")
	ErrInvalidAmount          = errors.New("amount must be greater than 0")
	ErrEmptyCategory          = errors.New("category is required")
	ErrEmptyMethod            = errors.New("method is required")
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
	transactionRepo repository.TransactionRepository
}

func NewTransactionService(transactionRepo repository.TransactionRepository) TransactionService {
	return &transactionService{
		transactionRepo: transactionRepo,
	}
}

func (s *transactionService) CreateTransaction(req *dto.CreateTransactionRequest) error {
	// Валидация и парсинг requestID
	requestID, err := uuid.Parse(req.RequestID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidRequestID, err)
	}

	// Валидация и парсинг transactionDate (дата + время в формате ISO8601)
	var transactionDate time.Time

	// Пробуем ISO8601 формат (2006-01-02T15:04:05Z07:00)
	transactionDate, err = time.Parse(time.RFC3339, req.TransactionDate)
	if err != nil {
		// Пробуем упрощенный ISO8601 без timezone (2006-01-02T15:04:05)
		transactionDate, err = time.Parse("2006-01-02T15:04:05", req.TransactionDate)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidDate, err)
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
			return fmt.Errorf("%w: %v", ErrInvalidDate, err)
		}
		// PostingDate - это только дата (без времени)
		postingDate = time.Date(postingDate.Year(), postingDate.Month(), postingDate.Day(), 0, 0, 0, 0, time.UTC)
	} else {
		// Если postingDate не указана, используем дату из transactionDate (без времени)
		postingDate = time.Date(transactionDate.Year(), transactionDate.Month(), transactionDate.Day(), 0, 0, 0, 0, time.UTC)
	}

	// Валидация типа транзакции
	// TODO: использовать типы для переменных
	if req.Type != "income" && req.Type != "expense" {
		return ErrInvalidTransactionType
	}

	// Валидация категории
	if req.Category == "" {
		return ErrEmptyCategory
	}

	// Валидация метода
	if req.Method == "" {
		return ErrEmptyMethod
	}
	// Проверка валидности метода платежа
	if _, err := value_objects.NewPaymentMethod(req.Method); err != nil {
		return fmt.Errorf("invalid payment method: %w", err)
	}

	// Валидация суммы
	if req.Amount <= 0 {
		return ErrInvalidAmount
	}

	// Генерация transactionID
	transactionID := utils.GenerateTransactionID("manual", 1)

	// Создание доменной сущности
	transaction := domain.NewGeneratedTransaction(requestID, transactionID, transactionDate, postingDate, req.Type, req.Category, req.Method, req.Amount)

	// Сохранение в БД
	return s.transactionRepo.Create(transaction)
}

func (s *transactionService) CreateBatchTransactions(req *dto.CreateBatchTransactionsRequest) error {
	if len(req.Transactions) == 0 {
		return errors.New("transactions array cannot be empty")
	}

	transactions := make([]*domain.GeneratedTransaction, len(req.Transactions))

	for i, tx := range req.Transactions {
		// Валидация и парсинг requestID
		requestID, err := uuid.Parse(tx.RequestID)
		if err != nil {
			return fmt.Errorf("invalid requestId format at index %d: %w", i, err)
		}

		// Проверка, что все транзакции имеют одинаковый requestID
		// TODO: могут ли они иметь разные requestID?
		if i > 0 {
			firstRequestID, _ := uuid.Parse(req.Transactions[0].RequestID)
			if requestID != firstRequestID {
				return fmt.Errorf("all transactions must have the same requestId, but found different at index %d", i)
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
				return fmt.Errorf("invalid transaction date format at index %d: %w", i, err)
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
				return fmt.Errorf("invalid posting date format at index %d: %w", i, err)
			}
			// PostingDate - это только дата (без времени)
			postingDate = time.Date(postingDate.Year(), postingDate.Month(), postingDate.Day(), 0, 0, 0, 0, time.UTC)
		} else {
			// Если postingDate не указана, используем дату из transactionDate (без времени)
			postingDate = time.Date(transactionDate.Year(), transactionDate.Month(), transactionDate.Day(), 0, 0, 0, 0, time.UTC)
		}

		// Валидация типа транзакции
		// TODO: использовать типы для переменных
		if tx.Type != "income" && tx.Type != "expense" {
			return fmt.Errorf("invalid transaction type at index %d: must be 'income' or 'expense'", i)
		}

		// Валидация категории
		if tx.Category == "" {
			return fmt.Errorf("category is required at index %d", i)
		}

		// Валидация метода
		if tx.Method == "" {
			return fmt.Errorf("method is required at index %d", i)
		}
		// Проверка валидности метода платежа
		if _, err := value_objects.NewPaymentMethod(tx.Method); err != nil {
			return fmt.Errorf("invalid payment method at index %d: %w", i, err)
		}

		// Валидация суммы
		if tx.Amount <= 0 {
			return fmt.Errorf("amount must be greater than 0 at index %d", i)
		}

		// Генерация transactionID для каждой транзакции
		transactionID := utils.GenerateTransactionID("manual", i+1)

		// Создание доменной сущности
		transactions[i] = domain.NewGeneratedTransaction(requestID, transactionID, transactionDate, postingDate, tx.Type, tx.Category, tx.Method, tx.Amount)
	}

	// Сохранение в БД
	return s.transactionRepo.CreateBatch(transactions)
}

func (s *transactionService) GetByRequestID(requestIDStr string) ([]domain.GeneratedTransaction, error) {
	requestID, err := uuid.Parse(requestIDStr)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidRequestID, err)
	}

	return s.transactionRepo.GetByRequestID(requestID)
}

func (s *transactionService) GetCountByRequestID(requestIDStr string) (int64, error) {
	requestID, err := uuid.Parse(requestIDStr)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrInvalidRequestID, err)
	}

	return s.transactionRepo.GetCountByRequestID(requestID)
}

func (s *transactionService) GetByTypeAndRequestID(transactionType string, requestIDStr string) ([]domain.GeneratedTransaction, error) {
	// Валидация типа транзакции
	// TODO: использовать типы для переменных
	if transactionType != "income" && transactionType != "expense" {
		return nil, ErrInvalidTransactionType
	}

	requestID, err := uuid.Parse(requestIDStr)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidRequestID, err)
	}

	return s.transactionRepo.GetByTypeAndRequestID(transactionType, requestID)
}

func (s *transactionService) GetByMethodAndRequestID(transactionMethod string, requestIDStr string) ([]domain.GeneratedTransaction, error) {
	if transactionMethod == "" {
		return nil, ErrEmptyMethod
	}

	requestID, err := uuid.Parse(requestIDStr)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidRequestID, err)
	}

	return s.transactionRepo.GetByMethodAndRequestID(transactionMethod, requestID)
}
