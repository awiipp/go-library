package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/awiipp/go-library/internal/config"

	_ "github.com/lib/pq"
)

const (
	maxRetries = 10
	retryDelay = 2 * time.Second

	maxOpenConns = 25
	maxIdleConns = 25
	connMaxIdle  = 5 * time.Minute
	connMaxLife  = 30 * time.Minute
)

func Connect(cfg *config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Name,
		cfg.DB.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres connection: %w", err)
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxIdleTime(connMaxIdle)
	db.SetConnMaxLifetime(connMaxLife)

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if err := db.Ping(); err == nil {
			log.Println("database connected")

			return db, nil
		} else {
			log.Printf("waiting for database (%d/%d): %v", attempt, maxRetries, err)
		}

		if attempt < maxRetries {
			time.Sleep(retryDelay)
		}
	}

	_ = db.Close()

	return nil, fmt.Errorf("could not connect to database after %d attemps", maxRetries)
}
