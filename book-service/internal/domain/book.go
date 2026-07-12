package domain

import (
	"context"
	"time"

	"github.com/awiipp/go-library/internal/dto"
)

type Book struct {
	ID          string
	Title       string
	Author      string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type BookRepository interface {
	FindAll(ctx context.Context) ([]*Book, error)
	FindByID(ctx context.Context, id string) (*Book, error)
	Save(ctx context.Context, book *Book) (*Book, error)
	Update(ctx context.Context, book *Book) (*Book, error)
	Delete(ctx context.Context, id string) error
}

type BookUsecase interface {
	GetAll(ctx context.Context) ([]*dto.BookResponse, error)
	GetByID(ctx context.Context, id string) (*dto.BookResponse, error)
	Create(ctx context.Context, req *dto.CreateBookRequest) (*dto.BookResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdateBookRequest) (*dto.BookResponse, error)
	Delete(ctx context.Context, id string) error
}
