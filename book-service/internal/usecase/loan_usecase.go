package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/awiipp/go-library/internal/domain"
	"github.com/awiipp/go-library/internal/dto"
	pkgerrors "github.com/awiipp/go-library/pkg/errors"
)

const loanDuration = 7 * 24 * time.Hour // 7 days

type LoanUsecase interface {
	BorrowBook(ctx context.Context, bookID, userID string) (*dto.LoanResponse, error)
	ReturnBook(ctx context.Context, loanID, userID string) error
	GetMyLoans(ctx context.Context, userID string) ([]*dto.LoanResponse, error)
}

type loanUsecase struct {
	db       *sql.DB
	bookRepo domain.BookRepository
	loanRepo domain.LoanRepository
}

func NewLoanUsecase(db *sql.DB, bookRepo domain.BookRepository, loanRepo domain.LoanRepository) LoanUsecase {
	return &loanUsecase{
		db:       db,
		bookRepo: bookRepo,
		loanRepo: loanRepo,
	}
}

func (u *loanUsecase) BorrowBook(ctx context.Context, bookID, userID string) (*dto.LoanResponse, error) {
	// check: book is already borrowed by user
	exists, err := u.loanRepo.FindActiveByBookAndUser(ctx, bookID, userID)
	if err != nil && errors.Is(err, pkgerrors.ErrNotFound) {
		return nil, fmt.Errorf("usecase.BorrowBook.FindActiveByBookAndUser: %w", err)
	}
	if exists != nil {
		return nil, pkgerrors.ErrAlreadyBorrowed
	}

	// begin transaction
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("usecase.BorrowBook.BeginTx: %w", err)
	}

	defer tx.Rollback()

	if err := u.bookRepo.DecreaseStockTx(ctx, tx, bookID); err != nil {
		return nil, err // ErrOutOfStock
	}

	now := time.Now()
	loan := &domain.Loan{
		BookID:     bookID,
		UserID:     userID,
		Status:     domain.LoanStatusActive,
		BorrowedAt: now,
		DueAt:      now.Add(loanDuration),
	}

	if err := u.loanRepo.Create(ctx, tx, loan); err != nil {
		return nil, fmt.Errorf("usecase.BorrowBook.Create: %w", err)
	}

	// commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("usecase.BorrowBook.Commit: %w", err)
	}

	return toLoanResponse(loan), nil
}

func (u *loanUsecase) ReturnBook(ctx context.Context, loanID, userID string) error {
	// check: loan exists and is active
	loan, err := u.loanRepo.FindByID(ctx, loanID)
	if err != nil {
		return fmt.Errorf("usecase.ReturnBook.FindByID: %w", err)
	}
	if loan.UserID != userID {
		return pkgerrors.ErrForbidden
	}
	if loan.Status != domain.LoanStatusActive {
		return pkgerrors.ErrLoanAlreadyReturned
	}

	// begin transaction
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("usecase.ReturnBook.BeginTx: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()
	// mark loan as returned
	if err := u.loanRepo.MarkReturned(ctx, tx, loanID, now); err != nil {
		return fmt.Errorf("usecase.ReturnBook.MarkReturned: %w", err)
	}
	// increase book stock
	if err := u.bookRepo.IncreaseStockTx(ctx, tx, loan.BookID); err != nil {
		return fmt.Errorf("usecase.ReturnBook.IncreaseStockTx: %w", err)
	}

	// commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("usecase.ReturnBook.Commit: %w", err)
	}

	return nil
}

func (u *loanUsecase) GetMyLoans(ctx context.Context, userID string) ([]*dto.LoanResponse, error) {
	loans, err := u.loanRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("usecase.GetMyLoans.FindByUserID: %w", err)
	}

	responses := make([]*dto.LoanResponse, 0, len(loans))
	for _, l := range loans {
		responses = append(responses, toLoanResponse(l))
	}

	return responses, nil
}
