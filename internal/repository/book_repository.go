package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/awiipp/go-library/internal/domain"
	pkgerrors "github.com/awiipp/go-library/pkg/errors"
	"github.com/google/uuid"
)

type bookRepository struct {
	db *sql.DB
}

func NewBookRepository(db *sql.DB) domain.BookRepository {
	return &bookRepository{db: db}
}

func (r *bookRepository) FindAll(ctx context.Context) ([]*domain.Book, error) {
	query := `SELECT id, title, author, description, created_at, updated_at FROM books ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("repository.FindAll: %w", err)
	}

	defer rows.Close()

	var books []*domain.Book
	for rows.Next() {
		book := &domain.Book{}

		err := rows.Scan(
			&book.ID,
			&book.Author,
			&book.Description,
			&book.CreatedAt,
			&book.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("repository.FindAll: %w", err)
		}

		books = append(books, book)
	}

	return books, nil
}

func (r *bookRepository) FindByID(ctx context.Context, id string) (*domain.Book, error) {
	query := `SELECT id, title, author, description, created_at, updated_at FROM books WHERE id = $1`

	book := &domain.Book{}

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&book.ID,
		&book.Author,
		&book.Description,
		&book.CreatedAt,
		&book.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pkgerrors.ErrNotFound
		}

		return nil, fmt.Errorf("repository.FindByID: %w", err)
	}

	return book, nil
}

func (r *bookRepository) Save(ctx context.Context, book *domain.Book) (*domain.Book, error) {
	book.ID = uuid.NewString()
	book.CreatedAt = time.Now()
	book.UpdatedAt = time.Now()

	query := `
		INSERT INTO books (id, title, author, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(ctx, query,
		book.ID,
		book.Title,
		book.Author,
		book.Description,
		book.CreatedAt,
		book.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("repository.Save: %w", err)
	}

	return book, nil
}

func (r *bookRepository) Update(ctx context.Context, book *domain.Book) (*domain.Book, error) {
	book.UpdatedAt = time.Now()

	query := `
		UPDATE books
		SET title = $1, author = $2, description = $3, updated_at = $4
		WHERE id = $5
	`

	result, err := r.db.ExecContext(ctx, query,
		book.Title,
		book.Author,
		book.Description,
		book.UpdatedAt,
		book.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("repository.Update: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("repository.Update Row Affected: %w", err)
	}
	if rowsAffected == 0 {
		return nil, pkgerrors.ErrNotFound
	}

	return book, nil
}

func (r *bookRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM books WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("repository.Delete: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository.Update Row Affected: %w", err)
	}
	if rowsAffected == 0 {
		return pkgerrors.ErrNotFound
	}

	return nil
}
