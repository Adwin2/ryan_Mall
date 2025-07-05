package handler

import (
	"ryan-mall/internal/middleware"
	"ryan-mall/internal/model"
	"ryan-mall/internal/service"
	"ryan-mall/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CartHandler 购物车HTTP处理器
type CartHandler struct {
	cartService service.CartService
}

// NewCartHandler 创建购物车处理器实例
func NewCartHandler(cartService service.CartService) *CartHandler {
	return &CartHandler{
		cartService: cartService,
	}
}

// AddToCart 添加商品到购物车
// POST /api/v1/cart
// 需要认证
func (h *CartHandler) AddToCart(c *gin.Context) {
	// 1. 获取用户ID
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}
	
	// 2. 绑定请求参数
	var req model.AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	
	// 3. 调用业务逻辑
	err := h.cartService.AddToCart(userID, &req)
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}
	
	// 4. 返回成功响应
	response.SuccessWithMessage(c, "商品已添加到购物车", nil)
}

// GetCart 获取购物车
// GET /api/v1/cart
// 需要认证
func (h *CartHandler) GetCart(c *gin.Context) {
	// 1. 获取用户ID
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}
	
	// 2. 调用业务逻辑
	cart, err := h.cartService.GetCart(userID)
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}
	
	// 3. 返回成功响应
	response.Success(c, cart)
}

// UpdateCartItem 更新购物车商品数量
// PUT /api/v1/cart/:id
// 需要认证
func (h *CartHandler) UpdateCartItem(c *gin.Context) {
	// 1. 获取用户ID
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}
	
	// 2. 获取路径参数
	idStr := c.Param("id")
	cartItemID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "购物车项ID格式错误")
		return
	}
	
	// 3. 绑定请求参数
	var req model.UpdateCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	
	// 4. 调用业务逻辑
	err = h.cartService.UpdateCartItem(userID, uint(cartItemID), &req)
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}
	
	// 5. 返回成功响应
	response.SuccessWithMessage(c, "购物车更新成功", nil)
}

// RemoveFromCart 从购物车移除商品
// DELETE /api/v1/cart/:id
// 需要认证
func (h *CartHandler) RemoveFromCart(c *gin.Context) {
	// 1. 获取用户ID
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}
	
	// 2. 获取路径参数
	idStr := c.Param("id")
	cartItemID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "购物车项ID格式错误")
		return
	}
	
	// 3. 调用业务逻辑
	err = h.cartService.RemoveFromCart(userID, uint(cartItemID))
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}
	
	// 4. 返回成功响应
	response.SuccessWithMessage(c, "商品已从购物车移除", nil)
}

// RemoveProduct 移除特定商品
// DELETE /api/v1/cart/product/:productId
// 需要认证
func (h *CartHandler) RemoveProduct(c *gin.Context) {
	// 1. 获取用户ID
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}
	
	// 2. 获取路径参数
	productIDStr := c.Param("productId")
	productID, err := strconv.ParseUint(productIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "商品ID格式错误")
		return
	}
	
	// 3. 调用业务逻辑
	err = h.cartService.RemoveProduct(userID, uint(productID))
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}
	
	// 4. 返回成功响应
	response.SuccessWithMessage(c, "商品已从购物车移除", nil)
}

// ClearCart 清空购物车
// DELETE /api/v1/cart
// 需要认证
func (h *CartHandler) ClearCart(c *gin.Context) {
	// 1. 获取用户ID
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}
	
	// 2. 调用业务逻辑
	err := h.cartService.ClearCart(userID)
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}
	
	// 3. 返回成功响应
	response.SuccessWithMessage(c, "购物车已清空", nil)
}

// GetCartSummary 获取购物车汇总
// GET /api/v1/cart/summary
// 需要认证
func (h *CartHandler) GetCartSummary(c *gin.Context) {
	// 1. 获取用户ID
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}
	
	// 2. 调用业务逻辑
	summary, err := h.cartService.GetCartSummary(userID)
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}
	
	// 3. 返回成功响应
	response.Success(c, summary)
}

// BatchAddToCart 批量添加商品到购物车
// POST /api/v1/cart/batch
// 需要认证
func (h *CartHandler) BatchAddToCart(c *gin.Context) {
	// 1. 获取用户ID
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}
	
	// 2. 绑定请求参数
	var requests []*model.AddToCartRequest
	if err := c.ShouldBindJSON(&requests); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	
	if len(requests) == 0 {
		response.BadRequest(c, "请求列表不能为空")
		return
	}
	
	// 3. 调用业务逻辑
	for _, req := range requests {
		if err := h.cartService.AddToCart(userID, req); err != nil {
			response.Error(c, response.ERROR, "添加商品失败: "+err.Error())
			return
		}
	}
	
	// 4. 返回成功响应
	response.SuccessWithMessage(c, "商品批量添加成功", nil)
}

// GetCartItemCount 获取购物车商品数量
// GET /api/v1/cart/count
// 需要认证
func (h *CartHandler) GetCartItemCount(c *gin.Context) {
	// 1. 获取用户ID
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}
	
	// 2. 调用业务逻辑
	summary, err := h.cartService.GetCartSummary(userID)
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}
	
	// 3. 返回成功响应
	response.Success(c, gin.H{
		"count": summary.TotalItems,
	})
}

// RegisterRoutes 注册购物车相关路由
func (h *CartHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	// 所有购物车路由都需要认证
	cart := r.Group("/cart")
	cart.Use(authMiddleware.RequireAuth())
	{
		cart.POST("", h.AddToCart)                    // 添加商品到购物车
		cart.GET("", h.GetCart)                       // 获取购物车
		cart.PUT("/:id", h.UpdateCartItem)            // 更新购物车商品数量
		cart.DELETE("/:id", h.RemoveFromCart)         // 移除购物车商品
		cart.DELETE("", h.ClearCart)                  // 清空购物车
		cart.GET("/summary", h.GetCartSummary)        // 获取购物车汇总
		cart.POST("/batch", h.BatchAddToCart)         // 批量添加商品
		cart.GET("/count", h.GetCartItemCount)        // 获取购物车商品数量
		cart.DELETE("/product/:productId", h.RemoveProduct) // 移除特定商品
	}
}
