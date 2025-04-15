package config

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"strings"
)

func NewLogger(cfg Config) *zap.Logger {
	var level zapcore.Level

	switch strings.ToLower(cfg.LogLevel) {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	default:
		level = zap.InfoLevel
	}

	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Development:      true,
		Encoding:         "console",
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:      "time",
			LevelKey:     "level",
			MessageKey:   "msg",
			CallerKey:    "caller",
			EncodeLevel:  zapcore.CapitalColorLevelEncoder,
			EncodeTime:   zapcore.ISO8601TimeEncoder,
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
	}

	logger, err := config.Build()
	if err != nil {
		panic("не удалось инициализировать логгер: " + err.Error())
	}
	return logger
}
