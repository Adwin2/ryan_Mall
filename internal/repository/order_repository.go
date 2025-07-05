package repository

import (
	"errors"
	"ryan-mall/internal/model"
	"time"

	"gorm.io/gorm"
)

// OrderRepository 订单数据访问层接口
type OrderRepository interface {
	Create(order *model.Order) error                                    // 创建订单
	GetByID(id uint) (*model.Order, error)                             // 根据ID获取订单
	GetByOrderNo(orderNo string) (*model.Order, error)                 // 根据订单号获取订单
	GetByUserID(userID uint, req *model.OrderListRequest) ([]*model.Order, int64, error) // 获取用户订单列表
	Update(order *model.Order) error                                    // 更新订单
	UpdateStatus(id uint, status model.OrderStatus) error              // 更新订单状态
	GetOrderItems(orderID uint) ([]*model.OrderItem, error)           // 获取订单项
	CreateOrderItems(items []*model.OrderItem) error                   // 创建订单项
	GetOrderStatistics(userID uint) (*model.OrderStatistics, error)   // 获取订单统计
	CancelExpiredOrders() error                                         // 取消过期订单
}

// orderRepository 订单数据访问层实现
type orderRepository struct {
	db *gorm.DB
}

// NewOrderRepository 创建订单数据访问层实例
func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{
		db: db,
	}
}

// Create 创建订单
// 使用事务确保订单和订单项的一致性
func (r *orderRepository) Create(order *model.Order) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. 创建订单
		if err := tx.Create(order).Error; err != nil {
			return err
		}
		
		// 2. 创建订单项
		if len(order.OrderItems) > 0 {
			for i := range order.OrderItems {
				order.OrderItems[i].OrderID = order.ID
			}
			if err := tx.Create(&order.OrderItems).Error; err != nil {
				return err
			}
		}
		
		return nil
	})
}

// GetByID 根据ID获取订单
// 包含订单项和商品信息的关联查询
func (r *orderRepository) GetByID(id uint) (*model.Order, error) {
	var order model.Order
	
	err := r.db.Where("id = ?", id).
		Preload("OrderItems").
		Preload("OrderItems.Product").
		Preload("OrderItems.Product.Category").
		First(&order).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	
	return &order, nil
}

// GetByOrderNo 根据订单号获取订单
func (r *orderRepository) GetByOrderNo(orderNo string) (*model.Order, error) {
	var order model.Order
	
	err := r.db.Where("order_no = ?", orderNo).
		Preload("OrderItems").
		Preload("OrderItems.Product").
		Preload("OrderItems.Product.Category").
		First(&order).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	
	return &order, nil
}

// GetByUserID 获取用户订单列表
// 支持分页和状态筛选
func (r *orderRepository) GetByUserID(userID uint, req *model.OrderListRequest) ([]*model.Order, int64, error) {
	var orders []*model.Order
	var total int64
	
	// 构建查询条件
	query := r.db.Where("user_id = ?", userID)
	
	// 状态筛选
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}
	
	// 时间范围筛选
	if req.StartDate != nil {
		query = query.Where("created_at >= ?", *req.StartDate)
	}
	if req.EndDate != nil {
		query = query.Where("created_at <= ?", *req.EndDate)
	}
	
	// 获取总数
	if err := query.Model(&model.Order{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	err := query.Preload("OrderItems").
		Preload("OrderItems.Product").
		Order("created_at DESC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&orders).Error
	
	return orders, total, err
}

// Update 更新订单
func (r *orderRepository) Update(order *model.Order) error {
	return r.db.Save(order).Error
}

// UpdateStatus 更新订单状态
func (r *orderRepository) UpdateStatus(id uint, status model.OrderStatus) error {
	return r.db.Model(&model.Order{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		}).Error
}

// GetOrderItems 获取订单项
func (r *orderRepository) GetOrderItems(orderID uint) ([]*model.OrderItem, error) {
	var items []*model.OrderItem
	
	err := r.db.Where("order_id = ?", orderID).
		Preload("Product").
		Preload("Product.Category").
		Find(&items).Error
	
	return items, err
}

// CreateOrderItems 创建订单项
func (r *orderRepository) CreateOrderItems(items []*model.OrderItem) error {
	return r.db.Create(&items).Error
}

// GetOrderStatistics 获取订单统计
func (r *orderRepository) GetOrderStatistics(userID uint) (*model.OrderStatistics, error) {
	var stats model.OrderStatistics
	stats.UserID = userID
	
	// 统计各状态订单数量
	var statusCounts []struct {
		Status model.OrderStatus
		Count  int64
	}
	
	err := r.db.Model(&model.Order{}).
		Select("status, count(*) as count").
		Where("user_id = ?", userID).
		Group("status").
		Find(&statusCounts).Error
	
	if err != nil {
		return nil, err
	}
	
	// 填充统计数据
	for _, sc := range statusCounts {
		switch sc.Status {
		case model.OrderStatusPending:
			stats.PendingCount = int(sc.Count)
		case model.OrderStatusPaid:
			stats.PaidCount = int(sc.Count)
		case model.OrderStatusShipped:
			stats.ShippedCount = int(sc.Count)
		case model.OrderStatusDelivered:
			stats.DeliveredCount = int(sc.Count)
		case model.OrderStatusCancelled:
			stats.CancelledCount = int(sc.Count)
		}
	}
	
	// 计算总订单数和总金额
	err = r.db.Model(&model.Order{}).
		Select("count(*) as total_orders, COALESCE(sum(total_amount), 0) as total_amount").
		Where("user_id = ?", userID).
		Row().Scan(&stats.TotalOrders, &stats.TotalAmount)
	
	return &stats, err
}

// CancelExpiredOrders 取消过期订单
// 取消超过30分钟未支付的订单
func (r *orderRepository) CancelExpiredOrders() error {
	expiredTime := time.Now().Add(-30 * time.Minute)
	
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. 查找过期的待支付订单
		var expiredOrders []model.Order
		err := tx.Where("status = ? AND created_at < ?", model.OrderStatusPending, expiredTime).
			Find(&expiredOrders).Error
		if err != nil {
			return err
		}
		
		// 2. 更新订单状态为已取消
		if len(expiredOrders) > 0 {
			var orderIDs []uint
			for _, order := range expiredOrders {
				orderIDs = append(orderIDs, order.ID)
			}
			
			err = tx.Model(&model.Order{}).
				Where("id IN ?", orderIDs).
				Updates(map[string]interface{}{
					"status":     model.OrderStatusCancelled,
					"updated_at": time.Now(),
				}).Error
			if err != nil {
				return err
			}
			
			// 3. 恢复库存（这里需要获取订单项信息）
			var orderItems []model.OrderItem
			err = tx.Where("order_id IN ?", orderIDs).Find(&orderItems).Error
			if err != nil {
				return err
			}
			
			// 4. 批量恢复库存
			for _, item := range orderItems {
				err = tx.Model(&model.Product{}).
					Where("id = ?", item.ProductID).
					Update("stock", gorm.Expr("stock + ?", item.Quantity)).Error
				if err != nil {
					return err
				}
			}
		}
		
		return nil
	})
}

// GetOrdersByStatus 根据状态获取订单列表（管理员功能）
func (r *orderRepository) GetOrdersByStatus(status model.OrderStatus, page, pageSize int) ([]*model.Order, int64, error) {
	var orders []*model.Order
	var total int64
	
	query := r.db.Where("status = ?", status)
	
	// 获取总数
	if err := query.Model(&model.Order{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Preload("OrderItems").
		Preload("OrderItems.Product").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&orders).Error
	
	return orders, total, err
}
