package domain

import (
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/pkg/helpers"
)

type Holiday struct {
	HolidayDate string
	Name string
	Country string
}

func NewHoliday(date time.Time, name string, country string) (*Holiday, error) {
	if !IsValidCountry(country) {
		return nil, helpers.ErrInvalidCountry
	}
	if !IsValidDate(date) {
		return nil, helpers.ErrInvalidDate
	}
	dateStr := date.Format("2006-01-02")
	return &Holiday{
		HolidayDate: dateStr,
		Name: name,
		Country: country,
	}, nil
}

func IsValidCountry(country string) bool {
	return country == "RU" || country == "US" || country == "BY"
}

func IsValidDate(date time.Time) bool {
	return !date.IsZero()
}