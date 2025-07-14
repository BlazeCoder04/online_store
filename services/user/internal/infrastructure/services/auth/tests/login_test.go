package tests

import (
	"context"
	"testing"
	"time"

	"github.com/BlazeCoder04/online_store/libs/hash"
	"github.com/BlazeCoder04/online_store/libs/logger"
	"github.com/BlazeCoder04/online_store/services/user/configs"
	"github.com/BlazeCoder04/online_store/services/user/internal/domain/models"
	mocksAdapter "github.com/BlazeCoder04/online_store/services/user/internal/domain/ports/adapters/cache/redis/mocks"
	mocksRepo "github.com/BlazeCoder04/online_store/services/user/internal/domain/ports/repositories/mocks"
	services "github.com/BlazeCoder04/online_store/services/user/internal/infrastructure/services/auth"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func TestAuthService_Login(t *testing.T) {
	type args struct {
		ctx      context.Context
		email    string
		password string
	}

	type expect struct {
		err   error
		user  *models.User
		token bool
	}

	var (
		ctx = context.Background()

		userID            = uuid.New()
		correctEmail      = "test1@test.ru"
		wrongEmail        = "test2@test.ru"
		correctPassword   = "correct_password"
		wrongPassword     = "wrong_passwod"
		hashedPassword, _ = hash.HashPassword(correctPassword)
		firstName         = gofakeit.FirstName()
		lastName          = gofakeit.LastName()

		accessTokenPrivateKey = generateRSAPrivateKeyBase64(t)
		accessTokenExpiresIn  = 15 * time.Minute

		refreshTokenPrivateKey = generateRSAPrivateKeyBase64(t)
		refreshTokenExpiresIn  = 10080 * time.Minute

		baseUser = &models.User{
			ID:        userID,
			Email:     correctEmail,
			Password:  hashedPassword,
			FirstName: firstName,
			LastName:  lastName,
			Role:      models.UserRole,
		}
	)

	tests := []struct {
		name                   string
		args                   args
		setupMocks             func(ctrl *gomock.Controller) (*mocksRepo.MockUserRepository, *mocksAdapter.MockTokenAdapter)
		expect                 expect
		accessTokenPrivateKey  string
		refreshTokenPrivateKey string
	}{
		{
			name: "success case",
			args: args{
				ctx:      ctx,
				email:    correctEmail,
				password: correctPassword,
			},
			setupMocks: func(ctrl *gomock.Controller) (*mocksRepo.MockUserRepository, *mocksAdapter.MockTokenAdapter) {
				userRepo := mocksRepo.NewMockUserRepository(ctrl)
				tokenAdapter := mocksAdapter.NewMockTokenAdapter(ctrl)

				userRepo.EXPECT().
					FindByEmail(ctx, correctEmail).
					Return(baseUser, nil)

				tokenAdapter.EXPECT().
					Set(ctx, baseUser.ID.String(), gomock.Any(), gomock.Any()).
					Return(nil)

				return userRepo, tokenAdapter
			},
			expect: expect{
				err:   nil,
				user:  baseUser,
				token: true,
			},
			accessTokenPrivateKey:  accessTokenPrivateKey,
			refreshTokenPrivateKey: refreshTokenPrivateKey,
		},
		{
			name: "user not found case",
			args: args{
				ctx:      ctx,
				email:    wrongEmail,
				password: correctPassword,
			},
			setupMocks: func(ctrl *gomock.Controller) (*mocksRepo.MockUserRepository, *mocksAdapter.MockTokenAdapter) {
				userRepo := mocksRepo.NewMockUserRepository(ctrl)
				tokenAdapter := mocksAdapter.NewMockTokenAdapter(ctrl)

				userRepo.EXPECT().
					FindByEmail(ctx, wrongEmail).
					Return(nil, pgx.ErrNoRows)

				return userRepo, tokenAdapter
			},
			expect: expect{
				err:   services.ErrUserNotFound,
				user:  nil,
				token: false,
			},
			accessTokenPrivateKey:  accessTokenPrivateKey,
			refreshTokenPrivateKey: refreshTokenPrivateKey,
		},
		{
			name: "password wrong case",
			args: args{
				ctx:      ctx,
				email:    correctEmail,
				password: wrongPassword,
			},
			setupMocks: func(ctrl *gomock.Controller) (*mocksRepo.MockUserRepository, *mocksAdapter.MockTokenAdapter) {
				userRepo := mocksRepo.NewMockUserRepository(ctrl)
				tokenAdapter := mocksAdapter.NewMockTokenAdapter(ctrl)

				userRepo.EXPECT().
					FindByEmail(ctx, correctEmail).
					Return(baseUser, nil)

				return userRepo, tokenAdapter
			},
			expect: expect{
				err:   services.ErrPasswordWrong,
				user:  nil,
				token: false,
			},
			accessTokenPrivateKey:  accessTokenPrivateKey,
			refreshTokenPrivateKey: refreshTokenPrivateKey,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userRepo, tokenAdapter := tt.setupMocks(ctrl)

			cfg := &configs.Config{
				AccessTokenPrivateKey:  tt.accessTokenPrivateKey,
				AccessTokenExpiresIn:   accessTokenExpiresIn,
				RefreshTokenPrivateKey: tt.refreshTokenPrivateKey,
				RefreshTokenExpiresIn:  refreshTokenExpiresIn,
			}

			log, _ := logger.NewAdapter(&logger.Config{
				Level: logger.LevelError,
			})

			authService, _ := services.NewAuthService(userRepo, tokenAdapter, log, cfg)

			user, accessToken, refreshToken, err := authService.Login(tt.args.ctx, tt.args.email, tt.args.password)

			if tt.expect.err != nil {
				require.Error(t, err)
				require.Equal(t, tt.expect.err.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			if tt.expect.user != nil {
				require.NotNil(t, user)
				require.Equal(t, tt.expect.user.Email, user.Email)
			} else {
				require.Nil(t, user)
			}

			if tt.expect.token {
				require.NotEmpty(t, accessToken)
				require.NotEmpty(t, refreshToken)
			} else {
				require.Empty(t, accessToken)
				require.Empty(t, refreshToken)
			}
		})
	}
}
