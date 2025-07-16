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
	services "github.com/BlazeCoder04/online_store/services/user/internal/infrastructure/services/profile"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/go-redis/redis/v8"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func TestProfileService_Delete(t *testing.T) {
	type args struct {
		ctx         context.Context
		userID      string
		password    string
		accessToken string
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

	tests := []struct {
		name   string
		args   args
		mock   func(ctrl *gomock.Controller) (*mocksRepo.MockUserRepository, *mocksAdapter.MockTokenAdapter)
		expect expect
	}{
		{
			name: "success case",
			args: args{
				ctx,
				userID.String(),
				password,
				accessToken,
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
					Delete(ctx, userID.String()).
					Return(nil)

				tokenAdapter.EXPECT().
					Del(ctx, userID.String()).
					Return(nil)

				return userRepo, tokenAdapter
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
				accessToken,
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
				err: services.ErrUserNotFound,
			},
		},
		{
			name: "wrong password case",
			args: args{
				ctx,
				userID.String(),
				wrongPassword,
				accessToken,
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
				err: services.ErrPasswordWrong,
			},
		},
		{
			name: "access token invalid case",
			args: args{
				ctx,
				userID.String(),
				password,
				wrongAccessToken,
			},
			mock: func(ctrl *gomock.Controller) (*mocksRepo.MockUserRepository, *mocksAdapter.MockTokenAdapter) {
				userRepo := mocksRepo.NewMockUserRepository(ctrl)
				tokenAdapter := mocksAdapter.NewMockTokenAdapter(ctrl)

				return userRepo, tokenAdapter
			},
			expect: expect{
				err: services.ErrTokenInvalid,
			},
		},
		{
			name: "refresh token not found in redis case",
			args: args{
				ctx,
				userID.String(),
				password,
				accessToken,
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
				err: services.ErrTokenInvalid,
			},
		},
		{
			name: "refresh token invalid case",
			args: args{
				ctx,
				userID.String(),
				password,
				accessToken,
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
				err: services.ErrTokenInvalid,
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

			err := profileService.Delete(tt.args.ctx, tt.args.userID, tt.args.password, tt.args.accessToken)

			if tt.expect.err != nil {
				require.Error(t, err)
				require.Equal(t, tt.expect.err.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
