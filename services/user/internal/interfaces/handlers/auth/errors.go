package handlers

import "errors"

var (
	ErrUserNotFound  = errors.New("user.not_found")
	ErrPasswordWrong = errors.New("password.wrong")
	ErrUserExists    = errors.New("user.exists")
	ErrTokenInvalid  = errors.New("token.invalid")
)
