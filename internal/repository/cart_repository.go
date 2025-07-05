package repository

import (
	"errors"
	"ryan-mall/internal/model"

	"gorm.io/gorm"
)

// CartRepository 购物车数据访问层接口
type CartRepository interface {
	Create(cartItem *model.CartItem) error                           // 添加商品到购物车
	GetByUserID(userID uint) ([]*model.CartItem, error)            // 获取用户购物车
	GetByUserAndProduct(userID, productID uint) (*model.CartItem, error) // 获取用户特定商品的购物车项
	Update(cartItem *model.CartItem) error                          // 更新购物车项
	Delete(id uint) error                                           // 删除购物车项
	DeleteByUserAndProduct(userID, productID uint) error           // 删除用户特定商品
	DeleteByUser(userID uint) error                                 // 清空用户购物车
	GetByIDs(ids []uint) ([]*model.CartItem, error)               // 根据ID列表获取购物车项
	GetCartSummary(userID uint) (*model.CartSummary, error)       // 获取购物车汇总信息
	GetCartItemsWithValidation(userID uint) ([]*model.CartItem, error) // 获取购物车项并验证商品状态
}

// cartRepository 购物车数据访问层实现
type cartRepository struct {
	db *gorm.DB
}

// NewCartRepository 创建购物车数据访问层实例
func NewCartRepository(db *gorm.DB) CartRepository {
	return &cartRepository{
		db: db,
	}
}

// Create 添加商品到购物车
func (r *cartRepository) Create(cartItem *model.CartItem) error {
	return r.db.Create(cartItem).Error
}

// GetByUserID 获取用户购物车
// 包含商品信息和分类信息的关联查询
func (r *cartRepository) GetByUserID(userID uint) ([]*model.CartItem, error) {
	var cartItems []*model.CartItem
	
	// 预加载商品信息和商品的分类信息
	// 只查询上架的商品
	err := r.db.Where("user_id = ?", userID).
		Preload("Product", "status = ?", model.ProductStatusOnline).
		Preload("Product.Category").
		Order("created_at DESC").
		Find(&cartItems).Error
	
	return cartItems, err
}

// GetByUserAndProduct 获取用户特定商品的购物车项
func (r *cartRepository) GetByUserAndProduct(userID, productID uint) (*model.CartItem, error) {
	var cartItem model.CartItem
	
	err := r.db.Where("user_id = ? AND product_id = ?", userID, productID).
		First(&cartItem).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 购物车项不存在
		}
		return nil, err
	}
	
	return &cartItem, nil
}

// Update 更新购物车项
func (r *cartRepository) Update(cartItem *model.CartItem) error {
	return r.db.Save(cartItem).Error
}

// Delete 删除购物车项
func (r *cartRepository) Delete(id uint) error {
	return r.db.Delete(&model.CartItem{}, id).Error
}

// DeleteByUserAndProduct 删除用户特定商品
func (r *cartRepository) DeleteByUserAndProduct(userID, productID uint) error {
	return r.db.Where("user_id = ? AND product_id = ?", userID, productID).
		Delete(&model.CartItem{}).Error
}

// DeleteByUser 清空用户购物车
func (r *cartRepository) DeleteByUser(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&model.CartItem{}).Error
}

// GetByIDs 根据ID列表获取购物车项
// 用于批量操作，如批量结算
func (r *cartRepository) GetByIDs(ids []uint) ([]*model.CartItem, error) {
	var cartItems []*model.CartItem
	
	err := r.db.Where("id IN ?", ids).
		Preload("Product", "status = ?", model.ProductStatusOnline).
		Preload("Product.Category").
		Find(&cartItems).Error
	
	return cartItems, err
}

// GetCartSummary 获取购物车汇总信息
// 计算购物车中的商品总数和总金额
func (r *cartRepository) GetCartSummary(userID uint) (*model.CartSummary, error) {
	var summary model.CartSummary
	
	// 查询购物车项，只包含上架的商品
	var cartItems []*model.CartItem
	err := r.db.Where("user_id = ?", userID).
		Preload("Product", "status = ?", model.ProductStatusOnline).
		Find(&cartItems).Error
	
	if err != nil {
		return nil, err
	}
	
	// 计算汇总信息
	summary.UserID = userID
	summary.TotalItems = 0
	summary.TotalAmount = 0
	summary.ItemCount = len(cartItems)
	
	for _, item := range cartItems {
		if item.Product.ID != 0 { // 确保商品存在且上架
			summary.TotalItems += item.Quantity
			summary.TotalAmount += float64(item.Quantity) * item.Product.Price
		}
	}
	
	return &summary, nil
}

// GetCartItemsWithValidation 获取购物车项并验证商品状态
// 这个方法会过滤掉已下架或删除的商品
func (r *cartRepository) GetCartItemsWithValidation(userID uint) ([]*model.CartItem, error) {
	var cartItems []*model.CartItem
	
	// 查询购物车项
	err := r.db.Where("user_id = ?", userID).
		Preload("Product").
		Preload("Product.Category").
		Find(&cartItems).Error
	
	if err != nil {
		return nil, err
	}
	
	// 过滤有效的购物车项
	var validItems []*model.CartItem
	var invalidItemIDs []uint
	
	for _, item := range cartItems {
		// 检查商品是否存在且上架
		if item.Product.ID == 0 || item.Product.Status != model.ProductStatusOnline {
			invalidItemIDs = append(invalidItemIDs, item.ID)
		} else {
			validItems = append(validItems, item)
		}
	}
	
	// 删除无效的购物车项
	if len(invalidItemIDs) > 0 {
		r.db.Where("id IN ?", invalidItemIDs).Delete(&model.CartItem{})
	}
	
	return validItems, nil
}

// UpdateQuantity 更新购物车商品数量
// 使用原子操作确保数据一致性
func (r *cartRepository) UpdateQuantity(userID, productID uint, quantity int) error {
	return r.db.Model(&model.CartItem{}).
		Where("user_id = ? AND product_id = ?", userID, productID).
		Update("quantity", quantity).Error
}

// IncrementQuantity 增加购物车商品数量
func (r *cartRepository) IncrementQuantity(userID, productID uint, increment int) error {
	return r.db.Model(&model.CartItem{}).
		Where("user_id = ? AND product_id = ?", userID, productID).
		Update("quantity", gorm.Expr("quantity + ?", increment)).Error
}
