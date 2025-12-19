// internal/domain/value_objects/payment_method.go
package value_objects

import "github.com/IbadT/business_bank_back/services/matematika/pkg/helpers"

// PaymentMethod - Value Object метода платежа [44]
type PaymentMethod string

const (
	ACHCredit         PaymentMethod = "ACH_CREDIT"
	ACHDebit          PaymentMethod = "ACH_DEBIT"
	ElectronicPayment PaymentMethod = "Electronic Payment"
	Card              PaymentMethod = "card"
	Account           PaymentMethod = "account"
	Wire              PaymentMethod = "wire"
	Zelle             PaymentMethod = "zelle"
)

// NewPaymentMethod создает новый метод платежа с валидацией
func NewPaymentMethod(m string) (PaymentMethod, error) {
	switch m {
	case string(ACHCredit), string(ACHDebit), string(ElectronicPayment),
	     string(Card), string(Account), string(Wire), string(Zelle):
		return PaymentMethod(m), nil
	default:
		return "", helpers.ErrInvalidPaymentMethod
	}
}

// String возвращает строковое представление
func (pm PaymentMethod) String() string {
	return string(pm)
}

// IsValid проверяет валидность метода
func (pm PaymentMethod) IsValid() bool {
	switch pm {
	case ACHCredit, ACHDebit, ElectronicPayment, Card, Account, Wire, Zelle:
		return true
	default:
		return false
	}
}

// IsAccountTransfer проверяет, является ли операция переводом по счету [32][33]
func (pm PaymentMethod) IsAccountTransfer() bool {
	return pm == ACHCredit || pm == ACHDebit || pm == Wire || pm == Account
}

// IsCardOperation проверяет, является ли операция карточной [33]
func (pm PaymentMethod) IsCardOperation() bool {
	return pm == Card
}
