package utils

import (
	"fmt"

	"github.com/awiipp/go-library/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func VerifyToken(cfg *config.Config, tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodRS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["Alg"])
		}

		return cfg.JWT.PublicKey, nil
	}, jwt.WithIssuer(cfg.JWT.Issuer))
	if err != nil {
		return nil, fmt.Errorf("utils.VerifyToken: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("utils.VerifyToken: invalid token")
	}

	return claims, nil
}
