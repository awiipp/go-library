package repository

import (
	"github.com/awiipp/go-library/user-service/internal/domain"
	"github.com/awiipp/go-library/user-service/internal/repository/model"
)

func toModel(u *domain.User) *model.User {
	return &model.User{
		ID:        u.ID,
		Email:     u.Email,
		Username:  u.Username,
		Password:  u.Password,
		FullName:  u.FullName,
		Role:      string(u.Role),
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func toDomain(u *model.User) *domain.User {
	return &domain.User{
		ID:        u.ID,
		Email:     u.Email,
		Username:  u.Username,
		Password:  u.Password,
		FullName:  u.FullName,
		Role:      domain.Role(u.Role),
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
