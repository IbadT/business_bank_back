// models/config.go
package models

type Holiday struct {
	Date    string `json:"date"`    // [32] дата праздника
	Name    string `json:"name"`    // [32] название праздника
	Country string `json:"country"` // [32] страна
}

type BusinessHours struct {
	Start string `json:"start"` // [33] начало рабочего времени
	End   string `json:"end"`   // [33] конец рабочего времени
}

type TransactionTemplate struct {
	Category        string       `json:"category"`        // Категория операции
	Type            string       `json:"type"`           // income или expense
	IsPercentage    bool         `json:"isPercentage"`   // [7-12] процентная операция
	PercentageMin   float64      `json:"percentageMin"`  // [7-12] минимальный процент
	PercentageMax   float64      `json:"percentageMax"`  // [7-12] максимальный процент
	FixedAmount     float64      `json:"fixedAmount"`    // [13-19] фиксированная сумма
	Frequency       string       `json:"frequency"`      // [22-31] частота
	PreferredDay    string       `json:"preferredDay"`   // [22-31] предпочтительный день
	WeekOfMonth     []int        `json:"weekOfMonth"`    // [7][8] недели месяца (2,4)
	BusinessHours   BusinessHours `json:"businessHours"` // [33] рабочие часы
	IsOptional      bool         `json:"isOptional"`     // [39-41] опциональная категория
	Method          string       `json:"method"`         // Метод платежа
	MinTransactions int          `json:"minTransactions"` // [2][3] мин. транзакций
	MaxTransactions int          `json:"maxTransactions"` // [2][3] макс. транзакций
}

type Gateway struct {
	Name string `json:"name"` // [35] название шлюза
}

type DefaultCustomer struct {
	Name            string  `json:"name"`            // [36] название клиента
	Category        string  `json:"category"`        // [36] категория (retails, wholesale)
	MinPercent      float64 `json:"minPercent"`      // [37] 5.5-8.5%
	MaxPercent      float64 `json:"maxPercent"`      // [37] 5.5-8.5%
	MinTransactions int     `json:"minTransactions"` // [36] 2-8 платежей
	MaxTransactions int     `json:"maxTransactions"` // [36] 2-8 платежей
}