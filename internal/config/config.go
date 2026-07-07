package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	App   AppConfig
	DB    DBConfig
	Redis RedisConfig
}

type AppConfig struct {
	Name string
	Env  string
	Port string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

func Load() *Config {
	_ = godotenv.Load()

	db, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		db = 0
	}

	return &Config{
		App: AppConfig{
			Name: getEnv("APP_NAME", "go-library"),
			Env:  getEnv("APP_ENV", "development"),
			Port: getEnv("APP_PORT", "9000"),
		},
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "postgres"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       db,
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}
