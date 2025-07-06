package migrator

import (
	"errors"
	"fmt"

	"github.com/BlazeCoder04/online_store/libs/logger"
	"github.com/BlazeCoder04/online_store/services/user/configs"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	loggerTag      = "migrator"
	migrationsPath = "migrations"
)

type Migrator struct {
	cfg    *configs.Config
	logger logger.Logger
}

func NewMigrator(cfg *configs.Config, logger logger.Logger) *Migrator {
	return &Migrator{
		cfg,
		logger,
	}
}

func (m *Migrator) Up() error {
	migration, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		m.cfg.PostgresMigrationDSN,
	)
	if err != nil {
		return fmt.Errorf("Error during initialization of migration: %v", err)
	}
	defer migration.Close()

	if err := migration.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			m.logger.Info(loggerTag, "No migrations to apply")
		}

		return fmt.Errorf("Migration error: %v", err)
	}

	m.logger.Info(loggerTag, "Migrations completed successfully")

	return nil
}
