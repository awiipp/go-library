package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/awiipp/go-library/user-service/internal/domain"
	"github.com/awiipp/go-library/user-service/internal/repository/model"
	pkgerrors "github.com/awiipp/go-library/user-service/pkg/errors"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	m := toModel(user)
	if result := r.db.WithContext(ctx).Create(m); result.Error != nil {
		return fmt.Errorf("repository.Create: %w", result.Error)
	}

	return nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	m := &model.User{}

	err := r.db.WithContext(ctx).Where("email = ?", email).First(m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, pkgerrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return toDomain(m), nil
}

func (r *userRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	m := &model.User{}

	err := r.db.WithContext(ctx).Where("username = ?", username).First(m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, pkgerrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return toDomain(m), nil
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	m := &model.User{}

	err := r.db.WithContext(ctx).Where("id = ?", id).First(m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, pkgerrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return toDomain(m), nil
}
