package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/BlazeCoder04/online_store/libs/hash"
	"github.com/BlazeCoder04/online_store/libs/logger"
	"github.com/BlazeCoder04/online_store/services/user/internal/domain/models"
	domainRepo "github.com/BlazeCoder04/online_store/services/user/internal/domain/ports/repositories"
	domainService "github.com/BlazeCoder04/online_store/services/user/internal/domain/ports/services"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type ProfileService struct {
	userRepo domainRepo.UserRepository
	logger   logger.Logger
}

func NewProfileService(userRepo domainRepo.UserRepository, logger logger.Logger) (domainService.ProfileService, error) {
	loggerTag := "profile.service.newProfileService"

	logger.Info(loggerTag, "Initializing profile service")
	logger.Info(loggerTag, "Successful initialization")

	return &ProfileService{
		userRepo,
		logger,
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
