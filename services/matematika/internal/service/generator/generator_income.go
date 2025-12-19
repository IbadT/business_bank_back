// internal/service/generator_income.go
package generatorservice

import (
	"github.com/IbadT/business_bank_back/services/matematika/internal/domain/entities"
	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/dto"
)

// generateIncomes генерирует доходы в зависимости от модели бизнеса
func (s *generatorService) generateIncomes(req *dto.GenerateRequest, userID *string) ([]*entities.Transaction, error) {
	switch req.Model {
	case "B2C":
		return s.generateB2CIncomes(req, userID)
	case "B2B":
		return s.generateB2BIncomes(req), nil
	default:
		return nil, ErrInvalidModel
	}
}
