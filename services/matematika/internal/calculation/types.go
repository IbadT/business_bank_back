package calculation

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
