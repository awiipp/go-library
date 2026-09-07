package usecase_test

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/awiipp/go-library/user-service/internal/config"
)

func newTestConfig(t *testing.T) *config.Config {
	t.Helper()
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate test RSA key: %v", err)
	}

	return &config.Config{
		JWT: config.JWTConfig{
			PrivateKey: privKey,
			PublicKey:  &privKey.PublicKey,
			Issuer:     "user-service",
			ExpiresIn:  900,
		},
	}
}
