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

	// Парсим дату
	// t, err := time.Parse("2006-01-02", dateStr)
	// if err != nil {
	// 	return "", err
	// }

	// // Создаём новое время с 10:00:00 в UTC
	// newTime := time.Date(
	// 	t.Year(), t.Month(), t.Day(),
	// 	10, 0, 0, 0,
	// 	time.UTC,
	// )

	// // Возвращаем в формате RFC3339
	// return newTime.Format(time.RFC3339), nil
}
