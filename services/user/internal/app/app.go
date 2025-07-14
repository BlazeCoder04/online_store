package app

import (
	"fmt"

	"github.com/BlazeCoder04/online_store/libs/logger"
	"github.com/BlazeCoder04/online_store/services/user/configs"
	domain "github.com/BlazeCoder04/online_store/services/user/internal/domain/ports"
	server "github.com/BlazeCoder04/online_store/services/user/internal/infrastructure"
	tokenAdapter "github.com/BlazeCoder04/online_store/services/user/internal/infrastructure/adapters/cache/redis/token"
	userRepo "github.com/BlazeCoder04/online_store/services/user/internal/infrastructure/repositories/user"
	authService "github.com/BlazeCoder04/online_store/services/user/internal/infrastructure/services/auth"
	authHandler "github.com/BlazeCoder04/online_store/services/user/internal/interfaces/handlers/auth"
)

type Application struct {
	server domain.Server
	logger logger.Logger
	cfg    *configs.Config
}

func NewApplication(logger logger.Logger, cfg *configs.Config) (domain.Application, error) {
	loggerTag := "application.newApplication"

	logger.Info(loggerTag, "Initializing application")

	userRepository, err := userRepo.NewUserRepository(logger, cfg)
	if err != nil {
		return nil, fmt.Errorf("error initializing user repository: %v", err)
	}

	tokenAdapter, err := tokenAdapter.NewTokenAdapter(logger, cfg)
	if err != nil {
		return nil, fmt.Errorf("error initializing token repository: %v", err)
	}

	authService, err := authService.NewAuthService(userRepository, tokenAdapter, logger, cfg)
	if err != nil {
		return nil, fmt.Errorf("error initializing auth service: %v", err)
	}

	authHandler, err := authHandler.NewAuthHandler(authService, logger, cfg)
	if err != nil {
		return nil, fmt.Errorf("error initializing auth handler: %v", err)
	}

	server, err := server.NewServer(authHandler, logger, cfg)
	if err != nil {
		return nil, fmt.Errorf("error initializing server: %v", err)
	}

	logger.Info(loggerTag, "Application initialized successfully")

	return &Application{
		server,
		logger,
		cfg,
	}, nil
}

func (a *Application) Run() error {
	loggerTag := "application.run"

	a.logger.Info(loggerTag, "Running the application")

	return a.server.Run()
}
