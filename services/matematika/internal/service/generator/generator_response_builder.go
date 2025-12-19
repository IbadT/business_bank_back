// internal/service/generator_response_builder.go
package generatorservice

import (
	"math"
	"sync"
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/internal/domain/entities"
	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/dto"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/logger"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/transport"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/utils"
	"github.com/google/uuid"
)

// buildResponse формирует ответ с транзакциями [44-54]
func (s *generatorService) buildResponse(transactions []*entities.Transaction, req *dto.GenerateRequest, requestID uuid.UUID, userID uuid.UUID) *dto.GenerateResponse {
	// Рассчитываем финансовую сводку
	finalBalance := req.InitialBalance
	totalRevenue := 0.0
	totalExpenses := 0.0

	for _, tx := range transactions {
		if tx.IsIncome() {
			totalRevenue += tx.Amount
		} else {
			totalExpenses += tx.Amount
		}
	}

	if len(transactions) > 0 {
		finalBalance = transactions[len(transactions)-1].BalanceAfter
	}

	// [42][45] Рассчитываем netProfit согласно ТЗ: сумма всех доходов минус сумма всех расходов
	// totalExpenses отрицательное, поэтому netProfit = totalRevenue + totalExpenses
	netProfit := totalRevenue + totalExpenses

	// [5][6] Проверка соответствия ТЗ: суммарные доходы должны быть = 100% оборота
	// Учитываем ручные доходы, если они есть
	expectedRevenue := req.Turnover
	manualIncomeAmount := 0.0
	for _, tx := range transactions {
		if tx.IsIncome() && tx.IsManualTransaction() {
			manualIncomeAmount += tx.Amount
		}
	}
	if manualIncomeAmount > 0 {
		expectedRevenue = req.Turnover + manualIncomeAmount
	}

	// [1][5] Проверка соответствия ТЗ: прибыль должна соответствовать desiredProfitPercent
	// Учитываем ручные транзакции
	expectedProfit := req.Turnover * (req.DesiredProfitPercent / 100)
	manualExpenseAmount := 0.0
	for _, tx := range transactions {
		if tx.IsExpense() && tx.IsManualTransaction() {
			manualExpenseAmount += tx.Amount // отрицательное
		}
	}
	if manualIncomeAmount > 0 || manualExpenseAmount < 0 {
		expectedProfit = expectedProfit + manualIncomeAmount + manualExpenseAmount
	}

	// Проверка и логирование отклонений (допустимая погрешность округления 0.05)
	revenueError := math.Abs(expectedRevenue - totalRevenue)
	profitError := math.Abs(expectedProfit - netProfit)

	op := "service.generator.buildResponse"
	log := logger.GetLogger().WithOperation(op)
	if revenueError > 0.05 {
		log.Warn("Revenue mismatch: expected=%.2f (turnover=%.2f + manualIncome=%.2f), actual=%.2f, diff=%.2f",
			expectedRevenue, req.Turnover, manualIncomeAmount, totalRevenue, revenueError)
	}
	if profitError > 0.05 {
		log.Warn("Profit mismatch: expected=%.2f (%.2f%% of %.2f + manual adjustments), actual=%.2f, diff=%.2f",
			expectedProfit, req.DesiredProfitPercent, req.Turnover, netProfit, profitError)
	}

	summary := dto.FinancialSummary{
		InitialBalance: utils.RoundToCents(req.InitialBalance),
		FinalBalance:   utils.RoundToCents(finalBalance),
		TotalRevenue:   utils.RoundToCents(totalRevenue),
		TotalExpenses:  utils.RoundToCents(totalExpenses),
		NetProfit:      utils.RoundToCents(netProfit),
	}

	// Рассчитываем ежедневные балансы
	dailyBalances := s.calculateDailyBalances(transactions, req.InitialBalance, req.Year, req.Month)

	// Конвертируем entities.Transaction в dto.Transaction
	dtoTransactions := make([]dto.Transaction, len(transactions))
	for i, tx := range transactions {
		dtoTransactions[i] = dto.Transaction{
			TransactionID:      tx.ID,
			TransactionDate:    tx.TransactionDate,
			PostingDate:        tx.PostingDate.Format("2006-01-02"),
			Type:               tx.Type.String(),
			Category:           tx.Category,
			Method:             tx.Method.String(),
			Amount:             tx.Amount,
			BalanceAfter:       tx.BalanceAfter,
			IsManual:           tx.IsManualTransaction(),
			FixAsFirst:         tx.FixAsFirst,
			CalculationDetails: tx.CalculationDetails,
		}
	}

	// Получаем associatedCard из таблицы users [51][52]
	// Номер карты должен быть задан пользователем через API /api/user/associated-card
	associatedCard := s.getAssociatedCard(userID)

	// Формируем forwardingInfo
	forwardingInfo := dto.ForwardingInfo{
		AssociatedCard:    associatedCard,
		OwnerName:         "",
		CompanyName:       "",
		CustomCustomers:   []string{},
		CustomContractors: []dto.CustomContractor{},
	}

	if req.CustomData != nil {
		forwardingInfo.OwnerName = req.CustomData.CompanyInfo.OwnerName
		forwardingInfo.CompanyName = req.CustomData.CompanyInfo.CompanyName
		forwardingInfo.CustomCustomers = req.CustomData.CustomCustomers
		forwardingInfo.CustomContractors = req.CustomData.CustomContractors
	}

	// Рассчитываем разбивку по методам платежа параллельно
	var revenueBreakdown transport.RevenueBreakdown
	var expensesBreakdown transport.ExpensesBreakdown
	var transactionCounts transport.TransactionCounts

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		revenueBreakdown = s.breakdownService.CalculateRevenueBreakdown(transactions)
	}()

	go func() {
		defer wg.Done()
		expensesBreakdown = s.breakdownService.CalculateExpensesBreakdown(transactions)
	}()

	go func() {
		defer wg.Done()
		transactionCounts = s.breakdownService.CalculateTransactionCounts(transactions)
	}()

	wg.Wait()

	return &dto.GenerateResponse{
		RequestID:            requestID.String(),
		Transactions:         dtoTransactions,
		FinancialSummary:     summary,
		DailyClosingBalances: dailyBalances,
		ForwardingInfo:       forwardingInfo,
		RevenueBreakdown:     revenueBreakdown,
		ExpensesBreakdown:    expensesBreakdown,
		TransactionCounts:    transactionCounts,
	}
}

// calculateDailyBalances рассчитывает ежедневные балансы [50]
// Генерирует балансы для всех дней месяца, а не только для дней с транзакциями
func (s *generatorService) calculateDailyBalances(transactions []*entities.Transaction, initialBalance float64, year, month int) []dto.DailyBalance {
	op := "service.generator.calculateDailyBalances"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"transactions_count": len(transactions),
		"initial_balance":    initialBalance,
		"year":                year,
		"month":               month,
	})
	log.Debug("Calculating daily balances")

	// Собираем балансы по датам из транзакций
	// Для каждой даты берем баланс после последней транзакции этого дня
	balancesByDate := make(map[string]float64)

	// Начальный баланс на первый день месяца
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	balancesByDate[firstDay.Format("2006-01-02")] = initialBalance

	// Обновляем балансы после каждой транзакции
	// Если в один день несколько транзакций, последняя перезапишет баланс
	for _, tx := range transactions {
		dateKey := tx.PostingDate.Format("2006-01-02")
		balancesByDate[dateKey] = tx.BalanceAfter
	}

	// Получаем последний день месяца
	lastDayOfMonth := time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.UTC)
	daysInMonth := lastDayOfMonth.Day()

	// Генерируем балансы для всех дней месяца
	var dailyBalances []dto.DailyBalance
	currentBalance := initialBalance

	for day := 1; day <= daysInMonth; day++ {
		currentDate := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		dateKey := currentDate.Format("2006-01-02")

		// Если есть транзакции в этот день, используем баланс после последней транзакции
		if balance, exists := balancesByDate[dateKey]; exists {
			currentBalance = balance
		}
		// Если транзакций нет, используем баланс предыдущего дня (currentBalance уже содержит его)

		dailyBalances = append(dailyBalances, dto.DailyBalance{
			Date:    dateKey,
			Balance: utils.RoundToCents(currentBalance),
		})
	}

	log.WithFields(logger.Fields{"daily_balances_count": len(dailyBalances)}).Debug("Daily balances calculated")
	return dailyBalances
}
