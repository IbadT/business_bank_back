package generatorservice

import (
	"math/rand"
	"strings"
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/internal/domain/entities"
	"github.com/IbadT/business_bank_back/services/matematika/internal/domain/value_objects"
	dateservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/date"
	holidayservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/holiday"
	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/dto"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/helpers"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/logger"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/utils"
)

// expenseGenerator генерирует транзакции расходов используя функции из patterns.go
type expenseGenerator struct {
	dateCalculator       *dateservice.DateCalculator
	holidayService       holidayservice.HolidayService
	fixedAmountCalculator *fixedAmountCalculator
}

// newExpenseGenerator создает новый генератор расходов
func newExpenseGenerator(dateCalculator *dateservice.DateCalculator, holidayService holidayservice.HolidayService, fixedAmountCalculator *fixedAmountCalculator) *expenseGenerator {
	return &expenseGenerator{
		dateCalculator:        dateCalculator,
		holidayService:        holidayService,
		fixedAmountCalculator: fixedAmountCalculator,
	}
}

// GenerateFromPatterns использует функции из patterns.go для соответствующих категорий расходов
// Возвращает (transactions, totalAmount, ok), где ok указывает, была ли использована функция из patterns.go
func (g *expenseGenerator) GenerateFromPatterns(
	req *dto.GenerateRequest,
	template *entities.TransactionTemplate,
	userID *string,
) ([]*entities.Transaction, float64, bool) {
	op := "service.generator.expenseGenerator.generateFromPatterns"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"category": template.Category,
		"template_id": template.ID,
	})
	log.Debug("Generating expenses from patterns")

	category := strings.ToLower(template.Category)

	// Нормализуем название категории для сопоставления
	switch {
	case strings.Contains(category, "аренда") || strings.Contains(category, "rent"):
		return g.generateRent(req, template)

	case strings.Contains(category, "коммунал") || strings.Contains(category, "utilit"):
		return g.generateUtilities(req, template, userID)

	case strings.Contains(category, "страхование") || strings.Contains(category, "insurance"):
		return g.generateInsurance(req, template)

	case strings.Contains(category, "irs") || strings.Contains(category, "налог"):
		return g.generateIRSTaxes(req, template)

	case strings.Contains(category, "payroll"):
		return g.generatePayroll(req, template)

	case strings.Contains(category, "перевод владельцу") || strings.Contains(category, "owner transfer"):
		return g.generateOwnerTransfer(req, template)

	case strings.Contains(category, "saas"):
		return g.generateSaaS(req, template, userID)

	case strings.Contains(category, "equipment") && strings.Contains(category, "lease"):
		return g.generateEquipmentLease(req, template, userID)

	case strings.Contains(category, "бухгалтер") || strings.Contains(category, "accountant"):
		return g.generateAccountant(req, template)

	case strings.Contains(category, "закупк") || strings.Contains(category, "purchas"):
		return g.generatePurchases(req, template)

	case strings.Contains(category, "inbound") && strings.Contains(category, "freight"):
		return g.generateInboundFreight(req, template)

	case strings.Contains(category, "outbound") && strings.Contains(category, "shipping"):
		return g.generateOutboundShipping(req, template)

	case strings.Contains(category, "fuel") || strings.Contains(category, "топливо"):
		return g.generateFuel(req, template)

	case strings.Contains(category, "упаковк") || strings.Contains(category, "packaging"):
		return g.generatePackaging(req, template)

	case strings.Contains(category, "маркетинг") || strings.Contains(category, "marketing"):
		return g.generateMarketing(req, template)

	case strings.Contains(category, "it") && strings.Contains(category, "secur"):
		return g.generateITSecurity(req, template)

	case strings.Contains(category, "usda") && strings.Contains(category, "inspect"):
		return g.generateUSDAInspect(req, template)

	case strings.Contains(category, "demurrage"):
		return g.generateDemurrage(req, template)

	case strings.Contains(category, "pallet") && strings.Contains(category, "fee"):
		return g.generatePalletFee(req, template)

	case strings.Contains(category, "swift") && strings.Contains(category, "fee"):
		return g.generateSwiftFee(req, template)
	}

	// Если категория не подходит под patterns.go, возвращаем false
	log.Debug("Category does not match any pattern, returning false")
	return nil, 0, false
}

func (g *expenseGenerator) generateRent(req *dto.GenerateRequest, template *entities.TransactionTemplate) ([]*entities.Transaction, float64, bool) {
	op := "service.generator.expenseGenerator.generateRent"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"category": template.Category})
	log.Debug("Generating rent transactions")

	rentPercentage := 0.0
	if template.IsPercentage {
		rentPercentage = utils.RandomPercentage(template.PercentageRange.Min, template.PercentageRange.Max)
	}
	rentData := helpers.CalculateRentExpense(req.Turnover, rentPercentage)
	// Устанавливаем дату на 1-е число месяца
	rentDate := utils.FirstDayOfMonth(req.Year, req.Month)
	if rentDate.Weekday() == time.Saturday {
		rentDate = rentDate.AddDate(0, 0, 2)
	} else if rentDate.Weekday() == time.Sunday {
		rentDate = rentDate.AddDate(0, 0, 1)
	}
	if g.holidayService.IsHoliday(rentDate) {
		rentDate = g.holidayService.GetNextBusinessDay(rentDate)
	}
	transactionTime := g.dateCalculator.GenerateBusinessTime(rentDate, 8, 18)
	
	// Сохраняем детали расчёта
	calculationDetails := map[string]interface{}{
		"type":              "percentage_expense",
		"percentage":       utils.FormatPercentage(rentPercentage),
		"percentage_percent": utils.FormatPercentagePercent(rentPercentage),
		"turnover":          req.Turnover,
		"total_amount":      rentData.Amount,
		"formula":           "turnover * percentage",
	}
	
	tx := entities.NewTransaction(
		utils.GenerateTemplateTransactionID(template.Category, 1),
		transactionTime,
		rentDate,
		value_objects.Expense,
		rentData.Category,
		template.PaymentMethod,
		-rentData.Amount,
	)
	tx.SetCalculationDetails(calculationDetails)
	log.WithFields(logger.Fields{"amount": rentData.Amount, "percentage": rentPercentage}).Debug("Rent transaction generated")
	return []*entities.Transaction{tx}, rentData.Amount, true
}

func (g *expenseGenerator) generateUtilities(req *dto.GenerateRequest, template *entities.TransactionTemplate, userID *string) ([]*entities.Transaction, float64, bool) {
	op := "service.generator.expenseGenerator.generateUtilities"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"category": template.Category})
	log.Debug("Generating utilities transactions")

	// [15][16] Коммунальные: фиксируются в первом месяце в диапазоне $200–500 и далее меняются ±15% от этой базы
	// Используем правильную логику из fixedAmountCalculator вместо helpers.CalculateUtilitiesExpense
	utilitiesAmount, calculationDetails := g.fixedAmountCalculator.CalculateFixedAmount(
		value_objects.CategoryUtilitiesRU,
		template.FixedAmount,
		req,
		userID,
	)
	
	// Находим 3-ю пятницу месяца
	utilitiesDate := helpers.GetNthWeekdayInMonth(req.Year, req.Month, time.Friday, 3)
	
	// [32] Проверка праздников для коммунальных (если это операция по счету)
	// Операции по счету (ACH, wire, internal transfers) не проводятся в праздничные дни
	if template.PaymentMethod.IsAccountTransfer() {
		if g.holidayService.IsHoliday(utilitiesDate) {
			utilitiesDate = g.holidayService.GetNextBusinessDay(utilitiesDate)
		}
	}
	
	// [33] Генерация времени: 09:00-20:00 для карт, 08:00-18:00 для операций по счету
	var transactionTime time.Time
	if template.PaymentMethod.IsCardOperation() {
		transactionTime = g.dateCalculator.GenerateBusinessTime(utilitiesDate, 9, 20)
	} else {
		transactionTime = g.dateCalculator.GenerateBusinessTime(utilitiesDate, 8, 18)
	}
	
	tx := entities.NewTransaction(
		utils.GenerateTemplateTransactionID(template.Category, 1),
		transactionTime,
		utilitiesDate,
		value_objects.Expense,
		template.Category,
		template.PaymentMethod,
		-utilitiesAmount,
	)
	tx.SetCalculationDetails(calculationDetails)
	log.WithFields(logger.Fields{"amount": utilitiesAmount}).Debug("Utilities transaction generated")
	return []*entities.Transaction{tx}, utilitiesAmount, true
}

func (g *expenseGenerator) generateInsurance(req *dto.GenerateRequest, template *entities.TransactionTemplate) ([]*entities.Transaction, float64, bool) {
	op := "service.generator.expenseGenerator.generateInsurance"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"category": template.Category})
	log.Debug("Generating insurance transactions")

	insurancePercentage := 0.0
	if template.IsPercentage {
		insurancePercentage = utils.RandomPercentage(template.PercentageRange.Min, template.PercentageRange.Max)
	}
	insuranceData := helpers.CalculateBusinessInsuranceExpense(req.Turnover, insurancePercentage)
	insuranceDate := g.dateCalculator.GenerateRandomBusinessDate(req.Year, req.Month)
	if g.holidayService.IsHoliday(insuranceDate) {
		insuranceDate = g.holidayService.GetNextBusinessDay(insuranceDate)
	}
	transactionTime := g.dateCalculator.GenerateBusinessTime(insuranceDate, 8, 18)
	
	// Сохраняем детали расчёта
	calculationDetails := map[string]interface{}{
		"type":              "percentage_expense",
		"percentage":        utils.FormatPercentage(insurancePercentage),
		"percentage_percent": utils.FormatPercentagePercent(insurancePercentage),
		"turnover":          req.Turnover,
		"total_amount":      insuranceData.Amount,
		"formula":           "turnover * percentage",
	}
	
	tx := entities.NewTransaction(
		utils.GenerateTemplateTransactionID(template.Category, 1),
		transactionTime,
		insuranceDate,
		value_objects.Expense,
		insuranceData.Category,
		template.PaymentMethod,
		-insuranceData.Amount,
	)
	tx.SetCalculationDetails(calculationDetails)
	log.WithFields(logger.Fields{"amount": insuranceData.Amount, "percentage": insurancePercentage}).Debug("Insurance transaction generated")
	return []*entities.Transaction{tx}, insuranceData.Amount, true
}

func (g *expenseGenerator) generateIRSTaxes(req *dto.GenerateRequest, template *entities.TransactionTemplate) ([]*entities.Transaction, float64, bool) {
	op := "service.generator.expenseGenerator.generateIRSTaxes"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"category": template.Category})
	log.Debug("Generating IRS taxes transactions")

	irsPercentage := 0.0
	if template.IsPercentage {
		irsPercentage = utils.RandomPercentage(template.PercentageRange.Min, template.PercentageRange.Max)
	}
	irsData := helpers.CalculateIRSTaxesExpense(req.Turnover, req.Year, req.Month, irsPercentage)
	
	// [23][24] Сохраняем детали расчёта для всех транзакций
	calculationDetails := map[string]interface{}{
		"type":              "percentage_expense",
		"percentage":        utils.FormatPercentage(irsPercentage),
		"percentage_percent": utils.FormatPercentagePercent(irsPercentage),
		"turnover":          req.Turnover,
		"total_amount":      irsData.Amount,
		"transaction_count": irsData.TransactionCount,
		"formula":           "turnover * percentage / transaction_count",
		"is_quarterly":      irsData.TransactionCount == 2,
	}
	
	var transactions []*entities.Transaction
	for i := 0; i < irsData.TransactionCount; i++ {
		// [23][24] Все транзакции IRS налогов должны быть 15-го числа
		// Для квартальных месяцев обе транзакции - 15-го числа (не разные даты)
		txDate := g.dateCalculator.CalculateIRSDate(req.Year, req.Month, i+1)
		
		transactionTime := g.dateCalculator.GenerateBusinessTime(txDate, 8, 18)
		amount := utils.DistributeAmount(irsData.Amount, i, irsData.TransactionCount)
		tx := entities.NewTransaction(
			utils.GenerateTemplateTransactionID(template.Category, i+1),
			transactionTime,
			txDate,
			value_objects.Expense,
			irsData.Category,
			template.PaymentMethod,
			-amount,
		)
		tx.SetCalculationDetails(calculationDetails)
		transactions = append(transactions, tx)
	}
	log.WithFields(logger.Fields{"amount": irsData.Amount, "count": irsData.TransactionCount, "percentage": irsPercentage}).Debug("IRS taxes transactions generated")
	return transactions, irsData.Amount, true
}

func (g *expenseGenerator) generatePayroll(req *dto.GenerateRequest, template *entities.TransactionTemplate) ([]*entities.Transaction, float64, bool) {
	op := "service.generator.expenseGenerator.generatePayroll"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"category": template.Category})
	log.Debug("Generating payroll transactions")

	// [7][8] Payroll ADP: ~27–27.5% от оборота, разбито на две выплаты (2-я и 4-я пятница)
	payrollPercentage := 0.0
	if template.IsPercentage {
		payrollPercentage = utils.RandomPercentage(template.PercentageRange.Min, template.PercentageRange.Max)
	}
	// Fallback: если процент не задан, используем диапазон по умолчанию [7][8]
	if payrollPercentage == 0.0 {
		// [7][8] Диапазон по умолчанию: 27–27.5%
		payrollPercentage = utils.RandomPercentage(0.27, 0.275)
		op := "service.generator.expenseGenerator.generatePayroll"
		log := logger.GetLogger().WithOperation(op)
		log.Warn("Payroll percentage not set in template, using default range 27–27.5%%: %.4f", payrollPercentage)
	}
	// Рассчитываем общую сумму: процент от оборота
	totalPayrollAmount := req.Turnover * payrollPercentage
	totalPayrollAmount = utils.RoundToCents(totalPayrollAmount)

	// [7][8][34] Генерируем транзакции: 2-я и 4-я пятница (при 5 пятницах в месяце добавляем 5-ю)
	firstDay := utils.FirstDayOfMonth(req.Year, req.Month)
	transactionCount := 2
	
	// [34] Учёт пятой недели: если в месяце 5 пятниц, добавляем 5-ю пятницу (2-я, 4-я и 5-я)
	fridaysCount := g.dateCalculator.GetFridaysCount(req.Year, req.Month)
	if fridaysCount == 5 {
		transactionCount = 3 // 2-я, 4-я и 5-я пятница
	}
	
	if len(template.Schedule.WeekOfMonth) > 0 {
		// Если в шаблоне указаны конкретные недели, используем их, но учитываем 5-ю пятницу
		if fridaysCount == 5 && len(template.Schedule.WeekOfMonth) == 2 {
			// Добавляем 5-ю неделю к существующим (2-я, 4-я) -> (2-я, 4-я, 5-я)
			transactionCount = 3
		} else {
			transactionCount = len(template.Schedule.WeekOfMonth)
		}
	}

	// [7][8] Сохраняем детали расчёта для всех транзакций
	calculationDetails := map[string]interface{}{
		"type":              "percentage_expense",
		"percentage":        utils.FormatPercentage(payrollPercentage),
		"percentage_percent": utils.FormatPercentagePercent(payrollPercentage),
		"turnover":          req.Turnover,
		"total_amount":      totalPayrollAmount,
		"transaction_count": transactionCount,
		"formula":           "turnover * percentage / transaction_count",
	}

	var transactions []*entities.Transaction
	var totalAmount float64
	for i := 0; i < transactionCount; i++ {
		var fridayDate time.Time
		var weekNum int
		
		// Определяем номер недели для транзакции
		if len(template.Schedule.WeekOfMonth) > 0 && i < len(template.Schedule.WeekOfMonth) {
			// Используем недели из шаблона (обычно 2-я и 4-я)
			weekNum = template.Schedule.WeekOfMonth[i]
		} else if transactionCount == 3 && fridaysCount == 5 {
			// [34] Если 5 пятниц и 3 транзакции: 2-я, 4-я и 5-я пятница
			if i == 0 {
				weekNum = 2
			} else if i == 1 {
				weekNum = 4
			} else {
				weekNum = 5 // 5-я пятница
			}
		} else {
			// Fallback: 2-я и 4-я пятница (обычный случай, 2 транзакции)
			if i == 0 {
				weekNum = 2
			} else {
				weekNum = 4
			}
		}
		
		// Находим дату пятницы
		if len(template.Schedule.WeekOfMonth) > 0 && i < len(template.Schedule.WeekOfMonth) {
			// Используем PreferredDay из шаблона
			fridayDate = g.dateCalculator.FindNthWeekdayInMonth(firstDay, template.Schedule.PreferredDay, weekNum)
		} else {
			// Используем Friday по умолчанию
			fridayDate = g.dateCalculator.FindNthWeekdayInMonth(firstDay, "Friday", weekNum)
		}

		// [32] Проверка праздников для Payroll (операция по счету - ACH_DEBIT)
		// Если пятница попадает на праздник, переносим на следующий рабочий день
		if g.holidayService.IsHoliday(fridayDate) {
			fridayDate = g.holidayService.GetNextBusinessDay(fridayDate)
		}

		// Распределяем сумму между транзакциями
		var amount float64
		if i == transactionCount-1 {
			// Последняя транзакция: корректируем для точного соответствия
			amount = totalPayrollAmount - totalAmount
			amount = utils.RoundToCents(amount)
		} else {
			amount = totalPayrollAmount / float64(transactionCount)
			amount = utils.RoundToCents(amount)
		}
		totalAmount += amount

		transactionTime := g.dateCalculator.GenerateBusinessTime(fridayDate, 8, 18)
		tx := entities.NewTransaction(
			utils.GenerateTemplateTransactionID(template.Category, i+1),
			transactionTime,
			fridayDate,
			value_objects.Expense,
			template.Category,
			template.PaymentMethod,
			-amount,
		)
		// Устанавливаем детали расчёта для каждой транзакции
		tx.SetCalculationDetails(calculationDetails)
		transactions = append(transactions, tx)
	}
	return transactions, totalAmount, true
}

func (g *expenseGenerator) generateOwnerTransfer(req *dto.GenerateRequest, template *entities.TransactionTemplate) ([]*entities.Transaction, float64, bool) {
	op := "service.generator.expenseGenerator.generateOwnerTransfer"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"category": template.Category})
	log.Debug("Generating owner transfer transactions")

	ownerPercentage := 0.0
	if template.IsPercentage {
		ownerPercentage = utils.RandomPercentage(template.PercentageRange.Min, template.PercentageRange.Max)
	}
	ownerData := helpers.CalculateOwnerTransferExpense(req.Turnover, ownerPercentage)
	ownerDate := g.dateCalculator.GenerateRandomBusinessDate(req.Year, req.Month)
	if g.holidayService.IsHoliday(ownerDate) {
		ownerDate = g.holidayService.GetNextBusinessDay(ownerDate)
	}
	transactionTime := g.dateCalculator.GenerateBusinessTime(ownerDate, 8, 18)
	
	// [22] Сохраняем детали расчёта
	calculationDetails := map[string]interface{}{
		"type":              "percentage_expense",
		"percentage":        utils.FormatPercentage(ownerPercentage),
		"percentage_percent": utils.FormatPercentagePercent(ownerPercentage),
		"turnover":          req.Turnover,
		"total_amount":      ownerData.Amount,
		"formula":           "turnover * percentage",
	}
	
	tx := entities.NewTransaction(
		utils.GenerateTemplateTransactionID(template.Category, 1),
		transactionTime,
		ownerDate,
		value_objects.Expense,
		ownerData.Category,
		template.PaymentMethod,
		-ownerData.Amount,
	)
	tx.SetCalculationDetails(calculationDetails)
	log.WithFields(logger.Fields{"amount": ownerData.Amount, "percentage": ownerPercentage}).Debug("Owner transfer transaction generated")
	return []*entities.Transaction{tx}, ownerData.Amount, true
}

func (g *expenseGenerator) generateSaaS(req *dto.GenerateRequest, template *entities.TransactionTemplate, userID *string) ([]*entities.Transaction, float64, bool) {
	op := "service.generator.expenseGenerator.generateSaaS"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"category": template.Category})
	log.Debug("Generating SaaS subscription transactions")

	// [13][14] Используем fixedAmountCalculator для получения фиксированной суммы из конфигурации
	saasAmount, calculationDetails := g.fixedAmountCalculator.CalculateFixedAmount(
		value_objects.CategorySoftwareSubscriptionRU,
		template.FixedAmount,
		req,
		userID,
	)
	
	// [25][14] Используем calculateSoftwareSubscriptionDate для сохранения дня недели между месяцами
	baseDate := utils.FirstDayOfMonth(req.Year, req.Month)
	userIDUUID := helpers.ParseUUIDOrNil(userID)
	saasDate := g.dateCalculator.CalculateSoftwareSubscriptionDate(baseDate, userIDUUID)
	
	// [33] Время транзакции для подписок: 00:01 (полночь)
	transactionTime := time.Date(
		saasDate.Year(), saasDate.Month(), saasDate.Day(),
		0, 1, 0, 0, time.UTC)
	
	// Используем категорию из шаблона или из helpers
	category := template.Category
	if category == "" {
		category = value_objects.CategorySoftwareSubscriptionRU
	}
	
	tx := entities.NewTransaction(
		utils.GenerateTemplateTransactionID(category, 1),
		transactionTime,
		saasDate,
		value_objects.Expense,
		category,
		template.PaymentMethod,
		-saasAmount,
	)
	
	// Сохраняем детали расчёта из fixedAmountCalculator
	if calculationDetails != nil {
		tx.SetCalculationDetails(calculationDetails)
	}
	
	log.WithFields(logger.Fields{"amount": saasAmount}).Debug("SaaS subscription transaction generated")
	return []*entities.Transaction{tx}, saasAmount, true
}

func (g *expenseGenerator) generateEquipmentLease(req *dto.GenerateRequest, template *entities.TransactionTemplate, userID *string) ([]*entities.Transaction, float64, bool) {
	op := "service.generator.expenseGenerator.generateEquipmentLease"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"category": template.Category})
	log.Debug("Generating equipment lease transactions")

	// [19] Лизинг: первый месяц ~11.5–12% оборота фиксируется и повторяется 1:1 в последующих месяцах
	// Используем fixedAmountCalculator для получения фиксированной суммы (по логике лизинга)
	totalLeaseAmount, calculationDetails := g.fixedAmountCalculator.CalculateFixedAmount(
		value_objects.CategoryEquipmentLease,
		0, // baseAmount не используется для лизинга
		req,
		userID,
	)

	// [30][31] Лизинг: последняя (4-я или 5-я) пятница месяца - всегда 1 транзакция
	transactionCount := 1

	// Обновляем calculationDetails с информацией о количестве транзакций
	if calculationDetails == nil {
		calculationDetails = make(map[string]interface{})
	}
	calculationDetails["transaction_count"] = transactionCount
	calculationDetails["total_amount"] = totalLeaseAmount
	if _, ok := calculationDetails["type"]; !ok {
		calculationDetails["type"] = "fixed_expense"
	}

	// [30][31] Генерируем 1 транзакцию на последней пятнице месяца
	var transactions []*entities.Transaction
	var totalAmount float64
	for i := 0; i < transactionCount; i++ {
		// Используем всю сумму для одной транзакции
		amount := totalLeaseAmount
		amount = utils.RoundToCents(amount)
		totalAmount += amount

		// [30][31] Последняя (4-я или 5-я) пятница месяца
		equipmentDate := g.dateCalculator.GetLastFridayInMonth(req.Year, req.Month)
		if g.holidayService.IsHoliday(equipmentDate) {
			equipmentDate = g.holidayService.GetNextBusinessDay(equipmentDate)
		}
		transactionTime := g.dateCalculator.GenerateBusinessTime(equipmentDate, 8, 18)
		tx := entities.NewTransaction(
			utils.GenerateTemplateTransactionID(template.Category, i+1),
			transactionTime,
			equipmentDate,
			value_objects.Expense,
			template.Category,
			template.PaymentMethod,
			-amount,
		)
		tx.SetCalculationDetails(calculationDetails)
		transactions = append(transactions, tx)
	}
	log.WithFields(logger.Fields{"amount": totalAmount, "count": transactionCount}).Debug("Equipment lease transactions generated")
	return transactions, totalAmount, true
}

func (g *expenseGenerator) generateAccountant(req *dto.GenerateRequest, template *entities.TransactionTemplate) ([]*entities.Transaction, float64, bool) {
	op := "service.generator.expenseGenerator.generateAccountant"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"category": template.Category})
	log.Debug("Generating accountant transactions")

	accountantPercentage := 0.0
	if template.IsPercentage {
		accountantPercentage = utils.RandomPercentage(template.PercentageRange.Min, template.PercentageRange.Max)
	}
	accountantData := helpers.CalculateAccountantExpense(req.Turnover, accountantPercentage)
	accountantDate := g.dateCalculator.GenerateRandomBusinessDate(req.Year, req.Month)
	if g.holidayService.IsHoliday(accountantDate) {
		accountantDate = g.holidayService.GetNextBusinessDay(accountantDate)
	}
	transactionTime := g.dateCalculator.GenerateBusinessTime(accountantDate, 8, 18)
	
	// Сохраняем детали расчёта
	calculationDetails := map[string]interface{}{
		"type":              "percentage_expense",
		"percentage":        utils.FormatPercentage(accountantPercentage),
		"percentage_percent": utils.FormatPercentagePercent(accountantPercentage),
		"turnover":          req.Turnover,
		"total_amount":      accountantData.Amount,
		"formula":           "turnover * percentage",
	}
	
	tx := entities.NewTransaction(
		utils.GenerateTemplateTransactionID(template.Category, 1),
		transactionTime,
		accountantDate,
		value_objects.Expense,
		accountantData.Category,
		template.PaymentMethod,
		-accountantData.Amount,
	)
	tx.SetCalculationDetails(calculationDetails)
	log.WithFields(logger.Fields{"amount": accountantData.Amount, "percentage": accountantPercentage}).Debug("Accountant transaction generated")
	return []*entities.Transaction{tx}, accountantData.Amount, true
}

func (g *expenseGenerator) generatePurchases(req *dto.GenerateRequest, template *entities.TransactionTemplate) ([]*entities.Transaction, float64, bool) {
	op := "service.generator.expenseGenerator.generatePurchases"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"category": template.Category})
	log.Debug("Generating purchases transactions")

	purchasesPercentage := 0.0
	if template.IsPercentage {
		purchasesPercentage = utils.RandomPercentage(template.PercentageRange.Min, template.PercentageRange.Max)
	}
	purchasesData := helpers.CalculatePurchasesExpense(req.Turnover, purchasesPercentage)
	
	// Сохраняем детали расчёта
	calculationDetails := map[string]interface{}{
		"type":              "percentage_expense",
		"percentage":        utils.FormatPercentage(purchasesPercentage),
		"percentage_percent": utils.FormatPercentagePercent(purchasesPercentage),
		"turnover":          req.Turnover,
		"total_amount":      purchasesData.Amount,
		"transaction_count": purchasesData.TransactionCount,
		"formula":           "turnover * percentage / transaction_count",
	}
	
	// Генерируем 15-22 транзакции на будние дни
	var transactions []*entities.Transaction
	for i := 0; i < purchasesData.TransactionCount; i++ {
		amount := utils.DistributeAmount(purchasesData.Amount, i, purchasesData.TransactionCount)
		purchasesDate := g.dateCalculator.GenerateRandomBusinessDate(req.Year, req.Month)
		if g.holidayService.IsHoliday(purchasesDate) {
			purchasesDate = g.holidayService.GetNextBusinessDay(purchasesDate)
		}
		transactionTime := g.dateCalculator.GenerateBusinessTime(purchasesDate, 8, 18)
		tx := entities.NewTransaction(
			utils.GenerateTemplateTransactionID(template.Category, i+1),
			transactionTime,
			purchasesDate,
			value_objects.Expense,
			purchasesData.Category,
			template.PaymentMethod,
			-amount,
		)
		tx.SetCalculationDetails(calculationDetails)
		transactions = append(transactions, tx)
	}
	log.WithFields(logger.Fields{"amount": purchasesData.Amount, "count": purchasesData.TransactionCount, "percentage": purchasesPercentage}).Debug("Purchases transactions generated")
	return transactions, purchasesData.Amount, true
}

func (g *expenseGenerator) generateInboundFreight(req *dto.GenerateRequest, template *entities.TransactionTemplate) ([]*entities.Transaction, float64, bool) {
	op := "service.generator.expenseGenerator.generateInboundFreight"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"category": template.Category})
	log.Debug("Generating inbound freight transactions")

	freightPercentage := 0.0
	if template.IsPercentage {
		freightPercentage = utils.RandomPercentage(template.PercentageRange.Min, template.PercentageRange.Max)
	}
	freightData := helpers.CalculateInboundFreightExpense(req.Turnover, freightPercentage)
	
	// Сохраняем детали расчёта
	calculationDetails := map[string]interface{}{
		"type":              "percentage_expense",
		"percentage":        utils.FormatPercentage(freightPercentage),
		"percentage_percent": utils.FormatPercentagePercent(freightPercentage),
		"turnover":          req.Turnover,
		"total_amount":      freightData.Amount,
		"transaction_count": freightData.TransactionCount,
		"formula":           "turnover * percentage / transaction_count",
	}
	
	// Генерируем 5-7 транзакций на будние дни
	var transactions []*entities.Transaction
	for i := 0; i < freightData.TransactionCount; i++ {
		amount := utils.DistributeAmount(freightData.Amount, i, freightData.TransactionCount)
		freightDate := g.dateCalculator.GenerateRandomBusinessDate(req.Year, req.Month)
		if g.holidayService.IsHoliday(freightDate) {
			freightDate = g.holidayService.GetNextBusinessDay(freightDate)
		}
		transactionTime := g.dateCalculator.GenerateBusinessTime(freightDate, 8, 18)
		tx := entities.NewTransaction(
			utils.GenerateTemplateTransactionID(template.Category, i+1),
			transactionTime,
			freightDate,
			value_objects.Expense,
			freightData.Category,
			template.PaymentMethod,
			-amount,
		)
		tx.SetCalculationDetails(calculationDetails)
		transactions = append(transactions, tx)
	}
	log.WithFields(logger.Fields{"amount": freightData.Amount, "count": freightData.TransactionCount, "percentage": freightPercentage}).Debug("Inbound freight transactions generated")
	return transactions, freightData.Amount, true
}

func (g *expenseGenerator) generateOutboundShipping(req *dto.GenerateRequest, template *entities.TransactionTemplate) ([]*entities.Transaction, float64, bool) {
	op := "service.generator.expenseGenerator.generateOutboundShipping"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"category": template.Category})
	log.Debug("Generating outbound shipping transactions")

	shippingPercentage := 0.0
	if template.IsPercentage {
		shippingPercentage = utils.RandomPercentage(template.PercentageRange.Min, template.PercentageRange.Max)
	}
	shippingData := helpers.CalculateOutboundShippingExpense(req.Turnover, shippingPercentage)
	
	// Сохраняем детали расчёта
	calculationDetails := map[string]interface{}{
		"type":              "percentage_expense",
		"percentage":        utils.FormatPercentage(shippingPercentage),
		"percentage_percent": utils.FormatPercentagePercent(shippingPercentage),
		"turnover":          req.Turnover,
		"total_amount":      shippingData.Amount,
		"transaction_count": shippingData.TransactionCount,
		"formula":           "turnover * percentage / transaction_count",
	}
	
	// Генерируем 3-5 транзакций на будние дни
	var transactions []*entities.Transaction
	for i := 0; i < shippingData.TransactionCount; i++ {
		amount := utils.DistributeAmount(shippingData.Amount, i, shippingData.TransactionCount)
		shippingDate := g.dateCalculator.GenerateRandomBusinessDate(req.Year, req.Month)
		if g.holidayService.IsHoliday(shippingDate) {
			shippingDate = g.holidayService.GetNextBusinessDay(shippingDate)
		}
		transactionTime := g.dateCalculator.GenerateBusinessTime(shippingDate, 8, 18)
		tx := entities.NewTransaction(
			utils.GenerateTemplateTransactionID(template.Category, i+1),
			transactionTime,
			shippingDate,
			value_objects.Expense,
			shippingData.Category,
			template.PaymentMethod,
			-amount,
		)
		tx.SetCalculationDetails(calculationDetails)
		transactions = append(transactions, tx)
	}
	log.WithFields(logger.Fields{"amount": shippingData.Amount, "count": shippingData.TransactionCount, "percentage": shippingPercentage}).Debug("Outbound shipping transactions generated")
	return transactions, shippingData.Amount, true
}

func (g *expenseGenerator) generateFuel(req *dto.GenerateRequest, template *entities.TransactionTemplate) ([]*entities.Transaction, float64, bool) {
	op := "service.generator.expenseGenerator.generateFuel"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"category": template.Category})
	log.Debug("Generating fuel transactions")

	// [9][10] Топливо: ~15–17.5% оборота, разбито на 7–9 транзакций
	fuelPercentage := 0.0
	if template.IsPercentage {
		fuelPercentage = utils.RandomPercentage(template.PercentageRange.Min, template.PercentageRange.Max)
	}
	// Fallback: если процент не задан, используем диапазон по умолчанию [9][10]
	if fuelPercentage == 0.0 {
		// [9][10] Диапазон по умолчанию: 15–17.5%
		fuelPercentage = utils.RandomPercentage(0.15, 0.175)
		op := "service.generator.expenseGenerator.generateFuel"
		log := logger.GetLogger().WithOperation(op)
		log.Warn("Fuel percentage not set in template, using default range 15–17.5%%: %.4f", fuelPercentage)
	}
	// Рассчитываем общую сумму: процент от оборота
	totalFuelAmount := req.Turnover * fuelPercentage
	totalFuelAmount = utils.RoundToCents(totalFuelAmount)

	// [9][10] Количество транзакций: 7-9 (из шаблона или fallback)
	transactionCount := template.Schedule.MinOccurrences
	if template.Schedule.MaxOccurrences > template.Schedule.MinOccurrences {
		transactionCount = template.Schedule.MinOccurrences + rand.Intn(template.Schedule.MaxOccurrences-template.Schedule.MinOccurrences+1)
	}
	// Fallback: если шаблон не настроен, используем диапазон 7-9 транзакций [9][10]
	if transactionCount == 0 || (template.Schedule.MinOccurrences == 0 && template.Schedule.MaxOccurrences == 0) {
		// [9][10] Диапазон по умолчанию: 7-9 транзакций
		transactionCount = 7 + rand.Intn(3) // 7-9 транзакций
		op := "service.generator.expenseGenerator.generateFuel"
		log := logger.GetLogger().WithOperation(op)
		log.Warn("Fuel transaction count not set in template, using default range 7-9: %d", transactionCount)
	}

	// [9][10] Сохраняем детали расчёта для всех транзакций
	calculationDetails := map[string]interface{}{
		"type":              "percentage_expense",
		"percentage":        utils.FormatPercentage(fuelPercentage),
		"percentage_percent": utils.FormatPercentagePercent(fuelPercentage),
		"turnover":          req.Turnover,
		"total_amount":      totalFuelAmount,
		"transaction_count": transactionCount,
		"formula":           "turnover * percentage / transaction_count",
	}

	// Генерируем транзакции на будние дни (операция по карте)
	// [33] Операции по карте могут происходить в праздники (время 09:00-20:00)
	var transactions []*entities.Transaction
	var totalAmount float64
	for i := 0; i < transactionCount; i++ {
		var amount float64
		if i == transactionCount-1 {
			// Последняя транзакция: корректируем для точного соответствия
			amount = totalFuelAmount - totalAmount
			amount = utils.RoundToCents(amount)
		} else {
			amount = totalFuelAmount / float64(transactionCount)
			amount = utils.RoundToCents(amount)
		}
		totalAmount += amount

		// [33] Для операций по карте используем generateRandomWeekdayDate (исключает только выходные, не праздники)
		fuelDate := g.dateCalculator.GenerateRandomWeekdayDate(req.Year, req.Month)
		transactionTime := g.dateCalculator.GenerateBusinessTime(fuelDate, 9, 20)
		tx := entities.NewTransaction(
			utils.GenerateTemplateTransactionID(template.Category, i+1),
			transactionTime,
			fuelDate,
			value_objects.Expense,
			template.Category,
			template.PaymentMethod,
			-amount,
		)
		// Устанавливаем детали расчёта для каждой транзакции
		tx.SetCalculationDetails(calculationDetails)
		transactions = append(transactions, tx)
	}
	log.WithFields(logger.Fields{"amount": totalAmount, "count": transactionCount, "percentage": fuelPercentage}).Debug("Fuel transactions generated")
	return transactions, totalAmount, true
}

func (g *expenseGenerator) generatePackaging(req *dto.GenerateRequest, template *entities.TransactionTemplate) ([]*entities.Transaction, float64, bool) {
	op := "service.generator.expenseGenerator.generatePackaging"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"category": template.Category})
	log.Debug("Generating packaging transactions")

	packagingPercentage := 0.0
	if template.IsPercentage {
		packagingPercentage = utils.RandomPercentage(template.PercentageRange.Min, template.PercentageRange.Max)
	}
	packagingData := helpers.CalculatePackagingExpense(req.Turnover, packagingPercentage)
	
	// Сохраняем детали расчёта
	calculationDetails := map[string]interface{}{
		"type":              "percentage_expense",
		"percentage":        utils.FormatPercentage(packagingPercentage),
		"percentage_percent": utils.FormatPercentagePercent(packagingPercentage),
		"turnover":          req.Turnover,
		"total_amount":      packagingData.Amount,
		"transaction_count": packagingData.TransactionCount,
		"formula":           "turnover * percentage / transaction_count",
	}
	
	// Генерируем 2-3 транзакции на будние дни (операция по карте)
	// [33] Операции по карте могут происходить в праздники (время 09:00-20:00)
	var transactions []*entities.Transaction
	for i := 0; i < packagingData.TransactionCount; i++ {
		amount := utils.DistributeAmount(packagingData.Amount, i, packagingData.TransactionCount)
		// [33] Для операций по карте используем generateRandomWeekdayDate (исключает только выходные, не праздники)
		packagingDate := g.dateCalculator.GenerateRandomWeekdayDate(req.Year, req.Month)
		transactionTime := g.dateCalculator.GenerateBusinessTime(packagingDate, 9, 20)
		tx := entities.NewTransaction(
			utils.GenerateTemplateTransactionID(template.Category, i+1),
			transactionTime,
			packagingDate,
			value_objects.Expense,
			packagingData.Category,
			template.PaymentMethod,
			-amount,
		)
		tx.SetCalculationDetails(calculationDetails)
		transactions = append(transactions, tx)
	}
	log.WithFields(logger.Fields{"amount": packagingData.Amount, "count": packagingData.TransactionCount, "percentage": packagingPercentage}).Debug("Packaging transactions generated")
	return transactions, packagingData.Amount, true
}

func (g *expenseGenerator) generateMarketing(req *dto.GenerateRequest, template *entities.TransactionTemplate) ([]*entities.Transaction, float64, bool) {
	op := "service.generator.expenseGenerator.generateMarketing"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"category": template.Category})
	log.Debug("Generating marketing transactions")

	// [11][12] Маркетинг: ~0.5–0.7% оборота
	marketingPercentage := 0.0
	if template.IsPercentage {
		marketingPercentage = utils.RandomPercentage(template.PercentageRange.Min, template.PercentageRange.Max)
	}
	marketingData := helpers.CalculateMarketingExpense(req.Turnover, marketingPercentage)

	// [11][12] Сохраняем детали расчёта для всех транзакций
	calculationDetails := map[string]interface{}{
		"type":              "percentage_expense",
		"percentage":        utils.FormatPercentage(marketingPercentage),
		"percentage_percent": utils.FormatPercentagePercent(marketingPercentage),
		"turnover":          req.Turnover,
		"total_amount":      marketingData.Amount,
		"transaction_count": marketingData.TransactionCount,
		"formula":           "turnover * percentage / transaction_count",
	}

	// Генерируем 1-2 транзакции на будние дни
	var transactions []*entities.Transaction
	for i := 0; i < marketingData.TransactionCount; i++ {
		amount := utils.DistributeAmount(marketingData.Amount, i, marketingData.TransactionCount)
		marketingDate := g.dateCalculator.GenerateRandomBusinessDate(req.Year, req.Month)
		if g.holidayService.IsHoliday(marketingDate) {
			marketingDate = g.holidayService.GetNextBusinessDay(marketingDate)
		}
		transactionTime := g.dateCalculator.GenerateBusinessTime(marketingDate, 8, 18)
		tx := entities.NewTransaction(
			utils.GenerateTemplateTransactionID(template.Category, i+1),
			transactionTime,
			marketingDate,
			value_objects.Expense,
			marketingData.Category,
			template.PaymentMethod,
			-amount,
		)
		// Устанавливаем детали расчёта для каждой транзакции
		tx.SetCalculationDetails(calculationDetails)
		transactions = append(transactions, tx)
	}
	log.WithFields(logger.Fields{"amount": marketingData.Amount, "count": marketingData.TransactionCount, "percentage": marketingPercentage}).Debug("Marketing transactions generated")
	return transactions, marketingData.Amount, true
}

func (g *expenseGenerator) generateITSecurity(req *dto.GenerateRequest, template *entities.TransactionTemplate) ([]*entities.Transaction, float64, bool) {
	op := "service.generator.expenseGenerator.generateITSecurity"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"category": template.Category})
	log.Debug("Generating IT security transactions")

	itSecurityPercentage := 0.0
	if template.IsPercentage {
		itSecurityPercentage = utils.RandomPercentage(template.PercentageRange.Min, template.PercentageRange.Max)
	}
	itSecurityData := helpers.CalculateITSecurityExpense(req.Turnover, itSecurityPercentage)
	itSecurityDate := g.dateCalculator.GenerateRandomBusinessDate(req.Year, req.Month)
	if g.holidayService.IsHoliday(itSecurityDate) {
		itSecurityDate = g.holidayService.GetNextBusinessDay(itSecurityDate)
	}
	transactionTime := g.dateCalculator.GenerateBusinessTime(itSecurityDate, 8, 18)
	
	// Сохраняем детали расчёта
	calculationDetails := map[string]interface{}{
		"type":              "percentage_expense",
		"percentage":        utils.FormatPercentage(itSecurityPercentage),
		"percentage_percent": utils.FormatPercentagePercent(itSecurityPercentage),
		"turnover":          req.Turnover,
		"total_amount":      itSecurityData.Amount,
		"formula":           "turnover * percentage",
	}
	
	tx := entities.NewTransaction(
		utils.GenerateTemplateTransactionID(template.Category, 1),
		transactionTime,
		itSecurityDate,
		value_objects.Expense,
		itSecurityData.Category,
		template.PaymentMethod,
		-itSecurityData.Amount,
	)
	tx.SetCalculationDetails(calculationDetails)
	log.WithFields(logger.Fields{"amount": itSecurityData.Amount, "percentage": itSecurityPercentage}).Debug("IT security transaction generated")
	return []*entities.Transaction{tx}, itSecurityData.Amount, true
}

func (g *expenseGenerator) generateUSDAInspect(req *dto.GenerateRequest, template *entities.TransactionTemplate) ([]*entities.Transaction, float64, bool) {
	op := "service.generator.expenseGenerator.generateUSDAInspect"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"category": template.Category})
	log.Debug("Generating USDA inspect transactions")

	usdaData := helpers.CalculateUSDAInspectExpense()
	if !usdaData.ShouldAppear {
		// Транзакция не появляется (шанс 20-25%)
		log.Debug("USDA inspect transaction should not appear (20-25% chance)")
		return []*entities.Transaction{}, 0, true
	}
	usdaDate := g.dateCalculator.GenerateRandomBusinessDate(req.Year, req.Month)
	if g.holidayService.IsHoliday(usdaDate) {
		usdaDate = g.holidayService.GetNextBusinessDay(usdaDate)
	}
	transactionTime := g.dateCalculator.GenerateBusinessTime(usdaDate, 8, 18)
	
	// Сохраняем детали расчёта
	calculationDetails := map[string]interface{}{
		"type":         "fixed_expense",
		"fixed_amount": usdaData.Amount,
		"chance":       "20-25%",
	}
	
	tx := entities.NewTransaction(
		utils.GenerateTemplateTransactionID(template.Category, 1),
		transactionTime,
		usdaDate,
		value_objects.Expense,
		usdaData.Category,
		template.PaymentMethod,
		-usdaData.Amount,
	)
	tx.SetCalculationDetails(calculationDetails)
	log.WithFields(logger.Fields{"amount": usdaData.Amount}).Debug("USDA inspect transaction generated")
	return []*entities.Transaction{tx}, usdaData.Amount, true
}

func (g *expenseGenerator) generateDemurrage(req *dto.GenerateRequest, template *entities.TransactionTemplate) ([]*entities.Transaction, float64, bool) {
	op := "service.generator.expenseGenerator.generateDemurrage"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"category": template.Category})
	log.Debug("Generating demurrage transactions")

	demurrageData := helpers.CalculateDemurrageExpense()
	if !demurrageData.ShouldAppear {
		// Транзакция не появляется (шанс 10-15%)
		log.Debug("Demurrage transaction should not appear (10-15% chance)")
		return []*entities.Transaction{}, 0, true
	}
	demurrageDate := g.dateCalculator.GenerateRandomBusinessDate(req.Year, req.Month)
	if g.holidayService.IsHoliday(demurrageDate) {
		demurrageDate = g.holidayService.GetNextBusinessDay(demurrageDate)
	}
	transactionTime := g.dateCalculator.GenerateBusinessTime(demurrageDate, 8, 18)
	
	// Сохраняем детали расчёта
	calculationDetails := map[string]interface{}{
		"type":         "fixed_expense",
		"fixed_amount": demurrageData.Amount,
		"chance":       "10-15%",
	}
	
	tx := entities.NewTransaction(
		utils.GenerateTemplateTransactionID(template.Category, 1),
		transactionTime,
		demurrageDate,
		value_objects.Expense,
		demurrageData.Category,
		template.PaymentMethod,
		-demurrageData.Amount,
	)
	tx.SetCalculationDetails(calculationDetails)
	log.WithFields(logger.Fields{"amount": demurrageData.Amount}).Debug("Demurrage transaction generated")
	return []*entities.Transaction{tx}, demurrageData.Amount, true
}

func (g *expenseGenerator) generatePalletFee(req *dto.GenerateRequest, template *entities.TransactionTemplate) ([]*entities.Transaction, float64, bool) {
	op := "service.generator.expenseGenerator.generatePalletFee"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"category": template.Category})
	log.Debug("Generating pallet fee transactions")

	palletFees := helpers.CalculatePalletFeeExpense()
	// Генерируем 1-2 транзакции, каждая по $5-8
	var transactions []*entities.Transaction
	var totalAmount float64
	for i, palletData := range palletFees {
		palletDate := g.dateCalculator.GenerateRandomBusinessDate(req.Year, req.Month)
		if g.holidayService.IsHoliday(palletDate) {
			palletDate = g.holidayService.GetNextBusinessDay(palletDate)
		}
		transactionTime := g.dateCalculator.GenerateBusinessTime(palletDate, 8, 18)
		
		// Сохраняем детали расчёта
		calculationDetails := map[string]interface{}{
			"type":         "fixed_expense",
			"fixed_amount": palletData.Amount,
			"range":        "$5-8",
		}
		
		tx := entities.NewTransaction(
			utils.GenerateTemplateTransactionID(template.Category, i+1),
			transactionTime,
			palletDate,
			value_objects.Expense,
			palletData.Category,
			template.PaymentMethod,
			-palletData.Amount,
		)
		tx.SetCalculationDetails(calculationDetails)
		transactions = append(transactions, tx)
		totalAmount += palletData.Amount
	}
	log.WithFields(logger.Fields{"amount": totalAmount, "count": len(transactions)}).Debug("Pallet fee transactions generated")
	return transactions, totalAmount, true
}

func (g *expenseGenerator) generateSwiftFee(req *dto.GenerateRequest, template *entities.TransactionTemplate) ([]*entities.Transaction, float64, bool) {
	op := "service.generator.expenseGenerator.generateSwiftFee"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"category": template.Category})
	log.Debug("Generating SWIFT fee transactions")

	// SWIFT Transfer Fee - зависит от количества транзакций "Закупка"
	// Нужно посчитать количество транзакций "Закупка" из уже сгенерированных
	// Но это сложно, так как мы еще не знаем, сколько будет транзакций "Закупка"
	// Поэтому используем template для определения количества
	purchasesCount := template.TransactionRange.Max
	if purchasesCount == 0 {
		purchasesCount = 15 // По умолчанию минимальное количество
	}
	swiftData := helpers.CalculateSwiftFeeExpense(purchasesCount)
	swiftDate := g.dateCalculator.GenerateRandomBusinessDate(req.Year, req.Month)
	if g.holidayService.IsHoliday(swiftDate) {
		swiftDate = g.holidayService.GetNextBusinessDay(swiftDate)
	}
	transactionTime := g.dateCalculator.GenerateBusinessTime(swiftDate, 8, 18)
	
	// Сохраняем детали расчёта
	calculationDetails := map[string]interface{}{
		"type":            "fixed_expense",
		"fixed_amount":    swiftData.Amount,
		"purchases_count": purchasesCount,
		"formula":         "based_on_purchases_count",
	}
	
	tx := entities.NewTransaction(
		utils.GenerateTemplateTransactionID(template.Category, 1),
		transactionTime,
		swiftDate,
		value_objects.Expense,
		swiftData.Category,
		template.PaymentMethod,
		-swiftData.Amount,
	)
	tx.SetCalculationDetails(calculationDetails)
	log.WithFields(logger.Fields{"amount": swiftData.Amount, "purchases_count": purchasesCount}).Debug("SWIFT fee transaction generated")
	return []*entities.Transaction{tx}, swiftData.Amount, true
}
