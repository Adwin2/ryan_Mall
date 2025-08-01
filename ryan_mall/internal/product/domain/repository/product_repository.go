package repository

import (
	"context"

	"ryan-mall-microservices/internal/product/domain/entity"
	"ryan-mall-microservices/internal/shared/domain"
)

// ProductRepository 商品仓储接口
type ProductRepository interface {
	// Save 保存商品
	Save(ctx context.Context, product *entity.Product) error
	
	// FindByID 根据ID查找商品
	FindByID(ctx context.Context, id domain.ProductID) (*entity.Product, error)
	
	// FindByCategoryID 根据分类ID查找商品列表
	FindByCategoryID(ctx context.Context, categoryID string, offset, limit int) ([]*entity.Product, int64, error)
	
	// Update 更新商品
	Update(ctx context.Context, product *entity.Product) error
	
	// Delete 删除商品
	Delete(ctx context.Context, id domain.ProductID) error
	
	// List 分页查询商品列表
	List(ctx context.Context, offset, limit int) ([]*entity.Product, int64, error)
	
	// Search 搜索商品
	Search(ctx context.Context, keyword string, offset, limit int) ([]*entity.Product, int64, error)
	
	// FindByIDs 根据ID列表查找商品
	FindByIDs(ctx context.Context, ids []domain.ProductID) ([]*entity.Product, error)
	
	// UpdateStock 更新库存（原子操作）
	UpdateStock(ctx context.Context, id domain.ProductID, quantity int) error
	
	// ReserveStock 预留库存（原子操作）
	ReserveStock(ctx context.Context, id domain.ProductID, quantity int) error
	
	// ReleaseStock 释放库存（原子操作）
	ReleaseStock(ctx context.Context, id domain.ProductID, quantity int) error
}

// CategoryRepository 分类仓储接口
type CategoryRepository interface {
	// Save 保存分类
	Save(ctx context.Context, category *entity.Category) error
	
	// FindByID 根据ID查找分类
	FindByID(ctx context.Context, id domain.ID) (*entity.Category, error)
	
	// FindByParentID 根据父分类ID查找子分类列表
	FindByParentID(ctx context.Context, parentID *domain.ID) ([]*entity.Category, error)
	
	// FindRootCategories 查找根分类列表
	FindRootCategories(ctx context.Context) ([]*entity.Category, error)
	
	// Update 更新分类
	Update(ctx context.Context, category *entity.Category) error
	
	// Delete 删除分类
	Delete(ctx context.Context, id domain.ID) error
	
	// List 分页查询分类列表
	List(ctx context.Context, offset, limit int) ([]*entity.Category, int64, error)
	
	// ExistsByName 检查分类名称是否存在
	ExistsByName(ctx context.Context, name string, parentID *domain.ID) (bool, error)
}
