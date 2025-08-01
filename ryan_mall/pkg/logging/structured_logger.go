package logging

import (
	"context"
	"fmt"
	"os"
	"time"

	"ryan-mall-microservices/pkg/tracing"

	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// StructuredLogger 结构化日志器
type StructuredLogger struct {
	logger *zap.Logger
	sugar  *zap.SugaredLogger
}

// LogConfig 日志配置
type LogConfig struct {
	Level       string `json:"level"`        // 日志级别
	Format      string `json:"format"`       // 日志格式 (json/console)
	Output      string `json:"output"`       // 输出目标 (stdout/file)
	Filename    string `json:"filename"`     // 文件名
	MaxSize     int    `json:"max_size"`     // 最大文件大小(MB)
	MaxBackups  int    `json:"max_backups"`  // 最大备份数
	MaxAge      int    `json:"max_age"`      // 最大保存天数
	Compress    bool   `json:"compress"`     // 是否压缩
	ServiceName string `json:"service_name"` // 服务名称
	Environment string `json:"environment"`  // 环境
}

// DefaultLogConfig 默认日志配置
func DefaultLogConfig(serviceName string) *LogConfig {
	return &LogConfig{
		Level:       "info",
		Format:      "json",
		Output:      "stdout",
		ServiceName: serviceName,
		Environment: "development",
	}
}

// NewStructuredLogger 创建结构化日志器
func NewStructuredLogger(config *LogConfig) (*StructuredLogger, error) {
	// 解析日志级别
	level, err := zapcore.ParseLevel(config.Level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}

	// 配置编码器
	var encoderConfig zapcore.EncoderConfig
	if config.Format == "json" {
		encoderConfig = zap.NewProductionEncoderConfig()
	} else {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// 自定义时间格式
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// 创建编码器
	var encoder zapcore.Encoder
	if config.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 配置输出
	var writeSyncer zapcore.WriteSyncer
	if config.Output == "stdout" {
		writeSyncer = zapcore.AddSync(os.Stdout)
	} else {
		// 文件输出配置（这里简化处理，实际项目中可以使用lumberjack）
		file, err := os.OpenFile(config.Filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		writeSyncer = zapcore.AddSync(file)
	}

	// 创建核心
	core := zapcore.NewCore(encoder, writeSyncer, level)

	// 创建logger
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	// 添加全局字段
	logger = logger.With(
		zap.String("service", config.ServiceName),
		zap.String("environment", config.Environment),
	)

	return &StructuredLogger{
		logger: logger,
		sugar:  logger.Sugar(),
	}, nil
}

// WithContext 添加上下文信息
func (l *StructuredLogger) WithContext(ctx context.Context) *StructuredLogger {
	fields := l.extractContextFields(ctx)
	if len(fields) == 0 {
		return l
	}

	return &StructuredLogger{
		logger: l.logger.With(fields...),
		sugar:  l.logger.With(fields...).Sugar(),
	}
}

// WithGinContext 添加Gin上下文信息
func (l *StructuredLogger) WithGinContext(c *gin.Context) *StructuredLogger {
	fields := []zap.Field{
		zap.String("request_id", l.getRequestID(c)),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
	}

	// 添加追踪信息
	if traceID := tracing.GetTraceID(c); traceID != "" {
		fields = append(fields, zap.String("trace_id", traceID))
	}
	if spanID := tracing.GetSpanID(c); spanID != "" {
		fields = append(fields, zap.String("span_id", spanID))
	}

	// 添加用户信息
	if userID, exists := c.Get("user_id"); exists {
		fields = append(fields, zap.String("user_id", fmt.Sprintf("%v", userID)))
	}

	return &StructuredLogger{
		logger: l.logger.With(fields...),
		sugar:  l.logger.With(fields...).Sugar(),
	}
}

// WithFields 添加字段
func (l *StructuredLogger) WithFields(fields map[string]interface{}) *StructuredLogger {
	zapFields := make([]zap.Field, 0, len(fields))
	for key, value := range fields {
		zapFields = append(zapFields, zap.Any(key, value))
	}

	return &StructuredLogger{
		logger: l.logger.With(zapFields...),
		sugar:  l.logger.With(zapFields...).Sugar(),
	}
}

// Debug 调试日志
func (l *StructuredLogger) Debug(msg string, fields ...zap.Field) {
	l.logger.Debug(msg, fields...)
}

// Info 信息日志
func (l *StructuredLogger) Info(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}

// Warn 警告日志
func (l *StructuredLogger) Warn(msg string, fields ...zap.Field) {
	l.logger.Warn(msg, fields...)
}

// Error 错误日志
func (l *StructuredLogger) Error(msg string, fields ...zap.Field) {
	l.logger.Error(msg, fields...)
}

// Fatal 致命错误日志
func (l *StructuredLogger) Fatal(msg string, fields ...zap.Field) {
	l.logger.Fatal(msg, fields...)
}

// Debugf 格式化调试日志
func (l *StructuredLogger) Debugf(template string, args ...interface{}) {
	l.sugar.Debugf(template, args...)
}

// Infof 格式化信息日志
func (l *StructuredLogger) Infof(template string, args ...interface{}) {
	l.sugar.Infof(template, args...)
}

// Warnf 格式化警告日志
func (l *StructuredLogger) Warnf(template string, args ...interface{}) {
	l.sugar.Warnf(template, args...)
}

// Errorf 格式化错误日志
func (l *StructuredLogger) Errorf(template string, args ...interface{}) {
	l.sugar.Errorf(template, args...)
}

// Fatalf 格式化致命错误日志
func (l *StructuredLogger) Fatalf(template string, args ...interface{}) {
	l.sugar.Fatalf(template, args...)
}

// LogRequest 记录HTTP请求
func (l *StructuredLogger) LogRequest(c *gin.Context, duration time.Duration) {
	l.WithGinContext(c).Info("HTTP request",
		zap.Int("status", c.Writer.Status()),
		zap.Duration("duration", duration),
		zap.Int("response_size", c.Writer.Size()),
	)
}

// LogError 记录错误
func (l *StructuredLogger) LogError(err error, msg string, fields ...zap.Field) {
	allFields := append(fields, zap.Error(err))
	l.logger.Error(msg, allFields...)
}

// LogPanic 记录panic
func (l *StructuredLogger) LogPanic(recovered interface{}, msg string, fields ...zap.Field) {
	allFields := append(fields, zap.Any("panic", recovered))
	l.logger.Error(msg, allFields...)
}

// Sync 同步日志
func (l *StructuredLogger) Sync() error {
	return l.logger.Sync()
}

// extractContextFields 从上下文提取字段
func (l *StructuredLogger) extractContextFields(ctx context.Context) []zap.Field {
	var fields []zap.Field

	// 提取追踪信息
	if span := opentracing.SpanFromContext(ctx); span != nil {
		// 这里可以添加更多的追踪信息提取逻辑
	}

	return fields
}

// getRequestID 获取请求ID
func (l *StructuredLogger) getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		return fmt.Sprintf("%v", requestID)
	}
	return c.GetHeader("X-Request-ID")
}

// RequestLoggingMiddleware 请求日志中间件
func RequestLoggingMiddleware(logger *StructuredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 记录请求开始
		logger.WithGinContext(c).Info("Request started")

		c.Next()

		// 记录请求完成
		duration := time.Since(start)
		logger.LogRequest(c, duration)
	}
}

// ErrorLoggingMiddleware 错误日志中间件
func ErrorLoggingMiddleware(logger *StructuredLogger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		logger.WithGinContext(c).LogPanic(recovered, "Panic recovered")
		c.JSON(500, gin.H{"error": "Internal server error"})
	})
}

// BusinessEventLogger 业务事件日志器
type BusinessEventLogger struct {
	logger *StructuredLogger
}

// NewBusinessEventLogger 创建业务事件日志器
func NewBusinessEventLogger(logger *StructuredLogger) *BusinessEventLogger {
	return &BusinessEventLogger{
		logger: logger,
	}
}

// LogUserRegistration 记录用户注册事件
func (b *BusinessEventLogger) LogUserRegistration(userID, email string) {
	b.logger.Info("User registered",
		zap.String("event_type", "user_registration"),
		zap.String("user_id", userID),
		zap.String("email", email),
		zap.Time("event_time", time.Now()),
	)
}

// LogUserLogin 记录用户登录事件
func (b *BusinessEventLogger) LogUserLogin(userID, email string, success bool) {
	b.logger.Info("User login attempt",
		zap.String("event_type", "user_login"),
		zap.String("user_id", userID),
		zap.String("email", email),
		zap.Bool("success", success),
		zap.Time("event_time", time.Now()),
	)
}

// LogOrderCreation 记录订单创建事件
func (b *BusinessEventLogger) LogOrderCreation(orderID, userID string, amount float64) {
	b.logger.Info("Order created",
		zap.String("event_type", "order_creation"),
		zap.String("order_id", orderID),
		zap.String("user_id", userID),
		zap.Float64("amount", amount),
		zap.Time("event_time", time.Now()),
	)
}

// LogPaymentTransaction 记录支付交易事件
func (b *BusinessEventLogger) LogPaymentTransaction(paymentID, orderID, userID string, amount float64, status string) {
	b.logger.Info("Payment transaction",
		zap.String("event_type", "payment_transaction"),
		zap.String("payment_id", paymentID),
		zap.String("order_id", orderID),
		zap.String("user_id", userID),
		zap.Float64("amount", amount),
		zap.String("status", status),
		zap.Time("event_time", time.Now()),
	)
}
