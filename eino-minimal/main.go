package main

import (
	"context"
	"log"

	"eino-minimal/api"

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

	// 添加健康检查端点
	h.GET("/health", func(c context.Context, ctx *app.RequestContext) {
		ctx.JSON(consts.StatusOK, map[string]string{
			"status": "ok",
			"service": "eino-minimal",
		})
	})

	log.Println("Eino Minimal API Server starting on :8083")
	log.Println("Chat API: GET /api/chat?id=<conversation_id>&message=<your_message>")
	log.Println("Health Check: GET /health")

	// 启动服务器
	h.Spin()
}
