package domain

import (
	"context"

	"github.com/BlazeCoder04/online_store/services/user/internal/domain/models"
)

type UserRepository interface {
	Create(ctx context.Context, email, password, firstName, lastName string) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context, ID string) (*models.User, error)
}
