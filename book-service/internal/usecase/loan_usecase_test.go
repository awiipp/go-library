package usecase_test

import (
	"context"
	"testing"

	"github.com/awiipp/go-library/internal/domain"
	"github.com/awiipp/go-library/internal/mocks"
	"github.com/awiipp/go-library/internal/usecase"
	pkgerrors "github.com/awiipp/go-library/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Borrow Book Tests
func TestBorrowBook_Success(t *testing.T) {
	bookRepo := new(mocks.MockBookRepository)
	loanRepo := new(mocks.MockLoanRepository)
	tx := &mocks.FakeTransactor{}

	loanRepo.On("FindActiveByBookAndUser", mock.Anything, "book-1", "user-1").Return(nil, pkgerrors.ErrNotFound)
	bookRepo.On("DecreaseStock", mock.Anything, "book-1").Return(nil)
	loanRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Loan")).Return(nil)
	bookRepo.On("InvalidateCache", mock.Anything, "book-1").Return(nil)

	uc := usecase.NewLoanUsecase(tx, bookRepo, loanRepo)
	result, err := uc.BorrowBook(context.Background(), "book-1", "user-1")

	assert.NoError(t, err)
	assert.Equal(t, "book-1", result.BookID)
	assert.Equal(t, "user-1", result.UserID)
	assert.Equal(t, "active", result.Status)
	bookRepo.AssertExpectations(t)
	loanRepo.AssertExpectations(t)
}

func TestBorrowBook_AlreadyBorrowed(t *testing.T) {
	bookRepo := new(mocks.MockBookRepository)
	loanRepo := new(mocks.MockLoanRepository)
	tx := &mocks.FakeTransactor{}

	existingLoan := &domain.Loan{
		ID:     "loan-1",
		BookID: "book-1",
		UserID: "user-1",
		Status: domain.LoanStatusActive,
	}

	loanRepo.On("FindActiveByBookAndUser", mock.Anything, "book-1", "user-1").Return(existingLoan, nil)

	uc := usecase.NewLoanUsecase(tx, bookRepo, loanRepo)
	result, err := uc.BorrowBook(context.Background(), "book-1", "user-1")

	assert.Nil(t, result)
	assert.ErrorIs(t, err, pkgerrors.ErrAlreadyBorrowed)
	bookRepo.AssertNotCalled(t, "DecreaseStock", mock.Anything, mock.Anything)
}

func TestBorrowBook_OutOfStock(t *testing.T) {
	bookRepo := new(mocks.MockBookRepository)
	loanRepo := new(mocks.MockLoanRepository)
	tx := &mocks.FakeTransactor{}

	loanRepo.On("FindActiveByBookAndUser", mock.Anything, "book-1", "user-1").Return(nil, pkgerrors.ErrNotFound)
	bookRepo.On("DecreaseStock", mock.Anything, "book-1").Return(pkgerrors.ErrOutOfStock)

	uc := usecase.NewLoanUsecase(tx, bookRepo, loanRepo)
	result, err := uc.BorrowBook(context.Background(), "book-1", "user-1")

	assert.Nil(t, result)
	assert.ErrorIs(t, err, pkgerrors.ErrOutOfStock)
	loanRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

// Return Book Tests
func TestReturnBook_Success(t *testing.T) {
	bookRepo := new(mocks.MockBookRepository)
	loanRepo := new(mocks.MockLoanRepository)
	tx := &mocks.FakeTransactor{}

	existingLoan := &domain.Loan{
		ID:     "loan-1",
		BookID: "book-1",
		UserID: "user-1",
		Status: domain.LoanStatusActive,
	}

	loanRepo.On("FindByID", mock.Anything, "loan-1").Return(existingLoan, nil)
	loanRepo.On("MarkReturned", mock.Anything, "loan-1", mock.AnythingOfType("time.Time")).Return(nil)
	bookRepo.On("IncreaseStock", mock.Anything, "book-1").Return(nil)
	bookRepo.On("InvalidateCache", mock.Anything, "book-1").Return(nil)

	uc := usecase.NewLoanUsecase(tx, bookRepo, loanRepo)
	err := uc.ReturnBook(context.Background(), "loan-1", "user-1")

	assert.NoError(t, err)
	loanRepo.AssertExpectations(t)
	bookRepo.AssertExpectations(t)
}

func TestReturnBook_Forbidden(t *testing.T) {
	bookRepo := new(mocks.MockBookRepository)
	loanRepo := new(mocks.MockLoanRepository)
	tx := &mocks.FakeTransactor{}

	existingLoan := &domain.Loan{
		ID:     "loan-1",
		BookID: "book-1",
		UserID: "someone-else",
		Status: domain.LoanStatusActive,
	}

	loanRepo.On("FindByID", mock.Anything, "loan-1").Return(existingLoan, nil)

	uc := usecase.NewLoanUsecase(tx, bookRepo, loanRepo)
	err := uc.ReturnBook(context.Background(), "loan-1", "user-1") // user-1 is NOT loan owner

	assert.ErrorIs(t, err, pkgerrors.ErrForbidden)
	loanRepo.AssertNotCalled(t, "MarkReturned", mock.Anything, mock.Anything, mock.Anything)
}

func TestReturnBook_AlreadyReturned(t *testing.T) {
	bookRepo := new(mocks.MockBookRepository)
	loanRepo := new(mocks.MockLoanRepository)
	tx := &mocks.FakeTransactor{}

	existingLoan := &domain.Loan{
		ID:     "loan-1",
		BookID: "book-1",
		UserID: "user-1",
		Status: domain.LoanStatusReturned,
	}

	loanRepo.On("FindByID", mock.Anything, "loan-1").Return(existingLoan, nil)

	uc := usecase.NewLoanUsecase(tx, bookRepo, loanRepo)
	err := uc.ReturnBook(context.Background(), "loan-1", "user-1")

	assert.ErrorIs(t, err, pkgerrors.ErrLoanAlreadyReturned)
}
