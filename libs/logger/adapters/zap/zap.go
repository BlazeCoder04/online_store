package adapters

import (
	"errors"
	"fmt"
	"sync"

	"github.com/BlazeCoder04/online_store/libs/logger/domain"
	"github.com/BlazeCoder04/online_store/libs/logger/pkg/colorise"
	"github.com/BlazeCoder04/online_store/libs/logger/pkg/formatter"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapAdapter struct {
	logger    *zap.Logger
	fields    []zap.Field
	formatter *formatter.Formatter
	mu        sync.RWMutex
}

func toRouterFields(fields []domain.Field) []zap.Field {
	converted := []zap.Field{}
	for _, value := range fields {
		field := zap.Field{}
		switch value.Value.(type) {
		case string:
			field = zap.String(value.Key, value.Value.(string))
		case int:
			field = zap.Int(value.Key, value.Value.(int))
		default:
			field = zap.Any(value.Key, value.Value)
		}
		converted = append(converted, field)
	}

	return converted
}

func (z *ZapAdapter) log(level zapcore.Level, msg string, fields []domain.Field, color colorise.Color) {
	formattedMsg := z.formatter.FormatMessage(msg)
	formattedMsg = colorise.ColorString(formattedMsg, color)

	zapFields := toRouterFields(fields)

	z.mu.RLock()
	allFields := append(zapFields, z.fields...)
	z.mu.RUnlock()

	switch level {
	case zap.DebugLevel:
		z.logger.Debug(formattedMsg, allFields...)
	case zap.InfoLevel:
		z.logger.Info(formattedMsg, allFields...)
	case zap.WarnLevel:
		z.logger.Info(formattedMsg, allFields...)
	case zap.ErrorLevel:
		z.logger.Error(formattedMsg, allFields...)
	case zap.FatalLevel:
		z.logger.Fatal(formattedMsg, allFields...)
	}
}

func (z *ZapAdapter) Debug(tag, msg string, fields ...domain.Field) {
	z.log(zap.DebugLevel, fmt.Sprintf("[%s] %s", tag, msg), fields, colorise.ColorReset)
}

func (z *ZapAdapter) Info(tag, msg string, fields ...domain.Field) {
	z.log(zap.InfoLevel, fmt.Sprintf("[%s] %s", tag, msg), fields, colorise.ColorGreen)
}

func (z *ZapAdapter) Warn(tag, msg string, fields ...domain.Field) {
	z.log(zap.WarnLevel, fmt.Sprintf("[%s] %s", tag, msg), fields, colorise.ColorYellow)
}

func (z *ZapAdapter) Error(tag, msg string, fields ...domain.Field) {
	z.log(zap.ErrorLevel, fmt.Sprintf("[%s] %s", tag, msg), fields, colorise.ColorOrange)
}

func (z *ZapAdapter) Fatal(tag, msg string, fields ...domain.Field) {
	z.log(zap.FatalLevel, fmt.Sprintf("[%s] %s", tag, msg), fields, colorise.ColorRed)
}

func (z *ZapAdapter) WithFields(fields ...domain.Field) domain.Logger {
	zapFields := toRouterFields(fields)

	z.mu.RLock()
	defer z.mu.RUnlock()

	return &ZapAdapter{
		logger:    z.logger,
		fields:    append(zapFields, z.fields...),
		formatter: z.formatter,
	}
}

func toRouterLevel(level domain.Level) zap.AtomicLevel {
	switch level {
	case domain.LevelDebug:
		return zap.NewAtomicLevelAt(zap.DebugLevel)
	case domain.LevelInfo:
		return zap.NewAtomicLevelAt(zap.InfoLevel)
	case domain.LevelWarn:
		return zap.NewAtomicLevelAt(zap.WarnLevel)
	case domain.LevelError:
		return zap.NewAtomicLevelAt(zap.ErrorLevel)
	case domain.LevelFatal:
		return zap.NewAtomicLevelAt(zap.FatalLevel)
	default:
		return zap.NewAtomicLevelAt(zap.DebugLevel)
	}
}

func NewAdapter(level domain.Level) (domain.Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.Encoding = "console"
	cfg.Level = toRouterLevel(level)
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.CallerKey = ""

	logger, err := cfg.Build()
	if err != nil {
		return nil, errors.New(ErrZapBuild)
	}

	return &ZapAdapter{
		logger:    logger,
		fields:    make([]zap.Field, 0),
		formatter: formatter.NewFormatter(""),
	}, nil
}
