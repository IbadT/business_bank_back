// internal/transport/http/dto/response.go
package dto

import "time"

// GenerateResponse - DTO ответа с транзакциями
// @Description Ответ с сгенерированными транзакциями и финансовой сводкой
type GenerateResponse struct {
	Transactions         []Transaction     `json:"transactions"`
	FinancialSummary     FinancialSummary  `json:"financialSummary"`
	DailyClosingBalances []DailyBalance   `json:"dailyClosingBalances"`
	ForwardingInfo       ForwardingInfo    `json:"forwardingInfo"`
}

// Transaction - транзакция
// @Description Информация об одной транзакции
type Transaction struct {
	TransactionID      string                 `json:"transactionId" example:"550e8400-e29b-41d4-a716-446655440000"`
	TransactionDate    time.Time              `json:"transactionDate" example:"2024-12-15T10:30:00Z" format:"date-time"`
	PostingDate        string                 `json:"postingDate" example:"2024-12-15"`
	Type               string                 `json:"type" example:"income"`
	Category           string                 `json:"category" example:"Sales"`
	Method             string                 `json:"method" example:"card"`
	Amount             float64                `json:"amount" example:"15000.00"`
	BalanceAfter       float64                `json:"balanceAfter" example:"65000.00"`
	IsManual           bool                   `json:"isManual" example:"false"`
	CalculationDetails map[string]interface{} `json:"calculationDetails,omitempty"`
}

// FinancialSummary - финансовая сводка
// @Description Сводная информация о финансах за период
type FinancialSummary struct {
	InitialBalance float64 `json:"initialBalance" example:"50000.00"`
	FinalBalance   float64 `json:"finalBalance" example:"150000.00"`
	TotalRevenue   float64 `json:"totalRevenue" example:"1200000.00"`
	TotalExpenses  float64 `json:"totalExpenses" example:"1100000.00"`
	NetProfit      float64 `json:"netProfit" example:"100000.00"`
}

// DailyBalance - ежедневный баланс
// @Description Баланс счета на конец дня
type DailyBalance struct {
	Date    string  `json:"date" example:"2024-12-15"`
	Balance float64 `json:"balance" example:"65000.00"`
}

// ForwardingInfo - информация для пересылки
// @Description Дополнительная информация о компании и контрагентах
type ForwardingInfo struct {
	AssociatedCard    string             `json:"associatedCard" example:"****1234"`
	OwnerName         string             `json:"ownerName" example:"Иван Иванов"`
	CompanyName       string             `json:"companyName" example:"ООО \"Пример\""`
	CustomCustomers   []string           `json:"customCustomers"`
	CustomContractors []CustomContractor `json:"customContractors"`
}

// TokenResponse - DTO ответа с токенами авторизации
// @Description Ответ с токенами доступа и обновления после успешной авторизации
type TokenResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}