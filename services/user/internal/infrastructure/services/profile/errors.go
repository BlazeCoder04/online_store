package services

import "errors"

var (
	ErrUserNotFound       = errors.New("user.not_found")
	ErrPasswordWrong      = errors.New("password.wrong")
	ErrEmailUnchanged     = errors.New("email.unchanged")
	ErrPasswordUnchanged  = errors.New("password.unchanged")
	ErrFirstNameUnchanged = errors.New("first_name.unchanged")
	ErrLastNameUnchanged  = errors.New("last_name.unchanged")
)
