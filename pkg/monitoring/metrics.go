package monitoring

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsManager 指标管理器
type MetricsManager struct {
	// HTTP指标
	httpRequestsTotal     *prometheus.CounterVec
	httpRequestDuration   *prometheus.HistogramVec
	httpRequestsInFlight  prometheus.Gauge
	
	// 业务指标
	userRegistrations     prometheus.Counter
	userLogins           prometheus.Counter
	orderCreated         *prometheus.CounterVec
	orderValue           *prometheus.HistogramVec
	productViews         *prometheus.CounterVec
	cartOperations       *prometheus.CounterVec
	
	// 系统指标
	databaseConnections  prometheus.Gauge
	redisConnections     prometheus.Gauge
	activeWebsockets     prometheus.Gauge
	
	// 错误指标
	errorTotal           *prometheus.CounterVec
	
	// 自定义指标
	customMetrics        map[string]prometheus.Collector
}

// NewMetricsManager 创建指标管理器
func NewMetricsManager() *MetricsManager {
	mm := &MetricsManager{
		customMetrics: make(map[string]prometheus.Collector),
	}
	
	mm.initMetrics()
	return mm
}

// initMetrics 初始化指标
func (mm *MetricsManager) initMetrics() {
	// HTTP指标
	mm.httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code"},
	)
	
	mm.httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
	
	mm.httpRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Number of HTTP requests currently being processed",
		},
	)
	
	// 业务指标
	mm.userRegistrations = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "user_registrations_total",
			Help: "Total number of user registrations",
		},
	)
	
	mm.userLogins = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "user_logins_total",
			Help: "Total number of user logins",
		},
	)
	
	mm.orderCreated = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "orders_created_total",
			Help: "Total number of orders created",
		},
		[]string{"status"},
	)
	
	mm.orderValue = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "order_value_yuan",
			Help:    "Order value in yuan",
			Buckets: []float64{10, 50, 100, 500, 1000, 5000, 10000, 50000},
		},
		[]string{"payment_method"},
	)
	
	mm.productViews = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "product_views_total",
			Help: "Total number of product views",
		},
		[]string{"product_id", "category"},
	)
	
	mm.cartOperations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cart_operations_total",
			Help: "Total number of cart operations",
		},
		[]string{"operation"}, // add, remove, update, clear
	)
	
	// 系统指标
	mm.databaseConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "database_connections_active",
			Help: "Number of active database connections",
		},
	)
	
	mm.redisConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "redis_connections_active",
			Help: "Number of active Redis connections",
		},
	)
	
	mm.activeWebsockets = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "websocket_connections_active",
			Help: "Number of active WebSocket connections",
		},
	)
	
	// 错误指标
	mm.errorTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "errors_total",
			Help: "Total number of errors",
		},
		[]string{"type", "component"},
	)
}

// HTTPMetricsMiddleware HTTP指标中间件
func (mm *MetricsManager) HTTPMetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// 增加正在处理的请求数
		mm.httpRequestsInFlight.Inc()
		defer mm.httpRequestsInFlight.Dec()
		
		// 处理请求
		c.Next()
		
		// 记录指标
		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(c.Writer.Status())
		
		mm.httpRequestsTotal.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			statusCode,
		).Inc()
		
		mm.httpRequestDuration.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
		).Observe(duration)
	}
}

// RecordUserRegistration 记录用户注册
func (mm *MetricsManager) RecordUserRegistration() {
	mm.userRegistrations.Inc()
}

// RecordUserLogin 记录用户登录
func (mm *MetricsManager) RecordUserLogin() {
	mm.userLogins.Inc()
}

// RecordOrderCreated 记录订单创建
func (mm *MetricsManager) RecordOrderCreated(status string, value float64, paymentMethod string) {
	mm.orderCreated.WithLabelValues(status).Inc()
	mm.orderValue.WithLabelValues(paymentMethod).Observe(value)
}

// RecordProductView 记录商品浏览
func (mm *MetricsManager) RecordProductView(productID uint, category string) {
	mm.productViews.WithLabelValues(
		strconv.FormatUint(uint64(productID), 10),
		category,
	).Inc()
}

// RecordCartOperation 记录购物车操作
func (mm *MetricsManager) RecordCartOperation(operation string) {
	mm.cartOperations.WithLabelValues(operation).Inc()
}

// UpdateDatabaseConnections 更新数据库连接数
func (mm *MetricsManager) UpdateDatabaseConnections(count float64) {
	mm.databaseConnections.Set(count)
}

// UpdateRedisConnections 更新Redis连接数
func (mm *MetricsManager) UpdateRedisConnections(count float64) {
	mm.redisConnections.Set(count)
}

// UpdateActiveWebsockets 更新活跃WebSocket连接数
func (mm *MetricsManager) UpdateActiveWebsockets(count float64) {
	mm.activeWebsockets.Set(count)
}

// RecordError 记录错误
func (mm *MetricsManager) RecordError(errorType, component string) {
	mm.errorTotal.WithLabelValues(errorType, component).Inc()
}

// BusinessMetrics 业务指标
type BusinessMetrics struct {
	metricsManager *MetricsManager
	
	// 实时业务指标
	dailyRevenue      prometheus.Gauge
	dailyOrders       prometheus.Gauge
	dailyUsers        prometheus.Gauge
	conversionRate    prometheus.Gauge
	averageOrderValue prometheus.Gauge
	
	// 商品指标
	topProducts       *prometheus.GaugeVec
	categoryRevenue   *prometheus.GaugeVec
	
	// 用户指标
	userActivity      *prometheus.GaugeVec
}

// NewBusinessMetrics 创建业务指标
func NewBusinessMetrics(mm *MetricsManager) *BusinessMetrics {
	bm := &BusinessMetrics{
		metricsManager: mm,
	}
	
	bm.initBusinessMetrics()
	return bm
}

// initBusinessMetrics 初始化业务指标
func (bm *BusinessMetrics) initBusinessMetrics() {
	bm.dailyRevenue = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "daily_revenue_yuan",
			Help: "Daily revenue in yuan",
		},
	)
	
	bm.dailyOrders = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "daily_orders_count",
			Help: "Daily orders count",
		},
	)
	
	bm.dailyUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "daily_active_users",
			Help: "Daily active users count",
		},
	)
	
	bm.conversionRate = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "conversion_rate_percent",
			Help: "Conversion rate percentage",
		},
	)
	
	bm.averageOrderValue = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "average_order_value_yuan",
			Help: "Average order value in yuan",
		},
	)
	
	bm.topProducts = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "top_products_sales",
			Help: "Top products by sales",
		},
		[]string{"product_id", "product_name"},
	)
	
	bm.categoryRevenue = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "category_revenue_yuan",
			Help: "Revenue by category in yuan",
		},
		[]string{"category_id", "category_name"},
	)
	
	bm.userActivity = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "user_activity_score",
			Help: "User activity score",
		},
		[]string{"activity_type"},
	)
}

// UpdateDailyRevenue 更新日收入
func (bm *BusinessMetrics) UpdateDailyRevenue(revenue float64) {
	bm.dailyRevenue.Set(revenue)
}

// UpdateDailyOrders 更新日订单数
func (bm *BusinessMetrics) UpdateDailyOrders(orders float64) {
	bm.dailyOrders.Set(orders)
}

// UpdateDailyUsers 更新日活用户
func (bm *BusinessMetrics) UpdateDailyUsers(users float64) {
	bm.dailyUsers.Set(users)
}

// UpdateConversionRate 更新转化率
func (bm *BusinessMetrics) UpdateConversionRate(rate float64) {
	bm.conversionRate.Set(rate)
}

// UpdateAverageOrderValue 更新平均订单价值
func (bm *BusinessMetrics) UpdateAverageOrderValue(value float64) {
	bm.averageOrderValue.Set(value)
}

// UpdateTopProductSales 更新热门商品销量
func (bm *BusinessMetrics) UpdateTopProductSales(productID uint, productName string, sales float64) {
	bm.topProducts.WithLabelValues(
		strconv.FormatUint(uint64(productID), 10),
		productName,
	).Set(sales)
}

// UpdateCategoryRevenue 更新分类收入
func (bm *BusinessMetrics) UpdateCategoryRevenue(categoryID uint, categoryName string, revenue float64) {
	bm.categoryRevenue.WithLabelValues(
		strconv.FormatUint(uint64(categoryID), 10),
		categoryName,
	).Set(revenue)
}

// AlertManager 告警管理器
type AlertManager struct {
	metricsManager *MetricsManager
	
	// 告警阈值
	errorRateThreshold    float64
	responseTimeThreshold float64
	connectionThreshold   float64
}

// NewAlertManager 创建告警管理器
func NewAlertManager(mm *MetricsManager) *AlertManager {
	return &AlertManager{
		metricsManager:        mm,
		errorRateThreshold:    0.05, // 5%错误率
		responseTimeThreshold: 2.0,  // 2秒响应时间
		connectionThreshold:   1000, // 1000连接数
	}
}

// CheckAlerts 检查告警条件
func (am *AlertManager) CheckAlerts() []Alert {
	var alerts []Alert
	
	// 这里可以添加具体的告警检查逻辑
	// 例如：检查错误率、响应时间、连接数等
	
	return alerts
}

// Alert 告警结构
type Alert struct {
	Level       string    `json:"level"`       // info, warning, critical
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
	Value       float64   `json:"value"`
	Threshold   float64   `json:"threshold"`
}

// MetricsCollector 指标收集器
type MetricsCollector struct {
	metricsManager  *MetricsManager
	businessMetrics *BusinessMetrics
	
	// 数据源
	dbStats    func() (int, int) // 返回活跃连接数和最大连接数
	redisStats func() int        // 返回Redis连接数
	wsStats    func() int        // 返回WebSocket连接数
}

// NewMetricsCollector 创建指标收集器
func NewMetricsCollector(mm *MetricsManager, bm *BusinessMetrics) *MetricsCollector {
	return &MetricsCollector{
		metricsManager:  mm,
		businessMetrics: bm,
	}
}

// SetDataSources 设置数据源
func (mc *MetricsCollector) SetDataSources(
	dbStats func() (int, int),
	redisStats func() int,
	wsStats func() int,
) {
	mc.dbStats = dbStats
	mc.redisStats = redisStats
	mc.wsStats = wsStats
}

// StartCollection 开始收集指标
func (mc *MetricsCollector) StartCollection() {
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for range ticker.C {
			mc.collectSystemMetrics()
		}
	}()
}

// collectSystemMetrics 收集系统指标
func (mc *MetricsCollector) collectSystemMetrics() {
	if mc.dbStats != nil {
		active, _ := mc.dbStats()
		mc.metricsManager.UpdateDatabaseConnections(float64(active))
	}
	
	if mc.redisStats != nil {
		connections := mc.redisStats()
		mc.metricsManager.UpdateRedisConnections(float64(connections))
	}
	
	if mc.wsStats != nil {
		connections := mc.wsStats()
		mc.metricsManager.UpdateActiveWebsockets(float64(connections))
	}
}

// GetMetricsHandler 获取Prometheus指标处理器
func GetMetricsHandler() gin.HandlerFunc {
	handler := promhttp.Handler()
	return gin.WrapH(handler)
}
