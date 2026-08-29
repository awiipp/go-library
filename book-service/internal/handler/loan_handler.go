package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/awiipp/go-library/internal/usecase"
	pkgerrors "github.com/awiipp/go-library/pkg/errors"
	"github.com/awiipp/go-library/pkg/response"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type LoanHandler struct {
	usecase usecase.LoanUsecase
}

func NewLoanHandler(usecase usecase.LoanUsecase) *LoanHandler {
	return &LoanHandler{usecase: usecase}
}

func (h *LoanHandler) Borrow(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), readTimeout)
	defer cancel()

	bookID := c.Params("id")
	if _, err := uuid.Parse(bookID); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid book id")
	}

	// get user id
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return response.Error(c, http.StatusUnauthorized, "unauthorized")
	}

	result, err := h.usecase.BorrowBook(ctx, bookID, userID)
	if err != nil {
		switch {
		case errors.Is(err, pkgerrors.ErrNotFound):
			return response.Error(c, http.StatusNotFound, "book not found")
		case errors.Is(err, pkgerrors.ErrOutOfStock):
			return response.Error(c, http.StatusConflict, "book is out of stock")
		case errors.Is(err, pkgerrors.ErrAlreadyBorrowed):
			return response.Error(c, http.StatusConflict, "you already borrowed this book")
		default:
			return response.Error(c, http.StatusInternalServerError, "internal server error")
		}
	}

	return response.Success(c, http.StatusCreated, result)
}

func (h *LoanHandler) Return(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), readTimeout)
	defer cancel()

	loanID := c.Params("id")
	if _, err := uuid.Parse(loanID); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid loan id")
	}

	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return response.Error(c, http.StatusUnauthorized, "unauthorized")
	}

	if err := h.usecase.ReturnBook(ctx, loanID, userID); err != nil {
		switch {
		case errors.Is(err, pkgerrors.ErrNotFound):
			return response.Error(c, http.StatusNotFound, "loan not found")
		case errors.Is(err, pkgerrors.ErrForbidden):
			return response.Error(c, http.StatusForbidden, "not your loan")
		case errors.Is(err, pkgerrors.ErrLoanAlreadyReturned):
			return response.Error(c, http.StatusConflict, "loan already returned")
		default:
			return response.Error(c, http.StatusInternalServerError, "internal server error")
		}
	}

	return response.Success(c, http.StatusOK, fiber.Map{"message": "book returned successfully"})
}

func (h *LoanHandler) MyLoan(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), readTimeout)
	defer cancel()

	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return response.Error(c, http.StatusUnauthorized, "unauthorized")
	}

	result, err := h.usecase.GetMyLoans(ctx, userID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "internal server error")
	}

	return response.Success(c, http.StatusOK, result)
}
