package repository

import (
	"context"
	"fmt"
	"time"

	"ryan-mall-microservices/internal/product/domain/entity"
	"ryan-mall-microservices/internal/product/domain/repository"
	"ryan-mall-microservices/internal/shared/domain"

	"gorm.io/gorm"
)

// ProductPO 商品持久化对象
type ProductPO struct {
	ID          uint      `gorm:"primaryKey;autoIncrement"`
	ProductID   string    `gorm:"uniqueIndex;size:36;not null"`
	Name        string    `gorm:"size:255;not null"`
	Description string    `gorm:"type:text"`
	CategoryID  string    `gorm:"size:36;not null;index"`
	Price       int64     `gorm:"not null"` // 以分为单位存储
	Stock       int       `gorm:"not null;default:0"`
	SalesCount  int       `gorm:"not null;default:0"`
	Status      int       `gorm:"not null;default:1"` // 1-可用，0-不可用
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

// TableName 表名
func (ProductPO) TableName() string {
	return "products"
}

// MySQLProductRepository MySQL商品仓储实现
type MySQLProductRepository struct {
	db *gorm.DB
}

// NewMySQLProductRepository 创建MySQL商品仓储
func NewMySQLProductRepository(db *gorm.DB) repository.ProductRepository {
	return &MySQLProductRepository{
		db: db,
	}
}

// Save 保存商品
func (r *MySQLProductRepository) Save(ctx context.Context, product *entity.Product) error {
	po := r.entityToPO(product)
	return r.db.WithContext(ctx).Create(&po).Error
}

// FindByID 根据ID查找商品
func (r *MySQLProductRepository) FindByID(ctx context.Context, id domain.ProductID) (*entity.Product, error) {
	var po ProductPO
	err := r.db.WithContext(ctx).Where("product_id = ?", id.String()).First(&po).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return r.poToEntity(&po)
}

// FindByCategoryID 根据分类ID查找商品列表
func (r *MySQLProductRepository) FindByCategoryID(ctx context.Context, categoryID string, offset, limit int) ([]*entity.Product, int64, error) {
	var pos []ProductPO
	var total int64

	// 查询总数
	if err := r.db.WithContext(ctx).Model(&ProductPO{}).Where("category_id = ?", categoryID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询数据
	if err := r.db.WithContext(ctx).Where("category_id = ?", categoryID).Offset(offset).Limit(limit).Find(&pos).Error; err != nil {
		return nil, 0, err
	}

	// 转换为实体
	products := make([]*entity.Product, len(pos))
	for i, po := range pos {
		product, err := r.poToEntity(&po)
		if err != nil {
			return nil, 0, err
		}
		products[i] = product
	}

	return products, total, nil
}

// Update 更新商品
func (r *MySQLProductRepository) Update(ctx context.Context, product *entity.Product) error {
	po := r.entityToPO(product)
	return r.db.WithContext(ctx).Model(&ProductPO{}).Where("product_id = ?", product.ID().String()).Updates(&po).Error
}

// Delete 删除商品
func (r *MySQLProductRepository) Delete(ctx context.Context, id domain.ProductID) error {
	return r.db.WithContext(ctx).Where("product_id = ?", id.String()).Delete(&ProductPO{}).Error
}

// List 分页查询商品列表
func (r *MySQLProductRepository) List(ctx context.Context, offset, limit int) ([]*entity.Product, int64, error) {
	var pos []ProductPO
	var total int64

	// 查询总数
	if err := r.db.WithContext(ctx).Model(&ProductPO{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询数据
	if err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&pos).Error; err != nil {
		return nil, 0, err
	}

	// 转换为实体
	products := make([]*entity.Product, len(pos))
	for i, po := range pos {
		product, err := r.poToEntity(&po)
		if err != nil {
			return nil, 0, err
		}
		products[i] = product
	}

	return products, total, nil
}

// Search 搜索商品
func (r *MySQLProductRepository) Search(ctx context.Context, keyword string, offset, limit int) ([]*entity.Product, int64, error) {
	var pos []ProductPO
	var total int64

	query := r.db.WithContext(ctx).Model(&ProductPO{})
	if keyword != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 查询总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询数据
	if err := query.Offset(offset).Limit(limit).Find(&pos).Error; err != nil {
		return nil, 0, err
	}

	// 转换为实体
	products := make([]*entity.Product, len(pos))
	for i, po := range pos {
		product, err := r.poToEntity(&po)
		if err != nil {
			return nil, 0, err
		}
		products[i] = product
	}

	return products, total, nil
}

// FindByIDs 根据ID列表查找商品
func (r *MySQLProductRepository) FindByIDs(ctx context.Context, ids []domain.ProductID) ([]*entity.Product, error) {
	if len(ids) == 0 {
		return []*entity.Product{}, nil
	}

	idStrings := make([]string, len(ids))
	for i, id := range ids {
		idStrings[i] = id.String()
	}

	var pos []ProductPO
	if err := r.db.WithContext(ctx).Where("product_id IN ?", idStrings).Find(&pos).Error; err != nil {
		return nil, err
	}

	// 转换为实体
	products := make([]*entity.Product, len(pos))
	for i, po := range pos {
		product, err := r.poToEntity(&po)
		if err != nil {
			return nil, err
		}
		products[i] = product
	}

	return products, nil
}

// UpdateStock 更新库存（原子操作）
func (r *MySQLProductRepository) UpdateStock(ctx context.Context, id domain.ProductID, quantity int) error {
	return r.db.WithContext(ctx).Model(&ProductPO{}).
		Where("product_id = ?", id.String()).
		Update("stock", quantity).Error
}

// ReserveStock 预留库存（原子操作）
func (r *MySQLProductRepository) ReserveStock(ctx context.Context, id domain.ProductID, quantity int) error {
	fmt.Printf("[DEBUG] 基础仓储的ReserveStock被调用: productID=%s, quantity=%d\n", id.String(), quantity)

	// 使用乐观锁进行库存扣减
	result := r.db.WithContext(ctx).Model(&ProductPO{}).
		Where("product_id = ? AND stock >= ?", id.String(), quantity).
		Update("stock", gorm.Expr("stock - ?", quantity))
	
	if result.Error != nil {
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		return domain.NewInsufficientStockError(id.String(), quantity, 0)
	}
	
	return nil
}

// ReleaseStock 释放库存（原子操作）
func (r *MySQLProductRepository) ReleaseStock(ctx context.Context, id domain.ProductID, quantity int) error {
	return r.db.WithContext(ctx).Model(&ProductPO{}).
		Where("product_id = ?", id.String()).
		Update("stock", gorm.Expr("stock + ?", quantity)).Error
}

// entityToPO 实体转持久化对象
func (r *MySQLProductRepository) entityToPO(product *entity.Product) ProductPO {
	status := 1
	if !product.IsAvailable() {
		status = 0
	}

	return ProductPO{
		ProductID:   product.ID().String(),
		Name:        product.Name(),
		Description: product.Description(),
		CategoryID:  product.CategoryID(),
		Price:       product.Price().Amount,
		Stock:       product.Stock(),
		SalesCount:  product.SalesCount(),
		Status:      status,
		CreatedAt:   product.CreatedAt().Time(),
		UpdatedAt:   product.UpdatedAt().Time(),
	}
}

// poToEntity 持久化对象转实体
func (r *MySQLProductRepository) poToEntity(po *ProductPO) (*entity.Product, error) {
	// 创建价格值对象
	price := domain.NewMoney(po.Price, "CNY")

	// 重建商品实体
	return entity.ReconstructProduct(
		domain.ProductID(po.ProductID),
		po.Name,
		po.Description,
		po.CategoryID,
		price,
		po.Stock,
		po.SalesCount,
		po.Status == 1,
		domain.NewTimestamp(po.CreatedAt),
		domain.NewTimestamp(po.UpdatedAt),
	), nil
}
