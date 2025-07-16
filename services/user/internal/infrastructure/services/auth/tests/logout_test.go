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
	services "github.com/BlazeCoder04/online_store/services/user/internal/infrastructure/services/auth"
	"github.com/go-redis/redis/v8"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestAuthService_Logout(t *testing.T) {
	type args struct {
		ctx         context.Context
		accessToken string
	}

	type expect struct {
		err error
	}

	var (
		ctx = context.Background()

		userID = uuid.New()
		role   = models.UserRole

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
	)

	tests := []struct {
		name   string
		args   args
		mock   func(ctrl *gomock.Controller) *mocksAdapter.MockTokenAdapter
		expect expect
	}{
		{
			name: "success case",
			args: args{
				ctx,
				accessToken,
			},
			mock: func(ctrl *gomock.Controller) *mocksAdapter.MockTokenAdapter {
				tokenAdapter := mocksAdapter.NewMockTokenAdapter(ctrl)

				tokenAdapter.EXPECT().
					Get(ctx, userID.String()).
					Return(refreshToken, nil)

				tokenAdapter.EXPECT().
					Del(ctx, userID.String()).
					Return(nil)

				return tokenAdapter
			},
			expect: expect{
				err: nil,
			},
		},
		{
			name: "access token invalid case",
			args: args{
				ctx,
				wrongAccessToken,
			},
			mock: func(ctrl *gomock.Controller) *mocksAdapter.MockTokenAdapter {
				tokenAdapter := mocksAdapter.NewMockTokenAdapter(ctrl)

				return tokenAdapter
			},
			expect: expect{
				err: services.ErrTokenInvalid,
			},
		},
		{
			name: "refresh token not found in redis case",
			args: args{
				ctx,
				accessToken,
			},
			mock: func(ctrl *gomock.Controller) *mocksAdapter.MockTokenAdapter {
				tokenAdapter := mocksAdapter.NewMockTokenAdapter(ctrl)

				tokenAdapter.EXPECT().
					Get(ctx, userID.String()).
					Return("", redis.Nil)

				return tokenAdapter
			},
			expect: expect{
				err: services.ErrTokenInvalid,
			},
		},
		{
			name: "refresh token invalid case",
			args: args{
				ctx,
				accessToken,
			},
			mock: func(ctrl *gomock.Controller) *mocksAdapter.MockTokenAdapter {
				tokenAdapter := mocksAdapter.NewMockTokenAdapter(ctrl)

				tokenAdapter.EXPECT().
					Get(ctx, userID.String()).
					Return(wrongRefreshToken, nil)

				return tokenAdapter
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

			tokenAdapter := tt.mock(ctrl)

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

			authService, _ := services.NewAuthService(nil, tokenAdapter, log, cfg)

			err := authService.Logout(tt.args.ctx, tt.args.accessToken)

			if tt.expect.err != nil {
				require.Error(t, err)
				require.Equal(t, tt.expect.err.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
