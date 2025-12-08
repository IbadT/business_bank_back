// models/response.go
package models

import "time"

type Transaction struct {
	TransactionID     string                 `json:"transactionId"`    // [44] transactionId
	TransactionDate  time.Time              `json:"transactionDate"`  // [44] transactionDate
	PostingDate      string                 `json:"postingDate"`      // [44] postingDate
	Type             TransactionType        `json:"type"`             // [44] type
	Category         string                 `json:"category"`         // [44] category
	Method           PaymentMethod          `json:"method"`           // [44] method
	Amount           float64                `json:"amount"`           // [44] amount
	BalanceAfter     float64                `json:"balanceAfter"`     // [44] balanceAfter
	IsManual         bool                   `json:"isManual"`         // [44] isManual
	CalculationDetails map[string]interface{} `json:"calculationDetails,omitempty"` // [20][21] calculationDetails
}

type DailyBalance struct {
	Date    string  `json:"date"`    // [50] date
	Balance float64 `json:"balance"` // [50] balance
}

type FinancialSummary struct {
	InitialBalance float64 `json:"initialBalance"` // [45] initialBalance
	FinalBalance   float64 `json:"finalBalance"`   // [45] finalBalance
	TotalRevenue   float64 `json:"totalRevenue"`   // [45] totalRevenue
	TotalExpenses  float64 `json:"totalExpenses"`  // [45] totalExpenses
	NetProfit      float64 `json:"netProfit"`      // [45] netProfit
}

type ForwardingInfo struct {
	AssociatedCard    string             `json:"associatedCard"`    // [51][52] associatedCard
	OwnerName         string             `json:"ownerName"`         // [forwardingInfo] ownerName
	CompanyName       string             `json:"companyName"`       // [forwardingInfo] companyName
	CustomCustomers   []string           `json:"customCustomers"`   // [6][forwardingInfo] customCustomers
	CustomContractors []CustomContractor `json:"customContractors"` // [53] customContractors
}

type GenerateResponse struct {
	Transactions        []Transaction     `json:"transactions"`              // [44] Список транзакций
	FinancialSummary    FinancialSummary  `json:"financialSummary"`          // [45] Сводные показатели
	DailyClosingBalances []DailyBalance   `json:"dailyClosingBalances"`      // [50] Ежедневные балансы
	ForwardingInfo      ForwardingInfo    `json:"forwardingInfo"`            // [51-54] Данные для маскировки
}