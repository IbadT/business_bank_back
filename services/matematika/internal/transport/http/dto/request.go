// internal/transport/http/dto/request.go
package dto

import "time"

// LoginRequest - DTO запроса авторизации
// @Description Параметры для авторизации пользователя
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email" example:"user@example.com"`
	Password string `json:"password" validate:"required,min=8" example:"password123"`
}

// RegisterRequest - DTO запроса регистрации пользователя
// @Description Параметры для регистрации пользователя
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email" example:"user@example.com"`
	Password string `json:"password" validate:"required,min=8" example:"password123"`
}

// GenerateRequest - DTO запроса генерации транзакций
// @Description Параметры для генерации финансовой выписки
type GenerateRequest struct {
	Month                int         `json:"month" validate:"required,min=1,max=12" example:"12" minimum:"1" maximum:"12"`
	Year                 int         `json:"year" validate:"required,min=2000,max=2100" example:"2024" minimum:"2000" maximum:"2100"`
	Turnover             float64     `json:"turnover" validate:"required,gt=0" example:"100000.00" minimum:"0"`
	DesiredProfitPercent float64     `json:"desiredProfitPercent" validate:"required,min=0,max=100" example:"10.0" minimum:"0" maximum:"100"`
	Model                string      `json:"model" validate:"required,oneof=B2C B2B" example:"B2C" enums:"B2C,B2B"`
	InitialBalance       float64     `json:"initialBalance" validate:"required,min=0" example:"300000.00" minimum:"0"`
	ScaleFactor          int         `json:"scaleFactor" validate:"min=1,max=10" example:"1" minimum:"1" maximum:"10"`
	CustomData           *CustomData `json:"customData,omitempty"`
}

// CustomData - дополнительные кастомные данные для генерации
// @Description Дополнительные данные для кастомизации генерации: ручные транзакции, информация о компании, кастомные клиенты и контрагенты
type CustomData struct {
	ManualTransactions []ManualTransaction `json:"manualTransactions"`
	CompanyInfo        CompanyInfo         `json:"companyInfo"`
	CustomCustomers    []string            `json:"customCustomers"`
	CustomContractors  []CustomContractor  `json:"customContractors"`
}

// ManualTransaction - ручная транзакция
// @Description Транзакция, добавленная вручную пользователем
type ManualTransaction struct {
	TransactionDate time.Time `json:"transactionDate" example:"2024-12-15T10:30:00Z" format:"date-time"`
	Type            string    `json:"type" example:"income"`
	Category        string    `json:"category" example:"Sales"`
	Method          string    `json:"method" example:"card"`
	Amount          float64   `json:"amount" example:"15000.00"`
}

// CompanyInfo - информация о компании
// @Description Информация о владельце и названии компании
type CompanyInfo struct {
	OwnerName   string `json:"ownerName" example:"Иван Иванов"`
	CompanyName string `json:"companyName" example:"ООО \"Пример\""`
}

// CustomContractor - кастомный контрагент
// @Description Кастомный контрагент с указанием типа транзакций
type CustomContractor struct {
	TransactionType string `json:"transactionType" example:"income"`
	Name            string `json:"name" example:"ООО \"Партнер\""`
}