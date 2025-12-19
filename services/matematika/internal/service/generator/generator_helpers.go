// internal/service/generator_helpers.go
package generatorservice

import (
	"github.com/IbadT/business_bank_back/services/matematika/internal/domain/entities"
	"github.com/IbadT/business_bank_back/services/matematika/internal/models"
	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/dto"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/logger"
	"github.com/google/uuid"
)

// convertCustomDataToJSONB конвертирует CustomData в JSONB формат
func (s *generatorService) convertCustomDataToJSONB(customData *dto.CustomData) models.JSONB {
	op := "service.generator.convertCustomDataToJSONB"
	log := logger.GetLogger().WithOperation(op)
	log.Debug("Converting custom data to JSONB")

	if customData == nil {
		log.Debug("Custom data is nil, returning nil")
		return nil
	}

	result := make(models.JSONB)
	if len(customData.ManualTransactions) > 0 {
		result["manualTransactions"] = customData.ManualTransactions
	}
	if customData.CompanyInfo.OwnerName != "" || customData.CompanyInfo.CompanyName != "" {
		result["companyInfo"] = customData.CompanyInfo
	}
	if len(customData.CustomCustomers) > 0 {
		result["customCustomers"] = customData.CustomCustomers
	}
	if len(customData.CustomContractors) > 0 {
		result["customContractors"] = customData.CustomContractors
	}
	log.WithFields(logger.Fields{"keys_count": len(result)}).Debug("Custom data converted to JSONB")
	return result
}

// updateRequestStatusOnError обновляет статус запроса при ошибке
func (s *generatorService) updateRequestStatusOnError(requestID uuid.UUID, err error) {
	errorMsg := err.Error()
	op := "service.generator.updateRequestStatusOnError"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"request_id": requestID})
	if updateErr := s.generationRequestRepo.UpdateStatus(requestID, "failed", &errorMsg); updateErr != nil {
		log.Warn("Failed to update request status: %v", updateErr)
	}
}

// addManualTransactions добавляет ручные транзакции и возвращает суммы
func (s *generatorService) addManualTransactions(transactions []*entities.Transaction, customData *dto.CustomData) ([]*entities.Transaction, float64, float64) {
	op := "service.generator.addManualTransactions"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"existing_count": len(transactions),
	})
	log.Debug("Adding manual transactions")

	if customData == nil || len(customData.ManualTransactions) == 0 {
		log.Debug("No manual transactions to add")
		return transactions, 0, 0
	}

	manualTransactions := s.convertManualTransactions(customData.ManualTransactions)
	allTransactions := append(transactions, manualTransactions...)

	var manualIncomeAmount, manualExpenseAmount float64
	for _, tx := range manualTransactions {
		if tx.IsIncome() {
			manualIncomeAmount += tx.Amount
		} else {
			manualExpenseAmount += tx.Amount // расходы отрицательные
		}
	}

	log.WithFields(logger.Fields{
		"total_count":        len(allTransactions),
		"manual_count":       len(manualTransactions),
		"manual_income":      manualIncomeAmount,
		"manual_expense":     manualExpenseAmount,
	}).Debug("Manual transactions added")
	return allTransactions, manualIncomeAmount, manualExpenseAmount
}
