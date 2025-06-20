package server

import (
	"fmt"
	"net"

	"github.com/BlazeCoder04/online_store/libs/logger"
	pb "github.com/BlazeCoder04/online_store/protobuf/gen/go/services/user/auth"
	"github.com/BlazeCoder04/online_store/services/user/configs"
	"github.com/BlazeCoder04/online_store/services/user/internal/domain"
	"github.com/BlazeCoder04/online_store/services/user/internal/interfaces/http/auth/handlers"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type AuthServer struct {
	authHandler *handlers.AuthHandler
	logger      logger.Logger
	cfg         *configs.Config
}

func NewAuthServer(authHandler *handlers.AuthHandler, logger logger.Logger, cfg *configs.Config) (domain.Server, error) {
	logger.Info("Initializing auth server")

	server := &AuthServer{
		authHandler,
		logger,
		cfg,
	}

	logger.OK("Success")

	return server, nil
}

func (s *AuthServer) Run() error {
	s.logger.Info("The server is running", logger.Field{
		Key:   "port",
		Value: s.cfg.ServerPort,
	})

	server := grpc.NewServer()

	pb.RegisterAuthServer(server, s.authHandler)

	reflection.Register(server)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.cfg.ServerPort))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	return server.Serve(listener)
}
