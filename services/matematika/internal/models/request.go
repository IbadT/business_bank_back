// models/request.go
package models

import (
	"time"
)

type TransactionType string
type PaymentMethod string

const (
	Income  TransactionType = "income"
	Expense TransactionType = "expense"
)

const (
	ACHCredit         PaymentMethod = "ACH_CREDIT"
	ACHDebit          PaymentMethod = "ACH_DEBIT"
	ElectronicPayment PaymentMethod = "Electronic Payment"
	Card              PaymentMethod = "card"
	Account           PaymentMethod = "account"
	Wire              PaymentMethod = "wire"
	Zelle             PaymentMethod = "zelle"
)

type CustomContractor struct {
	TransactionType string `json:"transactionType"` // [53] transactionType
	Name            string `json:"name"`            // [53] name
}

type CompanyInfo struct {
	OwnerName   string `json:"ownerName"`   // [forwardingInfo] ownerName
	CompanyName string `json:"companyName"` // [forwardingInfo] companyName
}

type ManualTransaction struct {
	TransactionDate time.Time      `json:"transactionDate"`
	Type            TransactionType `json:"type"`
	Category        string         `json:"category"`
	Method          PaymentMethod  `json:"method"`
	Amount          float64        `json:"amount"`
	IsManual        bool           `json:"isManual"` // [44] isManual
}

type CustomData struct {
	ManualTransactions []ManualTransaction `json:"manualTransactions"` // [38] manualTransactions
	CompanyInfo        CompanyInfo         `json:"companyInfo"`        // [forwardingInfo]
	CustomCustomers    []string            `json:"customCustomers"`    // [6][forwardingInfo] customCustomers
	CustomContractors  []CustomContractor  `json:"customContractors"`  // [53] customContractors
}

type GenerateRequest struct {
	Month                int        `json:"month"`                // Из примеров дат [44]
	Year                 int        `json:"year"`                 // Из примеров дат [44]
	Turnover             float64    `json:"turnover"`             // [1][5] оборот
	DesiredProfitPercent float64    `json:"desiredProfitPercent"` // [1] желаемая прибыль
	Model                string     `json:"model"`                // [1][2] B2C или B2B
	InitialBalance       float64    `json:"initialBalance"`       // [43] начальный баланс
	ScaleFactor          int        `json:"scaleFactor"`          // [38] масштабирование
	CustomData           *CustomData `json:"customData,omitempty"` // [38] CustomData
}