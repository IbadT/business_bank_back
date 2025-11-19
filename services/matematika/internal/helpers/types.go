package helpers

import "github.com/google/uuid"

// ============================================================================
// TRANSACTION TYPES
// ============================================================================

type TransactionType string

const (
	TransactionTypeIncome  TransactionType = "income"
	TransactionTypeExpense TransactionType = "expense"
)

func (t TransactionType) IsValid() bool {
	return t == TransactionTypeIncome || t == TransactionTypeExpense
}

func (t TransactionType) String() string {
	return string(t)
}

// ============================================================================
// TRANSACTION METHODS
// ============================================================================

type TransactionMethod string

const (
	TransactionMethodACHCredit    TransactionMethod = "ACH_CREDIT"
	TransactionMethodWireCredit   TransactionMethod = "WIRE_CREDIT"
	TransactionMethodCashCredit   TransactionMethod = "CASH_CREDIT"
	TransactionMethodBankTransfer TransactionMethod = "BANK_TRANSFER"
	TransactionMethodOther        TransactionMethod = "OTHER"
)

func (m TransactionMethod) IsValid() bool {
	switch m {
	case TransactionMethodACHCredit, TransactionMethodWireCredit,
		TransactionMethodCashCredit, TransactionMethodBankTransfer,
		TransactionMethodOther:
		return true
	}
	return false
}

func (m TransactionMethod) String() string {
	return string(m)
}

// ============================================================================
// BUSINESS TYPES
// ============================================================================

type BusinessType string

const (
	BusinessTypeB2B BusinessType = "B2B"
	BusinessTypeB2C BusinessType = "B2C"
)

func (b BusinessType) IsValid() bool {
	return b == BusinessTypeB2B || b == BusinessTypeB2C
}

func (b BusinessType) String() string {
	return string(b)
}

// ============================================================================
// STATEMENT STATUS
// ============================================================================

type StatementStatus string

const (
	StatusPending    StatementStatus = "pending"
	StatusProcessing StatementStatus = "processing"
	StatusCompleted  StatementStatus = "completed"
	StatusFailed     StatementStatus = "failed"
)

func (s StatementStatus) IsValid() bool {
	switch s {
	case StatusPending, StatusProcessing, StatusCompleted, StatusFailed:
		return true
	}
	return false
}

func (s StatementStatus) String() string {
	return string(s)
}

// IsFinal - проверяет является ли статус финальным (нельзя изменить)
func (s StatementStatus) IsFinal() bool {
	return s == StatusCompleted || s == StatusFailed
}

// ============================================================================
// ADMIN CONFIG TYPES
// ============================================================================

type AdminConfigResponse struct {
	ExpenseCategories []ExpenseCategoryResponse `json:"expense_categories"`
	Schedules         []SchedulesResponse       `json:"schedules"`
	IncomeTemplates   []IncomeTemplateResponse  `json:"income_templates"`
}
type ExpenseCategoryResponse struct {
	ID                uuid.UUID `json:"id"`
	Name              string    `json:"name"`
	DefaultPercentMin float64   `json:"default_percent_min"`
	DefaultPercentMax float64   `json:"default_percent_max"`
	IsOptional        bool      `json:"is_optional"`
	Priority          int       `json:"priority"`
}

type SchedulesResponse struct {
	ID           uuid.UUID `json:"id"`
	CategoryID   uuid.UUID `json:"category_id"`
	Frequency    string    `json:"frequency"`
	PreferredDay string    `json:"preferred_day"`
	WeekOfMonth  int       `json:"week_of_month"`
	NTimes       int       `json:"n_times"`
	TimeWindow   string    `json:"time_window"`
}

type IncomeTemplateResponse struct {
	ID                       uuid.UUID `json:"id"`
	BusinessModel            string    `json:"business_model"`
	Category                 string    `json:"category"`
	CountMin                 int       `json:"count_min"`
	CountMax                 int       `json:"count_max"`
	PercentPerTransactionMin float64   `json:"percent_per_transaction_min"`
	PercentPerTransactionMax float64   `json:"percent_per_transaction_max"`
	DefaultMethods           []string  `json:"default_methods"`
}

// ============================================================================
// HEALTH CHECK TYPES
// ============================================================================

type HealthCheckResponse struct {
	Status       string                  `json:"status"`
	Timestamp    string                  `json:"timestamp"`
	Version      string                  `json:"version"`
	Dependencies HealthCheckDependencies `json:"dependencies"`
}

type HealthCheckDependencies struct {
	Kafka        string `json:"kafka"`
	Database     string `json:"database"`
	Redis        string `json:"redis"`
	ConfigLoaded bool   `json:"config_loaded"`
	Service      string `json:"service"`
}

// ============================================================================
// REQUEST/RESPONSE MODELS
// ============================================================================

// GenerateStatementRequest - запрос на генерацию выписки
type GenerateStatementRequest struct {
	AccountID      string  `json:"accountId" validate:"required"`
	Month          string  `json:"month" validate:"required"`
	BusinessType   string  `json:"businessType" validate:"required,oneof=B2B B2C"`
	InitialBalance float64 `json:"initialBalance" validate:"required,gte=0"`
}

// ================================================
// STATE
// ================================================

type StatementStateRequest struct {
	CompanyInfo CompanyInfo `json:"companyInfo" validate:"required"`
	Financials  Financials  `json:"financials" validate:"required"`
	CustomData  CustomData  `json:"customData" validate:"required"`
}

type Financials struct {
	StartBalance        float64  `json:"startBalance" validate:"required,gte=0"`
	Turnover            float64  `json:"turnover" validate:"required,gte=0"`
	ProfitPercent       float64  `json:"profitPercent" validate:"required,gte=0,lte=50"`
	TargetProfit        float64  `json:"targetProfit" validate:"required,gte=0"`
	Months              int      `json:"months" validate:"omitempty,gte=1,lte=36"`
	StartMonth          string   `json:"startMonth" validate:"omitempty"`
	Periods             []string `json:"periods" validate:"omitempty,dive"`
	OperationMultiplier float64  `json:"operationMultiplier" validate:"required,gt=0,lte=5"`
}

type ManualIncome struct {
	Date        string  `json:"date" validate:"required,datetime=2006-01-02"`
	Amount      float64 `json:"amount" validate:"required,gte=0"`
	Category    string  `json:"category" validate:"required"`
	Description string  `json:"description" validate:"omitempty"`
}

type ManualExpense struct {
	Date        string  `json:"date" validate:"required,datetime=2006-01-02"`
	Amount      float64 `json:"amount" validate:"required,gte=0"`
	Category    string  `json:"category" validate:"required"`
	Description string  `json:"description" validate:"omitempty"`
}

type CustomContractor struct {
	TransactionType string `json:"transaction_type"` // Тип транзакции (можно сделать enum потом)
	Name            string `json:"name"`             // Имя контрагента
}

type CustomData struct {
	// ManualIncomes     []ManualIncome     `json:"manualIncomes,omitempty"`
	ManualIncomes     []ManualIncome     `json:"manualIncomes,omitempty"`
	ManualExpenses    []ManualExpense    `json:"manualExpenses,omitempty"`
	CustomCustomers   []string           `json:"customCustomers,omitempty"`
	CustomContractors []CustomContractor `json:"customContractors,omitempty"`
	DisableCategories []string           `json:"disableCategories,omitempty"`
}
type CompanyInfo struct {
	CompanyName    string `json:"companyName" validate:"required"`
	OwnerName      string `json:"ownerName" validate:"required"`
	AccountNumber  string `json:"accountNumber" validate:"required"`
	AssociatedCard string `json:"associatedCard" validate:"required"`
	Model          string `json:"model" validate:"required,oneof=B2B B2C"`
	State          string `json:"state" validate:"required"`
	Industry       string `json:"industry" validate:"required"`
}

// GenerateStatementResponse - ответ на генерацию выписки
type GenerateStatementResponse struct {
	StatementID string `json:"statementId"`
	Status      string `json:"status"`
	Message     string `json:"message"`
}

// ============================================================================
// ERROR RESPONSE TYPE
// ============================================================================

type ErrorResponse struct {
	Error   error       `json:"error"`
	Message error       `json:"message"`
	Details interface{} `json:"details,omitempty"`
}
