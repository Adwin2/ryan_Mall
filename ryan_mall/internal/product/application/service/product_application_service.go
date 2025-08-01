package service

import (
	"context"

	"ryan-mall-microservices/internal/product/application/command"
	"ryan-mall-microservices/internal/product/application/query"
	"ryan-mall-microservices/internal/product/domain/repository"
	"ryan-mall-microservices/internal/shared/events"
)

// ProductApplicationService 商品应用服务
type ProductApplicationService struct {
	// 命令处理器
	createProductHandler  *command.CreateProductHandler
	updateStockHandler    *command.UpdateStockHandler
	reserveStockHandler   *command.ReserveStockHandler
	releaseStockHandler   *command.ReleaseStockHandler
	updatePriceHandler    *command.UpdatePriceHandler

	// 查询处理器
	getProductHandler   *query.GetProductHandler
	listProductsHandler *query.ListProductsHandler
	checkStockHandler   *query.CheckStockHandler
}

// NewProductApplicationService 创建商品应用服务
func NewProductApplicationService(
	productRepo repository.ProductRepository,
	eventPublisher *events.EventPublisher,
) *ProductApplicationService {
	return &ProductApplicationService{
		// 初始化命令处理器
		updateStockHandler:    command.NewUpdateStockHandler(productRepo, eventPublisher),
		reserveStockHandler:   command.NewReserveStockHandler(productRepo, eventPublisher),
		releaseStockHandler:   command.NewReleaseStockHandler(productRepo, eventPublisher),
		updatePriceHandler:    command.NewUpdatePriceHandler(productRepo, eventPublisher),

		// 初始化查询处理器
		getProductHandler:   query.NewGetProductHandler(productRepo),
		listProductsHandler: query.NewListProductsHandler(productRepo),
		checkStockHandler:   query.NewCheckStockHandler(productRepo),
	}
}

// CreateProduct 创建商品
func (s *ProductApplicationService) CreateProduct(ctx context.Context, cmd *command.CreateProductCommand) (*command.CreateProductResult, error) {
	return s.createProductHandler.Handle(ctx, cmd)
}

// UpdateStock 更新库存
func (s *ProductApplicationService) UpdateStock(ctx context.Context, cmd *command.UpdateStockCommand) error {
	return s.updateStockHandler.Handle(ctx, cmd)
}

// ReserveStock 预留库存（用于订单创建）
func (s *ProductApplicationService) ReserveStock(ctx context.Context, cmd *command.ReserveStockCommand) error {
	return s.reserveStockHandler.Handle(ctx, cmd)
}

// ReleaseStock 释放库存（用于订单取消）
func (s *ProductApplicationService) ReleaseStock(ctx context.Context, cmd *command.ReleaseStockCommand) error {
	return s.releaseStockHandler.Handle(ctx, cmd)
}

// UpdatePrice 更新价格
func (s *ProductApplicationService) UpdatePrice(ctx context.Context, cmd *command.UpdatePriceCommand) error {
	return s.updatePriceHandler.Handle(ctx, cmd)
}

// GetProduct 获取商品信息
func (s *ProductApplicationService) GetProduct(ctx context.Context, qry *query.GetProductQuery) (*query.GetProductResult, error) {
	return s.getProductHandler.Handle(ctx, qry)
}

// ListProducts 获取商品列表
func (s *ProductApplicationService) ListProducts(ctx context.Context, qry *query.ListProductsQuery) (*query.ListProductsResult, error) {
	return s.listProductsHandler.Handle(ctx, qry)
}

// CheckStock 检查库存
func (s *ProductApplicationService) CheckStock(ctx context.Context, qry *query.CheckStockQuery) (*query.CheckStockResult, error) {
	return s.checkStockHandler.Handle(ctx, qry)
}
