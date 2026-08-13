package server

import (
	"github.com/awiipp/go-library/internal/config"
	"github.com/awiipp/go-library/internal/handler"
	"github.com/awiipp/go-library/internal/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func New(bookHandler *handler.BookHandler, cfg *config.Config) *fiber.App {
	app := fiber.New()

	// global middleware
	app.Use(recover.New())
	app.Use(middleware.Logger())

	// routes
	v1 := app.Group("/v1/api")

	// auth middleware
	protected := v1.Group("/", middleware.RequireAuth(cfg))

	books := protected.Group("/books")
	books.Get("/", bookHandler.GetAll)
	books.Get("/:id", bookHandler.GetByID)
	books.Post("/", bookHandler.Create)
	books.Put("/:id", bookHandler.Update)
	books.Delete("/:id", bookHandler.Delete)

	return app
}
