package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// ORM MODELS для v2 API - GORM модели для работы с базой данных
// ============================================================================

// GenerationRequest - модель таблицы generation_requests
// CREATE TABLE generation_requests (
//     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
//     user_id UUID,
//     month VARCHAR(7) NOT NULL,
//     year INTEGER NOT NULL,
//     turnover DECIMAL(15,2) NOT NULL,
//     desired_profit_percent DECIMAL(5,2) NOT NULL,
//     model VARCHAR(10) NOT NULL CHECK (model IN ('B2C', 'B2B')),
//     initial_balance DECIMAL(15,2) NOT NULL,
//     scale_factor INTEGER DEFAULT 1,
//     custom_data JSONB,
//     status VARCHAR(20) DEFAULT 'processing',
//     error_message TEXT,
//     created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
//     completed_at TIMESTAMP WITH TIME ZONE,
//     updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
// );
type GenerationRequest struct {
	ID                 uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID             *uuid.UUID      `gorm:"type:uuid;column:user_id"`
	Month              string           `gorm:"type:varchar(7);not null"`
	Year               int              `gorm:"type:integer;not null"`
	Turnover           float64          `gorm:"type:decimal(15,2);not null"`
	DesiredProfitPercent float64        `gorm:"type:decimal(5,2);not null;column:desired_profit_percent"`
	Model              string           `gorm:"type:varchar(10);not null;check:model IN ('B2C', 'B2B')"`
	InitialBalance     float64          `gorm:"type:decimal(15,2);not null;column:initial_balance"`
	ScaleFactor        int              `gorm:"type:integer;default:1;column:scale_factor"`
	CustomData         JSONB            `gorm:"type:jsonb;column:custom_data"`
	Status             string           `gorm:"type:varchar(20);default:'processing'"`
	ErrorMessage       *string          `gorm:"type:text;column:error_message"`
	CreatedAt          time.Time        `gorm:"type:timestamp with time zone;default:current_timestamp;column:created_at"`
	CompletedAt        *time.Time       `gorm:"type:timestamp with time zone;column:completed_at"`
	UpdatedAt          time.Time        `gorm:"type:timestamp with time zone;default:current_timestamp;column:updated_at"`
}

// TableName указывает имя таблицы для GORM
func (GenerationRequest) TableName() string {
	return "generation_requests"
}

// GeneratedTransaction - модель таблицы generated_transactions
// CREATE TABLE generated_transactions (
//     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
//     request_id UUID NOT NULL REFERENCES generation_requests(id),
//     transaction_id VARCHAR(50) NOT NULL,
//     transaction_date TIMESTAMP WITH TIME ZONE NOT NULL,
//     posting_date DATE NOT NULL,
//     type VARCHAR(10) NOT NULL CHECK (type IN ('income', 'expense')),
//     category VARCHAR(100) NOT NULL,
//     method VARCHAR(50) NOT NULL,
//     amount DECIMAL(15,2) NOT NULL,
//     balance_after DECIMAL(15,2),
//     is_manual BOOLEAN DEFAULT FALSE,
//     calculation_details JSONB,
//     sort_order INTEGER
// );
type GeneratedTransaction struct {
	ID                 uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	RequestID          uuid.UUID       `gorm:"type:uuid;not null;column:request_id"`
	TransactionID      string           `gorm:"type:varchar(50);not null;column:transaction_id"`
	TransactionDate    time.Time        `gorm:"type:timestamp with time zone;not null;column:transaction_date"`
	PostingDate        time.Time        `gorm:"type:date;not null;column:posting_date"`
	Type               string           `gorm:"type:varchar(10);not null;check:type IN ('income', 'expense')"`
	Category           string           `gorm:"type:varchar(100);not null"`
	Method             string           `gorm:"type:varchar(50);not null"`
	Amount             float64          `gorm:"type:decimal(15,2);not null"`
	BalanceAfter       *float64         `gorm:"type:decimal(15,2);column:balance_after"`
	IsManual           bool             `gorm:"type:boolean;default:false;column:is_manual"`
	CalculationDetails JSONB            `gorm:"type:jsonb;column:calculation_details"`
	SortOrder          *int             `gorm:"type:integer;column:sort_order"`
}

// TableName указывает имя таблицы для GORM
func (GeneratedTransaction) TableName() string {
	return "generated_transactions"
}

// FinancialSummaryDB - модель таблицы financial_summaries
// CREATE TABLE financial_summaries (
//     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
//     request_id UUID NOT NULL REFERENCES generation_requests(id),
//     initial_balance DECIMAL(15,2) NOT NULL,
//     final_balance DECIMAL(15,2) NOT NULL,
//     total_revenue DECIMAL(15,2) NOT NULL,
//     total_expenses DECIMAL(15,2) NOT NULL,
//     net_profit DECIMAL(15,2) NOT NULL
// );
type FinancialSummaryDB struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	RequestID      uuid.UUID `gorm:"type:uuid;not null;column:request_id"`
	InitialBalance float64   `gorm:"type:decimal(15,2);not null;column:initial_balance"`
	FinalBalance   float64   `gorm:"type:decimal(15,2);not null;column:final_balance"`
	TotalRevenue   float64   `gorm:"type:decimal(15,2);not null;column:total_revenue"`
	TotalExpenses  float64   `gorm:"type:decimal(15,2);not null;column:total_expenses"`
	NetProfit      float64   `gorm:"type:decimal(15,2);not null;column:net_profit"`
}

// TableName указывает имя таблицы для GORM
func (FinancialSummaryDB) TableName() string {
	return "financial_summaries"
}

// DailyBalanceV2 - модель таблицы daily_balances (v2)
// CREATE TABLE daily_balances (
//     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
//     request_id UUID NOT NULL REFERENCES generation_requests(id),
//     balance_date DATE NOT NULL,
//     balance DECIMAL(15,2) NOT NULL
// );
type DailyBalanceV2 struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	RequestID   uuid.UUID `gorm:"type:uuid;not null;column:request_id"`
	BalanceDate time.Time `gorm:"type:date;not null;column:balance_date"`
	Balance     float64   `gorm:"type:decimal(15,2);not null"`
}

// TableName указывает имя таблицы для GORM
func (DailyBalanceV2) TableName() string {
	return "daily_balances"
}

// TransactionTemplateDB - модель таблицы transaction_templates
// CREATE TABLE transaction_templates (
//     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
//     template_key VARCHAR(100) NOT NULL UNIQUE,
//     category VARCHAR(100) NOT NULL,
//     type VARCHAR(10) NOT NULL CHECK (type IN ('income', 'expense')),
//     is_percentage BOOLEAN NOT NULL DEFAULT FALSE,
//     percentage_min DECIMAL(5,4),
//     percentage_max DECIMAL(5,2),
//     fixed_amount DECIMAL(15,2),
//     frequency VARCHAR(50),
//     preferred_day VARCHAR(20),
//     week_of_month INTEGER[],
//     business_hours JSONB,
//     is_optional BOOLEAN DEFAULT FALSE,
//     priority INTEGER DEFAULT 100,
//     method VARCHAR(50),
//     min_transactions INTEGER DEFAULT 1,
//     max_transactions INTEGER DEFAULT 1
// );
type TransactionTemplateDB struct {
	ID              uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TemplateKey     string          `gorm:"type:varchar(100);not null;unique;column:template_key"`
	Category        string          `gorm:"type:varchar(100);not null"`
	Type            string          `gorm:"type:varchar(10);not null;check:type IN ('income', 'expense')"`
	IsPercentage    bool            `gorm:"type:boolean;not null;default:false;column:is_percentage"`
	PercentageMin   *float64         `gorm:"type:decimal(5,4);column:percentage_min"`
	PercentageMax   *float64         `gorm:"type:decimal(5,2);column:percentage_max"`
	FixedAmount     *float64         `gorm:"type:decimal(15,2);column:fixed_amount"`
	Frequency       *string          `gorm:"type:varchar(50)"`
	PreferredDay    *string          `gorm:"type:varchar(20);column:preferred_day"`
	WeekOfMonth     IntArray         `gorm:"type:integer[];column:week_of_month"`
	BusinessHours   JSONB            `gorm:"type:jsonb;column:business_hours"`
	IsOptional      bool             `gorm:"type:boolean;default:false;column:is_optional"`
	Priority        int              `gorm:"type:integer;default:100"`
	Method          *string          `gorm:"type:varchar(50)"`
	MinTransactions int              `gorm:"type:integer;default:1;column:min_transactions"`
	MaxTransactions int              `gorm:"type:integer;default:1;column:max_transactions"`
}

// TableName указывает имя таблицы для GORM
func (TransactionTemplateDB) TableName() string {
	return "transaction_templates"
}

// DefaultCustomerDB - модель таблицы default_customers
// CREATE TABLE default_customers (
//     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
//     name VARCHAR(100) NOT NULL,
//     category VARCHAR(50) NOT NULL,
//     min_percent DECIMAL(5,4) NOT NULL,
//     max_percent DECIMAL(5,4) NOT NULL,
//     min_transactions INTEGER NOT NULL,
//     max_transactions INTEGER NOT NULL
// );
type DefaultCustomerDB struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name            string    `gorm:"type:varchar(100);not null"`
	Category        string    `gorm:"type:varchar(50);not null"`
	MinPercent      float64   `gorm:"type:decimal(5,4);not null;column:min_percent"`
	MaxPercent      float64   `gorm:"type:decimal(5,4);not null;column:max_percent"`
	MinTransactions int       `gorm:"type:integer;not null;column:min_transactions"`
	MaxTransactions int       `gorm:"type:integer;not null;column:max_transactions"`
}

// ============================================================================
// Вспомогательные типы для работы с PostgreSQL специфичными типами
// ============================================================================

// JSONB - тип для работы с JSONB полями PostgreSQL
type JSONB map[string]interface{}

// Value реализует driver.Valuer для JSONB
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan реализует sql.Scanner для JSONB
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte(value.(string)), j)
	}
	return json.Unmarshal(bytes, j)
}

// IntArray - тип для работы с INTEGER[] полями PostgreSQL
type IntArray []int

// Value реализует driver.Valuer для IntArray
func (a IntArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a)
}

// Scan реализует sql.Scanner для IntArray
func (a *IntArray) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte(value.(string)), a)
	}
	return json.Unmarshal(bytes, a)
}

// HolidayDB - модель таблицы holidays
// CREATE TABLE holidays (
//     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
//     holiday_date DATE NOT NULL UNIQUE,
//     name VARCHAR(100) NOT NULL,
//     country VARCHAR(2) DEFAULT 'RU'
// );
type Holiday struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	HolidayDate time.Time `gorm:"type:date;not null;unique;column:holiday_date"`
	Name        string    `gorm:"type:varchar(100);not null"`
	Country     string    `gorm:"type:varchar(2);default:'RU'"`
}


// GenerationState - модель таблицы generation_state для сохранения состояния между генерациями
// CREATE TABLE generation_state (
//     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
//     user_id UUID,
//     state_key VARCHAR(100) NOT NULL,
//     state_value JSONB NOT NULL,
//     created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
//     updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
//     UNIQUE(user_id, state_key)
// );
type GenerationState struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    *uuid.UUID `gorm:"type:uuid;column:user_id"`
	StateKey  string     `gorm:"type:varchar(100);not null;column:state_key"`
	StateValue JSONB     `gorm:"type:jsonb;not null;column:state_value"`
	CreatedAt time.Time  `gorm:"type:timestamp with time zone;default:current_timestamp;column:created_at"`
	UpdatedAt time.Time  `gorm:"type:timestamp with time zone;default:current_timestamp;column:updated_at"`
}

// TableName указывает имя таблицы для GORM
func (GenerationState) TableName() string {
	return "generation_state"
}

// User - модель таблицы users
// CREATE TABLE users (
//     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
//     email VARCHAR(255) NOT NULL UNIQUE,
//     password_hash VARCHAR(255) NOT NULL,
//     role VARCHAR(50) NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin')),
//     created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
//     updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
// );
// type User struct {
// 	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
// 	Email        string    `gorm:"type:varchar(255);not null;unique"`
// 	Password string    `gorm:"type:varchar(255);not null;column:password"`
// 	Role         string    `gorm:"type:varchar(50);not null;default:'user';check:role IN ('user', 'admin')"`
// 	CreatedAt    time.Time `gorm:"type:timestamp with time zone;default:current_timestamp;column:created_at"`
// 	UpdatedAt    time.Time `gorm:"type:timestamp with time zone;default:current_timestamp;column:updated_at"`
// }

// // TableName указывает имя таблицы для GORM
// func (User) TableName() string {
// 	return "users"
// }

