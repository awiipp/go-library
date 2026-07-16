package pkgerrors

import "errors"

var (
	ErrNotFound              = errors.New("not found")
	ErrConflict              = errors.New("already exists")
	ErrUnauthorized          = errors.New("unauthorized")
	ErrBadRequest            = errors.New("bad request")
	ErrInternalServer        = errors.New("internal server")
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrUsernameAlreadyExists = errors.New("username already exists")
)
