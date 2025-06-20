package handlers

import (
	"context"

	"github.com/BlazeCoder04/online_store/libs/logger"
	"github.com/BlazeCoder04/online_store/libs/validate"
	pb "github.com/BlazeCoder04/online_store/protobuf/gen/go/services/user/auth"
	"github.com/BlazeCoder04/online_store/services/user/configs"
	domain "github.com/BlazeCoder04/online_store/services/user/internal/domain/ports/auth/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AuthHandler struct {
	pb.UnimplementedAuthServer
	authService domain.AuthService
	logger      logger.Logger
	cfg         *configs.Config
}

func NewAuthHandler(authService domain.AuthService, logger logger.Logger, cfg *configs.Config) (*AuthHandler, error) {
	logger.Info("Initializing auth handler")
	logger.OK("Successful initialization")

	return &AuthHandler{
		authService: authService,
		logger:      logger,
		cfg:         cfg,
	}, nil
}

func (h *AuthHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if err := validate.ValidateRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	user, accessToken, refreshToken, err := h.authService.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	md := metadata.Pairs(
		"access_token", accessToken,
		"refresh_token", refreshToken,
	)

	ctx = metadata.NewOutgoingContext(ctx, md)

	if err = grpc.SendHeader(ctx, md); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.LoginResponse{
		Status: "success",
		Data: &pb.User{
			Id:        user.ID.String(),
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Role:      pb.UserRole(pb.UserRole_value[string(user.Role)]),
			CreatedAt: timestamppb.New(user.CreatedAt),
			UpdatedAt: timestamppb.New(user.UpdatedAt),
		},
		AccessToken: accessToken,
	}, nil
}

func (h *AuthHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if err := validate.ValidateRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	user, accessToken, refreshToken, err := h.authService.Register(ctx, req.Email, req.Password, req.FirstName, req.LastName)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	md := metadata.Pairs(
		"access_token", accessToken,
		"refresh_token", refreshToken,
	)

	ctx = metadata.NewOutgoingContext(ctx, md)

	if err = grpc.SendHeader(ctx, md); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.RegisterResponse{
		Status: "success",
		Data: &pb.User{
			Id:        user.ID.String(),
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Role:      pb.UserRole(pb.UserRole_value[string(user.Role)]),
			CreatedAt: timestamppb.New(user.CreatedAt),
			UpdatedAt: timestamppb.New(user.UpdatedAt),
		},
		AccessToken: accessToken,
	}, nil
}

func (h *AuthHandler) UpdateToken(ctx context.Context, _ *emptypb.Empty) (*pb.UpdateTokenResponse, error) {
	md, _ := metadata.FromIncomingContext(ctx)

	token := md.Get("refresh_token")
	if len(token) == 0 {
		return nil, status.Error(codes.Unauthenticated, ErrUserUnauthorized)
	}
	refreshToken := token[0]

	token = md.Get("access_token")
	if len(token) == 0 {
		return nil, status.Error(codes.Unauthenticated, ErrUserUnauthorized)
	}
	accessToken := token[0]

	newAccessToken, err := h.authService.UpdateToken(ctx, accessToken, refreshToken)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	md = metadata.Pairs(
		"access_token", newAccessToken,
	)

	ctx = metadata.NewOutgoingContext(ctx, md)

	if err = grpc.SetHeader(ctx, md); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UpdateTokenResponse{
		Status:      "success",
		AccessToken: newAccessToken,
	}, nil
}

func (h *AuthHandler) Logout(ctx context.Context, _ *emptypb.Empty) (*pb.LogoutResponse, error) {
	md, _ := metadata.FromIncomingContext(ctx)

	token := md.Get("access_token")
	if len(token) == 0 {
		return nil, status.Error(codes.Unauthenticated, ErrUserUnauthorized)
	}
	accessToken := token[0]

	if err := h.authService.Logout(ctx, accessToken); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.LogoutResponse{
		Status: "success",
	}, nil
}
