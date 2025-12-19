package redis

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/pkg/helpers"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/logger"
	"github.com/redis/go-redis/v9"
)

type RDS struct {
	rdb *redis.Client
}

type Creds string

func New(creds Creds) (*RDS, error) {
	op := "redis.new"
	log := logger.GetLogger().WithOperation(op)
	log.Info("Creating new Redis client")

	pattern := regexp.MustCompile(`redis://(?P<password>[^@]+)@(?P<host>[^:]+):(?P<port>[^/]+)/(?P<dbname>[^?]+)`)
	sub := pattern.FindStringSubmatch(string(creds))

	if len(sub) != 5 {
		log.Error(helpers.ErrInvalidCredentials, "Invalid credentials format")
		return nil, helpers.ErrInvalidCredentials
	}

	password := sub[1]
	host := sub[2]

	port, err := strconv.Atoi(sub[3])
	if err != nil {
		log.Error(err, "Failed to parse port")
		return nil, err
	}

	dbIndex, err := strconv.Atoi(sub[4])
	if err != nil {
		log.Error(err, "Failed to parse database index")
		return nil, err
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: password,
		DB:       dbIndex,
	})

	log.WithFields(logger.Fields{
		"host": host,
		"port": port,
		"db":   dbIndex,
	}).Success("Redis client created successfully")

	return &RDS{rdb: rdb}, nil
}

// NewFromClient создает RDS из готового redis.Client
func NewFromClient(client *redis.Client) *RDS {
	op := "redis.newFromClient"
	log := logger.GetLogger().WithOperation(op)
	log.Debug("Creating RDS from existing Redis client")
	return &RDS{rdb: client}
}

// fetch str from redis
func (rds *RDS) GetStr(ctx context.Context, key string) (string, error) {
	op := "redis.getStr"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"key": key})
	log.Debug("Getting string from Redis")

	v, err := rds.rdb.Get(ctx, key).Result()

	if errors.Is(err, redis.Nil) {
		log.Debug("Key not found")
		return "", nil
	} else if err != nil {
		log.Error(err, "Failed to get string from Redis")
		return "", err
	}

	log.WithFields(logger.Fields{"value_length": len(v)}).Debug("String retrieved successfully")
	return v, nil
}

// ttl - in seconds
func (rds *RDS) SetStr(ctx context.Context, key, value string, ttl int) error {
	op := "redis.setStr"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"key":        key,
		"value_length": len(value),
		"ttl":        ttl,
	})
	log.Debug("Setting string in Redis")

	err := rds.rdb.Set(ctx, key, value, time.Duration(ttl)*time.Second).Err()
	if err != nil {
		log.Error(err, "Failed to set string in Redis")
		return err
	}

	log.Success("String set successfully")
	return nil
}

func (rds *RDS) GetStrSlice(ctx context.Context, key string) ([]string, error) {
	op := "redis.getStrSlice"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"key": key})
	log.Debug("Getting string slice from Redis")

	// Проверяем тип ключа перед использованием LRange
	keyType, err := rds.rdb.Type(ctx, key).Result()
	if err != nil {
		// Если ключа нет, возвращаем пустой список
		if errors.Is(err, redis.Nil) {
			log.Debug("Key not found")
			return []string{}, nil
		}
		log.Error(err, "Failed to get key type")
		return []string{}, err
	}

	// Если ключ не список, удаляем его и возвращаем пустой список
	if keyType != "list" {
		log.Warn("Key is not a list, deleting. Key type: %s", keyType)
		rds.rdb.Del(ctx, key)
		return []string{}, nil
	}

	v, err := rds.rdb.LRange(ctx, key, 0, -1).Result()

	if errors.Is(err, redis.Nil) {
		log.Debug("Key not found")
		return []string{}, nil
	} else if err != nil {
		log.Error(err, "Failed to get string slice from Redis")
		return []string{}, err
	}

	if len(v) == 0 {
		log.Debug("Empty slice")
		return []string{}, nil
	}

	log.WithFields(logger.Fields{"count": len(v)}).Debug("String slice retrieved successfully")
	return v, nil
}

func (rds *RDS) SetStrSlice(ctx context.Context, key string, value string, ttl int) error {
	op := "redis.setStrSlice"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"key":        key,
		"value_length": len(value),
		"ttl":        ttl,
	})
	log.Debug("Setting string slice in Redis")

	// Удаляем старый ключ, если он существует (может быть другого типа)
	rds.rdb.Del(ctx, key)

	// Сохраняем как список - добавляем элемент в список
	err := rds.rdb.LPush(ctx, key, value).Err()
	if err != nil {
		log.Error(err, "Failed to push value to list")
		return err
	}

	// Устанавливаем TTL для списка
	err = rds.rdb.Expire(ctx, key, time.Duration(ttl)*time.Second).Err()
	if err != nil {
		log.Error(err, "Failed to set TTL for list")
		return err
	}

	log.Success("String slice set successfully")
	return nil
}

// AddToStrSlice добавляет элемент в список без удаления существующих
func (rds *RDS) AddToStrSlice(ctx context.Context, key string, value string) error {
	op := "redis.addToStrSlice"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"key":        key,
		"value_length": len(value),
	})
	log.Debug("Adding value to string slice")

	err := rds.rdb.LPush(ctx, key, value).Err()
	if err != nil {
		log.Error(err, "Failed to add value to list")
		return err
	}

	log.Success("Value added to list successfully")
	return nil
}

// SetStrSliceTTL устанавливает TTL для ключа
func (rds *RDS) SetStrSliceTTL(ctx context.Context, key string, ttl int) error {
	op := "redis.setStrSliceTTL"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"key": key,
		"ttl": ttl,
	})
	log.Debug("Setting TTL for key")

	err := rds.rdb.Expire(ctx, key, time.Duration(ttl)*time.Second).Err()
	if err != nil {
		log.Error(err, "Failed to set TTL")
		return err
	}

	log.Success("TTL set successfully")
	return nil
}

func (rds *RDS) Del(ctx context.Context, key string) error {
	op := "redis.del"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"key": key})
	log.Debug("Deleting key from Redis")

	err := rds.rdb.Del(ctx, key).Err()
	if err != nil {
		log.Error(err, "Failed to delete key")
		return err
	}

	log.Success("Key deleted successfully")
	return nil
}

func (rds *RDS) Ping(ctx context.Context) error {
	op := "redis.ping"
	log := logger.GetLogger().WithOperation(op)
	log.Debug("Pinging Redis")

	err := rds.rdb.Ping(ctx).Err()
	if err != nil {
		log.Error(err, "Failed to ping Redis")
		return err
	}

	log.Success("Redis ping successful")
	return nil
}

func (rds *RDS) Exists(ctx context.Context, key string, value string) (bool, bool, error) {
	op := "redis.exists"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"key":   key,
		"value": value,
	})
	log.Debug("Checking if value exists in Redis list")

	// Проверяем, существует ли ключ
	keyExists, err := rds.rdb.Exists(ctx, key).Result()
	if err != nil {
		log.Error(err, "Failed to check if key exists")
		return false, false, err
	}

	// Если ключа нет, возвращаем false и флаг что данных нет
	if keyExists == 0 {
		log.Debug("Key does not exist")
		return false, false, nil
	}

	// Проверяем тип ключа - должен быть список
	keyType, err := rds.rdb.Type(ctx, key).Result()
	if err != nil {
		log.Error(err, "Failed to get key type")
		return false, true, err
	}

	// Если ключ не список, удаляем его и возвращаем false
	if keyType != "list" {
		log.Warn("Key is not a list, deleting and returning false. Key type: %s", keyType)
		rds.rdb.Del(ctx, key)
		return false, false, nil
	}

	// Ключ существует и является списком, проверяем наличие элемента в списке
	_, err = rds.rdb.LPos(ctx, key, value, redis.LPosArgs{}).Result()

	if errors.Is(err, redis.Nil) {
		// Элемент не найден, но ключ существует
		log.Debug("Value not found in list")
		return false, true, nil
	} else if err != nil {
		log.Error(err, "Failed to check value position in list")
		return false, true, err
	}

	// Элемент найден
	log.Debug("Value found in list")
	return true, true, nil
}
