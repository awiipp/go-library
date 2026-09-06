package domain

import (
	"context"
	"time"
)

type Book struct {
	ID          string
	Title       string
	Author      string
	Description string
	Stock       int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type BookRepository interface {
	FindAll(ctx context.Context) ([]*Book, error)
	FindByID(ctx context.Context, id string) (*Book, error)
	Save(ctx context.Context, book *Book) (*Book, error)
	Update(ctx context.Context, book *Book) (*Book, error)
	Delete(ctx context.Context, id string) error
	DecreaseStock(ctx context.Context, bookID string) error
	IncreaseStock(ctx context.Context, bookID string) error
}
