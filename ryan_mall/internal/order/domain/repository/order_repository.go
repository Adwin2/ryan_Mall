package repository

import (
	"context"

	"ryan-mall-microservices/internal/order/domain/entity"
	"ryan-mall-microservices/internal/shared/domain"
)

// OrderRepository 订单仓储接口（写模型）
type OrderRepository interface {
	// Save 保存订单
	Save(ctx context.Context, order *entity.Order) error
	
	// FindByID 根据ID查找订单
	FindByID(ctx context.Context, id domain.OrderID) (*entity.Order, error)
	
	// Update 更新订单
	Update(ctx context.Context, order *entity.Order) error
	
	// Delete 删除订单
	Delete(ctx context.Context, id domain.OrderID) error
	
	// FindByUserID 根据用户ID查找订单列表
	FindByUserID(ctx context.Context, userID domain.UserID, offset, limit int) ([]*entity.Order, int64, error)
	
	// FindByStatus 根据状态查找订单列表
	FindByStatus(ctx context.Context, status entity.OrderStatus, offset, limit int) ([]*entity.Order, int64, error)
}

// OrderQueryRepository 订单查询仓储接口（读模型）
type OrderQueryRepository interface {
	// GetOrderDetail 获取订单详情（包含商品信息）
	GetOrderDetail(ctx context.Context, orderID string) (*OrderDetailView, error)
	
	// ListUserOrders 获取用户订单列表
	ListUserOrders(ctx context.Context, userID string, status string, offset, limit int) ([]*OrderListView, int64, error)
	
	// GetOrderStatistics 获取订单统计信息
	GetOrderStatistics(ctx context.Context, userID string) (*OrderStatistics, error)
	
	// SearchOrders 搜索订单
	SearchOrders(ctx context.Context, criteria *OrderSearchCriteria) ([]*OrderListView, int64, error)
}

// OrderDetailView 订单详情视图（读模型）
type OrderDetailView struct {
	OrderID         string                `json:"order_id"`
	UserID          string                `json:"user_id"`
	Status          string                `json:"status"`
	TotalAmount     float64               `json:"total_amount"`
	ShippingAddress string                `json:"shipping_address"`
	CancelReason    string                `json:"cancel_reason"`
	Items           []*OrderItemDetailView `json:"items"`
	CreatedAt       int64                 `json:"created_at"`
	UpdatedAt       int64                 `json:"updated_at"`
}

// OrderItemDetailView 订单项详情视图
type OrderItemDetailView struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
	TotalPrice  float64 `json:"total_price"`
}

// OrderListView 订单列表视图（读模型）
type OrderListView struct {
	OrderID     string  `json:"order_id"`
	Status      string  `json:"status"`
	TotalAmount float64 `json:"total_amount"`
	ItemCount   int     `json:"item_count"`
	CreatedAt   int64   `json:"created_at"`
}

// OrderStatistics 订单统计信息
type OrderStatistics struct {
	TotalOrders     int64   `json:"total_orders"`
	PendingOrders   int64   `json:"pending_orders"`
	CompletedOrders int64   `json:"completed_orders"`
	CancelledOrders int64   `json:"cancelled_orders"`
	TotalAmount     float64 `json:"total_amount"`
}

// OrderSearchCriteria 订单搜索条件
type OrderSearchCriteria struct {
	UserID    string `json:"user_id"`
	Status    string `json:"status"`
	StartDate int64  `json:"start_date"`
	EndDate   int64  `json:"end_date"`
	MinAmount float64 `json:"min_amount"`
	MaxAmount float64 `json:"max_amount"`
	Offset    int    `json:"offset"`
	Limit     int    `json:"limit"`
}
