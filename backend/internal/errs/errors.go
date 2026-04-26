package errs

import "errors"

var (
	ErrUserLogin             = errors.New("invalid email or password")
	ErrUserEmailAlreadyExist = errors.New("email already exists")
	ErrUserInactive          = errors.New("user is inactive")
	ErrUserInvalidToken      = errors.New("invalid token")
)
