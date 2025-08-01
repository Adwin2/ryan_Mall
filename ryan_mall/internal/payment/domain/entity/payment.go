package entity

import (
	"ryan-mall-microservices/internal/shared/domain"
	"ryan-mall-microservices/internal/shared/events"
)

// Payment 支付聚合根
type Payment struct {
	id            domain.ID
	userID        domain.UserID
	orderID       domain.OrderID
	amount        domain.Money
	method        PaymentMethod
	status        PaymentStatus
	transactionID string
	refundID      string
	refundAmount  domain.Money
	refundReason  string
	failureReason string
	createdAt     domain.Timestamp
	updatedAt     domain.Timestamp
	domainEvents  []events.Event
}

// PaymentMethod 支付方式
type PaymentMethod string

const (
	PaymentMethodAlipay   PaymentMethod = "ALIPAY"   // 支付宝
	PaymentMethodWechat   PaymentMethod = "WECHAT"   // 微信支付
	PaymentMethodUnionPay PaymentMethod = "UNIONPAY" // 银联
	PaymentMethodCredit   PaymentMethod = "CREDIT"   // 信用卡
	PaymentMethodDebit    PaymentMethod = "DEBIT"    // 借记卡
)

// PaymentStatus 支付状态
type PaymentStatus string

const (
	PaymentStatusPending    PaymentStatus = "PENDING"    // 待支付
	PaymentStatusProcessing PaymentStatus = "PROCESSING" // 处理中
	PaymentStatusCompleted  PaymentStatus = "COMPLETED"  // 已完成
	PaymentStatusFailed     PaymentStatus = "FAILED"     // 失败
	PaymentStatusCancelled  PaymentStatus = "CANCELLED"  // 已取消
	PaymentStatusRefunded   PaymentStatus = "REFUNDED"   // 已退款
)

// NewPayment 创建新支付
func NewPayment(
	userID domain.UserID,
	orderID domain.OrderID,
	amount domain.Money,
	method PaymentMethod,
) (*Payment, error) {
	// 验证用户ID
	if userID.String() == "" {
		return nil, domain.NewValidationError("user ID cannot be empty")
	}

	// 验证订单ID
	if orderID.String() == "" {
		return nil, domain.NewValidationError("order ID cannot be empty")
	}

	// 验证金额
	if err := validateAmount(amount); err != nil {
		return nil, err
	}

	// 验证支付方式
	if err := validatePaymentMethod(method); err != nil {
		return nil, err
	}

	// 创建支付
	payment := &Payment{
		id:           domain.NewID(),
		userID:       userID,
		orderID:      orderID,
		amount:       amount,
		method:       method,
		status:       PaymentStatusPending,
		refundAmount: domain.NewMoney(0, amount.Currency),
		createdAt:    domain.Now(),
		updatedAt:    domain.Now(),
		domainEvents: make([]events.Event, 0),
	}

	// 添加支付创建事件
	event := events.NewPaymentCreatedEvent(
		payment.id.String(),
		payment.userID.String(),
		payment.orderID.String(),
		payment.amount.ToYuan(),
		string(payment.method),
	)
	payment.addDomainEvent(event)

	return payment, nil
}

// ID 获取支付ID
func (p *Payment) ID() domain.ID {
	return p.id
}

// UserID 获取用户ID
func (p *Payment) UserID() domain.UserID {
	return p.userID
}

// OrderID 获取订单ID
func (p *Payment) OrderID() domain.OrderID {
	return p.orderID
}

// Amount 获取支付金额
func (p *Payment) Amount() domain.Money {
	return p.amount
}

// Method 获取支付方式
func (p *Payment) Method() PaymentMethod {
	return p.method
}

// Status 获取支付状态
func (p *Payment) Status() PaymentStatus {
	return p.status
}

// TransactionID 获取交易ID
func (p *Payment) TransactionID() string {
	return p.transactionID
}

// RefundID 获取退款ID
func (p *Payment) RefundID() string {
	return p.refundID
}

// RefundAmount 获取退款金额
func (p *Payment) RefundAmount() domain.Money {
	return p.refundAmount
}

// RefundReason 获取退款原因
func (p *Payment) RefundReason() string {
	return p.refundReason
}

// FailureReason 获取失败原因
func (p *Payment) FailureReason() string {
	return p.failureReason
}

// CreatedAt 获取创建时间
func (p *Payment) CreatedAt() domain.Timestamp {
	return p.createdAt
}

// UpdatedAt 获取更新时间
func (p *Payment) UpdatedAt() domain.Timestamp {
	return p.updatedAt
}

// Process 处理支付
func (p *Payment) Process(transactionID string) error {
	if p.status != PaymentStatusPending {
		return domain.NewValidationError("only pending payments can be processed")
	}

	if transactionID == "" {
		return domain.NewValidationError("transaction ID cannot be empty")
	}

	p.status = PaymentStatusProcessing
	p.transactionID = transactionID
	p.updatedAt = domain.Now()

	// 添加支付处理事件
	event := events.NewPaymentProcessingEvent(
		p.id.String(),
		p.userID.String(),
		p.orderID.String(),
		transactionID,
	)
	p.addDomainEvent(event)

	return nil
}

// Complete 完成支付
func (p *Payment) Complete() error {
	if p.status != PaymentStatusProcessing {
		return domain.NewValidationError("only processing payments can be completed")
	}

	p.status = PaymentStatusCompleted
	p.updatedAt = domain.Now()

	// 添加支付完成事件
	event := events.NewPaymentCompletedEvent(
		p.id.String(),
		p.userID.String(),
		p.orderID.String(),
		p.amount.ToYuan(),
		p.transactionID,
	)
	p.addDomainEvent(event)

	return nil
}

// Fail 支付失败
func (p *Payment) Fail(reason string) error {
	if p.status != PaymentStatusProcessing {
		return domain.NewValidationError("only processing payments can be failed")
	}

	if reason == "" {
		return domain.NewValidationError("failure reason cannot be empty")
	}

	p.status = PaymentStatusFailed
	p.failureReason = reason
	p.updatedAt = domain.Now()

	// 添加支付失败事件
	event := events.NewPaymentFailedEvent(
		p.id.String(),
		p.userID.String(),
		p.orderID.String(),
		reason,
	)
	p.addDomainEvent(event)

	return nil
}

// Cancel 取消支付
func (p *Payment) Cancel() error {
	if p.status == PaymentStatusCompleted || p.status == PaymentStatusFailed || p.status == PaymentStatusRefunded {
		return domain.NewValidationError("cannot cancel completed, failed or refunded payments")
	}

	p.status = PaymentStatusCancelled
	p.updatedAt = domain.Now()

	// 添加支付取消事件
	event := events.NewPaymentCancelledEvent(
		p.id.String(),
		p.userID.String(),
		p.orderID.String(),
	)
	p.addDomainEvent(event)

	return nil
}

// Refund 退款
func (p *Payment) Refund(refundID string, refundAmount domain.Money, reason string) error {
	if p.status != PaymentStatusCompleted {
		return domain.NewValidationError("only completed payments can be refunded")
	}

	if refundID == "" {
		return domain.NewValidationError("refund ID cannot be empty")
	}

	if reason == "" {
		return domain.NewValidationError("refund reason cannot be empty")
	}

	if err := validateAmount(refundAmount); err != nil {
		return err
	}

	if refundAmount.GreaterThan(p.amount) {
		return domain.NewValidationError("refund amount cannot exceed payment amount")
	}

	p.status = PaymentStatusRefunded
	p.refundID = refundID
	p.refundAmount = refundAmount
	p.refundReason = reason
	p.updatedAt = domain.Now()

	// 添加退款事件
	event := events.NewPaymentRefundedEvent(
		p.id.String(),
		p.userID.String(),
		p.orderID.String(),
		refundID,
		refundAmount.ToYuan(),
		reason,
	)
	p.addDomainEvent(event)

	return nil
}

// IsCompleted 是否已完成
func (p *Payment) IsCompleted() bool {
	return p.status == PaymentStatusCompleted
}

// IsFailed 是否失败
func (p *Payment) IsFailed() bool {
	return p.status == PaymentStatusFailed
}

// IsCancelled 是否已取消
func (p *Payment) IsCancelled() bool {
	return p.status == PaymentStatusCancelled
}

// IsRefunded 是否已退款
func (p *Payment) IsRefunded() bool {
	return p.status == PaymentStatusRefunded
}

// CanBeProcessed 是否可以处理
func (p *Payment) CanBeProcessed() bool {
	return p.status == PaymentStatusPending
}

// CanBeCancelled 是否可以取消
func (p *Payment) CanBeCancelled() bool {
	return p.status == PaymentStatusPending || p.status == PaymentStatusProcessing
}

// CanBeRefunded 是否可以退款
func (p *Payment) CanBeRefunded() bool {
	return p.status == PaymentStatusCompleted
}

// DomainEvents 获取领域事件
func (p *Payment) DomainEvents() []events.Event {
	return p.domainEvents
}

// ClearDomainEvents 清除领域事件
func (p *Payment) ClearDomainEvents() {
	p.domainEvents = make([]events.Event, 0)
}

// addDomainEvent 添加领域事件
func (p *Payment) addDomainEvent(event events.Event) {
	p.domainEvents = append(p.domainEvents, event)
}

// validateAmount 验证金额
func validateAmount(amount domain.Money) error {
	if amount.IsNegative() {
		return domain.NewValidationError("amount cannot be negative")
	}
	
	if amount.IsZero() {
		return domain.NewValidationError("amount cannot be zero")
	}
	
	return nil
}

// validatePaymentMethod 验证支付方式
func validatePaymentMethod(method PaymentMethod) error {
	switch method {
	case PaymentMethodAlipay, PaymentMethodWechat, PaymentMethodUnionPay, PaymentMethodCredit, PaymentMethodDebit:
		return nil
	default:
		return domain.NewValidationError("invalid payment method")
	}
}

// ReconstructPayment 重建支付实体（用于从持久化存储重建）
func ReconstructPayment(
	id domain.ID,
	userID domain.UserID,
	orderID domain.OrderID,
	amount domain.Money,
	method PaymentMethod,
	status PaymentStatus,
	transactionID string,
	refundID string,
	refundAmount domain.Money,
	refundReason string,
	failureReason string,
	createdAt domain.Timestamp,
	updatedAt domain.Timestamp,
) *Payment {
	return &Payment{
		id:            id,
		userID:        userID,
		orderID:       orderID,
		amount:        amount,
		method:        method,
		status:        status,
		transactionID: transactionID,
		refundID:      refundID,
		refundAmount:  refundAmount,
		refundReason:  refundReason,
		failureReason: failureReason,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
		domainEvents:  make([]events.Event, 0),
	}
}
