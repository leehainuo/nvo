package log

import (
	"fmt"

	"nvo-api/core/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Init 初始化全局logger
func Init(c config.LogConfig) error {
	logger, err := buildLogger(c)
	if err != nil {
		return err
	}
	zap.ReplaceGlobals(logger)
	return nil
}

// buildLogger 构建logger实例
func buildLogger(c config.LogConfig) (*zap.Logger, error) {
	var level zapcore.Level
	switch c.Level {
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

	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Development:      c.Level == "debug",
		Encoding:         c.Format,
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{c.OutputPath},
		ErrorOutputPaths: []string{"stderr"},
	}

	if c.Format == "console" {
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	logger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return logger, nil
}

// Debug 记录调试信息
func Debug(msg string, fields ...zap.Field) {
	zap.L().Debug(msg, fields...)
}

// Info 记录一般信息
func Info(msg string, fields ...zap.Field) {
	zap.L().Info(msg, fields...)
}

// Warn 记录警告信息
func Warn(msg string, fields ...zap.Field) {
	zap.L().Warn(msg, fields...)
}

// Error 记录错误信息
func Error(msg string, fields ...zap.Field) {
	zap.L().Error(msg, fields...)
}

// Fatal 记录致命错误并退出程序
func Fatal(msg string, fields ...zap.Field) {
	zap.L().Fatal(msg, fields...)
}

// Panic 记录错误并panic
func Panic(msg string, fields ...zap.Field) {
	zap.L().Panic(msg, fields...)
}

// Sync 同步日志缓冲
func Sync() error {
	return zap.L().Sync()
}
