package tracing

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
)

// JaegerConfig Jaeger配置
type JaegerConfig struct {
	ServiceName     string  `json:"service_name"`
	AgentHost       string  `json:"agent_host"`
	AgentPort       int     `json:"agent_port"`
	CollectorURL    string  `json:"collector_url"`
	SamplingRate    float64 `json:"sampling_rate"`
	LogSpans        bool    `json:"log_spans"`
	Disabled        bool    `json:"disabled"`
}

// DefaultJaegerConfig 默认Jaeger配置
func DefaultJaegerConfig(serviceName string) *JaegerConfig {
	return &JaegerConfig{
		ServiceName:  serviceName,
		AgentHost:    "localhost",
		AgentPort:    6831,
		SamplingRate: 1.0,
		LogSpans:     false,
		Disabled:     false,
	}
}

// InitJaeger 初始化Jaeger追踪器
func InitJaeger(cfg *JaegerConfig) (opentracing.Tracer, io.Closer, error) {
	if cfg.Disabled {
		return opentracing.NoopTracer{}, &noopCloser{}, nil
	}

	// 配置Jaeger
	jaegerCfg := config.Configuration{
		ServiceName: cfg.ServiceName,
		Sampler: &config.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: cfg.SamplingRate,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:           cfg.LogSpans,
			LocalAgentHostPort: fmt.Sprintf("%s:%d", cfg.AgentHost, cfg.AgentPort),
		},
	}

	// 如果配置了Collector URL，使用HTTP Reporter
	if cfg.CollectorURL != "" {
		jaegerCfg.Reporter.CollectorEndpoint = cfg.CollectorURL
	}

	// 创建追踪器
	tracer, closer, err := jaegerCfg.NewTracer()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Jaeger tracer: %w", err)
	}

	// 设置全局追踪器
	opentracing.SetGlobalTracer(tracer)

	return tracer, closer, nil
}

// TracingMiddleware 链路追踪中间件
func TracingMiddleware(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中提取span上下文
		spanCtx, _ := opentracing.GlobalTracer().Extract(
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(c.Request.Header),
		)

		// 创建新的span
		operationName := fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path)
		span := opentracing.GlobalTracer().StartSpan(
			operationName,
			ext.RPCServerOption(spanCtx),
		)
		defer span.Finish()

		// 设置span标签
		ext.HTTPMethod.Set(span, c.Request.Method)
		ext.HTTPUrl.Set(span, c.Request.URL.String())
		ext.Component.Set(span, serviceName)

		// 将span上下文添加到请求上下文
		ctx := opentracing.ContextWithSpan(c.Request.Context(), span)
		c.Request = c.Request.WithContext(ctx)

		// 设置追踪ID到响应头
		if jaegerSpan, ok := span.(*jaeger.Span); ok {
			traceID := jaegerSpan.SpanContext().TraceID().String()
			c.Header("X-Trace-ID", traceID)
		}

		c.Next()

		// 设置响应状态码
		statusCode := c.Writer.Status()
		ext.HTTPStatusCode.Set(span, uint16(statusCode))

		// 如果是错误状态码，标记为错误
		if statusCode >= 400 {
			ext.Error.Set(span, true)
			span.SetTag("error.message", http.StatusText(statusCode))
		}
	}
}

// StartSpan 开始一个新的span
func StartSpan(ctx context.Context, operationName string) (opentracing.Span, context.Context) {
	span, ctx := opentracing.StartSpanFromContext(ctx, operationName)
	return span, ctx
}

// StartSpanFromGin 从Gin上下文开始一个新的span
func StartSpanFromGin(c *gin.Context, operationName string) opentracing.Span {
	parentSpan := opentracing.SpanFromContext(c.Request.Context())
	if parentSpan == nil {
		return opentracing.GlobalTracer().StartSpan(operationName)
	}
	return opentracing.GlobalTracer().StartSpan(
		operationName,
		opentracing.ChildOf(parentSpan.Context()),
	)
}

// LogError 记录错误到span
func LogError(span opentracing.Span, err error) {
	if span == nil || err == nil {
		return
	}

	ext.Error.Set(span, true)
	span.LogFields(
		log.String("error", err.Error()),
	)
}

// LogEvent 记录事件到span
func LogEvent(span opentracing.Span, event string, fields ...log.Field) {
	if span == nil {
		return
	}

	logFields := []log.Field{
		log.String("event", event),
	}
	logFields = append(logFields, fields...)
	span.LogFields(logFields...)
}

// SetTag 设置span标签
func SetTag(span opentracing.Span, key string, value interface{}) {
	if span == nil {
		return
	}
	span.SetTag(key, value)
}

// InjectToHTTPRequest 将span上下文注入到HTTP请求头
func InjectToHTTPRequest(span opentracing.Span, req *http.Request) error {
	if span == nil {
		return nil
	}

	return opentracing.GlobalTracer().Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header),
	)
}

// ExtractFromHTTPRequest 从HTTP请求头提取span上下文
func ExtractFromHTTPRequest(req *http.Request) (opentracing.SpanContext, error) {
	return opentracing.GlobalTracer().Extract(
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header),
	)
}

// TraceHTTPClient HTTP客户端追踪包装器
func TraceHTTPClient(client *http.Client, operationName string) *http.Client {
	if client == nil {
		client = &http.Client{}
	}

	originalTransport := client.Transport
	if originalTransport == nil {
		originalTransport = http.DefaultTransport
	}

	client.Transport = &tracingTransport{
		transport:     originalTransport,
		operationName: operationName,
	}

	return client
}

// tracingTransport HTTP传输层追踪包装器
type tracingTransport struct {
	transport     http.RoundTripper
	operationName string
}

// RoundTrip 实现http.RoundTripper接口
func (t *tracingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// 从请求上下文获取span
	parentSpan := opentracing.SpanFromContext(req.Context())
	
	var span opentracing.Span
	if parentSpan != nil {
		span = opentracing.GlobalTracer().StartSpan(
			t.operationName,
			opentracing.ChildOf(parentSpan.Context()),
		)
	} else {
		span = opentracing.GlobalTracer().StartSpan(t.operationName)
	}
	defer span.Finish()

	// 设置HTTP客户端标签
	ext.SpanKindRPCClient.Set(span)
	ext.HTTPMethod.Set(span, req.Method)
	ext.HTTPUrl.Set(span, req.URL.String())

	// 注入span上下文到请求头
	if err := InjectToHTTPRequest(span, req); err != nil {
		LogError(span, err)
	}

	// 执行请求
	resp, err := t.transport.RoundTrip(req)
	if err != nil {
		ext.Error.Set(span, true)
		LogError(span, err)
		return resp, err
	}

	// 设置响应状态码
	ext.HTTPStatusCode.Set(span, uint16(resp.StatusCode))
	if resp.StatusCode >= 400 {
		ext.Error.Set(span, true)
	}

	return resp, nil
}

// noopCloser 空的closer实现
type noopCloser struct{}

func (n *noopCloser) Close() error {
	return nil
}

// GetTraceID 获取当前请求的追踪ID
func GetTraceID(c *gin.Context) string {
	span := opentracing.SpanFromContext(c.Request.Context())
	if span == nil {
		return ""
	}

	if jaegerSpan, ok := span.(*jaeger.Span); ok {
		return jaegerSpan.SpanContext().TraceID().String()
	}

	return ""
}

// GetSpanID 获取当前span ID
func GetSpanID(c *gin.Context) string {
	span := opentracing.SpanFromContext(c.Request.Context())
	if span == nil {
		return ""
	}

	if jaegerSpan, ok := span.(*jaeger.Span); ok {
		return jaegerSpan.SpanContext().SpanID().String()
	}

	return ""
}
