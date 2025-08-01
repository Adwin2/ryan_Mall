package infrastructure

import (
	"context"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger 日志接口
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	With(fields ...Field) Logger
	WithContext(ctx context.Context) Logger
}

// Field 日志字段
type Field struct {
	Key   string
	Value interface{}
}

// String 创建字符串字段
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Int 创建整数字段
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Int64 创建64位整数字段
func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

// Float64 创建浮点数字段
func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

// Bool 创建布尔字段
func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

// Error 创建错误字段
func Error(err error) Field {
	return Field{Key: "error", Value: err}
}

// Duration 创建时长字段
func Duration(key string, val time.Duration) Field {
	return Field{Key: key, Value: val}
}

// Any 创建任意类型字段
func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// ZapLogger Zap日志实现
type ZapLogger struct {
	logger *zap.Logger
}

// NewLogger 创建日志器（NewZapLogger的别名）
func NewLogger(level string) (Logger, error) {
	return NewZapLogger(level)
}

// NewZapLogger 创建Zap日志器
func NewZapLogger(level string) (*ZapLogger, error) {
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	case "fatal":
		zapLevel = zapcore.FatalLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	config := zap.Config{
		Level:       zap.NewAtomicLevelAt(zapLevel),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding: "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "message",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	// 如果是开发环境，使用更友好的输出格式
	if os.Getenv("ENV") == "development" {
		config.Development = true
		config.Encoding = "console"
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return &ZapLogger{logger: logger}, nil
}

// fieldsToZapFields 转换字段格式
func (l *ZapLogger) fieldsToZapFields(fields []Field) []zap.Field {
	zapFields := make([]zap.Field, len(fields))
	for i, field := range fields {
		zapFields[i] = zap.Any(field.Key, field.Value)
	}
	return zapFields
}

// Debug 调试日志
func (l *ZapLogger) Debug(msg string, fields ...Field) {
	l.logger.Debug(msg, l.fieldsToZapFields(fields)...)
}

// Info 信息日志
func (l *ZapLogger) Info(msg string, fields ...Field) {
	l.logger.Info(msg, l.fieldsToZapFields(fields)...)
}

// Warn 警告日志
func (l *ZapLogger) Warn(msg string, fields ...Field) {
	l.logger.Warn(msg, l.fieldsToZapFields(fields)...)
}

// Error 错误日志
func (l *ZapLogger) Error(msg string, fields ...Field) {
	l.logger.Error(msg, l.fieldsToZapFields(fields)...)
}

// Fatal 致命错误日志
func (l *ZapLogger) Fatal(msg string, fields ...Field) {
	l.logger.Fatal(msg, l.fieldsToZapFields(fields)...)
}

// With 添加字段
func (l *ZapLogger) With(fields ...Field) Logger {
	return &ZapLogger{
		logger: l.logger.With(l.fieldsToZapFields(fields)...),
	}
}

// WithContext 添加上下文
func (l *ZapLogger) WithContext(ctx context.Context) Logger {
	// 从上下文中提取追踪信息
	fields := []Field{}
	
	if traceID := getTraceIDFromContext(ctx); traceID != "" {
		fields = append(fields, String("trace_id", traceID))
	}
	
	if spanID := getSpanIDFromContext(ctx); spanID != "" {
		fields = append(fields, String("span_id", spanID))
	}
	
	if userID := getUserIDFromContext(ctx); userID != "" {
		fields = append(fields, String("user_id", userID))
	}
	
	if requestID := getRequestIDFromContext(ctx); requestID != "" {
		fields = append(fields, String("request_id", requestID))
	}
	
	return l.With(fields...)
}

// Sync 同步日志
func (l *ZapLogger) Sync() error {
	return l.logger.Sync()
}

// 上下文键类型
type contextKey string

const (
	traceIDKey   contextKey = "trace_id"
	spanIDKey    contextKey = "span_id"
	userIDKey    contextKey = "user_id"
	requestIDKey contextKey = "request_id"
)

// getTraceIDFromContext 从上下文获取追踪ID
func getTraceIDFromContext(ctx context.Context) string {
	if traceID, ok := ctx.Value(traceIDKey).(string); ok {
		return traceID
	}
	return ""
}

// getSpanIDFromContext 从上下文获取Span ID
func getSpanIDFromContext(ctx context.Context) string {
	if spanID, ok := ctx.Value(spanIDKey).(string); ok {
		return spanID
	}
	return ""
}

// getUserIDFromContext 从上下文获取用户ID
func getUserIDFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value(userIDKey).(string); ok {
		return userID
	}
	return ""
}

// getRequestIDFromContext 从上下文获取请求ID
func getRequestIDFromContext(ctx context.Context) string {
	if requestID, ok := ctx.Value(requestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// WithTraceID 添加追踪ID到上下文
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// WithSpanID 添加Span ID到上下文
func WithSpanID(ctx context.Context, spanID string) context.Context {
	return context.WithValue(ctx, spanIDKey, spanID)
}

// WithUserID 添加用户ID到上下文
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// WithRequestID 添加请求ID到上下文
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// 全局日志器实例
var globalLogger Logger

// InitGlobalLogger 初始化全局日志器
func InitGlobalLogger(level string) error {
	logger, err := NewZapLogger(level)
	if err != nil {
		return err
	}
	globalLogger = logger
	return nil
}

// GetLogger 获取全局日志器
func GetLogger() Logger {
	if globalLogger == nil {
		// 如果没有初始化，使用默认配置
		logger, _ := NewZapLogger("info")
		globalLogger = logger
	}
	return globalLogger
}
