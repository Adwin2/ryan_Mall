package health

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// HealthStatus 健康状态
type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusUnhealthy HealthStatus = "unhealthy"
	StatusDegraded  HealthStatus = "degraded"
)

// CheckResult 检查结果
type CheckResult struct {
	Name        string                 `json:"name"`
	Status      HealthStatus           `json:"status"`
	Message     string                 `json:"message"`
	Duration    time.Duration          `json:"duration"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// HealthReport 健康报告
type HealthReport struct {
	Status    HealthStatus             `json:"status"`
	Timestamp time.Time                `json:"timestamp"`
	Duration  time.Duration            `json:"duration"`
	Checks    map[string]*CheckResult  `json:"checks"`
	Metadata  map[string]interface{}   `json:"metadata,omitempty"`
}

// HealthChecker 健康检查器接口
type HealthChecker interface {
	Check(ctx context.Context) *CheckResult
	Name() string
}

// HealthManager 健康管理器
type HealthManager struct {
	checkers []HealthChecker
	timeout  time.Duration
	metadata map[string]interface{}
	mutex    sync.RWMutex
}

// NewHealthManager 创建健康管理器
func NewHealthManager(timeout time.Duration) *HealthManager {
	return &HealthManager{
		checkers: make([]HealthChecker, 0),
		timeout:  timeout,
		metadata: make(map[string]interface{}),
	}
}

// AddChecker 添加检查器
func (h *HealthManager) AddChecker(checker HealthChecker) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.checkers = append(h.checkers, checker)
}

// SetMetadata 设置元数据
func (h *HealthManager) SetMetadata(key string, value interface{}) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.metadata[key] = value
}

// Check 执行健康检查
func (h *HealthManager) Check(ctx context.Context) *HealthReport {
	start := time.Now()
	
	// 创建带超时的上下文
	checkCtx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()

	h.mutex.RLock()
	checkers := make([]HealthChecker, len(h.checkers))
	copy(checkers, h.checkers)
	metadata := make(map[string]interface{})
	for k, v := range h.metadata {
		metadata[k] = v
	}
	h.mutex.RUnlock()

	// 并发执行检查
	results := make(chan *CheckResult, len(checkers))
	for _, checker := range checkers {
		go func(c HealthChecker) {
			results <- c.Check(checkCtx)
		}(checker)
	}

	// 收集结果
	checks := make(map[string]*CheckResult)
	overallStatus := StatusHealthy

	for i := 0; i < len(checkers); i++ {
		result := <-results
		checks[result.Name] = result

		// 确定整体状态
		switch result.Status {
		case StatusUnhealthy:
			overallStatus = StatusUnhealthy
		case StatusDegraded:
			if overallStatus == StatusHealthy {
				overallStatus = StatusDegraded
			}
		}
	}

	return &HealthReport{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Duration:  time.Since(start),
		Checks:    checks,
		Metadata:  metadata,
	}
}

// Handler 返回HTTP处理器
func (h *HealthManager) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		report := h.Check(c.Request.Context())
		
		var statusCode int
		switch report.Status {
		case StatusHealthy:
			statusCode = http.StatusOK
		case StatusDegraded:
			statusCode = http.StatusOK // 降级状态仍返回200
		case StatusUnhealthy:
			statusCode = http.StatusServiceUnavailable
		default:
			statusCode = http.StatusInternalServerError
		}

		c.JSON(statusCode, report)
	}
}

// DatabaseChecker 数据库健康检查器
type DatabaseChecker struct {
	name string
	db   *gorm.DB
}

// NewDatabaseChecker 创建数据库检查器
func NewDatabaseChecker(name string, db *gorm.DB) *DatabaseChecker {
	return &DatabaseChecker{
		name: name,
		db:   db,
	}
}

// Name 返回检查器名称
func (d *DatabaseChecker) Name() string {
	return d.name
}

// Check 执行数据库检查
func (d *DatabaseChecker) Check(ctx context.Context) *CheckResult {
	start := time.Now()
	result := &CheckResult{
		Name:      d.name,
		Timestamp: start,
		Metadata:  make(map[string]interface{}),
	}

	// 获取底层SQL DB
	sqlDB, err := d.db.DB()
	if err != nil {
		result.Status = StatusUnhealthy
		result.Message = fmt.Sprintf("Failed to get SQL DB: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	// 检查连接
	if err := sqlDB.PingContext(ctx); err != nil {
		result.Status = StatusUnhealthy
		result.Message = fmt.Sprintf("Database ping failed: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	// 获取连接池统计
	stats := sqlDB.Stats()
	result.Metadata["open_connections"] = stats.OpenConnections
	result.Metadata["in_use"] = stats.InUse
	result.Metadata["idle"] = stats.Idle

	// 检查连接池状态
	if stats.OpenConnections > 0 {
		result.Status = StatusHealthy
		result.Message = "Database is healthy"
	} else {
		result.Status = StatusDegraded
		result.Message = "No database connections"
	}

	result.Duration = time.Since(start)
	return result
}

// RedisChecker Redis健康检查器
type RedisChecker struct {
	name   string
	client *redis.Client
}

// NewRedisChecker 创建Redis检查器
func NewRedisChecker(name string, client *redis.Client) *RedisChecker {
	return &RedisChecker{
		name:   name,
		client: client,
	}
}

// Name 返回检查器名称
func (r *RedisChecker) Name() string {
	return r.name
}

// Check 执行Redis检查
func (r *RedisChecker) Check(ctx context.Context) *CheckResult {
	start := time.Now()
	result := &CheckResult{
		Name:      r.name,
		Timestamp: start,
		Metadata:  make(map[string]interface{}),
	}

	// 执行PING命令
	pong, err := r.client.Ping(ctx).Result()
	if err != nil {
		result.Status = StatusUnhealthy
		result.Message = fmt.Sprintf("Redis ping failed: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	if pong != "PONG" {
		result.Status = StatusUnhealthy
		result.Message = fmt.Sprintf("Unexpected ping response: %s", pong)
		result.Duration = time.Since(start)
		return result
	}

	// 获取Redis信息
	info, err := r.client.Info(ctx, "server").Result()
	if err == nil {
		result.Metadata["info"] = info
	}

	result.Status = StatusHealthy
	result.Message = "Redis is healthy"
	result.Duration = time.Since(start)
	return result
}

// HTTPChecker HTTP服务健康检查器
type HTTPChecker struct {
	name   string
	url    string
	client *http.Client
}

// NewHTTPChecker 创建HTTP检查器
func NewHTTPChecker(name, url string, timeout time.Duration) *HTTPChecker {
	return &HTTPChecker{
		name: name,
		url:  url,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Name 返回检查器名称
func (h *HTTPChecker) Name() string {
	return h.name
}

// Check 执行HTTP检查
func (h *HTTPChecker) Check(ctx context.Context) *CheckResult {
	start := time.Now()
	result := &CheckResult{
		Name:      h.name,
		Timestamp: start,
		Metadata:  make(map[string]interface{}),
	}

	req, err := http.NewRequestWithContext(ctx, "GET", h.url, nil)
	if err != nil {
		result.Status = StatusUnhealthy
		result.Message = fmt.Sprintf("Failed to create request: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	resp, err := h.client.Do(req)
	if err != nil {
		result.Status = StatusUnhealthy
		result.Message = fmt.Sprintf("HTTP request failed: %v", err)
		result.Duration = time.Since(start)
		return result
	}
	defer resp.Body.Close()

	result.Metadata["status_code"] = resp.StatusCode
	result.Metadata["url"] = h.url

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result.Status = StatusHealthy
		result.Message = "HTTP service is healthy"
	} else if resp.StatusCode >= 500 {
		result.Status = StatusUnhealthy
		result.Message = fmt.Sprintf("HTTP service returned %d", resp.StatusCode)
	} else {
		result.Status = StatusDegraded
		result.Message = fmt.Sprintf("HTTP service returned %d", resp.StatusCode)
	}

	result.Duration = time.Since(start)
	return result
}

// CustomChecker 自定义检查器
type CustomChecker struct {
	name     string
	checkFn  func(ctx context.Context) (HealthStatus, string, map[string]interface{})
}

// NewCustomChecker 创建自定义检查器
func NewCustomChecker(name string, checkFn func(ctx context.Context) (HealthStatus, string, map[string]interface{})) *CustomChecker {
	return &CustomChecker{
		name:    name,
		checkFn: checkFn,
	}
}

// Name 返回检查器名称
func (c *CustomChecker) Name() string {
	return c.name
}

// Check 执行自定义检查
func (c *CustomChecker) Check(ctx context.Context) *CheckResult {
	start := time.Now()
	result := &CheckResult{
		Name:      c.name,
		Timestamp: start,
	}

	status, message, metadata := c.checkFn(ctx)
	result.Status = status
	result.Message = message
	result.Metadata = metadata
	result.Duration = time.Since(start)

	return result
}
