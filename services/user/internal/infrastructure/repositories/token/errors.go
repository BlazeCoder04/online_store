package repositories

import "errors"

var (
	ErrConnecting    = "error connecting to the redis"
	ErrTokenNotFound = errors.New("token.not_found")
)
