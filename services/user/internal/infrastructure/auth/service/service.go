package service

import (
	"context"
	"errors"
	"strings"

	hashPassword "github.com/BlazeCoder04/online_store/libs/hash_password"
	"github.com/BlazeCoder04/online_store/libs/jwt"
	"github.com/BlazeCoder04/online_store/libs/logger"
	"github.com/BlazeCoder04/online_store/services/user/configs"
	"github.com/BlazeCoder04/online_store/services/user/internal/domain/models"
	domainRepo "github.com/BlazeCoder04/online_store/services/user/internal/domain/ports/auth/repository"
	domainService "github.com/BlazeCoder04/online_store/services/user/internal/domain/ports/auth/service"
	"github.com/jackc/pgx/v5"
)

type AuthService struct {
	authRepo  domainRepo.AuthRepository
	tokenRepo domainRepo.TokenRepository
	logger    logger.Logger
	cfg       *configs.Config
}

func NewAuthService(authRepo domainRepo.AuthRepository, tokenRepo domainRepo.TokenRepository, logger logger.Logger, cfg *configs.Config) (domainService.AuthService, error) {
	logger.Info("Initializing auth service")
	logger.OK("Successful initialization")

	return &AuthService{
		authRepo,
		tokenRepo,
		logger,
		cfg,
	}, nil
}

func (s *AuthService) generateAndStoreTokens(ctx context.Context, userID string) (string, string, error) {
	accessToken, err := jwt.Create(s.cfg.AccessTokenExpiresIn, userID, s.cfg.AccessTokenPrivateKey)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := jwt.Create(s.cfg.RefreshTokenExpiresIn, userID, s.cfg.RefreshTokenPrivateKey)
	if err != nil {
		return "", "", err
	}

	if err = s.tokenRepo.Set(ctx, userID, refreshToken, s.cfg.RefreshTokenExpiresIn); err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*models.User, string, string, error) {
	email = strings.ToLower(email)

	user, err := s.authRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", "", errors.New(ErrUserNotFound)
		}

		return nil, "", "", err
	}

	if err = hashPassword.ComparePassword(user.Password, password); err != nil {
		return nil, "", "", errors.New(ErrPasswordWrong)
	}

	accessToken, refreshToken, err := s.generateAndStoreTokens(ctx, user.ID.String())
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (s *AuthService) Register(ctx context.Context, email, password, firstName, lastName string) (*models.User, string, string, error) {
	email = strings.ToLower(email)

	existedUser, _ := s.authRepo.FindByEmail(ctx, email)
	if existedUser != nil {
		return nil, "", "", errors.New(ErrUserExists)
	}

	hashedPassword := hashPassword.HashPassword(password)

	user, err := s.authRepo.Create(ctx, email, hashedPassword, firstName, lastName)
	if err != nil {
		return nil, "", "", err
	}

	accessToken, refreshToken, err := s.generateAndStoreTokens(ctx, user.ID.String())
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (s *AuthService) UpdateTokens(ctx context.Context, refreshToken string) (string, string, error) {
	userID, err := jwt.Validate(refreshToken, s.cfg.RefreshTokenPublicKey)
	if err != nil {
		return "", "", errors.New(ErrTokenInvalid)
	}

	storedRefreshToken, err := s.tokenRepo.Get(ctx, userID)
	if err != nil {
		return "", "", errors.New(ErrTokenNotFound)
	}

	if storedRefreshToken != refreshToken {
		return "", "", errors.New(ErrTokenInvalid)
	}

	newAccessToken, newRefreshToken, err := s.generateAndStoreTokens(ctx, userID)
	if err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

func (s *AuthService) Logout(ctx context.Context, accessToken string) error {
	userID, err := jwt.Validate(accessToken, s.cfg.AccessTokenPublicKey)
	if err != nil {
		return errors.New(ErrTokenInvalid)
	}

	refreshToken, err := s.tokenRepo.Get(ctx, userID)
	if err != nil {
		return errors.New(ErrTokenNotFound)
	}

	if _, err = jwt.Validate(refreshToken, s.cfg.RefreshTokenPublicKey); err != nil {
		return errors.New(ErrTokenInvalid)
	}

	if err = s.tokenRepo.Delete(ctx, userID); err != nil {
		return err
	}

	return nil
}
