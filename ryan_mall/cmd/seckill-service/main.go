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

	"ryan-mall-microservices/internal/shared/infrastructure"
	seckillHttp "ryan-mall-microservices/internal/seckill/interfaces/http"
	"ryan-mall-microservices/pkg/health"
	"ryan-mall-microservices/pkg/monitoring"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 初始化日志
	logger, err := infrastructure.NewLogger(getEnv("LOG_LEVEL", "info"))
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}

	// 初始化数据库
	db, err := initDatabase()
	if err != nil {
		logger.Fatal("Failed to connect to database", infrastructure.Error(err))
	}

	// 初始化Redis
	redisClient := initRedis()

	// 初始化监控
	metrics := monitoring.NewPrometheusMetrics()

	// 初始化健康检查
	healthManager := health.NewHealthManager(30 * time.Second)
	healthManager.AddChecker(health.NewDatabaseChecker("mysql", db))
	healthManager.AddChecker(health.NewRedisChecker("redis", redisClient))

	// 设置Gin模式
	if getEnv("ENVIRONMENT", "development") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建Gin引擎
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// 添加监控中间件
	router.Use(metrics.PrometheusMiddleware())

	// 健康检查端点
	router.GET("/health", healthManager.Handler())

	// 监控指标端点
	router.GET("/metrics", metrics.MetricsHandler())

	// 注册秒杀服务路由
	seckillHandler := seckillHttp.NewSeckillHandler(nil, logger) // 这里需要传入实际的应用服务
	seckillHandler.RegisterRoutes(router)

	// 启动服务器
	port := getEnv("SECKILL_SERVICE_PORT", "8084")
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: router,
	}

	// 优雅关闭
	go func() {
		logger.Info("Starting Seckill Service",
			infrastructure.String("port", port),
			infrastructure.String("environment", getEnv("ENVIRONMENT", "development")),
		)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", infrastructure.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Seckill Service...")

	// 优雅关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", infrastructure.Error(err))
	}

	// 关闭数据库连接
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.Close()
	}

	// 关闭Redis连接
	if err := redisClient.Close(); err != nil {
		logger.Error("Failed to close Redis connection", infrastructure.Error(err))
	}

	logger.Info("Seckill Service stopped")
}

// initDatabase 初始化数据库连接
func initDatabase() (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		getEnv("DB_USER", "ryan_mall"),
		getEnv("DB_PASSWORD", "ryan_mall123"),
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "3306"),
		getEnv("DB_NAME", "ryan_mall"),
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

// initRedis 初始化Redis连接
func initRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     getEnv("REDIS_ADDRESS", "localhost:6379"),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       0,
	})
}

// getEnv 获取环境变量
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
