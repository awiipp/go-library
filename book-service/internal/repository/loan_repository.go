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

type loanRepository struct {
	db *sql.DB
}

func NewLoanRepository(db *sql.DB) domain.LoanRepository {
	return &loanRepository{db: db}
}

func (r *loanRepository) Create(ctx context.Context, tx *sql.Tx, loan *domain.Loan) error {
	loan.ID = uuid.NewString()

	query := `
		INSERT INTO loans (id, book_id, user_id, status, borrowed_at, due_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := tx.ExecContext(ctx, query, loan.ID, loan.BookID, loan.UserID, loan.Status, loan.BorrowedAt, loan.DueAt)
	if err != nil {
		return fmt.Errorf("repository.Create: %w", err)
	}

	return nil
}

func (r *loanRepository) FindByID(ctx context.Context, id string) (*domain.Loan, error) {
	query := `SELECT id, book_id, user_id, status, borrowed_at, due_at, returned_at FROM loans WHERE id = $1`

	loan := &domain.Loan{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&loan.ID,
		&loan.BookID,
		&loan.UserID,
		&loan.Status,
		&loan.BorrowedAt,
		&loan.DueAt,
		&loan.ReturnedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, pkgerrors.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository.FindByID: %w", err)
	}

	return loan, nil
}

func (r *loanRepository) FindActiveByBookAndUser(ctx context.Context, bookID, userID string) (*domain.Loan, error) {
	query := `
		SELECT id, book_id, user_id, status, borrowed_at, due_at, returned_at
		FROM loans WHERE book_id = $1 AND user_id = $2 AND status = 'active'
	`

	loan := &domain.Loan{}
	err := r.db.QueryRowContext(ctx, query, bookID, userID).Scan(
		&loan.ID,
		&loan.BookID,
		&loan.UserID,
		&loan.Status,
		&loan.BorrowedAt,
		&loan.DueAt,
		&loan.ReturnedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, pkgerrors.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository.FindActiveByBookAndUser: %w", err)
	}

	return loan, err
}

func (r *loanRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.Loan, error) {
	query := `
		SELECT id, book_id, user_id, status, borrowed_at, due_at, returned_at
		FROM loans WHERE user_id = $1 ORDER BY borrowed_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("repository.FindByUserID: %w", err)
	}

	defer rows.Close()

	var loans []*domain.Loan
	for rows.Next() {
		loan := &domain.Loan{}

		err := rows.Scan(
			&loan.ID,
			&loan.BookID,
			&loan.UserID,
			&loan.Status,
			&loan.BorrowedAt,
			&loan.DueAt,
			&loan.ReturnedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("repository.FindByUserID: %w", err)
		}

		loans = append(loans, loan)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository.FindByUserID: %w", err)
	}

	return loans, nil
}

func (r *loanRepository) MarkReturned(ctx context.Context, tx *sql.Tx, id string, returnedAt time.Time) error {
	query := `UPDATE loans SET status = 'returned', returned_at = $1 WHERE id = $2`

	_, err := tx.ExecContext(ctx, query, returnedAt, id)
	if err != nil {
		return fmt.Errorf("repository.MarkReturned: %w", err)
	}

	return nil
}
