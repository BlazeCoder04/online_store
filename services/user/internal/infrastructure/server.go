package server

import (
	"fmt"
	"net"

	"github.com/BlazeCoder04/online_store/libs/logger"
	"github.com/BlazeCoder04/online_store/services/user/configs"
	domain "github.com/BlazeCoder04/online_store/services/user/internal/domain/ports"
	auth "github.com/BlazeCoder04/online_store/services/user/internal/interfaces/handlers/auth"
	profile "github.com/BlazeCoder04/online_store/services/user/internal/interfaces/handlers/profile"
	authDesc "github.com/BlazeCoder04/online_store/services/user/pkg/auth/v1"
	profileDesc "github.com/BlazeCoder04/online_store/services/user/pkg/profile/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	authHandler    *auth.AuthHandler
	profileHandler *profile.ProfileHandler
	logger         logger.Logger
	cfg            *configs.Config
}

func NewServer(authHandler *auth.AuthHandler, profileHandler *profile.ProfileHandler, logger logger.Logger, cfg *configs.Config) (domain.Server, error) {
	loggerTag := "server.newServer"

	logger.Info(loggerTag, "Initializing server")
	logger.Info(loggerTag, "Successful initialization")

	return &Server{
		authHandler,
		profileHandler,
		logger,
		cfg,
	}, nil
}

func (s *Server) Run() error {
	loggerTag := "server.run"

	s.logger.Info(loggerTag, "The server is running", logger.Field{
		Key:   "port",
		Value: s.cfg.ServerPort,
	})

	server := grpc.NewServer()

	authDesc.RegisterAuthV1Server(server, s.authHandler)
	profileDesc.RegisterProfileV1Server(server, s.profileHandler)

	reflection.Register(server)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.cfg.ServerPort))
	if err != nil {
		return fmt.Errorf("failed listen: %v", err)
	}

	return server.Serve(listener)
}
