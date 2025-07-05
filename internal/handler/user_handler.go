package handler

import (
	"ryan-mall/internal/middleware"
	"ryan-mall/internal/model"
	"ryan-mall/internal/service"
	"ryan-mall/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户HTTP处理器
// 负责处理用户相关的HTTP请求
type UserHandler struct {
	userService service.UserService // 用户业务逻辑服务
}

// NewUserHandler 创建用户处理器实例
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// Register 用户注册
// POST /api/v1/register
func (h *UserHandler) Register(c *gin.Context) {
	// 1. 绑定请求参数
	var req model.UserRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Gin的ShouldBindJSON会自动验证binding标签
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	
	// 2. 调用业务逻辑
	result, err := h.userService.Register(&req)
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}
	
	// 3. 返回成功响应
	response.SuccessWithMessage(c, "注册成功", result)
}

// Login 用户登录
// POST /api/v1/login
func (h *UserHandler) Login(c *gin.Context) {
	// 1. 绑定请求参数
	var req model.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	
	// 2. 调用业务逻辑
	result, err := h.userService.Login(&req)
	if err != nil {
		response.Error(c, response.UNAUTHORIZED, err.Error())
		return
	}
	
	// 3. 返回成功响应
	response.SuccessWithMessage(c, "登录成功", result)
}

// GetProfile 获取用户资料
// GET /api/v1/profile
// 需要认证
func (h *UserHandler) GetProfile(c *gin.Context) {
	// 1. 从中间件获取用户ID
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}
	
	// 2. 调用业务逻辑
	profile, err := h.userService.GetProfile(userID)
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}
	
	// 3. 返回成功响应
	response.Success(c, profile)
}

// UpdateProfile 更新用户资料
// PUT /api/v1/profile
// 需要认证
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	// 1. 从中间件获取用户ID
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}
	
	// 2. 绑定请求参数
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	
	// 3. 过滤允许更新的字段
	allowedFields := map[string]bool{
		"phone":  true,
		"avatar": true,
		"email":  true,
	}
	
	filteredUpdates := make(map[string]interface{})
	for field, value := range updates {
		if allowedFields[field] {
			filteredUpdates[field] = value
		}
	}
	
	if len(filteredUpdates) == 0 {
		response.BadRequest(c, "没有可更新的字段")
		return
	}
	
	// 4. 调用业务逻辑
	err := h.userService.UpdateProfile(userID, filteredUpdates)
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}
	
	// 5. 返回成功响应
	response.SuccessWithMessage(c, "资料更新成功", nil)
}

// ChangePassword 修改密码
// POST /api/v1/change-password
// 需要认证
func (h *UserHandler) ChangePassword(c *gin.Context) {
	// 1. 从中间件获取用户ID
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}
	
	// 2. 定义请求结构体
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	
	// 3. 调用业务逻辑
	err := h.userService.ChangePassword(userID, req.OldPassword, req.NewPassword)
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}
	
	// 4. 返回成功响应
	response.SuccessWithMessage(c, "密码修改成功", nil)
}

// GetUserByID 根据ID获取用户信息（管理员功能）
// GET /api/v1/users/:id
// 需要认证和管理员权限
func (h *UserHandler) GetUserByID(c *gin.Context) {
	// 1. 获取路径参数
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "用户ID格式错误")
		return
	}
	
	// 2. 调用业务逻辑
	profile, err := h.userService.GetProfile(uint(id))
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}
	
	// 3. 返回成功响应
	response.Success(c, profile)
}

// Logout 用户登出
// POST /api/v1/logout
// 需要认证
func (h *UserHandler) Logout(c *gin.Context) {
	// JWT是无状态的，登出通常在客户端处理（删除本地存储的token）
	// 服务端可以维护一个黑名单来实现真正的登出，但这会增加复杂性
	// 这里只是返回成功响应，实际的登出逻辑在客户端实现
	
	response.SuccessWithMessage(c, "登出成功", nil)
}

// RefreshToken 刷新令牌
// POST /api/v1/refresh-token
// 需要认证
func (h *UserHandler) RefreshToken(c *gin.Context) {
	// 1. 从请求头获取当前令牌
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		response.Unauthorized(c, "缺少认证令牌")
		return
	}
	
	const bearerPrefix = "Bearer "
	if len(authHeader) <= len(bearerPrefix) {
		response.Unauthorized(c, "认证令牌格式错误")
		return
	}
	
	tokenString := authHeader[len(bearerPrefix):]
	
	// 2. 验证并刷新令牌
	// 注意：这里需要在JWT管理器中实现RefreshToken方法
	// 或者重新生成令牌
	claims, err := h.userService.ValidateToken(tokenString)
	if err != nil {
		response.Unauthorized(c, "令牌无效")
		return
	}
	
	// 3. 生成新令牌（简化实现，实际应该检查令牌是否即将过期）
	// 这里需要访问JWT管理器，可能需要调整服务接口
	response.SuccessWithMessage(c, "令牌刷新功能待实现", gin.H{
		"user_id": claims.UserID,
		"message": "当前令牌仍然有效",
	})
}

// RegisterRoutes 注册用户相关路由
// 这个方法用于在main.go中注册所有用户相关的路由
func (h *UserHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	// 公开路由（不需要认证）
	r.POST("/register", h.Register)
	r.POST("/login", h.Login)
	
	// 需要认证的路由
	auth := r.Group("")
	auth.Use(authMiddleware.RequireAuth())
	{
		auth.GET("/profile", h.GetProfile)
		auth.PUT("/profile", h.UpdateProfile)
		auth.POST("/change-password", h.ChangePassword)
		auth.POST("/logout", h.Logout)
		auth.POST("/refresh-token", h.RefreshToken)
		
		// 管理员路由（需要额外的权限检查）
		auth.GET("/users/:id", h.GetUserByID)
	}
}
