package validator

import (
	"fmt"
	"math/rand"

	"github.com/IbadT/business_bank_back/services/matematika/internal/domain/entities"
	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/dto"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/helpers"
	"github.com/sirupsen/logrus"
)

// ValidateGenerateRequest валидирует запрос на генерацию транзакций
// [1] По умолчанию чистая прибыль компании около 6–9% от оборота
// Если desiredProfitPercent не задан (0), устанавливаем случайное значение в диапазоне 6-9%
func ValidateGenerateRequest(req *dto.GenerateRequest) error {
	if req.Turnover <= 0 {
		return helpers.ErrTurnoverMustBeGreaterThanZero
	}

	// [1] По умолчанию чистая прибыль компании около 6–9% от оборота
	// Если desiredProfitPercent не задан (0), устанавливаем случайное значение в диапазоне 6-9%
	if req.DesiredProfitPercent == 0 {
		// Генерируем случайное значение в диапазоне 6-9%
		req.DesiredProfitPercent = 6.0 + rand.Float64()*(9.0-6.0)
		logrus.Infof("[INFO] DesiredProfitPercent not set, using default range 6-9%%: %.2f%%", req.DesiredProfitPercent)
	}

	if req.DesiredProfitPercent < 0 || req.DesiredProfitPercent > 100 {
		return helpers.ErrDesiredProfitPercentInvalid
	}
	if req.Model != "B2C" && req.Model != "B2B" {
		return helpers.ErrModelMustBeB2COrB2B
	}
	if req.InitialBalance < 0 {
		return helpers.ErrInitialBalanceCannotBeNegative
	}
	return nil
}

// ValidateTransactionCounts проверяет количество транзакций согласно требованиям README [2][3][4]
// [2] Входящие: ~4 (редко 5) для B2C или 10–20 для B2B
// [3] Исходящие: ~45 ± 10 операций (35-55)
// [4] Общее число: 39–75
func ValidateTransactionCounts(transactions []*entities.Transaction, model string) error {
	var incomeCount, expenseCount int

	for _, tx := range transactions {
		if tx.IsIncome() {
			incomeCount++
		} else {
			expenseCount++
		}
	}

	totalCount := len(transactions)

	// [2] Валидация входящих транзакций
	if model == "B2C" {
		if incomeCount < 4 || incomeCount > 5 {
			return fmt.Errorf("%w, got %d", helpers.ErrB2CIncomeTransactionsCountInvalid, incomeCount)
		}
	} else if model == "B2B" {
		if incomeCount < 10 || incomeCount > 20 {
			return fmt.Errorf("%w, got %d", helpers.ErrB2BIncomeTransactionsCountInvalid, incomeCount)
		}
	}

	// [3] Валидация исходящих транзакций (35-55)
	if expenseCount < 35 || expenseCount > 55 {
		return fmt.Errorf("%w, got %d", helpers.ErrExpenseTransactionsCountInvalid, expenseCount)
	}

	// [4] Валидация общего количества транзакций (39-75)
	if totalCount < 39 || totalCount > 75 {
		return fmt.Errorf("%w, got %d (income: %d, expense: %d)", helpers.ErrTotalTransactionsCountInvalid, totalCount, incomeCount, expenseCount)
	}

	logrus.Infof("[INFO] Transaction counts validated: total=%d, income=%d, expense=%d (model=%s)", totalCount, incomeCount, expenseCount, model)
	return nil
}

// CheckNegativeBalance проверяет отрицательный баланс [43]
func CheckNegativeBalance(transactions []*entities.Transaction) error {
	for _, tx := range transactions {
		if tx.BalanceAfter < 0 {
			return helpers.ErrNegativeBalance
		}
	}
	return nil
}
