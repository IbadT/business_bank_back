package breakdownservice

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/IbadT/business_bank_back/services/matematika/internal/domain"
	"github.com/IbadT/business_bank_back/services/matematika/internal/domain/entities"
	"github.com/IbadT/business_bank_back/services/matematika/internal/repository"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/transport"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/utils"
	"github.com/google/uuid"
)

var (
	ErrInvalidRequestID = errors.New("invalid requestId format")
)

type BreakdownService interface {
	GetRevenueBreakdown(requestIDStr string) (*transport.RevenueBreakdown, error)
	GetExpensesBreakdown(requestIDStr string) (*transport.ExpensesBreakdown, error)
	CalculateRevenueBreakdown(transactions []*entities.Transaction) transport.RevenueBreakdown
	CalculateExpensesBreakdown(transactions []*entities.Transaction) transport.ExpensesBreakdown
	CalculateTransactionCounts(transactions []*entities.Transaction) transport.TransactionCounts
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

// CalculateRevenueBreakdown рассчитывает разбивку доходов по методам платежа из транзакций в памяти
func (s *breakdownService) CalculateRevenueBreakdown(transactions []*entities.Transaction) transport.RevenueBreakdown {
	breakdown := transport.RevenueBreakdown{
		TotalAch:     0.0,
		TotalWire:    0.0,
		TotalZelle:   0.0,
		TotalGateway: 0.0,
		TotalOther:   0.0,
	}

	for _, tx := range transactions {
		if !tx.IsIncome() {
			continue
		}

		method := strings.ToLower(tx.Method.String())
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

	// Округление до центов
	breakdown.TotalAch = utils.RoundToCents(breakdown.TotalAch)
	breakdown.TotalWire = utils.RoundToCents(breakdown.TotalWire)
	breakdown.TotalZelle = utils.RoundToCents(breakdown.TotalZelle)
	breakdown.TotalGateway = utils.RoundToCents(breakdown.TotalGateway)
	breakdown.TotalOther = utils.RoundToCents(breakdown.TotalOther)

	return breakdown
}

// CalculateExpensesBreakdown рассчитывает разбивку расходов по методам платежа из транзакций в памяти
func (s *breakdownService) CalculateExpensesBreakdown(transactions []*entities.Transaction) transport.ExpensesBreakdown {
	breakdown := transport.ExpensesBreakdown{
		ByCard:    0.0,
		ByAccount: 0.0,
	}

	for _, tx := range transactions {
		if tx.IsIncome() {
			continue
		}

		method := strings.ToLower(tx.Method.String())
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

	// Округление до центов
	breakdown.ByCard = utils.RoundToCents(breakdown.ByCard)
	breakdown.ByAccount = utils.RoundToCents(breakdown.ByAccount)

	return breakdown
}

// CalculateTransactionCounts рассчитывает количество транзакций по типам и методам [48][49]
func (s *breakdownService) CalculateTransactionCounts(transactions []*entities.Transaction) transport.TransactionCounts {
	counts := transport.TransactionCounts{
		Total: len(transactions),
		Deposits: transport.DepositCounts{
			Total: 0,
			Ach:   0,
			Wire:  0,
			Zelle: 0,
		},
		Withdrawals: transport.WithdrawalCounts{
			Total:       0,
			FromAccount: 0,
			ByCard:      0,
		},
	}

	for _, tx := range transactions {
		method := strings.ToLower(tx.Method.String())

		if tx.IsIncome() {
			// Подсчет депозитов (доходов)
			counts.Deposits.Total++

			switch method {
			case "ach_credit":
				counts.Deposits.Ach++
			case "wire":
				counts.Deposits.Wire++
			case "zelle":
				counts.Deposits.Zelle++
			}
		} else {
			// Подсчет выводов (расходов)
			counts.Withdrawals.Total++

			switch method {
			case "card":
				counts.Withdrawals.ByCard++
			case "ach_debit", "account", "wire", "electronic payment", "bank_transfer":
				counts.Withdrawals.FromAccount++
			default:
				// Неизвестный метод - относим к account
				counts.Withdrawals.FromAccount++
			}
		}
	}

	return counts
}
