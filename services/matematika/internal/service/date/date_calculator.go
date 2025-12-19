package dateservice

import (
	"math/rand"
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/internal/domain"
	"github.com/IbadT/business_bank_back/services/matematika/internal/domain/entities"
	"github.com/IbadT/business_bank_back/services/matematika/internal/domain/value_objects"
	"github.com/IbadT/business_bank_back/services/matematika/internal/repository"
	baseamountservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/base"
	holidayservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/holiday"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/helpers"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/logger"
	"github.com/google/uuid"
)

type DateCalculator struct {
	holidays              []*domain.Holiday
	holidayMap            map[string]bool
	stateRepo             repository.StateRepository
	holidayService        holidayservice.HolidayService
	baseAmountService     baseamountservice.BaseAmountService
	generationRequestRepo repository.GenerationRequestRepository
}

func NewDateCalculator(holidays []*domain.Holiday, stateRepo repository.StateRepository, holidayService holidayservice.HolidayService, baseAmountService baseamountservice.BaseAmountService, generationRequestRepo repository.GenerationRequestRepository) *DateCalculator {
	holidayMap := make(map[string]bool)
	for _, holiday := range holidays {
		dateStr := holiday.HolidayDate
		holidayMap[dateStr] = true
	}

	return &DateCalculator{
		holidays:              holidays,
		holidayMap:            holidayMap,
		stateRepo:             stateRepo,
		holidayService:        holidayService,
		baseAmountService:     baseAmountService,
		generationRequestRepo: generationRequestRepo,
	}
}

// isQuarterlyMonth проверяет, является ли месяц квартальным [23][24]
// Квартальные месяцы: январь (1), апрель (4), июнь (6), сентябрь (9)
func (dc *DateCalculator) IsQuarterlyMonth(month int) bool {
	op := "service.date.isQuarterlyMonth"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"month": month})
	log.Debug("Checking if month is quarterly")

	quarterlyMonths := []int{1, 4, 6, 9} // Январь, Апрель, Июнь, Сентябрь
	for _, m := range quarterlyMonths {
		if month == m {
			log.Debug("Month is quarterly")
			return true
		}
	}
	log.Debug("Month is not quarterly")
	return false
}

// calculateIRSDate рассчитывает дату для IRS налогов [23][24]
// Всегда возвращает 15-е число месяца (или следующий рабочий день, если 15-е - выходной/праздник)
// seqNum не используется, так как обе транзакции в квартальный месяц должны быть 15-го числа
func (dc *DateCalculator) CalculateIRSDate(year, month int, seqNum int) time.Time {
	op := "service.date.calculateIRSDate"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"year": year, "month": month})
	log.Debug("Calculating IRS date")

	// [23][24] Всегда 15-е число
	date := time.Date(year, time.Month(month), 15, 0, 0, 0, 0, time.UTC)

	// Если 15-е число попадает на выходной или на праздник - переносим на следующий рабочий день
	if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday || dc.IsHoliday(date) {
		log.Debug("15th is weekend/holiday, moving to next business day")
		date = dc.GetNextBusinessDay(date)
	}

	log.WithFields(logger.Fields{"date": date.Format("2006-01-02")}).Debug("IRS date calculated")
	return date
}

func (dc *DateCalculator) IsHoliday(date time.Time) bool {
	op := "service.date.isHoliday"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"date": date.Format("2006-01-02")})
	log.Debug("Checking if date is holiday")

	// Используем HolidayService для проверки праздников из БД
	if dc.holidayService != nil {
		result := dc.holidayService.IsHoliday(date)
		log.WithFields(logger.Fields{"is_holiday": result}).Debug("Holiday check completed via service")
		return result
	}
	// Fallback на локальную карту из конфига
	dateStr := date.Format("2006-01-02")
	result := dc.holidayMap[dateStr]
	log.WithFields(logger.Fields{"is_holiday": result}).Debug("Holiday check completed via map")
	return result
}

func (dc *DateCalculator) GetFridaysInMonth(year, month int) []time.Time {
	op := "service.date.getFridaysInMonth"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"year": year, "month": month})
	log.Debug("Getting Fridays in month")
	fridays := dc.GetWeekdaysInMonth(year, month, time.Friday)
	log.WithFields(logger.Fields{"count": len(fridays)}).Debug("Fridays retrieved")
	return fridays
}

// getFridaysCount возвращает количество пятниц в месяце [34]
func (dc *DateCalculator) GetFridaysCount(year, month int) int {
	op := "service.date.getFridaysCount"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"year": year, "month": month})
	log.Debug("Getting Fridays count")
	count := len(dc.GetFridaysInMonth(year, month))
	log.WithFields(logger.Fields{"count": count}).Debug("Fridays count retrieved")
	return count
}

// getLastFridayInMonth возвращает последнюю пятницу месяца (4-ю или 5-ю) [30][31]
func (dc *DateCalculator) GetLastFridayInMonth(year, month int) time.Time {
	op := "service.date.getLastFridayInMonth"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"year": year, "month": month})
	log.Debug("Getting last Friday in month")
	fridays := dc.GetFridaysInMonth(year, month)
	if len(fridays) == 0 {
		// Если нет пятниц (крайне маловероятно), возвращаем последний день месяца
		lastDay := time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.UTC)
		log.Warn("No Fridays found, returning last day of month")
		return lastDay
	}
	lastFriday := fridays[len(fridays)-1]
	log.WithFields(logger.Fields{"date": lastFriday.Format("2006-01-02")}).Debug("Last Friday retrieved")
	return lastFriday
}

func (dc *DateCalculator) GetWeekdaysInMonth(year, month int, weekday time.Weekday) []time.Time {
	op := "service.date.getWeekdaysInMonth"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"year":    year,
		"month":   month,
		"weekday": weekday.String(),
	})
	log.Debug("Getting weekdays in month")
	var days []time.Time
	current := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)

	for current.Month() == time.Month(month) {
		if current.Weekday() == weekday {
			days = append(days, current)
		}
		current = current.AddDate(0, 0, 1)
	}

	log.WithFields(logger.Fields{"count": len(days)}).Debug("Weekdays retrieved")
	return days
}

// generateRandomBusinessDate генерирует случайную дату буднего дня (исключает выходные и праздники)
// Используется для операций по счету (ACH, wire, internal transfers)
func (dc *DateCalculator) GenerateRandomBusinessDate(year, month int) time.Time {
	op := "service.date.generateRandomBusinessDate"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"year": year, "month": month})
	log.Debug("Generating random business date")
	daysInMonth := time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.UTC).Day()

	// Собираем все рабочие дни месяца (исключаем выходные и праздники)
	var businessDays []time.Time
	for day := 1; day <= daysInMonth; day++ {
		date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		if date.Weekday() != time.Saturday && date.Weekday() != time.Sunday && !dc.IsHoliday(date) {
			businessDays = append(businessDays, date)
		}
	}

	// Если есть рабочие дни, выбираем случайный
	if len(businessDays) > 0 {
		selectedDate := businessDays[rand.Intn(len(businessDays))]
		log.WithFields(logger.Fields{"date": selectedDate.Format("2006-01-02"), "business_days_count": len(businessDays)}).Debug("Random business date generated")
		return selectedDate
	}

	// Если все дни месяца - выходные/праздники (крайне маловероятно),
	// используем getNextBusinessDay от последнего дня месяца
	// Это может вернуть дату следующего месяца, но это крайний случай
	log.Warn("No business days in month, using next business day from last day")
	lastDayOfMonth := time.Date(year, time.Month(month), daysInMonth, 0, 0, 0, 0, time.UTC)
	return dc.GetNextBusinessDay(lastDayOfMonth)
}

// generateRandomWeekdayDate генерирует случайную дату буднего дня (исключает только выходные, НЕ исключает праздники)
// Используется для операций по карте, которые могут происходить в праздники (время 09:00-20:00)
func (dc *DateCalculator) GenerateRandomWeekdayDate(year, month int) time.Time {
	op := "service.date.generateRandomWeekdayDate"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"year": year, "month": month})
	log.Debug("Generating random weekday date")
	daysInMonth := time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.UTC).Day()

	// Собираем все будние дни месяца (исключаем только выходные, НЕ исключаем праздники)
	var weekdays []time.Time
	for day := 1; day <= daysInMonth; day++ {
		date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		if date.Weekday() != time.Saturday && date.Weekday() != time.Sunday {
			weekdays = append(weekdays, date)
		}
	}

	// Если есть будние дни, выбираем случайный
	if len(weekdays) > 0 {
		selectedDate := weekdays[rand.Intn(len(weekdays))]
		log.WithFields(logger.Fields{"date": selectedDate.Format("2006-01-02"), "weekdays_count": len(weekdays)}).Debug("Random weekday date generated")
		return selectedDate
	}

	// Если все дни месяца - выходные (крайне маловероятно),
	// возвращаем последний день месяца
	log.Warn("No weekdays in month, returning last day")
	lastDayOfMonth := time.Date(year, time.Month(month), daysInMonth, 0, 0, 0, 0, time.UTC)
	return lastDayOfMonth
}

func (dc *DateCalculator) GenerateBusinessTime(date time.Time, startHour, endHour int) time.Time {
	op := "service.date.generateBusinessTime"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"date":       date.Format("2006-01-02"),
		"start_hour": startHour,
		"end_hour":   endHour,
	})
	log.Debug("Generating business time")
	hour := startHour + rand.Intn(endHour-startHour+1)
	minute := rand.Intn(60)
	result := time.Date(date.Year(), date.Month(), date.Day(), hour, minute, 0, 0, time.UTC)
	log.WithFields(logger.Fields{"time": result.Format("15:04")}).Debug("Business time generated")
	return result
}

func (dc *DateCalculator) CalculateTransactionDate(template *entities.TransactionTemplate, year, month int, seqNum int) time.Time {
	op := "service.date.calculateTransactionDate"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"category":  template.Category,
		"year":      year,
		"month":     month,
		"seq_num":   seqNum,
		"frequency": template.Schedule.Frequency,
	})
	log.Debug("Calculating transaction date")
	baseDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)

	var result time.Time
	switch template.Schedule.Frequency {
	case "monthly":
		result = dc.CalculateMonthlyDate(baseDate, template, seqNum)
	case "biweekly":
		result = dc.CalculateBiweeklyDate(baseDate, template, seqNum)
	case "weekly":
		result = dc.CalculateWeeklyDate(baseDate, template, seqNum)
	case "once":
		result = dc.CalculateOnceDate(baseDate, template)
	default:
		result = dc.CalculateMonthlyDate(baseDate, template, seqNum)
	}
	log.WithFields(logger.Fields{"date": result.Format("2006-01-02")}).Debug("Transaction date calculated")
	return result
}

func (dc *DateCalculator) CalculatePostingDate(template *entities.TransactionTemplate, year, month int, seqNum int) time.Time {
	op := "service.date.calculatePostingDate"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"category": template.Category,
		"year":     year,
		"month":    month,
		"seq_num":  seqNum,
	})
	log.Debug("Calculating posting date")
	transactionDate := dc.CalculateTransactionDate(template, year, month, seqNum)

	// [32] Если транзакция попадает на праздник и это операция по счету, переносим
	// Примечание: "Перевод владельцу" обрабатывается отдельно в generator.go
	if dc.IsHoliday(transactionDate) && template.PaymentMethod.IsAccountTransfer() {
		log.Debug("Transaction date is holiday and account transfer, moving to next business day")
		result := dc.GetNextBusinessDay(transactionDate)
		log.WithFields(logger.Fields{"date": result.Format("2006-01-02")}).Debug("Posting date calculated")
		return result
	}

	log.WithFields(logger.Fields{"date": transactionDate.Format("2006-01-02")}).Debug("Posting date calculated")
	return transactionDate
}

func (dc *DateCalculator) CalculateMonthlyDate(baseDate time.Time, template *entities.TransactionTemplate, seqNum int) time.Time {
	op := "service.date.calculateMonthlyDate"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"category": template.Category,
		"seq_num":   seqNum,
	})
	log.Debug("Calculating monthly date")

	// [22] Для "Перевод владельцу" - всегда будний день (не праздничный)
	if value_objects.IsOwnerTransfer(template.Category) {
		result := dc.GenerateRandomBusinessDate(baseDate.Year(), int(baseDate.Month()))
		log.WithFields(logger.Fields{"date": result.Format("2006-01-02")}).Debug("Monthly date calculated (owner transfer)")
		return result
	}

	// Упрощенная реализация - нужно доработать
	if len(template.Schedule.WeekOfMonth) > 0 && seqNum <= len(template.Schedule.WeekOfMonth) {
		weekNum := template.Schedule.WeekOfMonth[seqNum-1]
		result := dc.FindNthWeekdayInMonth(baseDate, template.Schedule.PreferredDay, weekNum)
		log.WithFields(logger.Fields{"date": result.Format("2006-01-02"), "week_num": weekNum}).Debug("Monthly date calculated (from schedule)")
		return result
	}

	// По умолчанию - случайный день месяца
	daysInMonth := time.Date(baseDate.Year(), baseDate.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
	day := 1 + rand.Intn(daysInMonth)
	result := time.Date(baseDate.Year(), baseDate.Month(), day, 0, 0, 0, 0, time.UTC)
	log.WithFields(logger.Fields{"date": result.Format("2006-01-02")}).Debug("Monthly date calculated (random)")
	return result
}

func (dc *DateCalculator) CalculateBiweeklyDate(baseDate time.Time, template *entities.TransactionTemplate, seqNum int) time.Time {
	op := "service.date.calculateBiweeklyDate"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"category": template.Category, "seq_num": seqNum})
	log.Debug("Calculating biweekly date")
	result := dc.CalculateMonthlyDate(baseDate, template, seqNum)
	log.WithFields(logger.Fields{"date": result.Format("2006-01-02")}).Debug("Biweekly date calculated")
	return result
}

func (dc *DateCalculator) CalculateWeeklyDate(baseDate time.Time, template *entities.TransactionTemplate, seqNum int) time.Time {
	op := "service.date.calculateWeeklyDate"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"category": template.Category, "seq_num": seqNum})
	log.Debug("Calculating weekly date")
	result := dc.CalculateMonthlyDate(baseDate, template, seqNum)
	log.WithFields(logger.Fields{"date": result.Format("2006-01-02")}).Debug("Weekly date calculated")
	return result
}

func (dc *DateCalculator) CalculateOnceDate(baseDate time.Time, template *entities.TransactionTemplate) time.Time {
	op := "service.date.calculateOnceDate"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"category": template.Category})
	log.Debug("Calculating once date")
	result := dc.CalculateMonthlyDate(baseDate, template, 1)
	log.WithFields(logger.Fields{"date": result.Format("2006-01-02")}).Debug("Once date calculated")
	return result
}

// calculateSoftwareSubscriptionDate рассчитывает дату для подписки ПО с сохранением дня недели [25][14]
func (dc *DateCalculator) CalculateSoftwareSubscriptionDate(baseDate time.Time, userID *uuid.UUID) time.Time {
	op := "service.date.calculateSoftwareSubscriptionDate"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"base_date": baseDate.Format("2006-01-02"),
		"user_id":   userID,
	})
	log.Debug("calculateSoftwareSubscriptionDate called")

	// Пытаемся получить сохраненный день недели
	if dc.stateRepo != nil {
		weekday, err := dc.stateRepo.GetSoftwareSubscriptionWeekday(*userID)
		log.Debug("GetSoftwareSubscriptionWeekday: weekday=%d, err=%v", weekday, err)

		if err == nil && weekday >= 0 && weekday <= 6 {
			// Используем сохраненный день недели
			date := dc.FindFirstWeekdayInMonth(baseDate, time.Weekday(weekday))
			log.Debug("Using saved weekday %d, date=%v", weekday, date.Format("2006-01-02"))
			return date
		}
	} else {
		log.Debug("stateRepo is nil, cannot save weekday")
	}

	// Если не найден - выбираем случайный будний день и сохраняем
	weekdayNum := time.Weekday(1 + rand.Intn(5)) // Понедельник-Пятница (1-5)
	date := dc.FindFirstWeekdayInMonth(baseDate, weekdayNum)
	log.Debug("Selected new weekday %d, date=%v", int(weekdayNum), date.Format("2006-01-02"))

	// Сохраняем выбранный день недели
	if dc.stateRepo != nil {
		if err := dc.stateRepo.SaveSoftwareSubscriptionWeekday(*userID, int(weekdayNum)); err != nil {
			log.Error(err, "Failed to save software subscription weekday")
		} else {
			log.Debug("Successfully saved weekday %d for userID=%v", int(weekdayNum), userID)
		}
	} else {
		log.Warn("stateRepo is nil, cannot save weekday")
	}

	return date
}

// findFirstWeekdayInMonth находит первый день указанного дня недели в месяце
func (dc *DateCalculator) FindFirstWeekdayInMonth(baseDate time.Time, weekday time.Weekday) time.Time {
	op := "service.date.findFirstWeekdayInMonth"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"date":    baseDate.Format("2006-01-02"),
		"weekday": weekday.String(),
	})
	log.Debug("Finding first weekday in month")
	current := time.Date(baseDate.Year(), baseDate.Month(), 1, 0, 0, 0, 0, time.UTC)

	// Находим первый день недели в месяце
	for current.Weekday() != weekday {
		current = current.AddDate(0, 0, 1)
		if current.Month() != baseDate.Month() {
			// Если вышли за пределы месяца, возвращаем первый день месяца
			result := time.Date(baseDate.Year(), baseDate.Month(), 1, 0, 0, 0, 0, time.UTC)
			log.Warn("Weekday not found in month, returning first day")
			return result
		}
	}

	log.WithFields(logger.Fields{"date": current.Format("2006-01-02")}).Debug("First weekday found")
	return current
}

func (dc *DateCalculator) FindNthWeekdayInMonth(baseDate time.Time, weekday string, n int) time.Time {
	op := "service.date.findNthWeekdayInMonth"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"date":    baseDate.Format("2006-01-02"),
		"weekday": weekday,
		"n":       n,
	})
	log.Debug("Finding nth weekday in month")
	weekdayNum := parseWeekday(weekday)
	current := time.Date(baseDate.Year(), baseDate.Month(), 1, 0, 0, 0, 0, time.UTC)

	for current.Weekday() != weekdayNum {
		current = current.AddDate(0, 0, 1)
	}

	if n > 1 {
		current = current.AddDate(0, 0, (n-1)*7)
	}

	if current.Month() != baseDate.Month() {
		for current.AddDate(0, 0, 7).Month() == baseDate.Month() {
			current = current.AddDate(0, 0, 7)
		}
	}

	log.WithFields(logger.Fields{"date": current.Format("2006-01-02")}).Debug("Nth weekday found")
	return current
}

func (dc *DateCalculator) GetNextBusinessDay(date time.Time) time.Time {
	op := "service.date.getNextBusinessDay"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"date": date.Format("2006-01-02")})
	log.Debug("Getting next business day")

	// Используем HolidayService для получения следующего рабочего дня
	if dc.holidayService != nil {
		result := dc.holidayService.GetNextBusinessDay(date)
		log.WithFields(logger.Fields{"next_date": result.Format("2006-01-02")}).Debug("Next business day retrieved via service")
		return result
	}
	// Fallback на локальную логику
	nextDay := date.AddDate(0, 0, 1)
	for nextDay.Weekday() == time.Saturday || nextDay.Weekday() == time.Sunday || dc.IsHoliday(nextDay) {
		nextDay = nextDay.AddDate(0, 0, 1)
	}
	log.WithFields(logger.Fields{"next_date": nextDay.Format("2006-01-02")}).Debug("Next business day calculated via fallback")
	return nextDay
}

func parseWeekday(weekday string) time.Weekday {
	switch weekday {
	case "Monday", "monday":
		return time.Monday
	case "Tuesday", "tuesday":
		return time.Tuesday
	case "Wednesday", "wednesday":
		return time.Wednesday
	case "Thursday", "thursday":
		return time.Thursday
	case "Friday", "friday":
		return time.Friday
	case "Saturday", "saturday":
		return time.Saturday
	case "Sunday", "sunday":
		return time.Sunday
	default:
		return time.Monday
	}
}

// isFirstMonthForCategory проверяет, является ли месяц первым для категории
// Использует проверку истории генераций для более надежного определения
func (dc *DateCalculator) IsFirstMonthForCategory(userID *string, categoryKey string, monthStr string) bool {
	if userID == nil || *userID == "" {
		// Если userID нет, используем fallback логику
		return true
	}

	op := "service.date.isFirstMonthForCategory"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"category_key": categoryKey,
		"month":        monthStr,
	})

	userUUID := helpers.ParseUUIDOrNil(userID)
	if userUUID == nil {
		log.Warn("Invalid userID in isFirstMonthForCategory")
		return true
	}

	// 1. Проверяем сохраненный first_month из state
	savedFirstMonth := ""
	switch categoryKey {
	case "leasing":
		savedFirstMonth, _ = dc.baseAmountService.GetLeasingFirstMonth(*userID)
	case "mobile":
		savedFirstMonth, _ = dc.baseAmountService.GetMobileFirstMonth(*userID)
	case "utilities":
		savedFirstMonth, _ = dc.baseAmountService.GetUtilitiesFirstMonth(*userID)
	}

	// Если есть сохраненный first_month и запрашиваемый месяц <= сохраненному, это первый месяц
	if savedFirstMonth != "" {
		return monthStr <= savedFirstMonth
	}

	// 2. Если сохраненного first_month нет, проверяем историю генераций
	// Если у пользователя есть завершенные генерации, это не первый месяц
	completedRequests, err := dc.generationRequestRepo.GetCompletedByUserID(*userUUID)
	if err != nil {
		// Если ошибка при проверке истории, логируем и считаем первым месяцем (fallback)
		log.Warn("Failed to check generation history for userID=%s: %v, treating as first month", *userID, err)
		return true
	}

	// Если есть хотя бы одна завершенная генерация, это не первый месяц
	if len(completedRequests) > 0 {
		return false
	}

	// Если нет завершенных генераций и нет сохраненного first_month, это первый месяц
	return true
}
