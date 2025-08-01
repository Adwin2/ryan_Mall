package command

import (
	"context"
	"fmt"

	"ryan-mall-microservices/internal/product/domain/repository"
	"ryan-mall-microservices/internal/shared/domain"
	"ryan-mall-microservices/internal/shared/events"
)

// ReserveStockCommand 预留库存命令
type ReserveStockCommand struct {
	ProductID string `json:"product_id" validate:"required"`
	Quantity  int    `json:"quantity" validate:"required,gt=0"`
	OrderID   string `json:"order_id" validate:"required"`
}

// ReserveStockHandler 预留库存命令处理器
type ReserveStockHandler struct {
	productRepo    repository.ProductRepository
	eventPublisher *events.EventPublisher
}

// NewReserveStockHandler 创建预留库存命令处理器
func NewReserveStockHandler(
	productRepo repository.ProductRepository,
	eventPublisher *events.EventPublisher,
) *ReserveStockHandler {
	return &ReserveStockHandler{
		productRepo:    productRepo,
		eventPublisher: eventPublisher,
	}
}

// Handle 处理预留库存命令
func (h *ReserveStockHandler) Handle(ctx context.Context, cmd *ReserveStockCommand) error {
	fmt.Printf("[DEBUG] ReserveStockHandler.Handle被调用: productID=%s, quantity=%d\n", cmd.ProductID, cmd.Quantity)

	// 直接调用仓储的ReserveStock方法（使用分布式锁）
	if err := h.productRepo.ReserveStock(ctx, domain.ProductID(cmd.ProductID), cmd.Quantity); err != nil {
		return err
	}

	// 查找商品以获取更新后的库存信息（用于事件发布）
	product, err := h.productRepo.FindByID(ctx, domain.ProductID(cmd.ProductID))
	if err != nil {
		return domain.NewInternalError("failed to find product after stock reservation", err)
	}
	if product == nil {
		return domain.NewNotFoundError("product", cmd.ProductID)
	}

	// 发布库存预留事件
	event := events.NewStockReservedEvent(
		cmd.ProductID,
		cmd.Quantity,
		cmd.OrderID,
		product.Stock(),
	)
	if h.eventPublisher != nil {
		if err := h.eventPublisher.PublishEvents(ctx, event); err != nil {
			// 记录日志，但不影响主流程
		}
	}

	// 发布领域事件
	if h.eventPublisher != nil {
		if err := h.eventPublisher.PublishEvents(ctx, product.DomainEvents()...); err != nil {
			// 事件发布失败不应该影响主流程，但应该记录日志
		}
	}

	// 清除领域事件
	product.ClearDomainEvents()

	return nil
}

// ReleaseStockCommand 释放库存命令
type ReleaseStockCommand struct {
	ProductID string `json:"product_id" validate:"required"`
	Quantity  int    `json:"quantity" validate:"required,gt=0"`
	OrderID   string `json:"order_id" validate:"required"`
}

// ReleaseStockHandler 释放库存命令处理器
type ReleaseStockHandler struct {
	productRepo    repository.ProductRepository
	eventPublisher *events.EventPublisher
}

// NewReleaseStockHandler 创建释放库存命令处理器
func NewReleaseStockHandler(
	productRepo repository.ProductRepository,
	eventPublisher *events.EventPublisher,
) *ReleaseStockHandler {
	return &ReleaseStockHandler{
		productRepo:    productRepo,
		eventPublisher: eventPublisher,
	}
}

// Handle 处理释放库存命令
func (h *ReleaseStockHandler) Handle(ctx context.Context, cmd *ReleaseStockCommand) error {
	// 查找商品
	product, err := h.productRepo.FindByID(ctx, domain.ProductID(cmd.ProductID))
	if err != nil {
		return domain.NewInternalError("failed to find product", err)
	}
	if product == nil {
		return domain.NewNotFoundError("product", cmd.ProductID)
	}

	// 释放库存
	if err := product.ReleaseStock(cmd.Quantity, cmd.OrderID); err != nil {
		return err
	}

	// 保存商品
	if err := h.productRepo.Update(ctx, product); err != nil {
		return domain.NewInternalError("failed to update product", err)
	}

	// 发布库存释放事件
	event := events.NewStockReleasedEvent(
		cmd.ProductID,
		cmd.Quantity,
		cmd.OrderID,
		product.Stock(),
	)
	if h.eventPublisher != nil {
		if err := h.eventPublisher.PublishEvents(ctx, event); err != nil {
			// 记录日志，但不影响主流程
		}
	}

	// 发布领域事件
	if h.eventPublisher != nil {
		if err := h.eventPublisher.PublishEvents(ctx, product.DomainEvents()...); err != nil {
			// 事件发布失败不应该影响主流程，但应该记录日志
		}
	}

	// 清除领域事件
	product.ClearDomainEvents()

	return nil
}

// UpdatePriceCommand 更新价格命令
type UpdatePriceCommand struct {
	ProductID string  `json:"product_id" validate:"required"`
	Price     float64 `json:"price" validate:"required,gt=0"`
}

// UpdatePriceHandler 更新价格命令处理器
type UpdatePriceHandler struct {
	productRepo    repository.ProductRepository
	eventPublisher *events.EventPublisher
}

// NewUpdatePriceHandler 创建更新价格命令处理器
func NewUpdatePriceHandler(
	productRepo repository.ProductRepository,
	eventPublisher *events.EventPublisher,
) *UpdatePriceHandler {
	return &UpdatePriceHandler{
		productRepo:    productRepo,
		eventPublisher: eventPublisher,
	}
}

// Handle 处理更新价格命令
func (h *UpdatePriceHandler) Handle(ctx context.Context, cmd *UpdatePriceCommand) error {
	// 查找商品
	product, err := h.productRepo.FindByID(ctx, domain.ProductID(cmd.ProductID))
	if err != nil {
		return domain.NewInternalError("failed to find product", err)
	}
	if product == nil {
		return domain.NewNotFoundError("product", cmd.ProductID)
	}

	// 记录旧价格
	oldPrice := product.Price().ToYuan()

	// 更新价格
	newPrice := domain.NewMoneyFromYuan(cmd.Price, "CNY")
	if err := product.UpdatePrice(newPrice); err != nil {
		return err
	}

	// 保存商品
	if err := h.productRepo.Update(ctx, product); err != nil {
		return domain.NewInternalError("failed to update product", err)
	}

	// 发布价格更新事件
	event := events.NewPriceUpdatedEvent(
		cmd.ProductID,
		oldPrice,
		cmd.Price,
	)
	if h.eventPublisher != nil {
		if err := h.eventPublisher.PublishEvents(ctx, event); err != nil {
			// 记录日志，但不影响主流程
		}
	}

	// 发布领域事件
	if h.eventPublisher != nil {
		if err := h.eventPublisher.PublishEvents(ctx, product.DomainEvents()...); err != nil {
			// 事件发布失败不应该影响主流程，但应该记录日志
		}
	}

	// 清除领域事件
	product.ClearDomainEvents()

	return nil
}
