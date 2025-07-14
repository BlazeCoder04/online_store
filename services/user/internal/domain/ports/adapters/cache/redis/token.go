package domain

import (
	"context"
	"time"
)

type TokenAdapter interface {
	Set(ctx context.Context, userID, refreshToken string, expiresIn time.Duration) error
	Get(ctx context.Context, userID string) (string, error)
	Del(ctx context.Context, userID string) error
}
