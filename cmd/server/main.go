package main

import (
	"log"
	"net/http"
	"ryan-mall/internal/config"
	"ryan-mall/internal/handler"
	"ryan-mall/internal/middleware"
	"ryan-mall/internal/model"
	"ryan-mall/internal/repository"
	"ryan-mall/internal/service"
	"ryan-mall/pkg/cache"
	"ryan-mall/pkg/database"
	"ryan-mall/pkg/jwt"
	"ryan-mall/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. 加载配置
	// 这里加载应用程序的所有配置信息
	cfg := config.LoadConfig()

	// 2. 初始化数据库连接
	// 使用GORM连接MySQL数据库
	if err := database.InitMySQL(cfg); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer database.Close() // 程序退出时关闭数据库连接

	// 3. 自动迁移数据库表结构
	// GORM会根据模型结构自动创建和更新表
	if err := database.AutoMigrate(
		&model.User{},
		&model.Category{},
		&model.Product{},
		&model.CartItem{},
		&model.Order{},
		&model.OrderItem{},
	); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// 4. 初始化缓存系统 - 必须在创建服务之前
	// 优化为16分片，减少哈希计算开销
	cache.SetGlobalCache(cache.NewShardedCache(16))
	log.Println("✅ 分片缓存系统初始化完成 (16分片，性能优化)")

	// 5. 初始化Redis集群（如果启用）
	var redisCluster interface{}
	if cfg.Redis.ClusterEnabled && len(cfg.Redis.ClusterNodes) > 0 {
		log.Println("🔗 初始化Redis集群...")
		// 这里暂时注释掉，因为需要先启动集群
		// redisCluster = redisManager.NewClusterManager(cfg.Redis.ClusterNodes, cfg.Redis.Password)
		// log.Printf("✅ Redis集群初始化完成，节点数: %d", len(cfg.Redis.ClusterNodes))
		log.Println("⚠️  Redis集群配置已检测到，但暂未启用（需要先启动集群）")
	} else {
		log.Println("📝 使用内存缓存模式（Redis集群未启用）")
	}
	_ = redisCluster // 避免未使用变量警告

	// 5. 初始化依赖组件
	// 创建JWT管理器
	jwtManager := jwt.NewJWTManager(cfg.JWT.SecretKey, cfg.JWT.ExpireHours)

	// 创建数据访问层
	userRepo := repository.NewUserRepository(database.GetDB())
	productRepo := repository.NewProductRepository(database.GetDB())
	categoryRepo := repository.NewCategoryRepository(database.GetDB())
	cartRepo := repository.NewCartRepository(database.GetDB())
	orderRepo := repository.NewOrderRepository(database.GetDB())

	// 创建业务逻辑层
	userService := service.NewUserService(userRepo, jwtManager)
	// 使用带缓存的商品服务
	productService := service.NewCachedProductService(productRepo, categoryRepo)
	categoryService := service.NewCategoryService(categoryRepo)
	cartService := service.NewCartService(cartRepo, productRepo)
	orderService := service.NewOrderService(orderRepo, cartRepo, productRepo, database.GetDB())

	// 创建HTTP处理器
	userHandler := handler.NewUserHandler(userService)
	productHandler := handler.NewProductHandler(productService, categoryService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	cartHandler := handler.NewCartHandler(cartService)
	orderHandler := handler.NewOrderHandler(orderService)

	// 创建监控处理器
	metricsHandler := handler.NewMetricsHandler(nil, productService)

	// 创建中间件
	authMiddleware := middleware.NewAuthMiddleware(userService)

	// 5. 设置Gin运行模式
	// debug: 开发模式，会输出详细日志
	// release: 生产模式，性能更好
	gin.SetMode(cfg.Server.Mode)

	// 6. 创建Gin引擎
	// Gin是一个高性能的HTTP Web框架
	r := gin.Default()

	// 添加CORS中间件
	r.Use(middleware.CORS())

	// 6. 设置基础路由
	// 这是一个健康检查接口，用于确认服务是否正常运行
	r.GET("/ping", func(c *gin.Context) {
		response.Success(c, gin.H{
			"message": "pong",
			"version": "1.0.0",
		})
	})

	// 监控相关路由
	r.GET("/health", metricsHandler.GetHealthCheck)
	r.GET("/cache/stats", metricsHandler.GetCacheStats)
	r.GET("/db/stats", metricsHandler.GetDBStats)
	
	// 7. 设置API路由组
	// 使用路由组可以为一组路由添加统一的前缀和中间件
	v1 := r.Group("/api/v1")
	{
		// 注册用户相关路由
		userHandler.RegisterRoutes(v1, authMiddleware)

		// 注册商品相关路由
		productHandler.RegisterRoutes(v1, authMiddleware)

		// 注册分类相关路由
		categoryHandler.RegisterRoutes(v1, authMiddleware)

		// 注册购物车相关路由
		cartHandler.RegisterRoutes(v1, authMiddleware)

		// 注册订单相关路由
		orderHandler.RegisterRoutes(v1, authMiddleware)

		// 临时测试路由
		v1.GET("/test", func(c *gin.Context) {
			response.Success(c, gin.H{
				"message": "API v1 is working!",
			})
		})
	}
	
	// 8. 启动服务器 - 优化HTTP服务器配置
	// 监听指定端口，开始处理HTTP请求
	log.Printf("Server starting on port %s", cfg.Server.Port)
	log.Printf("Health check: http://localhost:%s/ping", cfg.Server.Port)
	log.Printf("API test: http://localhost:%s/api/v1/test", cfg.Server.Port)

	// 创建优化的HTTP服务器配置 - 针对高并发优化
	server := &http.Server{
		Addr:           ":" + cfg.Server.Port,
		Handler:        r,
		ReadTimeout:    5 * time.Second,   // 减少读取超时
		WriteTimeout:   5 * time.Second,   // 减少写入超时
		IdleTimeout:    30 * time.Second,  // 减少空闲超时
		MaxHeaderBytes: 1 << 16,           // 减少最大请求头大小 64KB
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
