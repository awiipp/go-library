package usecase

import (
	"context"
	"fmt"

	"github.com/awiipp/go-library/internal/domain"
	"github.com/awiipp/go-library/internal/dto"
)

type bookUsecase struct {
	bookRepo domain.BookRepository
}

func NewBookUsecase(bookRepo domain.BookRepository) domain.BookUsecase {
	return &bookUsecase{bookRepo: bookRepo}
}

func (u *bookUsecase) GetAll(ctx context.Context) ([]*dto.BookResponse, error) {
	books, err := u.bookRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("usecase.GetAll: %w", err)
	}

	var result []*dto.BookResponse
	for _, book := range books {
		result = append(result, toBookResponse(book))
	}

	return result, nil
}

func (u *bookUsecase) GetByID(ctx context.Context, id string) (*dto.BookResponse, error) {
	book, err := u.bookRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("usecase.GetByID: %w", err)
	}

	return toBookResponse(book), nil
}

func (u *bookUsecase) Create(ctx context.Context, req *dto.CreateBookRequest) (*dto.BookResponse, error) {
	book := &domain.Book{
		Title:       req.Title,
		Author:      req.Author,
		Description: req.Description,
	}

	saved, err := u.bookRepo.Save(ctx, book)
	if err != nil {
		return nil, fmt.Errorf("usecase.Create: %w", err)
	}

	return toBookResponse(saved), nil
}

func (u *bookUsecase) Update(ctx context.Context, id string, req *dto.UpdateBookRequest) (*dto.BookResponse, error) {
	book, err := u.bookRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	book.Title = req.Title
	book.Author = req.Author
	book.Description = req.Description

	updated, err := u.bookRepo.Update(ctx, book)
	if err != nil {
		return nil, fmt.Errorf("usecase.Update: %w", err)
	}

	return toBookResponse(updated), err
}

func (u *bookUsecase) Delete(ctx context.Context, id string) error {
	_, err := u.bookRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if err := u.bookRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("usecase.Delete: %w", err)
	}

	return nil
}
