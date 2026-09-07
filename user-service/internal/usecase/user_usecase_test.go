package usecase_test

import (
	"context"
	"testing"

	"github.com/awiipp/go-library/user-service/internal/domain"
	"github.com/awiipp/go-library/user-service/internal/dto"
	"github.com/awiipp/go-library/user-service/internal/mocks"
	"github.com/awiipp/go-library/user-service/internal/usecase"
	pkgerrors "github.com/awiipp/go-library/user-service/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// Register Tests
func TestRegister_Success(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	refreshRepo := new(mocks.MockRefreshTokenRepository)
	cfg := newTestConfig(t)

	userRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(nil, pkgerrors.ErrNotFound)
	userRepo.On("FindByUsername", mock.Anything, "testuser").Return(nil, pkgerrors.ErrNotFound)
	userRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)

	uc := usecase.NewUserUsecase(userRepo, refreshRepo, cfg)

	req := &dto.RegisterUserRequest{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "password",
		FullName: "Test User",
	}

	result, err := uc.Register(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, "test@example.com", result.Email)
	assert.Equal(t, "reader", result.Role)
	assert.True(t, result.IsActive)
	userRepo.AssertExpectations(t)
}

func TestRegister_EmailAlreadyExists(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	refreshRepo := new(mocks.MockRefreshTokenRepository)
	cfg := newTestConfig(t)

	existingUser := &domain.User{ID: "user-id-example", Email: "test@example.com"}
	userRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(existingUser, nil)

	uc := usecase.NewUserUsecase(userRepo, refreshRepo, cfg)

	req := &dto.RegisterUserRequest{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "password",
		FullName: "Test User",
	}

	result, err := uc.Register(context.Background(), req)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, pkgerrors.ErrEmailAlreadyExists)
	userRepo.AssertExpectations(t)
	userRepo.AssertNotCalled(t, "FindByUsername", mock.Anything, mock.Anything)
}

func TestRegister_UsernameAlreadyExists(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	refreshRepo := new(mocks.MockRefreshTokenRepository)
	cfg := newTestConfig(t)

	userRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(nil, pkgerrors.ErrNotFound)
	existingUser := &domain.User{ID: "user-id-example", Username: "testuser"}
	userRepo.On("FindByUsername", mock.Anything, "testuser").Return(existingUser, nil)

	uc := usecase.NewUserUsecase(userRepo, refreshRepo, cfg)

	req := &dto.RegisterUserRequest{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "password",
		FullName: "Test User",
	}

	result, err := uc.Register(context.Background(), req)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, pkgerrors.ErrUsernameAlreadyExists)
	userRepo.AssertExpectations(t)
}

// Login Tests
func TestLogin_Success(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	refreshRepo := new(mocks.MockRefreshTokenRepository)
	cfg := newTestConfig(t)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	existingUser := &domain.User{
		ID:       "user-id-example",
		Email:    "test@example.com",
		Password: string(hashed),
		Role:     domain.RoleReader,
		IsActive: true,
	}

	userRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(existingUser, nil)
	refreshRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.RefreshToken")).Return(nil)

	uc := usecase.NewUserUsecase(userRepo, refreshRepo, cfg)

	req := &dto.LoginUserRequest{
		Email:    "test@example.com",
		Password: "password",
	}

	result, err := uc.Login(context.Background(), req)

	assert.NoError(t, err)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
	assert.Equal(t, "Bearer", result.TokenType)
	userRepo.AssertExpectations(t)
	refreshRepo.AssertExpectations(t)
}

func TestLogin_WrongPassword(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	refreshRepo := new(mocks.MockRefreshTokenRepository)
	cfg := newTestConfig(t)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	existingUser := &domain.User{
		ID:       "user-id-example",
		Email:    "test@example.com",
		Password: string(hashed),
		IsActive: true,
	}

	userRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(existingUser, nil)

	uc := usecase.NewUserUsecase(userRepo, refreshRepo, cfg)

	req := &dto.LoginUserRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	result, err := uc.Login(context.Background(), req)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, pkgerrors.ErrInvalidCredentials)
}

func TestLogin_UserNotFound(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	refreshRepo := new(mocks.MockRefreshTokenRepository)
	cfg := newTestConfig(t)

	userRepo.On("FindByEmail", mock.Anything, "notfound@example.com").Return(nil, pkgerrors.ErrNotFound)

	uc := usecase.NewUserUsecase(userRepo, refreshRepo, cfg)

	req := &dto.LoginUserRequest{
		Email:    "notfound@example.com",
		Password: "password",
	}

	result, err := uc.Login(context.Background(), req)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, pkgerrors.ErrInvalidCredentials)
}

func TestLogin_InactiveUser(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	refreshRepo := new(mocks.MockRefreshTokenRepository)
	cfg := newTestConfig(t)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	existingUser := &domain.User{
		ID:       "user-id",
		Email:    "test@example.com",
		Password: string(hashed),
		IsActive: false,
	}

	userRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(existingUser, nil)

	uc := usecase.NewUserUsecase(userRepo, refreshRepo, cfg)

	req := &dto.LoginUserRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	result, err := uc.Login(context.Background(), req)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, pkgerrors.ErrUserInactive)
}
