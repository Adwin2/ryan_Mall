package events

import (
	"context"
	"encoding/json"
	"time"

	"ryan-mall-microservices/internal/shared/domain"
)

// Event 领域事件接口
type Event interface {
	EventID() string
	EventType() string
	AggregateID() string
	AggregateType() string
	EventVersion() int
	OccurredAt() time.Time
	Data() interface{}
}

// BaseEvent 基础事件实现
type BaseEvent struct {
	ID              string      `json:"event_id"`
	Type            string      `json:"event_type"`
	AggregateRootID string      `json:"aggregate_id"`
	AggregateRootType string    `json:"aggregate_type"`
	Version         int         `json:"event_version"`
	Timestamp       time.Time   `json:"occurred_at"`
	EventData       interface{} `json:"data"`
}

// NewBaseEvent 创建基础事件
func NewBaseEvent(eventType, aggregateID, aggregateType string, data interface{}) *BaseEvent {
	return &BaseEvent{
		ID:                domain.NewID().String(),
		Type:              eventType,
		AggregateRootID:   aggregateID,
		AggregateRootType: aggregateType,
		Version:           1,
		Timestamp:         time.Now(),
		EventData:         data,
	}
}

// EventID 获取事件ID
func (e *BaseEvent) EventID() string {
	return e.ID
}

// EventType 获取事件类型
func (e *BaseEvent) EventType() string {
	return e.Type
}

// AggregateID 获取聚合根ID
func (e *BaseEvent) AggregateID() string {
	return e.AggregateRootID
}

// AggregateType 获取聚合根类型
func (e *BaseEvent) AggregateType() string {
	return e.AggregateRootType
}

// EventVersion 获取事件版本
func (e *BaseEvent) EventVersion() int {
	return e.Version
}

// OccurredAt 获取事件发生时间
func (e *BaseEvent) OccurredAt() time.Time {
	return e.Timestamp
}

// Data 获取事件数据
func (e *BaseEvent) Data() interface{} {
	return e.EventData
}

// ToJSON 转换为JSON
func (e *BaseEvent) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// EventHandler 事件处理器接口
type EventHandler interface {
	Handle(ctx context.Context, event Event) error
	EventType() string
}

// EventBus 事件总线接口
type EventBus interface {
	Publish(ctx context.Context, events ...Event) error
	Subscribe(handler EventHandler) error
	Unsubscribe(eventType string) error
}

// EventStore 事件存储接口
type EventStore interface {
	Save(ctx context.Context, events ...Event) error
	Load(ctx context.Context, aggregateID string) ([]Event, error)
	LoadFromVersion(ctx context.Context, aggregateID string, version int) ([]Event, error)
}

// InMemoryEventBus 内存事件总线实现
type InMemoryEventBus struct {
	handlers map[string][]EventHandler
}

// NewInMemoryEventBus 创建内存事件总线
func NewInMemoryEventBus() *InMemoryEventBus {
	return &InMemoryEventBus{
		handlers: make(map[string][]EventHandler),
	}
}

// Publish 发布事件
func (bus *InMemoryEventBus) Publish(ctx context.Context, events ...Event) error {
	for _, event := range events {
		handlers, exists := bus.handlers[event.EventType()]
		if !exists {
			continue
		}

		for _, handler := range handlers {
			// 异步处理事件，避免阻塞
			go func(h EventHandler, e Event) {
				if err := h.Handle(ctx, e); err != nil {
					// 在实际项目中，这里应该记录日志或者重试
					// log.Error("failed to handle event", "error", err, "event", e)
				}
			}(handler, event)
		}
	}
	return nil
}

// Subscribe 订阅事件
func (bus *InMemoryEventBus) Subscribe(handler EventHandler) error {
	eventType := handler.EventType()
	bus.handlers[eventType] = append(bus.handlers[eventType], handler)
	return nil
}

// Unsubscribe 取消订阅
func (bus *InMemoryEventBus) Unsubscribe(eventType string) error {
	delete(bus.handlers, eventType)
	return nil
}

// EventPublisher 事件发布器
type EventPublisher struct {
	eventBus   EventBus
	eventStore EventStore
}

// NewEventPublisher 创建事件发布器
func NewEventPublisher(eventBus EventBus, eventStore EventStore) *EventPublisher {
	return &EventPublisher{
		eventBus:   eventBus,
		eventStore: eventStore,
	}
}

// PublishEvents 发布事件
func (p *EventPublisher) PublishEvents(ctx context.Context, events ...Event) error {
	// 先保存事件到事件存储
	if p.eventStore != nil {
		if err := p.eventStore.Save(ctx, events...); err != nil {
			return err
		}
	}

	// 然后发布事件到事件总线
	return p.eventBus.Publish(ctx, events...)
}

// 常用事件类型常量
const (
	// 用户事件
	UserRegisteredEventType = "user.registered"
	UserUpdatedEventType    = "user.updated"
	UserDeletedEventType    = "user.deleted"

	// 商品事件
	ProductCreatedEventType = "product.created"
	ProductUpdatedEventType = "product.updated"
	ProductDeletedEventType = "product.deleted"
	StockUpdatedEventType   = "stock.updated"

	// 订单事件
	OrderCreatedEventType   = "order.created"
	OrderConfirmedEventType = "order.confirmed"
	OrderCancelledEventType = "order.cancelled"
	OrderCompletedEventType = "order.completed"

	// 支付事件
	PaymentInitiatedEventType = "payment.initiated"
	PaymentCompletedEventType = "payment.completed"
	PaymentFailedEventType    = "payment.failed"

	// 秒杀事件
	SeckillStartedEventType = "seckill.started"
	SeckillEndedEventType   = "seckill.ended"
	SeckillOrderEventType   = "seckill.order"
)

// UserRegisteredEvent 用户注册事件
type UserRegisteredEvent struct {
	*BaseEvent
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// NewUserRegisteredEvent 创建用户注册事件
func NewUserRegisteredEvent(userID, username, email string) *UserRegisteredEvent {
	data := map[string]interface{}{
		"user_id":  userID,
		"username": username,
		"email":    email,
	}
	
	return &UserRegisteredEvent{
		BaseEvent: NewBaseEvent(UserRegisteredEventType, userID, "User", data),
		UserID:    userID,
		Username:  username,
		Email:     email,
	}
}

// OrderCreatedEvent 订单创建事件
type OrderCreatedEvent struct {
	*BaseEvent
	OrderID     string  `json:"order_id"`
	UserID      string  `json:"user_id"`
	TotalAmount float64 `json:"total_amount"`
}

// NewOrderCreatedEvent 创建订单创建事件
func NewOrderCreatedEvent(orderID, userID string, totalAmount float64) *OrderCreatedEvent {
	data := map[string]interface{}{
		"order_id":     orderID,
		"user_id":      userID,
		"total_amount": totalAmount,
	}

	return &OrderCreatedEvent{
		BaseEvent:   NewBaseEvent(OrderCreatedEventType, orderID, "Order", data),
		OrderID:     orderID,
		UserID:      userID,
		TotalAmount: totalAmount,
	}
}

// ProductCreatedEvent 商品创建事件
type ProductCreatedEvent struct {
	*BaseEvent
	ProductID  string  `json:"product_id"`
	Name       string  `json:"name"`
	CategoryID string  `json:"category_id"`
	Price      float64 `json:"price"`
}

// NewProductCreatedEvent 创建商品创建事件
func NewProductCreatedEvent(productID, name, categoryID string, price float64) *ProductCreatedEvent {
	data := map[string]interface{}{
		"product_id":  productID,
		"name":        name,
		"category_id": categoryID,
		"price":       price,
	}

	return &ProductCreatedEvent{
		BaseEvent:  NewBaseEvent(ProductCreatedEventType, productID, "Product", data),
		ProductID:  productID,
		Name:       name,
		CategoryID: categoryID,
		Price:      price,
	}
}

// PriceUpdatedEvent 价格更新事件
type PriceUpdatedEvent struct {
	*BaseEvent
	ProductID string  `json:"product_id"`
	OldPrice  float64 `json:"old_price"`
	NewPrice  float64 `json:"new_price"`
}

// NewPriceUpdatedEvent 创建价格更新事件
func NewPriceUpdatedEvent(productID string, oldPrice, newPrice float64) *PriceUpdatedEvent {
	data := map[string]interface{}{
		"product_id": productID,
		"old_price":  oldPrice,
		"new_price":  newPrice,
	}

	return &PriceUpdatedEvent{
		BaseEvent: NewBaseEvent("product.price_updated", productID, "Product", data),
		ProductID: productID,
		OldPrice:  oldPrice,
		NewPrice:  newPrice,
	}
}

// StockUpdatedEvent 库存更新事件
type StockUpdatedEvent struct {
	*BaseEvent
	ProductID string `json:"product_id"`
	OldStock  int    `json:"old_stock"`
	NewStock  int    `json:"new_stock"`
}

// NewStockUpdatedEvent 创建库存更新事件
func NewStockUpdatedEvent(productID string, oldStock, newStock int) *StockUpdatedEvent {
	data := map[string]interface{}{
		"product_id": productID,
		"old_stock":  oldStock,
		"new_stock":  newStock,
	}

	return &StockUpdatedEvent{
		BaseEvent: NewBaseEvent(StockUpdatedEventType, productID, "Product", data),
		ProductID: productID,
		OldStock:  oldStock,
		NewStock:  newStock,
	}
}

// StockReservedEvent 库存预留事件
type StockReservedEvent struct {
	*BaseEvent
	ProductID     string `json:"product_id"`
	OrderID       string `json:"order_id"`
	ReservedQty   int    `json:"reserved_quantity"`
	RemainingQty  int    `json:"remaining_quantity"`
}

// NewStockReservedEvent 创建库存预留事件
func NewStockReservedEvent(productID string, reservedQty int, orderID string, remainingQty int) *StockReservedEvent {
	data := map[string]interface{}{
		"product_id":         productID,
		"order_id":           orderID,
		"reserved_quantity":  reservedQty,
		"remaining_quantity": remainingQty,
	}

	return &StockReservedEvent{
		BaseEvent:    NewBaseEvent("stock.reserved", productID, "Product", data),
		ProductID:    productID,
		OrderID:      orderID,
		ReservedQty:  reservedQty,
		RemainingQty: remainingQty,
	}
}

// StockReleasedEvent 库存释放事件
type StockReleasedEvent struct {
	*BaseEvent
	ProductID    string `json:"product_id"`
	OrderID      string `json:"order_id"`
	ReleasedQty  int    `json:"released_quantity"`
	CurrentQty   int    `json:"current_quantity"`
}

// NewStockReleasedEvent 创建库存释放事件
func NewStockReleasedEvent(productID string, releasedQty int, orderID string, currentQty int) *StockReleasedEvent {
	data := map[string]interface{}{
		"product_id":        productID,
		"order_id":          orderID,
		"released_quantity": releasedQty,
		"current_quantity":  currentQty,
	}

	return &StockReleasedEvent{
		BaseEvent:   NewBaseEvent("stock.released", productID, "Product", data),
		ProductID:   productID,
		OrderID:     orderID,
		ReleasedQty: releasedQty,
		CurrentQty:  currentQty,
	}
}

// CategoryCreatedEvent 分类创建事件
type CategoryCreatedEvent struct {
	*BaseEvent
	CategoryID string `json:"category_id"`
	Name       string `json:"name"`
	Level      int    `json:"level"`
}

// NewCategoryCreatedEvent 创建分类创建事件
func NewCategoryCreatedEvent(categoryID, name string, level int) *CategoryCreatedEvent {
	data := map[string]interface{}{
		"category_id": categoryID,
		"name":        name,
		"level":       level,
	}

	return &CategoryCreatedEvent{
		BaseEvent:  NewBaseEvent("category.created", categoryID, "Category", data),
		CategoryID: categoryID,
		Name:       name,
		Level:      level,
	}
}

// OrderConfirmedEvent 订单确认事件
type OrderConfirmedEvent struct {
	*BaseEvent
	OrderID string `json:"order_id"`
	UserID  string `json:"user_id"`
}

// NewOrderConfirmedEvent 创建订单确认事件
func NewOrderConfirmedEvent(orderID, userID string) *OrderConfirmedEvent {
	data := map[string]interface{}{
		"order_id": orderID,
		"user_id":  userID,
	}

	return &OrderConfirmedEvent{
		BaseEvent: NewBaseEvent(OrderConfirmedEventType, orderID, "Order", data),
		OrderID:   orderID,
		UserID:    userID,
	}
}

// OrderCancelledEvent 订单取消事件
type OrderCancelledEvent struct {
	*BaseEvent
	OrderID string `json:"order_id"`
	UserID  string `json:"user_id"`
	Reason  string `json:"reason"`
}

// NewOrderCancelledEvent 创建订单取消事件
func NewOrderCancelledEvent(orderID, userID, reason string) *OrderCancelledEvent {
	data := map[string]interface{}{
		"order_id": orderID,
		"user_id":  userID,
		"reason":   reason,
	}

	return &OrderCancelledEvent{
		BaseEvent: NewBaseEvent(OrderCancelledEventType, orderID, "Order", data),
		OrderID:   orderID,
		UserID:    userID,
		Reason:    reason,
	}
}

// OrderCompletedEvent 订单完成事件
type OrderCompletedEvent struct {
	*BaseEvent
	OrderID string `json:"order_id"`
	UserID  string `json:"user_id"`
}

// NewOrderCompletedEvent 创建订单完成事件
func NewOrderCompletedEvent(orderID, userID string) *OrderCompletedEvent {
	data := map[string]interface{}{
		"order_id": orderID,
		"user_id":  userID,
	}

	return &OrderCompletedEvent{
		BaseEvent: NewBaseEvent(OrderCompletedEventType, orderID, "Order", data),
		OrderID:   orderID,
		UserID:    userID,
	}
}

// SeckillCreatedEvent 秒杀活动创建事件
type SeckillCreatedEvent struct {
	*BaseEvent
	ActivityID   string  `json:"activity_id"`
	ProductID    string  `json:"product_id"`
	Name         string  `json:"name"`
	SeckillPrice float64 `json:"seckill_price"`
	TotalStock   int     `json:"total_stock"`
	StartTime    int64   `json:"start_time"`
	EndTime      int64   `json:"end_time"`
}

// NewSeckillCreatedEvent 创建秒杀活动创建事件
func NewSeckillCreatedEvent(activityID, productID, name string, seckillPrice float64, totalStock int, startTime, endTime int64) *SeckillCreatedEvent {
	data := map[string]interface{}{
		"activity_id":   activityID,
		"product_id":    productID,
		"name":          name,
		"seckill_price": seckillPrice,
		"total_stock":   totalStock,
		"start_time":    startTime,
		"end_time":      endTime,
	}

	return &SeckillCreatedEvent{
		BaseEvent:    NewBaseEvent("seckill.created", activityID, "SeckillActivity", data),
		ActivityID:   activityID,
		ProductID:    productID,
		Name:         name,
		SeckillPrice: seckillPrice,
		TotalStock:   totalStock,
		StartTime:    startTime,
		EndTime:      endTime,
	}
}

// SeckillStartedEvent 秒杀开始事件
type SeckillStartedEvent struct {
	*BaseEvent
	ActivityID string `json:"activity_id"`
	ProductID  string `json:"product_id"`
	Name       string `json:"name"`
}

// NewSeckillStartedEvent 创建秒杀开始事件
func NewSeckillStartedEvent(activityID, productID, name string) *SeckillStartedEvent {
	data := map[string]interface{}{
		"activity_id": activityID,
		"product_id":  productID,
		"name":        name,
	}

	return &SeckillStartedEvent{
		BaseEvent:  NewBaseEvent(SeckillStartedEventType, activityID, "SeckillActivity", data),
		ActivityID: activityID,
		ProductID:  productID,
		Name:       name,
	}
}

// SeckillEndedEvent 秒杀结束事件
type SeckillEndedEvent struct {
	*BaseEvent
	ActivityID     string `json:"activity_id"`
	ProductID      string `json:"product_id"`
	Name           string `json:"name"`
	SoldCount      int    `json:"sold_count"`
	RemainingStock int    `json:"remaining_stock"`
}

// NewSeckillEndedEvent 创建秒杀结束事件
func NewSeckillEndedEvent(activityID, productID, name string, soldCount, remainingStock int) *SeckillEndedEvent {
	data := map[string]interface{}{
		"activity_id":     activityID,
		"product_id":      productID,
		"name":            name,
		"sold_count":      soldCount,
		"remaining_stock": remainingStock,
	}

	return &SeckillEndedEvent{
		BaseEvent:      NewBaseEvent(SeckillEndedEventType, activityID, "SeckillActivity", data),
		ActivityID:     activityID,
		ProductID:      productID,
		Name:           name,
		SoldCount:      soldCount,
		RemainingStock: remainingStock,
	}
}

// SeckillStockReservedEvent 秒杀库存预留事件
type SeckillStockReservedEvent struct {
	*BaseEvent
	ActivityID     string `json:"activity_id"`
	ProductID      string `json:"product_id"`
	ReservedQty    int    `json:"reserved_quantity"`
	RemainingStock int    `json:"remaining_stock"`
}

// NewSeckillStockReservedEvent 创建秒杀库存预留事件
func NewSeckillStockReservedEvent(activityID, productID string, reservedQty, remainingStock int) *SeckillStockReservedEvent {
	data := map[string]interface{}{
		"activity_id":       activityID,
		"product_id":        productID,
		"reserved_quantity": reservedQty,
		"remaining_stock":   remainingStock,
	}

	return &SeckillStockReservedEvent{
		BaseEvent:      NewBaseEvent("seckill.stock_reserved", activityID, "SeckillActivity", data),
		ActivityID:     activityID,
		ProductID:      productID,
		ReservedQty:    reservedQty,
		RemainingStock: remainingStock,
	}
}

// SeckillOrderCreatedEvent 秒杀订单创建事件
type SeckillOrderCreatedEvent struct {
	*BaseEvent
	OrderID     string  `json:"order_id"`
	UserID      string  `json:"user_id"`
	ActivityID  string  `json:"activity_id"`
	ProductID   string  `json:"product_id"`
	Quantity    int     `json:"quantity"`
	TotalAmount float64 `json:"total_amount"`
}

// NewSeckillOrderCreatedEvent 创建秒杀订单创建事件
func NewSeckillOrderCreatedEvent(orderID, userID, activityID, productID string, quantity int, totalAmount float64) *SeckillOrderCreatedEvent {
	data := map[string]interface{}{
		"order_id":     orderID,
		"user_id":      userID,
		"activity_id":  activityID,
		"product_id":   productID,
		"quantity":     quantity,
		"total_amount": totalAmount,
	}

	return &SeckillOrderCreatedEvent{
		BaseEvent:   NewBaseEvent(SeckillOrderEventType, orderID, "SeckillOrder", data),
		OrderID:     orderID,
		UserID:      userID,
		ActivityID:  activityID,
		ProductID:   productID,
		Quantity:    quantity,
		TotalAmount: totalAmount,
	}
}

// SeckillOrderPaidEvent 秒杀订单支付事件
type SeckillOrderPaidEvent struct {
	*BaseEvent
	OrderID     string  `json:"order_id"`
	UserID      string  `json:"user_id"`
	ActivityID  string  `json:"activity_id"`
	TotalAmount float64 `json:"total_amount"`
}

// NewSeckillOrderPaidEvent 创建秒杀订单支付事件
func NewSeckillOrderPaidEvent(orderID, userID, activityID string, totalAmount float64) *SeckillOrderPaidEvent {
	data := map[string]interface{}{
		"order_id":     orderID,
		"user_id":      userID,
		"activity_id":  activityID,
		"total_amount": totalAmount,
	}

	return &SeckillOrderPaidEvent{
		BaseEvent:   NewBaseEvent("seckill.order_paid", orderID, "SeckillOrder", data),
		OrderID:     orderID,
		UserID:      userID,
		ActivityID:  activityID,
		TotalAmount: totalAmount,
	}
}

// SeckillOrderCancelledEvent 秒杀订单取消事件
type SeckillOrderCancelledEvent struct {
	*BaseEvent
	OrderID    string `json:"order_id"`
	UserID     string `json:"user_id"`
	ActivityID string `json:"activity_id"`
	Quantity   int    `json:"quantity"`
}

// NewSeckillOrderCancelledEvent 创建秒杀订单取消事件
func NewSeckillOrderCancelledEvent(orderID, userID, activityID string, quantity int) *SeckillOrderCancelledEvent {
	data := map[string]interface{}{
		"order_id":    orderID,
		"user_id":     userID,
		"activity_id": activityID,
		"quantity":    quantity,
	}

	return &SeckillOrderCancelledEvent{
		BaseEvent:  NewBaseEvent("seckill.order_cancelled", orderID, "SeckillOrder", data),
		OrderID:    orderID,
		UserID:     userID,
		ActivityID: activityID,
		Quantity:   quantity,
	}
}

// SeckillOrderExpiredEvent 秒杀订单过期事件
type SeckillOrderExpiredEvent struct {
	*BaseEvent
	OrderID    string `json:"order_id"`
	UserID     string `json:"user_id"`
	ActivityID string `json:"activity_id"`
	Quantity   int    `json:"quantity"`
}

// NewSeckillOrderExpiredEvent 创建秒杀订单过期事件
func NewSeckillOrderExpiredEvent(orderID, userID, activityID string, quantity int) *SeckillOrderExpiredEvent {
	data := map[string]interface{}{
		"order_id":    orderID,
		"user_id":     userID,
		"activity_id": activityID,
		"quantity":    quantity,
	}

	return &SeckillOrderExpiredEvent{
		BaseEvent:  NewBaseEvent("seckill.order_expired", orderID, "SeckillOrder", data),
		OrderID:    orderID,
		UserID:     userID,
		ActivityID: activityID,
		Quantity:   quantity,
	}
}

// PaymentCreatedEvent 支付创建事件
type PaymentCreatedEvent struct {
	*BaseEvent
	PaymentID string  `json:"payment_id"`
	UserID    string  `json:"user_id"`
	OrderID   string  `json:"order_id"`
	Amount    float64 `json:"amount"`
	Method    string  `json:"method"`
}

// NewPaymentCreatedEvent 创建支付创建事件
func NewPaymentCreatedEvent(paymentID, userID, orderID string, amount float64, method string) *PaymentCreatedEvent {
	data := map[string]interface{}{
		"payment_id": paymentID,
		"user_id":    userID,
		"order_id":   orderID,
		"amount":     amount,
		"method":     method,
	}

	return &PaymentCreatedEvent{
		BaseEvent: NewBaseEvent("payment.created", paymentID, "Payment", data),
		PaymentID: paymentID,
		UserID:    userID,
		OrderID:   orderID,
		Amount:    amount,
		Method:    method,
	}
}

// PaymentProcessingEvent 支付处理事件
type PaymentProcessingEvent struct {
	*BaseEvent
	PaymentID     string `json:"payment_id"`
	UserID        string `json:"user_id"`
	OrderID       string `json:"order_id"`
	TransactionID string `json:"transaction_id"`
}

// NewPaymentProcessingEvent 创建支付处理事件
func NewPaymentProcessingEvent(paymentID, userID, orderID, transactionID string) *PaymentProcessingEvent {
	data := map[string]interface{}{
		"payment_id":     paymentID,
		"user_id":        userID,
		"order_id":       orderID,
		"transaction_id": transactionID,
	}

	return &PaymentProcessingEvent{
		BaseEvent:     NewBaseEvent("payment.processing", paymentID, "Payment", data),
		PaymentID:     paymentID,
		UserID:        userID,
		OrderID:       orderID,
		TransactionID: transactionID,
	}
}

// PaymentCompletedEvent 支付完成事件
type PaymentCompletedEvent struct {
	*BaseEvent
	PaymentID     string  `json:"payment_id"`
	UserID        string  `json:"user_id"`
	OrderID       string  `json:"order_id"`
	Amount        float64 `json:"amount"`
	TransactionID string  `json:"transaction_id"`
}

// NewPaymentCompletedEvent 创建支付完成事件
func NewPaymentCompletedEvent(paymentID, userID, orderID string, amount float64, transactionID string) *PaymentCompletedEvent {
	data := map[string]interface{}{
		"payment_id":     paymentID,
		"user_id":        userID,
		"order_id":       orderID,
		"amount":         amount,
		"transaction_id": transactionID,
	}

	return &PaymentCompletedEvent{
		BaseEvent:     NewBaseEvent("payment.completed", paymentID, "Payment", data),
		PaymentID:     paymentID,
		UserID:        userID,
		OrderID:       orderID,
		Amount:        amount,
		TransactionID: transactionID,
	}
}

// PaymentFailedEvent 支付失败事件
type PaymentFailedEvent struct {
	*BaseEvent
	PaymentID     string `json:"payment_id"`
	UserID        string `json:"user_id"`
	OrderID       string `json:"order_id"`
	FailureReason string `json:"failure_reason"`
}

// NewPaymentFailedEvent 创建支付失败事件
func NewPaymentFailedEvent(paymentID, userID, orderID, failureReason string) *PaymentFailedEvent {
	data := map[string]interface{}{
		"payment_id":     paymentID,
		"user_id":        userID,
		"order_id":       orderID,
		"failure_reason": failureReason,
	}

	return &PaymentFailedEvent{
		BaseEvent:     NewBaseEvent("payment.failed", paymentID, "Payment", data),
		PaymentID:     paymentID,
		UserID:        userID,
		OrderID:       orderID,
		FailureReason: failureReason,
	}
}

// PaymentCancelledEvent 支付取消事件
type PaymentCancelledEvent struct {
	*BaseEvent
	PaymentID string `json:"payment_id"`
	UserID    string `json:"user_id"`
	OrderID   string `json:"order_id"`
}

// NewPaymentCancelledEvent 创建支付取消事件
func NewPaymentCancelledEvent(paymentID, userID, orderID string) *PaymentCancelledEvent {
	data := map[string]interface{}{
		"payment_id": paymentID,
		"user_id":    userID,
		"order_id":   orderID,
	}

	return &PaymentCancelledEvent{
		BaseEvent: NewBaseEvent("payment.cancelled", paymentID, "Payment", data),
		PaymentID: paymentID,
		UserID:    userID,
		OrderID:   orderID,
	}
}

// PaymentRefundedEvent 支付退款事件
type PaymentRefundedEvent struct {
	*BaseEvent
	PaymentID    string  `json:"payment_id"`
	UserID       string  `json:"user_id"`
	OrderID      string  `json:"order_id"`
	RefundID     string  `json:"refund_id"`
	RefundAmount float64 `json:"refund_amount"`
	RefundReason string  `json:"refund_reason"`
}

// NewPaymentRefundedEvent 创建支付退款事件
func NewPaymentRefundedEvent(paymentID, userID, orderID, refundID string, refundAmount float64, refundReason string) *PaymentRefundedEvent {
	data := map[string]interface{}{
		"payment_id":    paymentID,
		"user_id":       userID,
		"order_id":      orderID,
		"refund_id":     refundID,
		"refund_amount": refundAmount,
		"refund_reason": refundReason,
	}

	return &PaymentRefundedEvent{
		BaseEvent:    NewBaseEvent("payment.refunded", paymentID, "Payment", data),
		PaymentID:    paymentID,
		UserID:       userID,
		OrderID:      orderID,
		RefundID:     refundID,
		RefundAmount: refundAmount,
		RefundReason: refundReason,
	}
}

// SeckillActivityCreatedEvent 秒杀活动创建事件
type SeckillActivityCreatedEvent struct {
	*BaseEvent
	ActivityID string    `json:"activity_id"`
	ProductID  string    `json:"product_id"`
	Name       string    `json:"name"`
	TotalStock int       `json:"total_stock"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
}

// NewSeckillActivityCreatedEvent 创建秒杀活动创建事件
func NewSeckillActivityCreatedEvent(activityID, productID, name string, totalStock int, startTime, endTime time.Time) *SeckillActivityCreatedEvent {
	data := map[string]interface{}{
		"activity_id":  activityID,
		"product_id":   productID,
		"name":         name,
		"total_stock":  totalStock,
		"start_time":   startTime,
		"end_time":     endTime,
	}

	return &SeckillActivityCreatedEvent{
		BaseEvent:  NewBaseEvent("seckill.activity.created", activityID, "SeckillActivity", data),
		ActivityID: activityID,
		ProductID:  productID,
		Name:       name,
		TotalStock: totalStock,
		StartTime:  startTime,
		EndTime:    endTime,
	}
}

// SeckillPurchaseEvent 秒杀购买事件
type SeckillPurchaseEvent struct {
	*BaseEvent
	ActivityID     string  `json:"activity_id"`
	ProductID      string  `json:"product_id"`
	UserID         string  `json:"user_id"`
	Quantity       int     `json:"quantity"`
	Price          float64 `json:"price"`
	RemainingStock int     `json:"remaining_stock"`
}

// NewSeckillPurchaseEvent 创建秒杀购买事件
func NewSeckillPurchaseEvent(activityID, productID, userID string, quantity int, price float64, remainingStock int) *SeckillPurchaseEvent {
	data := map[string]interface{}{
		"activity_id":     activityID,
		"product_id":      productID,
		"user_id":         userID,
		"quantity":        quantity,
		"price":           price,
		"remaining_stock": remainingStock,
	}

	return &SeckillPurchaseEvent{
		BaseEvent:      NewBaseEvent("seckill.purchase", activityID, "SeckillActivity", data),
		ActivityID:     activityID,
		ProductID:      productID,
		UserID:         userID,
		Quantity:       quantity,
		Price:          price,
		RemainingStock: remainingStock,
	}
}

// SeckillActivityStartedEvent 秒杀活动开始事件
type SeckillActivityStartedEvent struct {
	*BaseEvent
	ActivityID string `json:"activity_id"`
	ProductID  string `json:"product_id"`
	Name       string `json:"name"`
}

// NewSeckillActivityStartedEvent 创建秒杀活动开始事件
func NewSeckillActivityStartedEvent(activityID, productID, name string) *SeckillActivityStartedEvent {
	data := map[string]interface{}{
		"activity_id": activityID,
		"product_id":  productID,
		"name":        name,
	}

	return &SeckillActivityStartedEvent{
		BaseEvent:  NewBaseEvent("seckill.activity.started", activityID, "SeckillActivity", data),
		ActivityID: activityID,
		ProductID:  productID,
		Name:       name,
	}
}

// SeckillActivityEndedEvent 秒杀活动结束事件
type SeckillActivityEndedEvent struct {
	*BaseEvent
	ActivityID     string `json:"activity_id"`
	ProductID      string `json:"product_id"`
	Name           string `json:"name"`
	RemainingStock int    `json:"remaining_stock"`
}

// NewSeckillActivityEndedEvent 创建秒杀活动结束事件
func NewSeckillActivityEndedEvent(activityID, productID, name string, remainingStock int) *SeckillActivityEndedEvent {
	data := map[string]interface{}{
		"activity_id":     activityID,
		"product_id":      productID,
		"name":            name,
		"remaining_stock": remainingStock,
	}

	return &SeckillActivityEndedEvent{
		BaseEvent:      NewBaseEvent("seckill.activity.ended", activityID, "SeckillActivity", data),
		ActivityID:     activityID,
		ProductID:      productID,
		Name:           name,
		RemainingStock: remainingStock,
	}
}

// SeckillActivityCanceledEvent 秒杀活动取消事件
type SeckillActivityCanceledEvent struct {
	*BaseEvent
	ActivityID string `json:"activity_id"`
	ProductID  string `json:"product_id"`
	Name       string `json:"name"`
}

// NewSeckillActivityCanceledEvent 创建秒杀活动取消事件
func NewSeckillActivityCanceledEvent(activityID, productID, name string) *SeckillActivityCanceledEvent {
	data := map[string]interface{}{
		"activity_id": activityID,
		"product_id":  productID,
		"name":        name,
	}

	return &SeckillActivityCanceledEvent{
		BaseEvent:  NewBaseEvent("seckill.activity.canceled", activityID, "SeckillActivity", data),
		ActivityID: activityID,
		ProductID:  productID,
		Name:       name,
	}
}


