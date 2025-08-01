package query

import (
	"context"

	"ryan-mall-microservices/internal/product/domain/repository"
	"ryan-mall-microservices/internal/shared/domain"
)

// GetProductQuery 获取商品查询
type GetProductQuery struct {
	ProductID string `json:"product_id" validate:"required"`
}

// GetProductResult 获取商品结果
type GetProductResult struct {
	ProductID   string  `json:"product_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	CategoryID  string  `json:"category_id"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	SalesCount  int     `json:"sales_count"`
	IsAvailable bool    `json:"is_available"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// GetProductHandler 获取商品查询处理器
type GetProductHandler struct {
	productRepo repository.ProductRepository
}

// NewGetProductHandler 创建获取商品查询处理器
func NewGetProductHandler(productRepo repository.ProductRepository) *GetProductHandler {
	return &GetProductHandler{
		productRepo: productRepo,
	}
}

// Handle 处理获取商品查询
func (h *GetProductHandler) Handle(ctx context.Context, qry *GetProductQuery) (*GetProductResult, error) {
	// 查找商品
	product, err := h.productRepo.FindByID(ctx, domain.ProductID(qry.ProductID))
	if err != nil {
		return nil, domain.NewInternalError("failed to find product", err)
	}
	if product == nil {
		return nil, domain.NewNotFoundError("product", qry.ProductID)
	}

	// 返回结果
	return &GetProductResult{
		ProductID:   product.ID().String(),
		Name:        product.Name(),
		Description: product.Description(),
		CategoryID:  product.CategoryID(),
		Price:       product.Price().ToYuan(),
		Stock:       product.Stock(),
		SalesCount:  product.SalesCount(),
		IsAvailable: product.IsAvailable(),
		CreatedAt:   product.CreatedAt().Time().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   product.UpdatedAt().Time().Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// ListProductsQuery 商品列表查询
type ListProductsQuery struct {
	CategoryID string `json:"category_id"`
	Keyword    string `json:"keyword"`
	Page       int    `json:"page" validate:"min=1"`
	PageSize   int    `json:"page_size" validate:"min=1,max=100"`
}

// ListProductsResult 商品列表结果
type ListProductsResult struct {
	Products   []*GetProductResult `json:"products"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
	TotalPages int                 `json:"total_pages"`
}

// ListProductsHandler 商品列表查询处理器
type ListProductsHandler struct {
	productRepo repository.ProductRepository
}

// NewListProductsHandler 创建商品列表查询处理器
func NewListProductsHandler(productRepo repository.ProductRepository) *ListProductsHandler {
	return &ListProductsHandler{
		productRepo: productRepo,
	}
}

// Handle 处理商品列表查询
func (h *ListProductsHandler) Handle(ctx context.Context, qry *ListProductsQuery) (*ListProductsResult, error) {
	// 设置默认值
	if qry.Page <= 0 {
		qry.Page = 1
	}
	if qry.PageSize <= 0 {
		qry.PageSize = 20
	}

	// 计算偏移量
	offset := (qry.Page - 1) * qry.PageSize

	// 查询商品列表
	products, total, err := h.productRepo.List(ctx, offset, qry.PageSize)
	if err != nil {
		return nil, domain.NewInternalError("failed to list products", err)
	}

	// 转换结果
	productResults := make([]*GetProductResult, len(products))
	for i, product := range products {
		productResults[i] = &GetProductResult{
			ProductID:   product.ID().String(),
			Name:        product.Name(),
			Description: product.Description(),
			CategoryID:  product.CategoryID(),
			Price:       product.Price().ToYuan(),
			Stock:       product.Stock(),
			SalesCount:  product.SalesCount(),
			IsAvailable: product.IsAvailable(),
			CreatedAt:   product.CreatedAt().Time().Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   product.UpdatedAt().Time().Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	// 计算总页数
	totalPages := int((total + int64(qry.PageSize) - 1) / int64(qry.PageSize))

	return &ListProductsResult{
		Products:   productResults,
		Total:      total,
		Page:       qry.Page,
		PageSize:   qry.PageSize,
		TotalPages: totalPages,
	}, nil
}

// CheckStockQuery 检查库存查询
type CheckStockQuery struct {
	ProductID string `json:"product_id" validate:"required"`
}

// CheckStockResult 检查库存结果
type CheckStockResult struct {
	ProductID   string `json:"product_id"`
	Stock       int    `json:"stock"`
	IsAvailable bool   `json:"is_available"`
}

// CheckStockHandler 检查库存查询处理器
type CheckStockHandler struct {
	productRepo repository.ProductRepository
}

// NewCheckStockHandler 创建检查库存查询处理器
func NewCheckStockHandler(productRepo repository.ProductRepository) *CheckStockHandler {
	return &CheckStockHandler{
		productRepo: productRepo,
	}
}

// Handle 处理检查库存查询
func (h *CheckStockHandler) Handle(ctx context.Context, qry *CheckStockQuery) (*CheckStockResult, error) {
	// 查找商品
	product, err := h.productRepo.FindByID(ctx, domain.ProductID(qry.ProductID))
	if err != nil {
		return nil, domain.NewInternalError("failed to find product", err)
	}
	if product == nil {
		return nil, domain.NewNotFoundError("product", qry.ProductID)
	}

	// 返回库存信息
	return &CheckStockResult{
		ProductID:   product.ID().String(),
		Stock:       product.Stock(),
		IsAvailable: product.IsAvailable() && product.Stock() > 0,
	}, nil
}
