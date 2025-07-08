package main

import (
	"log"

	"github.com/BlazeCoder04/online_store/libs/logger"
	"github.com/BlazeCoder04/online_store/services/user/configs"
	"github.com/BlazeCoder04/online_store/services/user/internal/infrastructure/migrator"
)

const (
	loggerTag         = "app"
	loggerMigratorTag = "migrator"
)

func main() {
	cfg, err := configs.Load()
	if err != nil {
		log.Fatalf("[%s] Error during initialization of config: %v", loggerTag, err)
	}

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
}
