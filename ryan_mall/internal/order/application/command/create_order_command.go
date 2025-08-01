package command

import (
	"context"

	"ryan-mall-microservices/internal/order/domain/entity"
	"ryan-mall-microservices/internal/order/domain/repository"
	"ryan-mall-microservices/internal/shared/domain"
	"ryan-mall-microservices/internal/shared/events"
)

// CreateOrderCommand 创建订单命令
type CreateOrderCommand struct {
	UserID          string             `json:"user_id" validate:"required"`
	Items           []OrderItemCommand `json:"items" validate:"required,min=1"`
	ShippingAddress string             `json:"shipping_address"`
}

// OrderItemCommand 订单项命令
type OrderItemCommand struct {
	ProductID string  `json:"product_id" validate:"required"`
	Quantity  int     `json:"quantity" validate:"required,min=1"`
	Price     float64 `json:"price" validate:"required,gt=0"`
}

// CreateOrderResult 创建订单结果
type CreateOrderResult struct {
	OrderID     string  `json:"order_id"`
	UserID      string  `json:"user_id"`
	TotalAmount float64 `json:"total_amount"`
	Status      string  `json:"status"`
}

// CreateOrderHandler 创建订单命令处理器
type CreateOrderHandler struct {
	orderRepo      repository.OrderRepository
	eventPublisher *events.EventPublisher
	// 这里可以注入商品服务的gRPC客户端来验证商品和库存
	// productServiceClient product.ProductServiceClient
}

// NewCreateOrderHandler 创建订单命令处理器
func NewCreateOrderHandler(
	orderRepo repository.OrderRepository,
	eventPublisher *events.EventPublisher,
) *CreateOrderHandler {
	return &CreateOrderHandler{
		orderRepo:      orderRepo,
		eventPublisher: eventPublisher,
	}
}

// Handle 处理创建订单命令
func (h *CreateOrderHandler) Handle(ctx context.Context, cmd *CreateOrderCommand) (*CreateOrderResult, error) {
	// TODO: 通过gRPC调用商品服务验证商品存在性和库存
	// 这里是分布式事务的关键点，需要确保库存预留成功
	
	// 转换订单项数据
	itemsData := make([]entity.OrderItemData, len(cmd.Items))
	for i, item := range cmd.Items {
		itemsData[i] = entity.OrderItemData{
			ProductID: domain.ProductID(item.ProductID),
			Quantity:  item.Quantity,
			Price:     domain.NewMoneyFromYuan(item.Price, "CNY"),
		}
	}

	// 创建订单实体
	order, err := entity.NewOrder(domain.UserID(cmd.UserID), itemsData)
	if err != nil {
		return nil, err
	}

	// 设置收货地址
	if cmd.ShippingAddress != "" {
		if err := order.UpdateShippingAddress(cmd.ShippingAddress); err != nil {
			return nil, err
		}
	}

	// 保存订单
	if err := h.orderRepo.Save(ctx, order); err != nil {
		return nil, domain.NewInternalError("failed to save order", err)
	}

	// 发布领域事件
	if h.eventPublisher != nil {
		if err := h.eventPublisher.PublishEvents(ctx, order.DomainEvents()...); err != nil {
			// 事件发布失败不应该影响主流程，但应该记录日志
		}
	}

	// 清除领域事件
	order.ClearDomainEvents()

	return &CreateOrderResult{
		OrderID:     order.ID().String(),
		UserID:      order.UserID().String(),
		TotalAmount: order.TotalAmount().ToYuan(),
		Status:      string(order.Status()),
	}, nil
}

// ConfirmOrderCommand 确认订单命令
type ConfirmOrderCommand struct {
	OrderID string `json:"order_id" validate:"required"`
}

// ConfirmOrderHandler 确认订单命令处理器
type ConfirmOrderHandler struct {
	orderRepo      repository.OrderRepository
	eventPublisher *events.EventPublisher
}

// NewConfirmOrderHandler 创建确认订单命令处理器
func NewConfirmOrderHandler(
	orderRepo repository.OrderRepository,
	eventPublisher *events.EventPublisher,
) *ConfirmOrderHandler {
	return &ConfirmOrderHandler{
		orderRepo:      orderRepo,
		eventPublisher: eventPublisher,
	}
}

// Handle 处理确认订单命令
func (h *ConfirmOrderHandler) Handle(ctx context.Context, cmd *ConfirmOrderCommand) error {
	// 查找订单
	order, err := h.orderRepo.FindByID(ctx, domain.OrderID(cmd.OrderID))
	if err != nil {
		return domain.NewInternalError("failed to find order", err)
	}
	if order == nil {
		return domain.NewNotFoundError("order", cmd.OrderID)
	}

	// 确认订单
	if err := order.Confirm(); err != nil {
		return err
	}

	// 更新订单
	if err := h.orderRepo.Update(ctx, order); err != nil {
		return domain.NewInternalError("failed to update order", err)
	}

	// 发布领域事件
	if h.eventPublisher != nil {
		if err := h.eventPublisher.PublishEvents(ctx, order.DomainEvents()...); err != nil {
			// 事件发布失败不应该影响主流程，但应该记录日志
		}
	}

	// 清除领域事件
	order.ClearDomainEvents()

	return nil
}

// CancelOrderCommand 取消订单命令
type CancelOrderCommand struct {
	OrderID string `json:"order_id" validate:"required"`
	Reason  string `json:"reason" validate:"required"`
}

// CancelOrderHandler 取消订单命令处理器
type CancelOrderHandler struct {
	orderRepo      repository.OrderRepository
	eventPublisher *events.EventPublisher
}

// NewCancelOrderHandler 创建取消订单命令处理器
func NewCancelOrderHandler(
	orderRepo repository.OrderRepository,
	eventPublisher *events.EventPublisher,
) *CancelOrderHandler {
	return &CancelOrderHandler{
		orderRepo:      orderRepo,
		eventPublisher: eventPublisher,
	}
}

// Handle 处理取消订单命令
func (h *CancelOrderHandler) Handle(ctx context.Context, cmd *CancelOrderCommand) error {
	// 查找订单
	order, err := h.orderRepo.FindByID(ctx, domain.OrderID(cmd.OrderID))
	if err != nil {
		return domain.NewInternalError("failed to find order", err)
	}
	if order == nil {
		return domain.NewNotFoundError("order", cmd.OrderID)
	}

	// 取消订单
	if err := order.Cancel(cmd.Reason); err != nil {
		return err
	}

	// 更新订单
	if err := h.orderRepo.Update(ctx, order); err != nil {
		return domain.NewInternalError("failed to update order", err)
	}

	// 发布领域事件
	if h.eventPublisher != nil {
		if err := h.eventPublisher.PublishEvents(ctx, order.DomainEvents()...); err != nil {
			// 事件发布失败不应该影响主流程，但应该记录日志
		}
	}

	// 清除领域事件
	order.ClearDomainEvents()

	return nil
}

// UpdateShippingAddressCommand 更新收货地址命令
type UpdateShippingAddressCommand struct {
	OrderID         string `json:"order_id" validate:"required"`
	ShippingAddress string `json:"shipping_address" validate:"required"`
}

// UpdateShippingAddressHandler 更新收货地址命令处理器
type UpdateShippingAddressHandler struct {
	orderRepo repository.OrderRepository
}

// NewUpdateShippingAddressHandler 创建更新收货地址命令处理器
func NewUpdateShippingAddressHandler(orderRepo repository.OrderRepository) *UpdateShippingAddressHandler {
	return &UpdateShippingAddressHandler{
		orderRepo: orderRepo,
	}
}

// Handle 处理更新收货地址命令
func (h *UpdateShippingAddressHandler) Handle(ctx context.Context, cmd *UpdateShippingAddressCommand) error {
	// 查找订单
	order, err := h.orderRepo.FindByID(ctx, domain.OrderID(cmd.OrderID))
	if err != nil {
		return domain.NewInternalError("failed to find order", err)
	}
	if order == nil {
		return domain.NewNotFoundError("order", cmd.OrderID)
	}

	// 更新收货地址
	if err := order.UpdateShippingAddress(cmd.ShippingAddress); err != nil {
		return err
	}

	// 更新订单
	if err := h.orderRepo.Update(ctx, order); err != nil {
		return domain.NewInternalError("failed to update order", err)
	}

	return nil
}
