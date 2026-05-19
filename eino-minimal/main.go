package main

import (
	"context"
	"log"
	"os"

	"eino-minimal/api"
	"eino-minimal/multiagent"
	"eino-minimal/multiagent/memory"
	"eino-minimal/multiagent/routing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/joho/godotenv"
)

func main() {
	// 加载.env文件
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// 创建Hertz服务器
	h := server.Default(server.WithHostPorts(":8083"))

	// 设置聊天API路由
	h.GET("/api/chat", api.ChatHandler())
	h.POST("/api/chat", api.ChatHandler())

	// 设置 Copilot API 路由
	h.GET("/api/copilot/stream", api.CopilotStreamHandler())
	h.POST("/api/copilot/plan", api.CopilotPlanHandler())
	h.POST("/api/copilot/diagnose", api.CopilotDiagnoseHandler())
	h.POST("/api/copilot/plan-from", api.CopilotPlanFromHandler())

	// Multi-Agent 路由 (ModelRouter + MemoryManager + ContextBuilder)
	infra := initMultiAgentInfra()
	h.GET("/api/ops/stream", api.OpsStreamHandler(infra))
	h.GET("/api/shop/stream", api.ShopStreamHandler(infra))
	h.GET("/api/routing/metrics", api.RoutingMetricsHandler(infra))
	h.POST("/api/routing/reload", api.RoutingReloadHandler(infra))

	// 添加健康检查端点
	h.GET("/health", func(c context.Context, ctx *app.RequestContext) {
		ctx.JSON(consts.StatusOK, map[string]string{
			"status":  "ok",
			"service": "eino-minimal",
		})
	})

	// 静态文件服务（生产模式 serve web/dist/）
	if _, err := os.Stat("./web/dist/index.html"); err == nil {
		h.GET("/assets/*filepath", func(_ context.Context, ctx *app.RequestContext) {
			filePath := "./web/dist/assets/" + ctx.Param("filepath")
			ctx.File(filePath)
		})
		h.GET("/favicon.svg", func(_ context.Context, ctx *app.RequestContext) {
			ctx.File("./web/dist/favicon.svg")
		})
		h.NoRoute(func(_ context.Context, ctx *app.RequestContext) {
			ctx.File("./web/dist/index.html")
		})
		log.Println("Serving frontend from ./web/dist/")
	}

	log.Println("Eino Minimal API Server starting on :8083")
	log.Println("Chat API: GET /api/chat?id=<conversation_id>&message=<your_message>")
	log.Println("Copilot Stream: GET /api/copilot/stream?session_id=<id>&message=<goal>")
	log.Println("Copilot Diagnose: POST /api/copilot/diagnose {session_id, message}")
	log.Println("Copilot PlanFrom: POST /api/copilot/plan-from {session_id, diagnosis, full_analysis}")
	log.Println("Copilot Plan:     POST /api/copilot/plan {session_id, message}")
	log.Println("Health Check: GET /health")

	// 启动服务器
	h.Spin()
}

func initMultiAgentInfra() *multiagent.AgentInfra {
	router, err := routing.NewModelRouter("./model_routing.yaml")
	if err != nil {
		log.Printf("[MultiAgent] ModelRouter init failed (multi-agent disabled): %v", err)
		return nil
	}

	storage := memory.NewJSONLStorage("./data/memory/multiagent")

	// 初始化 Qdrant LongTermStore
	var longStore memory.LongTermStore
	embedder := memory.NewArkEmbedder("doubao-embedding-large-text-240915")
	qdrantAddr := os.Getenv("QDRANT_ADDR")
	if qdrantAddr == "" {
		qdrantAddr = "localhost:6334"
	}
	qs, err := memory.NewQdrantStore(embedder, memory.QdrantConfig{
		Addr:       qdrantAddr,
		Collection: "agent_memory",
		VecDim:     1024,
	})
	if err != nil {
		log.Printf("[MultiAgent] Qdrant init failed (long-term memory disabled): %v", err)
	} else {
		if err := qs.EnsureCollection(context.Background()); err != nil {
			log.Printf("[MultiAgent] Qdrant collection init failed: %v", err)
		} else {
			longStore = qs
			log.Println("[MultiAgent] Qdrant LongTermStore ready")
		}
	}

	memMgr := memory.NewMemoryManager(storage, longStore)

	return &multiagent.AgentInfra{
		Router: router,
		Memory: memMgr,
	}
}
