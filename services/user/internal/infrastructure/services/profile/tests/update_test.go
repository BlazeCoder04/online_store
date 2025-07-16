package tests

import (
	"context"
	"testing"
	"time"

	"github.com/BlazeCoder04/online_store/libs/hash"
	"github.com/BlazeCoder04/online_store/libs/jwt"
	"github.com/BlazeCoder04/online_store/libs/logger"
	"github.com/BlazeCoder04/online_store/services/user/configs"
	"github.com/BlazeCoder04/online_store/services/user/internal/domain/models"
	mocksAdapter "github.com/BlazeCoder04/online_store/services/user/internal/domain/ports/adapters/cache/redis/mocks"
	mocksRepo "github.com/BlazeCoder04/online_store/services/user/internal/domain/ports/repositories/mocks"
	domain "github.com/BlazeCoder04/online_store/services/user/internal/domain/ports/services"
	services "github.com/BlazeCoder04/online_store/services/user/internal/infrastructure/services/profile"
	"github.com/go-redis/redis/v8"
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

		accessTokenPrivateKey = generateRSAPrivateKeyBase64(t)
		accessTokenPublicKey  = generateRSAPublicKeyBase64(t, accessTokenPrivateKey)
		accessTokenExpiresIn  = 15 * time.Minute

		refreshTokenPrivateKey = generateRSAPrivateKeyBase64(t)
		refreshTokenPublicKey  = generateRSAPublicKeyBase64(t, refreshTokenPrivateKey)
		refreshTokenExpiresIn  = 10080 * time.Minute

		accessToken, _  = jwt.Create(accessTokenExpiresIn, userID.String(), string(role), accessTokenPrivateKey)
		refreshToken, _ = jwt.Create(refreshTokenExpiresIn, userID.String(), string(role), refreshTokenPrivateKey)

		wrongAccessTokenPrivateKey = generateRSAPrivateKeyBase64(t)
		wrongAccessToken, _        = jwt.Create(accessTokenExpiresIn, userID.String(), string(role), wrongAccessTokenPrivateKey)

		wrongRefreshTokenPrivateKey = generateRSAPrivateKeyBase64(t)
		wrongRefreshToken, _        = jwt.Create(refreshTokenExpiresIn, userID.String(), string(role), wrongRefreshTokenPrivateKey)

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
		mock   func(ctrl *gomock.Controller) (*mocksRepo.MockUserRepository, *mocksAdapter.MockTokenAdapter)
		expect expect
	}{
		{
			name: "success update email case",
			args: args{
				ctx,
				&domain.UpdateProfileArgs{
					UserID:      userID.String(),
					Password:    password,
					NewEmail:    &newEmail,
					AccessToken: accessToken,
				},
			},
			mock: func(ctrl *gomock.Controller) (*mocksRepo.MockUserRepository, *mocksAdapter.MockTokenAdapter) {
				userRepo := mocksRepo.NewMockUserRepository(ctrl)
				tokenAdapter := mocksAdapter.NewMockTokenAdapter(ctrl)

				tokenAdapter.EXPECT().
					Get(ctx, userID.String()).
					Return(refreshToken, nil)

				userRepo.EXPECT().
					FindByID(ctx, userID.String()).
					Return(baseUser, nil)

				userRepo.EXPECT().
					Update(ctx, userID.String(), &newEmail, nil, nil, nil).
					Return(
						getUpdatedUser(&newEmail, nil, nil, nil),
						nil,
					)

				return userRepo, tokenAdapter
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
					AccessToken: accessToken,
				},
			},
			mock: func(ctrl *gomock.Controller) (*mocksRepo.MockUserRepository, *mocksAdapter.MockTokenAdapter) {
				userRepo := mocksRepo.NewMockUserRepository(ctrl)
				tokenAdapter := mocksAdapter.NewMockTokenAdapter(ctrl)

				tokenAdapter.EXPECT().
					Get(ctx, userID.String()).
					Return(refreshToken, nil)

				userRepo.EXPECT().
					FindByID(ctx, userID.String()).
					Return(baseUser, nil)

				userRepo.EXPECT().
					Update(ctx, userID.String(), nil, gomock.Any(), nil, nil).
					Return(baseUser, nil)

				return userRepo, tokenAdapter
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
					AccessToken:  accessToken,
				},
			},
			mock: func(ctrl *gomock.Controller) (*mocksRepo.MockUserRepository, *mocksAdapter.MockTokenAdapter) {
				userRepo := mocksRepo.NewMockUserRepository(ctrl)
				tokenAdapter := mocksAdapter.NewMockTokenAdapter(ctrl)

				tokenAdapter.EXPECT().
					Get(ctx, userID.String()).
					Return(refreshToken, nil)

				userRepo.EXPECT().
					FindByID(ctx, userID.String()).
					Return(baseUser, nil)

				userRepo.EXPECT().
					Update(ctx, userID.String(), nil, nil, &newFirstName, nil).
					Return(
						getUpdatedUser(nil, nil, &newFirstName, nil),
						nil,
					)

				return userRepo, tokenAdapter
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
					AccessToken: accessToken,
				},
			},
			mock: func(ctrl *gomock.Controller) (*mocksRepo.MockUserRepository, *mocksAdapter.MockTokenAdapter) {
				userRepo := mocksRepo.NewMockUserRepository(ctrl)
				tokenAdapter := mocksAdapter.NewMockTokenAdapter(ctrl)

				tokenAdapter.EXPECT().
					Get(ctx, userID.String()).
					Return(refreshToken, nil)

				userRepo.EXPECT().
					FindByID(ctx, userID.String()).
					Return(baseUser, nil)

				userRepo.EXPECT().
					Update(ctx, userID.String(), nil, nil, nil, &newLastName).
					Return(
						getUpdatedUser(nil, nil, nil, &newLastName),
						nil,
					)

				return userRepo, tokenAdapter
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
					AccessToken:  accessToken,
				},
			},
			mock: func(ctrl *gomock.Controller) (*mocksRepo.MockUserRepository, *mocksAdapter.MockTokenAdapter) {
				userRepo := mocksRepo.NewMockUserRepository(ctrl)
				tokenAdapter := mocksAdapter.NewMockTokenAdapter(ctrl)

				tokenAdapter.EXPECT().
					Get(ctx, userID.String()).
					Return(refreshToken, nil)

				userRepo.EXPECT().
					FindByID(ctx, userID.String()).
					Return(baseUser, nil)

				userRepo.EXPECT().
					Update(ctx, userID.String(), &newEmail, gomock.Any(), &newFirstName, &newLastName).
					Return(
						getUpdatedUser(&newEmail, nil, &newFirstName, &newLastName),
						nil,
					)

				return userRepo, tokenAdapter
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
					UserID:      userID.String(),
					Password:    password,
					NewEmail:    &email,
					AccessToken: accessToken,
				},
			},
			mock: func(ctrl *gomock.Controller) (*mocksRepo.MockUserRepository, *mocksAdapter.MockTokenAdapter) {
				userRepo := mocksRepo.NewMockUserRepository(ctrl)
				tokenAdapter := mocksAdapter.NewMockTokenAdapter(ctrl)

				tokenAdapter.EXPECT().
					Get(ctx, userID.String()).
					Return(refreshToken, nil)

				userRepo.EXPECT().
					FindByID(ctx, userID.String()).
					Return(baseUser, nil)

				return userRepo, tokenAdapter
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
					AccessToken: accessToken,
				},
			},
			mock: func(ctrl *gomock.Controller) (*mocksRepo.MockUserRepository, *mocksAdapter.MockTokenAdapter) {
				userRepo := mocksRepo.NewMockUserRepository(ctrl)
				tokenAdapter := mocksAdapter.NewMockTokenAdapter(ctrl)

				tokenAdapter.EXPECT().
					Get(ctx, userID.String()).
					Return(refreshToken, nil)

				userRepo.EXPECT().
					FindByID(ctx, userID.String()).
					Return(baseUser, nil)

				return userRepo, tokenAdapter
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
					AccessToken:  accessToken,
				},
			},
			mock: func(ctrl *gomock.Controller) (*mocksRepo.MockUserRepository, *mocksAdapter.MockTokenAdapter) {
				userRepo := mocksRepo.NewMockUserRepository(ctrl)
				tokenAdapter := mocksAdapter.NewMockTokenAdapter(ctrl)

				tokenAdapter.EXPECT().
					Get(ctx, userID.String()).
					Return(refreshToken, nil)

				userRepo.EXPECT().
					FindByID(ctx, userID.String()).
					Return(baseUser, nil)

				return userRepo, tokenAdapter
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
					AccessToken: accessToken,
				},
			},
			mock: func(ctrl *gomock.Controller) (*mocksRepo.MockUserRepository, *mocksAdapter.MockTokenAdapter) {
				userRepo := mocksRepo.NewMockUserRepository(ctrl)
				tokenAdapter := mocksAdapter.NewMockTokenAdapter(ctrl)

				tokenAdapter.EXPECT().
					Get(ctx, userID.String()).
					Return(refreshToken, nil)

				userRepo.EXPECT().
					FindByID(ctx, userID.String()).
					Return(baseUser, nil)

				return userRepo, tokenAdapter
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
					UserID:      userID.String(),
					Password:    password,
					AccessToken: accessToken,
				},
			},
			mock: func(ctrl *gomock.Controller) (*mocksRepo.MockUserRepository, *mocksAdapter.MockTokenAdapter) {
				userRepo := mocksRepo.NewMockUserRepository(ctrl)
				tokenAdapter := mocksAdapter.NewMockTokenAdapter(ctrl)

				tokenAdapter.EXPECT().
					Get(ctx, userID.String()).
					Return(refreshToken, nil)

				userRepo.EXPECT().
					FindByID(ctx, userID.String()).
					Return(nil, pgx.ErrNoRows)

				return userRepo, tokenAdapter
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
					UserID:      userID.String(),
					Password:    wrongPassword,
					AccessToken: accessToken,
				},
			},
			mock: func(ctrl *gomock.Controller) (*mocksRepo.MockUserRepository, *mocksAdapter.MockTokenAdapter) {
				userRepo := mocksRepo.NewMockUserRepository(ctrl)
				tokenAdapter := mocksAdapter.NewMockTokenAdapter(ctrl)

				tokenAdapter.EXPECT().
					Get(ctx, userID.String()).
					Return(refreshToken, nil)

				userRepo.EXPECT().
					FindByID(ctx, userID.String()).
					Return(baseUser, nil)

				return userRepo, tokenAdapter
			},
			expect: expect{
				services.ErrPasswordWrong,
				nil,
			},
		},
		{
			name: "access token invalid case",
			args: args{
				ctx,
				&domain.UpdateProfileArgs{
					UserID:      userID.String(),
					Password:    password,
					AccessToken: wrongAccessToken,
				},
			},
			mock: func(ctrl *gomock.Controller) (*mocksRepo.MockUserRepository, *mocksAdapter.MockTokenAdapter) {
				userRepo := mocksRepo.NewMockUserRepository(ctrl)
				tokenAdapter := mocksAdapter.NewMockTokenAdapter(ctrl)

				return userRepo, tokenAdapter
			},
			expect: expect{
				err:  services.ErrTokenInvalid,
				user: nil,
			},
		},
		{
			name: "refresh token not found in redis case",
			args: args{
				ctx,
				&domain.UpdateProfileArgs{
					UserID:      userID.String(),
					Password:    password,
					AccessToken: accessToken,
				},
			},
			mock: func(ctrl *gomock.Controller) (*mocksRepo.MockUserRepository, *mocksAdapter.MockTokenAdapter) {
				userRepo := mocksRepo.NewMockUserRepository(ctrl)
				tokenAdapter := mocksAdapter.NewMockTokenAdapter(ctrl)

				tokenAdapter.EXPECT().
					Get(ctx, userID.String()).
					Return("", redis.Nil)

				return userRepo, tokenAdapter
			},
			expect: expect{
				err:  services.ErrTokenInvalid,
				user: nil,
			},
		},
		{
			name: "refresh token invalid case",
			args: args{
				ctx,
				&domain.UpdateProfileArgs{
					UserID:      userID.String(),
					Password:    password,
					AccessToken: accessToken,
				},
			},
			mock: func(ctrl *gomock.Controller) (*mocksRepo.MockUserRepository, *mocksAdapter.MockTokenAdapter) {
				userRepo := mocksRepo.NewMockUserRepository(ctrl)
				tokenAdapter := mocksAdapter.NewMockTokenAdapter(ctrl)

				tokenAdapter.EXPECT().
					Get(ctx, userID.String()).
					Return(wrongRefreshToken, nil)

				return userRepo, tokenAdapter
			},
			expect: expect{
				err:  services.ErrTokenInvalid,
				user: nil,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userRepo, tokenAdapter := tt.mock(ctrl)

			log, _ := logger.NewAdapter(&logger.Config{
				Level: logger.LevelError,
			})

			cfg := &configs.Config{
				AccessTokenPrivateKey:  accessTokenPrivateKey,
				AccessTokenPublicKey:   accessTokenPublicKey,
				AccessTokenExpiresIn:   accessTokenExpiresIn,
				RefreshTokenPrivateKey: refreshTokenPrivateKey,
				RefreshTokenPublicKey:  refreshTokenPublicKey,
				RefreshTokenExpiresIn:  refreshTokenExpiresIn,
			}

			profileService, _ := services.NewProfileService(userRepo, tokenAdapter, log, cfg)

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
