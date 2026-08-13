package config

import (
	"crypto/rsa"
	"fmt"
	"os"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

type Config struct {
	App   AppConfig
	DB    DBConfig
	Redis RedisConfig
	JWT   JWTConfig
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

type JWTConfig struct {
	PublicKey *rsa.PublicKey
	Issuer    string
}

func Load() *Config {
	_ = godotenv.Load()

	jwtCfg, err := loadJWTConfig(
		getEnv("JWT_PUBLIC_KEY_PATH", "./certs/public.pem"),
		getEnv("JWT_ISSUER", "user-service"),
	)

	db, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		db = 0
	}

	return &Config{
		App: AppConfig{
			Name: getEnv("APP_NAME", "go-library-book-service"),
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
		JWT: *jwtCfg,
	}
}

func loadJWTConfig(publicPath, issuer string) (*JWTConfig, error) {
	pubBytes, err := os.ReadFile(publicPath)
	if err != nil {
		return nil, fmt.Errorf("config.Load read publicKey: %w", err)
	}

	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(pubBytes)
	if err != nil {
		return nil, fmt.Errorf("config.Load parse publicKey: %w", err)
	}

	return &JWTConfig{
		PublicKey: pubKey,
		Issuer:    issuer,
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}
