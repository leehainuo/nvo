package log

import (
	"fmt"
	"moka/pkg/config"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var format string

// Config 日志配置
type Config struct {
	Level      string `mapstructure:"level"`       // debug, info, warn, error
	Format     string `mapstructure:"format"`      // json, console
	OutputPath string `mapstructure:"output_path"` // stdout 或文件路径
	MaxSize    int    `mapstructure:"max_size"`    // 单个文件最大 MB
	MaxBackups int    `mapstructure:"max_backups"` // 保留旧文件数量
	MaxAge     int    `mapstructure:"max_age"`     // 保留天数
	Compress   bool   `mapstructure:"compress"`    // 是否压缩
}

func Init() error {
	var c Config
	if err := config.Viper.UnmarshalKey("log", &c); err != nil {
		return fmt.Errorf("failed to unmarshal log config: %w", err)
	}

	logger, err := build(c)
	if err != nil {
		return err
	}

	zap.ReplaceGlobals(logger)

	return nil
}

func build(c Config) (*zap.Logger, error) {
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(c.Level)); err != nil {
		level = zapcore.InfoLevel
	}

	format = c.Format

	encoder := getEncoder()

	writeSyncer, err := getWriteSyncer(c)
	if err != nil {
		return nil, err
	}

	core := zapcore.NewCore(encoder, writeSyncer, level)

	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return logger, nil
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",                         // 日志中时间字段的键名
		LevelKey:       "level",                        // 日志中级别字段的键名
		NameKey:        "logger",                       // 日志中 logger 名称字段的键名
		CallerKey:      "caller",                       // 日志中调用者信息字段的键名
		FunctionKey:    zapcore.OmitKey,                // 日志中函数信息字段的键名
		MessageKey:     "msg",                          // 日志中消息字段的键名
		StacktraceKey:  "stacktrace",                   // 日志中堆栈信息字段的键名
		LineEnding:     zapcore.DefaultLineEnding,      // 每条日志结束时的换行符
		EncodeLevel:    zapcore.CapitalLevelEncoder,    // 日志级别的显示格式
		EncodeTime:     zapcore.ISO8601TimeEncoder,     // ISO8601 时间格式编码器
		EncodeDuration: zapcore.SecondsDurationEncoder, // 持续时间以秒为单位编码
		EncodeCaller:   zapcore.ShortCallerEncoder,     // 短格式调用者编码器
	}

	if format == "json" {
		return zapcore.NewJSONEncoder(encoderConfig)
	}

	encoderConfig.CallerKey   = zapcore.OmitKey
	encoderConfig.EncodeTime  = TimeEncoder
	encoderConfig.EncodeLevel = LevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getWriteSyncer(c Config) (zapcore.WriteSyncer, error) {
	// 输出到标准输出/错误
	if c.OutputPath == "stdout" {
		return zapcore.AddSync(os.Stdout), nil
	}
	if c.OutputPath == "stderr" {
		return zapcore.AddSync(os.Stderr), nil
	}

	// 输出到文件，自动创建目录
	if err := os.MkdirAll(filepath.Dir(c.OutputPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// 使用 lumberjack 实现日志轮转
	lumberjackLogger := &lumberjack.Logger{
		Filename:   c.OutputPath,
		MaxSize:    c.MaxSize,
		MaxBackups: c.MaxBackups,
		MaxAge:     c.MaxAge,
		Compress:   c.Compress,
	}

	return zapcore.AddSync(lumberjackLogger), nil
}

// customTimeEncoder 自定义时间编码器（加粗）
func TimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString("\033[1m" + t.Format("2006-01-02 15:04:05") + "\033[0m")
}

// customLevelEncoder 自定义级别编码器（彩色加粗）
func LevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	var levelStr string
	switch level {
	case zapcore.DebugLevel:
		levelStr = "\033[1;35mDEBUG\033[0m"  // 品红加粗
	case zapcore.InfoLevel:
		levelStr = "\033[1;34mINFO\033[0m"   // 蓝色加粗
	case zapcore.WarnLevel:
		levelStr = "\033[1;33mWARN\033[0m"   // 黄色加粗
	case zapcore.ErrorLevel:
		levelStr = "\033[1;31mERROR\033[0m"  // 红色加粗
	case zapcore.DPanicLevel:
		levelStr = "\033[1;31mDPANIC\033[0m" // 红色加粗
	case zapcore.PanicLevel:
		levelStr = "\033[1;31mPANIC\033[0m"  // 红色加粗
	case zapcore.FatalLevel:
		levelStr = "\033[1;31mFATAL\033[0m" // 红色加粗
	default:
		levelStr = level.CapitalString()
	}
	enc.AppendString(levelStr)
}

// style 为消息添加样式（仅 console 格式）
func style(msg, ansi string) string {
	if format == "json" {
		return msg
	}
	return ansi + msg + "\033[0m"
}

// Debug 记录调试信息
func Debug(msg string, fields ...zap.Field) {
	zap.L().Debug(style(msg, "\033[1m"), fields...)
}

// Info 记录一般信息
func Info(msg string, fields ...zap.Field) {
	zap.L().Info(style(msg, "\033[1m"), fields...)
}

// Warn 记录警告信息
func Warn(msg string, fields ...zap.Field) {
	zap.L().Warn(style(msg, "\033[1m"), fields...)
}

// Error 记录错误信息
func Error(msg string, fields ...zap.Field) {
	zap.L().Error(style(msg, "\033[1;31m"), fields...)
}

// Fatal 记录致命错误并退出程序
func Fatal(msg string, fields ...zap.Field) {
	zap.L().Fatal(style(msg, "\033[1;31m"), fields...)
}

// Panic 记录错误并panic
func Panic(msg string, fields ...zap.Field) {
	zap.L().Panic(style(msg, "\033[1;31m"), fields...)
}

// Sync 同步日志缓冲
func Sync() error {
	return zap.L().Sync()
}
