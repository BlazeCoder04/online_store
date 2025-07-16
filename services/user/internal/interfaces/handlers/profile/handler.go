package handlers

import (
	"context"
	"errors"
	"strings"

	"github.com/BlazeCoder04/online_store/libs/logger"
	"github.com/BlazeCoder04/online_store/libs/validate"
	domain "github.com/BlazeCoder04/online_store/services/user/internal/domain/ports/services"
	"github.com/BlazeCoder04/online_store/services/user/internal/interfaces/converters"
	desc "github.com/BlazeCoder04/online_store/services/user/pkg/profile/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ProfileHandler struct {
	desc.UnimplementedProfileV1Server
	profileService domain.ProfileService
	logger         logger.Logger
}

const tokenPrefix = "Bearer "

func NewProfileHandler(profileService domain.ProfileService, logger logger.Logger) (*ProfileHandler, error) {
	loggerTag := "profile.handler.newAuthHandler"

	logger.Info(loggerTag, "Profile handler initialized")

	return &ProfileHandler{
		profileService: profileService,
		logger:         logger,
	}, nil
}

func (h *ProfileHandler) GetAccessToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", ErrMetadataNotProvided
	}

	authHeader, ok := md["authorization"]
	if !ok || len(authHeader) == 0 {
		return "", ErrHeaderNotProvided
	}

	if !strings.HasPrefix(authHeader[0], tokenPrefix) {
		return "", ErrTokenInvalid
	}

	accessToken := strings.TrimPrefix(authHeader[0], tokenPrefix)

	return accessToken, nil
}

func (h *ProfileHandler) Get(ctx context.Context, req *desc.GetRequest) (*desc.GetResponse, error) {
	if err := validate.ValidateRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	accessToken, err := h.GetAccessToken(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	user, err := h.profileService.Get(ctx, req.UserId, accessToken)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		case errors.Is(err, ErrPasswordWrong):
			return nil, status.Error(codes.Unauthenticated, err.Error())
		case errors.Is(err, ErrTokenInvalid):
			return nil, status.Error(codes.Unauthenticated, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &desc.GetResponse{
		Data: converters.UserToDesc(user),
	}, nil
}

func (h *ProfileHandler) Update(ctx context.Context, req *desc.UpdateRequest) (*desc.UpdateResponse, error) {
	if err := validate.ValidateRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	accessToken, err := h.GetAccessToken(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	updatedUser, err := h.profileService.Update(ctx, &domain.UpdateProfileArgs{
		UserID:       req.UserId,
		Password:     req.Password,
		NewEmail:     req.NewEmail,
		NewPassword:  req.NewPassword,
		NewFirstName: req.NewFirstName,
		NewLastName:  req.NewLastName,
		AccessToken:  accessToken,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		case errors.Is(err, ErrPasswordWrong):
			return nil, status.Error(codes.Unauthenticated, err.Error())
		case errors.Is(err, ErrTokenInvalid):
			return nil, status.Error(codes.Unauthenticated, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &desc.UpdateResponse{
		Data: converters.UserToDesc(updatedUser),
	}, nil
}

func (h *ProfileHandler) Delete(ctx context.Context, req *desc.DeleteRequest) (*emptypb.Empty, error) {
	if err := validate.ValidateRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	accessToken, err := h.GetAccessToken(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	if err := h.profileService.Delete(ctx, req.UserId, req.Password, accessToken); err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		case errors.Is(err, ErrPasswordWrong):
			return nil, status.Error(codes.Unauthenticated, err.Error())
		case errors.Is(err, ErrTokenInvalid):
			return nil, status.Error(codes.Unauthenticated, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &emptypb.Empty{}, nil
}
