// internal/service/generator_balance.go
package generatorservice

import (
	"fmt"
	"strings"

	"github.com/IbadT/business_bank_back/services/matematika/internal/domain/entities"
	balanceservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/balance"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/validator"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// calculateAndAdjustBalances рассчитывает балансы и корректирует их при необходимости
func (s *generatorService) calculateAndAdjustBalances(
	transactions []*entities.Transaction,
	initialBalance float64,
	year, month int,
	requestID uuid.UUID,
) ([]*entities.Transaction, error) {
	// Расчет балансов
	transactionsWithBalance, err := s.balanceAdjustmentService.CalculateBalances(transactions, initialBalance)
	if err != nil {
		// Если есть ошибка недостатка баланса, пытаемся скорректировать
		if strings.Contains(err.Error(), "insufficient balance") {
			transactionsWithBalance, err = s.adjustBalancesWithStrategy(
				transactions, initialBalance, year, month, requestID, balanceservice.StrategyPostpone, "adjustment")
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	// [43] Проверка отрицательного баланса (после корректировки)
	if err := validator.CheckNegativeBalance(transactionsWithBalance); err != nil {
		// Если после корректировки все еще есть отрицательный баланс, пробуем уменьшить суммы
		transactionsWithBalance, err = s.adjustBalancesWithStrategy(
			transactionsWithBalance, initialBalance, year, month, requestID, balanceservice.StrategyReduce, "reduction")
		if err != nil {
			return nil, err
		}

		// Финальная проверка
		if err := validator.CheckNegativeBalance(transactionsWithBalance); err != nil {
			s.updateRequestStatusOnError(requestID, err)
			return nil, fmt.Errorf("negative balance still exists after all adjustments: %w", err)
		}
	}

	return transactionsWithBalance, nil
}

// adjustBalancesWithStrategy корректирует балансы используя указанную стратегию
func (s *generatorService) adjustBalancesWithStrategy(
	transactions []*entities.Transaction,
	initialBalance float64,
	year, month int,
	requestID uuid.UUID,
	strategy balanceservice.BalanceHandlingStrategy,
	operationType string,
) ([]*entities.Transaction, error) {
	adjustedTransactions, adjustments, adjustErr := s.balanceAdjustmentService.AdjustTransactionsForBalance(
		transactions,
		initialBalance,
		strategy,
		s.dateCalculator,
		year,
		month,
	)
	if adjustErr != nil {
		errorMsg := fmt.Sprintf("failed to adjust transactions by %s: %v", operationType, adjustErr)
		s.updateRequestStatusOnError(requestID, fmt.Errorf(errorMsg))
		return nil, fmt.Errorf("failed to adjust transactions by %s: %w", operationType, adjustErr)
	}

	// Пересчитываем балансы после корректировки
	transactionsWithBalance, err := s.balanceAdjustmentService.RecalculateBalances(adjustedTransactions, initialBalance)
	if err != nil {
		errorMsg := fmt.Sprintf("failed to recalculate balances after %s: %v", operationType, err)
		s.updateRequestStatusOnError(requestID, fmt.Errorf(errorMsg))
		return nil, fmt.Errorf("failed to recalculate balances after %s: %w", operationType, err)
	}

	// Логируем корректировки
	if len(adjustments) > 0 {
		logrus.Infof("[INFO] Applied %d balance adjustments by %s", len(adjustments), operationType)
	}

	return transactionsWithBalance, nil
}
