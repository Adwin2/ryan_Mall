package service

import (
	"context"

	"ryan-mall-microservices/internal/product/domain/entity"
	"ryan-mall-microservices/internal/product/domain/repository"
	"ryan-mall-microservices/internal/shared/domain"
)

// ProductDomainService 商品领域服务
type ProductDomainService struct {
	productRepo  repository.ProductRepository
	categoryRepo repository.CategoryRepository
}

// NewProductDomainService 创建商品领域服务
func NewProductDomainService(
	productRepo repository.ProductRepository,
	categoryRepo repository.CategoryRepository,
) *ProductDomainService {
	return &ProductDomainService{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
	}
}

// ValidateProductForCreation 验证商品创建
func (s *ProductDomainService) ValidateProductForCreation(ctx context.Context, categoryID string) error {
	// 检查分类是否存在
	category, err := s.categoryRepo.FindByID(ctx, domain.ID(categoryID))
	if err != nil {
		return domain.NewInternalError("failed to find category", err)
	}
	if category == nil {
		return domain.NewNotFoundError("category", categoryID)
	}

	// 检查分类是否激活
	if !category.IsActive() {
		return domain.NewBusinessError(domain.ErrCodeValidation, "category is not active")
	}

	return nil
}

// CanProductBeDeleted 检查商品是否可以删除
func (s *ProductDomainService) CanProductBeDeleted(ctx context.Context, product *entity.Product) error {
	if product == nil {
		return domain.NewNotFoundError("product", "")
	}

	// 如果商品有库存，不能删除
	if product.Stock() > 0 {
		return domain.NewBusinessError(domain.ErrCodeConflict, "cannot delete product with stock")
	}

	return nil
}

// ValidateCategoryForCreation 验证分类创建
func (s *ProductDomainService) ValidateCategoryForCreation(ctx context.Context, name string, parentID *domain.ID) error {
	// 检查分类名称是否重复
	exists, err := s.categoryRepo.ExistsByName(ctx, name, parentID)
	if err != nil {
		return domain.NewInternalError("failed to check category name", err)
	}
	if exists {
		return domain.NewAlreadyExistsError("category", "name", name)
	}

	// 如果有父分类，检查父分类是否存在且激活
	if parentID != nil {
		parent, err := s.categoryRepo.FindByID(ctx, *parentID)
		if err != nil {
			return domain.NewInternalError("failed to find parent category", err)
		}
		if parent == nil {
			return domain.NewNotFoundError("parent category", parentID.String())
		}
		if !parent.IsActive() {
			return domain.NewBusinessError(domain.ErrCodeValidation, "parent category is not active")
		}
	}

	return nil
}

// CanCategoryBeDeleted 检查分类是否可以删除
func (s *ProductDomainService) CanCategoryBeDeleted(ctx context.Context, categoryID domain.ID) error {
	// 检查是否有子分类
	children, err := s.categoryRepo.FindByParentID(ctx, &categoryID)
	if err != nil {
		return domain.NewInternalError("failed to find child categories", err)
	}
	if len(children) > 0 {
		return domain.NewBusinessError(domain.ErrCodeConflict, "cannot delete category with child categories")
	}

	// 检查是否有商品使用此分类
	products, _, err := s.productRepo.FindByCategoryID(ctx, categoryID.String(), 0, 1)
	if err != nil {
		return domain.NewInternalError("failed to find products in category", err)
	}
	if len(products) > 0 {
		return domain.NewBusinessError(domain.ErrCodeConflict, "cannot delete category with products")
	}

	return nil
}

// CalculateDiscountPrice 计算折扣价格
func (s *ProductDomainService) CalculateDiscountPrice(originalPrice domain.Money, discountPercent float64) domain.Money {
	if discountPercent <= 0 || discountPercent >= 100 {
		return originalPrice
	}

	discountAmount := int64(float64(originalPrice.Amount) * discountPercent / 100)
	finalAmount := originalPrice.Amount - discountAmount

	if finalAmount < 0 {
		finalAmount = 0
	}

	return domain.NewMoney(finalAmount, originalPrice.Currency)
}

// ValidateStockOperation 验证库存操作
func (s *ProductDomainService) ValidateStockOperation(ctx context.Context, productID domain.ProductID, quantity int) error {
	if quantity <= 0 {
		return domain.NewValidationError("quantity must be positive")
	}

	product, err := s.productRepo.FindByID(ctx, productID)
	if err != nil {
		return domain.NewInternalError("failed to find product", err)
	}
	if product == nil {
		return domain.NewNotFoundError("product", productID.String())
	}

	if !product.IsAvailable() {
		return domain.NewBusinessError(domain.ErrCodeValidation, "product is not available")
	}

	return nil
}

// BatchValidateProducts 批量验证商品
func (s *ProductDomainService) BatchValidateProducts(ctx context.Context, productIDs []domain.ProductID) ([]*entity.Product, error) {
	products, err := s.productRepo.FindByIDs(ctx, productIDs)
	if err != nil {
		return nil, domain.NewInternalError("failed to find products", err)
	}

	if len(products) != len(productIDs) {
		return nil, domain.NewBusinessError(domain.ErrCodeNotFound, "some products not found")
	}

	// 验证所有商品都可用
	for _, product := range products {
		if !product.IsAvailable() {
			return nil, domain.NewBusinessError(domain.ErrCodeValidation, 
				"product "+product.ID().String()+" is not available")
		}
	}

	return products, nil
}
