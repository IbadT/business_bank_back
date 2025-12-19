package cache

import (
	"context"
	"encoding/json"

	"github.com/IbadT/business_bank_back/services/matematika/internal/database"
	"github.com/IbadT/business_bank_back/services/matematika/internal/domain"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/helpers"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/logger"
)

var (
	HOLIDAYS_KEY string = "holidays"
	GATEWAYS_KEY string = "gateways"
)

type CacheService struct {
	repo             *Repository
	cacheHolidaysTTL int
	cacheGatewaysTTL int
}

func New(repo *Repository) *CacheService {
	op := "cache.new"
	log := logger.GetLogger().WithOperation(op)
	log.Info("Creating new cache service")

	holidaysTTL := database.GetEnvInt("CACHE_HOLIDAYS_TTL", 60*60*24)
	gatewaysTTL := database.GetEnvInt("CACHE_GATEWAYS_TTL", 60*60*24)
	
	log.WithFields(logger.Fields{
		"holidays_ttl": holidaysTTL,
		"gateways_ttl": gatewaysTTL,
	}).Success("Cache service created successfully")

	return &CacheService{
		repo:             repo,
		cacheHolidaysTTL: holidaysTTL,
		cacheGatewaysTTL: gatewaysTTL,
	}
}

func (cs *CacheService) GetHolidays(ctx context.Context) ([]domain.Holiday, error) {
	op := "cache.getHolidays"
	log := logger.GetLogger().WithOperation(op)
	log.Debug("Getting holidays from cache")

	if cs.cacheHolidaysTTL == 0 {
		log.Error(helpers.ErrCacheHolidaysTTLZero, "Cache holidays TTL is zero")
		return []domain.Holiday{}, helpers.ErrCacheHolidaysTTLZero
	}

	holidaysJSON, err := cs.repo.rds.GetStrSlice(ctx, HOLIDAYS_KEY)
	if err == nil && len(holidaysJSON) > 0 {
		holidays := make([]domain.Holiday, len(holidaysJSON))
		for i, holidayJSON := range holidaysJSON {
			err = json.Unmarshal([]byte(holidayJSON), &holidays[i])
			if err != nil {
				log.Error(err, "Failed to unmarshal holiday JSON at index %d", i)
				return []domain.Holiday{}, err
			}
		}
		log.WithFields(logger.Fields{"count": len(holidays)}).Debug("Holidays retrieved from Redis")
		return holidays, nil
	} else if err != nil {
		log.Error(err, "Failed to get holidays from Redis")
	}
	log.Debug("No holidays data in cache")
	return []domain.Holiday{}, nil
}

func (cs *CacheService) SetHolidays(ctx context.Context, holidays []domain.Holiday) {
	op := "cache.setHolidays"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"count": len(holidays)})
	log.Info("Setting holidays in cache")

	if cs.cacheHolidaysTTL == 0 {
		log.Error(helpers.ErrCacheHolidaysTTLZero, "Cache holidays TTL is zero")
		return
	}

	// Удаляем старый ключ
	cs.repo.rds.Del(ctx, HOLIDAYS_KEY)

	// Сохраняем каждый праздник отдельно в список
	for i, holiday := range holidays {
		holidayJSON, err := json.Marshal(holiday)
		if err != nil {
			log.Error(err, "Failed to marshal holiday at index %d", i)
			continue
		}

		// Добавляем элемент в список
		err = cs.repo.rds.AddToStrSlice(ctx, HOLIDAYS_KEY, string(holidayJSON))
		if err != nil {
			log.Error(err, "Failed to add holiday to Redis list at index %d", i)
			continue
		}
	}

	// Устанавливаем TTL для всего списка один раз
	err := cs.repo.rds.SetStrSliceTTL(ctx, HOLIDAYS_KEY, cs.cacheHolidaysTTL)
	if err != nil {
		log.Error(err, "Failed to set TTL for holidays list")
		return
	}

	log.Success("Holidays set in Redis successfully")
}

func (cs *CacheService) DelHolidays(ctx context.Context) {
	op := "cache.delHolidays"
	log := logger.GetLogger().WithOperation(op)
	log.Info("Deleting holidays from cache")

	err := cs.repo.rds.Del(ctx, HOLIDAYS_KEY)
	if err != nil {
		log.Error(err, "Failed to delete holidays from Redis")
		return
	}

	log.Success("Holidays deleted from Redis successfully")
}

func (cs *CacheService) IsHoliday(ctx context.Context, date string) (bool, bool) {
	op := "cache.isHoliday"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"date": date})
	log.Debug("Checking if date is holiday")

	// Получаем все праздники из кеша
	holidaysJSON, err := cs.repo.rds.GetStrSlice(ctx, HOLIDAYS_KEY)
	if err != nil {
		log.Error(err, "Failed to get holidays from Redis")
		return false, false
	}

	// Если данных нет в кеше
	if len(holidaysJSON) == 0 {
		log.Debug("No holidays data in cache")
		return false, false
	}

	// Проверяем наличие даты в каждом празднике
	for i, holidayJSON := range holidaysJSON {
		var holiday domain.Holiday
		err = json.Unmarshal([]byte(holidayJSON), &holiday)
		if err != nil {
			log.Error(err, "Failed to unmarshal holiday JSON at index %d", i)
			continue
		}

		// Сравниваем дату (формат: YYYY-MM-DD)
		if holiday.HolidayDate == date {
			log.Debug("Holiday found for date")
			return true, true
		}
	}

	log.Debug("Date not found in holidays list")
	return false, true
}

func (cs *CacheService) GetGateways(ctx context.Context) ([]domain.Gateway, error) {
	op := "cache.getGateways"
	log := logger.GetLogger().WithOperation(op)
	log.Debug("Getting gateways from cache")

	if cs.cacheGatewaysTTL == 0 {
		log.Error(helpers.ErrCacheGatewaysTTLZero, "Cache gateways TTL is zero")
		return []domain.Gateway{}, helpers.ErrCacheGatewaysTTLZero
	}

	gatewaysJSON, err := cs.repo.rds.GetStrSlice(ctx, GATEWAYS_KEY)
	if err == nil && len(gatewaysJSON) > 0 {
		gateways := make([]domain.Gateway, len(gatewaysJSON))
		for i, gatewayJSON := range gatewaysJSON {
			err = json.Unmarshal([]byte(gatewayJSON), &gateways[i])
			if err != nil {
				log.Error(err, "Failed to unmarshal gateway JSON at index %d", i)
				return []domain.Gateway{}, err
			}
		}
		log.WithFields(logger.Fields{"count": len(gateways)}).Debug("Gateways retrieved from Redis")
		return gateways, nil
	} else if err != nil {
		log.Error(err, "Failed to get gateways from Redis")
	}
	log.Debug("No gateways data in cache")
	return []domain.Gateway{}, nil
}

func (cs *CacheService) SetGateways(ctx context.Context, gateways []domain.Gateway) {
	op := "cache.setGateways"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"count": len(gateways)})
	log.Info("Setting gateways in cache")

	if cs.cacheGatewaysTTL == 0 {
		log.Error(helpers.ErrCacheGatewaysTTLZero, "Cache gateways TTL is zero")
		return
	}

	cs.repo.rds.Del(ctx, GATEWAYS_KEY)

	for i, gateway := range gateways {
		gatewaysJSON, err := json.Marshal(gateway)
		if err != nil {
			log.Error(err, "Failed to marshal gateway at index %d", i)
			continue
		}

		err = cs.repo.rds.AddToStrSlice(ctx, GATEWAYS_KEY, string(gatewaysJSON))
		if err != nil {
			log.Error(err, "Failed to add gateway to Redis list at index %d", i)
			continue
		}
	}

	err := cs.repo.rds.SetStrSliceTTL(ctx, GATEWAYS_KEY, cs.cacheGatewaysTTL)
	if err != nil {
		log.Error(err, "Failed to set TTL for gateways list")
		return
	}

	log.Success("Gateways set in Redis successfully")
}
