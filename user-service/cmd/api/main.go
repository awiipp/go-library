package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/awiipp/go-library/user-service/internal/config"
	"github.com/awiipp/go-library/user-service/internal/database"
	"github.com/awiipp/go-library/user-service/internal/handler"
	"github.com/awiipp/go-library/user-service/internal/repository"
	"github.com/awiipp/go-library/user-service/internal/server"
	"github.com/awiipp/go-library/user-service/internal/usecase"
)

func main() {
	// load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// database connection
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal(err)
	}

	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// wiring repository, usecase, handler
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	userUsecase := usecase.NewUserUsecase(userRepo, refreshTokenRepo, cfg)
	userHandler := handler.NewUserHandler(userUsecase)

	// http server
	app := server.New(userHandler, cfg)

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
