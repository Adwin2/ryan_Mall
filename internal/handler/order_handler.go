package handler

import (
	"ryan-mall/internal/middleware"
	"ryan-mall/internal/model"
	"ryan-mall/internal/service"
	"ryan-mall/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// OrderHandler 订单HTTP处理器
type OrderHandler struct {
	orderService service.OrderService
}

// NewOrderHandler 创建订单处理器实例
func NewOrderHandler(orderService service.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

// CreateOrder 创建订单
// POST /api/v1/orders
// 需要认证
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	// 1. 获取用户ID
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}
	
	// 2. 绑定请求参数
	var req model.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	
	// 3. 调用业务逻辑
	order, err := h.orderService.CreateOrder(userID, &req)
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}
	
	// 4. 返回成功响应
	response.SuccessWithMessage(c, "订单创建成功", order)
}

// GetOrder 获取订单详情
// GET /api/v1/orders/:id
// 需要认证
func (h *OrderHandler) GetOrder(c *gin.Context) {
	// 1. 获取用户ID
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}
	
	// 2. 获取路径参数
	idStr := c.Param("id")
	orderID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "订单ID格式错误")
		return
	}
	
	// 3. 调用业务逻辑
	order, err := h.orderService.GetOrder(userID, uint(orderID))
	if err != nil {
		response.Error(c, response.NOT_FOUND, err.Error())
		return
	}
	
	// 4. 返回成功响应
	response.Success(c, order)
}

// GetOrderByNo 根据订单号获取订单
// GET /api/v1/orders/no/:orderNo
// 需要认证
func (h *OrderHandler) GetOrderByNo(c *gin.Context) {
	// 1. 获取用户ID
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}
	
	// 2. 获取路径参数
	orderNo := c.Param("orderNo")
	if orderNo == "" {
		response.BadRequest(c, "订单号不能为空")
		return
	}
	
	// 3. 调用业务逻辑
	order, err := h.orderService.GetOrderByNo(userID, orderNo)
	if err != nil {
		response.Error(c, response.NOT_FOUND, err.Error())
		return
	}
	
	// 4. 返回成功响应
	response.Success(c, order)
}

// GetOrderList 获取订单列表
// GET /api/v1/orders
// 需要认证
func (h *OrderHandler) GetOrderList(c *gin.Context) {
	// 1. 获取用户ID
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}
	
	// 2. 绑定查询参数
	var req model.OrderListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "查询参数错误: "+err.Error())
		return
	}
	
	// 3. 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}
	
	// 4. 调用业务逻辑
	result, err := h.orderService.GetOrderList(userID, &req)
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}
	
	// 5. 返回成功响应
	response.Success(c, result)
}

// CancelOrder 取消订单
// PUT /api/v1/orders/:id/cancel
// 需要认证
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	// 1. 获取用户ID
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}
	
	// 2. 获取路径参数
	idStr := c.Param("id")
	orderID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "订单ID格式错误")
		return
	}
	
	// 3. 调用业务逻辑
	err = h.orderService.CancelOrder(userID, uint(orderID))
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}
	
	// 4. 返回成功响应
	response.SuccessWithMessage(c, "订单取消成功", nil)
}

// PayOrder 支付订单
// POST /api/v1/orders/:id/pay
// 需要认证
func (h *OrderHandler) PayOrder(c *gin.Context) {
	// 1. 获取用户ID
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}
	
	// 2. 获取路径参数
	idStr := c.Param("id")
	orderID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "订单ID格式错误")
		return
	}
	
	// 3. 绑定请求参数
	var req model.PayOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	
	// 4. 调用业务逻辑
	err = h.orderService.PayOrder(userID, uint(orderID), &req)
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}
	
	// 5. 返回成功响应
	response.SuccessWithMessage(c, "支付成功", nil)
}

// ConfirmOrder 确认收货
// PUT /api/v1/orders/:id/confirm
// 需要认证
func (h *OrderHandler) ConfirmOrder(c *gin.Context) {
	// 1. 获取用户ID
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}
	
	// 2. 获取路径参数
	idStr := c.Param("id")
	orderID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "订单ID格式错误")
		return
	}
	
	// 3. 调用业务逻辑
	err = h.orderService.ConfirmOrder(userID, uint(orderID))
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}
	
	// 4. 返回成功响应
	response.SuccessWithMessage(c, "确认收货成功", nil)
}

// GetOrderStatistics 获取订单统计
// GET /api/v1/orders/statistics
// 需要认证
func (h *OrderHandler) GetOrderStatistics(c *gin.Context) {
	// 1. 获取用户ID
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}
	
	// 2. 调用业务逻辑
	stats, err := h.orderService.GetOrderStatistics(userID)
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}
	
	// 3. 返回成功响应
	response.Success(c, stats)
}

// ProcessExpiredOrders 处理过期订单（管理员功能）
// POST /api/v1/orders/process-expired
// 需要认证（管理员权限）
func (h *OrderHandler) ProcessExpiredOrders(c *gin.Context) {
	// 调用业务逻辑
	err := h.orderService.ProcessExpiredOrders()
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}
	
	// 返回成功响应
	response.SuccessWithMessage(c, "过期订单处理完成", nil)
}

// RegisterRoutes 注册订单相关路由
func (h *OrderHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	// 所有订单路由都需要认证
	orders := r.Group("/orders")
	orders.Use(authMiddleware.RequireAuth())
	{
		orders.POST("", h.CreateOrder)                    // 创建订单
		orders.GET("", h.GetOrderList)                    // 获取订单列表
		orders.GET("/statistics", h.GetOrderStatistics)  // 获取订单统计
		orders.GET("/:id", h.GetOrder)                    // 获取订单详情
		orders.GET("/no/:orderNo", h.GetOrderByNo)        // 根据订单号获取订单
		orders.PUT("/:id/cancel", h.CancelOrder)          // 取消订单
		orders.POST("/:id/pay", h.PayOrder)               // 支付订单
		orders.PUT("/:id/confirm", h.ConfirmOrder)        // 确认收货
		
		// 管理员功能
		orders.POST("/process-expired", h.ProcessExpiredOrders) // 处理过期订单
	}
}
