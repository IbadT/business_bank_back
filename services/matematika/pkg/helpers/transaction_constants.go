package helpers

// ============================================================================
// TRANSACTION TYPE CONSTANTS (string values for API/DB)
// ============================================================================

const (
	// TransactionTypeIncomeStr строковое значение для типа транзакции "income"
	TransactionTypeIncomeStr = "income"
	// TransactionTypeExpenseStr строковое значение для типа транзакции "expense"
	TransactionTypeExpenseStr = "expense"
)

// ============================================================================
// PAYMENT METHOD CONSTANTS (string values for API/DB)
// ============================================================================

const (
	// PaymentMethodCardStr строковое значение для метода платежа "card"
	PaymentMethodCardStr = "card"
	// PaymentMethodAccountStr строковое значение для метода платежа "account"
	PaymentMethodAccountStr = "account"
	// PaymentMethodWireStr строковое значение для метода платежа "wire"
	PaymentMethodWireStr = "wire"
	// PaymentMethodZelleStr строковое значение для метода платежа "zelle"
	PaymentMethodZelleStr = "zelle"
	// PaymentMethodACHCreditStr строковое значение для метода платежа "ACH_CREDIT"
	PaymentMethodACHCreditStr = "ACH_CREDIT"
	// PaymentMethodACHCreditLowerStr строковое значение для метода платежа "ach_credit" (lowercase)
	PaymentMethodACHCreditLowerStr = "ach_credit"
	// PaymentMethodACHDebitStr строковое значение для метода платежа "ACH_DEBIT"
	PaymentMethodACHDebitStr = "ACH_DEBIT"
	// PaymentMethodACHDebitLowerStr строковое значение для метода платежа "ach_debit" (lowercase)
	PaymentMethodACHDebitLowerStr = "ach_debit"
	// PaymentMethodElectronicPaymentStr строковое значение для метода платежа "Electronic Payment"
	PaymentMethodElectronicPaymentStr = "Electronic Payment"
	// PaymentMethodElectronicPaymentLowerStr строковое значение для метода платежа "electronic payment" (lowercase)
	PaymentMethodElectronicPaymentLowerStr = "electronic payment"
	// PaymentMethodBankTransferStr строковое значение для метода платежа "bank_transfer"
	PaymentMethodBankTransferStr = "bank_transfer"
	// PaymentMethodACHCreditDashStr строковое значение для метода платежа "ACH-credit" (with dash)
	PaymentMethodACHCreditDashStr = "ACH-credit"
	// PaymentMethodWireTitleStr строковое значение для метода платежа "Wire" (title case)
	PaymentMethodWireTitleStr = "Wire"
	// PaymentMethodZelleTitleStr строковое значение для метода платежа "Zelle" (title case)
	PaymentMethodZelleTitleStr = "Zelle"
)
