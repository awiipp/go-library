package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
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
	tx       domain.Transactor
	bookRepo domain.BookRepository
	loanRepo domain.LoanRepository
}

func NewLoanUsecase(tx domain.Transactor, bookRepo domain.BookRepository, loanRepo domain.LoanRepository) LoanUsecase {
	return &loanUsecase{
		tx:       tx,
		bookRepo: bookRepo,
		loanRepo: loanRepo,
	}
}

func (u *loanUsecase) BorrowBook(ctx context.Context, bookID, userID string) (*dto.LoanResponse, error) {
	// check: book is already borrowed by user
	exists, err := u.loanRepo.FindActiveByBookAndUser(ctx, bookID, userID)
	if err != nil && !errors.Is(err, pkgerrors.ErrNotFound) {
		return nil, fmt.Errorf("usecase.BorrowBook.FindActiveByBookAndUser: %w", err)
	}
	if exists != nil {
		return nil, pkgerrors.ErrAlreadyBorrowed
	}

	loan := &domain.Loan{}
	err = u.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := u.bookRepo.DecreaseStock(txCtx, bookID); err != nil {
			return err
		}

		now := time.Now()
		loan = &domain.Loan{
			BookID:     bookID,
			UserID:     userID,
			Status:     domain.LoanStatusActive,
			BorrowedAt: now,
			DueAt:      now.Add(loanDuration),
		}

		return u.loanRepo.Create(txCtx, loan)
	})
	if err != nil {
		return nil, fmt.Errorf("usecase.BorrowBook: %w", err)
	}

	if err := u.bookRepo.InvalidateCache(ctx, bookID); err != nil {
		log.Printf("failed to invalidate book cache %s: %v", bookID, err)
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

	err = u.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		now := time.Now()
		// mark loan as returned
		if err := u.loanRepo.MarkReturned(txCtx, loanID, now); err != nil {
			return fmt.Errorf("usecase.ReturnBook.MarkReturned: %w", err)
		}

		// increase book stock
		return u.bookRepo.IncreaseStock(txCtx, loan.BookID)
	})
	if err != nil {
		return fmt.Errorf("usecase.ReturnBook: %w", err)
	}

	if err := u.bookRepo.InvalidateCache(ctx, loan.BookID); err != nil {
		log.Printf("failed to invalidate book cache %s: %v", loan.BookID, err)
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
