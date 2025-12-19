// internal/service/generator_persistence.go
package generatorservice

import (
	"fmt"
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/internal/domain"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/logger"
	"github.com/google/uuid"
)

// saveTransactionsAndUpdateStatus сохраняет транзакции и обновляет статус запроса
func (s *generatorService) saveTransactionsAndUpdateStatus(domainTransactions []*domain.GeneratedTransaction, requestID uuid.UUID) {
	op := "service.generator.saveTransactionsAndUpdateStatus"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"request_id": requestID,
		"count":      len(domainTransactions),
	})
	log.Info("Saving transactions and updating status")

	// Сохраняем транзакции в БД
	if err := s.transactionRepo.CreateBatch(domainTransactions); err != nil {
		log.Error(err, "Failed to save transactions to database")
		errorMsg := fmt.Sprintf("failed to save transactions to database: %v", err)
		s.generationRequestRepo.UpdateStatus(requestID, "failed", &errorMsg)
		return
	}

	log.Success("Transactions saved to database")

	// Обновляем статус GenerationRequest на "completed"
	completedAt := time.Now()
	if err := s.generationRequestRepo.UpdateCompletedAt(requestID, completedAt); err != nil {
		log.Error(err, "Failed to update generation request status")
	} else {
		log.Success("Generation request status updated to completed")
	}
}
