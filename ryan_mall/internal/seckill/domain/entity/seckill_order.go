package entity

import (
	"ryan-mall-microservices/internal/shared/domain"
	"ryan-mall-microservices/internal/shared/events"
)

// SeckillOrder 秒杀订单聚合根
type SeckillOrder struct {
	id           domain.OrderID
	userID       domain.UserID
	activityID   domain.ID
	productID    domain.ProductID
	quantity     int
	price        domain.Money
	totalAmount  domain.Money
	status       SeckillOrderStatus
	createdAt    domain.Timestamp
	updatedAt    domain.Timestamp
	domainEvents []events.Event
}

// SeckillOrderStatus 秒杀订单状态
type SeckillOrderStatus string

const (
	SeckillOrderStatusPending   SeckillOrderStatus = "PENDING"   // 待支付
	SeckillOrderStatusPaid      SeckillOrderStatus = "PAID"      // 已支付
	SeckillOrderStatusCancelled SeckillOrderStatus = "CANCELLED" // 已取消
	SeckillOrderStatusExpired   SeckillOrderStatus = "EXPIRED"   // 已过期
)

// NewSeckillOrder 创建新的秒杀订单
func NewSeckillOrder(
	userID domain.UserID,
	activityID domain.ID,
	productID domain.ProductID,
	quantity int,
	price domain.Money,
) (*SeckillOrder, error) {
	// 验证用户ID
	if userID.String() == "" {
		return nil, domain.NewValidationError("user ID cannot be empty")
	}

	// 验证活动ID
	if activityID.String() == "" {
		return nil, domain.NewValidationError("activity ID cannot be empty")
	}

	// 验证商品ID
	if productID.String() == "" {
		return nil, domain.NewValidationError("product ID cannot be empty")
	}

	// 验证数量
	if quantity <= 0 {
		return nil, domain.NewValidationError("quantity must be positive")
	}

	// 验证价格
	if price.IsNegative() || price.IsZero() {
		return nil, domain.NewValidationError("price must be positive")
	}

	// 计算总金额
	totalAmount := price.Multiply(int64(quantity))

	// 创建秒杀订单
	order := &SeckillOrder{
		id:           domain.NewOrderID(),
		userID:       userID,
		activityID:   activityID,
		productID:    productID,
		quantity:     quantity,
		price:        price,
		totalAmount:  totalAmount,
		status:       SeckillOrderStatusPending,
		createdAt:    domain.Now(),
		updatedAt:    domain.Now(),
		domainEvents: make([]events.Event, 0),
	}

	// 添加秒杀订单创建事件
	event := events.NewSeckillOrderCreatedEvent(
		order.id.String(),
		order.userID.String(),
		order.activityID.String(),
		order.productID.String(),
		order.quantity,
		order.totalAmount.ToYuan(),
	)
	order.addDomainEvent(event)

	return order, nil
}

// ID 获取订单ID
func (s *SeckillOrder) ID() domain.OrderID {
	return s.id
}

// UserID 获取用户ID
func (s *SeckillOrder) UserID() domain.UserID {
	return s.userID
}

// ActivityID 获取活动ID
func (s *SeckillOrder) ActivityID() domain.ID {
	return s.activityID
}

// ProductID 获取商品ID
func (s *SeckillOrder) ProductID() domain.ProductID {
	return s.productID
}

// Quantity 获取数量
func (s *SeckillOrder) Quantity() int {
	return s.quantity
}

// Price 获取单价
func (s *SeckillOrder) Price() domain.Money {
	return s.price
}

// TotalAmount 获取总金额
func (s *SeckillOrder) TotalAmount() domain.Money {
	return s.totalAmount
}

// Status 获取状态
func (s *SeckillOrder) Status() SeckillOrderStatus {
	return s.status
}

// CreatedAt 获取创建时间
func (s *SeckillOrder) CreatedAt() domain.Timestamp {
	return s.createdAt
}

// UpdatedAt 获取更新时间
func (s *SeckillOrder) UpdatedAt() domain.Timestamp {
	return s.updatedAt
}

// Pay 支付订单
func (s *SeckillOrder) Pay() error {
	if s.status != SeckillOrderStatusPending {
		return domain.NewValidationError("only pending orders can be paid")
	}

	s.status = SeckillOrderStatusPaid
	s.updatedAt = domain.Now()

	// 添加支付完成事件
	event := events.NewSeckillOrderPaidEvent(
		s.id.String(),
		s.userID.String(),
		s.activityID.String(),
		s.totalAmount.ToYuan(),
	)
	s.addDomainEvent(event)

	return nil
}

// Cancel 取消订单
func (s *SeckillOrder) Cancel() error {
	if s.status != SeckillOrderStatusPending {
		return domain.NewValidationError("only pending orders can be cancelled")
	}

	s.status = SeckillOrderStatusCancelled
	s.updatedAt = domain.Now()

	// 添加订单取消事件
	event := events.NewSeckillOrderCancelledEvent(
		s.id.String(),
		s.userID.String(),
		s.activityID.String(),
		s.quantity,
	)
	s.addDomainEvent(event)

	return nil
}

// Expire 过期订单
func (s *SeckillOrder) Expire() error {
	if s.status != SeckillOrderStatusPending {
		return domain.NewValidationError("only pending orders can be expired")
	}

	s.status = SeckillOrderStatusExpired
	s.updatedAt = domain.Now()

	// 添加订单过期事件
	event := events.NewSeckillOrderExpiredEvent(
		s.id.String(),
		s.userID.String(),
		s.activityID.String(),
		s.quantity,
	)
	s.addDomainEvent(event)

	return nil
}

// IsPaid 是否已支付
func (s *SeckillOrder) IsPaid() bool {
	return s.status == SeckillOrderStatusPaid
}

// IsCancelled 是否已取消
func (s *SeckillOrder) IsCancelled() bool {
	return s.status == SeckillOrderStatusCancelled
}

// IsExpired 是否已过期
func (s *SeckillOrder) IsExpired() bool {
	return s.status == SeckillOrderStatusExpired
}

// CanBeCancelled 是否可以取消
func (s *SeckillOrder) CanBeCancelled() bool {
	return s.status == SeckillOrderStatusPending
}

// DomainEvents 获取领域事件
func (s *SeckillOrder) DomainEvents() []events.Event {
	return s.domainEvents
}

// ClearDomainEvents 清除领域事件
func (s *SeckillOrder) ClearDomainEvents() {
	s.domainEvents = make([]events.Event, 0)
}

// addDomainEvent 添加领域事件
func (s *SeckillOrder) addDomainEvent(event events.Event) {
	s.domainEvents = append(s.domainEvents, event)
}

// ReconstructSeckillOrder 重建秒杀订单实体（用于从持久化存储重建）
func ReconstructSeckillOrder(
	id domain.OrderID,
	userID domain.UserID,
	activityID domain.ID,
	productID domain.ProductID,
	quantity int,
	price domain.Money,
	totalAmount domain.Money,
	status SeckillOrderStatus,
	createdAt domain.Timestamp,
	updatedAt domain.Timestamp,
) *SeckillOrder {
	return &SeckillOrder{
		id:           id,
		userID:       userID,
		activityID:   activityID,
		productID:    productID,
		quantity:     quantity,
		price:        price,
		totalAmount:  totalAmount,
		status:       status,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
		domainEvents: make([]events.Event, 0),
	}
}
