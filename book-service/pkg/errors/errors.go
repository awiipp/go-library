package pkgerrors

import "errors"

var (
	ErrNotFound            = errors.New("not found")
	ErrConflict            = errors.New("already exists")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrBadRequest          = errors.New("bad request")
	ErrInternalServer      = errors.New("internal server")
	ErrForbidden           = errors.New("you don't have access to this resource")
	ErrOutOfStock          = errors.New("book is out of stock")
	ErrAlreadyBorrowed     = errors.New("you already have an active loan for this book")
	ErrLoanAlreadyReturned = errors.New("loan has already been returned")
)
