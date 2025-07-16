package tests

import (
	"context"
	"testing"
	"time"

	"github.com/BlazeCoder04/online_store/libs/jwt"
	"github.com/BlazeCoder04/online_store/libs/logger"
	"github.com/BlazeCoder04/online_store/services/user/configs"
	"github.com/BlazeCoder04/online_store/services/user/internal/domain/models"
	mocksAdapter "github.com/BlazeCoder04/online_store/services/user/internal/domain/ports/adapters/cache/redis/mocks"
	mocksRepo "github.com/BlazeCoder04/online_store/services/user/internal/domain/ports/repositories/mocks"
	services "github.com/BlazeCoder04/online_store/services/user/internal/infrastructure/services/auth"
	"github.com/go-redis/redis/v8"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func TestAuthService_RefreshToken(t *testing.T) {
	type args struct {
		ctx          context.Context
		refreshToken string
	}

	type expect struct {
		err   error
		token bool
	}

	var (
		ctx = context.Background()

		userID = uuid.New()
		email  = "test@test.ru"
		role   = models.UserRole

		accessTokenPrivateKey = generateRSAPrivateKeyBase64(t)
		accessTokenPublicKey  = generateRSAPublicKeyBase64(t, accessTokenPrivateKey)
		accessTokenExpiresIn  = 15 * time.Minute

		refreshTokenPrivateKey = generateRSAPrivateKeyBase64(t)
		refreshTokenPublicKey  = generateRSAPublicKeyBase64(t, refreshTokenPrivateKey)
		refreshTokenExpiresIn  = 10080 * time.Minute

		refreshToken, _ = jwt.Create(refreshTokenExpiresIn, userID.String(), string(role), refreshTokenPrivateKey)

		wrongRefreshTokenPrivateKey = generateRSAPrivateKeyBase64(t)
		wrongRefreshToken, _        = jwt.Create(refreshTokenExpiresIn, userID.String(), string(role), wrongRefreshTokenPrivateKey)

		baseUser = &models.User{
			ID:    userID,
			Email: email,
			Role:  role,
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
				refreshToken,
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
				err:   nil,
				token: true,
			},
		},
		{
			name: "token invalid case",
			args: args{
				ctx,
				wrongRefreshToken,
			},
			mock: func(ctrl *gomock.Controller) (*mocksRepo.MockUserRepository, *mocksAdapter.MockTokenAdapter) {
				userRepo := mocksRepo.NewMockUserRepository(ctrl)
				tokenAdapter := mocksAdapter.NewMockTokenAdapter(ctrl)

				return userRepo, tokenAdapter
			},
			expect: expect{
				err:   services.ErrTokenInvalid,
				token: false,
			},
		},
		{
			name: "token not found in redis case",
			args: args{
				ctx,
				refreshToken,
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
				err:   services.ErrTokenInvalid,
				token: false,
			},
		},
		{
			name: "token mismatch case",
			args: args{
				ctx,
				refreshToken,
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
				err:   services.ErrTokenInvalid,
				token: false,
			},
		},
		{
			name: "user not found case",
			args: args{
				ctx,
				refreshToken,
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
				err:   services.ErrUserNotFound,
				token: false,
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

			cfg := &configs.Config{
				AccessTokenPrivateKey:  accessTokenPrivateKey,
				AccessTokenPublicKey:   accessTokenPublicKey,
				AccessTokenExpiresIn:   accessTokenExpiresIn,
				RefreshTokenPrivateKey: refreshTokenPrivateKey,
				RefreshTokenPublicKey:  refreshTokenPublicKey,
				RefreshTokenExpiresIn:  refreshTokenExpiresIn,
			}

			log, _ := logger.NewAdapter(&logger.Config{
				Level: logger.LevelError,
			})

			authService, _ := services.NewAuthService(userRepo, tokenAdapter, log, cfg)

			accessToken, err := authService.RefreshToken(tt.args.ctx, tt.args.refreshToken)

			if tt.expect.err != nil {
				require.Error(t, err)
				require.Equal(t, tt.expect.err.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			if tt.expect.token {
				require.NotEmpty(t, accessToken)
			} else {
				require.Empty(t, accessToken)
			}
		})
	}
}
