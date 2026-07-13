package domain

import "time"

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
