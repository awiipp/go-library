package usecase

import (
	"github.com/awiipp/go-library/internal/domain"
	"github.com/awiipp/go-library/internal/dto"
)

func toBookResponse(book *domain.Book) *dto.BookResponse {
	return &dto.BookResponse{
		ID:          book.ID,
		Title:       book.Title,
		Author:      book.Author,
		Description: book.Description,
		CreatedAt:   book.CreatedAt,
		UpdatedAt:   book.UpdatedAt,
	}
}
