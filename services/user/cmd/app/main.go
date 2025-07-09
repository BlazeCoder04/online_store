package main

import (
	"fmt"
	"log"

	"github.com/BlazeCoder04/online_store/libs/logger"
	"github.com/BlazeCoder04/online_store/services/user/configs"
	"github.com/BlazeCoder04/online_store/services/user/internal/app"
	"github.com/BlazeCoder04/online_store/services/user/internal/infrastructure/migrator"
)

const (
	loggerTag         = "main"
	loggerMigratorTag = "main.migrator"
)

func main() {
	cfg := configs.Load()

	logger, err := logger.NewAdapter(&logger.Config{
		Level: logger.LevelDebug - 1,
	})
	if err != nil {
		log.Fatalf("[%s] Error during initialization of logger: %v", loggerTag, err)
	}

	migrator := migrator.NewMigrator(&cfg, logger)
	if err = migrator.Up(); err != nil {
		logger.Fatal(loggerMigratorTag, err.Error())
	}

	app, err := app.NewApplication(logger, &cfg)
	if err != nil {
		logger.Fatal(loggerTag, fmt.Sprintf("Error during initializtion application: %v", err))
	}

	if err := app.Run(); err != nil {
		logger.Fatal(loggerTag, fmt.Sprintf("Error during run application: %v", err))
	}
}
