// internal/transport/http/mapper/mapper.go
package mapper

import (
	"github.com/IbadT/business_bank_back/services/matematika/pkg/utils"

	"github.com/IbadT/business_bank_back/services/matematika/internal/domain/entities"
	"github.com/IbadT/business_bank_back/services/matematika/internal/domain/value_objects"
	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/dto"
)

// EntityToDTO маппит доменную сущность в DTO
func EntityToDTO(entity *entities.Transaction) *dto.Transaction {
	return &dto.Transaction{
		TransactionID:     entity.ID,
		TransactionDate:  entity.TransactionDate,
		PostingDate:      entity.PostingDate.Format("2006-01-02"),
		Type:             entity.Type.String(),
		Category:         entity.Category,
		Method:           entity.Method.String(),
		Amount:           utils.RoundToCents(entity.Amount),
		BalanceAfter:     utils.RoundToCents(entity.BalanceAfter),
		IsManual:         entity.IsManual,
		CalculationDetails: entity.CalculationDetails,
	}
}

// DTOToEntity маппит DTO в доменную сущность (для ручных транзакций)
func DTOToEntity(dtoTx dto.ManualTransaction, id string) (*entities.Transaction, error) {
	transactionType, err := value_objects.NewTransactionType(dtoTx.Type)
	if err != nil {
		return nil, err
	}
	
	paymentMethod, err := value_objects.NewPaymentMethod(dtoTx.Method)
	if err != nil {
		return nil, err
	}
	
	transaction := entities.NewTransaction(
		id,
		dtoTx.TransactionDate,
		dtoTx.TransactionDate,
		transactionType,
		dtoTx.Category,
		paymentMethod,
		utils.RoundToCents(dtoTx.Amount),
	)
	transaction.SetManual(true)
	
	return transaction, nil
}

// EntitiesToResponse маппит доменные сущности в полный ответ
func EntitiesToResponse(
	transactions []*entities.Transaction,
	financialSummary *dto.FinancialSummary,
	dailyBalances []*dto.DailyBalance,
	forwardingInfo *dto.ForwardingInfo,
) *dto.GenerateResponse {
	var dtoTransactions []dto.Transaction
	for _, tx := range transactions {
		dtoTransactions = append(dtoTransactions, *EntityToDTO(tx))
	}
	
	return &dto.GenerateResponse{
		Transactions:        dtoTransactions,
		FinancialSummary:    *financialSummary,
		DailyClosingBalances: DailyBalancesToDTO(dailyBalances),
		ForwardingInfo:      *forwardingInfo,
	}
}

// FinancialSummaryToDTO конвертирует FinancialSummary
func FinancialSummaryToDTO(summary *dto.FinancialSummary) dto.FinancialSummary {
	if summary == nil {
		return dto.FinancialSummary{}
	}
	return *summary
}

// DailyBalancesToDTO конвертирует массив DailyBalance
func DailyBalancesToDTO(balances []*dto.DailyBalance) []dto.DailyBalance {
	if balances == nil {
		return []dto.DailyBalance{}
	}
	result := make([]dto.DailyBalance, len(balances))
	for i, b := range balances {
		if b != nil {
			result[i] = *b
		}
	}
	return result
}

// ForwardingInfoToDTO конвертирует ForwardingInfo
func ForwardingInfoToDTO(info *dto.ForwardingInfo) dto.ForwardingInfo {
	if info == nil {
		return dto.ForwardingInfo{}
	}
	return *info
}