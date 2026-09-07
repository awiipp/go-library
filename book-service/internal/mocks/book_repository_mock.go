package mocks

import (
	"context"

	"github.com/awiipp/go-library/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockBookRepository struct {
	domain.BookRepository // embed interface
	mock.Mock
}

func (m *MockBookRepository) DecreaseStock(ctx context.Context, bookID string) error {
	args := m.Called(ctx, bookID)
	return args.Error(0)
}

func (m *MockBookRepository) IncreaseStock(ctx context.Context, bookID string) error {
	args := m.Called(ctx, bookID)
	return args.Error(0)
}

func (m *MockBookRepository) InvalidateCache(ctx context.Context, bookID string) error {
	args := m.Called(ctx, bookID)
	return args.Error(0)
}
