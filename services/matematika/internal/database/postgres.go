package database

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

type Postgres struct {
	Config Config
	Client *sql.DB
}

func NewConfig() *Config {
	return &Config{
		Host:     GetEnv("POSTGRES_HOST", "postgres"),
		Port:     5432,
		User:     GetEnv("POSTGRES_USER", "postgres"),
		Password: GetEnv("POSTGRES_PASSWORD", "postgres"),
		DBName:   GetEnv("POSTGRES_DB", "matematika"),
		// Host:     "localhost",
		// Port:     5432,
		// User:     "postgres",
		// Password: "postgres",
		// DBName:   "matematika",
	}
}

func InitDB() (*gorm.DB, error) {
	config := NewConfig()

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.User, config.Password, config.DBName)

	logrus.Infof("🔌 Connecting to database: host=%s port=%d user=%s dbname=%s",
		config.Host, config.Port, config.User, config.DBName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: nil, // Можно включить логирование для отладки
	})
	if err != nil {
		return nil, err
	}

	// Получаем sql.DB для проверки подключения
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Проверяем подключение
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func GetEnv(key, defaulValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaulValue
}

// GetEnvInt получает переменную окружения как int
// Параметры:
//   - key: имя переменной окружения
//   - defaultValue: значение по умолчанию (int)
//
// Возвращает:
//   - int значение переменной окружения или defaultValue если не найдено/невалидно
func GetEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if result, err := strconv.Atoi(value); err == nil {
			return result
		}
	}
	return defaultValue
}

// HealthCheckDB проверяет подключение к базе данных
// Параметры:
//   - gormDB: GORM подключение к БД
//
// Возвращает:
//   - "connected" если БД доступна
//   - "disconnected" если БД недоступна или gormDB == nil
func HealthCheckDB(gormDB *gorm.DB) string {
	if gormDB == nil {
		return "disconnected"
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return "disconnected"
	}

	if err := sqlDB.Ping(); err != nil {
		return "disconnected"
	}

	return "connected"
}
