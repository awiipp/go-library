package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/awiipp/go-library/internal/cache"
	"github.com/awiipp/go-library/internal/config"
	"github.com/awiipp/go-library/internal/database"
	"github.com/awiipp/go-library/internal/handler"
	"github.com/awiipp/go-library/internal/repository"
	"github.com/awiipp/go-library/internal/server"
	"github.com/awiipp/go-library/internal/usecase"
)

func main() {
	cfg := config.Load()

	// database connection
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	// redis connection
	redisClient, err := cache.RedisClient(cfg.Redis)
	if err != nil {
		log.Fatalf("failed to connect redis: %v", err)
	}

	defer redisClient.Close()

	// wiring repository, usecase, handler
	bookCache := cache.NewBookCache(redisClient)
	bookRepo := repository.NewBookRepository(db, bookCache)
	bookUsecase := usecase.NewBookUsecase(bookRepo)
	bookHandler := handler.NewBookHandler(bookUsecase)

	// http server
	app := server.New(bookHandler)

	go func() {
		if err := app.Listen(":" + cfg.App.Port); err != nil {
			log.Fatal(err)
		}
	}()

	log.Printf("server listening on: %s", cfg.App.Port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")
	if err := app.Shutdown(); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}
