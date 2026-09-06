package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/awiipp/go-library/internal/cache"
	"github.com/awiipp/go-library/internal/domain"
	pkgerrors "github.com/awiipp/go-library/pkg/errors"
	"github.com/google/uuid"
)

type bookRepository struct {
	db        *sql.DB
	bookCache *cache.BookCache
}

func NewBookRepository(db *sql.DB, bookCache *cache.BookCache) domain.BookRepository {
	return &bookRepository{
		db:        db,
		bookCache: bookCache,
	}
}

func (r *bookRepository) FindAll(ctx context.Context) ([]*domain.Book, error) {
	query := `SELECT id, title, author, description, stock, created_at, updated_at FROM books ORDER BY created_at DESC`

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
			&book.Title,
			&book.Author,
			&book.Description,
			&book.Stock,
			&book.CreatedAt,
			&book.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("repository.FindAll: %w", err)
		}

		books = append(books, book)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository.FindAll: %w", err)
	}

	return books, nil
}

func (r *bookRepository) FindByID(ctx context.Context, id string) (*domain.Book, error) {
	book, err := r.bookCache.Get(ctx, id)
	if err != nil {
		log.Printf("repository.FindByID cache: %v", err)
	}

	if book != nil {
		return book, nil
	}

	query := `SELECT id, title, author, description, stock, created_at, updated_at FROM books WHERE id = $1`

	book = &domain.Book{}

	err = r.db.QueryRowContext(ctx, query, id).Scan(
		&book.ID,
		&book.Title,
		&book.Author,
		&book.Description,
		&book.Stock,
		&book.CreatedAt,
		&book.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pkgerrors.ErrNotFound
		}

		return nil, fmt.Errorf("repository.FindByID: %w", err)
	}

	if err := r.bookCache.Set(ctx, book); err != nil {
		log.Printf("failed to cache book %s: %v", id, err)
	}

	return book, nil
}

func (r *bookRepository) Save(ctx context.Context, book *domain.Book) (*domain.Book, error) {
	book.ID = uuid.NewString()
	book.CreatedAt = time.Now()
	book.UpdatedAt = time.Now()

	query := `
		INSERT INTO books (id, title, author, description, stock, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(
		ctx, query,
		book.ID,
		book.Title,
		book.Author,
		book.Description,
		book.Stock,
		book.CreatedAt,
		book.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("repository.Save: %w", err)
	}

	if err := r.bookCache.Set(ctx, book); err != nil {
		fmt.Printf("failed to cache book %s: %v", book.ID, err)
	}

	return book, nil
}

func (r *bookRepository) Update(ctx context.Context, book *domain.Book) (*domain.Book, error) {
	book.UpdatedAt = time.Now()

	query := `
		UPDATE books
		SET title = $1, author = $2, description = $3, stock = $4, updated_at = $5
		WHERE id = $6
	`

	result, err := r.db.ExecContext(
		ctx, query,
		book.Title,
		book.Author,
		book.Description,
		book.Stock,
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

	if err := r.bookCache.Set(ctx, book); err != nil {
		fmt.Printf("failed to update cache for book %s: %v", book.ID, err)
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

	if err := r.bookCache.Delete(ctx, id); err != nil {
		fmt.Printf("failed to delete cache for book %s: %v", id, err)
	}

	return nil
}

func (r *bookRepository) DecreaseStock(ctx context.Context, bookID string) error {
	exec := getExecutor(ctx, r.db)
	query := `UPDATE books SET stock = stock - 1 WHERE id = $1 AND stock > 0`

	result, err := exec.ExecContext(ctx, query, bookID)
	if err != nil {
		return fmt.Errorf("repository.DecreaseStock: %w", err)
	}

	rowAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository.DecreaseStock: %w", err)
	}
	if rowAffected == 0 {
		return pkgerrors.ErrOutOfStock
	}

	return nil
}

func (r *bookRepository) IncreaseStock(ctx context.Context, bookID string) error {
	exec := getExecutor(ctx, r.db)
	query := `UPDATE books SET stock = stock + 1 WHERE id = $1`

	_, err := exec.ExecContext(ctx, query, bookID)
	if err != nil {
		return fmt.Errorf("repository.IncreaseStock: %w", err)
	}

	return nil
}

func (r *bookRepository) InvalidateCache(ctx context.Context, bookID string) error {
	return r.bookCache.Delete(ctx, bookID)
}
