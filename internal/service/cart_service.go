package service

import (
	"errors"
	"ryan-mall/internal/model"
	"ryan-mall/internal/repository"
)

// CartService 购物车业务逻辑层接口
type CartService interface {
	AddToCart(userID uint, req *model.AddToCartRequest) error                    // 添加商品到购物车
	GetCart(userID uint) (*model.CartListResponse, error)                       // 获取用户购物车
	UpdateCartItem(userID, cartItemID uint, req *model.UpdateCartRequest) error // 更新购物车商品数量
	RemoveFromCart(userID, cartItemID uint) error                               // 从购物车移除商品
	RemoveProduct(userID, productID uint) error                                 // 移除特定商品
	ClearCart(userID uint) error                                                // 清空购物车
	GetCartSummary(userID uint) (*model.CartSummary, error)                     // 获取购物车汇总
	ValidateCartItems(userID uint, cartItemIDs []uint) ([]*model.CartItem, error) // 验证购物车项（用于下单）
}

// cartService 购物车业务逻辑层实现
type cartService struct {
	cartRepo    repository.CartRepository
	productRepo repository.ProductRepository
}

// NewCartService 创建购物车业务逻辑层实例
func NewCartService(cartRepo repository.CartRepository, productRepo repository.ProductRepository) CartService {
	return &cartService{
		cartRepo:    cartRepo,
		productRepo: productRepo,
	}
}

// AddToCart 添加商品到购物车
func (s *cartService) AddToCart(userID uint, req *model.AddToCartRequest) error {
	// 1. 验证商品是否存在且上架
	product, err := s.productRepo.GetByID(req.ProductID)
	if err != nil {
		return err
	}
	if product == nil {
		return errors.New("商品不存在")
	}
	if product.Status != model.ProductStatusOnline {
		return errors.New("商品已下架")
	}
	
	// 2. 检查库存是否充足
	if product.Stock < req.Quantity {
		return errors.New("库存不足")
	}
	
	// 3. 检查购物车中是否已存在该商品
	existingItem, err := s.cartRepo.GetByUserAndProduct(userID, req.ProductID)
	if err != nil {
		return err
	}
	
	if existingItem != nil {
		// 商品已存在，更新数量
		newQuantity := existingItem.Quantity + req.Quantity
		
		// 检查新数量是否超过库存
		if newQuantity > product.Stock {
			return errors.New("添加数量超过库存限制")
		}
		
		// 更新数量
		existingItem.Quantity = newQuantity
		return s.cartRepo.Update(existingItem)
	} else {
		// 商品不存在，创建新的购物车项
		cartItem := &model.CartItem{
			UserID:    userID,
			ProductID: req.ProductID,
			Quantity:  req.Quantity,
		}
		return s.cartRepo.Create(cartItem)
	}
}

// GetCart 获取用户购物车
func (s *cartService) GetCart(userID uint) (*model.CartListResponse, error) {
	// 1. 获取购物车项（包含商品验证）
	cartItems, err := s.cartRepo.GetCartItemsWithValidation(userID)
	if err != nil {
		return nil, err
	}
	
	// 2. 转换为响应格式
	var cartResponses []*model.CartResponse
	var totalPrice float64
	
	for _, item := range cartItems {
		// 检查商品库存，如果库存不足则调整数量
		if item.Quantity > item.Product.Stock {
			item.Quantity = item.Product.Stock
			if item.Quantity > 0 {
				s.cartRepo.Update(item) // 更新数量
			} else {
				s.cartRepo.Delete(item.ID) // 库存为0则删除
				continue
			}
		}
		
		itemTotalPrice := float64(item.Quantity) * item.Product.Price
		totalPrice += itemTotalPrice
		
		cartResponse := &model.CartResponse{
			ID:          item.ID,
			ProductID:   item.ProductID,
			ProductName: item.Product.Name,
			Price:       item.Product.Price,
			Quantity:    item.Quantity,
			TotalPrice:  itemTotalPrice,
			Stock:       item.Product.Stock,
			CreatedAt:   item.CreatedAt,
		}
		
		// 设置商品图片
		if item.Product.MainImage != nil {
			cartResponse.ProductImage = item.Product.MainImage
		}
		
		cartResponses = append(cartResponses, cartResponse)
	}
	
	// 3. 构建响应
	response := &model.CartListResponse{
		Items:      cartResponses,
		TotalCount: len(cartResponses),
		TotalPrice: totalPrice,
	}
	
	return response, nil
}

// UpdateCartItem 更新购物车商品数量
func (s *cartService) UpdateCartItem(userID, cartItemID uint, req *model.UpdateCartRequest) error {
	// 1. 获取购物车项
	cartItems, err := s.cartRepo.GetByUserID(userID)
	if err != nil {
		return err
	}
	
	var targetItem *model.CartItem
	for _, item := range cartItems {
		if item.ID == cartItemID {
			targetItem = item
			break
		}
	}
	
	if targetItem == nil {
		return errors.New("购物车项不存在")
	}
	
	// 2. 验证商品状态
	if targetItem.Product.Status != model.ProductStatusOnline {
		return errors.New("商品已下架")
	}
	
	// 3. 检查库存
	if req.Quantity > targetItem.Product.Stock {
		return errors.New("数量超过库存限制")
	}
	
	// 4. 更新数量
	targetItem.Quantity = req.Quantity
	return s.cartRepo.Update(targetItem)
}

// RemoveFromCart 从购物车移除商品
func (s *cartService) RemoveFromCart(userID, cartItemID uint) error {
	// 1. 验证购物车项是否属于该用户
	cartItems, err := s.cartRepo.GetByUserID(userID)
	if err != nil {
		return err
	}
	
	var found bool
	for _, item := range cartItems {
		if item.ID == cartItemID {
			found = true
			break
		}
	}
	
	if !found {
		return errors.New("购物车项不存在")
	}
	
	// 2. 删除购物车项
	return s.cartRepo.Delete(cartItemID)
}

// RemoveProduct 移除特定商品
func (s *cartService) RemoveProduct(userID, productID uint) error {
	return s.cartRepo.DeleteByUserAndProduct(userID, productID)
}

// ClearCart 清空购物车
func (s *cartService) ClearCart(userID uint) error {
	return s.cartRepo.DeleteByUser(userID)
}

// GetCartSummary 获取购物车汇总
func (s *cartService) GetCartSummary(userID uint) (*model.CartSummary, error) {
	return s.cartRepo.GetCartSummary(userID)
}

// ValidateCartItems 验证购物车项（用于下单）
func (s *cartService) ValidateCartItems(userID uint, cartItemIDs []uint) ([]*model.CartItem, error) {
	// 1. 获取指定的购物车项
	cartItems, err := s.cartRepo.GetByIDs(cartItemIDs)
	if err != nil {
		return nil, err
	}
	
	// 2. 验证购物车项是否属于该用户
	var validItems []*model.CartItem
	for _, item := range cartItems {
		if item.UserID != userID {
			return nil, errors.New("购物车项不属于当前用户")
		}
		
		// 3. 验证商品状态
		if item.Product.Status != model.ProductStatusOnline {
			return nil, errors.New("商品 " + item.Product.Name + " 已下架")
		}
		
		// 4. 验证库存
		if item.Quantity > item.Product.Stock {
			return nil, errors.New("商品 " + item.Product.Name + " 库存不足")
		}
		
		validItems = append(validItems, item)
	}
	
	if len(validItems) == 0 {
		return nil, errors.New("没有有效的购物车项")
	}
	
	return validItems, nil
}

// BatchAddToCart 批量添加商品到购物车
func (s *cartService) BatchAddToCart(userID uint, requests []*model.AddToCartRequest) error {
	for _, req := range requests {
		if err := s.AddToCart(userID, req); err != nil {
			return err
		}
	}
	return nil
}

// GetCartItemCount 获取购物车商品数量
func (s *cartService) GetCartItemCount(userID uint) (int, error) {
	summary, err := s.cartRepo.GetCartSummary(userID)
	if err != nil {
		return 0, err
	}
	return summary.TotalItems, nil
}
