package proxy

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"ryan-mall-microservices/internal/shared/infrastructure"
	"ryan-mall-microservices/pkg/discovery"

	"github.com/gin-gonic/gin"
)

// ProxyManager 代理管理器
type ProxyManager struct {
	serviceDiscovery discovery.ServiceDiscovery
	loadBalancer     discovery.LoadBalancer
	proxies          map[string]*httputil.ReverseProxy
	proxiesMutex     sync.RWMutex
	logger           infrastructure.Logger
	circuitBreakers  map[string]*CircuitBreaker
	cbMutex          sync.RWMutex
}

// NewProxyManager 创建代理管理器
func NewProxyManager(serviceDiscovery discovery.ServiceDiscovery, logger infrastructure.Logger) *ProxyManager {
	return &ProxyManager{
		serviceDiscovery: serviceDiscovery,
		loadBalancer:     discovery.NewRoundRobinLoadBalancer(),
		proxies:          make(map[string]*httputil.ReverseProxy),
		logger:           logger,
		circuitBreakers:  make(map[string]*CircuitBreaker),
	}
}

// ProxyHandler 代理处理器
func (pm *ProxyManager) ProxyHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 解析路径获取服务名
		serviceName := pm.extractServiceName(c.Request.URL.Path)
		if serviceName == "" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "service not found",
				"path":  c.Request.URL.Path,
			})
			return
		}

		// 检查熔断器状态
		if pm.isCircuitBreakerOpen(serviceName) {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "service temporarily unavailable",
				"service": serviceName,
			})
			return
		}

		// 发现服务实例
		services, err := pm.serviceDiscovery.Discover(c.Request.Context(), serviceName)
		if err != nil {
			pm.logger.Error("Failed to discover service",
				infrastructure.String("service", serviceName),
				infrastructure.Error(err),
			)
			pm.recordFailure(serviceName)
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "service discovery failed",
				"service": serviceName,
			})
			return
		}

		if len(services) == 0 {
			pm.logger.Warn("No available service instances",
				infrastructure.String("service", serviceName),
			)
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "no available service instances",
				"service": serviceName,
			})
			return
		}

		// 负载均衡选择服务实例
		selectedService := pm.loadBalancer.Select(services)
		if selectedService == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "load balancer failed to select service",
				"service": serviceName,
			})
			return
		}

		// 获取或创建代理
		proxy := pm.getOrCreateProxy(selectedService)

		// 修改请求路径
		originalPath := c.Request.URL.Path
		c.Request.URL.Path = pm.rewritePath(originalPath, serviceName)

		// 设置请求头
		c.Request.Header.Set("X-Forwarded-For", c.ClientIP())
		c.Request.Header.Set("X-Forwarded-Proto", "http")
		c.Request.Header.Set("X-Gateway-Service", serviceName)

		// 记录请求开始时间
		startTime := time.Now()

		// 代理请求
		proxy.ServeHTTP(c.Writer, c.Request)

		// 记录请求耗时
		duration := time.Since(startTime)
		pm.logger.Info("Request proxied",
			infrastructure.String("service", serviceName),
			infrastructure.String("path", originalPath),
			infrastructure.String("target", selectedService.Address),
			infrastructure.Duration("duration", duration),
		)

		// 记录成功
		pm.recordSuccess(serviceName)
	}
}

// extractServiceName 从路径中提取服务名
func (pm *ProxyManager) extractServiceName(path string) string {
	// 路径格式: /api/v1/{service}/...
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 3 && parts[0] == "api" && parts[1] == "v1" {
		switch parts[2] {
		case "users":
			return "user-service"
		case "products":
			return "product-service"
		case "orders":
			return "order-service"
		case "seckill":
			return "seckill-service"
		case "payments":
			return "payment-service"
		}
	}
	return ""
}

// rewritePath 重写请求路径
func (pm *ProxyManager) rewritePath(originalPath, serviceName string) string {
	// 移除网关前缀，保留服务路径
	// 例如: /api/v1/users/login -> /api/v1/users/login
	return originalPath
}

// getOrCreateProxy 获取或创建代理
func (pm *ProxyManager) getOrCreateProxy(service *discovery.ServiceInfo) *httputil.ReverseProxy {
	target := fmt.Sprintf("http://%s:%d", service.Address, service.Port)
	
	pm.proxiesMutex.RLock()
	proxy, exists := pm.proxies[target]
	pm.proxiesMutex.RUnlock()

	if exists {
		return proxy
	}

	pm.proxiesMutex.Lock()
	defer pm.proxiesMutex.Unlock()

	// 双重检查
	if proxy, exists := pm.proxies[target]; exists {
		return proxy
	}

	// 创建新的代理
	targetURL, _ := url.Parse(target)
	proxy = httputil.NewSingleHostReverseProxy(targetURL)

	// 自定义错误处理
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		pm.logger.Error("Proxy error",
			infrastructure.String("target", target),
			infrastructure.String("path", r.URL.Path),
			infrastructure.Error(err),
		)

		// 记录失败
		serviceName := pm.extractServiceName(r.URL.Path)
		if serviceName != "" {
			pm.recordFailure(serviceName)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(`{"error": "bad gateway", "message": "service temporarily unavailable"}`))
	}

	// 自定义请求修改
	proxy.ModifyResponse = func(resp *http.Response) error {
		// 添加响应头
		resp.Header.Set("X-Gateway", "ryan-mall-gateway")
		resp.Header.Set("X-Gateway-Version", "1.0.0")
		return nil
	}

	pm.proxies[target] = proxy
	return proxy
}

// isCircuitBreakerOpen 检查熔断器是否打开
func (pm *ProxyManager) isCircuitBreakerOpen(serviceName string) bool {
	pm.cbMutex.RLock()
	cb, exists := pm.circuitBreakers[serviceName]
	pm.cbMutex.RUnlock()

	if !exists {
		return false
	}

	return cb.IsOpen()
}

// recordSuccess 记录成功
func (pm *ProxyManager) recordSuccess(serviceName string) {
	pm.cbMutex.Lock()
	defer pm.cbMutex.Unlock()

	cb, exists := pm.circuitBreakers[serviceName]
	if !exists {
		cb = NewCircuitBreaker(CircuitBreakerConfig{
			FailureThreshold: 5,
			RecoveryTimeout:  60 * time.Second,
			SuccessThreshold: 3,
		})
		pm.circuitBreakers[serviceName] = cb
	}

	cb.RecordSuccess()
}

// recordFailure 记录失败
func (pm *ProxyManager) recordFailure(serviceName string) {
	pm.cbMutex.Lock()
	defer pm.cbMutex.Unlock()

	cb, exists := pm.circuitBreakers[serviceName]
	if !exists {
		cb = NewCircuitBreaker(CircuitBreakerConfig{
			FailureThreshold: 5,
			RecoveryTimeout:  60 * time.Second,
			SuccessThreshold: 3,
		})
		pm.circuitBreakers[serviceName] = cb
	}

	cb.RecordFailure()
}

// HealthCheck 健康检查代理
func (pm *ProxyManager) HealthCheck(serviceName string) error {
	services, err := pm.serviceDiscovery.Discover(context.Background(), serviceName)
	if err != nil {
		return fmt.Errorf("service discovery failed: %w", err)
	}

	if len(services) == 0 {
		return fmt.Errorf("no available service instances")
	}

	// 检查第一个可用的服务实例
	service := services[0]
	healthURL := fmt.Sprintf("http://%s:%d/health", service.Address, service.Port)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(healthURL)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("health check failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetServiceStats 获取服务统计信息
func (pm *ProxyManager) GetServiceStats() map[string]interface{} {
	pm.cbMutex.RLock()
	defer pm.cbMutex.RUnlock()

	stats := make(map[string]interface{})
	for serviceName, cb := range pm.circuitBreakers {
		stats[serviceName] = map[string]interface{}{
			"state":         cb.State(),
			"failure_count": cb.FailureCount(),
			"success_count": cb.SuccessCount(),
			"last_failure":  cb.LastFailureTime(),
		}
	}

	return stats
}
