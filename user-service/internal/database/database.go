package database

import (
	"fmt"
	"log"
	"time"

	"github.com/awiipp/go-library/user-service/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	maxRetries = 10
	retryDelay = 2 * time.Second

	maxOpenConns = 25
	maxIdleConns = 25
	ConnMaxIdle  = 5 * time.Minute
	ConnMaxLife  = 30 * time.Minute
)

func Connect(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Name,
		cfg.DB.SSLMode,
	)

	var db *gorm.DB
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Warn),
		})
		if err == nil {
			log.Println("database connected")
			break
		}

		log.Printf("Waiting for database (%d/%d): %v", attempt, maxRetries, err)
		if attempt < maxRetries {
			time.Sleep(retryDelay)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("could not connect to database after %d attempts: %v", maxRetries, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("database.Connect get sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxIdleTime(ConnMaxIdle)
	sqlDB.SetConnMaxLifetime(ConnMaxLife)

	return db, nil
}
