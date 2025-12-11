package service

import (
	"fmt"
	"math"
	"strings"

	"github.com/IbadT/business_bank_back/services/matematika/internal/domain"
	"github.com/IbadT/business_bank_back/services/matematika/internal/repository"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/transport"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/utils"
	"github.com/google/uuid"
)

type BreakdownService interface {
	GetRevenueBreakdown(requestIDStr string) (*transport.RevenueBreakdown, error)
	GetExpensesBreakdown(requestIDStr string) (*transport.ExpensesBreakdown, error)
}

type breakdownService struct {
	transactionRepo repository.TransactionRepository
}

func NewBreakdownService(transactionRepo repository.TransactionRepository) BreakdownService {
	return &breakdownService{
		transactionRepo: transactionRepo,
	}
}

func (s *breakdownService) GetRevenueBreakdown(requestIDStr string) (*transport.RevenueBreakdown, error) {
	requestID, err := uuid.Parse(requestIDStr)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidRequestID, err)
	}
	transactions, err := s.transactionRepo.GetIncomeTransactionsByRequestID(requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to get income transactions: %w", err)
	}

	incomeTransactions := make([]domain.GeneratedTransaction, len(transactions))
	for i, tx := range transactions {
		incomeTransactions[i] = domain.GeneratedTransaction{
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

	breakdown := &transport.RevenueBreakdown{
		TotalAch:     0.0,
		TotalWire:    0.0,
		TotalZelle:   0.0,
		TotalGateway: 0.0,
		TotalOther:   0.0,
	}

	// группировка и суммирование
	for _, tx := range incomeTransactions {
		method := strings.ToLower(tx.Method)
		category := strings.ToLower(tx.Category)
		amount := tx.Amount

		switch {
		case method == "ach_credit":
			breakdown.TotalAch += amount
		case method == "wire":
			breakdown.TotalWire += amount
		case method == "zelle":
			breakdown.TotalZelle += amount
		case strings.Contains(category, "шлюз") ||
			strings.Contains(category, "gateway") ||
			strings.Contains(category, "stripe") ||
			strings.Contains(category, "paypal") ||
			strings.Contains(category, "square"):
			breakdown.TotalGateway += amount
		default:
			breakdown.TotalOther += amount
		}
	}

	// округление до центов
	breakdown.TotalAch = utils.RoundToCents(breakdown.TotalAch)
	breakdown.TotalWire = utils.RoundToCents(breakdown.TotalWire)
	breakdown.TotalZelle = utils.RoundToCents(breakdown.TotalZelle)
	breakdown.TotalGateway = utils.RoundToCents(breakdown.TotalGateway)
	breakdown.TotalOther = utils.RoundToCents(breakdown.TotalOther)

	return breakdown, nil
}

func (s *breakdownService) GetExpensesBreakdown(requestIDStr string) (*transport.ExpensesBreakdown, error) {
	requestID, err := uuid.Parse(requestIDStr)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidRequestID, err)
	}
	transactions, err := s.transactionRepo.GetExpenseTransactionsByRequestID(requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to get expense transactions: %w", err)
	}

	expenseTransactions := make([]domain.GeneratedTransaction, len(transactions))
	for i, tx := range transactions {
		expenseTransactions[i] = domain.GeneratedTransaction{
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

	breakdown := &transport.ExpensesBreakdown{
		ByCard:    0.0,
		ByAccount: 0.0,
	}

	// группировка и суммирование
	for _, tx := range expenseTransactions {
		method := strings.ToLower(tx.Method)
		amount := math.Abs(tx.Amount) // расходы отрицательные, берем модуль

		switch {
		case method == "card":
			breakdown.ByCard += amount
		case method == "ach_debit" ||
			method == "account" ||
			method == "ach_credit" ||
			method == "wire" ||
			method == "electronic payment" ||
			method == "bank_transfer":
			breakdown.ByAccount += amount
		default:
			// Неизвестный метод - относим к account
			breakdown.ByAccount += amount
		}
	}

	// округление до центов
	breakdown.ByCard = utils.RoundToCents(breakdown.ByCard)
	breakdown.ByAccount = utils.RoundToCents(breakdown.ByAccount)
	return breakdown, nil
}
