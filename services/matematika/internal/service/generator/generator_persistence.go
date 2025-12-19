// internal/service/generator_persistence.go
package generatorservice

import (
	"fmt"
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/internal/domain"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// saveTransactionsAndUpdateStatus сохраняет транзакции и обновляет статус запроса
func (s *generatorService) saveTransactionsAndUpdateStatus(domainTransactions []*domain.GeneratedTransaction, requestID uuid.UUID) {
	// Сохраняем транзакции в БД
	if err := s.transactionRepo.CreateBatch(domainTransactions); err != nil {
		logrus.Infof("[ERROR] Failed to save transactions to database: %v", err)
		errorMsg := fmt.Sprintf("failed to save transactions to database: %v", err)
		s.generationRequestRepo.UpdateStatus(requestID, "failed", &errorMsg)
		return
	}

	logrus.Infof("[INFO] Saved %d transactions to database for request_id: %s", len(domainTransactions), requestID)

	// Обновляем статус GenerationRequest на "completed"
	completedAt := time.Now()
	if err := s.generationRequestRepo.UpdateCompletedAt(requestID, completedAt); err != nil {
		logrus.Infof("[ERROR] Failed to update generation request status: %v", err)
	}
}
