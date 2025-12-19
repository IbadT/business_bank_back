// internal/service/generator_income.go
package generatorservice

import (
	"github.com/IbadT/business_bank_back/services/matematika/internal/domain/entities"
	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/dto"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/helpers"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/logger"
)

// generateIncomes генерирует доходы в зависимости от модели бизнеса
func (s *generatorService) generateIncomes(req *dto.GenerateRequest, userID *string) ([]*entities.Transaction, error) {
	op := "service.generator.generateIncomes"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"model": req.Model})
	log.Debug("Generating incomes")

	switch req.Model {
	case "B2C":
		result, err := s.generateB2CIncomes(req, userID)
		if err != nil {
			log.Error(err, "Failed to generate B2C incomes")
			return nil, err
		}
		log.WithFields(logger.Fields{"count": len(result)}).Debug("B2C incomes generated")
		return result, nil
	case "B2B":
		result := s.generateB2BIncomes(req)
		log.WithFields(logger.Fields{"count": len(result)}).Debug("B2B incomes generated")
		return result, nil
	default:
		log.Error(nil, "Invalid model: %s", req.Model)
		return nil, helpers.ErrInvalidModel
	}
}
