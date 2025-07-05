package service

import (
	"errors"
	"math"
	"ryan-mall/internal/model"
	"ryan-mall/internal/repository"
)

// ProductService 商品业务逻辑层接口
type ProductService interface {
	CreateProduct(req *model.ProductCreateRequest) (*model.Product, error)           // 创建商品
	GetProduct(id uint) (*model.Product, error)                                     // 获取商品详情
	UpdateProduct(id uint, req *model.ProductUpdateRequest) error                   // 更新商品
	DeleteProduct(id uint) error                                                    // 删除商品
	GetProductList(req *model.ProductListRequest) (*model.ProductListResponse, error) // 获取商品列表
	GetProductsByCategory(categoryID uint) ([]*model.Product, error)                // 根据分类获取商品
	UpdateStock(id uint, stock int) error                                           // 更新库存
	DecrementStock(id uint, quantity int) error                                     // 减少库存（下单时使用）
	IncrementSalesCount(id uint, quantity int) error                                // 增加销售数量
}

// productService 商品业务逻辑层实现
type productService struct {
	productRepo  repository.ProductRepository
	categoryRepo repository.CategoryRepository
}

// NewProductService 创建商品业务逻辑层实例
func NewProductService(productRepo repository.ProductRepository, categoryRepo repository.CategoryRepository) ProductService {
	return &productService{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
	}
}

// CreateProduct 创建商品
func (s *productService) CreateProduct(req *model.ProductCreateRequest) (*model.Product, error) {
	// 1. 验证分类是否存在
	category, err := s.categoryRepo.GetByID(req.CategoryID)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, errors.New("商品分类不存在")
	}
	
	// 2. 创建商品对象
	product := &model.Product{
		Name:          req.Name,
		CategoryID:    req.CategoryID,
		Price:         req.Price,
		Stock:         req.Stock,
		Status:        model.ProductStatusOnline, // 默认上架
	}
	
	// 处理可选字段
	if req.Description != "" {
		product.Description = &req.Description
	}
	if req.OriginalPrice != nil {
		product.OriginalPrice = req.OriginalPrice
	}
	if req.MainImage != "" {
		product.MainImage = &req.MainImage
	}
	if len(req.Images) > 0 {
		product.Images = model.JSONArray(req.Images)
	}
	
	// 3. 保存商品
	if err := s.productRepo.Create(product); err != nil {
		return nil, err
	}
	
	// 4. 返回创建的商品（包含分类信息）
	return s.GetProduct(product.ID)
}

// GetProduct 获取商品详情
func (s *productService) GetProduct(id uint) (*model.Product, error) {
	product, err := s.productRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, errors.New("商品不存在")
	}
	
	return product, nil
}

// UpdateProduct 更新商品
func (s *productService) UpdateProduct(id uint, req *model.ProductUpdateRequest) error {
	// 1. 获取现有商品
	product, err := s.productRepo.GetByID(id)
	if err != nil {
		return err
	}
	if product == nil {
		return errors.New("商品不存在")
	}
	
	// 2. 验证分类（如果要更新分类）
	if req.CategoryID != nil {
		category, err := s.categoryRepo.GetByID(*req.CategoryID)
		if err != nil {
			return err
		}
		if category == nil {
			return errors.New("商品分类不存在")
		}
		product.CategoryID = *req.CategoryID
	}
	
	// 3. 更新字段
	if req.Name != nil {
		product.Name = *req.Name
	}
	if req.Description != nil {
		product.Description = req.Description
	}
	if req.Price != nil {
		product.Price = *req.Price
	}
	if req.OriginalPrice != nil {
		product.OriginalPrice = req.OriginalPrice
	}
	if req.Stock != nil {
		product.Stock = *req.Stock
	}
	if req.MainImage != nil {
		product.MainImage = req.MainImage
	}
	if req.Images != nil {
		product.Images = model.JSONArray(req.Images)
	}
	if req.Status != nil {
		product.Status = *req.Status
	}
	
	// 4. 保存更新
	return s.productRepo.Update(product)
}

// DeleteProduct 删除商品
func (s *productService) DeleteProduct(id uint) error {
	// 1. 检查商品是否存在
	product, err := s.productRepo.GetByID(id)
	if err != nil {
		return err
	}
	if product == nil {
		return errors.New("商品不存在")
	}
	
	// 2. 执行软删除
	return s.productRepo.Delete(id)
}

// GetProductList 获取商品列表
func (s *productService) GetProductList(req *model.ProductListRequest) (*model.ProductListResponse, error) {
	// 1. 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100 // 限制最大页面大小
	}
	if req.SortBy == "" {
		req.SortBy = model.SortByCreatedAt
	}
	if req.SortOrder == "" {
		req.SortOrder = model.SortOrderDesc
	}
	
	// 2. 查询商品列表
	products, total, err := s.productRepo.List(req)
	if err != nil {
		return nil, err
	}
	
	// 3. 计算总页数
	totalPages := int(math.Ceil(float64(total) / float64(req.PageSize)))
	
	// 4. 构建响应
	response := &model.ProductListResponse{
		Products:   products,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}
	
	return response, nil
}

// GetProductsByCategory 根据分类获取商品
func (s *productService) GetProductsByCategory(categoryID uint) ([]*model.Product, error) {
	// 1. 验证分类是否存在
	category, err := s.categoryRepo.GetByID(categoryID)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, errors.New("商品分类不存在")
	}
	
	// 2. 获取分类下的商品
	return s.productRepo.GetByCategoryID(categoryID)
}

// UpdateStock 更新库存
func (s *productService) UpdateStock(id uint, stock int) error {
	// 1. 验证库存数量
	if stock < 0 {
		return errors.New("库存数量不能为负数")
	}
	
	// 2. 检查商品是否存在
	product, err := s.productRepo.GetByID(id)
	if err != nil {
		return err
	}
	if product == nil {
		return errors.New("商品不存在")
	}
	
	// 3. 更新库存
	return s.productRepo.UpdateStock(id, stock)
}

// DecrementStock 减少库存（下单时使用）
func (s *productService) DecrementStock(id uint, quantity int) error {
	// 1. 验证数量
	if quantity <= 0 {
		return errors.New("数量必须大于0")
	}
	
	// 2. 获取当前商品信息
	product, err := s.productRepo.GetByID(id)
	if err != nil {
		return err
	}
	if product == nil {
		return errors.New("商品不存在")
	}
	
	// 3. 检查库存是否充足
	if product.Stock < quantity {
		return errors.New("库存不足")
	}
	
	// 4. 减少库存
	newStock := product.Stock - quantity
	return s.productRepo.UpdateStock(id, newStock)
}

// IncrementSalesCount 增加销售数量
func (s *productService) IncrementSalesCount(id uint, quantity int) error {
	// 1. 验证数量
	if quantity <= 0 {
		return errors.New("数量必须大于0")
	}
	
	// 2. 检查商品是否存在
	product, err := s.productRepo.GetByID(id)
	if err != nil {
		return err
	}
	if product == nil {
		return errors.New("商品不存在")
	}
	
	// 3. 增加销售数量
	return s.productRepo.UpdateSalesCount(id, quantity)
}
