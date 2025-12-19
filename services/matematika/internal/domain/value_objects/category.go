// internal/domain/value_objects/category.go
package value_objects

// Константы категорий транзакций
const (
	// Software Subscription
	CategorySoftwareSubscriptionRU = "Подписка ПО"
	CategorySoftwareSubscriptionEN = "Software Subscription"
	CategorySaaSSubscription       = "SaaS Subscription"

	// IRS Taxes
	CategoryIRSTaxesRU = "IRS налоги"
	CategoryIRS        = "IRS"

	// Toll Roads
	CategoryTollRoadsRU = "Платные дороги"
	CategoryTollRoadsEN = "Toll Roads"
	CategoryToll        = "Toll"

	// Overload
	CategoryOverloadRU = "Перегруз"
	CategoryOverloadEN = "Overload"

	// Owner Transfer
	CategoryOwnerTransferRU = "Перевод владельцу"
	CategoryOwnerTransferEN = "Owner Transfer"

	// Mobile
	CategoryMobileRU = "Мобильная связь"
	CategoryMobileEN = "Mobile"

	// Utilities
	CategoryUtilitiesRU = "Коммунальные"
	CategoryUtilitiesEN = "Utilities"

	// Leasing
	CategoryLeasingRU      = "Лизинг"
	CategoryLeasingEN      = "Leasing"
	CategoryEquipmentLease = "Equipment lease"

	// Rent
	CategoryRentRU         = "Аренда"
	CategoryRentEN         = "Rent"
	CategoryWarehouseRent = "Warehouse rent"

	// Insurance
	CategoryInsuranceRU = "Страхование"
	CategoryInsuranceEN = "Insurance"

	// Payroll
	CategoryPayroll = "Payroll"
)

// IsSoftwareSubscription проверяет, является ли категория подпиской ПО
func IsSoftwareSubscription(category string) bool {
	return category == CategorySoftwareSubscriptionRU ||
		category == CategorySoftwareSubscriptionEN ||
		category == CategorySaaSSubscription
}

// IsIRSTaxes проверяет, является ли категория налогами IRS
func IsIRSTaxes(category string) bool {
	return category == CategoryIRSTaxesRU || category == CategoryIRS
}

// IsTollRoads проверяет, является ли категория платными дорогами
func IsTollRoads(category string) bool {
	return category == CategoryTollRoadsRU ||
		category == CategoryTollRoadsEN ||
		category == CategoryToll
}

// IsOverload проверяет, является ли категория перегрузом
func IsOverload(category string) bool {
	return category == CategoryOverloadRU || category == CategoryOverloadEN
}

// IsOwnerTransfer проверяет, является ли категория переводом владельцу
func IsOwnerTransfer(category string) bool {
	return category == CategoryOwnerTransferRU || category == CategoryOwnerTransferEN
}

// IsMobile проверяет, является ли категория мобильной связью
func IsMobile(category string) bool {
	return category == CategoryMobileRU || category == CategoryMobileEN
}

// IsUtilities проверяет, является ли категория коммунальными
func IsUtilities(category string) bool {
	return category == CategoryUtilitiesRU || category == CategoryUtilitiesEN
}

// IsLeasing проверяет, является ли категория лизингом
func IsLeasing(category string) bool {
	return category == CategoryLeasingRU ||
		category == CategoryLeasingEN ||
		category == CategoryEquipmentLease
}

// IsRent проверяет, является ли категория арендой
func IsRent(category string) bool {
	return category == CategoryRentRU ||
		category == CategoryRentEN ||
		category == CategoryWarehouseRent
}

// IsInsurance проверяет, является ли категория страхованием
func IsInsurance(category string) bool {
	return category == CategoryInsuranceRU || category == CategoryInsuranceEN
}

// IsPayroll проверяет, является ли категория зарплатой
func IsPayroll(category string) bool {
	return category == CategoryPayroll
}

// GetFixedMonthlyCategories возвращает карту категорий, которые не масштабируются (фиксированные 1 раз в месяц)
func GetFixedMonthlyCategories() map[string]bool {
	return map[string]bool{
		CategoryPayroll:              true,
		CategorySoftwareSubscriptionRU: true,
		CategorySoftwareSubscriptionEN: true,
		CategorySaaSSubscription:      true,
		CategoryMobileRU:             true,
		CategoryMobileEN:             true,
		CategoryUtilitiesRU:          true,
		CategoryUtilitiesEN:          true,
		CategoryLeasingRU:            true,
		CategoryLeasingEN:            true,
		CategoryEquipmentLease:       true,
		CategoryRentRU:               true,
		CategoryRentEN:               true,
		CategoryWarehouseRent:        true,
		CategoryInsuranceRU:          true,
		CategoryInsuranceEN:          true,
		CategoryOwnerTransferRU:      true,
		CategoryOwnerTransferEN:      true,
	}
}
