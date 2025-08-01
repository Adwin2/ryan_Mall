package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ryan-mall-microservices/internal/gateway/config"
	"ryan-mall-microservices/internal/gateway/middleware"
	"ryan-mall-microservices/internal/gateway/proxy"
	"ryan-mall-microservices/internal/shared/infrastructure"
	"ryan-mall-microservices/pkg/discovery"
	"ryan-mall-microservices/pkg/monitoring"
	"ryan-mall-microservices/pkg/ratelimiter"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func main() {
	// 加载配置
	cfg, err := config.LoadGatewayConfig()
	if err != nil {
		log.Fatalf("Failed to load gateway config: %v", err)
	}

	// 初始化日志
	logger, err := infrastructure.NewLogger(cfg.LogLevel)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}

	// 初始化Redis客户端
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// 测试Redis连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Fatal("Failed to connect to Redis", infrastructure.Error(err))
	}

	// 初始化服务发现（使用内存版本用于开发）
	serviceDiscovery := discovery.NewMemoryServiceDiscovery()
	serviceDiscovery.PreregisterServices() // 预注册服务
	defer serviceDiscovery.Close()

	// 初始化限流器管理器
	rateLimiterManager := ratelimiter.NewRateLimiterManager(redisClient)

	// 初始化监控指标
	prometheusMetrics := monitoring.NewPrometheusMetrics()

	// 初始化代理管理器
	proxyManager := proxy.NewProxyManager(serviceDiscovery, logger)

	// 设置Gin模式
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建Gin引擎
	router := gin.New()

	// 添加中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// 添加监控中间件
	router.Use(prometheusMetrics.PrometheusMiddleware())

	// 添加CORS中间件
	router.Use(middleware.CORSMiddleware())

	// 添加限流中间件
	rateLimitConfig := middleware.DefaultRateLimitConfig(rateLimiterManager)
	router.Use(middleware.RateLimitMiddleware(rateLimitConfig))

	// 添加认证中间件
	authConfig := &middleware.AuthConfig{
		JWTSecret:     cfg.JWT.Secret,
		SkipPaths:     []string{"/health", "/metrics", "/gateway/services", "/api/v1/users/login", "/api/v1/users/register"},
		RedisClient:   redisClient,
		TokenExpiry:   time.Duration(cfg.JWT.ExpiryHours) * time.Hour,
		RefreshExpiry: time.Duration(cfg.JWT.RefreshExpiryHours) * time.Hour,
	}
	router.Use(middleware.AuthMiddleware(authConfig))

	// 添加请求追踪中间件
	router.Use(middleware.RequestTracingMiddleware())

	// 健康检查端点
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
			"version":   "1.0.0",
		})
	})

	// 监控指标端点
	router.GET("/metrics", prometheusMetrics.MetricsHandler())

	// 服务发现端点
	router.GET("/gateway/services", func(c *gin.Context) {
		// 这里可以返回已注册的服务列表
		c.JSON(http.StatusOK, gin.H{
			"message": "Service discovery endpoint",
		})
	})

	// 代理所有API请求
	router.Any("/api/*path", proxyManager.ProxyHandler())

	// 启动服务器
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router,
	}

	// 优雅关闭
	go func() {
		logger.Info("Starting API Gateway",
			infrastructure.Int("port", cfg.Port),
			infrastructure.String("environment", cfg.Environment),
		)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", infrastructure.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down API Gateway...")

	// 优雅关闭服务器
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", infrastructure.Error(err))
	}

	// 关闭Redis连接
	if err := redisClient.Close(); err != nil {
		logger.Error("Failed to close Redis connection", infrastructure.Error(err))
	}

	logger.Info("API Gateway stopped")
}
