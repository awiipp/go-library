package usecase

import (
	"context"
	"fmt"

	"github.com/awiipp/go-library/user-service/internal/domain"
	"github.com/awiipp/go-library/user-service/internal/dto"
	pkgerrors "github.com/awiipp/go-library/user-service/pkg/errors"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type userUsecase struct {
	userRepo domain.UserRepository
}

func NewUserUsecase(userRepo domain.UserRepository) domain.UserUsecase {
	return &userUsecase{userRepo: userRepo}
}

func (u *userUsecase) Register(ctx context.Context, req *dto.RegisterUserRequest) (*dto.UserResponse, error) {
	existing, err := u.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("usecase.Register.FindByEmail: %w", err)
	}
	if existing != nil {
		return nil, pkgerrors.ErrEmailAlreadyExists
	}

	existing, err = u.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("usecase.Register.FindByUsername: %w", err)
	}
	if existing != nil {
		return nil, pkgerrors.ErrUsernameAlreadyExists
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("usecase.Register.GenerateFromPassword: %w", err)
	}

	user := &domain.User{
		ID:       uuid.NewString(),
		Email:    req.Email,
		Username: req.Username,
		Password: string(hashed),
		FullName: req.FullName,
		Role:     domain.RoleReader,
		IsActive: true,
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("usecase.Register.Create: %w", err)
	}

	return &dto.UserResponse{
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
		FullName: user.FullName,
		Role:     string(user.Role),
		IsActive: user.IsActive,
	}, nil
}

func (u *userUsecase) Login(ctx context.Context, req *dto.LoginUserRequest) (*domain.User, error) {
	return nil, nil // WIP
}
