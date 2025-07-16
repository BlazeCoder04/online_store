package tests

import (
	"context"
	"testing"

	"github.com/BlazeCoder04/online_store/libs/hash"
	"github.com/BlazeCoder04/online_store/libs/logger"
	"github.com/BlazeCoder04/online_store/services/user/internal/domain/models"
	mocksRepo "github.com/BlazeCoder04/online_store/services/user/internal/domain/ports/repositories/mocks"
	services "github.com/BlazeCoder04/online_store/services/user/internal/infrastructure/services/profile"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func TestProfileService_Delete(t *testing.T) {
	type args struct {
		ctx      context.Context
		userID   string
		password string
	}

	type expect struct {
		err error
	}

	var (
		ctx = context.Background()

		userID            = uuid.New()
		email             = gofakeit.Email()
		password          = "password"
		wrongPassword     = "wrong_password"
		hashedPassword, _ = hash.HashPassword(password)
		firstName         = gofakeit.FirstName()
		lastName          = gofakeit.LastName()
		role              = models.UserRole

		baseUser = &models.User{
			ID:        userID,
			Email:     email,
			Password:  hashedPassword,
			FirstName: firstName,
			LastName:  lastName,
			Role:      role,
		}
	)

	tests := []struct {
		name   string
		args   args
		mock   func(ctrl *gomock.Controller) *mocksRepo.MockUserRepository
		expect expect
	}{
		{
			name: "success case",
			args: args{
				ctx,
				userID.String(),
				password,
			},
			mock: func(ctrl *gomock.Controller) *mocksRepo.MockUserRepository {
				userRepo := mocksRepo.NewMockUserRepository(ctrl)

				userRepo.EXPECT().
					FindByID(ctx, userID.String()).
					Return(baseUser, nil)

				userRepo.EXPECT().
					Delete(ctx, userID.String()).
					Return(nil)

				return userRepo
			},
			expect: expect{
				err: nil,
			},
		},
		{
			name: "user not found case",
			args: args{
				ctx,
				userID.String(),
				password,
			},
			mock: func(ctrl *gomock.Controller) *mocksRepo.MockUserRepository {
				userRepo := mocksRepo.NewMockUserRepository(ctrl)

				userRepo.EXPECT().
					FindByID(ctx, userID.String()).
					Return(nil, pgx.ErrNoRows)

				return userRepo
			},
			expect: expect{
				err: services.ErrUserNotFound,
			},
		},
		{
			name: "wrong password case",
			args: args{
				ctx,
				userID.String(),
				wrongPassword,
			},
			mock: func(ctrl *gomock.Controller) *mocksRepo.MockUserRepository {
				userRepo := mocksRepo.NewMockUserRepository(ctrl)

				userRepo.EXPECT().
					FindByID(ctx, userID.String()).
					Return(baseUser, nil)

				return userRepo
			},
			expect: expect{
				err: services.ErrPasswordWrong,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userRepo := tt.mock(ctrl)

			log, _ := logger.NewAdapter(&logger.Config{
				Level: logger.LevelError,
			})

			profileService, _ := services.NewProfileService(userRepo, log)

			err := profileService.Delete(tt.args.ctx, tt.args.userID, tt.args.password)

			if tt.expect.err != nil {
				require.Error(t, err)
				require.Equal(t, tt.expect.err.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
