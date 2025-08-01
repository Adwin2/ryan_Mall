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

	"ryan-mall-microservices/internal/product/application/service"
	"ryan-mall-microservices/internal/product/infrastructure/repository"
	productHttp "ryan-mall-microservices/internal/product/interfaces/http"
	"ryan-mall-microservices/internal/shared/events"
	"ryan-mall-microservices/internal/shared/infrastructure"
	"ryan-mall-microservices/pkg/health"
	"ryan-mall-microservices/pkg/monitoring"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
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

	// 初始化分布式锁
	distributedLock := infrastructure.NewRedisDistributedLock(redisClient)
	lockManager := infrastructure.NewLockManager(distributedLock)

	// 初始化事件发布器
	eventBus := events.NewInMemoryEventBus()
	eventPublisher := events.NewEventPublisher(eventBus, nil)

	// 初始化仓储（使用分布式锁版本）
	productRepo := repository.NewMySQLProductRepositoryWithLock(db, lockManager)

	// 初始化健康检查
	healthManager := health.NewHealthManager(30 * time.Second)
	healthManager.AddChecker(health.NewDatabaseChecker("mysql", db))
	// TODO: 添加Redis健康检查（需要解决版本兼容性问题）

	// 设置Gin模式
	if getEnv("ENVIRONMENT", "development") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建Gin引擎
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// 添加健康检查路由
	router.GET("/health", healthManager.Handler())

	// 添加监控指标路由
	prometheusMetrics := monitoring.NewPrometheusMetrics()
	router.Use(prometheusMetrics.PrometheusMiddleware())
	router.GET("/metrics", prometheusMetrics.MetricsHandler())

	// 初始化商品应用服务
	productAppSvc := service.NewProductApplicationService(
		productRepo,
		eventPublisher,
	)

	// 注册商品服务路由
	productHandler := productHttp.NewProductHandler(productAppSvc, logger)
	productHandler.RegisterRoutes(router)

	// 启动服务器
	port := getEnv("PRODUCT_SERVICE_PORT", "8082")
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: router,
	}

	// 优雅关闭
	go func() {
		logger.Info("Starting Product Service",
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

	logger.Info("Shutting down Product Service...")

	// 优雅关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Product Service forced to shutdown", infrastructure.Error(err))
	}

	logger.Info("Product Service exited")
}

// initDatabase 初始化数据库连接
func initDatabase() (*gorm.DB, error) {
	// MySQL配置
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		getEnv("DB_USER", "root"),
		getEnv("DB_PASSWORD", "123456"),
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

// initRedis 初始化Redis连接
func initRedis() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       0,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	return client
}

// autoMigrate 自动迁移数据库表
func autoMigrate(db *gorm.DB) error {
	// 定义数据库表结构
	type Product struct {
		ID          uint      `gorm:"primaryKey;autoIncrement"`
		ProductID   string    `gorm:"type:varchar(36);uniqueIndex;not null;comment:商品UUID"`
		Name        string    `gorm:"type:varchar(255);not null;comment:商品名称"`
		Description string    `gorm:"type:text;comment:商品描述"`
		CategoryID  string    `gorm:"type:varchar(36);not null;index;comment:分类ID"`
		Price       int64     `gorm:"not null;comment:价格（分）"`
		Stock       int       `gorm:"not null;default:0;comment:库存"`
		SalesCount  int       `gorm:"not null;default:0;comment:销量"`
		Status      int       `gorm:"not null;default:1;comment:状态：1-可用，0-不可用"`
		CreatedAt   time.Time `gorm:"autoCreateTime"`
		UpdatedAt   time.Time `gorm:"autoUpdateTime"`
	}

	// 执行自动迁移
	return db.AutoMigrate(&Product{})
}

// getEnv 获取环境变量
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
