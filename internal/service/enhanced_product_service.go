package service

import (
	"encoding/json"
	"fmt"
	"ryan-mall/internal/model"
	"ryan-mall/internal/repository"
	"ryan-mall/pkg/elasticsearch"
	"ryan-mall/pkg/monitoring"
	"ryan-mall/pkg/redis"
	"ryan-mall/pkg/websocket"
	"strconv"
	"time"
)

// EnhancedProductService 增强版商品服务
// 集成了Redis缓存、Elasticsearch搜索、WebSocket通知、监控指标
type EnhancedProductService struct {
	// 基础服务
	productRepo  repository.ProductRepository
	categoryRepo repository.CategoryRepository
	
	// 增强功能
	redisManager     *redis.RedisManager
	searchEngine     *elasticsearch.SearchEngine
	wsManager        *websocket.WebSocketManager
	metricsManager   *monitoring.MetricsManager
	
	// 缓存管理器
	cacheManager     *redis.CacheManager
	hotDataManager   *redis.HotDataManager
	
	// 通知服务
	notificationService *websocket.NotificationService
}

// NewEnhancedProductService 创建增强版商品服务
func NewEnhancedProductService(
	productRepo repository.ProductRepository,
	categoryRepo repository.CategoryRepository,
	redisManager *redis.RedisManager,
	searchEngine *elasticsearch.SearchEngine,
	wsManager *websocket.WebSocketManager,
	metricsManager *monitoring.MetricsManager,
) *EnhancedProductService {
	return &EnhancedProductService{
		productRepo:         productRepo,
		categoryRepo:        categoryRepo,
		redisManager:        redisManager,
		searchEngine:        searchEngine,
		wsManager:          wsManager,
		metricsManager:     metricsManager,
		cacheManager:       redis.NewCacheManager(redisManager),
		hotDataManager:     redis.NewHotDataManager(redisManager),
		notificationService: websocket.NewNotificationService(wsManager),
	}
}

// GetProduct 获取商品详情（带缓存）
func (s *EnhancedProductService) GetProduct(id uint) (*model.Product, error) {
	// 记录监控指标
	defer func() {
		s.metricsManager.RecordProductView(id, "unknown")
	}()
	
	// 1. 尝试从缓存获取
	cacheKey := fmt.Sprintf("product:%d", id)
	cachedData, err := s.cacheManager.GetMulti([]string{cacheKey})
	if err == nil && len(cachedData) > 0 {
		var product model.Product
		if err := json.Unmarshal([]byte(cachedData[cacheKey]), &product); err == nil {
			// 异步更新浏览次数
			go s.updateViewCount(id)
			return &product, nil
		}
	}
	
	// 2. 从数据库获取
	product, err := s.productRepo.GetByID(id)
	if err != nil {
		s.metricsManager.RecordError("database_error", "product_service")
		return nil, err
	}
	
	if product == nil {
		return nil, fmt.Errorf("product not found")
	}
	
	// 3. 更新缓存
	go s.updateProductCache(product)
	
	// 4. 异步更新浏览次数和热度
	go s.updateViewCount(id)
	go s.updateHotRanking(id, product.SalesCount)
	
	// 5. 记录监控指标
	if product.Category.ID != 0 {
		s.metricsManager.RecordProductView(id, product.Category.Name)
	}
	
	return product, nil
}

// SearchProducts 智能商品搜索
func (s *EnhancedProductService) SearchProducts(req *elasticsearch.SearchRequest) (*elasticsearch.SearchResponse, error) {
	// 1. 使用Elasticsearch进行搜索
	response, err := s.searchEngine.SearchProducts(req)
	if err != nil {
		s.metricsManager.RecordError("search_error", "product_service")
		// 降级到数据库搜索
		return s.fallbackSearch(req)
	}
	
	// 2. 异步更新搜索热词
	if req.Query != "" {
		go s.updateSearchHotwords(req.Query)
	}
	
	return response, nil
}

// CreateProduct 创建商品（增强版）
func (s *EnhancedProductService) CreateProduct(product *model.Product) error {
	// 1. 创建商品
	err := s.productRepo.Create(product)
	if err != nil {
		s.metricsManager.RecordError("database_error", "product_service")
		return err
	}
	
	// 2. 异步索引到Elasticsearch
	go s.indexProductToES(product)
	
	// 3. 异步更新缓存
	go s.updateProductCache(product)
	
	// 4. 发送WebSocket通知
	go s.notifyProductCreated(product)
	
	return nil
}

// UpdateStock 更新库存（原子操作）
func (s *EnhancedProductService) UpdateStock(productID uint, quantity int) error {
	// 1. 使用Redis Lua脚本原子性更新库存
	inventoryManager := redis.NewInventoryManager(s.redisManager)
	newStock, err := inventoryManager.DecreaseStock(productID, quantity)
	if err != nil {
		s.metricsManager.RecordError("inventory_error", "product_service")
		return err
	}
	
	// 2. 同步到数据库
	err = s.productRepo.UpdateStock(productID, newStock)
	if err != nil {
		// 如果数据库更新失败，需要回滚Redis
		// 这里可以实现补偿机制
		s.metricsManager.RecordError("database_error", "product_service")
		return err
	}
	
	// 3. 清除相关缓存
	go s.clearProductCache(productID)
	
	// 4. 发送库存更新通知
	go s.notifyStockUpdate(productID, newStock)
	
	// 5. 更新Elasticsearch索引
	go s.updateProductStockInES(productID, newStock)
	
	return nil
}

// GetRecommendations 获取推荐商品
func (s *EnhancedProductService) GetRecommendations(userID, productID uint, size int) ([]model.Product, error) {
	// 1. 尝试从缓存获取
	cacheKey := fmt.Sprintf("recommendations:user:%d:product:%d", userID, productID)
	cachedData, err := s.cacheManager.GetMulti([]string{cacheKey})
	if err == nil && len(cachedData) > 0 {
		var products []model.Product
		if err := json.Unmarshal([]byte(cachedData[cacheKey]), &products); err == nil {
			return products, nil
		}
	}
	
	// 2. 使用Elasticsearch获取相似商品
	recommendEngine := elasticsearch.NewRecommendationEngine(s.searchEngine)
	esProducts, err := recommendEngine.GetSimilarProducts(productID, size)
	if err != nil {
		// 降级到基于分类的推荐
		return s.fallbackRecommendations(productID, size)
	}
	
	// 3. 转换为内部模型
	var products []model.Product
	for _, esProduct := range esProducts {
		product := s.convertESProductToModel(&esProduct)
		products = append(products, *product)
	}
	
	// 4. 异步更新缓存
	go s.updateRecommendationCache(cacheKey, products)
	
	return products, nil
}

// GetHotProducts 获取热门商品
func (s *EnhancedProductService) GetHotProducts(categoryID uint, size int) ([]model.Product, error) {
	// 1. 从Redis获取热门商品排行
	hotProductIDs, err := s.hotDataManager.GetHotProducts(size)
	if err != nil || len(hotProductIDs) == 0 {
		// 降级到Elasticsearch
		return s.getHotProductsFromES(categoryID, size)
	}
	
	// 2. 批量获取商品详情
	var products []model.Product
	for _, productIDStr := range hotProductIDs {
		if productID, err := strconv.ParseUint(productIDStr[8:], 10, 32); err == nil {
			if product, err := s.GetProduct(uint(productID)); err == nil {
				if categoryID == 0 || product.CategoryID == categoryID {
					products = append(products, *product)
				}
			}
		}
	}
	
	return products, nil
}

// 私有方法

// updateViewCount 更新浏览次数
func (s *EnhancedProductService) updateViewCount(productID uint) {
	// 1. 更新Redis中的浏览次数
	count, err := s.hotDataManager.IncrementViewCount(productID)
	if err != nil {
		return
	}
	
	// 2. 每100次浏览同步一次到数据库
	if count%100 == 0 {
		// TODO: 实现IncrementViewCount方法
		// s.productRepo.IncrementViewCount(productID, 100)
	}
}

// updateHotRanking 更新热门排行
func (s *EnhancedProductService) updateHotRanking(productID uint, salesCount int) {
	// 计算热度分数（浏览量 + 销量*10）
	viewCount, _ := s.hotDataManager.IncrementViewCount(productID)
	score := float64(viewCount) + float64(salesCount)*10
	
	s.hotDataManager.UpdateHotProductsRanking(productID, score)
}

// updateProductCache 更新商品缓存
func (s *EnhancedProductService) updateProductCache(product *model.Product) {
	cacheKey := fmt.Sprintf("product:%d", product.ID)
	data, err := json.Marshal(product)
	if err != nil {
		return
	}
	
	cacheData := map[string]interface{}{
		cacheKey: string(data),
	}
	s.cacheManager.BatchSet(cacheData, 1*time.Hour)
}

// clearProductCache 清除商品缓存
func (s *EnhancedProductService) clearProductCache(productID uint) {
	cacheKey := fmt.Sprintf("product:%d", productID)
	// TODO: 实现删除缓存的方法
	_ = cacheKey
}

// indexProductToES 索引商品到Elasticsearch
func (s *EnhancedProductService) indexProductToES(product *model.Product) {
	esProduct := &elasticsearch.ProductDocument{
		ID:          product.ID,
		Name:        product.Name,
		Description: *product.Description,
		CategoryID:  product.CategoryID,
		Price:       product.Price,
		Stock:       product.Stock,
		Status:      int(product.Status),
		SalesCount:  product.SalesCount,
		CreatedAt:   product.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   product.UpdatedAt.Format(time.RFC3339),
	}
	
	if product.Category.ID != 0 {
		esProduct.Category = product.Category.Name
	}
	
	s.searchEngine.IndexProduct(esProduct)
}

// updateProductStockInES 更新ES中的库存
func (s *EnhancedProductService) updateProductStockInES(productID uint, stock int) {
	// 这里需要实现ES的部分更新
}

// notifyProductCreated 通知商品创建
func (s *EnhancedProductService) notifyProductCreated(product *model.Product) {
	_ = map[string]interface{}{
		"type":         "product_created",
		"product_id":   product.ID,
		"product_name": product.Name,
		"message":      fmt.Sprintf("新商品 %s 已上架", product.Name),
	}

	s.notificationService.SendSystemAlert(fmt.Sprintf("新商品 %s 已上架", product.Name))
}

// notifyStockUpdate 通知库存更新
func (s *EnhancedProductService) notifyStockUpdate(productID uint, newStock int) {
	// 获取商品信息
	product, err := s.productRepo.GetByID(productID)
	if err != nil {
		return
	}
	
	notification := &websocket.StockUpdateNotification{
		ProductID:   productID,
		ProductName: product.Name,
		NewStock:    newStock,
		Message:     fmt.Sprintf("商品 %s 库存已更新为 %d", product.Name, newStock),
	}
	
	s.notificationService.SendStockUpdateNotification(notification)
}

// updateSearchHotwords 更新搜索热词
func (s *EnhancedProductService) updateSearchHotwords(query string) {
	// 实现搜索热词统计
}

// fallbackSearch 降级搜索
func (s *EnhancedProductService) fallbackSearch(req *elasticsearch.SearchRequest) (*elasticsearch.SearchResponse, error) {
	// 使用数据库进行基础搜索
	searchReq := &model.ProductListRequest{
		Keyword:  req.Query,
		Page:     req.Page,
		PageSize: req.PageSize,
	}
	
	productPtrs, total, err := s.productRepo.List(searchReq)
	if err != nil {
		return nil, err
	}

	// 转换为ES响应格式
	var esProducts []elasticsearch.ProductDocument
	for _, productPtr := range productPtrs {
		esProduct := s.convertModelToESProduct(productPtr)
		esProducts = append(esProducts, *esProduct)
	}
	
	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))
	
	return &elasticsearch.SearchResponse{
		Products:   esProducts,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
		TimeTaken:  "0ms",
	}, nil
}

// fallbackRecommendations 降级推荐
func (s *EnhancedProductService) fallbackRecommendations(productID uint, size int) ([]model.Product, error) {
	// 基于分类的简单推荐
	product, err := s.productRepo.GetByID(productID)
	if err != nil {
		return nil, err
	}
	
	req := &model.ProductListRequest{
		CategoryID: &product.CategoryID,
		Page:       1,
		PageSize:   size,
	}
	
	productPtrs, _, err := s.productRepo.List(req)
	if err != nil {
		return nil, err
	}

	// 转换指针切片为值切片
	var products []model.Product
	for _, productPtr := range productPtrs {
		products = append(products, *productPtr)
	}
	return products, nil
}

// getHotProductsFromES 从ES获取热门商品
func (s *EnhancedProductService) getHotProductsFromES(categoryID uint, size int) ([]model.Product, error) {
	recommendEngine := elasticsearch.NewRecommendationEngine(s.searchEngine)
	esProducts, err := recommendEngine.GetHotProducts(categoryID, size)
	if err != nil {
		return nil, err
	}
	
	var products []model.Product
	for _, esProduct := range esProducts {
		product := s.convertESProductToModel(&esProduct)
		products = append(products, *product)
	}
	
	return products, nil
}

// updateRecommendationCache 更新推荐缓存
func (s *EnhancedProductService) updateRecommendationCache(cacheKey string, products []model.Product) {
	data, err := json.Marshal(products)
	if err != nil {
		return
	}
	
	cacheData := map[string]interface{}{
		cacheKey: string(data),
	}
	s.cacheManager.BatchSet(cacheData, 30*time.Minute)
}

// 转换方法
func (s *EnhancedProductService) convertESProductToModel(esProduct *elasticsearch.ProductDocument) *model.Product {
	// 实现ES文档到模型的转换
	return &model.Product{
		ID:          esProduct.ID,
		Name:        esProduct.Name,
		CategoryID:  esProduct.CategoryID,
		Price:       esProduct.Price,
		Stock:       esProduct.Stock,
		SalesCount:  esProduct.SalesCount,
		// ... 其他字段
	}
}

func (s *EnhancedProductService) convertModelToESProduct(product *model.Product) *elasticsearch.ProductDocument {
	// 实现模型到ES文档的转换
	return &elasticsearch.ProductDocument{
		ID:         product.ID,
		Name:       product.Name,
		CategoryID: product.CategoryID,
		Price:      product.Price,
		Stock:      product.Stock,
		SalesCount: product.SalesCount,
		// ... 其他字段
	}
}
