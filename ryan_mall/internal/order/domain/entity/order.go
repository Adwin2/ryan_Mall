package entity

import (
	"ryan-mall-microservices/internal/shared/domain"
	"ryan-mall-microservices/internal/shared/events"
)

// Order 订单聚合根
type Order struct {
	id              domain.OrderID
	userID          domain.UserID
	items           []*OrderItem
	totalAmount     domain.Money
	status          OrderStatus
	shippingAddress string
	cancelReason    string
	createdAt       domain.Timestamp
	updatedAt       domain.Timestamp
	domainEvents    []events.Event
}

// OrderStatus 订单状态
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "PENDING"   // 待处理
	OrderStatusConfirmed OrderStatus = "CONFIRMED" // 已确认
	OrderStatusPaid      OrderStatus = "PAID"      // 已支付
	OrderStatusShipped   OrderStatus = "SHIPPED"   // 已发货
	OrderStatusCompleted OrderStatus = "COMPLETED" // 已完成
	OrderStatusCancelled OrderStatus = "CANCELLED" // 已取消
)

// OrderItem 订单项
type OrderItem struct {
	productID  domain.ProductID
	quantity   int
	price      domain.Money
	totalPrice domain.Money
}

// OrderItemData 订单项数据（用于创建订单）
type OrderItemData struct {
	ProductID domain.ProductID
	Quantity  int
	Price     domain.Money
}

// NewOrder 创建新订单
func NewOrder(userID domain.UserID, itemsData []OrderItemData) (*Order, error) {
	// 验证用户ID
	if userID.String() == "" {
		return nil, domain.NewValidationError("user ID cannot be empty")
	}

	// 验证订单项
	if len(itemsData) == 0 {
		return nil, domain.NewValidationError("order must have at least one item")
	}

	// 创建订单项
	items := make([]*OrderItem, len(itemsData))
	totalAmount := domain.NewMoney(0, "CNY")

	for i, itemData := range itemsData {
		if err := validateOrderItemData(itemData); err != nil {
			return nil, err
		}

		item := &OrderItem{
			productID:  itemData.ProductID,
			quantity:   itemData.Quantity,
			price:      itemData.Price,
			totalPrice: itemData.Price.Multiply(int64(itemData.Quantity)),
		}

		items[i] = item
		totalAmount = totalAmount.Add(item.totalPrice)
	}

	// 创建订单
	order := &Order{
		id:           domain.NewOrderID(),
		userID:       userID,
		items:        items,
		totalAmount:  totalAmount,
		status:       OrderStatusPending,
		createdAt:    domain.Now(),
		updatedAt:    domain.Now(),
		domainEvents: make([]events.Event, 0),
	}

	// 添加订单创建事件
	event := events.NewOrderCreatedEvent(
		order.id.String(),
		order.userID.String(),
		order.totalAmount.ToYuan(),
	)
	order.addDomainEvent(event)

	return order, nil
}

// ID 获取订单ID
func (o *Order) ID() domain.OrderID {
	return o.id
}

// UserID 获取用户ID
func (o *Order) UserID() domain.UserID {
	return o.userID
}

// Items 获取订单项
func (o *Order) Items() []*OrderItem {
	return o.items
}

// TotalAmount 获取总金额
func (o *Order) TotalAmount() domain.Money {
	return o.totalAmount
}

// Status 获取订单状态
func (o *Order) Status() OrderStatus {
	return o.status
}

// ShippingAddress 获取收货地址
func (o *Order) ShippingAddress() string {
	return o.shippingAddress
}

// CancelReason 获取取消原因
func (o *Order) CancelReason() string {
	return o.cancelReason
}

// CreatedAt 获取创建时间
func (o *Order) CreatedAt() domain.Timestamp {
	return o.createdAt
}

// UpdatedAt 获取更新时间
func (o *Order) UpdatedAt() domain.Timestamp {
	return o.updatedAt
}

// Confirm 确认订单
func (o *Order) Confirm() error {
	if o.status != OrderStatusPending {
		return domain.NewValidationError("only pending orders can be confirmed")
	}

	o.status = OrderStatusConfirmed
	o.updatedAt = domain.Now()

	// 添加订单确认事件
	event := events.NewOrderConfirmedEvent(
		o.id.String(),
		o.userID.String(),
	)
	o.addDomainEvent(event)

	return nil
}

// Cancel 取消订单
func (o *Order) Cancel(reason string) error {
	if o.status == OrderStatusCompleted || o.status == OrderStatusCancelled {
		return domain.NewValidationError("cannot cancel completed or already cancelled order")
	}

	o.status = OrderStatusCancelled
	o.cancelReason = reason
	o.updatedAt = domain.Now()

	// 添加订单取消事件
	event := events.NewOrderCancelledEvent(
		o.id.String(),
		o.userID.String(),
		reason,
	)
	o.addDomainEvent(event)

	return nil
}

// Complete 完成订单
func (o *Order) Complete() error {
	if o.status != OrderStatusConfirmed && o.status != OrderStatusPaid && o.status != OrderStatusShipped {
		return domain.NewValidationError("only confirmed, paid or shipped orders can be completed")
	}

	o.status = OrderStatusCompleted
	o.updatedAt = domain.Now()

	// 添加订单完成事件
	event := events.NewOrderCompletedEvent(
		o.id.String(),
		o.userID.String(),
	)
	o.addDomainEvent(event)

	return nil
}

// MarkAsPaid 标记为已支付
func (o *Order) MarkAsPaid() error {
	if o.status != OrderStatusConfirmed {
		return domain.NewValidationError("only confirmed orders can be marked as paid")
	}

	o.status = OrderStatusPaid
	o.updatedAt = domain.Now()

	return nil
}

// MarkAsShipped 标记为已发货
func (o *Order) MarkAsShipped() error {
	if o.status != OrderStatusPaid {
		return domain.NewValidationError("only paid orders can be marked as shipped")
	}

	o.status = OrderStatusShipped
	o.updatedAt = domain.Now()

	return nil
}

// UpdateShippingAddress 更新收货地址
func (o *Order) UpdateShippingAddress(address string) error {
	if o.status == OrderStatusCompleted || o.status == OrderStatusCancelled {
		return domain.NewValidationError("cannot update address for completed or cancelled orders")
	}

	o.shippingAddress = address
	o.updatedAt = domain.Now()

	return nil
}

// DomainEvents 获取领域事件
func (o *Order) DomainEvents() []events.Event {
	return o.domainEvents
}

// ClearDomainEvents 清除领域事件
func (o *Order) ClearDomainEvents() {
	o.domainEvents = make([]events.Event, 0)
}

// addDomainEvent 添加领域事件
func (o *Order) addDomainEvent(event events.Event) {
	o.domainEvents = append(o.domainEvents, event)
}

// ProductID 获取商品ID
func (oi *OrderItem) ProductID() domain.ProductID {
	return oi.productID
}

// Quantity 获取数量
func (oi *OrderItem) Quantity() int {
	return oi.quantity
}

// Price 获取单价
func (oi *OrderItem) Price() domain.Money {
	return oi.price
}

// TotalPrice 获取总价
func (oi *OrderItem) TotalPrice() domain.Money {
	return oi.totalPrice
}

// validateOrderItemData 验证订单项数据
func validateOrderItemData(data OrderItemData) error {
	if data.ProductID.String() == "" {
		return domain.NewValidationError("product ID cannot be empty")
	}

	if data.Quantity <= 0 {
		return domain.NewValidationError("quantity must be positive")
	}

	if data.Price.IsNegative() || data.Price.IsZero() {
		return domain.NewValidationError("price must be positive")
	}

	return nil
}

// ReconstructOrder 重建订单实体（用于从持久化存储重建）
func ReconstructOrder(
	id domain.OrderID,
	userID domain.UserID,
	items []*OrderItem,
	totalAmount domain.Money,
	status OrderStatus,
	shippingAddress string,
	cancelReason string,
	createdAt domain.Timestamp,
	updatedAt domain.Timestamp,
) *Order {
	return &Order{
		id:              id,
		userID:          userID,
		items:           items,
		totalAmount:     totalAmount,
		status:          status,
		shippingAddress: shippingAddress,
		cancelReason:    cancelReason,
		createdAt:       createdAt,
		updatedAt:       updatedAt,
		domainEvents:    make([]events.Event, 0),
	}
}

// ReconstructOrderItem 重建订单项
func ReconstructOrderItem(
	productID domain.ProductID,
	quantity int,
	price domain.Money,
	totalPrice domain.Money,
) *OrderItem {
	return &OrderItem{
		productID:  productID,
		quantity:   quantity,
		price:      price,
		totalPrice: totalPrice,
	}
}
