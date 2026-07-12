package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/awiipp/go-library/internal/domain"
	"github.com/awiipp/go-library/internal/dto"
	pkgerrors "github.com/awiipp/go-library/pkg/errors"
	"github.com/awiipp/go-library/pkg/response"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type BookHandler struct {
	usecase  domain.BookUsecase
	validate *validator.Validate
}

const (
	readTimeout  = 5 * time.Second
	writeTimeout = 10 * time.Second
)

func NewBookHandler(usecase domain.BookUsecase) *BookHandler {
	return &BookHandler{
		usecase:  usecase,
		validate: validator.New(),
	}
}

func (h *BookHandler) Getall(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), readTimeout)
	defer cancel()

	result, err := h.usecase.GetAll(ctx)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "internal server error")
	}

	return response.Success(c, http.StatusOK, result)
}

func (h *BookHandler) GetByID(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), readTimeout)
	defer cancel()

	id := c.Params("id")

	if _, err := uuid.Parse(id); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid book id")
	}

	result, err := h.usecase.GetByID(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, pkgerrors.ErrNotFound):
			return response.Error(c, http.StatusNotFound, "book not found")
		default:
			return response.Error(c, http.StatusInternalServerError, "internal server error")
		}
	}

	return response.Success(c, http.StatusOK, result)
}

func (h *BookHandler) Create(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), writeTimeout)
	defer cancel()

	req := &dto.CreateBookRequest{}

	if err := c.BodyParser(req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body")
	}

	if err := h.validate.Struct(req); err != nil {
		return response.Error(c, http.StatusBadRequest, err.Error())
	}

	result, err := h.usecase.Create(ctx, req)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "internal server error")
	}

	return response.Success(c, http.StatusCreated, result)
}

func (h *BookHandler) Update(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), writeTimeout)
	defer cancel()

	id := c.Params("id")
	req := &dto.UpdateBookRequest{}

	if _, err := uuid.Parse(id); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid book id")
	}

	if err := c.BodyParser(req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body")
	}

	if err := h.validate.Struct(req); err != nil {
		return response.Error(c, http.StatusBadRequest, err.Error())
	}

	result, err := h.usecase.Update(ctx, id, req)
	if err != nil {
		switch {
		case errors.Is(err, pkgerrors.ErrNotFound):
			return response.Error(c, http.StatusNotFound, "book not found")
		default:
			return response.Error(c, http.StatusInternalServerError, "internal server error")
		}
	}

	return response.Success(c, http.StatusOK, result)
}

func (h *BookHandler) Delete(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), writeTimeout)
	defer cancel()

	id := c.Params("id")

	if _, err := uuid.Parse(id); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid book id")
	}

	if err := h.usecase.Delete(ctx, id); err != nil {
		switch {
		case errors.Is(err, pkgerrors.ErrNotFound):
			return response.Error(c, http.StatusNotFound, "book not found")
		default:
			return response.Error(c, http.StatusInternalServerError, "internal server error")
		}
	}

	return response.Success(c, http.StatusOK, nil)
}
