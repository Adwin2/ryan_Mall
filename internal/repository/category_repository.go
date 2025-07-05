package repository

import (
	"errors"
	"ryan-mall/internal/model"

	"gorm.io/gorm"
)

// CategoryRepository 分类数据访问层接口
type CategoryRepository interface {
	Create(category *model.Category) error                    // 创建分类
	GetByID(id uint) (*model.Category, error)               // 根据ID获取分类
	Update(category *model.Category) error                   // 更新分类
	Delete(id uint) error                                    // 删除分类（软删除）
	GetAll() ([]*model.Category, error)                     // 获取所有分类
	GetByParentID(parentID uint) ([]*model.Category, error) // 根据父分类ID获取子分类
	GetTopLevel() ([]*model.Category, error)                // 获取顶级分类
	GetWithChildren(id uint) (*model.Category, error)       // 获取分类及其子分类
	ExistsByName(name string) (bool, error)                 // 检查分类名称是否存在
	HasProducts(id uint) (bool, error)                      // 检查分类下是否有商品
	HasChildren(id uint) (bool, error)                      // 检查分类下是否有子分类
	GetCategoryTree() ([]*model.Category, error)            // 获取完整的分类树结构
}

// categoryRepository 分类数据访问层实现
type categoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository 创建分类数据访问层实例
func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{
		db: db,
	}
}

// Create 创建分类
func (r *categoryRepository) Create(category *model.Category) error {
	return r.db.Create(category).Error
}

// GetByID 根据ID获取分类
func (r *categoryRepository) GetByID(id uint) (*model.Category, error) {
	var category model.Category
	
	err := r.db.First(&category, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 分类不存在
		}
		return nil, err
	}
	
	return &category, nil
}

// Update 更新分类
func (r *categoryRepository) Update(category *model.Category) error {
	return r.db.Save(category).Error
}

// Delete 删除分类（软删除）
func (r *categoryRepository) Delete(id uint) error {
	return r.db.Delete(&model.Category{}, id).Error
}

// GetAll 获取所有分类
// 按照父分类ID和排序权重排序
func (r *categoryRepository) GetAll() ([]*model.Category, error) {
	var categories []*model.Category
	
	err := r.db.Where("status = ?", 1).
		Order("parent_id ASC, sort_order ASC, created_at ASC").
		Find(&categories).Error
	
	return categories, err
}

// GetByParentID 根据父分类ID获取子分类
func (r *categoryRepository) GetByParentID(parentID uint) ([]*model.Category, error) {
	var categories []*model.Category
	
	err := r.db.Where("parent_id = ? AND status = ?", parentID, 1).
		Order("sort_order ASC, created_at ASC").
		Find(&categories).Error
	
	return categories, err
}

// GetTopLevel 获取顶级分类
// 父分类ID为0的分类
func (r *categoryRepository) GetTopLevel() ([]*model.Category, error) {
	return r.GetByParentID(0)
}

// GetWithChildren 获取分类及其子分类
// 使用递归查询获取完整的分类树
func (r *categoryRepository) GetWithChildren(id uint) (*model.Category, error) {
	// 1. 获取当前分类
	category, err := r.GetByID(id)
	if err != nil || category == nil {
		return nil, err
	}
	
	// 2. 获取子分类
	children, err := r.GetByParentID(id)
	if err != nil {
		return nil, err
	}
	
	// 3. 递归获取每个子分类的子分类
	for _, child := range children {
		grandChildren, err := r.GetByParentID(child.ID)
		if err != nil {
			return nil, err
		}
		child.Children = grandChildren
	}
	
	category.Children = children
	return category, nil
}

// ExistsByName 检查分类名称是否存在
// 在同一父分类下，分类名称应该唯一
func (r *categoryRepository) ExistsByName(name string) (bool, error) {
	var count int64
	
	err := r.db.Model(&model.Category{}).
		Where("name = ?", name).
		Count(&count).Error
	
	return count > 0, err
}

// HasProducts 检查分类下是否有商品
// 删除分类前需要检查是否有关联的商品
func (r *categoryRepository) HasProducts(id uint) (bool, error) {
	var count int64
	
	err := r.db.Model(&model.Product{}).
		Where("category_id = ?", id).
		Count(&count).Error
	
	return count > 0, err
}

// HasChildren 检查分类下是否有子分类
// 删除分类前需要检查是否有子分类
func (r *categoryRepository) HasChildren(id uint) (bool, error) {
	var count int64
	
	err := r.db.Model(&model.Category{}).
		Where("parent_id = ?", id).
		Count(&count).Error
	
	return count > 0, err
}

// GetCategoryTree 获取完整的分类树结构
// 这是一个辅助方法，用于构建层级分类结构
func (r *categoryRepository) GetCategoryTree() ([]*model.Category, error) {
	// 1. 获取所有分类
	allCategories, err := r.GetAll()
	if err != nil {
		return nil, err
	}
	
	// 2. 构建分类映射
	categoryMap := make(map[uint]*model.Category)
	var topCategories []*model.Category
	
	// 初始化映射
	for _, category := range allCategories {
		categoryMap[category.ID] = category
		category.Children = []*model.Category{} // 初始化子分类切片
	}
	
	// 3. 构建树结构
	for _, category := range allCategories {
		if category.ParentID == 0 {
			// 顶级分类
			topCategories = append(topCategories, category)
		} else {
			// 子分类，添加到父分类的Children中
			if parent, exists := categoryMap[category.ParentID]; exists {
				parent.Children = append(parent.Children, category)
			}
		}
	}
	
	return topCategories, nil
}
