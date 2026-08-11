package utils

import (
	"fmt"
	"time"

	"github.com/awiipp/go-library/user-service/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateToken(cfg *config.Config, userID, role string) (string, error) {
	now := time.Now()

	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    cfg.JWT.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(cfg.JWT.ExpiresIn) * time.Second)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(cfg.JWT.PrivateKey)
}

func VerifyToken(cfg *config.Config, tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
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
