package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/BlazeCoder04/online_store/libs/hash"
	"github.com/BlazeCoder04/online_store/libs/jwt"
	"github.com/BlazeCoder04/online_store/libs/logger"
	"github.com/BlazeCoder04/online_store/services/user/configs"
	"github.com/BlazeCoder04/online_store/services/user/internal/domain/models"
	domainAdapter "github.com/BlazeCoder04/online_store/services/user/internal/domain/ports/adapters/cache/redis"
	domainRepo "github.com/BlazeCoder04/online_store/services/user/internal/domain/ports/repositories"
	domainService "github.com/BlazeCoder04/online_store/services/user/internal/domain/ports/services"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type ProfileService struct {
	userRepo     domainRepo.UserRepository
	tokenAdapter domainAdapter.TokenAdapter
	logger       logger.Logger
	cfg          *configs.Config
}

func NewProfileService(userRepo domainRepo.UserRepository, tokenAdapter domainAdapter.TokenAdapter, logger logger.Logger, cfg *configs.Config) (domainService.ProfileService, error) {
	loggerTag := "profile.service.newProfileService"

	logger.Info(loggerTag, "Profile service initialized")

	return &ProfileService{
		userRepo,
		tokenAdapter,
		logger,
		cfg,
	}, nil
}

func (s *ProfileService) VerifyToken(ctx context.Context, accessToken string) error {
	loggerTag := "profile.service.verifyToken"

	accessTokenClaims, err := jwt.Verify(accessToken, s.cfg.AccessTokenPublicKey)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenInvalid) {
			return err
		}

		s.logger.Error(loggerTag, fmt.Sprintf("failed verify access token: %v", err))

		return err
	}

	userID := accessTokenClaims["sub"].(string)

	refreshToken, err := s.tokenAdapter.Get(ctx, userID)
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

	return nil
}

func (s *ProfileService) Get(ctx context.Context, userID, accessToken string) (*models.User, error) {
	loggerTag := "profile.service.get"

	if err := s.VerifyToken(ctx, accessToken); err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		s.logger.Error(loggerTag, fmt.Sprintf("failed find user: %v", err))

		return nil, err
	}

	return user, nil
}

func (s *ProfileService) Update(ctx context.Context, args *domainService.UpdateProfileArgs) (*models.User, error) {
	loggerTag := "profile.service.update"

	if err := s.VerifyToken(ctx, args.AccessToken); err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByID(ctx, args.UserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		s.logger.Error(loggerTag, fmt.Sprintf("failed find user: %v", err))

		return nil, err
	}

	if err = hash.ComparePassword(user.Password, args.Password); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, ErrPasswordWrong
		}

		s.logger.Error(loggerTag, fmt.Sprintf("failed compare password: %v", err))

		return nil, err
	}

	switch {
	case args.NewEmail != nil && *args.NewEmail == user.Email:
		return nil, ErrEmailUnchanged
	case args.NewFirstName != nil && *args.NewFirstName == user.FirstName:
		return nil, ErrFirstNameUnchanged
	case args.NewLastName != nil && *args.NewLastName == user.LastName:
		return nil, ErrLastNameUnchanged
	}

	var hashedPassword *string
	if args.NewPassword != nil {
		if passCheckErr := hash.ComparePassword(user.Password, *args.NewPassword); passCheckErr == nil {
			return nil, ErrPasswordUnchanged
		}

		hashed, hashErr := hash.HashPassword(*args.NewPassword)
		if hashErr != nil {
			s.logger.Error(loggerTag, fmt.Sprintf("failed hash password: %v", err))

			return nil, err
		}

		hashedPassword = &hashed
	}

	updatedUser, err := s.userRepo.Update(ctx, args.UserID, args.NewEmail, hashedPassword, args.NewFirstName, args.NewLastName)
	if err != nil {
		s.logger.Error(loggerTag, fmt.Sprintf("failed update user: %v", err))

		return nil, err
	}

	return updatedUser, nil
}

func (s *ProfileService) Delete(ctx context.Context, userID, password, accessToken string) error {
	loggerTag := "profile.service.delete"

	if err := s.VerifyToken(ctx, accessToken); err != nil {
		return err
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}

		s.logger.Error(loggerTag, fmt.Sprintf("failed find user: %v", err))

		return err
	}

	if err = hash.ComparePassword(user.Password, password); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrPasswordWrong
		}

		s.logger.Error(loggerTag, fmt.Sprintf("failed compare password: %v", err))

		return err
	}

	if err = s.userRepo.Delete(ctx, userID); err != nil {
		s.logger.Error(loggerTag, fmt.Sprintf("failed delete user: %v", err))

		return err
	}

	if err = s.tokenAdapter.Del(ctx, userID); err != nil {
		s.logger.Error(loggerTag, fmt.Sprintf("failed delete refresh token to redis: %v", err))

		return err
	}

	return nil
}
