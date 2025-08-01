package http

import (
	"net/http"
	"strconv"
	"time"

	"ryan-mall-microservices/internal/seckill/application/command"
	"ryan-mall-microservices/internal/seckill/application/query"
	"ryan-mall-microservices/internal/seckill/application/service"
	"ryan-mall-microservices/internal/shared/domain"
	"ryan-mall-microservices/internal/shared/infrastructure"

	"github.com/gin-gonic/gin"
)

// SeckillHandler 秒杀HTTP处理器
type SeckillHandler struct {
	seckillAppSvc *service.SeckillApplicationService
	logger        infrastructure.Logger
}

// NewSeckillHandler 创建秒杀HTTP处理器
func NewSeckillHandler(seckillAppSvc *service.SeckillApplicationService, logger infrastructure.Logger) *SeckillHandler {
	return &SeckillHandler{
		seckillAppSvc: seckillAppSvc,
		logger:        logger,
	}
}

// RegisterRoutes 注册路由
func (h *SeckillHandler) RegisterRoutes(router *gin.Engine) {
	seckillGroup := router.Group("/api/v1/seckill")
	{
		// 活动管理
		seckillGroup.POST("/activities", h.CreateSeckillActivity)
		seckillGroup.POST("/activities/:id/start", h.StartSeckillActivity)
		seckillGroup.GET("/activities/:id", h.GetSeckillActivityDetail)
		seckillGroup.GET("/activities/:id/statistics", h.GetSeckillActivityStatistics)
		
		// 活动列表
		seckillGroup.GET("/activities/active", h.ListActiveSeckillActivities)
		seckillGroup.GET("/activities/upcoming", h.ListUpcomingSeckillActivities)
		seckillGroup.GET("/activities/search", h.SearchSeckillActivities)
		
		// 参与秒杀
		seckillGroup.POST("/participate", h.ParticipateInSeckill)
		
		// 订单管理
		seckillGroup.GET("/orders/user/:user_id", h.GetUserSeckillOrders)
		seckillGroup.POST("/orders/:id/pay", h.PaySeckillOrder)
		seckillGroup.POST("/orders/:id/cancel", h.CancelSeckillOrder)
		
		// 仪表板
		seckillGroup.GET("/dashboard", h.GetSeckillDashboard)
	}
}

// CreateSeckillActivity 创建秒杀活动
func (h *SeckillHandler) CreateSeckillActivity(c *gin.Context) {
	var req struct {
		Name          string `json:"name" binding:"required"`
		ProductID     string `json:"product_id" binding:"required"`
		OriginalPrice float64 `json:"original_price" binding:"required,gt=0"`
		SeckillPrice  float64 `json:"seckill_price" binding:"required,gt=0"`
		TotalStock    int     `json:"total_stock" binding:"required,gt=0"`
		StartTime     int64   `json:"start_time" binding:"required"`
		EndTime       int64   `json:"end_time" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	cmd := &command.CreateSeckillActivityCommand{
		Name:          req.Name,
		ProductID:     req.ProductID,
		OriginalPrice: req.OriginalPrice,
		SeckillPrice:  req.SeckillPrice,
		TotalStock:    req.TotalStock,
		StartTime:     time.Unix(req.StartTime, 0),
		EndTime:       time.Unix(req.EndTime, 0),
	}

	result, err := h.seckillAppSvc.CreateSeckillActivity(c.Request.Context(), cmd)
	if err != nil {
		h.handleDomainError(c, err)
		return
	}

	h.respondSuccess(c, http.StatusCreated, "seckill activity created successfully", result)
}

// StartSeckillActivity 启动秒杀活动
func (h *SeckillHandler) StartSeckillActivity(c *gin.Context) {
	activityID := c.Param("id")
	if activityID == "" {
		h.respondError(c, http.StatusBadRequest, "activity id is required", nil)
		return
	}

	cmd := &command.StartSeckillActivityCommand{
		ActivityID: activityID,
	}

	err := h.seckillAppSvc.StartSeckillActivity(c.Request.Context(), cmd)
	if err != nil {
		h.handleDomainError(c, err)
		return
	}

	h.respondSuccess(c, http.StatusOK, "seckill activity started successfully", nil)
}

// GetSeckillActivityDetail 获取秒杀活动详情
func (h *SeckillHandler) GetSeckillActivityDetail(c *gin.Context) {
	activityID := c.Param("id")
	if activityID == "" {
		h.respondError(c, http.StatusBadRequest, "activity id is required", nil)
		return
	}

	query := &query.GetSeckillActivityDetailQuery{
		ActivityID: activityID,
	}

	result, err := h.seckillAppSvc.GetSeckillActivityDetail(c.Request.Context(), query)
	if err != nil {
		h.handleDomainError(c, err)
		return
	}

	h.respondSuccess(c, http.StatusOK, "activity detail retrieved successfully", result)
}

// ListActiveSeckillActivities 获取激活的秒杀活动列表
func (h *SeckillHandler) ListActiveSeckillActivities(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	query := &query.ListActiveSeckillActivitiesQuery{
		Page:     page,
		PageSize: pageSize,
	}

	result, err := h.seckillAppSvc.ListActiveSeckillActivities(c.Request.Context(), query)
	if err != nil {
		h.handleDomainError(c, err)
		return
	}

	h.respondSuccess(c, http.StatusOK, "active activities retrieved successfully", result)
}

// ListUpcomingSeckillActivities 获取即将开始的秒杀活动列表
func (h *SeckillHandler) ListUpcomingSeckillActivities(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	query := &query.ListUpcomingSeckillActivitiesQuery{
		Page:     page,
		PageSize: pageSize,
	}

	result, err := h.seckillAppSvc.ListUpcomingSeckillActivities(c.Request.Context(), query)
	if err != nil {
		h.handleDomainError(c, err)
		return
	}

	h.respondSuccess(c, http.StatusOK, "upcoming activities retrieved successfully", result)
}

// ParticipateInSeckill 参与秒杀
func (h *SeckillHandler) ParticipateInSeckill(c *gin.Context) {
	var req struct {
		UserID     string `json:"user_id" binding:"required"`
		ActivityID string `json:"activity_id" binding:"required"`
		Quantity   int    `json:"quantity" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	cmd := &command.ParticipateInSeckillCommand{
		UserID:     req.UserID,
		ActivityID: req.ActivityID,
		Quantity:   req.Quantity,
	}

	result, err := h.seckillAppSvc.ParticipateInSeckill(c.Request.Context(), cmd)
	if err != nil {
		h.handleDomainError(c, err)
		return
	}

	h.respondSuccess(c, http.StatusCreated, "participated in seckill successfully", result)
}

// GetUserSeckillOrders 获取用户秒杀订单
func (h *SeckillHandler) GetUserSeckillOrders(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		h.respondError(c, http.StatusBadRequest, "user id is required", nil)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	query := &query.GetUserSeckillOrdersQuery{
		UserID:   userID,
		Page:     page,
		PageSize: pageSize,
	}

	result, err := h.seckillAppSvc.GetUserSeckillOrders(c.Request.Context(), query)
	if err != nil {
		h.handleDomainError(c, err)
		return
	}

	h.respondSuccess(c, http.StatusOK, "user seckill orders retrieved successfully", result)
}

// PaySeckillOrder 支付秒杀订单
func (h *SeckillHandler) PaySeckillOrder(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		h.respondError(c, http.StatusBadRequest, "order id is required", nil)
		return
	}

	var req struct {
		UserID string `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	cmd := &command.PaySeckillOrderCommand{
		OrderID: orderID,
		UserID:  req.UserID,
	}

	err := h.seckillAppSvc.PaySeckillOrder(c.Request.Context(), cmd)
	if err != nil {
		h.handleDomainError(c, err)
		return
	}

	h.respondSuccess(c, http.StatusOK, "seckill order paid successfully", nil)
}

// CancelSeckillOrder 取消秒杀订单
func (h *SeckillHandler) CancelSeckillOrder(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		h.respondError(c, http.StatusBadRequest, "order id is required", nil)
		return
	}

	var req struct {
		UserID string `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	cmd := &command.CancelSeckillOrderCommand{
		OrderID: orderID,
		UserID:  req.UserID,
	}

	err := h.seckillAppSvc.CancelSeckillOrder(c.Request.Context(), cmd)
	if err != nil {
		h.handleDomainError(c, err)
		return
	}

	h.respondSuccess(c, http.StatusOK, "seckill order cancelled successfully", nil)
}

// GetSeckillActivityStatistics 获取秒杀活动统计
func (h *SeckillHandler) GetSeckillActivityStatistics(c *gin.Context) {
	activityID := c.Param("id")
	if activityID == "" {
		h.respondError(c, http.StatusBadRequest, "activity id is required", nil)
		return
	}

	query := &query.GetSeckillActivityStatisticsQuery{
		ActivityID: activityID,
	}

	result, err := h.seckillAppSvc.GetSeckillActivityStatistics(c.Request.Context(), query)
	if err != nil {
		h.handleDomainError(c, err)
		return
	}

	h.respondSuccess(c, http.StatusOK, "activity statistics retrieved successfully", result)
}

// SearchSeckillActivities 搜索秒杀活动
func (h *SeckillHandler) SearchSeckillActivities(c *gin.Context) {
	keyword := c.Query("keyword")
	status := c.Query("status")
	startDate, _ := strconv.ParseInt(c.Query("start_date"), 10, 64)
	endDate, _ := strconv.ParseInt(c.Query("end_date"), 10, 64)
	minPrice, _ := strconv.ParseFloat(c.Query("min_price"), 64)
	maxPrice, _ := strconv.ParseFloat(c.Query("max_price"), 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	query := &query.SearchSeckillActivitiesQuery{
		Keyword:   keyword,
		Status:    status,
		StartDate: startDate,
		EndDate:   endDate,
		MinPrice:  minPrice,
		MaxPrice:  maxPrice,
		Page:      page,
		PageSize:  pageSize,
	}

	result, err := h.seckillAppSvc.SearchSeckillActivities(c.Request.Context(), query)
	if err != nil {
		h.handleDomainError(c, err)
		return
	}

	h.respondSuccess(c, http.StatusOK, "activities searched successfully", result)
}

// GetSeckillDashboard 获取秒杀仪表板
func (h *SeckillHandler) GetSeckillDashboard(c *gin.Context) {
	result, err := h.seckillAppSvc.GetSeckillDashboard(c.Request.Context())
	if err != nil {
		h.handleDomainError(c, err)
		return
	}

	h.respondSuccess(c, http.StatusOK, "seckill dashboard retrieved successfully", result)
}

// APIResponse API响应结构
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// respondSuccess 成功响应
func (h *SeckillHandler) respondSuccess(c *gin.Context, statusCode int, message string, data interface{}) {
	response := APIResponse{
		Code:    statusCode,
		Message: message,
		Data:    data,
	}
	c.JSON(statusCode, response)
}

// respondError 错误响应
func (h *SeckillHandler) respondError(c *gin.Context, statusCode int, message string, err error) {
	response := APIResponse{
		Code:    statusCode,
		Message: message,
	}

	if err != nil {
		response.Error = err.Error()
		h.logger.Error("HTTP request error",
			infrastructure.String("method", c.Request.Method),
			infrastructure.String("path", c.Request.URL.Path),
			infrastructure.String("error", err.Error()),
		)
	}

	c.JSON(statusCode, response)
}

// handleDomainError 处理领域错误
func (h *SeckillHandler) handleDomainError(c *gin.Context, err error) {
	if domainErr, ok := err.(domain.DomainError); ok {
		switch domainErr.Code() {
		case domain.ErrCodeValidation:
			h.respondError(c, http.StatusBadRequest, "validation error", err)
		case domain.ErrCodeNotFound:
			h.respondError(c, http.StatusNotFound, "resource not found", err)
		case domain.ErrCodeAlreadyExists:
			h.respondError(c, http.StatusConflict, "resource already exists", err)
		case domain.ErrCodeUnauthorized:
			h.respondError(c, http.StatusUnauthorized, "unauthorized", err)
		case domain.ErrCodeForbidden:
			h.respondError(c, http.StatusForbidden, "forbidden", err)
		case domain.ErrCodeInsufficientStock:
			h.respondError(c, http.StatusConflict, "insufficient stock", err)
		case domain.ErrCodeTooManyRequests:
			h.respondError(c, http.StatusTooManyRequests, "too many requests", err)
		case domain.ErrCodeInternalError:
			h.respondError(c, http.StatusInternalServerError, "internal server error", err)
		default:
			h.respondError(c, http.StatusInternalServerError, "unknown error", err)
		}
	} else {
		h.respondError(c, http.StatusInternalServerError, "internal server error", err)
	}
}
