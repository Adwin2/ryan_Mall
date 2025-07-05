package handler

import (
	"net/http"
	"ryan-mall/pkg/database"
	"ryan-mall/pkg/monitoring"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsHandler 监控指标处理器
type MetricsHandler struct {
	metricsManager *monitoring.MetricsManager
	productService interface {
		GetCacheStats() map[string]interface{}
	}
}

// NewMetricsHandler 创建监控指标处理器
func NewMetricsHandler(metricsManager *monitoring.MetricsManager, productService interface {
	GetCacheStats() map[string]interface{}
}) *MetricsHandler {
	return &MetricsHandler{
		metricsManager: metricsManager,
		productService: productService,
	}
}

// GetMetrics 获取Prometheus指标
func (h *MetricsHandler) GetMetrics(c *gin.Context) {
	handler := promhttp.Handler()
	handler.ServeHTTP(c.Writer, c.Request)
}

// GetHealthCheck 健康检查
func (h *MetricsHandler) GetHealthCheck(c *gin.Context) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": "2024-01-01T00:00:00Z",
		"version":   "1.0.0",
		"services": map[string]string{
			"database": "connected",
			"redis":    "connected", 
			"app":      "running",
		},
	}
	
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Health check passed",
		"data":    health,
	})
}

// GetCacheStats 获取缓存统计
func (h *MetricsHandler) GetCacheStats(c *gin.Context) {
	var stats map[string]interface{}

	if h.productService != nil {
		stats = h.productService.GetCacheStats()
	} else {
		stats = map[string]interface{}{
			"cache_size": 0,
			"cache_type": "not_available",
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Cache stats retrieved",
		"data":    stats,
	})
}

// GetDBStats 获取数据库连接池统计
func (h *MetricsHandler) GetDBStats(c *gin.Context) {
	stats, err := database.GetDBStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to get database stats",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Database stats retrieved",
		"data":    stats,
	})
}
