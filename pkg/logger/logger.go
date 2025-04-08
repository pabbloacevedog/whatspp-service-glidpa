package logger

import (
	"github.com/pabbloacevedog/whatspp-service-glidpa/pkg/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is the interface for logging
type Logger interface {
	Debug(msg string, fields ...zapcore.Field)
	Info(msg string, fields ...zapcore.Field)
	Warn(msg string, fields ...zapcore.Field)
	Error(msg string, fields ...zapcore.Field)
	Fatal(msg string, fields ...zapcore.Field)
	With(fields ...zapcore.Field) Logger
}

// ZapLogger implements the Logger interface using zap
type ZapLogger struct {
	logger *zap.Logger
}

// New creates a new logger based on the application configuration
func New(cfg *config.Config) (Logger, error) {
	var zapCfg zap.Config
	var appEnv, logLevel string

	// Handle nil config by setting default values
	if cfg == nil {
		// Default to development environment and debug log level
		appEnv = "development"
		logLevel = "debug"
	} else {
		appEnv = cfg.AppEnv
		logLevel = cfg.LogLevel
	}

	// Configure logger based on environment
	if appEnv == "production" {
		// JSON logger for production
		zapCfg = zap.NewProductionConfig()
	} else {
		// Console logger for development
		zapCfg = zap.NewDevelopmentConfig()
	}

	// Set log level from configuration
	var level zapcore.Level
	switch logLevel {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}
	zapCfg.Level = zap.NewAtomicLevelAt(level)

	// Build the logger
	logger, err := zapCfg.Build()
	if err != nil {
		return nil, err
	}

	return &ZapLogger{logger: logger}, nil
}

// Debug logs a debug message
func (l *ZapLogger) Debug(msg string, fields ...zapcore.Field) {
	l.logger.Debug(msg, fields...)
}

// Info logs an info message
func (l *ZapLogger) Info(msg string, fields ...zapcore.Field) {
	l.logger.Info(msg, fields...)
}

// Warn logs a warning message
func (l *ZapLogger) Warn(msg string, fields ...zapcore.Field) {
	l.logger.Warn(msg, fields...)
}

// Error logs an error message
func (l *ZapLogger) Error(msg string, fields ...zapcore.Field) {
	l.logger.Error(msg, fields...)
}

// Fatal logs a fatal message and then exits
func (l *ZapLogger) Fatal(msg string, fields ...zapcore.Field) {
	l.logger.Fatal(msg, fields...)
}

// With returns a logger with the given fields
func (l *ZapLogger) With(fields ...zapcore.Field) Logger {
	return &ZapLogger{logger: l.logger.With(fields...)}
}
