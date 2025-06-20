package app

import (
	"fmt"

	"github.com/BlazeCoder04/online_store/libs/logger"
	"github.com/BlazeCoder04/online_store/services/user/configs"
	"github.com/BlazeCoder04/online_store/services/user/internal/domain"
	server "github.com/BlazeCoder04/online_store/services/user/internal/infrastructure/auth"
	authRepo "github.com/BlazeCoder04/online_store/services/user/internal/infrastructure/auth/repository/auth"
	tokenRepo "github.com/BlazeCoder04/online_store/services/user/internal/infrastructure/auth/repository/token"
	"github.com/BlazeCoder04/online_store/services/user/internal/infrastructure/auth/service"
	"github.com/BlazeCoder04/online_store/services/user/internal/interfaces/http/auth/handlers"
)

type Application struct {
	server domain.Server
	logger logger.Logger
	cfg    *configs.Config
}

func NewApplication(logger logger.Logger, cfg *configs.Config) (domain.Application, error) {
	logger.Info("Initializing application")

	authRepoLogger := logger.WithLayer("AUTH_REPO")
	authRepository, err := authRepo.NewAuthRepository(authRepoLogger, cfg)
	if err != nil {
		return nil, fmt.Errorf("error initializing auth repository: %v", err)
	}

	tokenRepoLogger := logger.WithLayer("TOKEN_REPO")
	tokenRepository, err := tokenRepo.NewTokenRepository(tokenRepoLogger, cfg)
	if err != nil {
		return nil, fmt.Errorf("error initializing token repository: %v", err)
	}

	authServiceLogger := logger.WithLayer("AUTH_SERVICE")
	authService, err := service.NewAuthService(authRepository, tokenRepository, authServiceLogger, cfg)
	if err != nil {
		return nil, fmt.Errorf("error initializing auth service: %v", err)
	}

	authHandlerLogger := logger.WithLayer("AUTH_HANDLER")
	authHandler, err := handlers.NewAuthHandler(authService, authHandlerLogger, cfg)
	if err != nil {
		return nil, fmt.Errorf("error initializing auth handler: %v", err)
	}

	authServerLogger := logger.WithLayer("AUTH_SERVER")
	authServer, err := server.NewAuthServer(authHandler, authServerLogger, cfg)
	if err != nil {
		return nil, fmt.Errorf("error initializing auth server: %v", err)
	}

	logger.OK("Application initialized successfully")

	return &Application{
		authServer,
		logger,
		cfg,
	}, nil
}

func (a *Application) Run() error {
	a.logger.Info("Running the application")

	return a.server.Run()
}
