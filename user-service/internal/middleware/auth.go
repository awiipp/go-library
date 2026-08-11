package middleware

import (
	"net/http"
	"strings"

	"github.com/awiipp/go-library/user-service/internal/config"
	"github.com/awiipp/go-library/user-service/pkg/response"
	"github.com/awiipp/go-library/user-service/pkg/utils"
	"github.com/gofiber/fiber/v2"
)

func RequireAuth(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return response.Error(c, http.StatusUnauthorized, "missing authorization header")
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			return response.Error(c, http.StatusUnauthorized, "invalid authorization header format")
		}

		claims, err := utils.VerifyToken(cfg, parts[1])
		if err != nil {
			return response.Error(c, http.StatusUnauthorized, "invalid or expired token")
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("role", claims.Role)

		return c.Next()
	}
}
