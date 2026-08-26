package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/awiipp/go-library/user-service/internal/config"
	"github.com/awiipp/go-library/user-service/internal/domain"
	"github.com/awiipp/go-library/user-service/internal/dto"
	pkgerrors "github.com/awiipp/go-library/user-service/pkg/errors"
	"github.com/awiipp/go-library/user-service/pkg/utils"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserUsecase interface {
	Register(ctx context.Context, req *dto.RegisterUserRequest) (*dto.UserResponse, error)
	Login(ctx context.Context, req *dto.LoginUserRequest) (*dto.LoginResponse, error)
	RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.LoginResponse, error)
	GetProfile(ctx context.Context, userID string) (*dto.UserResponse, error)
}

type userUsecase struct {
	userRepo         domain.UserRepository
	refreshTokenRepo domain.RefreshTokenRepository
	cfg              *config.Config
}

func NewUserUsecase(userRepo domain.UserRepository, refreshTokenRepo domain.RefreshTokenRepository, config *config.Config) UserUsecase {
	return &userUsecase{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		cfg:              config,
	}
}

func (u *userUsecase) Register(ctx context.Context, req *dto.RegisterUserRequest) (*dto.UserResponse, error) {
	existing, err := u.userRepo.FindByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, pkgerrors.ErrNotFound) {
		return nil, fmt.Errorf("usecase.Register.FindByEmail: %w", err)
	}
	if existing != nil {
		return nil, pkgerrors.ErrEmailAlreadyExists
	}

	existing, err = u.userRepo.FindByUsername(ctx, req.Username)
	if err != nil && !errors.Is(err, pkgerrors.ErrNotFound) {
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

	return toUserResponse(user), nil
}

func (u *userUsecase) Login(ctx context.Context, req *dto.LoginUserRequest) (*dto.LoginResponse, error) {
	// user check
	user, err := u.userRepo.FindByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, pkgerrors.ErrNotFound) {
		return nil, fmt.Errorf("usecase.Login.FindByEmail: %w", err)
	}

	if user == nil {
		return nil, pkgerrors.ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, pkgerrors.ErrUserInactive
	}

	// password check
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, pkgerrors.ErrInvalidCredentials
	}

	// access token
	accessToken, err := utils.GenerateToken(u.cfg, user.ID, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("usecase.Login.GenerateToken: %w", err)
	}

	// refresh token
	rawRefreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("usecase.Login.GenerateRefreshToken: %w", err)
	}

	rt := &domain.RefreshToken{
		ID:        uuid.NewString(),
		UserID:    user.ID,
		TokenHash: utils.HashRefreshToken(rawRefreshToken),
		ExpiresAt: time.Now().Add(time.Duration(u.cfg.Auth.RefreshTokenTTL) * 24 * time.Hour),
	}

	if err := u.refreshTokenRepo.Create(ctx, rt); err != nil {
		return nil, fmt.Errorf("usecase.Login.Create: %w", err)
	}

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    u.cfg.JWT.ExpiresIn,
	}, nil
}

func (u *userUsecase) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.LoginResponse, error) {
	tokenHash := utils.HashRefreshToken(req.RefreshToken)

	rt, err := u.refreshTokenRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil && !errors.Is(err, pkgerrors.ErrNotFound) {
		return nil, fmt.Errorf("usecase.RefreshToken.FindByTokenHash: %w", err)
	}
	if rt == nil {
		return nil, pkgerrors.ErrInvalidRefreshToken
	}

	// reuse detection
	// if an already-revoked token is used again, it may indicate the token was stolen.
	if rt.RevokedAt != nil {
		_ = u.refreshTokenRepo.RevokeAllByUserID(ctx, rt.UserID)
		return nil, pkgerrors.ErrInvalidRefreshToken
	}

	if time.Now().After(rt.ExpiresAt) {
		return nil, pkgerrors.ErrInvalidRefreshToken
	}

	user, err := u.userRepo.FindByID(ctx, rt.UserID)
	if err != nil {
		return nil, fmt.Errorf("usecase.RefreshToken.FindByID: %w", err)
	}
	if !user.IsActive {
		return nil, pkgerrors.ErrUserInactive
	}

	// revoke old token
	if err := u.refreshTokenRepo.Revoke(ctx, tokenHash); err != nil {
		return nil, fmt.Errorf("usecase.RefreshToken.Revoke: %w", err)
	}

	// issue access token and new refresh token
	accessToken, err := utils.GenerateToken(u.cfg, user.ID, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("usecase.RefreshToken.GenerateToken: %w", err)
	}

	rawRefreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("usecase.RefreshToken.GenerateRefreshToken: %w", err)
	}

	newRT := &domain.RefreshToken{
		ID:        uuid.NewString(),
		UserID:    user.ID,
		TokenHash: utils.HashRefreshToken(rawRefreshToken),
		ExpiresAt: time.Now().Add(time.Duration(u.cfg.Auth.RefreshTokenTTL) * 24 * time.Hour),
	}
	if err := u.refreshTokenRepo.Create(ctx, newRT); err != nil {
		return nil, fmt.Errorf("usecase.RefreshToken.Create: %w", err)
	}

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    u.cfg.JWT.ExpiresIn,
	}, nil
}

func (u *userUsecase) GetProfile(ctx context.Context, userID string) (*dto.UserResponse, error) {
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("usecase.GetProfile: %w", err)
	}

	return toUserResponse(user), nil
}
