package generatorservice

import (
	"math/rand"

	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/dto"
)

type amountCalculator struct {
	previousLeaseAmount float64
	firstMonthTurnover  float64
	isFirstMonthFlag    bool
}

func newAmountCalculator() *amountCalculator {
	return &amountCalculator{
		isFirstMonthFlag: true,
	}
}

func (ac *amountCalculator) generateDeviation(maxDeviation float64) float64 {
	return (rand.Float64()*2 - 1) * maxDeviation // -maxDeviation to +maxDeviation
}

func (ac *amountCalculator) isFirstMonth(req *dto.GenerateRequest) bool {
	// Упрощенная проверка - можно доработать с учетом истории
	return ac.isFirstMonthFlag
}

func (ac *amountCalculator) getCurrentTurnover() float64 {
	return ac.firstMonthTurnover
}

func (ac *amountCalculator) saveLeaseAmount(amount float64) {
	ac.previousLeaseAmount = amount
	ac.isFirstMonthFlag = false
}

func (ac *amountCalculator) getSavedLeaseAmount() float64 {
	return ac.previousLeaseAmount
}

