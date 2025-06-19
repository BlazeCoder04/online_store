package domain

import (
	"context"
	"time"
)

type TokenRepository interface {
	Set(ctx context.Context, userID, token string, expiresIn time.Duration) error
	Get(ctx context.Context, userID string) (string, error)
	Delete(ctx context.Context, userID string) error
}
