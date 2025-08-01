package query

import (
	"context"

	"ryan-mall-microservices/internal/order/domain/repository"
	"ryan-mall-microservices/internal/shared/domain"
)

// GetOrderDetailQuery 获取订单详情查询
type GetOrderDetailQuery struct {
	OrderID string `json:"order_id" validate:"required"`
}

// GetOrderDetailHandler 获取订单详情查询处理器
type GetOrderDetailHandler struct {
	orderQueryRepo repository.OrderQueryRepository
}

// NewGetOrderDetailHandler 创建获取订单详情查询处理器
func NewGetOrderDetailHandler(orderQueryRepo repository.OrderQueryRepository) *GetOrderDetailHandler {
	return &GetOrderDetailHandler{
		orderQueryRepo: orderQueryRepo,
	}
}

// Handle 处理获取订单详情查询
func (h *GetOrderDetailHandler) Handle(ctx context.Context, query *GetOrderDetailQuery) (*repository.OrderDetailView, error) {
	orderDetail, err := h.orderQueryRepo.GetOrderDetail(ctx, query.OrderID)
	if err != nil {
		return nil, domain.NewInternalError("failed to get order detail", err)
	}
	if orderDetail == nil {
		return nil, domain.NewNotFoundError("order", query.OrderID)
	}

	return orderDetail, nil
}

// ListUserOrdersQuery 获取用户订单列表查询
type ListUserOrdersQuery struct {
	UserID   string `json:"user_id" validate:"required"`
	Status   string `json:"status"`
	Page     int    `json:"page" validate:"min=1"`
	PageSize int    `json:"page_size" validate:"min=1,max=100"`
}

// ListUserOrdersResult 用户订单列表结果
type ListUserOrdersResult struct {
	Orders     []*repository.OrderListView `json:"orders"`
	Total      int64                       `json:"total"`
	Page       int                         `json:"page"`
	PageSize   int                         `json:"page_size"`
	TotalPages int                         `json:"total_pages"`
}

// ListUserOrdersHandler 获取用户订单列表查询处理器
type ListUserOrdersHandler struct {
	orderQueryRepo repository.OrderQueryRepository
}

// NewListUserOrdersHandler 创建获取用户订单列表查询处理器
func NewListUserOrdersHandler(orderQueryRepo repository.OrderQueryRepository) *ListUserOrdersHandler {
	return &ListUserOrdersHandler{
		orderQueryRepo: orderQueryRepo,
	}
}

// Handle 处理获取用户订单列表查询
func (h *ListUserOrdersHandler) Handle(ctx context.Context, query *ListUserOrdersQuery) (*ListUserOrdersResult, error) {
	// 计算偏移量
	offset := (query.Page - 1) * query.PageSize

	// 查询订单列表
	orders, total, err := h.orderQueryRepo.ListUserOrders(ctx, query.UserID, query.Status, offset, query.PageSize)
	if err != nil {
		return nil, domain.NewInternalError("failed to list user orders", err)
	}

	// 计算总页数
	totalPages := int((total + int64(query.PageSize) - 1) / int64(query.PageSize))

	return &ListUserOrdersResult{
		Orders:     orders,
		Total:      total,
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetOrderStatisticsQuery 获取订单统计查询
type GetOrderStatisticsQuery struct {
	UserID string `json:"user_id" validate:"required"`
}

// GetOrderStatisticsHandler 获取订单统计查询处理器
type GetOrderStatisticsHandler struct {
	orderQueryRepo repository.OrderQueryRepository
}

// NewGetOrderStatisticsHandler 创建获取订单统计查询处理器
func NewGetOrderStatisticsHandler(orderQueryRepo repository.OrderQueryRepository) *GetOrderStatisticsHandler {
	return &GetOrderStatisticsHandler{
		orderQueryRepo: orderQueryRepo,
	}
}

// Handle 处理获取订单统计查询
func (h *GetOrderStatisticsHandler) Handle(ctx context.Context, query *GetOrderStatisticsQuery) (*repository.OrderStatistics, error) {
	statistics, err := h.orderQueryRepo.GetOrderStatistics(ctx, query.UserID)
	if err != nil {
		return nil, domain.NewInternalError("failed to get order statistics", err)
	}

	return statistics, nil
}

// SearchOrdersQuery 搜索订单查询
type SearchOrdersQuery struct {
	UserID    string  `json:"user_id"`
	Status    string  `json:"status"`
	StartDate int64   `json:"start_date"`
	EndDate   int64   `json:"end_date"`
	MinAmount float64 `json:"min_amount"`
	MaxAmount float64 `json:"max_amount"`
	Page      int     `json:"page" validate:"min=1"`
	PageSize  int     `json:"page_size" validate:"min=1,max=100"`
}

// SearchOrdersResult 搜索订单结果
type SearchOrdersResult struct {
	Orders     []*repository.OrderListView `json:"orders"`
	Total      int64                       `json:"total"`
	Page       int                         `json:"page"`
	PageSize   int                         `json:"page_size"`
	TotalPages int                         `json:"total_pages"`
}

// SearchOrdersHandler 搜索订单查询处理器
type SearchOrdersHandler struct {
	orderQueryRepo repository.OrderQueryRepository
}

// NewSearchOrdersHandler 创建搜索订单查询处理器
func NewSearchOrdersHandler(orderQueryRepo repository.OrderQueryRepository) *SearchOrdersHandler {
	return &SearchOrdersHandler{
		orderQueryRepo: orderQueryRepo,
	}
}

// Handle 处理搜索订单查询
func (h *SearchOrdersHandler) Handle(ctx context.Context, query *SearchOrdersQuery) (*SearchOrdersResult, error) {
	// 构建搜索条件
	criteria := &repository.OrderSearchCriteria{
		UserID:    query.UserID,
		Status:    query.Status,
		StartDate: query.StartDate,
		EndDate:   query.EndDate,
		MinAmount: query.MinAmount,
		MaxAmount: query.MaxAmount,
		Offset:    (query.Page - 1) * query.PageSize,
		Limit:     query.PageSize,
	}

	// 搜索订单
	orders, total, err := h.orderQueryRepo.SearchOrders(ctx, criteria)
	if err != nil {
		return nil, domain.NewInternalError("failed to search orders", err)
	}

	// 计算总页数
	totalPages := int((total + int64(query.PageSize) - 1) / int64(query.PageSize))

	return &SearchOrdersResult{
		Orders:     orders,
		Total:      total,
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetOrdersByStatusQuery 根据状态获取订单查询
type GetOrdersByStatusQuery struct {
	Status   string `json:"status" validate:"required"`
	Page     int    `json:"page" validate:"min=1"`
	PageSize int    `json:"page_size" validate:"min=1,max=100"`
}

// GetOrdersByStatusResult 根据状态获取订单结果
type GetOrdersByStatusResult struct {
	Orders     []*repository.OrderListView `json:"orders"`
	Total      int64                       `json:"total"`
	Page       int                         `json:"page"`
	PageSize   int                         `json:"page_size"`
	TotalPages int                         `json:"total_pages"`
}

// GetOrdersByStatusHandler 根据状态获取订单查询处理器
type GetOrdersByStatusHandler struct {
	orderQueryRepo repository.OrderQueryRepository
}

// NewGetOrdersByStatusHandler 创建根据状态获取订单查询处理器
func NewGetOrdersByStatusHandler(orderQueryRepo repository.OrderQueryRepository) *GetOrdersByStatusHandler {
	return &GetOrdersByStatusHandler{
		orderQueryRepo: orderQueryRepo,
	}
}

// Handle 处理根据状态获取订单查询
func (h *GetOrdersByStatusHandler) Handle(ctx context.Context, query *GetOrdersByStatusQuery) (*GetOrdersByStatusResult, error) {
	// 构建搜索条件
	criteria := &repository.OrderSearchCriteria{
		Status: query.Status,
		Offset: (query.Page - 1) * query.PageSize,
		Limit:  query.PageSize,
	}

	// 搜索订单
	orders, total, err := h.orderQueryRepo.SearchOrders(ctx, criteria)
	if err != nil {
		return nil, domain.NewInternalError("failed to get orders by status", err)
	}

	// 计算总页数
	totalPages := int((total + int64(query.PageSize) - 1) / int64(query.PageSize))

	return &GetOrdersByStatusResult{
		Orders:     orders,
		Total:      total,
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalPages: totalPages,
	}, nil
}
