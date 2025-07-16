package domain

import (
	"context"

	"github.com/BlazeCoder04/online_store/services/user/internal/domain/models"
)

type UpdateProfileArgs struct {
	UserID   string
	Password string

	NewEmail     *string
	NewPassword  *string
	NewFirstName *string
	NewLastName  *string

	AccessToken string
}

type ProfileService interface {
	Get(ctx context.Context, userID, accessToken string) (*models.User, error)
	Update(ctx context.Context, args *UpdateProfileArgs) (*models.User, error)
	Delete(ctx context.Context, userID, password, accessToken string) error
}
