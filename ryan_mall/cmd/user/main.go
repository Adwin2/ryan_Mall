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

	"ryan-mall-microservices/internal/shared/events"
	"ryan-mall-microservices/internal/shared/infrastructure"
	userService "ryan-mall-microservices/internal/user/application/service"
	userInfra "ryan-mall-microservices/internal/user/infrastructure/repository"
	userHttp "ryan-mall-microservices/internal/user/interfaces/http"
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
	// defer logger.Sync() // 简化版logger没有Sync方法

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

	// 初始化用户仓储
	userRepo := userInfra.NewMySQLUserRepository(db)

	// 初始化事件发布器
	eventBus := events.NewInMemoryEventBus()
	eventPublisher := events.NewEventPublisher(eventBus, nil)

	// 初始化用户应用服务
	userAppSvc := userService.NewUserApplicationService(
		userRepo,
		eventPublisher,
		getEnv("JWT_SECRET", "ryan-mall-secret-key"),
		24*time.Hour, // JWT过期时间
	)

	// 注册用户服务路由
	userHandler := userHttp.NewUserHandler(userAppSvc, logger)
	userHandler.RegisterRoutes(router)

	// 启动服务器
	port := getEnv("USER_SERVICE_PORT", "8081")
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: router,
	}

	// 优雅关闭
	go func() {
		logger.Info("Starting User Service",
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

	logger.Info("Shutting down User Service...")

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

	logger.Info("User Service stopped")
}

// initDatabase 初始化数据库连接
func initDatabase() (*gorm.DB, error) {
	// MySQL配置
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		getEnv("DB_USER", "root"),
		getEnv("DB_PASSWORD", ""),
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

	// 自动迁移数据库表
	if err := autoMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %v", err)
	}

	return db, nil
}

// autoMigrate 自动迁移数据库表
func autoMigrate(db *gorm.DB) error {
	// 定义数据库表结构
	type User struct {
		ID           uint       `gorm:"primaryKey;autoIncrement"`
		UserID       string     `gorm:"type:varchar(36);uniqueIndex;not null;comment:用户UUID"`
		Username     string     `gorm:"type:varchar(50);uniqueIndex;not null;comment:用户名"`
		Email        string     `gorm:"type:varchar(100);uniqueIndex;not null;comment:邮箱"`
		PasswordHash string     `gorm:"type:varchar(255);not null;comment:密码哈希"`
		Phone        string     `gorm:"type:varchar(20);comment:手机号"`
		Status       int8       `gorm:"type:tinyint;default:1;comment:状态：1-正常，0-禁用"`
		CreatedAt    *time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
		UpdatedAt    *time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
	}

	type UserProfile struct {
		ID        uint       `gorm:"primaryKey;autoIncrement"`
		UserID    string     `gorm:"type:varchar(36);uniqueIndex;not null;comment:用户UUID"`
		Nickname  string     `gorm:"type:varchar(50);comment:昵称"`
		AvatarURL string     `gorm:"type:varchar(255);comment:头像URL"`
		Gender    int8       `gorm:"type:tinyint;comment:性别：1-男，2-女，0-未知"`
		Birthday  *time.Time `gorm:"type:date;comment:生日"`
		Bio       string     `gorm:"type:text;comment:个人简介"`
		CreatedAt *time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
		UpdatedAt *time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
	}

	// 执行自动迁移
	return db.AutoMigrate(&User{}, &UserProfile{})
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
