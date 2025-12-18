package helpers

import (
	"fmt"
	"time"
)

func ParseDates(dateStart string, dateEnd string) (string, string, error) {
	newTimeStart, err := time.Parse("2006-01-02", dateStart)
	if err != nil {
		return "", "", fmt.Errorf("%w: %s", ErrFailedToParseDate, err.Error())
	}

	newTimeEnd, err := time.Parse("2006-01-02", dateEnd)
	if err != nil {
		return "", "", fmt.Errorf("%w: %s", ErrFailedToParseDate, err.Error())
	}

	newTimeStart = time.Date(newTimeStart.Year(), newTimeStart.Month(), newTimeStart.Day(), 0, 0, 0, 0, time.UTC)
	newTimeEnd = time.Date(newTimeEnd.Year(), newTimeEnd.Month(), newTimeEnd.Day(), 23, 59, 59, 999999999, time.UTC)

	return newTimeStart.Format(time.RFC3339), newTimeEnd.Format(time.RFC3339), nil
}
