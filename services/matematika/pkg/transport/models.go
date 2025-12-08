package transport

import helpers "github.com/IbadT/business_bank_back/services/matematika/pkg/helpers"

// ============================================================================
// RESPONSE/DTO MODELS - Модели для API ответов
// ============================================================================

// MatematikaResponse - response для выписок (один или несколько месяцев)
// Формат: { "JANUARY 2025": {...}, "FEBRUARY 2025": {...}, ... }
type MatematikaResponse map[string]MonthlyStatement

// MonthlyStatement - данные выписки за один месяц
type MonthlyStatement struct {
	FinancialSummary     FinancialSummary      `json:"financial_summary"`
	Totals               *Totals               `json:"totals,omitempty"`             // Опционально для January
	RevenueBreakdown     *RevenueBreakdown     `json:"revenue_breakdown,omitempty"`  // Опционально для January
	ExpensesBreakdown    *ExpensesBreakdown    `json:"expenses_breakdown,omitempty"` // Опционально для January
	TransactionCounts    *TransactionCounts    `json:"transaction_counts,omitempty"` // Опционально для January
	Transactions         []TransactionResponse `json:"transactions"`
	ForwardingInfo       ForwardingInfo        `json:"forwarding_info"`
	DailyClosingBalances []DailyClosingBalance `json:"daily_closing_balances"`
}

// ============================================================================
// TRANSACTION RESPONSE
// ============================================================================

// TransactionResponse - модель транзакции для API ответа
type TransactionResponse struct {
	TransactionID      string                    `json:"transaction_id"`
	TransactionDate    string                    `json:"transaction_date"` // ISO8601: "2025-01-06T11:00:00"
	PostingDate        string                    `json:"posting_date"`     // YYYY-MM-DD: "2025-01-06"
	Type               helpers.TransactionType   `json:"type"`
	Category           string                    `json:"category"`
	Method             helpers.TransactionMethod `json:"method"`
	Amount             float64                   `json:"amount"`
	IsManual           bool                      `json:"is_manual"`
	BalanceAfter       float64                   `json:"balance_after"`
	CalculationDetails *CalculationDetails       `json:"calculation_details,omitempty"` // Опционально
	FixAsFirst         bool                      `json:"fix_as_first"`
	FixAsFirstCount    *int                      `json:"fix_as_first_count,omitempty"` // Опционально
}

// CalculationDetails - детали расчета для транзакции
type CalculationDetails struct {
	WeightLb    *float64 `json:"weight_lb,omitempty"`     // Для взвешивания
	RatePerLb   *float64 `json:"rate_per_lb,omitempty"`   // Для взвешивания
	Hours       *int     `json:"hours,omitempty"`         // Для аренды (chassis usage)
	RatePerHour *float64 `json:"rate_per_hour,omitempty"` // Для аренды (chassis usage)
}

// ============================================================================
// DAILY CLOSING BALANCES
// ============================================================================

// DailyClosingBalance - дневной баланс для API ответа
type DailyClosingBalance struct {
	Date    string  `json:"date"` // YYYY-MM-DD: "2025-01-01"
	Balance float64 `json:"balance"`
}

// ============================================================================
// FORWARDING INFO
// ============================================================================

// ForwardingInfo - информация о переадресации для API ответа
type ForwardingInfo struct {
	AssociatedCard    string             `json:"associated_card"`
	OwnerName         string             `json:"owner_name"`
	CustomCustomers   []string           `json:"custom_customers"` // Массив строк: ["Super LLC", "Lulu Inc."]
	CustomContractors []CustomContractor `json:"custom_contractors"`
}

// CustomContractor - кастомный контрагент для API ответа
type CustomContractor struct {
	TransactionType string `json:"transaction_type"` // Тип транзакции
	Name            string `json:"name"`             // Имя контрагента
}

// ============================================================================
// FINANCIAL SUMMARY
// ============================================================================

// FinancialSummary - финансовое резюме для API ответа
type FinancialSummary struct {
	CompanyName    string  `json:"company_name"`
	AccountNumber  string  `json:"account_number"`
	Period         string  `json:"period"` // "2025-02-01 - 2025-02-28"
	InitialBalance float64 `json:"initial_balance"`
	FinalBalance   float64 `json:"final_balance"`
	TotalRevenue   float64 `json:"total_revenue"`
	TotalExpenses  float64 `json:"total_expenses"`
	NetProfit      float64 `json:"net_profit"`
}

// ============================================================================
// TOTALS & BREAKDOWNS (опционально для January с детальной статистикой)
// ============================================================================

// Totals - итоговые суммы для API ответа
type Totals struct {
	TotalRevenue  float64 `json:"total_revenue"`
	TotalExpenses float64 `json:"total_expenses"`
	NetProfit     float64 `json:"net_profit"`
}

// RevenueBreakdown - разбивка доходов для API ответа
type RevenueBreakdown struct {
	TotalAch     float64 `json:"total_ach"`
	TotalWire    float64 `json:"total_wire"`
	TotalZelle   float64 `json:"total_zelle"`
	TotalGateway float64 `json:"total_gateway"`
	TotalOther   float64 `json:"total_other"`
}

// ExpensesBreakdown - разбивка расходов для API ответа
type ExpensesBreakdown struct {
	ByCard    float64 `json:"by_card"`
	ByAccount float64 `json:"by_account"`
}

// TransactionCounts - количество транзакций для API ответа
type TransactionCounts struct {
	Total       int              `json:"total"`
	Deposits    DepositCounts    `json:"deposits"`
	Withdrawals WithdrawalCounts `json:"withdrawals"`
}

// DepositCounts - количество депозитов для API ответа
type DepositCounts struct {
	Total int `json:"total"`
	Ach   int `json:"ach"`
	Wire  int `json:"wire"`
	Zelle int `json:"zelle"`
}

// WithdrawalCounts - количество выводов для API ответа
type WithdrawalCounts struct {
	Total       int `json:"total"`
	FromAccount int `json:"from_account"`
	ByCard      int `json:"by_card"`
}
