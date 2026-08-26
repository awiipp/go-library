package server

import (
	"github.com/awiipp/go-library/user-service/internal/config"
	"github.com/awiipp/go-library/user-service/internal/handler"
	"github.com/awiipp/go-library/user-service/internal/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func New(userHandler *handler.UserHandler, cfg *config.Config) *fiber.App {
	app := fiber.New()

	// global middleware
	app.Use(recover.New())
	app.Use(middleware.Logger())

	// routes
	v1 := app.Group("/v1/api")

	auth := v1.Group("/auth")
	auth.Post("/register", userHandler.Register)
	auth.Post("/login", userHandler.Login)
	auth.Post("/refresh", userHandler.RefreshToken)

	// auth middleware
	protected := v1.Group("/users", middleware.RequireAuth(cfg))
	protected.Get("/profile", userHandler.Profile)

	return app
}
