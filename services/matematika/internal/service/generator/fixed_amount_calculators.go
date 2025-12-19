// internal/service/generator/fixed_amount_calculators.go
package generatorservice

import (
	"fmt"
	"math/rand"

	"github.com/IbadT/business_bank_back/services/matematika/internal/domain/value_objects"
	baseamountservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/base"
	dateservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/date"
	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/dto"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/utils"
	"github.com/sirupsen/logrus"
)

// fixedAmountCalculator рассчитывает фиксированные суммы для различных категорий
type fixedAmountCalculator struct {
	baseAmountService baseamountservice.BaseAmountService
	dateCalculator    *dateservice.DateCalculator
}

// newFixedAmountCalculator создает новый калькулятор фиксированных сумм
func newFixedAmountCalculator(
	baseAmountService baseamountservice.BaseAmountService,
	dateCalculator *dateservice.DateCalculator,
) *fixedAmountCalculator {
	return &fixedAmountCalculator{
		baseAmountService: baseAmountService,
		dateCalculator:    dateCalculator,
	}
}

// CalculateFixedAmount рассчитывает фиксированную сумму для категории
func (c *fixedAmountCalculator) CalculateFixedAmount(
	category string,
	baseAmount float64,
	req *dto.GenerateRequest,
	userID *string,
) (float64, map[string]interface{}) {
	switch category {
	case value_objects.CategoryOverloadRU, value_objects.CategoryOverloadEN:
		return c.calculateOverload()

	case value_objects.CategoryLeasingRU, value_objects.CategoryLeasingEN, value_objects.CategoryEquipmentLease:
		return c.calculateLeasing(req, userID)

	case value_objects.CategoryMobileRU, value_objects.CategoryMobileEN:
		return c.calculateMobile(req, baseAmount, userID)

	case value_objects.CategoryUtilitiesRU, value_objects.CategoryUtilitiesEN:
		return c.calculateUtilities(req, baseAmount, userID)

	case value_objects.CategoryTollRoadsRU, value_objects.CategoryTollRoadsEN, value_objects.CategoryToll:
		return c.calculateTollRoads()

	case value_objects.CategorySoftwareSubscriptionRU, value_objects.CategorySoftwareSubscriptionEN, value_objects.CategorySaaSSubscription:
		return c.calculateSoftwareSubscription(baseAmount)

	default:
		return baseAmount, nil
	}
}

func (c *fixedAmountCalculator) calculateOverload() (float64, map[string]interface{}) {
	// [20][21] вес (200–1000 lb) * ставку ($0.011–$0.039)
	weight := 200 + rand.Intn(801) // 200-1000
	rate := 0.011 + rand.Float64()*(0.039-0.011)
	amount := float64(weight) * rate
	amount = utils.RoundToCents(amount) // Округляем до центов
	details := map[string]interface{}{
		"weight_lb":   weight,
		"rate_per_lb": fmt.Sprintf("%.3f", rate),
		"formula":     "weight * rate",
	}
	return amount, details
}

func (c *fixedAmountCalculator) calculateLeasing(req *dto.GenerateRequest, userID *string) (float64, map[string]interface{}) {
	// [19] Используем BaseAmountService для расчета лизинга
	if userID == nil || *userID == "" {
		// Fallback на старую логику если userID не указан
		// Используем dateCalculator для проверки первого месяца
		monthStr := utils.FormatMonth(req.Year, req.Month)
		emptyUserID := ""
		isFirstMonth := c.dateCalculator.IsFirstMonthForCategory(&emptyUserID, "leasing", monthStr)
		
		if isFirstMonth {
			percentage := 0.115 + rand.Float64()*(0.12-0.115)
			amount := req.Turnover * percentage
			details := map[string]interface{}{
				"type":                    "first_month_lease",
				"percentage_of_turnover":  percentage,
				"fixed_for_future_months": true,
			}
			// Сохраняем через baseAmountService если возможно
			if err := c.baseAmountService.SaveLeasingBaseAmount("", amount, monthStr, req.Turnover); err != nil {
				logrus.Infof("[WARN] Failed to save leasing amount: %v", err)
			}
			return amount, details
		} else {
			// Получаем сохраненную сумму
			amount, err := c.baseAmountService.GetLeasingBaseAmount("")
			if err != nil || amount == 0 {
				// Если не найдено, используем дефолтное значение
				amount = req.Turnover * 0.115
				logrus.Infof("[WARN] Leasing base amount not found, using default: %.2f", amount)
			}
			details := map[string]interface{}{
				"type": "recurring_lease",
			}
			return amount, details
		}
	}

	// Проверяем, является ли это первым месяцем
	// Используем проверку истории генераций для более надежного определения
	monthStr := utils.FormatMonth(req.Year, req.Month)
	isFirstMonth := c.dateCalculator.IsFirstMonthForCategory(userID, "leasing", monthStr)

	amount, err := c.baseAmountService.CalculateLeasingAmount(*userID, req.Turnover, isFirstMonth, monthStr)
	if err != nil {
		// Fallback на старую логику при ошибке
		logrus.Infof("[WARN] Failed to calculate leasing amount via BaseAmountService: %v, using fallback", err)
		if isFirstMonth {
			// TODO: мне кажется, рандомное значение не должно быть !!!!!!!!!!!
			percentage := 0.115 + rand.Float64()*(0.12-0.115)
			amount = req.Turnover * percentage
		} else {
			// Получаем сохраненную сумму
			savedAmount, getErr := c.baseAmountService.GetLeasingBaseAmount(*userID)
			if getErr != nil || savedAmount == 0 {
				// Если не найдено, используем дефолтное значение
				amount = req.Turnover * 0.115
				logrus.Infof("[WARN] Leasing base amount not found in fallback, using default: %.2f", amount)
			} else {
				amount = savedAmount
			}
		}
	}

	details := map[string]interface{}{
		"type": "lease",
	}
	if isFirstMonth {
		details["is_first_month"] = true
	} else {
		details["is_first_month"] = false
	}
	return amount, details
}

func (c *fixedAmountCalculator) calculateMobile(req *dto.GenerateRequest, baseAmount float64, userID *string) (float64, map[string]interface{}) {
	// [15][16] Мобильная связь: фиксируется в первом месяце в диапазоне $200–500 и далее меняется ±15% от этой базы
	if userID == nil || *userID == "" {
		// Fallback логика для случаев без userID
		// Используем dateCalculator для проверки первого месяца
		monthStr := utils.FormatMonth(req.Year, req.Month)
		emptyUserID := ""
		firstMonth := c.dateCalculator.IsFirstMonthForCategory(&emptyUserID, "mobile", monthStr)
		var amount float64
		details := map[string]interface{}{
			"type":         "mobile",
			"fallback_mode": true,
		}

		if firstMonth {
			// Первый месяц: $200–500
			amount = 200.0 + rand.Float64()*(500.0-200.0)
			amount = utils.RoundToCents(amount)
			details["is_first_month"] = true
			details["amount_range"] = "$200–500"
		} else {
			// Последующие месяцы: используем baseAmount если есть, иначе fallback
			if baseAmount > 0 {
				// Применяем ±15% к baseAmount
				deviation := (rand.Float64()*2 - 1) * 0.15 // от -0.15 до +0.15
				amount = baseAmount * (1.0 + deviation)
				amount = utils.RoundToCents(amount)
				details["base_amount"] = baseAmount
				details["deviation_percent"] = utils.FormatPercentagePercent(deviation)
			} else {
				// Если baseAmount не задан, используем среднее значение
				amount = 350.0
				amount = utils.RoundToCents(amount)
				details["fallback_amount"] = amount
			}
			details["is_first_month"] = false
		}
		return amount, details
	}

	// Проверяем, является ли это первым месяцем
	// Используем проверку истории генераций для более надежного определения
	monthStr := utils.FormatMonth(req.Year, req.Month)
	isFirstMonth := c.dateCalculator.IsFirstMonthForCategory(userID, "mobile", monthStr)

	amount, err := c.baseAmountService.CalculateMobileAmount(*userID, isFirstMonth, monthStr)
	if err != nil {
		logrus.Infof("[WARN] Failed to calculate mobile amount via BaseAmountService: %v, using fallback", err)
		// Fallback на базовую логику
		if isFirstMonth {
			amount = 200.0 + rand.Float64()*(500.0-200.0)
			amount = utils.RoundToCents(amount)
		} else if baseAmount > 0 {
			deviation := (rand.Float64()*2 - 1) * 0.15
			amount = baseAmount * (1.0 + deviation)
			amount = utils.RoundToCents(amount)
		} else {
			amount = baseAmount
		}
	}

	details := map[string]interface{}{
		"type": "mobile",
	}
	if isFirstMonth {
		details["is_first_month"] = true
		details["amount_range"] = "$200–500"
	} else {
		details["is_first_month"] = false
		details["variation"] = "±15%"
	}
	return amount, details
}

func (c *fixedAmountCalculator) calculateUtilities(req *dto.GenerateRequest, baseAmount float64, userID *string) (float64, map[string]interface{}) {
	// [15][16] Коммунальные: фиксируются в первом месяце в диапазоне $200–500 и далее меняются ±15% от этой базы
	if userID == nil || *userID == "" {
		// Fallback логика для случаев без userID
		// Используем dateCalculator для проверки первого месяца
		monthStr := utils.FormatMonth(req.Year, req.Month)
		emptyUserID := ""
		firstMonth := c.dateCalculator.IsFirstMonthForCategory(&emptyUserID, "utilities", monthStr)
		var amount float64
		details := map[string]interface{}{
			"type":         "utilities",
			"fallback_mode": true,
		}

		if firstMonth {
			// Первый месяц: $200–500
			amount = 200.0 + rand.Float64()*(500.0-200.0)
			amount = utils.RoundToCents(amount)
			details["is_first_month"] = true
			details["amount_range"] = "$200–500"
		} else {
			// Последующие месяцы: используем baseAmount если есть, иначе fallback
			if baseAmount > 0 {
				// Применяем ±15% к baseAmount
				deviation := (rand.Float64()*2 - 1) * 0.15 // от -0.15 до +0.15
				amount = baseAmount * (1.0 + deviation)
				amount = utils.RoundToCents(amount)
				details["base_amount"] = baseAmount
				details["deviation_percent"] = utils.FormatPercentagePercent(deviation)
			} else {
				// Если baseAmount не задан, используем среднее значение
				amount = 350.0
				amount = utils.RoundToCents(amount)
				details["fallback_amount"] = amount
			}
			details["is_first_month"] = false
		}
		return amount, details
	}

	// Проверяем, является ли это первым месяцем
	// Используем проверку истории генераций для более надежного определения
	monthStr := utils.FormatMonth(req.Year, req.Month)
	isFirstMonth := c.dateCalculator.IsFirstMonthForCategory(userID, "utilities", monthStr)

	amount, err := c.baseAmountService.CalculateUtilitiesAmount(*userID, isFirstMonth, monthStr)
	if err != nil {
		logrus.Infof("[WARN] Failed to calculate utilities amount via BaseAmountService: %v, using fallback", err)
		// Fallback на базовую логику
		if isFirstMonth {
			amount = 200.0 + rand.Float64()*(500.0-200.0)
			amount = utils.RoundToCents(amount)
		} else if baseAmount > 0 {
			deviation := (rand.Float64()*2 - 1) * 0.15
			amount = baseAmount * (1.0 + deviation)
			amount = utils.RoundToCents(amount)
		} else {
			amount = baseAmount
		}
	}

	details := map[string]interface{}{
		"type": "utilities",
	}
	if isFirstMonth {
		details["is_first_month"] = true
		details["amount_range"] = "$200–500"
	} else {
		details["is_first_month"] = false
		details["variation"] = "±15%"
	}
	return amount, details
}

func (c *fixedAmountCalculator) calculateTollRoads() (float64, map[string]interface{}) {
	// [17][18] Фиксированные значения $20/$35/$50 за транзакцию
	tollOptions := []float64{20.0, 35.0, 50.0}
	selectedIndex := rand.Intn(len(tollOptions))
	amount := tollOptions[selectedIndex]
	details := map[string]interface{}{
		"type":         "toll_road",
		"fixed_amount": amount,
		"available_values": tollOptions,
		"selected_value": amount,
		"description":  "Random selection from fixed values: $20, $35, or $50 per transaction",
	}
	return amount, details
}

func (c *fixedAmountCalculator) calculateSoftwareSubscription(baseAmount float64) (float64, map[string]interface{}) {
	// [13][14] Фиксированная цена из конфигурации (одна транзакция в месяц)
	if baseAmount > 0 {
		// Используем фиксированную цену из конфигурации (template.FixedAmount)
		details := map[string]interface{}{
			"type":         "software_subscription",
			"fixed_amount": baseAmount,
			"source":       "configuration",
		}
		return baseAmount, details
	}
	// Fallback если не задано в конфигурации (не должно происходить в нормальной работе)
	fallbackAmount := 299.0
	logrus.Infof("[WARN] Software subscription fixed amount not set in configuration, using fallback: %.2f", fallbackAmount)
	details := map[string]interface{}{
		"type":     "software_subscription",
		"fallback": true,
		"amount":   fallbackAmount,
	}
	return fallbackAmount, details
}
