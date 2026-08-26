package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/awiipp/go-library/user-service/internal/domain"
	"github.com/awiipp/go-library/user-service/internal/repository/model"
	pkgerrors "github.com/awiipp/go-library/user-service/pkg/errors"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	m := toUserModel(user)

	if result := r.db.WithContext(ctx).Create(m); result.Error != nil {
		var pgErr *pgconn.PgError
		if errors.As(result.Error, &pgErr) && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "users_email_key":
				return pkgerrors.ErrEmailAlreadyExists
			case "users_username_key":
				return pkgerrors.ErrUsernameAlreadyExists
			}
			return pkgerrors.ErrConflict
		}

		return fmt.Errorf("repository.Create: %w", result.Error)
	}

	*user = *toUserDomain(m)

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

	return toUserDomain(m), nil
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

	return toUserDomain(m), nil
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

	return toUserDomain(m), nil
}
