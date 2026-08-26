package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/awiipp/go-library/user-service/internal/dto"
	"github.com/awiipp/go-library/user-service/internal/usecase"
	pkgerrors "github.com/awiipp/go-library/user-service/pkg/errors"
	"github.com/awiipp/go-library/user-service/pkg/response"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

const readTimeout = 5 * time.Second

type UserHandler struct {
	usecase  usecase.UserUsecase
	validate *validator.Validate
}

func NewUserHandler(usecase usecase.UserUsecase) *UserHandler {
	return &UserHandler{
		usecase:  usecase,
		validate: validator.New(),
	}
}

func (u *UserHandler) Register(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), readTimeout)
	defer cancel()

	req := &dto.RegisterUserRequest{}

	if err := c.BodyParser(req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body")
	}

	if err := u.validate.Struct(req); err != nil {
		return response.Error(c, http.StatusBadRequest, err.Error())
	}

	result, err := u.usecase.Register(ctx, req)
	if err != nil {
		switch {
		case errors.Is(err, pkgerrors.ErrEmailAlreadyExists):
			return response.Error(c, http.StatusConflict, "email already exist")
		case errors.Is(err, pkgerrors.ErrUsernameAlreadyExists):
			return response.Error(c, http.StatusConflict, "username already exist")
		default:
			return response.Error(c, http.StatusInternalServerError, "internal server error")
		}
	}

	return response.Success(c, http.StatusCreated, result)
}

func (u *UserHandler) Login(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), readTimeout)
	defer cancel()

	req := &dto.LoginUserRequest{}

	if err := c.BodyParser(req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid body request")
	}
	if err := u.validate.Struct(req); err != nil {
		return response.Error(c, http.StatusBadRequest, err.Error())
	}

	result, err := u.usecase.Login(ctx, req)
	if err != nil {
		switch {
		case errors.Is(err, pkgerrors.ErrInvalidCredentials):
			return response.Error(c, http.StatusUnauthorized, "invalid email or password")
		case errors.Is(err, pkgerrors.ErrUserInactive):
			return response.Error(c, http.StatusForbidden, "user is inactive")
		default:
			return response.Error(c, http.StatusInternalServerError, "internal server error")
		}
	}

	return response.Success(c, http.StatusOK, result)
}

func (u *UserHandler) RefreshToken(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), readTimeout)
	defer cancel()

	req := &dto.RefreshTokenRequest{}

	if err := c.BodyParser(req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid body request")
	}
	if err := u.validate.Struct(req); err != nil {
		return response.Error(c, http.StatusBadRequest, err.Error())
	}

	result, err := u.usecase.RefreshToken(ctx, req)
	if err != nil {
		switch {
		case errors.Is(err, pkgerrors.ErrInvalidRefreshToken):
			return response.Error(c, http.StatusUnauthorized, "invalid or expired refresh token")
		case errors.Is(err, pkgerrors.ErrUserInactive):
			return response.Error(c, http.StatusForbidden, "user is inactive")
		default:
			return response.Error(c, http.StatusInternalServerError, "internal server error")
		}
	}

	return response.Success(c, http.StatusOK, result)
}

func (u *UserHandler) Profile(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), readTimeout)
	defer cancel()

	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return response.Error(c, http.StatusUnauthorized, "unauthorized")
	}

	result, err := u.usecase.GetProfile(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, pkgerrors.ErrNotFound):
			return response.Error(c, http.StatusNotFound, "user not found")
		default:
			return response.Error(c, http.StatusInternalServerError, "internal server error")
		}
	}

	return response.Success(c, http.StatusOK, result)
}
