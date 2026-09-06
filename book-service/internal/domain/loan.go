package domain

import (
	"context"
	"time"
)

type LoanStatus string

const (
	LoanStatusActive   LoanStatus = "active"
	LoanStatusReturned LoanStatus = "returned"
)

type Loan struct {
	ID         string
	BookID     string
	UserID     string
	Status     LoanStatus
	BorrowedAt time.Time
	DueAt      time.Time
	ReturnedAt *time.Time
}

type LoanRepository interface {
	Create(ctx context.Context, loan *Loan) error
	FindByID(ctx context.Context, id string) (*Loan, error)
	FindActiveByBookAndUser(ctx context.Context, bookID, userID string) (*Loan, error)
	FindByUserID(ctx context.Context, userID string) ([]*Loan, error)
	MarkReturned(ctx context.Context, id string, returnedAt time.Time) error
}
