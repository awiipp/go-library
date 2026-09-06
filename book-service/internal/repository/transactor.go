package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/awiipp/go-library/internal/domain"
)

type txKey struct{}

// general abstraction
type sqlExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type sqlTransactor struct {
	db *sql.DB
}

func NewTransactor(db *sql.DB) domain.Transactor {
	return &sqlTransactor{db: db}
}

func (t *sqlTransactor) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("transactor.WithinTransaction.BeginTx: %w", err)
	}

	defer tx.Rollback()

	txCtx := context.WithValue(ctx, txKey{}, tx)
	if err := fn(txCtx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("transactor.WithinTransaction.Commit: %w", err)
	}

	return nil
}

// If there is a `tx` in the context (called from within `WithinTransaction`), use the `tx`.
// If not (called normally outside of a transaction), it falls back to the standard database.
func getExecutor(ctx context.Context, db *sql.DB) sqlExecutor {
	if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok {
		return tx
	}

	return db
}
