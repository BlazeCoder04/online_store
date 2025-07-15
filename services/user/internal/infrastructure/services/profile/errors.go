package services

import "errors"

var (
	ErrUserNotFound  = errors.New("user.not_found")
	ErrPasswordWrong = errors.New("password.wrong")
)
