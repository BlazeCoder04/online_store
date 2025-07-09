package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/BlazeCoder04/online_store/libs/hash"
	"github.com/BlazeCoder04/online_store/libs/jwt"
	"github.com/BlazeCoder04/online_store/libs/logger"
	"github.com/BlazeCoder04/online_store/services/user/configs"
	"github.com/BlazeCoder04/online_store/services/user/internal/domain/models"
	domainRepo "github.com/BlazeCoder04/online_store/services/user/internal/domain/ports/repositories"
	domainService "github.com/BlazeCoder04/online_store/services/user/internal/domain/ports/services"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo  domainRepo.UserRepository
	tokenRepo domainRepo.TokenRepository
	logger    logger.Logger
	cfg       *configs.Config
}

func NewAuthService(userRepo domainRepo.UserRepository, tokenRepo domainRepo.TokenRepository, logger logger.Logger, cfg *configs.Config) (domainService.AuthService, error) {
	loggerTag := "auth.service.newAuthService"

	logger.Info(loggerTag, "Initializing auth service")
	logger.Info(loggerTag, "Successful initialization")

	return &AuthService{
		userRepo,
		tokenRepo,
		logger,
		cfg,
	}, nil
}

func (s *AuthService) generateAndStoreTokens(ctx context.Context, userID, userRole string) (string, string, error) {
	loggerTag := "auth.service.generateAndStoreTokens"

	accessToken, err := jwt.Create(s.cfg.AccessTokenExpiresIn, userID, userRole, s.cfg.AccessTokenPrivateKey)
	if err != nil {
		s.logger.Error(loggerTag, fmt.Sprintf("failed create access token: %v", err))

		return "", "", err
	}

	refreshToken, err := jwt.Create(s.cfg.RefreshTokenExpiresIn, userID, userRole, s.cfg.RefreshTokenPrivateKey)
	if err != nil {
		s.logger.Error(loggerTag, fmt.Sprintf("failed create refresh token: %v", err))

		return "", "", err
	}

	if err = s.tokenRepo.Set(ctx, userID, refreshToken, s.cfg.RefreshTokenExpiresIn); err != nil {
		s.logger.Error(loggerTag, fmt.Sprintf("failed add refresh token to redis: %v", err))

		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*models.User, string, string, error) {
	loggerTag := "auth.service.login"

	email = strings.ToLower(email)

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", "", ErrUserNotFound
		}

		s.logger.Error(loggerTag, fmt.Sprintf("failed find user: %v", err))

		return nil, "", "", err
	}

	if err = hash.ComparePassword(user.Password, password); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, "", "", ErrPasswordWrong
		}

		s.logger.Error(loggerTag, fmt.Sprintf("failed compare password: %v", err))

		return nil, "", "", err
	}

	accessToken, refreshToken, err := s.generateAndStoreTokens(ctx, user.ID.String(), string(user.Role))
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (s *AuthService) Register(ctx context.Context, email, password, firstName, lastName string) (*models.User, string, string, error) {
	loggerTag := "auth.service.register"

	email = strings.ToLower(email)

	existedUser, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		s.logger.Error(loggerTag, fmt.Sprintf("failed find user: %v", err))

		return nil, "", "", err
	}
	if existedUser != nil {
		return nil, "", "", ErrUserExists
	}

	hashedPassword, err := hash.HashPassword(password)
	if err != nil {
		s.logger.Error(loggerTag, fmt.Sprintf("failed hash password: %v", err))

		return nil, "", "", err
	}

	user, err := s.userRepo.Create(ctx, email, hashedPassword, firstName, lastName)
	if err != nil {
		s.logger.Error(loggerTag, fmt.Sprintf("failed create user: %v", err))

		return nil, "", "", err
	}

	accessToken, refreshToken, err := s.generateAndStoreTokens(ctx, user.ID.String(), string(user.Role))
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	loggerTag := "auth.service.refreshToken"

	claims, err := jwt.Verify(refreshToken, s.cfg.RefreshTokenPublicKey)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenInvalid) {
			return "", err
		}

		s.logger.Error(loggerTag, fmt.Sprintf("failed verify token: %v", err))

		return "", err
	}

	userID := claims["sub"].(string)

	storedRefreshToken, err := s.tokenRepo.Get(ctx, userID)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ErrTokenInvalid
		}

		s.logger.Error(loggerTag, fmt.Sprintf("failed get refresh token to redis: %v", err))

		return "", err
	}

	if storedRefreshToken != refreshToken {
		return "", ErrTokenInvalid
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrUserNotFound
		}

		s.logger.Error(loggerTag, fmt.Sprintf("failed find user: %v", err))

		return "", err
	}

	accessToken, err := jwt.Create(s.cfg.AccessTokenExpiresIn, userID, string(user.Role), s.cfg.AccessTokenPrivateKey)
	if err != nil {
		s.logger.Error(loggerTag, fmt.Sprintf("failed create access token: %v", err))

		return "", err
	}

	return accessToken, nil
}

func (s *AuthService) Logout(ctx context.Context, accessToken string) error {
	loggerTag := "auth.service.logout"

	accessTokenClaims, err := jwt.Verify(accessToken, s.cfg.AccessTokenPublicKey)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenInvalid) {
			return err
		}

		s.logger.Error(loggerTag, fmt.Sprintf("failed verify access token: %v", err))

		return err
	}

	userID := accessTokenClaims["sub"].(string)

	refreshToken, err := s.tokenRepo.Get(ctx, userID)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrTokenInvalid
		}

		s.logger.Error(loggerTag, fmt.Sprintf("failed get refresh token to redis: %v", err))

		return err
	}

	_, err = jwt.Verify(refreshToken, s.cfg.RefreshTokenPublicKey)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenInvalid) {
			return err
		}

		s.logger.Error(loggerTag, fmt.Sprintf("failed verify refresh token: %v", err))

		return err
	}

	if err = s.tokenRepo.Del(ctx, userID); err != nil {
		s.logger.Error(loggerTag, fmt.Sprintf("failed delete refresh token to redis: %v", err))

		return err
	}

	return nil
}
