package server

import (
	"github.com/awiipp/go-library/internal/config"
	"github.com/awiipp/go-library/internal/handler"
	"github.com/awiipp/go-library/internal/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func New(bookHandler *handler.BookHandler, loanHandler *handler.LoanHandler, cfg *config.Config) *fiber.App {
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
	books.Post("/", middleware.RequireRole("admin"), bookHandler.Create)
	books.Put("/:id", middleware.RequireRole("admin"), bookHandler.Update)
	books.Delete("/:id", middleware.RequireRole("admin"), bookHandler.Delete)
	books.Post("/:id/borrow", loanHandler.Borrow)

	loans := protected.Group("/loans")
	loans.Post("/:id/return", loanHandler.Return)
	loans.Get("/me", loanHandler.MyLoan)

	return app
}
