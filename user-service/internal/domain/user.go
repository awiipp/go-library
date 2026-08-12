package domain

import (
	"context"
	"time"
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
