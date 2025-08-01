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
	"ryan-mall-microservices/internal/user/application/service"
	"ryan-mall-microservices/internal/user/infrastructure/repository"
	userHttp "ryan-mall-microservices/internal/user/interfaces/http"
	"ryan-mall-microservices/pkg/config"
	"ryan-mall-microservices/pkg/database"
	"ryan-mall-microservices/pkg/redis"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	if err := infrastructure.InitGlobalLogger("info"); err != nil {
		log.Fatalf("Failed to init logger: %v", err)
	}
	logger := infrastructure.GetLogger()

	// 连接数据库
	dbConn, err := database.NewMySQLConnection(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", infrastructure.Error(err))
	}
	defer dbConn.Close()

	// 自动迁移表结构
	if err := dbConn.AutoMigrate(
		&repository.UserPO{},
		&repository.UserProfilePO{},
	); err != nil {
		logger.Fatal("Failed to migrate database", infrastructure.Error(err))
	}

	// 连接Redis
	redisClient, err := redis.NewRedisClient(&cfg.Redis)
	if err != nil {
		logger.Fatal("Failed to connect to redis", infrastructure.Error(err))
	}
	defer redisClient.Close()

	// 创建事件总线和发布器
	eventBus := events.NewInMemoryEventBus()
	eventPublisher := events.NewEventPublisher(eventBus, nil)

	// 创建仓储
	userRepo := repository.NewMySQLUserRepository(dbConn.GetDB())

	// 创建应用服务
	userAppSvc := service.NewUserApplicationService(
		userRepo,
		eventPublisher,
		cfg.JWT.Secret,
		cfg.JWT.ExpireTime,
	)

	// 创建HTTP处理器
	userHandler := userHttp.NewUserHandler(userAppSvc, logger)

	// 设置Gin模式
	if os.Getenv("ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// 添加CORS中间件
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "user-service",
			"timestamp": time.Now().Unix(),
		})
	})

	// 注册用户路由
	userHandler.RegisterRoutes(router)

	// 创建HTTP服务器
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// 启动服务器
	go func() {
		logger.Info("Starting user service",
			infrastructure.String("address", server.Addr),
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", infrastructure.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down user service...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", infrastructure.Error(err))
	}

	logger.Info("User service stopped")
}
