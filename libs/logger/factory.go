package logger

import (
	adapters "github.com/BlazeCoder04/online_store/libs/logger/adapters/zap"
	"github.com/BlazeCoder04/online_store/libs/logger/domain"
)

type (
	Level  = domain.Level
	Field  = domain.Field
	Logger = domain.Logger
)

const (
	LevelDebug = domain.LevelDebug
	LevelInfo  = domain.LevelInfo
	LevelWarn  = domain.LevelWarn
	LevelError = domain.LevelError
	LevelFatal = domain.LevelFatal
)

type Config struct {
	Level Level
}

func NewAdapter(config *Config) (Logger, error) {
	return adapters.NewAdapter(config.Level)
}
