package domain

import (
	"context"

	"github.com/BlazeCoder04/online_store/services/user/internal/domain/models"
)

type AuthService interface {
	Login(ctx context.Context, email, password string) (*models.User, string, string, error)
	Register(ctx context.Context, email, password, firstName, lastName string) (*models.User, string, string, error)
	UpdateTokens(ctx context.Context, refreshToken string) (string, string, error)
	Logout(ctx context.Context, accessToken string) error
}
