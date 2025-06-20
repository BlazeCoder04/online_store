package main

import (
	"log"

	"github.com/BlazeCoder04/online_store/libs/logger"
	"github.com/BlazeCoder04/online_store/services/user/configs"
	"github.com/BlazeCoder04/online_store/services/user/internal/app"
)

func main() {
	cfg, err := configs.Load(".")
	if err != nil {
		log.Fatalf("Error during initialization of config: %v", err)
	}

	logger, err := logger.NewAdapter(&logger.Config{
		Level: logger.LevelDebug - 1,
	})
	if err != nil {
		log.Fatalf("Error during initialization of logger: %v", err)
	}

	appLogger := logger.WithLayer("APP")
	app, err := app.NewApplication(appLogger, &cfg)
	if err != nil {
		appLogger.Fatal("Error during initialization application: " + err.Error())
	}

	if err := app.Run(); err != nil {
		appLogger.Fatal(err.Error())
	}
}
