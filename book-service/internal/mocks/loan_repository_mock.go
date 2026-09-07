package mocks

import (
	"context"
	"time"

	"github.com/awiipp/go-library/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockLoanRepository struct {
	mock.Mock
}

func (m *MockLoanRepository) Create(ctx context.Context, loan *domain.Loan) error {
	args := m.Called(ctx, loan)
	return args.Error(0)
}

func (m *MockLoanRepository) FindByID(ctx context.Context, id string) (*domain.Loan, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*domain.Loan), args.Error(1)
}

func (m *MockLoanRepository) FindActiveByBookAndUser(ctx context.Context, bookID, userID string) (*domain.Loan, error) {
	args := m.Called(ctx, bookID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Loan), args.Error(1)
}

func (m *MockLoanRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.Loan, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Loan), args.Error(1)
}

func (m *MockLoanRepository) MarkReturned(ctx context.Context, id string, returnedAt time.Time) error {
	args := m.Called(ctx, id, returnedAt)
	return args.Error(0)
}
