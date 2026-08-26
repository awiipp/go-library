package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/awiipp/go-library/user-service/internal/domain"
	"github.com/awiipp/go-library/user-service/internal/repository/model"
	pkgerrors "github.com/awiipp/go-library/user-service/pkg/errors"
	"gorm.io/gorm"
)

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) domain.RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(ctx context.Context, rt *domain.RefreshToken) error {
	m := toRefreshTokenModel(rt)

	if result := r.db.WithContext(ctx).Create(m); result.Error != nil {
		return fmt.Errorf("repository.Create: %w", result.Error)
	}

	*rt = *toRefreshTokenDomain(m)
	return nil
}

func (r *refreshTokenRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	m := &model.RefreshToken{}

	err := r.db.WithContext(ctx).Where("token_hash = ?", tokenHash).First(m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, pkgerrors.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository.FindByTokenHash: %w", err)
	}

	return toRefreshTokenDomain(m), nil
}

func (r *refreshTokenRepository) Revoke(ctx context.Context, tokenHash string) error {
	now := time.Now()

	result := r.db.WithContext(ctx).
		Model(&model.RefreshToken{}).
		Where("token_hash = ?", tokenHash).
		Update("revoked_at", now)
	if result.Error != nil {
		return fmt.Errorf("repository.Revoke: %w", result.Error)
	}

	return nil
}

func (r *refreshTokenRepository) RevokeAllByUserID(ctx context.Context, id string) error {
	now := time.Now()

	result := r.db.WithContext(ctx).
		Model(&model.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", id).
		Update("revoked_at", now)
	if result.Error != nil {
		return fmt.Errorf("repository.RevokeAllByUserID: %w", result.Error)
	}

	return nil
}
