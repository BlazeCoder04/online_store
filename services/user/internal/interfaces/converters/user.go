package converters

import (
	"github.com/BlazeCoder04/online_store/services/user/internal/domain/models"
	desc "github.com/BlazeCoder04/online_store/services/user/pkg/user"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func UserToDesc(user *models.User) *desc.User {
	return &desc.User{
		Id:        user.ID.String(),
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      desc.UserRole(desc.UserRole_value[string(user.Role)]),
		CreatedAt: timestamppb.New(user.CreatedAt.UTC()),
		UpdatedAt: timestamppb.New(user.UpdatedAt.UTC()),
	}
}
