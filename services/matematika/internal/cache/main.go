package cache

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/IbadT/business_bank_back/services/matematika/internal/database"
	"github.com/IbadT/business_bank_back/services/matematika/internal/domain"
	"github.com/sirupsen/logrus"
)

var (
	HOLIDAYS_KEY string = "holidays"
)

type CacheService struct {
	repo             *Repository
	cacheHolidaysTTL int
}

func New(repo *Repository) *CacheService {
	holidaysTTL := database.GetEnvInt("CACHE_HOLIDAYS_TTL", 60*60*24)

	return &CacheService{
		repo:             repo,
		cacheHolidaysTTL: holidaysTTL,
	}
}

func (cs *CacheService) GetHolidays(ctx context.Context) ([]domain.Holiday, error) {
	if cs.cacheHolidaysTTL == 0 {
		return []domain.Holiday{}, errors.New("cache holidays ttl is 0")
	}

	holidaysJSON, err := cs.repo.rds.GetStrSlice(ctx, HOLIDAYS_KEY)
	if err == nil && len(holidaysJSON) > 0 {
		holidays := make([]domain.Holiday, len(holidaysJSON))
		for i, holidayJSON := range holidaysJSON {
			err = json.Unmarshal([]byte(holidayJSON), &holidays[i])
			if err != nil {
				logrus.Error("[module:cache] GetHolidays: unmarshal error: ", err)
				return []domain.Holiday{}, err
			}
		}
		logrus.Error("[module:cache] GetHolidays: from redis")
		return holidays, nil
	} else if err != nil {
		logrus.Error("[module:cache] GetHolidays: get str slice error: ", err)
	}
	logrus.Error("[module:cache] GetHolidays: no data")
	return []domain.Holiday{}, nil
}

func (cs *CacheService) SetHolidays(ctx context.Context, holidays []domain.Holiday) {
	if cs.cacheHolidaysTTL == 0 {
		logrus.Error("[module:cache] SetHolidays: cache holidays ttl is 0")
		return
	}

	// Удаляем старый ключ
	cs.repo.rds.Del(ctx, HOLIDAYS_KEY)

	// Сохраняем каждый праздник отдельно в список
	for _, holiday := range holidays {
		holidayJSON, err := json.Marshal(holiday)
		if err != nil {
			logrus.Error("[module:cache] SetHolidays: marshal error: ", err)
			continue
		}

		// Добавляем элемент в список
		err = cs.repo.rds.AddToStrSlice(ctx, HOLIDAYS_KEY, string(holidayJSON))
		if err != nil {
			logrus.Error("[module:cache] SetHolidays: add to str slice error: ", err)
			continue
		}
	}

	// Устанавливаем TTL для всего списка один раз
	err := cs.repo.rds.SetStrSliceTTL(ctx, HOLIDAYS_KEY, cs.cacheHolidaysTTL)
	if err != nil {
		logrus.Error("[module:cache] SetHolidays: set ttl error: ", err)
	}

	logrus.Error("[module:cache] SetHolidays: to redis")
}

func (cs *CacheService) DelHolidays(ctx context.Context) {
	err := cs.repo.rds.Del(ctx, HOLIDAYS_KEY)
	if err != nil {
		logrus.Error("[module:cache] DelHolidays: del error: ", err)
		return
	}

	logrus.Error("[module:cache] DelHolidays: from redis")
}

func (cs *CacheService) IsHoliday(ctx context.Context, date string) (bool, bool) {
	// Получаем все праздники из кеша
	holidaysJSON, err := cs.repo.rds.GetStrSlice(ctx, HOLIDAYS_KEY)
	if err != nil {
		logrus.Error("[module:cache] IsHoliday: get str slice error: ", err)
		return false, false
	}

	// Если данных нет в кеше
	if len(holidaysJSON) == 0 {
		return false, false
	}

	// Проверяем наличие даты в каждом празднике
	for _, holidayJSON := range holidaysJSON {
		var holiday domain.Holiday
		err = json.Unmarshal([]byte(holidayJSON), &holiday)
		if err != nil {
			logrus.Error("[module:cache] IsHoliday: unmarshal error: ", err)
			continue
		}

		// Сравниваем дату (формат: YYYY-MM-DD)
		if holiday.HolidayDate == date {
			logrus.Error("[module:cache] IsHoliday: found holiday: ", date)
			return true, true
		}
	}

	logrus.Error("[module:cache] IsHoliday: date not found: ", date, " hasData: true")
	return false, true
}
