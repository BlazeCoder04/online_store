package tests

import (
	"context"
	"testing"

	"github.com/BlazeCoder04/online_store/libs/hash"
	"github.com/BlazeCoder04/online_store/libs/logger"
	"github.com/BlazeCoder04/online_store/services/user/internal/domain/models"
	mocksRepo "github.com/BlazeCoder04/online_store/services/user/internal/domain/ports/repositories/mocks"
	domain "github.com/BlazeCoder04/online_store/services/user/internal/domain/ports/services"
	services "github.com/BlazeCoder04/online_store/services/user/internal/infrastructure/services/profile"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func TestProfileService_Update(t *testing.T) {
	type args struct {
		ctx context.Context
		in  *domain.UpdateProfileArgs
	}

	type expect struct {
		err  error
		user *models.User
	}

	var (
		ctx = context.Background()

		userID            = uuid.New()
		email             = "test@test.ru"
		newEmail          = "test1@test.ru"
		password          = "password"
		hashedPassword, _ = hash.HashPassword(password)
		wrongPassword     = "wrong_password"
		newPassword       = "new_password"
		firstName         = "John"
		newFirstName      = "Mike"
		lastName          = "Doe"
		newLastName       = "Smith"
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

	getUpdatedUser := func(newEmail, newPassword, newFirstName, newLastName *string) *models.User {
		u := *baseUser

		if newEmail != nil {
			u.Email = *newEmail
		}
		if newPassword != nil {
			u.Password = *newPassword
		}
		if newFirstName != nil {
			u.FirstName = *newFirstName
		}
		if newLastName != nil {
			u.LastName = *newLastName
		}

		return &u
	}

	tests := []struct {
		name   string
		args   args
		mock   func(ctrl *gomock.Controller) *mocksRepo.MockUserRepository
		expect expect
	}{
		{
			name: "success update email case",
			args: args{
				ctx,
				&domain.UpdateProfileArgs{
					UserID:   userID.String(),
					Password: password,
					NewEmail: &newEmail,
				},
			},
			mock: func(ctrl *gomock.Controller) *mocksRepo.MockUserRepository {
				userRepo := mocksRepo.NewMockUserRepository(ctrl)

				userRepo.EXPECT().
					FindByID(ctx, userID.String()).
					Return(baseUser, nil)

				userRepo.EXPECT().
					Update(ctx, userID.String(), &newEmail, nil, nil, nil).
					Return(
						getUpdatedUser(&newEmail, nil, nil, nil),
						nil,
					)

				return userRepo
			},
			expect: expect{
				nil,
				getUpdatedUser(&newEmail, nil, nil, nil),
			},
		},
		{
			name: "success update password case",
			args: args{
				ctx,
				&domain.UpdateProfileArgs{
					UserID:      userID.String(),
					Password:    password,
					NewPassword: &newPassword,
				},
			},
			mock: func(ctrl *gomock.Controller) *mocksRepo.MockUserRepository {
				userRepo := mocksRepo.NewMockUserRepository(ctrl)

				userRepo.EXPECT().
					FindByID(ctx, userID.String()).
					Return(baseUser, nil)

				userRepo.EXPECT().
					Update(ctx, userID.String(), nil, gomock.Any(), nil, nil).
					Return(baseUser, nil)

				return userRepo
			},
			expect: expect{
				nil,
				baseUser,
			},
		},
		{
			name: "success update firstName case",
			args: args{
				ctx,
				&domain.UpdateProfileArgs{
					UserID:       userID.String(),
					Password:     password,
					NewFirstName: &newFirstName,
				},
			},
			mock: func(ctrl *gomock.Controller) *mocksRepo.MockUserRepository {
				userRepo := mocksRepo.NewMockUserRepository(ctrl)

				userRepo.EXPECT().
					FindByID(ctx, userID.String()).
					Return(baseUser, nil)

				userRepo.EXPECT().
					Update(ctx, userID.String(), nil, nil, &newFirstName, nil).
					Return(
						getUpdatedUser(nil, nil, &newFirstName, nil),
						nil,
					)

				return userRepo
			},
			expect: expect{
				nil,
				getUpdatedUser(nil, nil, &newFirstName, nil),
			},
		},
		{
			name: "success update lastName case",
			args: args{
				ctx,
				&domain.UpdateProfileArgs{
					UserID:      userID.String(),
					Password:    password,
					NewLastName: &newLastName,
				},
			},
			mock: func(ctrl *gomock.Controller) *mocksRepo.MockUserRepository {
				userRepo := mocksRepo.NewMockUserRepository(ctrl)

				userRepo.EXPECT().
					FindByID(ctx, userID.String()).
					Return(baseUser, nil)

				userRepo.EXPECT().
					Update(ctx, userID.String(), nil, nil, nil, &newLastName).
					Return(
						getUpdatedUser(nil, nil, nil, &newLastName),
						nil,
					)

				return userRepo
			},
			expect: expect{
				nil,
				getUpdatedUser(nil, nil, nil, &newLastName),
			},
		},
		{
			name: "success update all feilds user case",
			args: args{
				ctx,
				&domain.UpdateProfileArgs{
					UserID:       userID.String(),
					Password:     password,
					NewEmail:     &newEmail,
					NewPassword:  &newPassword,
					NewFirstName: &newFirstName,
					NewLastName:  &newLastName,
				},
			},
			mock: func(ctrl *gomock.Controller) *mocksRepo.MockUserRepository {
				userRepo := mocksRepo.NewMockUserRepository(ctrl)

				userRepo.EXPECT().
					FindByID(ctx, userID.String()).
					Return(baseUser, nil)

				userRepo.EXPECT().
					Update(ctx, userID.String(), &newEmail, gomock.Any(), &newFirstName, &newLastName).
					Return(
						getUpdatedUser(&newEmail, nil, &newFirstName, &newLastName),
						nil,
					)

				return userRepo
			},
			expect: expect{
				nil,
				getUpdatedUser(&newEmail, nil, &newFirstName, &newLastName),
			},
		},
		{
			name: "email unchanged case",
			args: args{
				ctx,
				&domain.UpdateProfileArgs{
					UserID:   userID.String(),
					Password: password,
					NewEmail: &email,
				},
			},
			mock: func(ctrl *gomock.Controller) *mocksRepo.MockUserRepository {
				userRepo := mocksRepo.NewMockUserRepository(ctrl)

				userRepo.EXPECT().
					FindByID(ctx, userID.String()).
					Return(baseUser, nil)

				return userRepo
			},
			expect: expect{
				services.ErrEmailUnchanged,
				nil,
			},
		},
		{
			name: "password unchanged case",
			args: args{
				ctx,
				&domain.UpdateProfileArgs{
					UserID:      userID.String(),
					Password:    password,
					NewPassword: &password,
				},
			},
			mock: func(ctrl *gomock.Controller) *mocksRepo.MockUserRepository {
				userRepo := mocksRepo.NewMockUserRepository(ctrl)

				userRepo.EXPECT().
					FindByID(ctx, userID.String()).
					Return(baseUser, nil)

				return userRepo
			},
			expect: expect{
				services.ErrPasswordUnchanged,
				nil,
			},
		},
		{
			name: "firstName unchanged case",
			args: args{
				ctx,
				&domain.UpdateProfileArgs{
					UserID:       userID.String(),
					Password:     password,
					NewFirstName: &firstName,
				},
			},
			mock: func(ctrl *gomock.Controller) *mocksRepo.MockUserRepository {
				userRepo := mocksRepo.NewMockUserRepository(ctrl)

				userRepo.EXPECT().
					FindByID(ctx, userID.String()).
					Return(baseUser, nil)

				return userRepo
			},
			expect: expect{
				services.ErrFirstNameUnchanged,
				nil,
			},
		},
		{
			name: "lastName unchanged case",
			args: args{
				ctx,
				&domain.UpdateProfileArgs{
					UserID:      userID.String(),
					Password:    password,
					NewLastName: &lastName,
				},
			},
			mock: func(ctrl *gomock.Controller) *mocksRepo.MockUserRepository {
				userRepo := mocksRepo.NewMockUserRepository(ctrl)

				userRepo.EXPECT().
					FindByID(ctx, userID.String()).
					Return(baseUser, nil)

				return userRepo
			},
			expect: expect{
				services.ErrLastNameUnchanged,
				nil,
			},
		},
		{
			name: "user not found case",
			args: args{
				ctx,
				&domain.UpdateProfileArgs{
					UserID:   userID.String(),
					Password: password,
				},
			},
			mock: func(ctrl *gomock.Controller) *mocksRepo.MockUserRepository {
				userRepo := mocksRepo.NewMockUserRepository(ctrl)

				userRepo.EXPECT().
					FindByID(ctx, userID.String()).
					Return(nil, pgx.ErrNoRows)

				return userRepo
			},
			expect: expect{
				services.ErrUserNotFound,
				nil,
			},
		},
		{
			name: "password wrong case",
			args: args{
				ctx,
				&domain.UpdateProfileArgs{
					UserID:   userID.String(),
					Password: wrongPassword,
				},
			},
			mock: func(ctrl *gomock.Controller) *mocksRepo.MockUserRepository {
				userRepo := mocksRepo.NewMockUserRepository(ctrl)

				userRepo.EXPECT().
					FindByID(ctx, userID.String()).
					Return(baseUser, nil)

				return userRepo
			},
			expect: expect{
				services.ErrPasswordWrong,
				nil,
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

			user, err := profileService.Update(ctx, tt.args.in)

			if tt.expect.err != nil {
				require.Error(t, err)
				require.Equal(t, tt.expect.err.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			if tt.expect.user != nil {
				require.NotNil(t, user)
				require.Equal(t, tt.expect.user.ID, user.ID)
				require.Equal(t, tt.expect.user.Email, user.Email)
				require.Equal(t, tt.expect.user.Password, user.Password)
				require.Equal(t, tt.expect.user.FirstName, user.FirstName)
				require.Equal(t, tt.expect.user.LastName, user.LastName)
			} else {
				require.Nil(t, user)
			}
		})
	}
}
