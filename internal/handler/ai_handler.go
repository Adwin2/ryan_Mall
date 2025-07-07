package handler

import (
	"ryan-mall/internal/middleware"
	"ryan-mall/internal/service"
	"ryan-mall/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
)

type AIHandler struct {
    aiService service.AIService
}

func NewAIHandler(aiService service.AIService) *AIHandler {
    return &AIHandler{
        aiService: aiService,
    }
}

// Chat AI聊天接口
// POST /api/v1/ai/chat
func (h *AIHandler) Chat(c *gin.Context) {
    // 获取当前用户ID
    userID, exists := middleware.GetCurrentUserID(c)
    if !exists {
        response.Unauthorized(c, "请先登录")
        return
    }
    
    // 绑定请求参数
    var req struct {
        Message string `json:"message" binding:"required"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, "消息内容不能为空")
        return
    }
    
    // 调用AI服务
    reply, err := h.aiService.ChatWithAI(userID, req.Message)
    if err != nil {
        response.Error(c, response.ERROR, "AI助手暂时不可用，请稍后再试")
        return
    }
    
    // 返回成功响应
    response.Success(c, gin.H{
        "reply": reply,
        "timestamp": time.Now().Unix(),
    })
}

// RegisterRoutes 注册AI相关路由
func (h *AIHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
    ai := r.Group("/ai")
    ai.Use(authMiddleware.RequireAuth()) // 需要登录
    {
        ai.POST("/chat", h.Chat)
    }
}