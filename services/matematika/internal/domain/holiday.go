package domain

import (
	"errors"
	"time"
)

type Holiday struct {
	HolidayDate string
	Name string
	Country string
}

func NewHoliday(date time.Time, name string, country string) (*Holiday, error) {
	if !IsValidCountry(country) {
		return nil, errors.New("invalid country")
	}
	if !IsValidDate(date) {
		return nil, errors.New("invalid date")
	}
	dateStr := date.Format("2006-01-02")
	return &Holiday{
		HolidayDate: dateStr,
		Name: name,
		Country: country,
	}, nil
}

func IsValidCountry(country string) bool {
	return country == "RU" || country == "US"
}

func IsValidDate(date time.Time) bool {
	return !date.IsZero()
}