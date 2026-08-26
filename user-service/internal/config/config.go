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
	App  AppConfig
	DB   DBConfig
	JWT  JWTConfig
	Auth AuthConfig
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

type JWTConfig struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
	Issuer     string
	ExpiresIn  int64 // seconds
}

type AuthConfig struct {
	RefreshTokenTTL int64 // days
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	jwtCfg, err := loadJWTConfig(
		getEnv("JWT_PRIVATE_KEY_PATH", "./certs/private.pem"),
		getEnv("JWT_PUBLIC_KEY_PATH", "./certs/public.pem"),
		getEnv("JWT_ISSUER", "user-service"),
		getEnvInt64("JWT_EXPIRES_IN", 900),
	)
	if err != nil {
		return nil, fmt.Errorf("config.Load: %w", err)
	}

	return &Config{
		App: AppConfig{
			Name: getEnv("APP_NAME", "go-library-user-service"),
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
		JWT: *jwtCfg,
		Auth: AuthConfig{
			RefreshTokenTTL: getEnvInt64("REFRESH_TOKEN_TTL", 7),
		},
	}, nil
}

func loadJWTConfig(privatePath, publicPath, issuer string, expiresIn int64) (*JWTConfig, error) {
	// private key
	privBytes, err := os.ReadFile(privatePath)
	if err != nil {
		return nil, fmt.Errorf("config.Load read private: %w", err)
	}

	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(privBytes)
	if err != nil {
		return nil, fmt.Errorf("config.Load parse private: %w", err)
	}

	// public key
	pubBytes, err := os.ReadFile(publicPath)
	if err != nil {
		return nil, fmt.Errorf("config.Load read public: %w", err)
	}

	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(pubBytes)
	if err != nil {
		return nil, fmt.Errorf("config.Load parse private: %w", err)
	}

	return &JWTConfig{
		PrivateKey: privKey,
		PublicKey:  pubKey,
		Issuer:     issuer,
		ExpiresIn:  expiresIn,
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return defaultValue
	}

	return parsed
}
