// internal/service/date_calculator.go
package service

import (
	"math/rand"
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/internal/domain"
	"github.com/IbadT/business_bank_back/services/matematika/internal/domain/entities"
	"github.com/IbadT/business_bank_back/services/matematika/internal/repository"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type dateCalculator struct {
	holidays       []*domain.Holiday
	holidayMap     map[string]bool
	stateRepo      repository.StateRepository
	holidayService HolidayService
}

func newDateCalculator(holidays []*domain.Holiday, stateRepo repository.StateRepository, holidayService HolidayService) *dateCalculator {
	holidayMap := make(map[string]bool)
	for _, holiday := range holidays {
		dateStr := holiday.HolidayDate
		holidayMap[dateStr] = true
	}

	return &dateCalculator{
		holidays:       holidays,
		holidayMap:     holidayMap,
		stateRepo:      stateRepo,
		holidayService: holidayService,
	}
}

// isQuarterlyMonth проверяет, является ли месяц квартальным [23][24]
// Квартальные месяцы: январь (1), апрель (4), июнь (6), сентябрь (9)
func (dc *dateCalculator) isQuarterlyMonth(month int) bool {
	quarterlyMonths := []int{1, 4, 6, 9} // Январь, Апрель, Июнь, Сентябрь
	for _, m := range quarterlyMonths {
		if month == m {
			return true
		}
	}
	return false
}

// calculateIRSDate рассчитывает дату для IRS налогов [23][24]
// Всегда возвращает 15-е число месяца (или следующий рабочий день, если 15-е - выходной/праздник)
// seqNum не используется, так как обе транзакции в квартальный месяц должны быть 15-го числа
func (dc *dateCalculator) calculateIRSDate(year, month int, seqNum int) time.Time {
	// [23][24] Всегда 15-е число
	date := time.Date(year, time.Month(month), 15, 0, 0, 0, 0, time.UTC)

	// Если 15-е число попадает на выходной или на праздник - переносим на следующий рабочий день
	if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday || dc.isHoliday(date) {
		date = dc.getNextBusinessDay(date)
	}

	return date
}

func (dc *dateCalculator) isHoliday(date time.Time) bool {
	// Используем HolidayService для проверки праздников из БД
	if dc.holidayService != nil {
		return dc.holidayService.IsHoliday(date)
	}
	// Fallback на локальную карту из конфига
	dateStr := date.Format("2006-01-02")
	return dc.holidayMap[dateStr]
}

func (dc *dateCalculator) getFridaysInMonth(year, month int) []time.Time {
	return dc.getWeekdaysInMonth(year, month, time.Friday)
}

func (dc *dateCalculator) getWeekdaysInMonth(year, month int, weekday time.Weekday) []time.Time {
	var days []time.Time
	current := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)

	for current.Month() == time.Month(month) {
		if current.Weekday() == weekday {
			days = append(days, current)
		}
		current = current.AddDate(0, 0, 1)
	}

	return days
}

func (dc *dateCalculator) generateRandomBusinessDate(year, month int) time.Time {
	daysInMonth := time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.UTC).Day()

	// Собираем все рабочие дни месяца
	var businessDays []time.Time
	for day := 1; day <= daysInMonth; day++ {
		date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		if date.Weekday() != time.Saturday && date.Weekday() != time.Sunday && !dc.isHoliday(date) {
			businessDays = append(businessDays, date)
		}
	}

	// Если есть рабочие дни, выбираем случайный
	if len(businessDays) > 0 {
		return businessDays[rand.Intn(len(businessDays))]
	}

	// Если все дни месяца - выходные/праздники (крайне маловероятно),
	// используем getNextBusinessDay от последнего дня месяца
	// Это может вернуть дату следующего месяца, но это крайний случай
	lastDayOfMonth := time.Date(year, time.Month(month), daysInMonth, 0, 0, 0, 0, time.UTC)
	return dc.getNextBusinessDay(lastDayOfMonth)
}

func (dc *dateCalculator) generateBusinessTime(date time.Time, startHour, endHour int) time.Time {
	hour := startHour + rand.Intn(endHour-startHour+1)
	minute := rand.Intn(60)

	return time.Date(date.Year(), date.Month(), date.Day(), hour, minute, 0, 0, time.UTC)
}

func (dc *dateCalculator) calculateTransactionDate(template *entities.TransactionTemplate, year, month int, seqNum int) time.Time {
	baseDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)

	switch template.Schedule.Frequency {
	case "monthly":
		return dc.calculateMonthlyDate(baseDate, template, seqNum)
	case "biweekly":
		return dc.calculateBiweeklyDate(baseDate, template, seqNum)
	case "weekly":
		return dc.calculateWeeklyDate(baseDate, template, seqNum)
	case "once":
		return dc.calculateOnceDate(baseDate, template)
	default:
		return dc.calculateMonthlyDate(baseDate, template, seqNum)
	}
}

func (dc *dateCalculator) calculatePostingDate(template *entities.TransactionTemplate, year, month int, seqNum int) time.Time {
	transactionDate := dc.calculateTransactionDate(template, year, month, seqNum)

	// [32] Если транзакция попадает на праздник и это операция по счету, переносим
	// Примечание: "Перевод владельцу" обрабатывается отдельно в generator.go
	if dc.isHoliday(transactionDate) && template.PaymentMethod.IsAccountTransfer() {
		return dc.getNextBusinessDay(transactionDate)
	}

	return transactionDate
}

func (dc *dateCalculator) calculateMonthlyDate(baseDate time.Time, template *entities.TransactionTemplate, seqNum int) time.Time {
	// [22] Для "Перевод владельцу" - всегда будний день (не праздничный)
	if template.Category == "Перевод владельцу" || template.Category == "Owner Transfer" {
		return dc.generateRandomBusinessDate(baseDate.Year(), int(baseDate.Month()))
	}

	// Упрощенная реализация - нужно доработать
	if len(template.Schedule.WeekOfMonth) > 0 && seqNum <= len(template.Schedule.WeekOfMonth) {
		weekNum := template.Schedule.WeekOfMonth[seqNum-1]
		return dc.findNthWeekdayInMonth(baseDate, template.Schedule.PreferredDay, weekNum)
	}

	// По умолчанию - случайный день месяца
	daysInMonth := time.Date(baseDate.Year(), baseDate.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
	day := 1 + rand.Intn(daysInMonth)
	return time.Date(baseDate.Year(), baseDate.Month(), day, 0, 0, 0, 0, time.UTC)
}

func (dc *dateCalculator) calculateBiweeklyDate(baseDate time.Time, template *entities.TransactionTemplate, seqNum int) time.Time {
	// Упрощенная реализация
	return dc.calculateMonthlyDate(baseDate, template, seqNum)
}

func (dc *dateCalculator) calculateWeeklyDate(baseDate time.Time, template *entities.TransactionTemplate, seqNum int) time.Time {
	// Упрощенная реализация
	return dc.calculateMonthlyDate(baseDate, template, seqNum)
}

func (dc *dateCalculator) calculateOnceDate(baseDate time.Time, template *entities.TransactionTemplate) time.Time {
	// Упрощенная реализация
	return dc.calculateMonthlyDate(baseDate, template, 1)
}

// calculateSoftwareSubscriptionDate рассчитывает дату для подписки ПО с сохранением дня недели [25][14]
func (dc *dateCalculator) calculateSoftwareSubscriptionDate(baseDate time.Time, userID *uuid.UUID) time.Time {
	logrus.Debugf("[DEBUG] calculateSoftwareSubscriptionDate called: baseDate=%v, userID=%v, stateRepo=%v",
		baseDate.Format("2006-01-02"), userID, dc.stateRepo != nil)

	// Пытаемся получить сохраненный день недели
	if dc.stateRepo != nil {
		weekday, err := dc.stateRepo.GetSoftwareSubscriptionWeekday(*userID)
		logrus.Debugf("[DEBUG] GetSoftwareSubscriptionWeekday: weekday=%d, err=%v", weekday, err)

		if err == nil && weekday >= 0 && weekday <= 6 {
			// Используем сохраненный день недели
			date := dc.findFirstWeekdayInMonth(baseDate, time.Weekday(weekday))
			logrus.Debugf("[DEBUG] Using saved weekday %d, date=%v", weekday, date.Format("2006-01-02"))
			return date
		}
	} else {
		logrus.Debugf("[DEBUG] stateRepo is nil, cannot save weekday")
	}

	// Если не найден - выбираем случайный будний день и сохраняем
	weekdayNum := time.Weekday(1 + rand.Intn(5)) // Понедельник-Пятница (1-5)
	date := dc.findFirstWeekdayInMonth(baseDate, weekdayNum)
	logrus.Debugf("[DEBUG] Selected new weekday %d, date=%v", int(weekdayNum), date.Format("2006-01-02"))

	// Сохраняем выбранный день недели
	if dc.stateRepo != nil {
		if err := dc.stateRepo.SaveSoftwareSubscriptionWeekday(*userID, int(weekdayNum)); err != nil {
			logrus.Debugf("[ERROR] Failed to save software subscription weekday: %v (userID=%v)", err, userID)
		} else {
			logrus.Debugf("[DEBUG] Successfully saved weekday %d for userID=%v", int(weekdayNum), userID)
		}
	} else {
		logrus.Debugf("[WARN] stateRepo is nil, cannot save weekday")
	}

	return date
}

// findFirstWeekdayInMonth находит первый день указанного дня недели в месяце
func (dc *dateCalculator) findFirstWeekdayInMonth(baseDate time.Time, weekday time.Weekday) time.Time {
	current := time.Date(baseDate.Year(), baseDate.Month(), 1, 0, 0, 0, 0, time.UTC)

	// Находим первый день недели в месяце
	for current.Weekday() != weekday {
		current = current.AddDate(0, 0, 1)
		if current.Month() != baseDate.Month() {
			// Если вышли за пределы месяца, возвращаем первый день месяца
			return time.Date(baseDate.Year(), baseDate.Month(), 1, 0, 0, 0, 0, time.UTC)
		}
	}

	return current
}

func (dc *dateCalculator) findNthWeekdayInMonth(baseDate time.Time, weekday string, n int) time.Time {
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

	return current
}

func (dc *dateCalculator) getNextBusinessDay(date time.Time) time.Time {
	// Используем HolidayService для получения следующего рабочего дня
	if dc.holidayService != nil {
		return dc.holidayService.GetNextBusinessDay(date)
	}
	// Fallback на локальную логику
	nextDay := date.AddDate(0, 0, 1)
	for nextDay.Weekday() == time.Saturday || nextDay.Weekday() == time.Sunday || dc.isHoliday(nextDay) {
		nextDay = nextDay.AddDate(0, 0, 1)
	}
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
