package service

import (
	"errors"
	"ryan-mall/internal/model"
	"ryan-mall/internal/repository"
)

// CategoryService 分类业务逻辑层接口
type CategoryService interface {
	CreateCategory(req *model.CategoryCreateRequest) (*model.Category, error)    // 创建分类
	GetCategory(id uint) (*model.Category, error)                               // 获取分类详情
	UpdateCategory(id uint, req *model.CategoryUpdateRequest) error             // 更新分类
	DeleteCategory(id uint) error                                               // 删除分类
	GetAllCategories() ([]*model.Category, error)                               // 获取所有分类
	GetCategoryTree() ([]*model.Category, error)                                // 获取分类树
	GetTopCategories() ([]*model.Category, error)                               // 获取顶级分类
	GetSubCategories(parentID uint) ([]*model.Category, error)                  // 获取子分类
}

// categoryService 分类业务逻辑层实现
type categoryService struct {
	categoryRepo repository.CategoryRepository
}

// NewCategoryService 创建分类业务逻辑层实例
func NewCategoryService(categoryRepo repository.CategoryRepository) CategoryService {
	return &categoryService{
		categoryRepo: categoryRepo,
	}
}

// CreateCategory 创建分类
func (s *categoryService) CreateCategory(req *model.CategoryCreateRequest) (*model.Category, error) {
	// 1. 验证分类名称是否已存在
	exists, err := s.categoryRepo.ExistsByName(req.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("分类名称已存在")
	}
	
	// 2. 如果是子分类，验证父分类是否存在
	if req.ParentID != 0 {
		parent, err := s.categoryRepo.GetByID(req.ParentID)
		if err != nil {
			return nil, err
		}
		if parent == nil {
			return nil, errors.New("父分类不存在")
		}
	}
	
	// 3. 创建分类对象
	category := &model.Category{
		Name:      req.Name,
		ParentID:  req.ParentID,
		SortOrder: req.SortOrder,
		Status:    1, // 默认启用
	}
	
	// 4. 保存分类
	if err := s.categoryRepo.Create(category); err != nil {
		return nil, err
	}
	
	return category, nil
}

// GetCategory 获取分类详情
func (s *categoryService) GetCategory(id uint) (*model.Category, error) {
	category, err := s.categoryRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, errors.New("分类不存在")
	}
	
	return category, nil
}

// UpdateCategory 更新分类
func (s *categoryService) UpdateCategory(id uint, req *model.CategoryUpdateRequest) error {
	// 1. 获取现有分类
	category, err := s.categoryRepo.GetByID(id)
	if err != nil {
		return err
	}
	if category == nil {
		return errors.New("分类不存在")
	}
	
	// 2. 验证分类名称（如果要更新名称）
	if req.Name != nil && *req.Name != category.Name {
		exists, err := s.categoryRepo.ExistsByName(*req.Name)
		if err != nil {
			return err
		}
		if exists {
			return errors.New("分类名称已存在")
		}
		category.Name = *req.Name
	}
	
	// 3. 验证父分类（如果要更新父分类）
	if req.ParentID != nil {
		// 不能将分类设置为自己的子分类
		if *req.ParentID == id {
			return errors.New("不能将分类设置为自己的子分类")
		}
		
		// 验证父分类是否存在（除非设置为顶级分类）
		if *req.ParentID != 0 {
			parent, err := s.categoryRepo.GetByID(*req.ParentID)
			if err != nil {
				return err
			}
			if parent == nil {
				return errors.New("父分类不存在")
			}
			
			// 检查是否会形成循环引用
			if err := s.checkCircularReference(id, *req.ParentID); err != nil {
				return err
			}
		}
		
		category.ParentID = *req.ParentID
	}
	
	// 4. 更新其他字段
	if req.SortOrder != nil {
		category.SortOrder = *req.SortOrder
	}
	if req.Status != nil {
		category.Status = *req.Status
	}
	
	// 5. 保存更新
	return s.categoryRepo.Update(category)
}

// checkCircularReference 检查是否会形成循环引用
func (s *categoryService) checkCircularReference(categoryID, newParentID uint) error {
	// 递归检查新父分类的父分类链，确保不会形成循环
	currentParentID := newParentID
	
	for currentParentID != 0 {
		if currentParentID == categoryID {
			return errors.New("不能形成循环引用")
		}
		
		parent, err := s.categoryRepo.GetByID(currentParentID)
		if err != nil {
			return err
		}
		if parent == nil {
			break
		}
		
		currentParentID = parent.ParentID
	}
	
	return nil
}

// DeleteCategory 删除分类
func (s *categoryService) DeleteCategory(id uint) error {
	// 1. 检查分类是否存在
	category, err := s.categoryRepo.GetByID(id)
	if err != nil {
		return err
	}
	if category == nil {
		return errors.New("分类不存在")
	}
	
	// 2. 检查是否有子分类
	hasChildren, err := s.categoryRepo.HasChildren(id)
	if err != nil {
		return err
	}
	if hasChildren {
		return errors.New("该分类下还有子分类，无法删除")
	}
	
	// 3. 检查是否有关联的商品
	hasProducts, err := s.categoryRepo.HasProducts(id)
	if err != nil {
		return err
	}
	if hasProducts {
		return errors.New("该分类下还有商品，无法删除")
	}
	
	// 4. 执行删除
	return s.categoryRepo.Delete(id)
}

// GetAllCategories 获取所有分类
func (s *categoryService) GetAllCategories() ([]*model.Category, error) {
	return s.categoryRepo.GetAll()
}

// GetCategoryTree 获取分类树
func (s *categoryService) GetCategoryTree() ([]*model.Category, error) {
	return s.categoryRepo.GetCategoryTree()
}

// GetTopCategories 获取顶级分类
func (s *categoryService) GetTopCategories() ([]*model.Category, error) {
	return s.categoryRepo.GetTopLevel()
}

// GetSubCategories 获取子分类
func (s *categoryService) GetSubCategories(parentID uint) ([]*model.Category, error) {
	// 1. 验证父分类是否存在
	if parentID != 0 {
		parent, err := s.categoryRepo.GetByID(parentID)
		if err != nil {
			return nil, err
		}
		if parent == nil {
			return nil, errors.New("父分类不存在")
		}
	}
	
	// 2. 获取子分类
	return s.categoryRepo.GetByParentID(parentID)
}
