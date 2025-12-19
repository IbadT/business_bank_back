package generatorservice

import (
	"math/rand"

	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/dto"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/logger"
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
	op := "service.generator.amountCalculator.generateDeviation"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"max_deviation": maxDeviation})
	log.Debug("Generating deviation")
	deviation := (rand.Float64()*2 - 1) * maxDeviation // -maxDeviation to +maxDeviation
	log.WithFields(logger.Fields{"deviation": deviation}).Debug("Deviation generated")
	return deviation
}

func (ac *amountCalculator) isFirstMonth(req *dto.GenerateRequest) bool {
	op := "service.generator.amountCalculator.isFirstMonth"
	log := logger.GetLogger().WithOperation(op)
	log.Debug("Checking if first month")
	result := ac.isFirstMonthFlag
	log.WithFields(logger.Fields{"is_first_month": result}).Debug("First month check completed")
	return result
}

func (ac *amountCalculator) getCurrentTurnover() float64 {
	op := "service.generator.amountCalculator.getCurrentTurnover"
	log := logger.GetLogger().WithOperation(op)
	log.Debug("Getting current turnover")
	turnover := ac.firstMonthTurnover
	log.WithFields(logger.Fields{"turnover": turnover}).Debug("Current turnover retrieved")
	return turnover
}

func (ac *amountCalculator) saveLeaseAmount(amount float64) {
	op := "service.generator.amountCalculator.saveLeaseAmount"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"amount": amount})
	log.Debug("Saving lease amount")
	ac.previousLeaseAmount = amount
	ac.isFirstMonthFlag = false
	log.Debug("Lease amount saved")
}

func (ac *amountCalculator) getSavedLeaseAmount() float64 {
	op := "service.generator.amountCalculator.getSavedLeaseAmount"
	log := logger.GetLogger().WithOperation(op)
	log.Debug("Getting saved lease amount")
	amount := ac.previousLeaseAmount
	log.WithFields(logger.Fields{"amount": amount}).Debug("Saved lease amount retrieved")
	return amount
}

