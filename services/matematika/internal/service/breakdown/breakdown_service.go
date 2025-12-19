package breakdownservice

import (
	"fmt"
	"math"
	"strings"

	"github.com/IbadT/business_bank_back/services/matematika/internal/domain"
	"github.com/IbadT/business_bank_back/services/matematika/internal/domain/entities"
	"github.com/IbadT/business_bank_back/services/matematika/internal/repository"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/helpers"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/logger"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/transport"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/utils"
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
	op := "service.breakdown.getRevenueBreakdown"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"request_id": requestIDStr})
	log.Info("Getting revenue breakdown")

	requestID, err := helpers.ParseUUID(requestIDStr)
	if err != nil {
		log.Error(err, "Invalid requestID format")
		return nil, err
	}
	transactions, err := s.transactionRepo.GetIncomeTransactionsByRequestID(requestID)
	if err != nil {
		log.Error(err, "Failed to get income transactions from repository")
		return nil, fmt.Errorf("%w: %v", helpers.ErrFailedToGetIncomeTransactions, err)
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
		case method == helpers.PaymentMethodACHCreditLowerStr:
			breakdown.TotalAch += amount
		case method == helpers.PaymentMethodWireStr:
			breakdown.TotalWire += amount
		case method == helpers.PaymentMethodZelleStr:
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

	log.WithFields(logger.Fields{
		"total_ach":     breakdown.TotalAch,
		"total_wire":    breakdown.TotalWire,
		"total_zelle":   breakdown.TotalZelle,
		"total_gateway": breakdown.TotalGateway,
		"total_other":   breakdown.TotalOther,
	}).Success("Revenue breakdown calculated")

	return breakdown, nil
}

func (s *breakdownService) GetExpensesBreakdown(requestIDStr string) (*transport.ExpensesBreakdown, error) {
	op := "service.breakdown.getExpensesBreakdown"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"request_id": requestIDStr})
	log.Info("Getting expenses breakdown")

	requestID, err := helpers.ParseUUID(requestIDStr)
	if err != nil {
		log.Error(err, "Invalid requestID format")
		return nil, err
	}
	transactions, err := s.transactionRepo.GetExpenseTransactionsByRequestID(requestID)
	if err != nil {
		log.Error(err, "Failed to get expense transactions from repository")
		return nil, fmt.Errorf("%w: %v", helpers.ErrFailedToGetExpenseTransactions, err)
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
		case method == helpers.PaymentMethodCardStr:
			breakdown.ByCard += amount
		case method == helpers.PaymentMethodACHDebitLowerStr ||
			method == helpers.PaymentMethodAccountStr ||
			method == helpers.PaymentMethodACHCreditLowerStr ||
			method == helpers.PaymentMethodWireStr ||
			method == helpers.PaymentMethodElectronicPaymentLowerStr ||
			method == helpers.PaymentMethodBankTransferStr:
			breakdown.ByAccount += amount
		default:
			// Неизвестный метод - относим к account
			breakdown.ByAccount += amount
		}
	}

	// округление до центов
	breakdown.ByCard = utils.RoundToCents(breakdown.ByCard)
	breakdown.ByAccount = utils.RoundToCents(breakdown.ByAccount)
	
	log.WithFields(logger.Fields{
		"by_card":    breakdown.ByCard,
		"by_account": breakdown.ByAccount,
	}).Success("Expenses breakdown calculated")
	
	return breakdown, nil
}

// CalculateRevenueBreakdown рассчитывает разбивку доходов по методам платежа из транзакций в памяти
func (s *breakdownService) CalculateRevenueBreakdown(transactions []*entities.Transaction) transport.RevenueBreakdown {
	op := "service.breakdown.calculateRevenueBreakdown"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"transactions_count": len(transactions)})
	log.Debug("Calculating revenue breakdown from transactions")

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
		case method == helpers.PaymentMethodACHCreditLowerStr:
			breakdown.TotalAch += amount
		case method == helpers.PaymentMethodWireStr:
			breakdown.TotalWire += amount
		case method == helpers.PaymentMethodZelleStr:
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

	log.WithFields(logger.Fields{
		"total_ach":     breakdown.TotalAch,
		"total_wire":    breakdown.TotalWire,
		"total_zelle":   breakdown.TotalZelle,
		"total_gateway": breakdown.TotalGateway,
		"total_other":   breakdown.TotalOther,
	}).Debug("Revenue breakdown calculated")

	return breakdown
}

// CalculateExpensesBreakdown рассчитывает разбивку расходов по методам платежа из транзакций в памяти
func (s *breakdownService) CalculateExpensesBreakdown(transactions []*entities.Transaction) transport.ExpensesBreakdown {
	op := "service.breakdown.calculateExpensesBreakdown"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"transactions_count": len(transactions)})
	log.Debug("Calculating expenses breakdown from transactions")

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
		case method == helpers.PaymentMethodCardStr:
			breakdown.ByCard += amount
		case method == helpers.PaymentMethodACHDebitLowerStr ||
			method == helpers.PaymentMethodAccountStr ||
			method == helpers.PaymentMethodACHCreditLowerStr ||
			method == helpers.PaymentMethodWireStr ||
			method == helpers.PaymentMethodElectronicPaymentLowerStr ||
			method == helpers.PaymentMethodBankTransferStr:
			breakdown.ByAccount += amount
		default:
			// Неизвестный метод - относим к account
			breakdown.ByAccount += amount
		}
	}

	// Округление до центов
	breakdown.ByCard = utils.RoundToCents(breakdown.ByCard)
	breakdown.ByAccount = utils.RoundToCents(breakdown.ByAccount)

	log.WithFields(logger.Fields{
		"by_card":    breakdown.ByCard,
		"by_account": breakdown.ByAccount,
	}).Debug("Expenses breakdown calculated")

	return breakdown
}

// CalculateTransactionCounts рассчитывает количество транзакций по типам и методам [48][49]
func (s *breakdownService) CalculateTransactionCounts(transactions []*entities.Transaction) transport.TransactionCounts {
	op := "service.breakdown.calculateTransactionCounts"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"transactions_count": len(transactions)})
	log.Debug("Calculating transaction counts")

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
			case helpers.PaymentMethodACHCreditLowerStr:
				counts.Deposits.Ach++
			case helpers.PaymentMethodWireStr:
				counts.Deposits.Wire++
			case helpers.PaymentMethodZelleStr:
				counts.Deposits.Zelle++
			}
		} else {
			// Подсчет выводов (расходов)
			counts.Withdrawals.Total++

			switch method {
			case helpers.PaymentMethodCardStr:
				counts.Withdrawals.ByCard++
			case helpers.PaymentMethodACHDebitLowerStr, helpers.PaymentMethodAccountStr, helpers.PaymentMethodWireStr, helpers.PaymentMethodElectronicPaymentLowerStr, helpers.PaymentMethodBankTransferStr:
				counts.Withdrawals.FromAccount++
			default:
				// Неизвестный метод - относим к account
				counts.Withdrawals.FromAccount++
			}
		}
	}

	log.WithFields(logger.Fields{
		"total":              counts.Total,
		"deposits_total":     counts.Deposits.Total,
		"deposits_ach":       counts.Deposits.Ach,
		"deposits_wire":      counts.Deposits.Wire,
		"deposits_zelle":     counts.Deposits.Zelle,
		"withdrawals_total":  counts.Withdrawals.Total,
		"withdrawals_account": counts.Withdrawals.FromAccount,
		"withdrawals_card":   counts.Withdrawals.ByCard,
	}).Debug("Transaction counts calculated")

	return counts
}
