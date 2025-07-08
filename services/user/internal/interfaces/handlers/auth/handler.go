package handlers

import (
	"context"
	"errors"
	"fmt"

	"github.com/BlazeCoder04/online_store/libs/logger"
	"github.com/BlazeCoder04/online_store/libs/validate"
	"github.com/BlazeCoder04/online_store/services/user/configs"
	domain "github.com/BlazeCoder04/online_store/services/user/internal/domain/ports/services"
	converter "github.com/BlazeCoder04/online_store/services/user/internal/interfaces/converter/auth"
	desc "github.com/BlazeCoder04/online_store/services/user/pkg/auth/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type AuthHandler struct {
	desc.UnimplementedAuthV1Server
	authService domain.AuthService
	logger      logger.Logger
	cfg         *configs.Config
}

func NewAuthHandler(authService domain.AuthService, logger logger.Logger, cfg *configs.Config) (*AuthHandler, error) {
	loggerTag := "auth.handler.newAuthHandler"

	logger.Info(loggerTag, "Initializing auth handler")
	logger.Info(loggerTag, "Successful initialization")

	return &AuthHandler{
		authService: authService,
		logger:      logger,
		cfg:         cfg,
	}, nil
}

func (h *AuthHandler) Login(ctx context.Context, req *desc.LoginRequest) (*desc.LoginResponse, error) {
	loggerTag := "auth.handler.login"

	if err := validate.ValidateRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	user, accessToken, refreshToken, err := h.authService.Login(ctx, req.Email, req.Password)
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

	if err := grpc.SendHeader(ctx, metadata.Pairs(
		"access_token", accessToken,
		"refresh_token", refreshToken,
	)); err != nil {
		h.logger.Error(loggerTag, fmt.Sprintf("failed send header: %v", err))

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &desc.LoginResponse{
		Data:        converter.UserToDesc(user),
		AccessToken: accessToken,
	}, nil
}

func (h *AuthHandler) Register(ctx context.Context, req *desc.RegisterRequest) (*desc.RegisterResponse, error) {
	loggerTag := "auth.handler.register"

	if err := validate.ValidateRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	user, accessToken, refreshToken, err := h.authService.Register(ctx, req.Email, req.Password, req.FirstName, req.LastName)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserExists):
			return nil, status.Error(codes.AlreadyExists, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	if err := grpc.SendHeader(ctx, metadata.Pairs(
		"access_token", accessToken,
		"refresh_token", refreshToken,
	)); err != nil {
		h.logger.Error(loggerTag, fmt.Sprintf("failed send header: %v", err))

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &desc.RegisterResponse{
		Data:        converter.UserToDesc(user),
		AccessToken: accessToken,
	}, nil
}

func (h *AuthHandler) RefreshToken(ctx context.Context, req *desc.RefreshTokenRequest) (*desc.RefreshTokenResponse, error) {
	loggerTag := "auth.handler.refreshToken"

	if err := validate.ValidateRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	accessToken, err := h.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		case errors.Is(err, ErrTokenInvalid):
			return nil, status.Error(codes.Unauthenticated, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	if err := grpc.SendHeader(ctx, metadata.Pairs(
		"access_token", accessToken,
	)); err != nil {
		h.logger.Error(loggerTag, fmt.Sprintf("failed send header: %v", err))

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &desc.RefreshTokenResponse{
		AccessToken: accessToken,
	}, nil
}

func (h *AuthHandler) Logout(ctx context.Context, req *desc.LogoutRequest) (*emptypb.Empty, error) {
	if err := validate.ValidateRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err := h.authService.Logout(ctx, req.AccessToken)
	if err != nil {
		switch {
		case errors.Is(err, ErrTokenInvalid):
			return nil, status.Error(codes.Unauthenticated, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &emptypb.Empty{}, nil
}
