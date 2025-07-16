package handlers

import "errors"

var (
	ErrUserNotFound        = errors.New("user.not_found")
	ErrPasswordWrong       = errors.New("password.wrong")
	ErrMetadataNotProvided = errors.New("metadata.not_provided")
	ErrHeaderNotProvided   = errors.New("header.not_provided")
	ErrTokenInvalid        = errors.New("token.invalid")
)
