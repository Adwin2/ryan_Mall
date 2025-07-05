package service

import (
	"fmt"
	"ryan-mall/internal/model"
	"ryan-mall/internal/repository"
	"ryan-mall/pkg/cache"
	"time"
)

// CachedProductService 带缓存的商品服务
type CachedProductService struct {
	productRepo  repository.ProductRepository
	categoryRepo repository.CategoryRepository
	cache        cache.CacheManager
}

// NewCachedProductService 创建带缓存的商品服务
func NewCachedProductService(
	productRepo repository.ProductRepository,
	categoryRepo repository.CategoryRepository,
) *CachedProductService {
	return &CachedProductService{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
		cache:        cache.GetCache(), // 获取全局缓存实例
	}
}

// GetByID 获取商品详情（带缓存）
func (s *CachedProductService) GetByID(id uint) (*model.Product, error) {
	// 1. 尝试从缓存获取
	cacheKey := fmt.Sprintf("product:%d", id)
	var product model.Product
	if err := s.cache.GetJSON(cacheKey, &product); err == nil {
		// 缓存命中，记录浏览次数
		go s.incrementViewCount(id)
		return &product, nil
	}
	
	// 2. 缓存未命中，从数据库获取
	productPtr, err := s.productRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	
	if productPtr == nil {
		return nil, fmt.Errorf("product not found")
	}
	
	// 3. 存入缓存（5分钟过期）
	s.cache.SetJSON(cacheKey, productPtr, 5*time.Minute)
	
	// 4. 异步记录浏览次数
	go s.incrementViewCount(id)
	
	return productPtr, nil
}

// List 获取商品列表（带缓存）
func (s *CachedProductService) List(req *model.ProductListRequest) ([]*model.Product, int64, error) {
	// 生成缓存键
	cacheKey := s.generateListCacheKey(req)
	
	// 1. 尝试从缓存获取
	type CachedListResult struct {
		Products []*model.Product `json:"products"`
		Total    int64            `json:"total"`
	}
	
	var cached CachedListResult
	if err := s.cache.GetJSON(cacheKey, &cached); err == nil {
		return cached.Products, cached.Total, nil
	}
	
	// 2. 缓存未命中，从数据库获取
	products, total, err := s.productRepo.List(req)
	if err != nil {
		return nil, 0, err
	}
	
	// 3. 存入缓存（2分钟过期，列表数据变化较快）
	cached = CachedListResult{
		Products: products,
		Total:    total,
	}
	s.cache.SetJSON(cacheKey, cached, 2*time.Minute)
	
	return products, total, nil
}

// CreateProduct 创建商品
func (s *CachedProductService) CreateProduct(req *model.ProductCreateRequest) (*model.Product, error) {
	// 转换请求为商品模型
	product := &model.Product{
		Name:        req.Name,
		Description: &req.Description,
		CategoryID:  req.CategoryID,
		Price:       req.Price,
		Stock:       req.Stock,
		Status:      1, // 默认上架
		MainImage:   &req.MainImage,
		Images:      req.Images,
	}

	err := s.productRepo.Create(product)
	if err != nil {
		return nil, err
	}

	// 清除相关缓存
	s.clearProductCaches()

	return product, nil
}

// GetProduct 获取商品详情（实现接口）
func (s *CachedProductService) GetProduct(id uint) (*model.Product, error) {
	return s.GetByID(id)
}

// UpdateProduct 更新商品
func (s *CachedProductService) UpdateProduct(id uint, req *model.ProductUpdateRequest) error {
	// 先获取现有商品
	existingProduct, err := s.productRepo.GetByID(id)
	if err != nil {
		return err
	}
	if existingProduct == nil {
		return fmt.Errorf("product not found")
	}

	// 更新字段（只更新非nil的字段）
	if req.Name != nil {
		existingProduct.Name = *req.Name
	}
	if req.Description != nil {
		existingProduct.Description = req.Description
	}
	if req.CategoryID != nil {
		existingProduct.CategoryID = *req.CategoryID
	}
	if req.Price != nil {
		existingProduct.Price = *req.Price
	}
	if req.Stock != nil {
		existingProduct.Stock = *req.Stock
	}
	if req.MainImage != nil {
		existingProduct.MainImage = req.MainImage
	}
	if req.Images != nil {
		existingProduct.Images = req.Images
	}

	err = s.productRepo.Update(existingProduct)
	if err != nil {
		return err
	}

	// 清除相关缓存
	s.clearProductCache(id)
	s.clearProductCaches()

	return nil
}

// DeleteProduct 删除商品
func (s *CachedProductService) DeleteProduct(id uint) error {
	err := s.productRepo.Delete(id)
	if err != nil {
		return err
	}

	// 清除相关缓存
	s.clearProductCache(id)
	s.clearProductCaches()

	return nil
}

// GetProductList 获取商品列表（实现接口）
func (s *CachedProductService) GetProductList(req *model.ProductListRequest) (*model.ProductListResponse, error) {
	products, total, err := s.List(req)
	if err != nil {
		return nil, err
	}

	// 计算总页数
	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))

	return &model.ProductListResponse{
		Products:   products,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetProductsByCategory 根据分类获取商品
func (s *CachedProductService) GetProductsByCategory(categoryID uint) ([]*model.Product, error) {
	req := &model.ProductListRequest{
		CategoryID: &categoryID,
		Page:       1,
		PageSize:   100, // 获取前100个商品
		SortBy:     "created_at",
		SortOrder:  "desc",
	}

	products, _, err := s.List(req)
	return products, err
}

// DecrementStock 减少库存（下单时使用）
func (s *CachedProductService) DecrementStock(id uint, quantity int) error {
	// 先获取当前库存
	product, err := s.productRepo.GetByID(id)
	if err != nil {
		return err
	}
	if product == nil {
		return fmt.Errorf("product not found")
	}

	// 检查库存是否充足
	if product.Stock < quantity {
		return fmt.Errorf("insufficient stock")
	}

	// 更新库存
	newStock := product.Stock - quantity
	err = s.UpdateStock(id, newStock)
	if err != nil {
		return err
	}

	return nil
}

// IncrementSalesCount 增加销售数量
func (s *CachedProductService) IncrementSalesCount(id uint, quantity int) error {
	// 获取当前商品
	product, err := s.productRepo.GetByID(id)
	if err != nil {
		return err
	}
	if product == nil {
		return fmt.Errorf("product not found")
	}

	// 更新销售数量
	product.SalesCount += quantity
	err = s.productRepo.Update(product)
	if err != nil {
		return err
	}

	// 清除缓存
	s.clearProductCache(id)
	s.clearProductCaches()

	return nil
}

// UpdateStock 更新库存
func (s *CachedProductService) UpdateStock(id uint, stock int) error {
	err := s.productRepo.UpdateStock(id, stock)
	if err != nil {
		return err
	}
	
	// 清除商品缓存
	s.clearProductCache(id)
	
	return nil
}

// GetHotProducts 获取热门商品（带缓存）
func (s *CachedProductService) GetHotProducts(limit int) ([]*model.Product, error) {
	cacheKey := fmt.Sprintf("hot_products:%d", limit)
	
	// 1. 尝试从缓存获取
	var products []*model.Product
	if err := s.cache.GetJSON(cacheKey, &products); err == nil {
		return products, nil
	}
	
	// 2. 从数据库获取（按销量排序）
	req := &model.ProductListRequest{
		Page:      1,
		PageSize:  limit,
		SortBy:    "sales_count",
		SortOrder: "desc",
	}
	
	products, _, err := s.productRepo.List(req)
	if err != nil {
		return nil, err
	}
	
	// 3. 存入缓存（10分钟过期）
	s.cache.SetJSON(cacheKey, products, 10*time.Minute)
	
	return products, nil
}

// SearchProducts 搜索商品（带缓存）
func (s *CachedProductService) SearchProducts(keyword string, page, pageSize int) ([]*model.Product, int64, error) {
	cacheKey := fmt.Sprintf("search:%s:%d:%d", keyword, page, pageSize)
	
	// 1. 尝试从缓存获取
	type SearchResult struct {
		Products []*model.Product `json:"products"`
		Total    int64            `json:"total"`
	}
	
	var cached SearchResult
	if err := s.cache.GetJSON(cacheKey, &cached); err == nil {
		return cached.Products, cached.Total, nil
	}
	
	// 2. 从数据库搜索
	req := &model.ProductListRequest{
		Keyword:  keyword,
		Page:     page,
		PageSize: pageSize,
	}
	
	products, total, err := s.productRepo.List(req)
	if err != nil {
		return nil, 0, err
	}
	
	// 3. 存入缓存（5分钟过期）
	cached = SearchResult{
		Products: products,
		Total:    total,
	}
	s.cache.SetJSON(cacheKey, cached, 5*time.Minute)
	
	return products, total, nil
}

// GetCacheStats 获取缓存统计
func (s *CachedProductService) GetCacheStats() map[string]interface{} {
	stats := s.cache.Stats()
	return map[string]interface{}{
		"cache_size": s.cache.Size(),
		"cache_type": stats["cache_type"],
		"cache_stats": stats,
	}
}

// 私有方法

// generateListCacheKey 生成列表缓存键
func (s *CachedProductService) generateListCacheKey(req *model.ProductListRequest) string {
	key := fmt.Sprintf("product_list:%d:%d", req.Page, req.PageSize)
	
	if req.Keyword != "" {
		key += fmt.Sprintf(":kw:%s", req.Keyword)
	}
	
	if req.CategoryID != nil {
		key += fmt.Sprintf(":cat:%d", *req.CategoryID)
	}
	
	if req.MinPrice != nil {
		key += fmt.Sprintf(":minp:%.2f", *req.MinPrice)
	}
	
	if req.MaxPrice != nil {
		key += fmt.Sprintf(":maxp:%.2f", *req.MaxPrice)
	}
	
	if req.SortBy != "" {
		key += fmt.Sprintf(":sort:%s", req.SortBy)
		if req.SortOrder == "desc" {
			key += ":desc"
		}
	}
	
	return key
}

// clearProductCache 清除单个商品缓存
func (s *CachedProductService) clearProductCache(id uint) {
	cacheKey := fmt.Sprintf("product:%d", id)
	s.cache.Delete(cacheKey)
}

// clearProductCaches 清除商品相关缓存
func (s *CachedProductService) clearProductCaches() {
	// 这里简单实现，实际可以使用更精确的缓存失效策略
	// 清除热门商品缓存
	for i := 1; i <= 20; i++ {
		s.cache.Delete(fmt.Sprintf("hot_products:%d", i))
	}
	
	// 注意：在生产环境中，应该使用更智能的缓存失效策略
	// 比如使用缓存标签或者模式匹配删除
}

// incrementViewCount 增加浏览次数
func (s *CachedProductService) incrementViewCount(productID uint) {
	// 使用缓存计数器
	countKey := fmt.Sprintf("product_view_count:%d", productID)
	
	// 获取当前计数
	var count int
	if err := s.cache.GetJSON(countKey, &count); err != nil {
		count = 0
	}
	
	count++
	
	// 更新缓存计数（1小时过期）
	s.cache.SetJSON(countKey, count, 1*time.Hour)
	
	// 每10次浏览同步一次到数据库
	if count%10 == 0 {
		// 这里可以异步更新数据库
		// s.productRepo.IncrementViewCount(productID, 10)
	}
}
