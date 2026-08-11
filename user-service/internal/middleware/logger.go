package middleware

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
)

func Logger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		status := c.Response().StatusCode()
		duration := time.Since(start).Round(time.Microsecond)

		message := fmt.Sprintf(
			"%s %s %d %s %s",
			c.Method(),
			c.Path(),
			status,
			duration,
			c.IP(),
		)

		if err != nil || status >= 500 {
			slog.Info(message)
			return err
		}

		slog.Info(message)
		return nil
	}
}
