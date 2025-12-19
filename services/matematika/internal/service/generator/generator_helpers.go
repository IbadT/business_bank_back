// internal/service/generator_helpers.go
package generatorservice

import (
	"github.com/IbadT/business_bank_back/services/matematika/internal/domain/entities"
	"github.com/IbadT/business_bank_back/services/matematika/internal/models"
	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/dto"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// convertCustomDataToJSONB конвертирует CustomData в JSONB формат
func (s *generatorService) convertCustomDataToJSONB(customData *dto.CustomData) models.JSONB {
	if customData == nil {
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
	return result
}

// updateRequestStatusOnError обновляет статус запроса при ошибке
func (s *generatorService) updateRequestStatusOnError(requestID uuid.UUID, err error) {
	errorMsg := err.Error()
	if updateErr := s.generationRequestRepo.UpdateStatus(requestID, "failed", &errorMsg); updateErr != nil {
		logrus.Infof("[WARN] Failed to update request status: %v", updateErr)
	}
}

// addManualTransactions добавляет ручные транзакции и возвращает суммы
func (s *generatorService) addManualTransactions(transactions []*entities.Transaction, customData *dto.CustomData) ([]*entities.Transaction, float64, float64) {
	if customData == nil || len(customData.ManualTransactions) == 0 {
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

	return allTransactions, manualIncomeAmount, manualExpenseAmount
}
