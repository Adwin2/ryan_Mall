package command

import (
	"context"

	"ryan-mall-microservices/internal/product/domain/entity"
	"ryan-mall-microservices/internal/product/domain/repository"
	"ryan-mall-microservices/internal/product/domain/service"
	"ryan-mall-microservices/internal/shared/domain"
	"ryan-mall-microservices/internal/shared/events"
)

// CreateProductCommand 创建商品命令
type CreateProductCommand struct {
	Name        string  `json:"name" validate:"required,min=2,max=255"`
	Description string  `json:"description"`
	CategoryID  string  `json:"category_id" validate:"required"`
	Price       float64 `json:"price" validate:"required,gt=0"`
	Stock       int     `json:"stock" validate:"gte=0"`
}

// CreateProductResult 创建商品结果
type CreateProductResult struct {
	ProductID   string  `json:"product_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	CategoryID  string  `json:"category_id"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
}

// CreateProductHandler 创建商品命令处理器
type CreateProductHandler struct {
	productRepo    repository.ProductRepository
	productDomainSvc *service.ProductDomainService
	eventPublisher *events.EventPublisher
}

// NewCreateProductHandler 创建商品命令处理器
func NewCreateProductHandler(
	productRepo repository.ProductRepository,
	productDomainSvc *service.ProductDomainService,
	eventPublisher *events.EventPublisher,
) *CreateProductHandler {
	return &CreateProductHandler{
		productRepo:      productRepo,
		productDomainSvc: productDomainSvc,
		eventPublisher:   eventPublisher,
	}
}

// Handle 处理创建商品命令
func (h *CreateProductHandler) Handle(ctx context.Context, cmd *CreateProductCommand) (*CreateProductResult, error) {
	// 验证商品创建
	if err := h.productDomainSvc.ValidateProductForCreation(ctx, cmd.CategoryID); err != nil {
		return nil, err
	}

	// 创建价格值对象
	price := domain.NewMoneyFromYuan(cmd.Price, "CNY")

	// 创建商品实体
	product, err := entity.NewProduct(cmd.Name, cmd.Description, cmd.CategoryID, price, cmd.Stock)
	if err != nil {
		return nil, err
	}

	// 保存商品
	if err := h.productRepo.Save(ctx, product); err != nil {
		return nil, domain.NewInternalError("failed to save product", err)
	}

	// 发布领域事件
	if h.eventPublisher != nil {
		if err := h.eventPublisher.PublishEvents(ctx, product.DomainEvents()...); err != nil {
			// 事件发布失败不应该影响主流程，但应该记录日志
		}
	}

	// 清除领域事件
	product.ClearDomainEvents()

	return &CreateProductResult{
		ProductID:   product.ID().String(),
		Name:        product.Name(),
		Description: product.Description(),
		CategoryID:  product.CategoryID(),
		Price:       product.Price().ToYuan(),
		Stock:       product.Stock(),
	}, nil
}

// UpdateProductCommand 更新商品命令
type UpdateProductCommand struct {
	ProductID   string  `json:"product_id" validate:"required"`
	Name        string  `json:"name" validate:"required,min=2,max=255"`
	Description string  `json:"description"`
	Price       float64 `json:"price" validate:"required,gt=0"`
}

// UpdateProductHandler 更新商品命令处理器
type UpdateProductHandler struct {
	productRepo    repository.ProductRepository
	eventPublisher *events.EventPublisher
}

// NewUpdateProductHandler 创建更新商品命令处理器
func NewUpdateProductHandler(
	productRepo repository.ProductRepository,
	eventPublisher *events.EventPublisher,
) *UpdateProductHandler {
	return &UpdateProductHandler{
		productRepo:    productRepo,
		eventPublisher: eventPublisher,
	}
}

// Handle 处理更新商品命令
func (h *UpdateProductHandler) Handle(ctx context.Context, cmd *UpdateProductCommand) error {
	// 查找商品
	product, err := h.productRepo.FindByID(ctx, domain.ProductID(cmd.ProductID))
	if err != nil {
		return domain.NewInternalError("failed to find product", err)
	}
	if product == nil {
		return domain.NewNotFoundError("product", cmd.ProductID)
	}

	// 更新商品信息
	if err := product.UpdateInfo(cmd.Name, cmd.Description); err != nil {
		return err
	}

	// 更新价格
	newPrice := domain.NewMoneyFromYuan(cmd.Price, "CNY")
	if err := product.UpdatePrice(newPrice); err != nil {
		return err
	}

	// 保存商品
	if err := h.productRepo.Update(ctx, product); err != nil {
		return domain.NewInternalError("failed to update product", err)
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

// UpdateStockCommand 更新库存命令
type UpdateStockCommand struct {
	ProductID string `json:"product_id" validate:"required"`
	Stock     int    `json:"stock" validate:"gte=0"`
}

// UpdateStockHandler 更新库存命令处理器
type UpdateStockHandler struct {
	productRepo    repository.ProductRepository
	eventPublisher *events.EventPublisher
}

// NewUpdateStockHandler 创建更新库存命令处理器
func NewUpdateStockHandler(
	productRepo repository.ProductRepository,
	eventPublisher *events.EventPublisher,
) *UpdateStockHandler {
	return &UpdateStockHandler{
		productRepo:    productRepo,
		eventPublisher: eventPublisher,
	}
}

// Handle 处理更新库存命令
func (h *UpdateStockHandler) Handle(ctx context.Context, cmd *UpdateStockCommand) error {
	// 查找商品
	product, err := h.productRepo.FindByID(ctx, domain.ProductID(cmd.ProductID))
	if err != nil {
		return domain.NewInternalError("failed to find product", err)
	}
	if product == nil {
		return domain.NewNotFoundError("product", cmd.ProductID)
	}

	// 更新库存
	if err := product.UpdateStock(cmd.Stock); err != nil {
		return err
	}

	// 保存商品
	if err := h.productRepo.Update(ctx, product); err != nil {
		return domain.NewInternalError("failed to update product", err)
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
