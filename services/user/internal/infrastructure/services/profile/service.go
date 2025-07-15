package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/BlazeCoder04/online_store/libs/hash"
	"github.com/BlazeCoder04/online_store/libs/logger"
	"github.com/BlazeCoder04/online_store/services/user/configs"
	"github.com/BlazeCoder04/online_store/services/user/internal/domain/models"
	domainRepo "github.com/BlazeCoder04/online_store/services/user/internal/domain/ports/repositories"
	domainService "github.com/BlazeCoder04/online_store/services/user/internal/domain/ports/services"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type ProfileService struct {
	userRepo domainRepo.UserRepository
	logger   logger.Logger
	cfg      *configs.Config
}

func NewProfileService(userRepo domainRepo.UserRepository, logger logger.Logger, cfg *configs.Config) (domainService.ProfileService, error) {
	loggerTag := "profile.service.newProfileService"

	logger.Info(loggerTag, "Initializing profile service")
	logger.Info(loggerTag, "Successful initialization")

	return &ProfileService{
		userRepo,
		logger,
		cfg,
	}, nil
}

func (s *ProfileService) Get(ctx context.Context, userID string) (*models.User, error) {
	loggerTag := "profile.service.get"

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

	var hashedPassword *string
	if args.NewPassword != nil {
		hashed, hashErr := hash.HashPassword(*args.NewPassword)
		if hashErr != nil {
			s.logger.Error(loggerTag, fmt.Sprintf("failed hash password: %v", err))

			return nil, err
		}

		hashedPassword = &hashed
	}

	newUser, err := s.userRepo.Update(ctx, args.UserID, args.NewEmail, hashedPassword, args.NewFirstName, args.NewLastName)
	if err != nil {
		s.logger.Error(loggerTag, fmt.Sprintf("failed update user: %v", err))

		return nil, err
	}

	return newUser, nil
}

func (s *ProfileService) Delete(ctx context.Context, userID, password string) error {
	loggerTag := "profile.service.delete"

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

	return nil
}
