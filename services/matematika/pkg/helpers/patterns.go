package helpers

import (
	"math"      // Для округления до центов (roundToCents)
	"math/rand" // Для генерации случайных значений согласно паттернам
	"time"      // Для работы с датами транзакций
)

// NumericalCoreResult - результат расчета числового ядра (README строка 1-2)
// Содержит целевую прибыль 6-9% оборота (README строка 202-203, [1]) и целевые расходы
type NumericalCoreResult struct {
	TargetProfit        float64 // Чистая прибыль 6-9% оборота (README строка 202-203, [1])
	TotalExpensesTarget float64 // Целевая сумма расходов (оборот - прибыль)
	ExpensesPercentage  float64 // Процент расходов от оборота
}

// B2CTransactionData - данные транзакции B2C пополнения (README строка 9-10)
// 4 gateway-payout депозита по пятницам: три по 22% ± 3%, последняя 34% ± 4%
type B2CTransactionData struct {
	Amount          float64   // Сумма транзакции (22% ± 3% или 34% ± 4% от оборота)
	Date            time.Time // Дата транзакции (пятница месяца)
	Deviation       float64   // Отклонение от базового процента (±3% или ±4%)
	BasePercentage  float64   // Базовый процент (22% или 34%)
	FinalPercentage float64   // Финальный процент с учетом отклонения
}

// B2BTransactionData - данные транзакции B2B пополнения (README строка 11-17)
// 13-17 пополнений; 70% ± 10% ACH, остальные Electronic; 60% ± 10% компаний повторяются
type B2BTransactionData struct {
	Amount           float64   // Сумма транзакции (6-8.5% от оборота каждая)
	CustomerName     string    // Имя клиента (60% ± 10% повторяются между месяцами)
	TransactionCount int       // Количество транзакций для клиента (1-3)
	PaymentMethod    string    // Метод платежа: ACH-credit (70% ± 10%) или Electronic
	Date             time.Time // Дата транзакции (будний день, определяется на уровне оркестратора)
}

// RentExpenseData - данные расхода на аренду помещения (README строка 21-22)
// 1 транзакция, 1-е число, 3-5% от оборота, операция по счёту
type RentExpenseData struct {
	Amount     float64   // Сумма (3-5% от оборота)
	Date       time.Time // Дата: 1-е число месяца
	Method     string    // Метод: ACH_DEBIT (операция по счёту)
	Category   string    // Категория: "Аренда помещений"
	Percentage float64   // Процент от оборота (3-5%)
}

// UtilitiesExpenseData - данные расхода на коммунальные (README строка 23-25)
// 1 транзакция, 3-я пятница, $1000-2000, операция по карте
type UtilitiesExpenseData struct {
	Amount       float64   // Сумма ($1000-2000)
	Date         time.Time // Дата: 3-я пятница месяца
	Method       string    // Метод: card (операция по карте)
	Category     string    // Категория: "Коммунальные"
	IsFirstMonth bool      // Флаг первого месяца (не используется)
	BaseAmount   float64   // Базовая сумма для расчета
}

// BusinessInsuranceExpenseData - данные расхода на бизнес-страхование (README строка 29-31)
// 1 транзакция, будний день, 1-2% от оборота, операция по счёту
type BusinessInsuranceExpenseData struct {
	Amount     float64   // Сумма (1-2% от оборота)
	Date       time.Time // Дата: будний день
	Method     string    // Метод: ACH_DEBIT (операция по счёту)
	Category   string    // Категория: "Бизнес-страхование"
	Percentage float64   // Процент от оборота (1-2%)
}

// IRSTaxesExpenseData - данные расхода на IRS налоги (README строка 32-33)
// 1 транзакция в обычные месяцы, 2 в квартальные, 2-я среда, 4-6% от оборота
type IRSTaxesExpenseData struct {
	Amount           float64   // Сумма (4-6% от оборота)
	Date             time.Time // Дата: 2-я среда месяца
	Method           string    // Метод: ACH_DEBIT (операция по счёту)
	Category         string    // Категория: "IRS налоги"
	Percentage       float64   // Процент от оборота (4-6%)
	TransactionCount int       // Количество транзакций (1 или 2 в квартальные месяцы)
	IsQuarterly      bool      // Флаг квартального месяца (январь, апрель, июнь, сентябрь)
}

// EquipmentLeaseExpenseData - данные расхода на Equipment lease (README строка 47-49)
// 2-3 транзакции, будние дни, 1.5-2.5% от оборота, операция по счёту
type EquipmentLeaseExpenseData struct {
	Amount           float64   // Сумма (1.5-2.5% от оборота)
	Date             time.Time // Дата: будний день
	Method           string    // Метод: ACH_DEBIT (операция по счёту)
	Category         string    // Категория: "Equipment lease"
	Percentage       float64   // Процент от оборота (1.5-2.5%)
	TransactionCount int       // Количество транзакций (2-3)
}

// AccountantExpenseData - данные расхода на бухгалтера (README строка 59-61)
// 1 транзакция, будние дни, 1-1.5% от оборота, операция по счёту
type AccountantExpenseData struct {
	Amount     float64   // Сумма (1-1.5% от оборота)
	Date       time.Time // Дата: будний день
	Method     string    // Метод: ACH_DEBIT (операция по счёту)
	Category   string    // Категория: "Бухгалтер"
	Percentage float64   // Процент от оборота (1-1.5%)
}

// OwnerTransferExpenseData - данные перевода владельцу (README строка 19-20)
// 1 транзакция, будний день, 2-3% от оборота, операция по счёту
type OwnerTransferExpenseData struct {
	Amount     float64   // Сумма (2-3% от оборота)
	Date       time.Time // Дата: будний день
	Method     string    // Метод: ACH_DEBIT (операция по счёту)
	Category   string    // Категория: "Перевод владельцу"
	Percentage float64   // Процент от оборота (2-3%)
}

// SaaSExpenseData - данные расхода на SaaS (README строка 26-28)
// 1 транзакция, 1-я пятница, $250-600, операция по карте
type SaaSExpenseData struct {
	Amount   float64   // Сумма ($250-600)
	Date     time.Time // Дата: 1-я пятница месяца
	Method   string    // Метод: card (операция по карте)
	Category string    // Категория: "SaaS"
}

// PayrollExpenseData - данные расхода на Payroll ADP (README строка 7-8, 210)
// 2 транзакции, 2-я и 4-я пятница, 27-27.5% от оборота, операция по счёту
type PayrollExpenseData struct {
	Amount           float64   // Сумма (27-27.5% от оборота, распределяется на 2 транзакции)
	Date             time.Time // Дата: 2-я пятница месяца (первая из двух транзакций)
	Method           string    // Метод: ACH_DEBIT (операция по счёту)
	Category         string    // Категория: "Payroll ADP"
	Percentage       float64   // Процент от оборота (27-27.5%)
	TransactionCount int       // Количество транзакций (2, 2-я и 4-я пятница)
}

// PurchasesExpenseData - данные расхода на Закупки (README строка 36-38)
// 15-22 транзакции, будние дни, 45-70% от оборота, операция по счёту
type PurchasesExpenseData struct {
	Amount           float64   // Сумма (45-70% от оборота, распределяется на 15-22 транзакции)
	Date             time.Time // Дата: будний день
	Method           string    // Метод: ACH_DEBIT (операция по счёту)
	Category         string    // Категория: "Закупки"
	Percentage       float64   // Процент от оборота (45-70%)
	TransactionCount int       // Количество транзакций (15-22)
}

// InboundFreightExpenseData - данные расхода на Inbound freight (README строка 39-41)
// 5-7 транзакций, будние дни, 3-5% от оборота, операция по счёту
type InboundFreightExpenseData struct {
	Amount           float64   // Сумма (3-5% от оборота, распределяется на 5-7 транзакций)
	Date             time.Time // Дата: будний день
	Method           string    // Метод: ACH_DEBIT (операция по счёту)
	Category         string    // Категория: "Inbound freight"
	Percentage       float64   // Процент от оборота (3-5%)
	TransactionCount int       // Количество транзакций (5-7)
}

// OutboundShippingExpenseData - данные расхода на Outbound shipping (README строка 42-43)
// 3-5 транзакций, будние дни, 2-3.5% от оборота, операция по счёту
type OutboundShippingExpenseData struct {
	Amount           float64   // Сумма (2-3.5% от оборота, распределяется на 3-5 транзакций)
	Date             time.Time // Дата: будний день
	Method           string    // Метод: ACH_DEBIT (операция по счёту)
	Category         string    // Категория: "Outbound shipping"
	Percentage       float64   // Процент от оборота (2-3.5%)
	TransactionCount int       // Количество транзакций (3-5)
}

// FuelExpenseData - данные расхода на Fuel (README строка 44-46)
// 2-4 транзакции, будние дни, 1-2% от оборота, операция по карте
type FuelExpenseData struct {
	Amount           float64   // Сумма (1-2% от оборота, распределяется на 2-4 транзакции)
	Date             time.Time // Дата: будний день
	Method           string    // Метод: card (операция по карте)
	Category         string    // Категория: "Fuel"
	Percentage       float64   // Процент от оборота (1-2%)
	TransactionCount int       // Количество транзакций (2-4)
}

// PackagingExpenseData - данные расхода на Упаковка (README строка 50-52)
// 2-3 транзакции, будние дни, 0.8-1.5% от оборота, операция по карте
type PackagingExpenseData struct {
	Amount           float64   // Сумма (0.8-1.5% от оборота, распределяется на 2-3 транзакции)
	Date             time.Time // Дата: будний день
	Method           string    // Метод: card (операция по карте)
	Category         string    // Категория: "Упаковка"
	Percentage       float64   // Процент от оборота (0.8-1.5%)
	TransactionCount int       // Количество транзакций (2-3)
}

// MarketingExpenseData - данные расхода на Маркетинг (README строка 53-55)
// 1-2 транзакции, будние дни, 1-2% от оборота, операция по счёту
type MarketingExpenseData struct {
	Amount           float64   // Сумма (1-2% от оборота, распределяется на 1-2 транзакции)
	Date             time.Time // Дата: будний день
	Method           string    // Метод: ACH_DEBIT (операция по счёту)
	Category         string    // Категория: "Маркетинг"
	Percentage       float64   // Процент от оборота (1-2%)
	TransactionCount int       // Количество транзакций (1-2)
}

// ITSecurityExpenseData - данные расхода на IT-security (README строка 56-58)
// 1 транзакция, будний день, 0.5-1% от оборота, операция по счёту
type ITSecurityExpenseData struct {
	Amount     float64   // Сумма (0.5-1% от оборота)
	Date       time.Time // Дата: будний день
	Method     string    // Метод: ACH_DEBIT (операция по счёту)
	Category   string    // Категория: "IT-security"
	Percentage float64   // Процент от оборота (0.5-1%)
}

// USDAInspectExpenseData - данные расхода на USDA INSPEC (README строка 62-64)
// Шанс 20-25% появления 1 транзакции, будние дни, $25-40, операция по счёту
type USDAInspectExpenseData struct {
	Amount       float64   // Сумма ($25-40, если транзакция появляется)
	Date         time.Time // Дата: будний день
	Method       string    // Метод: ACH_DEBIT (операция по счёту)
	Category     string    // Категория: "USDA INSPEC"
	ShouldAppear bool      // Флаг появления транзакции (шанс 20-25%)
}

// DemurrageExpenseData - данные расхода на DEMURRAGE (README строка 65-67)
// Шанс 10-15% появления 1 транзакции, будние дни, $50-70, операция по счёту
type DemurrageExpenseData struct {
	Amount       float64   // Сумма ($50-70, если транзакция появляется)
	Date         time.Time // Дата: будний день
	Method       string    // Метод: ACH_DEBIT (операция по счёту)
	Category     string    // Категория: "DEMURRAGE"
	ShouldAppear bool      // Флаг появления транзакции (шанс 10-15%)
}

// PalletFeeExpenseData - данные расхода на PALLET FEE (README строка 68-69)
// 1-2 транзакции, будние дни, $5-8 каждая, операция по счёту
type PalletFeeExpenseData struct {
	Amount           float64   // Сумма ($5-8 за транзакцию)
	Date             time.Time // Дата: будний день
	Method           string    // Метод: ACH_DEBIT (операция по счёту)
	Category         string    // Категория: "PALLET FEE"
	TransactionCount int       // Количество транзакций (1-2)
}

// SwiftFeeExpenseData - данные расхода на SWIFT Transfer Fee (README строка 70-72)
// $45 после каждой транзакции "Закупка" в пользу компании из importers.csv
type SwiftFeeExpenseData struct {
	Amount      float64   // Сумма ($45 × количество транзакций "Закупка")
	Date        time.Time // Дата: будний день
	Method      string    // Метод: ACH_DEBIT (операция по счёту - FEE)
	Category    string    // Категория: "SWIFT Transfer Fee"
	PerPurchase bool      // Флаг: $45 за каждую транзакцию "Закупка"
}

// SequentialStatementsRules содержит правила для последовательных выписок (README строка 74-78)
// Правила повторяемости контрагентов и сохранения паттернов между месяцами
type SequentialStatementsRules struct {
	// Процент повторяющихся клиентов между месяцами: 60% ± 10% (README строка 75)
	RepeatingCustomersPercent float64 // 0.50 - 0.70 (50% - 70%)

	// Процент повторяющихся подрядчиков между месяцами: 70% ± 10% (README строка 75)
	RepeatingContractorsPercent float64 // 0.60 - 0.80 (60% - 80%)

	// Фиксированные контрагенты, которые неизменны между месяцами (README строка 76)
	FixedContractors []string // Warehouse rent, ERP, Insurance

	// Сохранять шаблоны дней недели при смещении дат (README строка 77)
	PreserveWeekdayPatterns bool

	// Целевая чистая прибыль при пересчете долей: 4-8% (README строка 78)
	TargetProfitPercentMin float64 // 4%
	TargetProfitPercentMax float64 // 8%
}

// CalculateNumericalCore рассчитывает числовое ядро (README строка 1-2)
// Чистая прибыль: 6-9% оборота по умолчанию (README строка 202-203, [1])
// При пользовательских операциях доли пересчитываются для сохранения прибыли
func CalculateNumericalCore(turnover float64, desiredProfitPercent float64) NumericalCoreResult {
	// Рассчитываем целевую прибыль (используется desiredProfitPercent, по умолчанию 6-9% от оборота)
	targetProfit := turnover * (desiredProfitPercent / 100)
	// Рассчитываем целевую сумму расходов (оборот - прибыль)
	totalExpensesTarget := turnover - targetProfit
	// Рассчитываем процент расходов от оборота
	expensesPercentage := (totalExpensesTarget / turnover) * 100

	return NumericalCoreResult{
		TargetProfit:        targetProfit,        // Чистая прибыль (по умолчанию 6-9% оборота)
		TotalExpensesTarget: totalExpensesTarget, // Целевая сумма расходов
		ExpensesPercentage:  expensesPercentage,  // Процент расходов
	}
}

// CalculateB2CReplenishment рассчитывает B2C пополнения (README строка 9-10)
// 4 gateway-payout депозита по пятницам: три по 22% ± 3%, последняя 34% ± 4% → 100% оборота
// Входящие платежи: 4 (B2C) (README строка 4)
func CalculateB2CReplenishment(turnover float64, year int, month int) []B2CTransactionData {
	// Находим все пятницы месяца (README строка 10: "по пятницам")
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	var fridays []time.Time
	current := firstDay
	for current.Month() == time.Month(month) {
		if current.Weekday() == time.Friday {
			fridays = append(fridays, current)
		}
		current = current.AddDate(0, 0, 1)
	}

	// [2][34] Количество входящих транзакций ~4 (редко 5) для B2C-модели
	// [34] Учёт пятой недели: если в месяце 5 пятниц, то "банк" суммы делится на 5 вместо 4
	numTransactions := 4
	if len(fridays) == 5 {
		// [34] Если пятниц 5, автоматически используем все 5 для распределения суммы
		numTransactions = 5
	} else if len(fridays) < 4 {
		numTransactions = len(fridays)
	}

	var results []B2CTransactionData
	totalGenerated := 0.0

	for i := 0; i < numTransactions; i++ {
		transactionDate := fridays[i] // Дата: пятница месяца

		var basePercentage float64
		var maxDeviation float64

		// Для 5 транзакций: первые 4 по 20% ± 2.5%, последняя 20% ± 2.5%
		// Для 4 транзакций: первые 3 по 22% ± 3%, последняя 34% ± 4%
		if numTransactions == 5 {
			if i < numTransactions-1 {
				// Первые 4: 20% ± 2.5% каждая
				basePercentage = 0.20
				maxDeviation = 0.025
			} else {
				// Последняя: 20% ± 2.5% (корректируется до 100%)
				basePercentage = 0.20
				maxDeviation = 0.025
			}
		} else {
			if i < numTransactions-1 {
				// Первые три: 22% ± 3% (README строка 10)
				basePercentage = 0.22
				maxDeviation = 0.03
			} else {
				// Последняя: 34% ± 4% (README строка 10)
				basePercentage = 0.34
				maxDeviation = 0.04
			}
		}

		// Отклонение от базового процента (±3% или ±4%)
		deviation := (rand.Float64()*2 - 1) * maxDeviation // от -maxDeviation до +maxDeviation
		finalPercentage := basePercentage + deviation
		amount := turnover * finalPercentage

		// Корректировка последней транзакции для точного оборота (→ 100% оборота)
		if i == numTransactions-1 {
			amount = turnover - totalGenerated
			finalPercentage = amount / turnover
		} else {
			totalGenerated += amount
		}

		results = append(results, B2CTransactionData{
			Amount:          roundToCents(amount),
			Date:            transactionDate,
			Deviation:       deviation,
			BasePercentage:  basePercentage,
			FinalPercentage: finalPercentage,
		})
	}

	return results
}

// B2BCategoryConfig конфигурация категории B2B источников (README строка 12-16)
// Категории: retail_chains_federal (5-8), retails_штат (3-5), exporter_federal (2-3), marketplaces (0-2)
// Каждая категория: 6-8.5% оборота каждая
type B2BCategoryConfig struct {
	CategoryName  string   // Название категории (retail_chains_federal, retails_штат, и т.д.)
	MinPayments   int      // Минимальное количество платежей для категории
	MaxPayments   int      // Максимальное количество платежей для категории
	MinPercentage float64  // Минимальный процент от оборота (6%)
	MaxPercentage float64  // Максимальный процент от оборота (8.5%)
	Customers     []string // Список клиентов для категории (60% ± 10% повторяются между месяцами)
}

// CalculateB2BReplenishment рассчитывает B2B пополнения согласно паттерну:
// - 10-20 пополнений (гарантировано) [2]
// - Категории источников:
//   - retail_chains_federal: 5-8 платежей, каждый 6-8.5%
//   - retails_штат: 3-5 платежей, каждый 6-8.5%
//   - exporter_federal: 2-3 платежа, каждый 6-8.5%
//   - marketplaces: 0-2 платежа, каждый 6-8.5%
//
// - Методы платежа: 70% ± 10% ACH-credit (60-80%), остальные Electronic (20-40%)
// - Дата: будние дни (конкретная дата определяется на уровне оркестратора)
func CalculateB2BReplenishment(turnover float64, categories []B2BCategoryConfig, year int, month int) []B2BTransactionData {
	if len(categories) == 0 {
		return []B2BTransactionData{}
	}

	var results []B2BTransactionData

	// Дата: базовая дата для транзакций (конкретная дата определяется на уровне оркестратора)
	// Используем 15-е число как базовую точку для распределения транзакций
	baseDate := time.Date(year, time.Month(month), 15, 0, 0, 0, 0, time.UTC)
	if baseDate.Weekday() == time.Saturday {
		baseDate = baseDate.AddDate(0, 0, 2) // Перенос на понедельник
	} else if baseDate.Weekday() == time.Sunday {
		baseDate = baseDate.AddDate(0, 0, 1) // Перенос на понедельник
	}

	// Генерируем платежи для каждой категории
	var allPayments []struct {
		amount        float64
		customerName  string
		date          time.Time
		paymentMethod string
		categoryName  string
	}

	for _, category := range categories {
		// Определяем количество платежей для категории (README строка 12-16)
		// retail_chains_federal: 5-8, retails_штат: 3-5, exporter_federal: 2-3, marketplaces: 0-2
		numPayments := category.MinPayments + rand.Intn(category.MaxPayments-category.MinPayments+1)

		if len(category.Customers) == 0 {
			continue
		}

		// Генерируем процент для каждого платежа (6-8.5% от оборота каждая) (README строка 12-16)
		var categoryPayments []float64
		for i := 0; i < numPayments; i++ {
			percentage := category.MinPercentage + rand.Float64()*(category.MaxPercentage-category.MinPercentage)
			categoryPayments = append(categoryPayments, percentage)
		}

		// Генерируем платежи для категории
		for i, percentage := range categoryPayments {
			// Выбираем клиента
			customerIndex := i % len(category.Customers)
			customerName := category.Customers[customerIndex]

			// Сумма платежа (6-8.5% от оборота)
			amount := turnover * percentage

			// Выбор метода платежа: 70% ± 10% ACH-credit (60-80%), остальные Electronic (20-40%)
			// (README строка 17: "70% ± 10% ACH, остальные Electronic")
			paymentMethod := selectB2BPaymentMethod()

			// Дата: будний день около базовой даты для разнообразия
			transactionDate := baseDate
			if len(allPayments) > 0 || i > 0 {
				dayOffset := rand.Intn(5) - 2 // от -2 до +2 дней
				transactionDate = baseDate.AddDate(0, 0, dayOffset)
				// Убеждаемся, что это будний день
				for transactionDate.Weekday() == time.Saturday || transactionDate.Weekday() == time.Sunday {
					transactionDate = transactionDate.AddDate(0, 0, 1)
				}
				// Убеждаемся, что не вышли за пределы месяца
				if transactionDate.Month() != time.Month(month) {
					transactionDate = baseDate
				}
			}

			allPayments = append(allPayments, struct {
				amount        float64
				customerName  string
				date          time.Time
				paymentMethod string
				categoryName  string
			}{
				amount:        amount,
				customerName:  customerName,
				date:          transactionDate,
				paymentMethod: paymentMethod,
				categoryName:  category.CategoryName,
			})
		}
	}

	// [2] Гарантируем 10-20 пополнений согласно требованиям README (строка 202-203)
	// Входящие платежи: 10-20 для B2B
	targetCount := 10 + rand.Intn(11) // 10-20 пополнений
	currentCount := len(allPayments)

	if currentCount < targetCount {
		// Нужно добавить платежи
		// Собираем всех клиентов из всех категорий
		allCustomers := make(map[string]bool)
		for _, payment := range allPayments {
			allCustomers[payment.customerName] = true
		}
		customerList := make([]string, 0, len(allCustomers))
		for customer := range allCustomers {
			customerList = append(customerList, customer)
		}
		if len(customerList) == 0 {
			// Если нет клиентов, берем из первой категории
			if len(categories) > 0 && len(categories[0].Customers) > 0 {
				customerList = categories[0].Customers
			}
		}

		// Добавляем недостающие платежи для достижения 10-20 пополнений
		for i := currentCount; i < targetCount; i++ {
			customerName := customerList[i%len(customerList)]
			// Каждый платеж: 6-8.5% от оборота (README строка 12-16)
			percentage := 0.06 + rand.Float64()*(0.085-0.06) // 6-8.5%
			amount := turnover * percentage

			paymentMethod := selectB2BPaymentMethod()
			transactionDate := baseDate
			if len(allPayments) > 0 {
				dayOffset := rand.Intn(5) - 2
				transactionDate = baseDate.AddDate(0, 0, dayOffset)
				for transactionDate.Weekday() == time.Saturday || transactionDate.Weekday() == time.Sunday {
					transactionDate = transactionDate.AddDate(0, 0, 1)
				}
				if transactionDate.Month() != time.Month(month) {
					transactionDate = baseDate
				}
			}

			allPayments = append(allPayments, struct {
				amount        float64
				customerName  string
				date          time.Time
				paymentMethod string
				categoryName  string
			}{
				amount:        amount,
				customerName:  customerName,
				date:          transactionDate,
				paymentMethod: paymentMethod,
				categoryName:  "additional",
			})
		}
	} else if currentCount > targetCount {
		// Нужно уменьшить количество платежей до целевого
		// Удаляем случайные платежи, оставляя приоритет категориям с большим количеством
		toRemove := currentCount - targetCount
		for i := 0; i < toRemove && len(allPayments) > targetCount; i++ {
			// Удаляем случайный платеж (можно улучшить логику, удаляя из категорий с переизбытком)
			removeIndex := rand.Intn(len(allPayments))
			allPayments = append(allPayments[:removeIndex], allPayments[removeIndex+1:]...)
		}
	}

	// Нормализуем все платежи, чтобы сумма была 100% оборота
	// Все платежи должны в сумме давать 100% оборота
	totalGenerated := 0.0
	for _, payment := range allPayments {
		totalGenerated += payment.amount
	}

	if len(allPayments) > 0 && totalGenerated > 0 {
		// Масштабируем все платежи пропорционально для достижения 100% оборота
		normalizationFactor := turnover / totalGenerated
		for i := range allPayments {
			allPayments[i].amount *= normalizationFactor
		}

		// Корректировка последнего платежа для точного оборота (100%)
		totalGenerated = 0.0
		for i := 0; i < len(allPayments)-1; i++ {
			totalGenerated += allPayments[i].amount
		}
		allPayments[len(allPayments)-1].amount = turnover - totalGenerated
	}

	// Преобразуем в результат
	for _, payment := range allPayments {
		results = append(results, B2BTransactionData{
			Amount:           roundToCents(payment.amount),
			CustomerName:     payment.customerName,
			TransactionCount: 1,
			PaymentMethod:    payment.paymentMethod,
			Date:             payment.date,
		})
	}

	return results
}

// selectPaymentMethod выбирает метод платежа для B2B (устаревший, используйте selectB2BPaymentMethod)
func selectPaymentMethod() string {
	return selectB2BPaymentMethod()
}

// selectB2BPaymentMethod выбирает метод платежа для B2B согласно паттерну (README строка 7, 17)
// ACH vs Electronic: 70% ± 10% ACH-credit / 30% ± 10% Electronic
// 70% ± 10% ACH, остальные Electronic
func selectB2BPaymentMethod() string {
	// 70% ± 10% = от 60% до 80% для ACH-credit (README строка 7, 17)
	achPercent := 0.60 + rand.Float64()*(0.80-0.60) // 60-80%

	r := rand.Float64()
	if r < achPercent {
		return "ACH-credit" // 60-80% транзакций
	} else {
		// Остальные Electronic: Wire или Zelle (равномерно) (20-40%)
		if rand.Float64() < 0.5 {
			return "Wire"
		}
		return "Zelle"
	}
}

// CalculateRentExpense рассчитывает расход на аренду помещения (README строка 21-22)
// 1 транзакция, 1-е число, 3-5% от оборота, операция по счёту
func CalculateRentExpense(turnover float64, rentPercentage float64) RentExpenseData {
	// Если процент не указан, используем диапазон 3-5% (README строка 22)
	if rentPercentage == 0 {
		rentPercentage = 0.03 + rand.Float64()*(0.05-0.03) // 3-5%
	}

	amount := turnover * rentPercentage
	// Дата: 1-е число месяца (README строка 22: "1‑е число")
	// Будет установлена позже с учетом выходных
	date := time.Time{}

	return RentExpenseData{
		Amount:     roundToCents(amount),
		Date:       date,
		Method:     "ACH_DEBIT", // Операция по счёту (README строка 22)
		Category:   "Аренда помещений",
		Percentage: rentPercentage,
	}
}

// CalculateUtilitiesExpense рассчитывает расход на коммунальные (README строка 23-25)
// 1 транзакция, 3-я пятница, $1000-2000, операция по карте
func CalculateUtilitiesExpense(turnover float64, utilitiesPercentage float64) UtilitiesExpenseData {
	// Согласно требованиям: $1000-2000 (README строка 25: "$1 000 – 2 000")
	// Не зависит от оборота, фиксированная сумма
	amount := 1000.0 + rand.Float64()*(2000.0-1000.0)

	// Дата: 3-я пятница месяца (README строка 25: "3‑я пт")
	// Будет установлена позже
	date := time.Time{}

	return UtilitiesExpenseData{
		Amount:       roundToCents(amount),
		Date:         date,
		Method:       "card", // Операция по карте (README строка 25)
		Category:     "Коммунальные",
		IsFirstMonth: false, // Не используется для коммунальных
		BaseAmount:   roundToCents(amount),
	}
}

// CalculateBusinessInsuranceExpense рассчитывает расход на бизнес-страхование (README строка 29-31)
// 1 транзакция, будний день, 1-2% от оборота, операция по счёту
func CalculateBusinessInsuranceExpense(turnover float64, businessInsurancePercentage float64) BusinessInsuranceExpenseData {
	// Если процент не указан, используем диапазон 1-2% (README строка 31)
	if businessInsurancePercentage == 0 {
		businessInsurancePercentage = 0.01 + rand.Float64()*(0.02-0.01) // 1-2%
	}

	amount := turnover * businessInsurancePercentage
	// Дата: будний день (README строка 31: "Будний")
	// Будет установлена позже
	date := time.Time{}

	return BusinessInsuranceExpenseData{
		Amount:     roundToCents(amount),
		Date:       date,
		Method:     "ACH_DEBIT", // Операция по счёту (README строка 31)
		Category:   "Бизнес-страхование",
		Percentage: businessInsurancePercentage,
	}
}

// CalculateIRSTaxesExpense рассчитывает IRS налоги согласно паттерну:
// - 4-6% от оборота (README строка 33)
// - 1 транзакция в обычные месяцы, 2 транзакции в квартальные (январь, апрель, июнь, сентябрь)
// - Дата: 2-я среда месяца (README строка 33: "2‑я ср")
// - Метод: ACH_DEBIT
func CalculateIRSTaxesExpense(turnover float64, year int, month int, irsPercentage float64) IRSTaxesExpenseData {
	// Если процент не указан, используем диапазон 4-6%
	if irsPercentage == 0 {
		irsPercentage = 0.04 + rand.Float64()*(0.06-0.04) // 4-6%
	}

	// Проверяем, является ли месяц квартальным
	// Квартальные месяцы: январь (1), апрель (4), июнь (6), сентябрь (9)
	isQuarterly := isQuarterlyMonth(month)
	transactionCount := 1 // Обычные месяцы: 1 транзакция
	if isQuarterly {
		transactionCount = 2 // Квартальные месяцы: 2 транзакции
	}

	// Общая сумма для всех транзакций (4-6% от оборота) (README строка 33)
	totalAmount := turnover * irsPercentage

	// Дата: 2-я среда месяца (README строка 33: "2‑я ср")
	// Будет установлена позже
	date := time.Time{}

	return IRSTaxesExpenseData{
		Amount:           roundToCents(totalAmount),
		Date:             date,        // 2-я среда месяца (README строка 33: "2‑я ср")
		Method:           "ACH_DEBIT", // Операция по счёту (README строка 33)
		Category:         "IRS налоги",
		Percentage:       irsPercentage,    // 4-6% от оборота (README строка 33)
		TransactionCount: transactionCount, // 1 или 2 транзакции
		IsQuarterly:      isQuarterly,
	}
}

// CalculateEquipmentLeaseExpense рассчитывает Equipment lease согласно паттерну:
// - 1.5-2.5% от оборота
// - 2-3 транзакции
// - Будние дни
// - Метод: ACH_DEBIT
func CalculateEquipmentLeaseExpense(turnover float64, year int, month int, equipmentLeasePercentage float64) EquipmentLeaseExpenseData {
	// Если процент не указан, используем диапазон 1.5-2.5%
	if equipmentLeasePercentage == 0 {
		equipmentLeasePercentage = 0.015 + rand.Float64()*(0.025-0.015) // 1.5-2.5%
	}

	// Количество транзакций: 2-3 (README строка 49)
	transactionCount := 2 + rand.Intn(2) // 2 или 3 транзакции

	// Общая сумма для всех транзакций (1.5-2.5% от оборота) (README строка 49)
	totalAmount := turnover * equipmentLeasePercentage

	// Дата: будний день (будет установлена позже)
	date := time.Time{}

	return EquipmentLeaseExpenseData{
		Amount:           roundToCents(totalAmount),
		Date:             date,        // Будний день (README строка 49: "Будни")
		Method:           "ACH_DEBIT", // Операция по счёту (README строка 49)
		Category:         "Equipment lease",
		Percentage:       equipmentLeasePercentage, // 1.5-2.5% от оборота (README строка 49)
		TransactionCount: transactionCount,         // 2-3 транзакции
	}
}

// CalculateAccountantExpense рассчитывает расходы на бухгалтера согласно паттерну:
// - 1-1.5% от оборота
// - 1 транзакция в месяц
// - Будние дни
// - Метод: ACH_DEBIT
func CalculateAccountantExpense(turnover float64, accountantPercentage float64) AccountantExpenseData {
	// Если процент не указан, используем диапазон 1-1.5%
	if accountantPercentage == 0 {
		accountantPercentage = 0.01 + rand.Float64()*(0.015-0.01) // 1-1.5%
	}

	amount := turnover * accountantPercentage
	// Дата: будний день (README строка 61: "Будни")
	// Будет установлена позже
	date := time.Time{}

	return AccountantExpenseData{
		Amount:     roundToCents(amount),
		Date:       date,
		Method:     "ACH_DEBIT", // Операция по счёту (README строка 61)
		Category:   "Бухгалтер",
		Percentage: accountantPercentage, // 1-1.5% от оборота (README строка 61)
	}
}

// GetSequentialStatementsRules возвращает правила для последовательных выписок
// согласно паттерну:
// - 60% ± 10% клиентов повторяются между месяцами (50-70%)
// - 70% ± 10% подрядчиков повторяются между месяцами (60-80%)
// - Фиксированные контрагенты (Warehouse rent, ERP, Insurance) неизменны
// - Даты смещаются, шаблоны дней недели сохраняются
// - При ручных операциях доли пересчитываются для чистой прибыли 4-8%
func GetSequentialStatementsRules() SequentialStatementsRules {
	// 60% ± 10% = от 50% до 70% (README строка 75: "60% ± 10% клиентов")
	repeatingCustomersPercent := 0.50 + rand.Float64()*(0.70-0.50)

	// 70% ± 10% = от 60% до 80% (README строка 75: "70% ± 10% подрядчиков")
	repeatingContractorsPercent := 0.60 + rand.Float64()*(0.80-0.60)

	// Фиксированные контрагенты (README строка 76: "Warehouse rent, ERP, Insurance")
	fixedContractors := []string{
		"Warehouse rent",
		"ERP",
		"Insurance",
	}

	return SequentialStatementsRules{
		RepeatingCustomersPercent:   repeatingCustomersPercent,   // 60% ± 10% клиентов
		RepeatingContractorsPercent: repeatingContractorsPercent, // 70% ± 10% подрядчиков
		FixedContractors:            fixedContractors,            // Warehouse rent, ERP, Insurance
		PreserveWeekdayPatterns:     true,                        // Даты смещаются, шаблоны дней недели сохраняются (README строка 77)
		TargetProfitPercentMin:      4.0,                         // 4% (README строка 78)
		TargetProfitPercentMax:      8.0,                         // 8% (README строка 78)
	}
}

// CalculatePurchasesExpense рассчитывает Закупки согласно паттерну:
// - 45-70% от оборота
// - 15-22 транзакции
// - Будние дни
// - Метод: ACH_DEBIT
func CalculatePurchasesExpense(turnover float64, purchasesPercentage float64) PurchasesExpenseData {
	// Если процент не указан, используем диапазон 45-70%
	if purchasesPercentage == 0 {
		purchasesPercentage = 0.45 + rand.Float64()*(0.70-0.45) // 45-70%
	}

	// Количество транзакций: 15-22 (README строка 38)
	transactionCount := 15 + rand.Intn(8) // 15-22 транзакции

	// Общая сумма для всех транзакций (45-70% от оборота) (README строка 38)
	totalAmount := turnover * purchasesPercentage

	// Дата: будний день (будет установлена позже)
	date := time.Time{}

	return PurchasesExpenseData{
		Amount:           roundToCents(totalAmount),
		Date:             date,        // Будний день (README строка 38: "Будни")
		Method:           "ACH_DEBIT", // Операция по счёту (README строка 38)
		Category:         "Закупки",
		Percentage:       purchasesPercentage, // 45-70% от оборота (README строка 38)
		TransactionCount: transactionCount,    // 15-22 транзакции
	}
}

// CalculateInboundFreightExpense рассчитывает Inbound freight согласно паттерну:
// - 3-5% от оборота
// - 5-7 транзакций
// - Будние дни
// - Метод: ACH_DEBIT
func CalculateInboundFreightExpense(turnover float64, inboundFreightPercentage float64) InboundFreightExpenseData {
	// Если процент не указан, используем диапазон 3-5%
	if inboundFreightPercentage == 0 {
		inboundFreightPercentage = 0.03 + rand.Float64()*(0.05-0.03) // 3-5%
	}

	// Количество транзакций: 5-7 (README строка 41)
	transactionCount := 5 + rand.Intn(3) // 5-7 транзакций

	// Общая сумма для всех транзакций (3-5% от оборота) (README строка 41)
	totalAmount := turnover * inboundFreightPercentage

	// Дата: будний день (будет установлена позже)
	date := time.Time{}

	return InboundFreightExpenseData{
		Amount:           roundToCents(totalAmount),
		Date:             date,        // Будний день (README строка 41: "Будни")
		Method:           "ACH_DEBIT", // Операция по счёту (README строка 41)
		Category:         "Inbound freight",
		Percentage:       inboundFreightPercentage, // 3-5% от оборота (README строка 41)
		TransactionCount: transactionCount,         // 5-7 транзакций
	}
}

// CalculateOutboundShippingExpense рассчитывает Outbound shipping согласно паттерну:
// - 2-3.5% от оборота
// - 3-5 транзакций
// - Будние дни
// - Метод: ACH_DEBIT
func CalculateOutboundShippingExpense(turnover float64, outboundShippingPercentage float64) OutboundShippingExpenseData {
	// Если процент не указан, используем диапазон 2-3.5%
	if outboundShippingPercentage == 0 {
		outboundShippingPercentage = 0.02 + rand.Float64()*(0.035-0.02) // 2-3.5%
	}

	// Количество транзакций: 3-5 (README строка 43)
	transactionCount := 3 + rand.Intn(3) // 3-5 транзакций

	// Общая сумма для всех транзакций (2-3.5% от оборота) (README строка 43)
	totalAmount := turnover * outboundShippingPercentage

	// Дата: будний день (будет установлена позже)
	date := time.Time{}

	return OutboundShippingExpenseData{
		Amount:           roundToCents(totalAmount),
		Date:             date,        // Будний день (README строка 43: "Будни")
		Method:           "ACH_DEBIT", // Операция по счёту (README строка 43)
		Category:         "Outbound shipping",
		Percentage:       outboundShippingPercentage, // 2-3.5% от оборота (README строка 43)
		TransactionCount: transactionCount,           // 3-5 транзакций
	}
}

// CalculateFuelExpense рассчитывает Fuel согласно паттерну:
// - 15-17.5% от оборота [9][10]
// - 7-9 транзакций [9][10]
// - Будние дни
// - Метод: card
func CalculateFuelExpense(turnover float64, fuelPercentage float64) FuelExpenseData {
	// [9][10] Если процент не указан, используем диапазон 15-17.5%
	if fuelPercentage == 0 {
		fuelPercentage = 0.15 + rand.Float64()*(0.175-0.15) // 15-17.5%
	}

	// [9][10] Количество транзакций: 7-9
	transactionCount := 7 + rand.Intn(3) // 7-9 транзакций

	// [9][10] Общая сумма для всех транзакций (15-17.5% от оборота)
	totalAmount := turnover * fuelPercentage

	// Дата: будний день (будет установлена позже)
	date := time.Time{}

	return FuelExpenseData{
		Amount:           roundToCents(totalAmount),
		Date:             date,   // Будний день
		Method:           "card", // Операция по карте
		Category:         "Fuel",
		Percentage:       fuelPercentage,   // 15-17.5% от оборота [9][10]
		TransactionCount: transactionCount, // 7-9 транзакций [9][10]
	}
}

// CalculatePackagingExpense рассчитывает Упаковка согласно паттерну:
// - 0.8-1.5% от оборота
// - 2-3 транзакции
// - Будние дни
// - Метод: card
func CalculatePackagingExpense(turnover float64, packagingPercentage float64) PackagingExpenseData {
	// Если процент не указан, используем диапазон 0.8-1.5%
	if packagingPercentage == 0 {
		packagingPercentage = 0.008 + rand.Float64()*(0.015-0.008) // 0.8-1.5%
	}

	// Количество транзакций: 2-3 (README строка 52)
	transactionCount := 2 + rand.Intn(2) // 2 или 3 транзакции

	// Общая сумма для всех транзакций (0.8-1.5% от оборота) (README строка 52)
	totalAmount := turnover * packagingPercentage

	// Дата: будний день (будет установлена позже)
	date := time.Time{}

	return PackagingExpenseData{
		Amount:           roundToCents(totalAmount),
		Date:             date,   // Будний день (README строка 52: "Будни")
		Method:           "card", // Операция по карте (README строка 52)
		Category:         "Упаковка",
		Percentage:       packagingPercentage, // 0.8-1.5% от оборота (README строка 52)
		TransactionCount: transactionCount,    // 2-3 транзакции
	}
}

// CalculateMarketingExpense рассчитывает Маркетинг согласно паттерну:
// - 0.5-0.7% от оборота [11][12]
// - 1-2 транзакции
// - Будние дни
// - Метод: ACH_DEBIT
func CalculateMarketingExpense(turnover float64, marketingPercentage float64) MarketingExpenseData {
	// [11][12] Если процент не указан, используем диапазон 0.5-0.7%
	if marketingPercentage == 0 {
		marketingPercentage = 0.005 + rand.Float64()*(0.007-0.005) // 0.5-0.7%
	}

	// Количество транзакций: 1-2 (README строка 55)
	transactionCount := 1 + rand.Intn(2) // 1 или 2 транзакции

	// [11][12] Общая сумма для всех транзакций (0.5-0.7% от оборота)
	totalAmount := turnover * marketingPercentage

	// Дата: будний день (будет установлена позже)
	date := time.Time{}

	return MarketingExpenseData{
		Amount:           roundToCents(totalAmount),
		Date:             date,        // Будний день (README строка 55: "Будни")
		Method:           "ACH_DEBIT", // Операция по счёту (README строка 55)
		Category:         "Маркетинг",
		Percentage:       marketingPercentage, // 0.5-0.7% от оборота [11][12]
		TransactionCount: transactionCount,    // 1-2 транзакции
	}
}

// CalculateITSecurityExpense рассчитывает IT-security согласно паттерну:
// - 0.5-1% от оборота
// - 1 транзакция
// - Будний день
// - Метод: ACH_DEBIT
func CalculateITSecurityExpense(turnover float64, itSecurityPercentage float64) ITSecurityExpenseData {
	// Если процент не указан, используем диапазон 0.5-1%
	if itSecurityPercentage == 0 {
		itSecurityPercentage = 0.005 + rand.Float64()*(0.01-0.005) // 0.5-1%
	}

	amount := turnover * itSecurityPercentage
	// Дата: будний день (README строка 58: "Будний")
	// Будет установлена позже
	date := time.Time{}

	return ITSecurityExpenseData{
		Amount:     roundToCents(amount),
		Date:       date,
		Method:     "ACH_DEBIT", // Операция по счёту (README строка 58)
		Category:   "IT-security",
		Percentage: itSecurityPercentage, // 0.5-1% от оборота (README строка 58)
	}
}

// CalculateUSDAInspectExpense рассчитывает USDA INSPEC согласно паттерну:
// - Шанс 20-25% появления 1 транзакции на выписку
// - $25-40
// - Будние дни
// - Метод: ACH_DEBIT
func CalculateUSDAInspectExpense() USDAInspectExpenseData {
	// Шанс 20-25%: выбираем случайный процент в диапазоне для каждого вызова (README строка 64)
	chance := 0.20 + rand.Float64()*(0.25-0.20) // 20-25%
	shouldAppear := rand.Float64() < chance

	amount := 0.0
	if shouldAppear {
		// Сумма: $25-40 (README строка 64)
		amount = 25.0 + rand.Float64()*(40.0-25.0) // $25-40
	}

	// Дата: будний день (будет установлена позже)
	date := time.Time{}

	return USDAInspectExpenseData{
		Amount:       roundToCents(amount),
		Date:         date,        // Будний день (README строка 64: "Будни")
		Method:       "ACH_DEBIT", // Операция по счёту (README строка 64)
		Category:     "USDA INSPEC",
		ShouldAppear: shouldAppear, // Шанс 20-25% появления (README строка 64)
	}
}

// CalculateDemurrageExpense рассчитывает DEMURRAGE согласно паттерну:
// - Шанс 10-15% появления 1 транзакции на выписку
// - $50-70
// - Будние дни
// - Метод: ACH_DEBIT
func CalculateDemurrageExpense() DemurrageExpenseData {
	// Шанс 10-15%: выбираем случайный процент в диапазоне для каждого вызова (README строка 67)
	chance := 0.10 + rand.Float64()*(0.15-0.10) // 10-15%
	shouldAppear := rand.Float64() < chance

	amount := 0.0
	if shouldAppear {
		// Сумма: $50-70 (README строка 67)
		amount = 50.0 + rand.Float64()*(70.0-50.0) // $50-70
	}

	// Дата: будний день (будет установлена позже)
	date := time.Time{}

	return DemurrageExpenseData{
		Amount:       roundToCents(amount),
		Date:         date,        // Будний день (README строка 67: "Будни")
		Method:       "ACH_DEBIT", // Операция по счёту (README строка 67)
		Category:     "DEMURRAGE",
		ShouldAppear: shouldAppear, // Шанс 10-15% появления (README строка 67)
	}
}

// CalculatePalletFeeExpense рассчитывает PALLET FEE согласно паттерну:
// - 1-2 транзакции, каждая по $5-8
// - Будние дни
// - Метод: ACH_DEBIT
// Возвращает массив транзакций, каждая по $5-8
func CalculatePalletFeeExpense() []PalletFeeExpenseData {
	// Количество транзакций: 1-2 (README строка 69)
	transactionCount := 1 + rand.Intn(2) // 1 или 2 транзакции

	var results []PalletFeeExpenseData

	// Генерируем отдельные транзакции, каждая по $5-8 (README строка 69)
	for i := 0; i < transactionCount; i++ {
		// Каждая транзакция отдельно $5-8
		amountPerTransaction := 5.0 + rand.Float64()*(8.0-5.0) // $5-8 за транзакцию

		// Дата: будний день (будет установлена позже)
		date := time.Time{}

		results = append(results, PalletFeeExpenseData{
			Amount:           roundToCents(amountPerTransaction),
			Date:             date,        // Будний день (README строка 69: "Будни")
			Method:           "ACH_DEBIT", // Операция по счёту (README строка 69)
			Category:         "PALLET FEE",
			TransactionCount: 1, // Каждая транзакция отдельно
		})
	}

	return results
}

// CalculateSwiftFeeExpense рассчитывает SWIFT Transfer Fee согласно паттерну:
// - $45 после каждой транзакции по параметру "Закупка"
// - Метод: ACH_DEBIT (FEE)
func CalculateSwiftFeeExpense(purchasesCount int) SwiftFeeExpenseData {
	// $45 за каждую транзакцию "Закупка" (README строка 72)
	// После каждой транзакции по параметру "Закупка" в пользу компании из importers.csv
	amount := 45.0 * float64(purchasesCount)

	// Дата: будний день (будет установлена позже)
	date := time.Time{}

	return SwiftFeeExpenseData{
		Amount:      roundToCents(amount),
		Date:        date,        // Будний день
		Method:      "ACH_DEBIT", // Операция по счёту - FEE (README строка 71)
		Category:    "SWIFT Transfer Fee",
		PerPurchase: true, // $45 после каждой транзакции "Закупка" (README строка 72)
	}
}

// CalculateOwnerTransferExpense рассчитывает перевод владельцу согласно паттерну:
// - 2-3% от оборота
// - 1 транзакция
// - Будний день
// - Метод: ACH_DEBIT
func CalculateOwnerTransferExpense(turnover float64, ownerTransferPercentage float64) OwnerTransferExpenseData {
	// Если процент не указан, используем диапазон 2-3%
	if ownerTransferPercentage == 0 {
		ownerTransferPercentage = 0.02 + rand.Float64()*(0.03-0.02) // 2-3%
	}

	amount := turnover * ownerTransferPercentage
	// Дата: будний день (README строка 20: "Будний")
	// Будет установлена позже
	date := time.Time{}

	return OwnerTransferExpenseData{
		Amount:     roundToCents(amount),
		Date:       date,
		Method:     "ACH_DEBIT", // Операция по счёту (README строка 20)
		Category:   "Перевод владельцу",
		Percentage: ownerTransferPercentage, // 2-3% от оборота (README строка 20)
	}
}

// CalculateSaaSExpense рассчитывает SaaS расходы согласно паттерну:
// - $250-600
// - 1 транзакция
// - 1-я пятница месяца
// - Метод: card
func CalculateSaaSExpense(saasAmount float64) SaaSExpenseData {
	// Если сумма не указана, используем диапазон $250-600 (README строка 28)
	if saasAmount == 0 {
		saasAmount = 250.0 + rand.Float64()*(600.0-250.0)
	}

	// Дата: 1-я пятница месяца (README строка 28: "1‑я пт")
	// Будет установлена позже
	date := time.Time{}

	return SaaSExpenseData{
		Amount:   roundToCents(saasAmount),
		Date:     date,
		Method:   "card", // Операция по карте (README строка 28)
		Category: "SaaS",
	}
}

// CalculatePayrollExpense рассчитывает Payroll ADP согласно паттерну:
// - 27-27.5% от оборота (README строка 7-8, 206)
// - 2 транзакции (README строка 7-8, 210)
// - 2-я и 4-я пятница месяца (README строка 7-8, 210)
// - Метод: ACH_DEBIT (операция по счёту)
func CalculatePayrollExpense(turnover float64, year int, month int, payrollPercentage float64) PayrollExpenseData {
	// Если процент не указан, используем диапазон 27-27.5% (README строка 7-8, 206)
	if payrollPercentage == 0 {
		payrollPercentage = 0.27 + rand.Float64()*(0.275-0.27) // 27-27.5%
	}

	// [7][8] Генерируем 2 транзакции: 2-я и 4-я пятница (README строка 210)
	transactionCount := 2

	// Находим 2-ю пятницу месяца (README строка 7-8, 210)
	date := GetNthWeekdayInMonth(year, month, time.Friday, 2)

	// Общая сумма для всех транзакций (27-27.5% от оборота) (README строка 7-8, 206)
	totalAmount := turnover * payrollPercentage

	return PayrollExpenseData{
		Amount:           roundToCents(totalAmount),
		Date:             date,        // 2-я пятница (первая из двух транзакций)
		Method:           "ACH_DEBIT", // Операция по счёту
		Category:         "Payroll ADP",
		Percentage:       payrollPercentage, // 27-27.5% от оборота (README строка 7-8, 206)
		TransactionCount: transactionCount,  // 2 транзакции, 2-я и 4-я пятница (README строка 7-8, 210)
	}
}

// isQuarterlyMonth проверяет, является ли месяц квартальным
// Квартальные месяцы: январь (1), апрель (4), июнь (6), сентябрь (9)
// Используется для IRS налогов: 2 транзакции в квартальные месяцы вместо 1 (README строка 33)
func isQuarterlyMonth(month int) bool {
	quarterlyMonths := []int{1, 4, 6, 9} // Январь, апрель, июнь, сентябрь
	for _, m := range quarterlyMonths {
		if month == m {
			return true
		}
	}
	return false
}

// roundToCents округляет сумму до центов
// Используется для всех денежных сумм в транзакциях для точности до центов
func roundToCents(amount float64) float64 {
	return math.Round(amount*100) / 100 // Округление до 2 знаков после запятой
}
