package service

import (
	"errors"
	"fmt"
	"ryan-mall/internal/model"
	"ryan-mall/internal/repository"
	"time"

	"gorm.io/gorm"
)

// OrderService 订单业务逻辑层接口
type OrderService interface {
	CreateOrder(userID uint, req *model.CreateOrderRequest) (*model.Order, error)     // 创建订单
	GetOrder(userID, orderID uint) (*model.Order, error)                             // 获取订单详情
	GetOrderByNo(userID uint, orderNo string) (*model.Order, error)                  // 根据订单号获取订单
	GetOrderList(userID uint, req *model.OrderListRequest) (*model.OrderListResponse, error) // 获取订单列表
	CancelOrder(userID, orderID uint) error                                          // 取消订单
	PayOrder(userID, orderID uint, req *model.PayOrderRequest) error                 // 支付订单
	ConfirmOrder(userID, orderID uint) error                                         // 确认收货
	GetOrderStatistics(userID uint) (*model.OrderStatistics, error)                 // 获取订单统计
	ProcessExpiredOrders() error                                                     // 处理过期订单
}

// orderService 订单业务逻辑层实现
type orderService struct {
	orderRepo   repository.OrderRepository
	cartRepo    repository.CartRepository
	productRepo repository.ProductRepository
	db          *gorm.DB
}

// NewOrderService 创建订单业务逻辑层实例
func NewOrderService(orderRepo repository.OrderRepository, cartRepo repository.CartRepository, productRepo repository.ProductRepository, db *gorm.DB) OrderService {
	return &orderService{
		orderRepo:   orderRepo,
		cartRepo:    cartRepo,
		productRepo: productRepo,
		db:          db,
	}
}

// CreateOrder 创建订单
// 从购物车创建订单，包含库存扣减和购物车清理
func (s *orderService) CreateOrder(userID uint, req *model.CreateOrderRequest) (*model.Order, error) {
	// 1. 验证购物车项
	cartItems, err := s.cartRepo.GetByIDs(req.CartItemIDs)
	if err != nil {
		return nil, err
	}

	if len(cartItems) == 0 {
		return nil, errors.New("购物车为空")
	}

	// 验证购物车项是否属于当前用户
	for _, item := range cartItems {
		if item.UserID != userID {
			return nil, errors.New("购物车项不属于当前用户")
		}

		// 验证商品状态
		if item.Product.Status != model.ProductStatusOnline {
			return nil, errors.New("商品 " + item.Product.Name + " 已下架")
		}

		// 验证库存
		if item.Quantity > item.Product.Stock {
			return nil, errors.New("商品 " + item.Product.Name + " 库存不足")
		}
	}
	
	// 2. 使用事务处理订单创建
	var order *model.Order
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 3. 计算订单金额和创建订单项
		var totalAmount float64
		var orderItems []*model.OrderItem
		
		for _, cartItem := range cartItems {
			// 再次检查库存（防止并发问题）
			var currentStock int
			err := tx.Model(&model.Product{}).
				Select("stock").
				Where("id = ?", cartItem.ProductID).
				Row().Scan(&currentStock)
			if err != nil {
				return err
			}
			
			if currentStock < cartItem.Quantity {
				return fmt.Errorf("商品 %s 库存不足，当前库存：%d", cartItem.Product.Name, currentStock)
			}
			
			// 扣减库存
			err = tx.Model(&model.Product{}).
				Where("id = ?", cartItem.ProductID).
				Update("stock", gorm.Expr("stock - ?", cartItem.Quantity)).Error
			if err != nil {
				return err
			}
			
			// 创建订单项
			itemTotal := cartItem.Product.Price * float64(cartItem.Quantity)
			totalAmount += itemTotal
			
			orderItem := &model.OrderItem{
				ProductID:    cartItem.ProductID,
				ProductName:  cartItem.Product.Name,
				ProductImage: cartItem.Product.MainImage,
				Price:        cartItem.Product.Price,
				Quantity:     cartItem.Quantity,
				TotalPrice:   itemTotal,
			}
			orderItems = append(orderItems, orderItem)
		}
		
		// 4. 生成订单号
		orderNo := s.generateOrderNo()
		
		// 5. 创建订单
		var remark *string
		if req.Remark != "" {
			remark = &req.Remark
		}

		// 转换订单项切片类型
		var orderItemsSlice []model.OrderItem
		for _, item := range orderItems {
			orderItemsSlice = append(orderItemsSlice, *item)
		}

		order = &model.Order{
			OrderNo:         orderNo,
			UserID:          userID,
			TotalAmount:     totalAmount,
			Status:          model.OrderStatusPending,
			PaymentMethod:   req.PaymentMethod,
			ContactPhone:    req.ContactPhone,
			ShippingAddress: req.ShippingAddress,
			Remark:          remark,
			OrderItems:      orderItemsSlice,
		}
		
		// 6. 保存订单
		if err := tx.Create(order).Error; err != nil {
			return err
		}
		
		// 7. 清理购物车
		for _, cartItem := range cartItems {
			if err := tx.Delete(cartItem).Error; err != nil {
				return err
			}
		}
		
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	
	// 8. 重新查询订单（包含关联数据）
	return s.orderRepo.GetByID(order.ID)
}

// GetOrder 获取订单详情
func (s *orderService) GetOrder(userID, orderID uint) (*model.Order, error) {
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return nil, err
	}
	
	if order == nil {
		return nil, errors.New("订单不存在")
	}
	
	// 验证订单所有权
	if order.UserID != userID {
		return nil, errors.New("无权访问此订单")
	}
	
	return order, nil
}

// GetOrderByNo 根据订单号获取订单
func (s *orderService) GetOrderByNo(userID uint, orderNo string) (*model.Order, error) {
	order, err := s.orderRepo.GetByOrderNo(orderNo)
	if err != nil {
		return nil, err
	}
	
	if order == nil {
		return nil, errors.New("订单不存在")
	}
	
	// 验证订单所有权
	if order.UserID != userID {
		return nil, errors.New("无权访问此订单")
	}
	
	return order, nil
}

// GetOrderList 获取订单列表
func (s *orderService) GetOrderList(userID uint, req *model.OrderListRequest) (*model.OrderListResponse, error) {
	orders, total, err := s.orderRepo.GetByUserID(userID, req)
	if err != nil {
		return nil, err
	}
	
	// 计算总页数
	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))
	
	return &model.OrderListResponse{
		Orders:     orders,
		Total:      int(total),
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// CancelOrder 取消订单
func (s *orderService) CancelOrder(userID, orderID uint) error {
	// 1. 获取订单
	order, err := s.GetOrder(userID, orderID)
	if err != nil {
		return err
	}
	
	// 2. 检查订单状态
	if order.Status != model.OrderStatusPending {
		return errors.New("只能取消待支付的订单")
	}
	
	// 3. 使用事务处理取消逻辑
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 4. 更新订单状态
		err := tx.Model(order).Updates(map[string]interface{}{
			"status":     model.OrderStatusCancelled,
			"updated_at": time.Now(),
		}).Error
		if err != nil {
			return err
		}
		
		// 5. 恢复库存
		for _, item := range order.OrderItems {
			err = tx.Model(&model.Product{}).
				Where("id = ?", item.ProductID).
				Update("stock", gorm.Expr("stock + ?", item.Quantity)).Error
			if err != nil {
				return err
			}
		}
		
		return nil
	})
}

// PayOrder 支付订单
func (s *orderService) PayOrder(userID, orderID uint, req *model.PayOrderRequest) error {
	// 1. 获取订单
	order, err := s.GetOrder(userID, orderID)
	if err != nil {
		return err
	}
	
	// 2. 检查订单状态
	if order.Status != model.OrderStatusPending {
		return errors.New("订单状态不正确，无法支付")
	}
	
	// 3. 模拟支付处理
	// 在实际项目中，这里会调用第三方支付接口
	if !s.simulatePayment(req.PaymentMethod, order.TotalAmount) {
		return errors.New("支付失败")
	}
	
	// 4. 更新订单状态
	return s.orderRepo.UpdateStatus(orderID, model.OrderStatusPaid)
}

// ConfirmOrder 确认收货
func (s *orderService) ConfirmOrder(userID, orderID uint) error {
	// 1. 获取订单
	order, err := s.GetOrder(userID, orderID)
	if err != nil {
		return err
	}
	
	// 2. 检查订单状态
	if order.Status != model.OrderStatusShipped {
		return errors.New("只能确认已发货的订单")
	}
	
	// 3. 更新订单状态
	return s.orderRepo.UpdateStatus(orderID, model.OrderStatusDelivered)
}

// GetOrderStatistics 获取订单统计
func (s *orderService) GetOrderStatistics(userID uint) (*model.OrderStatistics, error) {
	return s.orderRepo.GetOrderStatistics(userID)
}

// ProcessExpiredOrders 处理过期订单
func (s *orderService) ProcessExpiredOrders() error {
	return s.orderRepo.CancelExpiredOrders()
}

// generateOrderNo 生成订单号
// 格式：年月日时分秒 + 6位随机数
func (s *orderService) generateOrderNo() string {
	now := time.Now()
	return fmt.Sprintf("%s%06d", 
		now.Format("20060102150405"), 
		now.Nanosecond()%1000000)
}

// simulatePayment 模拟支付处理
func (s *orderService) simulatePayment(paymentMethod string, amount float64) bool {
	// 模拟支付逻辑
	// 在实际项目中，这里会调用支付宝、微信支付等第三方接口
	
	// 简单模拟：金额大于0且支付方式有效
	validMethods := map[string]bool{
		"alipay":  true,
		"wechat":  true,
		"balance": true,
	}
	
	return amount > 0 && validMethods[paymentMethod]
}
