package dto

import "time"

type LoanResponse struct {
	ID         string     `json:"id"`
	BookID     string     `json:"book_id"`
	UserID     string     `json:"user_id"`
	Status     string     `json:"status"`
	BorrowedAt time.Time  `json:"borrowed_at"`
	DueAt      time.Time  `json:"due_at"`
	ReturnedAt *time.Time `json:"returned_at,omitempty"`
}
