package calculation

import (
	"time"

	"github.com/google/uuid"
)

// CREATE TABLE statements (
//     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
//     account_id VARCHAR(50) NOT NULL,
//     month VARCHAR(7) NOT NULL, -- Format: YYYY-MM
//     business_type VARCHAR(10) NOT NULL, -- B2B, B2C
//     initial_balance DECIMAL(15,2) NOT NULL,
//     final_balance DECIMAL(15,2),
//     total_income DECIMAL(15,2),
//     total_expenses DECIMAL(15,2),
//     net_profit DECIMAL(15,2),
//     profit_percentage DECIMAL(5,2),
//     status VARCHAR(20) NOT NULL, -- pending, processing, completed, failed
//     correlation_id UUID,
//     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
//     updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
//     completed_at TIMESTAMP
// );

type Statement struct {
	ID        uuid.UUID `gorm:"primaryKey"`
	Status    string    `gorm:"status; not null"`
	CreatedAt time.Time `gorm:"created_at"`
	UpdatedAt time.Time `gorm:"updated_at"`
}

// CREATE TABLE transactions (
//     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
//     statement_id UUID NOT NULL REFERENCES statements(id) ON DELETE CASCADE,
//     transaction_date DATE NOT NULL,
//     transaction_type VARCHAR(20) NOT NULL, -- income, expense
//     category VARCHAR(50) NOT NULL, -- ACH, Wire, Payroll, etc.
//     amount DECIMAL(15,2) NOT NULL,
//     balance_after DECIMAL(15,2) NOT NULL,
//     is_user_defined BOOLEAN DEFAULT FALSE,
//     user_notes TEXT,
//     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
// );

type Transaction struct {
	ID              uuid.UUID `gorm:"primaryKey"`
	TransactionDate time.Time `gorm:"transaction_date; not null"`
	TransactionType string    `gorm:"transaction_type; not null"`
	Category        string    `gorm:"category; not null"`
	Amount          float64   `gorm:"amount; not null"`
	BalanceAfter    float64   `gorm:"balance_after; not null"`
	IsUserDefined   bool      `gorm:"is_user_defined; default:false"`
	UserNotes       string    `gorm:"user_notes"`
	CreatedAt       time.Time `gorm:"created_at"`
	UpdatedAt       time.Time `gorm:"updated_at"`
}

// CREATE TABLE business_rules (
//
//	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
//	rule_name VARCHAR(100) NOT NULL UNIQUE,
//	rule_type VARCHAR(50) NOT NULL, -- profit_range, transaction_frequency, etc.
//	business_type VARCHAR(10) NOT NULL, -- B2B, B2C
//	min_value DECIMAL(10,2),
//	max_value DECIMAL(10,2),
//	default_value DECIMAL(10,2),
//	description TEXT,
//	is_active BOOLEAN DEFAULT TRUE,
//	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
//	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
//
// );
type BusinessRule struct {
	ID           uuid.UUID `gorm:"primaryKey"`
	BusinessType string    `gorm:"business_type; not null"`
	Description  string    `gorm:"description"`
	CreatedAt    time.Time `gorm:"created_at"`
	UpdatedAt    time.Time `gorm:"updated_at"`
}
