package main

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/awiipp/go-library/user-service/internal/config"
	"github.com/awiipp/go-library/user-service/internal/database"
	"github.com/awiipp/go-library/user-service/internal/domain"
	"github.com/awiipp/go-library/user-service/internal/repository"
	pkgerrors "github.com/awiipp/go-library/user-service/pkg/errors"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	userRepo := repository.NewUserRepository(db, nil)

	email := os.Getenv("SEED_ADMIN_EMAIL")
	password := os.Getenv("SEED_ADMIN_PASSWORD")
	if email == "" || password == "" {
		log.Fatal("SEED_ADMIN_EMAIL and SEED_ADMIN_PASSWORD must be set")
	}

	existing, err := userRepo.FindByEmail(context.Background(), email)
	if err != nil && !errors.Is(err, pkgerrors.ErrNotFound) {
		log.Fatalf("failed to check existing admin: %v", err)
	}
	if existing != nil {
		log.Println("admin already exist, skipped seed")
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("failed to hash password: %v", err)
	}

	admin := &domain.User{
		ID:       uuid.NewString(),
		Email:    email,
		Username: "admin",
		Password: string(hashed),
		FullName: "Admin Library",
		Role:     domain.RoleAdmin,
		IsActive: true,
	}

	if err := userRepo.Create(context.Background(), admin); err != nil {
		log.Fatalf("failed to seed admin: %v", err)
	}

	log.Println("admin user seeded successfully")
}
