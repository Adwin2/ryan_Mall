package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"ryan-mall-microservices/internal/shared/infrastructure"
	"ryan-mall-microservices/pkg/ratelimiter"

	"github.com/gin-gonic/gin"
)

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	// 限流器管理器
	Manager *ratelimiter.RateLimiterManager
	
	// 限流策略
	Strategy RateLimitStrategy
	
	// 限流规则
	Rules []RateLimitRule
	
	// 跳过限流的路径
	SkipPaths []string
	
	// 错误处理器
	ErrorHandler func(c *gin.Context, err error)
	
	// 限流触发处理器
	LimitHandler func(c *gin.Context, key string, remaining int)
}

// RateLimitStrategy 限流策略
type RateLimitStrategy string

const (
	StrategyUser   RateLimitStrategy = "user"   // 按用户限流
	StrategyIP     RateLimitStrategy = "ip"     // 按IP限流
	StrategyAPI    RateLimitStrategy = "api"    // 按API限流
	StrategyGlobal RateLimitStrategy = "global" // 全局限流
)

// RateLimitRule 限流规则
type RateLimitRule struct {
	// 路径模式（支持通配符）
	PathPattern string
	
	// HTTP方法
	Methods []string
	
	// 限流类型
	Type LimiterType
	
	// 限流参数
	Rate     int           // 速率
	Capacity int           // 容量（令牌桶）
	Interval time.Duration // 时间间隔
	
	// 是否启用
	Enabled bool
}

// LimiterType 限流器类型
type LimiterType string

const (
	LimiterTypeSlidingWindow LimiterType = "sliding_window"
	LimiterTypeTokenBucket   LimiterType = "token_bucket"
)

// RateLimitMiddleware 创建限流中间件
func RateLimitMiddleware(config *RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否跳过限流
		if shouldSkip(c.Request.URL.Path, config.SkipPaths) {
			c.Next()
			return
		}

		// 查找匹配的限流规则
		rule := findMatchingRule(c, config.Rules)
		if rule == nil || !rule.Enabled {
			c.Next()
			return
		}

		// 生成限流键
		key, err := generateRateLimitKey(c, config.Strategy, rule)
		if err != nil {
			if config.ErrorHandler != nil {
				config.ErrorHandler(c, err)
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to generate rate limit key",
				})
			}
			c.Abort()
			return
		}

		// 获取限流器
		limiter := getLimiter(config.Manager, rule)

		// 检查限流
		allowed, err := limiter.Allow(c.Request.Context(), key)
		if err != nil {
			if config.ErrorHandler != nil {
				config.ErrorHandler(c, err)
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Rate limit check failed",
				})
			}
			c.Abort()
			return
		}

		if !allowed {
			// 获取剩余请求数
			remaining, _ := limiter.GetRemaining(c.Request.Context(), key)
			
			if config.LimitHandler != nil {
				config.LimitHandler(c, key, remaining)
			} else {
				c.Header("X-RateLimit-Limit", strconv.Itoa(rule.Rate))
				c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
				c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(rule.Interval).Unix(), 10))
				
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":     "Rate limit exceeded",
					"message":   "Too many requests, please try again later",
					"remaining": remaining,
					"reset_at":  time.Now().Add(rule.Interval).Unix(),
				})
			}
			c.Abort()
			return
		}

		// 设置响应头
		remaining, _ := limiter.GetRemaining(c.Request.Context(), key)
		c.Header("X-RateLimit-Limit", strconv.Itoa(rule.Rate))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(rule.Interval).Unix(), 10))

		c.Next()
	}
}

// shouldSkip 检查是否应该跳过限流
func shouldSkip(path string, skipPaths []string) bool {
	for _, skipPath := range skipPaths {
		if matchPath(path, skipPath) {
			return true
		}
	}
	return false
}

// findMatchingRule 查找匹配的限流规则
func findMatchingRule(c *gin.Context, rules []RateLimitRule) *RateLimitRule {
	path := c.Request.URL.Path
	method := c.Request.Method

	for _, rule := range rules {
		// 检查路径是否匹配
		if !matchPath(path, rule.PathPattern) {
			continue
		}

		// 检查方法是否匹配
		if len(rule.Methods) > 0 && !contains(rule.Methods, method) {
			continue
		}

		return &rule
	}

	return nil
}

// generateRateLimitKey 生成限流键
func generateRateLimitKey(c *gin.Context, strategy RateLimitStrategy, rule *RateLimitRule) (string, error) {
	action := fmt.Sprintf("%s:%s", c.Request.Method, c.Request.URL.Path)

	switch strategy {
	case StrategyUser:
		userID := getUserID(c)
		if userID == "" {
			return "", fmt.Errorf("user ID not found")
		}
		return ratelimiter.GenerateUserKey(userID, action), nil

	case StrategyIP:
		ip := getClientIP(c)
		return ratelimiter.GenerateIPKey(ip, action), nil

	case StrategyAPI:
		return ratelimiter.GenerateAPIKey(c.Request.URL.Path, c.Request.Method), nil

	case StrategyGlobal:
		return fmt.Sprintf("rate_limit:global:%s", action), nil

	default:
		return "", fmt.Errorf("unknown rate limit strategy: %s", strategy)
	}
}

// getLimiter 获取限流器
func getLimiter(manager *ratelimiter.RateLimiterManager, rule *RateLimitRule) ratelimiter.RateLimiter {
	switch rule.Type {
	case LimiterTypeTokenBucket:
		return manager.GetTokenBucketLimiter(
			rule.PathPattern,
			rule.Capacity,
			rule.Rate,
			rule.Interval,
		)
	default: // 默认使用滑动窗口
		return manager.GetSlidingWindowLimiter(
			rule.PathPattern,
			rule.Rate,
			rule.Interval,
		)
	}
}

// getUserID 从上下文获取用户ID
func getUserID(c *gin.Context) string {
	// 从JWT token中获取用户ID
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			return id
		}
	}

	// 从Authorization header中解析
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		// 这里可以解析JWT token获取用户ID
		// 简化处理，实际项目中需要完整的JWT解析
		return "anonymous"
	}

	return ""
}

// getClientIP 获取客户端IP
func getClientIP(c *gin.Context) string {
	// 检查X-Forwarded-For头
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 检查X-Real-IP头
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return xri
	}

	// 使用RemoteAddr
	return c.ClientIP()
}

// matchPath 检查路径是否匹配模式
func matchPath(path, pattern string) bool {
	// 简单的通配符匹配，实际项目中可以使用更复杂的匹配算法
	if pattern == "*" {
		return true
	}

	if pattern == path {
		return true
	}

	// 支持前缀匹配
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(path, prefix)
	}

	return false
}

// contains 检查切片是否包含元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// DefaultRateLimitConfig 默认限流配置
func DefaultRateLimitConfig(manager *ratelimiter.RateLimiterManager) *RateLimitConfig {
	return &RateLimitConfig{
		Manager:  manager,
		Strategy: StrategyUser,
		Rules: []RateLimitRule{
			{
				PathPattern: "/api/v1/users/login",
				Methods:     []string{"POST"},
				Type:        LimiterTypeSlidingWindow,
				Rate:        5,
				Interval:    time.Minute,
				Enabled:     true,
			},
			{
				PathPattern: "/api/v1/users/register",
				Methods:     []string{"POST"},
				Type:        LimiterTypeSlidingWindow,
				Rate:        3,
				Interval:    time.Minute,
				Enabled:     true,
			},
			{
				PathPattern: "/api/v1/orders",
				Methods:     []string{"POST"},
				Type:        LimiterTypeTokenBucket,
				Rate:        10,
				Capacity:    20,
				Interval:    time.Minute,
				Enabled:     true,
			},
			{
				PathPattern: "/api/v1/*",
				Methods:     []string{"GET", "POST", "PUT", "DELETE"},
				Type:        LimiterTypeSlidingWindow,
				Rate:        100,
				Interval:    time.Minute,
				Enabled:     true,
			},
		},
		SkipPaths: []string{
			"/health",
			"/metrics",
			"/api/v1/ping",
		},
		ErrorHandler: func(c *gin.Context, err error) {
			infrastructure.GetLogger().Error("Rate limit error",
				infrastructure.String("path", c.Request.URL.Path),
				infrastructure.String("method", c.Request.Method),
				infrastructure.Error(err),
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
			})
		},
		LimitHandler: func(c *gin.Context, key string, remaining int) {
			infrastructure.GetLogger().Warn("Rate limit exceeded",
				infrastructure.String("key", key),
				infrastructure.String("path", c.Request.URL.Path),
				infrastructure.String("method", c.Request.Method),
				infrastructure.String("ip", getClientIP(c)),
				infrastructure.Int("remaining", remaining),
			)
		},
	}
}
