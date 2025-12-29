// internal/transport/http/dto/response.go
package dto

import (
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/internal/domain"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/transport"
)

// GenerateResponse - DTO ответа с транзакциями
// @Description Ответ с сгенерированными транзакциями и финансовой сводкой
type GenerateResponse struct {
	RequestID            string           `json:"requestId" example:"550e8400-e29b-41d4-a716-446655440000"`
	Transactions         []Transaction    `json:"transactions"`
	FinancialSummary     FinancialSummary `json:"financialSummary"`
	DailyClosingBalances []DailyBalance   `json:"dailyClosingBalances"`
	ForwardingInfo       ForwardingInfo   `json:"forwardingInfo"`
	// TODO: изменить пакеты
	RevenueBreakdown  transport.RevenueBreakdown  `json:"revenueBreakdown,omitempty"`
	ExpensesBreakdown transport.ExpensesBreakdown `json:"expensesBreakdown,omitempty"`
	TransactionCounts transport.TransactionCounts `json:"transactionCounts,omitempty"`
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
	FixAsFirst         bool                   `json:"fixAsFirst" example:"false"`
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

// HolidayResponse - DTO ответа с информацией о празднике
// @Description Информация о празднике
type HolidayResponse struct {
	ID          string `json:"id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	HolidayDate string `json:"holidayDate" example:"2024-12-15"`
	Name        string `json:"name" example:"Новый Год"`
	Country     string `json:"country" example:"RU"`
}

// IsHolidayResponse - DTO ответа проверки праздника
// @Description Результат проверки, является ли дата праздником
type IsHolidayResponse struct {
	IsHoliday bool   `json:"isHoliday" example:"true"`
	Date      string `json:"date" example:"2024-12-15"`
}

// GetHolidaysResponse - DTO ответа со списком праздников
// @Description Список праздников за указанный год
type GetHolidaysResponse struct {
	Holidays []HolidayResponse `json:"holidays"`
	Year     string            `json:"year" example:"2024"`
}

// MessageResponse - DTO ответа с сообщением
// @Description Стандартный ответ с сообщением об успешной операции
type MessageResponse struct {
	Message string `json:"message" example:"Operation completed successfully"`
	Code    int    `json:"code" example:"200"`
}

// GetTransactionsResponse - DTO ответа со списком транзакций
// @Description Список транзакций
type GetTransactionsResponse struct {
	Transactions []domain.GeneratedTransaction `json:"transactions"`
	Code         int                           `json:"code" example:"200"`
}

// GetTransactionsCountResponse - DTO ответа с количеством транзакций
// @Description Количество транзакций
type GetTransactionsCountResponse struct {
	Count int64 `json:"count" example:"100"`
	Code  int   `json:"code" example:"200"`
}

// B2CGatewayResponse - DTO ответа с информацией о шлюзе для B2C
// @Description Информация о шлюзе для B2C
type B2CGatewayResponse struct {
	Gateway domain.Gateway `json:"gateway"`
	Code    int            `json:"code" example:"200"`
}

// CalculateRevenueBreakdownResponse - DTO ответа с разбивкой доходов
// @Description Разбивка доходов по методам платежа
type CalculateRevenueBreakdownResponse struct {
	RequestID        string           `json:"requestId" example:"550e8400-e29b-41d4-a716-446655440000"`
	RevenueBreakdown RevenueBreakdown `json:"revenueBreakdown"`
	Code             int              `json:"code" example:"200"`
}

// CalculateExpensesBreakdownResponse - DTO ответа с разбивкой расходов
// @Description Разбивка расходов по методам платежа
type CalculateExpensesBreakdownResponse struct {
	RequestID         string            `json:"requestId" example:"550e8400-e29b-41d4-a716-446655440000"`
	ExpensesBreakdown ExpensesBreakdown `json:"expensesBreakdown"`
	Code              int               `json:"code" example:"200"`
}

// RevenueBreakdown - разбивка доходов по методам платежа
// @Description Разбивка доходов по методам платежа
type RevenueBreakdown struct {
	TotalAch     float64 `json:"totalAch" example:"100000.00"`
	TotalWire    float64 `json:"totalWire" example:"100000.00"`
	TotalZelle   float64 `json:"totalZelle" example:"100000.00"`
	TotalGateway float64 `json:"totalGateway" example:"100000.00"`
	TotalOther   float64 `json:"totalOther" example:"100000.00"`
}

// ExpensesBreakdown - разбивка расходов по методам платежа
// @Description Разбивка расходов по методам платежа
type ExpensesBreakdown struct {
	ByCard    float64 `json:"byCard" example:"100000.00"`
	ByAccount float64 `json:"byAccount" example:"100000.00"`
}

// BaseAmountsResponse - ответ с базовыми суммами
// @Description Базовые суммы для мобильной связи, коммунальных и лизинга
type BaseAmountsResponse struct {
	UserID              string          `json:"userId" example:"550e8400-e29b-41d4-a716-446655440000"`
	MobileBaseAmount    *BaseAmountInfo `json:"mobileBaseAmount,omitempty"`
	UtilitiesBaseAmount *BaseAmountInfo `json:"utilitiesBaseAmount,omitempty"`
	LeasingBaseAmount   *BaseAmountInfo `json:"leasingBaseAmount,omitempty"`
	Code                int             `json:"code" example:"200"`
}

// BaseAmountInfo - информация о базовой сумме
// @Description Информация о базовой сумме
type BaseAmountInfo struct {
	Amount             float64 `json:"amount" example:"350.00"`
	FirstMonth         string  `json:"firstMonth" example:"2025-01"`
	FirstMonthTurnover float64 `json:"firstMonthTurnover,omitempty" example:"1000000.00"` // только для лизинга
	CreatedAt          string  `json:"createdAt" example:"2025-01-15T10:30:00Z"`
	UpdatedAt          string  `json:"updatedAt" example:"2025-01-15T10:30:00Z"`
}

// CalculateMobileAmountResponse - ответ с рассчитанной суммой мобильной связи
// @Description Рассчитанная сумма мобильной связи
type CalculateMobileAmountResponse struct {
	UserID       string  `json:"userId" example:"550e8400-e29b-41d4-a716-446655440000"`
	Amount       float64 `json:"amount" example:"350.00"`
	IsFirstMonth bool    `json:"isFirstMonth" example:"true"`
	Code         int     `json:"code" example:"200"`
}

// CalculateUtilitiesAmountResponse - ответ с рассчитанной суммой коммунальных
// @Description Рассчитанная сумма коммунальных
type CalculateUtilitiesAmountResponse struct {
	UserID       string  `json:"userId" example:"550e8400-e29b-41d4-a716-446655440000"`
	Amount       float64 `json:"amount" example:"425.50"`
	IsFirstMonth bool    `json:"isFirstMonth" example:"false"`
	Code         int     `json:"code" example:"200"`
}

// CalculateLeasingAmountResponse - ответ с рассчитанной суммой лизинга
// @Description Рассчитанная сумма лизинга
type CalculateLeasingAmountResponse struct {
	UserID       string  `json:"userId" example:"550e8400-e29b-41d4-a716-446655440000"`
	Amount       float64 `json:"amount" example:"11500.00"`
	Turnover     float64 `json:"turnover" example:"100000.00"`
	IsFirstMonth bool    `json:"isFirstMonth" example:"true"`
	Code         int     `json:"code" example:"200"`
}

// GetBalanceAdjustmentResponse - ответ с корректировкой баланса
// @Description Корректировка баланса
type GetBalanceAdjustmentResponse struct {
	RequestID    string                        `json:"requestId" example:"550e8400-e29b-41d4-a716-446655440000"`
	Transactions []domain.GeneratedTransaction `json:"transactions"`
	Code         int                           `json:"code" example:"200"`
}

// ValidateBalanceResponse - ответ валидации баланса
// @Description Результат валидации баланса транзакций
type ValidateBalanceResponse struct {
	RequestID    string                        `json:"requestId" example:"550e8400-e29b-41d4-a716-446655440000"`
	IsValid      bool                          `json:"isValid" example:"true"`
	Issues       []BalanceIssue                `json:"issues,omitempty"`
	Transactions []domain.GeneratedTransaction `json:"transactions,omitempty"`
	Code         int                           `json:"code" example:"200"`
}

// BalanceIssue - информация о проблеме с балансом
// @Description Информация о проблеме с балансом транзакции
type BalanceIssue struct {
	TransactionID    string  `json:"transactionId" example:"t_exp_045"`
	Date             string  `json:"date" example:"2025-01-15"`
	RequiredBalance  float64 `json:"requiredBalance" example:"5000.00"`
	AvailableBalance float64 `json:"availableBalance" example:"3000.00"`
	Shortage         float64 `json:"shortage" example:"2000.00"`
	ActionTaken      string  `json:"actionTaken,omitempty" example:"postponed" enums:"postponed,reduced,none"`
	NewDate          string  `json:"newDate,omitempty" example:"2025-01-17"`
	OriginalAmount   float64 `json:"originalAmount,omitempty" example:"-5000.00"`
	AdjustedAmount   float64 `json:"adjustedAmount,omitempty" example:"-3000.00"`
}

// SaveAssociatedCardResponse - ответ с сохранением номера карты
// @Description Сохранение номера карты
type SaveAssociatedCardResponse struct {
	Message string `json:"message" example:"Associated card saved successfully"`
	Code    int    `json:"code" example:"200"`
}

// ErrorResponse - единообразный ответ об ошибке
// @Description Стандартный формат ответа об ошибке для всех эндпоинтов
type ErrorResponse struct {
	Error   string `json:"error" example:"Invalid request body"`
	Details string `json:"details,omitempty" example:"Invalid JSON format"`
	Code    int    `json:"code" example:"400"`
}
