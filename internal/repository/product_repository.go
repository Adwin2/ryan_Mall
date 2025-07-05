package repository

import (
	"errors"
	"ryan-mall/internal/model"

	"gorm.io/gorm"
)

// ProductRepository 商品数据访问层接口
type ProductRepository interface {
	Create(product *model.Product) error                                    // 创建商品
	GetByID(id uint) (*model.Product, error)                              // 根据ID获取商品
	Update(product *model.Product) error                                   // 更新商品
	Delete(id uint) error                                                  // 删除商品（软删除）
	List(req *model.ProductListRequest) ([]*model.Product, int64, error)  // 分页查询商品列表
	GetByCategoryID(categoryID uint) ([]*model.Product, error)            // 根据分类ID获取商品
	UpdateStock(id uint, stock int) error                                 // 更新库存
	UpdateSalesCount(id uint, count int) error                            // 更新销售数量
}

// productRepository 商品数据访问层实现
type productRepository struct {
	db *gorm.DB
}

// NewProductRepository 创建商品数据访问层实例
func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{
		db: db,
	}
}

// Create 创建商品
func (r *productRepository) Create(product *model.Product) error {
	return r.db.Create(product).Error
}

// GetByID 根据ID获取商品
// 包含分类信息的关联查询
func (r *productRepository) GetByID(id uint) (*model.Product, error) {
	var product model.Product
	
	// 使用Preload预加载关联的分类信息
	err := r.db.Preload("Category").First(&product, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 商品不存在
		}
		return nil, err
	}
	
	return &product, nil
}

// Update 更新商品
func (r *productRepository) Update(product *model.Product) error {
	return r.db.Save(product).Error
}

// Delete 删除商品（软删除）
func (r *productRepository) Delete(id uint) error {
	return r.db.Delete(&model.Product{}, id).Error
}

// List 分页查询商品列表
// 支持关键词搜索、分类筛选、价格筛选、排序
func (r *productRepository) List(req *model.ProductListRequest) ([]*model.Product, int64, error) {
	var products []*model.Product
	var total int64
	
	// 构建查询条件
	query := r.db.Model(&model.Product{})
	
	// 1. 关键词搜索（商品名称）
	if req.Keyword != "" {
		query = query.Where("name LIKE ?", "%"+req.Keyword+"%")
	}
	
	// 2. 分类筛选
	if req.CategoryID != nil {
		query = query.Where("category_id = ?", *req.CategoryID)
	}
	
	// 3. 价格筛选
	if req.MinPrice != nil {
		query = query.Where("price >= ?", *req.MinPrice)
	}
	if req.MaxPrice != nil {
		query = query.Where("price <= ?", *req.MaxPrice)
	}
	
	// 4. 只查询上架的商品（状态为1）
	query = query.Where("status = ?", model.ProductStatusOnline)
	
	// 5. 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// 6. 排序
	orderBy := r.buildOrderBy(req.SortBy, req.SortOrder)
	query = query.Order(orderBy)
	
	// 7. 分页
	offset := (req.Page - 1) * req.PageSize
	query = query.Offset(offset).Limit(req.PageSize)
	
	// 8. 预加载分类信息并执行查询
	err := query.Preload("Category").Find(&products).Error
	if err != nil {
		return nil, 0, err
	}
	
	return products, total, nil
}

// buildOrderBy 构建排序条件
func (r *productRepository) buildOrderBy(sortBy, sortOrder string) string {
	// 验证排序字段
	validSortFields := map[string]bool{
		model.SortByCreatedAt:  true,
		model.SortByPrice:      true,
		model.SortBySalesCount: true,
	}
	
	if !validSortFields[sortBy] {
		sortBy = model.SortByCreatedAt // 默认按创建时间排序
	}
	
	// 验证排序方向
	if sortOrder != model.SortOrderAsc && sortOrder != model.SortOrderDesc {
		sortOrder = model.SortOrderDesc // 默认降序
	}
	
	return sortBy + " " + sortOrder
}

// GetByCategoryID 根据分类ID获取商品
func (r *productRepository) GetByCategoryID(categoryID uint) ([]*model.Product, error) {
	var products []*model.Product
	
	err := r.db.Where("category_id = ? AND status = ?", categoryID, model.ProductStatusOnline).
		Preload("Category").
		Find(&products).Error
	
	return products, err
}

// UpdateStock 更新库存
// 使用原子操作确保库存更新的安全性
func (r *productRepository) UpdateStock(id uint, stock int) error {
	return r.db.Model(&model.Product{}).
		Where("id = ?", id).
		Update("stock", stock).Error
}

// UpdateSalesCount 更新销售数量
func (r *productRepository) UpdateSalesCount(id uint, count int) error {
	// 使用原子操作增加销售数量
	return r.db.Model(&model.Product{}).
		Where("id = ?", id).
		Update("sales_count", gorm.Expr("sales_count + ?", count)).Error
}
