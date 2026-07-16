package model

import "time"

type User struct {
	ID        string    `gorm:"column:id,primaryKey"`
	Email     string    `gorm:"column:email"`
	Username  string    `gorm:"column:username"`
	Password  string    `gorm:"column:password"`
	FullName  string    `gorm:"column:full_name"`
	Role      string    `gorm:"column:role"`
	IsActive  bool      `gorm:"column:is_active"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (User) TableName() string {
	return "users"
}
