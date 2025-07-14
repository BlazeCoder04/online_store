package tests

import (
	"context"
	"testing"
	"time"

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

func TestAuthService_Register(t *testing.T) {
	type args struct {
		ctx       context.Context
		email     string
		password  string
		firstName string
		lastName  string
	}

	type expect struct {
		err   error
		user  *models.User
		token bool
	}

	var (
		ctx = context.Background()

		userID    = uuid.New()
		email     = "test1@test.ru"
		password  = gofakeit.Password(true, true, true, true, false, 12)
		firstName = gofakeit.FirstName()
		lastName  = gofakeit.LastName()

		accessTokenPrivateKey = generateRSAPrivateKeyBase64(t)
		accessTokenExpiresIn  = 15 * time.Minute

		refreshTokenPrivateKey = generateRSAPrivateKeyBase64(t)
		refreshTokenExpiresIn  = 10080 * time.Minute

		baseUser = &models.User{
			ID:        userID,
			Email:     email,
			Password:  password,
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
				ctx:       ctx,
				email:     email,
				password:  password,
				firstName: firstName,
				lastName:  lastName,
			},
			setupMocks: func(ctrl *gomock.Controller) (*mocksRepo.MockUserRepository, *mocksAdapter.MockTokenAdapter) {
				userRepo := mocksRepo.NewMockUserRepository(ctrl)
				tokenAdapter := mocksAdapter.NewMockTokenAdapter(ctrl)

				userRepo.EXPECT().
					FindByEmail(ctx, email).
					Return(nil, pgx.ErrNoRows)

				userRepo.EXPECT().
					Create(ctx, email, gomock.Any(), firstName, lastName).
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
			name: "user exists case",
			args: args{
				ctx:       ctx,
				email:     email,
				password:  password,
				firstName: firstName,
				lastName:  lastName,
			},
			setupMocks: func(ctrl *gomock.Controller) (*mocksRepo.MockUserRepository, *mocksAdapter.MockTokenAdapter) {
				userRepo := mocksRepo.NewMockUserRepository(ctrl)
				tokenAdapter := mocksAdapter.NewMockTokenAdapter(ctrl)

				userRepo.EXPECT().
					FindByEmail(ctx, email).
					Return(baseUser, nil)

				return userRepo, tokenAdapter
			},
			expect: expect{
				err:   services.ErrUserExists,
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
				AccessTokenPrivateKey:  accessTokenPrivateKey,
				AccessTokenExpiresIn:   accessTokenExpiresIn,
				RefreshTokenPrivateKey: refreshTokenPrivateKey,
				RefreshTokenExpiresIn:  refreshTokenExpiresIn,
			}

			log, _ := logger.NewAdapter(&logger.Config{
				Level: logger.LevelError,
			})

			authService, _ := services.NewAuthService(userRepo, tokenAdapter, log, cfg)

			user, accessToken, refreshToken, err := authService.Register(tt.args.ctx, tt.args.email, tt.args.password, tt.args.firstName, tt.args.lastName)

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
