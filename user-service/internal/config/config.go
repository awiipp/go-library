package config

import (
	"crypto/rsa"
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

type Config struct {
	JWT JWTConfig
}

type JWTConfig struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
	Issuer     string
	ExpiresIn  int64 // second
}

func Load(privatePath, publicPath, issuer string, expiresIn int64) (*Config, error) {
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

	return &Config{
		JWT: JWTConfig{
			PrivateKey: privKey,
			PublicKey:  pubKey,
			Issuer:     issuer,
			ExpiresIn:  expiresIn,
		},
	}, nil
}
