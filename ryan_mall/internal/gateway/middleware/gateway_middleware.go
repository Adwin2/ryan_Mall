package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"ryan-mall-microservices/internal/shared/infrastructure"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// CORSMiddleware CORS中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// AuthConfig 认证配置
type AuthConfig struct {
	JWTSecret     string
	SkipPaths     []string
	RedisClient   *redis.Client
	TokenExpiry   time.Duration
	RefreshExpiry time.Duration
}

// AuthMiddleware JWT认证中间件
func AuthMiddleware(config *AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否跳过认证
		if shouldSkipAuth(c.Request.URL.Path, config.SkipPaths) {
			c.Next()
			return
		}

		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header required",
			})
			c.Abort()
			return
		}

		// 检查Bearer前缀
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization header format",
			})
			c.Abort()
			return
		}

		// 提取token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// 验证JWT token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(config.JWTSecret), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token",
			})
			c.Abort()
			return
		}

		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "token is not valid",
			})
			c.Abort()
			return
		}

		// 提取claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token claims",
			})
			c.Abort()
			return
		}

		// 检查token是否在Redis黑名单中
		if config.RedisClient != nil {
			blacklistKey := fmt.Sprintf("blacklist:token:%s", tokenString)
			exists, err := config.RedisClient.Exists(c.Request.Context(), blacklistKey).Result()
			if err == nil && exists > 0 {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "token has been revoked",
				})
				c.Abort()
				return
			}
		}

		// 设置用户信息到上下文
		if userID, ok := claims["user_id"].(string); ok {
			c.Set("user_id", userID)
		}
		if username, ok := claims["username"].(string); ok {
			c.Set("username", username)
		}
		if email, ok := claims["email"].(string); ok {
			c.Set("email", email)
		}

		c.Next()
	}
}

// RequestTracingMiddleware 请求追踪中间件
func RequestTracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 生成请求ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// 设置请求ID到响应头和上下文
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)

		// 记录请求开始时间
		startTime := time.Now()

		// 记录请求信息
		logger := infrastructure.GetLogger()
		logger.Info("Request started",
			infrastructure.String("request_id", requestID),
			infrastructure.String("method", c.Request.Method),
			infrastructure.String("path", c.Request.URL.Path),
			infrastructure.String("query", c.Request.URL.RawQuery),
			infrastructure.String("user_agent", c.Request.UserAgent()),
			infrastructure.String("client_ip", c.ClientIP()),
		)

		c.Next()

		// 记录请求完成信息
		duration := time.Since(startTime)
		logger.Info("Request completed",
			infrastructure.String("request_id", requestID),
			infrastructure.String("method", c.Request.Method),
			infrastructure.String("path", c.Request.URL.Path),
			infrastructure.Int("status", c.Writer.Status()),
			infrastructure.Duration("duration", duration),
			infrastructure.Int("response_size", c.Writer.Size()),
		)
	}
}

// SecurityMiddleware 安全中间件
func SecurityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置安全头
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// 移除服务器信息
		c.Header("Server", "")

		c.Next()
	}
}

// TimeoutMiddleware 超时中间件
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 简化的超时实现
		// 在实际项目中，可以使用context.WithTimeout
		c.Next()
	}
}

// CompressionMiddleware 压缩中间件
func CompressionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查客户端是否支持gzip
		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		// 检查内容类型是否需要压缩
		contentType := c.GetHeader("Content-Type")
		if !shouldCompress(contentType) {
			c.Next()
			return
		}

		// 设置压缩头
		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")

		c.Next()
	}
}

// MetricsMiddleware 指标中间件
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		c.Next()

		// 记录指标
		duration := time.Since(startTime)
		method := c.Request.Method
		path := c.Request.URL.Path
		status := c.Writer.Status()

		// 这里可以集成Prometheus等指标系统
		logger := infrastructure.GetLogger()
		logger.Debug("Request metrics",
			infrastructure.String("method", method),
			infrastructure.String("path", path),
			infrastructure.Int("status", status),
			infrastructure.Duration("duration", duration),
		)
	}
}

// shouldSkipAuth 检查是否跳过认证
func shouldSkipAuth(path string, skipPaths []string) bool {
	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// shouldCompress 检查是否需要压缩
func shouldCompress(contentType string) bool {
	compressibleTypes := []string{
		"application/json",
		"application/xml",
		"text/html",
		"text/css",
		"text/javascript",
		"text/plain",
	}

	for _, t := range compressibleTypes {
		if strings.Contains(contentType, t) {
			return true
		}
	}
	return false
}

// ErrorHandlerMiddleware 错误处理中间件
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		logger := infrastructure.GetLogger()
		
		if err, ok := recovered.(string); ok {
			logger.Error("Panic recovered",
				infrastructure.String("error", err),
				infrastructure.String("path", c.Request.URL.Path),
				infrastructure.String("method", c.Request.Method),
			)
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
	})
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	GlobalLimit int
	UserLimit   int
	IPLimit     int
	Manager     interface{} // 简化的限流管理器接口
}

// DefaultRateLimitConfig 默认限流配置
func DefaultRateLimitConfig(manager interface{}) *RateLimitConfig {
	return &RateLimitConfig{
		GlobalLimit: 10000,
		UserLimit:   100,
		IPLimit:     1000,
		Manager:     manager,
	}
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(config *RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 简化的限流实现
		// 在实际项目中，这里应该使用Redis或其他存储来实现分布式限流

		// 获取客户端IP
		clientIP := c.ClientIP()

		// 这里可以实现基于IP、用户、全局的限流逻辑
		// 为了简化，我们暂时跳过实际的限流检查

		// 设置限流相关的响应头
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", config.IPLimit))
		c.Header("X-RateLimit-Remaining", "999") // 简化的剩余次数
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Hour).Unix()))

		// 记录请求信息
		logger := infrastructure.GetLogger()
		logger.Debug("Rate limit check",
			infrastructure.String("client_ip", clientIP),
			infrastructure.String("path", c.Request.URL.Path),
		)

		c.Next()
	}
}
