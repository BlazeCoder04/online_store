package handlers

import (
	"context"
	"errors"

	"github.com/BlazeCoder04/online_store/libs/logger"
	"github.com/BlazeCoder04/online_store/libs/validate"
	domain "github.com/BlazeCoder04/online_store/services/user/internal/domain/ports/services"
	desc "github.com/BlazeCoder04/online_store/services/user/pkg/profile/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ProfileHandler struct {
	desc.UnimplementedProfileV1Server
	profileService domain.ProfileService
	logger         logger.Logger
}

func NewProfileHandler(profileService domain.ProfileService, logger logger.Logger) (*ProfileHandler, error) {
	loggerTag := "profile.handler.newAuthHandler"

	logger.Info(loggerTag, "Initializing profile handler")
	logger.Info(loggerTag, "Successful initialization")

	return &ProfileHandler{
		profileService: profileService,
		logger:         logger,
	}, nil
}

func (h *ProfileHandler) Get(ctx context.Context, req *desc.GetRequest) (*desc.GetResponse, error) {
	if err := validate.ValidateRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	user, err := h.profileService.Get(ctx, req.UserId)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		case errors.Is(err, ErrPasswordWrong):
			return nil, status.Error(codes.Unauthenticated, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &desc.GetResponse{
		Data: userToDesc(user),
	}, nil
}

func (h *ProfileHandler) Update(ctx context.Context, req *desc.UpdateRequest) (*desc.UpdateResponse, error) {
	if err := validate.ValidateRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	updatedUser, err := h.profileService.Update(ctx, &domain.UpdateProfileArgs{
		UserID:       req.UserId,
		Password:     req.Password,
		NewEmail:     req.NewEmail,
		NewPassword:  req.NewPassword,
		NewFirstName: req.NewFirstName,
		NewLastName:  req.NewLastName,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		case errors.Is(err, ErrPasswordWrong):
			return nil, status.Error(codes.Unauthenticated, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &desc.UpdateResponse{
		Data: userToDesc(updatedUser),
	}, nil
}

func (h *ProfileHandler) Delete(ctx context.Context, req *desc.DeleteRequest) (*emptypb.Empty, error) {
	if err := validate.ValidateRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := h.profileService.Delete(ctx, req.UserId, req.Password); err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		case errors.Is(err, ErrPasswordWrong):
			return nil, status.Error(codes.Unauthenticated, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &emptypb.Empty{}, nil
}
