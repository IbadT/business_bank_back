package database

import (
	"database/sql"
	"fmt"
	"os"

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
		Host:     GetEnv("POSTGRES_HOST", "localhost"),
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

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func GetEnv(key, defaulValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaulValue
}
