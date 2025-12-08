package mocks

import (
	"fmt"

	helpers "github.com/IbadT/business_bank_back/services/matematika/pkg/helpers"
)

// GenerateStatementMocks - возвращает валидный запрос для тестирования
func GenerateStatementMocks() helpers.StatementStateRequest {
	return helpers.StatementStateRequest{
		CompanyInfo: helpers.CompanyInfo{
			CompanyName:    "Test Company",
			OwnerName:      "Test Owner",
			AccountNumber:  "1234567890",
			AssociatedCard: "1234567890",
			Model:          "B2B",
			State:          "CA",
			Industry:       "Test Industry",
		},
		Financials: helpers.Financials{
			StartBalance:        100000.0,
			Turnover:            100000.0,
			ProfitPercent:       7.5,
			TargetProfit:        7500.0,
			Months:              3,
			StartMonth:          "2024-12",  // Не в будущем для валидации
			Periods:             []string{}, // Пустой, используем months + startMonth
			OperationMultiplier: 1.0,
		},
		CustomData: helpers.CustomData{
			ManualIncomes: []helpers.ManualIncome{
				{
					Date:        "2024-12-01", // Дата в пределах StartMonth
					Amount:      10000.0,
					Category:    "Test Category",
					Description: "Test Description",
				},
			},
			ManualExpenses: []helpers.ManualExpense{
				{
					Date:        "2024-12-15", // Дата в пределах StartMonth
					Amount:      10000.0,
					Category:    "Test Category",
					Description: "Test Description",
				},
			},
			CustomCustomers: []string{
				"Test Customer 1",
				"Test Customer 2",
			},
			CustomContractors: []helpers.CustomContractor{
				{
					TransactionType: "Test Transaction Type",
					Name:            "Test Name",
				},
			},
			DisableCategories: []string{
				"Test Category 1",
				"Test Category 2",
			},
		},
	}
}

// ================================================
// Негативные тестовые случаи для валидации CompanyInfo
// ================================================

// MockMissingCompanyName - запрос без companyName
func MockMissingCompanyName() helpers.StatementStateRequest {
	req := GenerateStatementMocks()
	req.CompanyInfo.CompanyName = ""
	return req
}

// MockMissingOwnerName - запрос без ownerName
func MockMissingOwnerName() helpers.StatementStateRequest {
	req := GenerateStatementMocks()
	req.CompanyInfo.OwnerName = ""
	return req
}

// MockMissingAccountNumber - запрос без accountNumber
func MockMissingAccountNumber() helpers.StatementStateRequest {
	req := GenerateStatementMocks()
	req.CompanyInfo.AccountNumber = ""
	return req
}

// MockMissingAssociatedCard - запрос без associatedCard
func MockMissingAssociatedCard() helpers.StatementStateRequest {
	req := GenerateStatementMocks()
	req.CompanyInfo.AssociatedCard = ""
	return req
}

// MockMissingModel - запрос без model
func MockMissingModel() helpers.StatementStateRequest {
	req := GenerateStatementMocks()
	req.CompanyInfo.Model = ""
	return req
}

// MockMissingState - запрос без state
func MockMissingState() helpers.StatementStateRequest {
	req := GenerateStatementMocks()
	req.CompanyInfo.State = ""
	return req
}

// MockMissingIndustry - запрос без industry
func MockMissingIndustry() helpers.StatementStateRequest {
	req := GenerateStatementMocks()
	req.CompanyInfo.Industry = ""
	return req
}

// ================================================
// Негативные тестовые случаи для валидации Financials
// ================================================

// MockNegativeStartBalance - отрицательный startBalance
func MockNegativeStartBalance() helpers.StatementStateRequest {
	req := GenerateStatementMocks()
	req.Financials.StartBalance = -1000.0
	return req
}

// MockZeroTurnover - нулевой turnover
func MockZeroTurnover() helpers.StatementStateRequest {
	req := GenerateStatementMocks()
	req.Financials.Turnover = 0
	return req
}

// MockNegativeTurnover - отрицательный turnover
func MockNegativeTurnover() helpers.StatementStateRequest {
	req := GenerateStatementMocks()
	req.Financials.Turnover = -1000.0
	return req
}

// MockInvalidProfitPercent - profitPercent > 50%
func MockInvalidProfitPercent() helpers.StatementStateRequest {
	req := GenerateStatementMocks()
	req.Financials.ProfitPercent = 60.0
	return req
}

// MockNegativeProfitPercent - отрицательный profitPercent
func MockNegativeProfitPercent() helpers.StatementStateRequest {
	req := GenerateStatementMocks()
	req.Financials.ProfitPercent = -5.0
	return req
}

// MockTargetProfitExceedsTurnover - targetProfit >= turnover
func MockTargetProfitExceedsTurnover() helpers.StatementStateRequest {
	req := GenerateStatementMocks()
	req.Financials.TargetProfit = 150000.0 // Больше turnover
	return req
}

// MockNegativeTargetProfit - отрицательный targetProfit
func MockNegativeTargetProfit() helpers.StatementStateRequest {
	req := GenerateStatementMocks()
	req.Financials.TargetProfit = -1000.0
	return req
}

// MockInvalidMonths - months > 36
func MockInvalidMonths() helpers.StatementStateRequest {
	req := GenerateStatementMocks()
	req.Financials.Months = 50
	return req
}

// MockZeroMonths - months = 0
func MockZeroMonths() helpers.StatementStateRequest {
	req := GenerateStatementMocks()
	req.Financials.Months = 0
	return req
}

// MockFutureMonth - месяц в будущем
func MockFutureMonth() helpers.StatementStateRequest {
	req := GenerateStatementMocks()
	req.Financials.StartMonth = "2030-01"
	return req
}

// MockInvalidMonthFormat - неправильный формат месяца
func MockInvalidMonthFormat() helpers.StatementStateRequest {
	req := GenerateStatementMocks()
	req.Financials.StartMonth = "2024/12" // Неправильный формат
	return req
}

// MockEmptyMonth - пустой месяц
func MockEmptyMonth() helpers.StatementStateRequest {
	req := GenerateStatementMocks()
	req.Financials.StartMonth = ""
	return req
}

// MockInvalidOperationMultiplier - operationMultiplier > 5
func MockInvalidOperationMultiplier() helpers.StatementStateRequest {
	req := GenerateStatementMocks()
	req.Financials.OperationMultiplier = 10.0
	return req
}

// MockZeroOperationMultiplier - operationMultiplier = 0
func MockZeroOperationMultiplier() helpers.StatementStateRequest {
	req := GenerateStatementMocks()
	req.Financials.OperationMultiplier = 0
	return req
}

// MockNegativeOperationMultiplier - отрицательный operationMultiplier
func MockNegativeOperationMultiplier() helpers.StatementStateRequest {
	req := GenerateStatementMocks()
	req.Financials.OperationMultiplier = -1.0
	return req
}

// MockBothPeriodsAndMonths - указаны и periods и months одновременно
func MockBothPeriodsAndMonths() helpers.StatementStateRequest {
	req := GenerateStatementMocks()
	req.Financials.Periods = []string{"2024-12", "2025-01"}
	req.Financials.Months = 3
	req.Financials.StartMonth = "2024-12"
	return req
}

// MockNeitherPeriodsNorMonths - не указаны ни periods ни months
func MockNeitherPeriodsNorMonths() helpers.StatementStateRequest {
	req := GenerateStatementMocks()
	req.Financials.Periods = []string{}
	req.Financials.Months = 0
	req.Financials.StartMonth = ""
	return req
}

// MockInsufficientBalance - недостаточный баланс
func MockInsufficientBalance() helpers.StatementStateRequest {
	req := GenerateStatementMocks()
	req.Financials.StartBalance = 1000.0 // Слишком мало для расходов
	req.Financials.Turnover = 100000.0
	req.Financials.ProfitPercent = 7.5
	return req
}

// MockInvalidPeriodsFormat - неправильный формат в periods
func MockInvalidPeriodsFormat() helpers.StatementStateRequest {
	req := GenerateStatementMocks()
	req.Financials.Periods = []string{"2024/12", "2025-01"} // Первый период неправильного формата
	req.Financials.Months = 0
	req.Financials.StartMonth = ""
	return req
}

// MockFuturePeriods - periods содержат будущие месяцы
func MockFuturePeriods() helpers.StatementStateRequest {
	req := GenerateStatementMocks()
	req.Financials.Periods = []string{"2030-01", "2030-02"}
	req.Financials.Months = 0
	req.Financials.StartMonth = ""
	return req
}

// ================================================
// Негативные тестовые случаи для валидации CustomData
// ================================================

// MockInvalidManualIncomeDate - неправильная дата в manualIncome
func MockInvalidManualIncomeDate() helpers.StatementStateRequest {
	req := GenerateStatementMocks()
	req.CustomData.ManualIncomes = []helpers.ManualIncome{
		{
			Date:        "invalid-date", // Неправильный формат
			Amount:      10000.0,
			Category:    "Test Category",
			Description: "Test Description",
		},
	}
	return req
}

// MockNegativeManualIncomeAmount - отрицательная сумма в manualIncome
func MockNegativeManualIncomeAmount() helpers.StatementStateRequest {
	req := GenerateStatementMocks()
	req.CustomData.ManualIncomes = []helpers.ManualIncome{
		{
			Date:        "2024-12-01",
			Amount:      -1000.0, // Отрицательная сумма
			Category:    "Test Category",
			Description: "Test Description",
		},
	}
	return req
}

// MockInvalidManualExpenseDate - неправильная дата в manualExpense
func MockInvalidManualExpenseDate() helpers.StatementStateRequest {
	req := GenerateStatementMocks()
	req.CustomData.ManualExpenses = []helpers.ManualExpense{
		{
			Date:        "invalid-date", // Неправильный формат
			Amount:      10000.0,
			Category:    "Test Category",
			Description: "Test Description",
		},
	}
	return req
}

// MockNegativeManualExpenseAmount - отрицательная сумма в manualExpense
func MockNegativeManualExpenseAmount() helpers.StatementStateRequest {
	req := GenerateStatementMocks()
	req.CustomData.ManualExpenses = []helpers.ManualExpense{
		{
			Date:        "2024-12-15",
			Amount:      -1000.0, // Отрицательная сумма
			Category:    "Test Category",
			Description: "Test Description",
		},
	}
	return req
}

// MockMissingManualExpenseCategory - отсутствует category в manualExpense
func MockMissingManualExpenseCategory() helpers.StatementStateRequest {
	req := GenerateStatementMocks()
	req.CustomData.ManualExpenses = []helpers.ManualExpense{
		{
			Date:        "2024-12-15",
			Amount:      10000.0,
			Category:    "", // Пустая категория
			Description: "Test Description",
		},
	}
	return req
}

// MockTooManyCustomCustomers - слишком много customCustomers (> 20)
func MockTooManyCustomCustomers() helpers.StatementStateRequest {
	req := GenerateStatementMocks()
	customers := make([]string, 21) // 21 элемент
	for i := range customers {
		customers[i] = fmt.Sprintf("Customer %d", i+1)
	}
	req.CustomData.CustomCustomers = customers
	return req
}

// MockEmptyCustomContractorName - пустое имя в customContractor
func MockEmptyCustomContractorName() helpers.StatementStateRequest {
	req := GenerateStatementMocks()
	req.CustomData.CustomContractors = []helpers.CustomContractor{
		{
			TransactionType: "Test Transaction Type",
			Name:            "", // Пустое имя
		},
	}
	return req
}

// MockEmptyCustomContractorType - пустой тип в customContractor
func MockEmptyCustomContractorType() helpers.StatementStateRequest {
	req := GenerateStatementMocks()
	req.CustomData.CustomContractors = []helpers.CustomContractor{
		{
			TransactionType: "", // Пустой тип
			Name:            "Test Name",
		},
	}
	return req
}
