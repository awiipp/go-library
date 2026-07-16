package domain

import (
	"context"
	"time"

	"github.com/awiipp/go-library/user-service/internal/dto"
)

type Role string

const (
	RoleReader Role = "reader"
	RoleAdmin  Role = "admin"
)

type User struct {
	ID        string
	Email     string
	Username  string
	Password  string
	FullName  string
	Role      Role
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
}

type UserUsecase interface {
	Register(ctx context.Context, req *dto.RegisterUserRequest) (*dto.UserResponse, error)
	Login(ctx context.Context, req *dto.LoginUserRequest) (User, error)
}
