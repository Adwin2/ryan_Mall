package monitoring

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// PrometheusMetrics Prometheus指标收集器
type PrometheusMetrics struct {
	// HTTP请求相关指标
	httpRequestsTotal     *prometheus.CounterVec
	httpRequestDuration   *prometheus.HistogramVec
	httpRequestsInFlight  *prometheus.GaugeVec
	httpResponseSize      *prometheus.HistogramVec

	// 业务指标
	userRegistrations     prometheus.Counter
	userLogins           prometheus.Counter
	orderCreations       prometheus.Counter
	paymentTransactions  prometheus.Counter
	seckillParticipations prometheus.Counter

	// 系统指标
	databaseConnections  *prometheus.GaugeVec
	redisConnections     *prometheus.GaugeVec
	cacheHitRate        *prometheus.GaugeVec
	queueLength         *prometheus.GaugeVec

	// 错误指标
	errorRate           *prometheus.CounterVec
	panicRecoveries     prometheus.Counter
}

// NewPrometheusMetrics 创建Prometheus指标收集器
func NewPrometheusMetrics() *PrometheusMetrics {
	metrics := &PrometheusMetrics{
		// HTTP请求指标
		httpRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status_code"},
		),
		httpRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
		httpRequestsInFlight: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "http_requests_in_flight",
				Help: "Number of HTTP requests currently being processed",
			},
			[]string{"method", "endpoint"},
		),
		httpResponseSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_response_size_bytes",
				Help:    "HTTP response size in bytes",
				Buckets: []float64{100, 1000, 10000, 100000, 1000000},
			},
			[]string{"method", "endpoint"},
		),

		// 业务指标
		userRegistrations: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "user_registrations_total",
				Help: "Total number of user registrations",
			},
		),
		userLogins: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "user_logins_total",
				Help: "Total number of user logins",
			},
		),
		orderCreations: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "order_creations_total",
				Help: "Total number of order creations",
			},
		),
		paymentTransactions: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "payment_transactions_total",
				Help: "Total number of payment transactions",
			},
		),
		seckillParticipations: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "seckill_participations_total",
				Help: "Total number of seckill participations",
			},
		),

		// 系统指标
		databaseConnections: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "database_connections",
				Help: "Number of database connections",
			},
			[]string{"database", "status"},
		),
		redisConnections: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "redis_connections",
				Help: "Number of Redis connections",
			},
			[]string{"instance", "status"},
		),
		cacheHitRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "cache_hit_rate",
				Help: "Cache hit rate percentage",
			},
			[]string{"cache_type"},
		),
		queueLength: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "queue_length",
				Help: "Number of items in queue",
			},
			[]string{"queue_name"},
		),

		// 错误指标
		errorRate: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "errors_total",
				Help: "Total number of errors",
			},
			[]string{"service", "error_type"},
		),
		panicRecoveries: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "panic_recoveries_total",
				Help: "Total number of panic recoveries",
			},
		),
	}

	// 注册所有指标
	prometheus.MustRegister(
		metrics.httpRequestsTotal,
		metrics.httpRequestDuration,
		metrics.httpRequestsInFlight,
		metrics.httpResponseSize,
		metrics.userRegistrations,
		metrics.userLogins,
		metrics.orderCreations,
		metrics.paymentTransactions,
		metrics.seckillParticipations,
		metrics.databaseConnections,
		metrics.redisConnections,
		metrics.cacheHitRate,
		metrics.queueLength,
		metrics.errorRate,
		metrics.panicRecoveries,
	)

	return metrics
}

// PrometheusMiddleware Prometheus监控中间件
func (m *PrometheusMetrics) PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// 增加正在处理的请求数
		m.httpRequestsInFlight.WithLabelValues(method, path).Inc()

		c.Next()

		// 减少正在处理的请求数
		m.httpRequestsInFlight.WithLabelValues(method, path).Dec()

		// 记录请求完成指标
		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(c.Writer.Status())
		responseSize := float64(c.Writer.Size())

		m.httpRequestsTotal.WithLabelValues(method, path, statusCode).Inc()
		m.httpRequestDuration.WithLabelValues(method, path).Observe(duration)
		m.httpResponseSize.WithLabelValues(method, path).Observe(responseSize)
	}
}

// RecordUserRegistration 记录用户注册
func (m *PrometheusMetrics) RecordUserRegistration() {
	m.userRegistrations.Inc()
}

// RecordUserLogin 记录用户登录
func (m *PrometheusMetrics) RecordUserLogin() {
	m.userLogins.Inc()
}

// RecordOrderCreation 记录订单创建
func (m *PrometheusMetrics) RecordOrderCreation() {
	m.orderCreations.Inc()
}

// RecordPaymentTransaction 记录支付交易
func (m *PrometheusMetrics) RecordPaymentTransaction() {
	m.paymentTransactions.Inc()
}

// RecordSeckillParticipation 记录秒杀参与
func (m *PrometheusMetrics) RecordSeckillParticipation() {
	m.seckillParticipations.Inc()
}

// SetDatabaseConnections 设置数据库连接数
func (m *PrometheusMetrics) SetDatabaseConnections(database, status string, count float64) {
	m.databaseConnections.WithLabelValues(database, status).Set(count)
}

// SetRedisConnections 设置Redis连接数
func (m *PrometheusMetrics) SetRedisConnections(instance, status string, count float64) {
	m.redisConnections.WithLabelValues(instance, status).Set(count)
}

// SetCacheHitRate 设置缓存命中率
func (m *PrometheusMetrics) SetCacheHitRate(cacheType string, rate float64) {
	m.cacheHitRate.WithLabelValues(cacheType).Set(rate)
}

// SetQueueLength 设置队列长度
func (m *PrometheusMetrics) SetQueueLength(queueName string, length float64) {
	m.queueLength.WithLabelValues(queueName).Set(length)
}

// RecordError 记录错误
func (m *PrometheusMetrics) RecordError(service, errorType string) {
	m.errorRate.WithLabelValues(service, errorType).Inc()
}

// RecordPanicRecovery 记录panic恢复
func (m *PrometheusMetrics) RecordPanicRecovery() {
	m.panicRecoveries.Inc()
}

// Handler 返回Prometheus指标处理器
func (m *PrometheusMetrics) Handler() http.Handler {
	return promhttp.Handler()
}

// MetricsHandler Gin路由处理器
func (m *PrometheusMetrics) MetricsHandler() gin.HandlerFunc {
	handler := promhttp.Handler()
	return func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}
}

// HealthMetrics 健康检查指标
type HealthMetrics struct {
	ServiceUp    *prometheus.GaugeVec
	ServiceInfo  *prometheus.GaugeVec
	BuildInfo    *prometheus.GaugeVec
}

// NewHealthMetrics 创建健康检查指标
func NewHealthMetrics(serviceName, version, buildTime string) *HealthMetrics {
	metrics := &HealthMetrics{
		ServiceUp: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "service_up",
				Help: "Service availability (1 = up, 0 = down)",
			},
			[]string{"service"},
		),
		ServiceInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "service_info",
				Help: "Service information",
			},
			[]string{"service", "version"},
		),
		BuildInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "build_info",
				Help: "Build information",
			},
			[]string{"service", "version", "build_time"},
		),
	}

	// 注册指标
	prometheus.MustRegister(
		metrics.ServiceUp,
		metrics.ServiceInfo,
		metrics.BuildInfo,
	)

	// 设置静态信息
	metrics.ServiceUp.WithLabelValues(serviceName).Set(1)
	metrics.ServiceInfo.WithLabelValues(serviceName, version).Set(1)
	metrics.BuildInfo.WithLabelValues(serviceName, version, buildTime).Set(1)

	return metrics
}

// SetServiceStatus 设置服务状态
func (h *HealthMetrics) SetServiceStatus(serviceName string, up bool) {
	if up {
		h.ServiceUp.WithLabelValues(serviceName).Set(1)
	} else {
		h.ServiceUp.WithLabelValues(serviceName).Set(0)
	}
}
