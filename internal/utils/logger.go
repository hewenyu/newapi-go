package utils

import (
	"context"
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger 日志器接口
type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	With(fields ...zap.Field) Logger
	WithContext(ctx context.Context) Logger
	Sync() error
}

// logger 实现Logger接口
type logger struct {
	zap *zap.Logger
}

// 全局日志器
var (
	globalLogger Logger
	once         sync.Once
)

// LogLevel 日志级别
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

// LogConfig 日志配置
type LogConfig struct {
	Level       LogLevel
	Development bool
	OutputPaths []string
	Encoding    string
}

// DefaultLogConfig 默认日志配置
func DefaultLogConfig() *LogConfig {
	return &LogConfig{
		Level:       InfoLevel,
		Development: false,
		OutputPaths: []string{"stdout"},
		Encoding:    "json",
	}
}

// NewLogger 创建新的日志器
func NewLogger(config *LogConfig) (Logger, error) {
	if config == nil {
		config = DefaultLogConfig()
	}

	zapConfig := zap.Config{
		Level:             zap.NewAtomicLevelAt(toZapLevel(config.Level)),
		Development:       config.Development,
		DisableCaller:     false,
		DisableStacktrace: false,
		Sampling:          nil,
		Encoding:          config.Encoding,
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "message",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      config.OutputPaths,
		ErrorOutputPaths: []string{"stderr"},
		InitialFields:    map[string]interface{}{"service": "newapi-go-sdk"},
	}

	zapLogger, err := zapConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return &logger{zap: zapLogger}, nil
}

// GetLogger 获取全局日志器
func GetLogger() Logger {
	once.Do(func() {
		config := DefaultLogConfig()
		// 检查环境变量
		if os.Getenv("NEWAPI_DEBUG") == "true" {
			config.Level = DebugLevel
			config.Development = true
		}

		l, err := NewLogger(config)
		if err != nil {
			// 如果创建失败，使用nop logger
			globalLogger = &logger{zap: zap.NewNop()}
		} else {
			globalLogger = l
		}
	})
	return globalLogger
}

// SetGlobalLogger 设置全局日志器
func SetGlobalLogger(l Logger) {
	globalLogger = l
}

// Debug 记录调试信息
func (l *logger) Debug(msg string, fields ...zap.Field) {
	l.zap.Debug(msg, fields...)
}

// Info 记录信息
func (l *logger) Info(msg string, fields ...zap.Field) {
	l.zap.Info(msg, fields...)
}

// Warn 记录警告
func (l *logger) Warn(msg string, fields ...zap.Field) {
	l.zap.Warn(msg, fields...)
}

// Error 记录错误
func (l *logger) Error(msg string, fields ...zap.Field) {
	l.zap.Error(msg, fields...)
}

// Fatal 记录致命错误
func (l *logger) Fatal(msg string, fields ...zap.Field) {
	l.zap.Fatal(msg, fields...)
}

// With 添加字段
func (l *logger) With(fields ...zap.Field) Logger {
	return &logger{zap: l.zap.With(fields...)}
}

// WithContext 从上下文中添加字段
func (l *logger) WithContext(ctx context.Context) Logger {
	fields := []zap.Field{}

	// 从上下文中提取请求ID
	if requestID := GetRequestID(ctx); requestID != "" {
		fields = append(fields, zap.String("request_id", requestID))
	}

	// 从上下文中提取用户ID
	if userID := GetUserID(ctx); userID != "" {
		fields = append(fields, zap.String("user_id", userID))
	}

	// 从上下文中提取跟踪ID
	if traceID := GetTraceID(ctx); traceID != "" {
		fields = append(fields, zap.String("trace_id", traceID))
	}

	if len(fields) > 0 {
		return &logger{zap: l.zap.With(fields...)}
	}

	return l
}

// Sync 同步日志
func (l *logger) Sync() error {
	return l.zap.Sync()
}

// toZapLevel 转换为zap日志级别
func toZapLevel(level LogLevel) zapcore.Level {
	switch level {
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	case FatalLevel:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// LogAPIRequest 记录API请求日志
func LogAPIRequest(ctx context.Context, method, url string, headers map[string]string, body interface{}) {
	logger := GetLogger().WithContext(ctx)

	// 过滤敏感信息
	safeHeaders := make(map[string]string)
	for k, v := range headers {
		if isSensitiveHeader(k) {
			safeHeaders[k] = "[REDACTED]"
		} else {
			safeHeaders[k] = v
		}
	}

	logger.Info("API request",
		zap.String("method", method),
		zap.String("url", url),
		zap.Any("headers", safeHeaders),
		zap.Any("body", body),
	)
}

// LogAPIResponse 记录API响应日志
func LogAPIResponse(ctx context.Context, statusCode int, headers map[string]string, body interface{}, duration int64) {
	logger := GetLogger().WithContext(ctx)

	if statusCode >= 400 {
		logger.Error("API response",
			zap.Int("status_code", statusCode),
			zap.Any("headers", headers),
			zap.Any("body", body),
			zap.Int64("duration_ms", duration),
		)
	} else {
		logger.Info("API response",
			zap.Int("status_code", statusCode),
			zap.Any("headers", headers),
			zap.Any("body", body),
			zap.Int64("duration_ms", duration),
		)
	}
}

// LogStreamEvent 记录流式事件日志
func LogStreamEvent(ctx context.Context, eventType, data string) {
	logger := GetLogger().WithContext(ctx)
	logger.Debug("Stream event",
		zap.String("event_type", eventType),
		zap.String("data", data),
	)
}

// LogError 记录错误日志
func LogError(ctx context.Context, err error, msg string, fields ...zap.Field) {
	logger := GetLogger().WithContext(ctx)
	allFields := append(fields, zap.Error(err))
	logger.Error(msg, allFields...)
}

// isSensitiveHeader 检查是否为敏感头部
func isSensitiveHeader(key string) bool {
	sensitiveHeaders := []string{
		"authorization",
		"api-key",
		"x-api-key",
		"cookie",
		"set-cookie",
		"password",
		"token",
	}

	for _, header := range sensitiveHeaders {
		if key == header {
			return true
		}
	}
	return false
}
