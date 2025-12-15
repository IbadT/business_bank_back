package redis

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type RDS struct {
	rdb *redis.Client
}

type Creds string

func New(creds Creds) (*RDS, error) {
	pattern := regexp.MustCompile(`redis://(?P<password>[^@]+)@(?P<host>[^:]+):(?P<port>[^/]+)/(?P<dbname>[^?]+)`)
	sub := pattern.FindStringSubmatch(string(creds))

	if len(sub) != 5 {
		return nil, errors.New("invalid credentials")
	}

	password := sub[1]
	host := sub[2]

	port, err := strconv.Atoi(sub[3])
	if err != nil {
		return nil, err
	}

	dbIndex, err := strconv.Atoi(sub[4])
	if err != nil {
		return nil, err
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: password,
		DB:       dbIndex,
	})

	return &RDS{rdb: rdb}, nil
}

// NewFromClient создает RDS из готового redis.Client
func NewFromClient(client *redis.Client) *RDS {
	return &RDS{rdb: client}
}

// fetch str from redis
func (rds *RDS) GetStr(ctx context.Context, key string) (string, error) {
	v, err := rds.rdb.Get(ctx, key).Result()

	if errors.Is(err, redis.Nil) {
		return "", nil
	} else if err != nil {
		return "", err
	}

	return v, nil
}

// ttl - in seconds
func (rds *RDS) SetStr(ctx context.Context, key, value string, ttl int) error {
	err := rds.rdb.Set(ctx, key, value, time.Duration(ttl)*time.Second).Err()

	return err
}

func (rds *RDS) GetStrSlice(ctx context.Context, key string) ([]string, error) {
	// Проверяем тип ключа перед использованием LRange
	keyType, err := rds.rdb.Type(ctx, key).Result()
	if err != nil {
		// Если ключа нет, возвращаем пустой список
		if errors.Is(err, redis.Nil) {
			return []string{}, nil
		}
		return []string{}, err
	}

	// Если ключ не список, удаляем его и возвращаем пустой список
	if keyType != "list" {
		logrus.Warn("[module:redis] GetStrSlice: key is not a list, deleting. Key type: ", keyType)
		rds.rdb.Del(ctx, key)
		return []string{}, nil
	}

	v, err := rds.rdb.LRange(ctx, key, 0, -1).Result()

	if errors.Is(err, redis.Nil) {
		return []string{}, nil
	} else if err != nil {
		return []string{}, err
	}

	if len(v) == 0 {
		return []string{}, nil
	}

	return v, nil
}

func (rds *RDS) SetStrSlice(ctx context.Context, key string, value string, ttl int) error {
	// Удаляем старый ключ, если он существует (может быть другого типа)
	rds.rdb.Del(ctx, key)

	// Сохраняем как список - добавляем элемент в список
	err := rds.rdb.LPush(ctx, key, value).Err()
	if err != nil {
		return err
	}

	// Устанавливаем TTL для списка
	err = rds.rdb.Expire(ctx, key, time.Duration(ttl)*time.Second).Err()
	return err
}

// AddToStrSlice добавляет элемент в список без удаления существующих
func (rds *RDS) AddToStrSlice(ctx context.Context, key string, value string) error {
	err := rds.rdb.LPush(ctx, key, value).Err()
	return err
}

// SetStrSliceTTL устанавливает TTL для ключа
func (rds *RDS) SetStrSliceTTL(ctx context.Context, key string, ttl int) error {
	err := rds.rdb.Expire(ctx, key, time.Duration(ttl)*time.Second).Err()
	return err
}

func (rds *RDS) Del(ctx context.Context, key string) error {
	err := rds.rdb.Del(ctx, key).Err()
	return err
}

func (rds *RDS) Ping(ctx context.Context) error {
	err := rds.rdb.Ping(ctx).Err()
	return err
}

func (rds *RDS) Exists(ctx context.Context, key string, value string) (bool, bool, error) {
	// Проверяем, существует ли ключ
	keyExists, err := rds.rdb.Exists(ctx, key).Result()
	if err != nil {
		logrus.Error("[module:redis] Exists: key exists check error: ", err)
		return false, false, err
	}

	// Если ключа нет, возвращаем false и флаг что данных нет
	if keyExists == 0 {
		return false, false, nil
	}

	// Проверяем тип ключа - должен быть список
	keyType, err := rds.rdb.Type(ctx, key).Result()
	if err != nil {
		logrus.Error("[module:redis] Exists: type check error: ", err)
		return false, true, err
	}

	// Если ключ не список, удаляем его и возвращаем false
	if keyType != "list" {
		logrus.Warn("[module:redis] Exists: key is not a list, deleting and returning false. Key type: ", keyType)
		rds.rdb.Del(ctx, key)
		return false, false, nil
	}

	// Ключ существует и является списком, проверяем наличие элемента в списке
	_, err = rds.rdb.LPos(ctx, key, value, redis.LPosArgs{}).Result()

	if errors.Is(err, redis.Nil) {
		// Элемент не найден, но ключ существует
		return false, true, nil
	} else if err != nil {
		logrus.Error("[module:redis] Exists: lpos error: ", err)
		return false, true, err
	}

	// Элемент найден
	return true, true, nil
}
