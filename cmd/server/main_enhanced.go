package main

import (
	"log"
	"ryan-mall/internal/config"
	"ryan-mall/internal/handler"
	"ryan-mall/internal/middleware"
	"ryan-mall/internal/model"
	"ryan-mall/internal/repository"
	"ryan-mall/internal/service"
	"ryan-mall/pkg/database"
	"ryan-mall/pkg/jwt"
	"ryan-mall/pkg/monitoring"
	"ryan-mall/pkg/response"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. 加载配置
	cfg := config.LoadConfig()

	// 2. 初始化监控系统
	metricsManager := monitoring.NewMetricsManager()
	businessMetrics := monitoring.NewBusinessMetrics(metricsManager)
	
	log.Println("✅ 监控系统初始化完成")

	// 3. 初始化数据库连接
	if err := database.InitMySQL(cfg); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer database.Close()
	
	log.Println("✅ 数据库连接初始化完成")

	// 4. 自动迁移数据库表结构
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
	
	log.Println("✅ 数据库迁移完成")

	// 5. 初始化依赖组件
	jwtManager := jwt.NewJWTManager(cfg.JWT.SecretKey, cfg.JWT.ExpireHours)

	// 创建数据访问层
	userRepo := repository.NewUserRepository(database.GetDB())
	productRepo := repository.NewProductRepository(database.GetDB())
	categoryRepo := repository.NewCategoryRepository(database.GetDB())
	cartRepo := repository.NewCartRepository(database.GetDB())
	orderRepo := repository.NewOrderRepository(database.GetDB())

	// 创建业务逻辑层
	userService := service.NewUserService(userRepo, jwtManager)
	productService := service.NewProductService(productRepo, categoryRepo)
	categoryService := service.NewCategoryService(categoryRepo)
	cartService := service.NewCartService(cartRepo, productRepo)
	orderService := service.NewOrderService(orderRepo, cartRepo, productRepo)

	// 创建控制器层
	userHandler := handler.NewUserHandler(userService)
	productHandler := handler.NewProductHandler(productService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	cartHandler := handler.NewCartHandler(cartService)
	orderHandler := handler.NewOrderHandler(orderService)
	
	// 创建监控处理器
	metricsHandler := handler.NewMetricsHandler(metricsManager)
	
	log.Println("✅ 所有组件初始化完成")

	// 6. 设置Gin路由
	// 设置Gin模式
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建Gin引擎
	r := gin.New()

	// 添加中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(response.CORSMiddleware())
	
	// 添加监控中间件
	r.Use(metricsManager.HTTPMetricsMiddleware())
	
	log.Println("✅ 中间件配置完成")

	// 7. 注册路由
	
	// 健康检查路由
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})
	
	// 监控路由
	r.GET("/metrics", metricsHandler.GetMetrics)
	r.GET("/health", metricsHandler.GetHealthCheck)

	// API路由组
	api := r.Group("/api/v1")
	{
		// 用户相关路由
		api.POST("/register", func(c *gin.Context) {
			// 记录用户注册指标
			defer metricsManager.RecordUserRegistration()
			userHandler.Register(c)
		})
		
		api.POST("/login", func(c *gin.Context) {
			// 记录用户登录指标
			defer metricsManager.RecordUserLogin()
			userHandler.Login(c)
		})

		// 需要认证的路由
		auth := api.Group("")
		auth.Use(middleware.AuthMiddleware(jwtManager))
		{
			// 用户路由
			auth.GET("/profile", userHandler.GetProfile)
			auth.PUT("/profile", userHandler.UpdateProfile)
			auth.POST("/change-password", userHandler.ChangePassword)

			// 商品管理路由（需要管理员权限的可以后续添加权限中间件）
			auth.POST("/products", productHandler.Create)
			auth.PUT("/products/:id", productHandler.Update)
			auth.DELETE("/products/:id", productHandler.Delete)

			// 分类管理路由
			auth.POST("/categories", categoryHandler.Create)
			auth.PUT("/categories/:id", categoryHandler.Update)
			auth.DELETE("/categories/:id", categoryHandler.Delete)

			// 购物车路由
			auth.POST("/cart", func(c *gin.Context) {
				defer metricsManager.RecordCartOperation("add")
				cartHandler.AddToCart(c)
			})
			auth.GET("/cart", cartHandler.GetCart)
			auth.PUT("/cart/:id", func(c *gin.Context) {
				defer metricsManager.RecordCartOperation("update")
				cartHandler.UpdateCartItem(c)
			})
			auth.DELETE("/cart/:id", func(c *gin.Context) {
				defer metricsManager.RecordCartOperation("remove")
				cartHandler.RemoveFromCart(c)
			})
			auth.DELETE("/cart", func(c *gin.Context) {
				defer metricsManager.RecordCartOperation("clear")
				cartHandler.ClearCart(c)
			})

			// 订单路由
			auth.POST("/orders", func(c *gin.Context) {
				defer metricsManager.RecordOrderCreated("pending", 0, "unknown")
				orderHandler.CreateOrder(c)
			})
			auth.GET("/orders", orderHandler.GetOrders)
			auth.GET("/orders/:id", orderHandler.GetOrderByID)
			auth.POST("/orders/:id/pay", orderHandler.PayOrder)
			auth.PUT("/orders/:id/cancel", orderHandler.CancelOrder)
		}

		// 公开路由（不需要认证）
		api.GET("/products", func(c *gin.Context) {
			// 这里可以添加商品浏览统计
			productHandler.GetList(c)
		})
		api.GET("/products/:id", func(c *gin.Context) {
			// 记录商品浏览
			// 这里需要从URL参数获取商品ID，暂时使用0
			defer metricsManager.RecordProductView(0, "unknown")
			productHandler.GetByID(c)
		})
		api.GET("/categories", categoryHandler.GetList)
		api.GET("/categories/:id", categoryHandler.GetByID)
	}
	
	log.Println("✅ 路由注册完成")

	// 8. 启动指标收集
	metricsCollector := monitoring.NewMetricsCollector(metricsManager, businessMetrics)
	
	// 设置数据源（这里需要实际的统计函数）
	metricsCollector.SetDataSources(
		func() (int, int) { 
			// 返回数据库连接统计
			return 10, 100 // 活跃连接数, 最大连接数
		},
		func() int { 
			// 返回Redis连接数
			return 5 
		},
		func() int { 
			// 返回WebSocket连接数
			return 0 
		},
	)
	
	// 启动指标收集
	metricsCollector.StartCollection()
	
	log.Println("✅ 指标收集启动完成")

	// 9. 启动服务器
	serverAddr := ":" + cfg.Server.Port
	log.Printf("🚀 服务器启动在端口 %s", cfg.Server.Port)
	log.Printf("📊 监控指标: http://localhost%s/metrics", serverAddr)
	log.Printf("🏥 健康检查: http://localhost%s/health", serverAddr)
	log.Printf("📡 API文档: http://localhost%s/api/v1", serverAddr)
	
	if err := r.Run(serverAddr); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
