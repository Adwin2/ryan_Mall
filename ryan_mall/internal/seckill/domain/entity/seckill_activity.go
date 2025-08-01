package entity

import (
	"strings"
	"time"

	"ryan-mall-microservices/internal/shared/domain"
	"ryan-mall-microservices/internal/shared/events"
)

// SeckillActivity 秒杀活动聚合根
type SeckillActivity struct {
	id             domain.ID
	name           string
	productID      domain.ProductID
	originalPrice  domain.Money
	seckillPrice   domain.Money
	totalStock     int
	remainingStock int
	soldCount      int
	status         SeckillStatus
	startTime      time.Time
	endTime        time.Time
	createdAt      domain.Timestamp
	updatedAt      domain.Timestamp
	domainEvents   []events.Event
}

// SeckillStatus 秒杀状态
type SeckillStatus string

const (
	SeckillStatusPending SeckillStatus = "PENDING" // 待开始
	SeckillStatusActive  SeckillStatus = "ACTIVE"  // 进行中
	SeckillStatusEnded   SeckillStatus = "ENDED"   // 已结束
)

// NewSeckillActivity 创建新的秒杀活动
func NewSeckillActivity(
	name string,
	productID domain.ProductID,
	originalPrice domain.Money,
	seckillPrice domain.Money,
	totalStock int,
	startTime time.Time,
	endTime time.Time,
) (*SeckillActivity, error) {
	// 验证活动名称
	if err := validateActivityName(name); err != nil {
		return nil, err
	}

	// 验证商品ID
	if productID.String() == "" {
		return nil, domain.NewValidationError("product ID cannot be empty")
	}

	// 验证价格
	if err := validatePrices(originalPrice, seckillPrice); err != nil {
		return nil, err
	}

	// 验证库存
	if err := validateStock(totalStock); err != nil {
		return nil, err
	}

	// 验证时间
	if err := validateTimeRange(startTime, endTime); err != nil {
		return nil, err
	}

	// 创建秒杀活动
	activity := &SeckillActivity{
		id:             domain.NewID(),
		name:           strings.TrimSpace(name),
		productID:      productID,
		originalPrice:  originalPrice,
		seckillPrice:   seckillPrice,
		totalStock:     totalStock,
		remainingStock: totalStock,
		soldCount:      0,
		status:         SeckillStatusPending,
		startTime:      startTime,
		endTime:        endTime,
		createdAt:      domain.Now(),
		updatedAt:      domain.Now(),
		domainEvents:   make([]events.Event, 0),
	}

	// 添加秒杀活动创建事件
	event := events.NewSeckillCreatedEvent(
		activity.id.String(),
		activity.productID.String(),
		activity.name,
		activity.seckillPrice.ToYuan(),
		activity.totalStock,
		activity.startTime.Unix(),
		activity.endTime.Unix(),
	)
	activity.addDomainEvent(event)

	return activity, nil
}

// ID 获取活动ID
func (s *SeckillActivity) ID() domain.ID {
	return s.id
}

// Name 获取活动名称
func (s *SeckillActivity) Name() string {
	return s.name
}

// ProductID 获取商品ID
func (s *SeckillActivity) ProductID() domain.ProductID {
	return s.productID
}

// OriginalPrice 获取原价
func (s *SeckillActivity) OriginalPrice() domain.Money {
	return s.originalPrice
}

// SeckillPrice 获取秒杀价
func (s *SeckillActivity) SeckillPrice() domain.Money {
	return s.seckillPrice
}

// TotalStock 获取总库存
func (s *SeckillActivity) TotalStock() int {
	return s.totalStock
}

// RemainingStock 获取剩余库存
func (s *SeckillActivity) RemainingStock() int {
	return s.remainingStock
}

// SoldCount 获取已售数量
func (s *SeckillActivity) SoldCount() int {
	return s.soldCount
}

// Status 获取状态
func (s *SeckillActivity) Status() SeckillStatus {
	return s.status
}

// StartTime 获取开始时间
func (s *SeckillActivity) StartTime() time.Time {
	return s.startTime
}

// EndTime 获取结束时间
func (s *SeckillActivity) EndTime() time.Time {
	return s.endTime
}

// CreatedAt 获取创建时间
func (s *SeckillActivity) CreatedAt() domain.Timestamp {
	return s.createdAt
}

// UpdatedAt 获取更新时间
func (s *SeckillActivity) UpdatedAt() domain.Timestamp {
	return s.updatedAt
}

// Start 启动秒杀活动
func (s *SeckillActivity) Start() error {
	if s.status != SeckillStatusPending {
		return domain.NewValidationError("only pending activities can be started")
	}

	s.status = SeckillStatusActive
	s.updatedAt = domain.Now()

	// 添加秒杀开始事件
	event := events.NewSeckillStartedEvent(
		s.id.String(),
		s.productID.String(),
		s.name,
	)
	s.addDomainEvent(event)

	return nil
}

// End 结束秒杀活动
func (s *SeckillActivity) End() error {
	if s.status != SeckillStatusActive {
		return domain.NewValidationError("only active activities can be ended")
	}

	s.status = SeckillStatusEnded
	s.updatedAt = domain.Now()

	// 添加秒杀结束事件
	event := events.NewSeckillEndedEvent(
		s.id.String(),
		s.productID.String(),
		s.name,
		s.soldCount,
		s.remainingStock,
	)
	s.addDomainEvent(event)

	return nil
}

// ReserveStock 预留库存
func (s *SeckillActivity) ReserveStock(quantity int) error {
	if quantity <= 0 {
		return domain.NewValidationError("quantity must be positive")
	}

	if s.status != SeckillStatusActive {
		return domain.NewValidationError("activity is not active")
	}

	if s.remainingStock < quantity {
		return domain.NewInsufficientStockError(s.productID.String(), quantity, s.remainingStock)
	}

	s.remainingStock -= quantity
	s.soldCount += quantity
	s.updatedAt = domain.Now()

	// 添加库存预留事件
	event := events.NewSeckillStockReservedEvent(
		s.id.String(),
		s.productID.String(),
		quantity,
		s.remainingStock,
	)
	s.addDomainEvent(event)

	// 如果库存售罄，自动结束活动
	if s.remainingStock == 0 {
		s.End()
	}

	return nil
}

// ReleaseStock 释放库存（取消订单时使用）
func (s *SeckillActivity) ReleaseStock(quantity int) error {
	if quantity <= 0 {
		return domain.NewValidationError("quantity must be positive")
	}

	if quantity > s.soldCount {
		return domain.NewValidationError("cannot release more than sold")
	}

	s.remainingStock += quantity
	s.soldCount -= quantity
	s.updatedAt = domain.Now()

	return nil
}

// IsActive 检查活动是否激活（状态为激活且在时间范围内）
func (s *SeckillActivity) IsActive(checkTime time.Time) bool {
	return s.status == SeckillStatusActive &&
		!checkTime.Before(s.startTime) &&
		!checkTime.After(s.endTime)
}

// IsInTimeRange 检查是否在时间范围内
func (s *SeckillActivity) IsInTimeRange(checkTime time.Time) bool {
	return !checkTime.Before(s.startTime) && !checkTime.After(s.endTime)
}

// GetDiscountRate 获取折扣率
func (s *SeckillActivity) GetDiscountRate() float64 {
	if s.originalPrice.IsZero() {
		return 0
	}
	
	discount := s.originalPrice.Subtract(s.seckillPrice)
	return float64(discount.Amount) / float64(s.originalPrice.Amount)
}

// DomainEvents 获取领域事件
func (s *SeckillActivity) DomainEvents() []events.Event {
	return s.domainEvents
}

// ClearDomainEvents 清除领域事件
func (s *SeckillActivity) ClearDomainEvents() {
	s.domainEvents = make([]events.Event, 0)
}

// addDomainEvent 添加领域事件
func (s *SeckillActivity) addDomainEvent(event events.Event) {
	s.domainEvents = append(s.domainEvents, event)
}

// validateActivityName 验证活动名称
func validateActivityName(name string) error {
	name = strings.TrimSpace(name)
	
	if name == "" {
		return domain.NewValidationError("activity name cannot be empty")
	}
	
	if len(name) < 2 {
		return domain.NewValidationError("activity name must be at least 2 characters long")
	}
	
	if len(name) > 100 {
		return domain.NewValidationError("activity name must be at most 100 characters long")
	}
	
	return nil
}

// validatePrices 验证价格
func validatePrices(originalPrice, seckillPrice domain.Money) error {
	if originalPrice.IsNegative() || originalPrice.IsZero() {
		return domain.NewValidationError("original price must be positive")
	}
	
	if seckillPrice.IsNegative() || seckillPrice.IsZero() {
		return domain.NewValidationError("seckill price must be positive")
	}
	
	if seckillPrice.GreaterThan(originalPrice) || seckillPrice.Equal(originalPrice) {
		return domain.NewValidationError("seckill price must be lower than original price")
	}
	
	return nil
}

// validateStock 验证库存
func validateStock(stock int) error {
	if stock <= 0 {
		return domain.NewValidationError("stock must be positive")
	}
	
	return nil
}

// validateTimeRange 验证时间范围
func validateTimeRange(startTime, endTime time.Time) error {
	if endTime.Before(startTime) || endTime.Equal(startTime) {
		return domain.NewValidationError("end time must be after start time")
	}
	
	return nil
}

// ReconstructSeckillActivity 重建秒杀活动实体（用于从持久化存储重建）
func ReconstructSeckillActivity(
	id domain.ID,
	name string,
	productID domain.ProductID,
	originalPrice domain.Money,
	seckillPrice domain.Money,
	totalStock int,
	remainingStock int,
	soldCount int,
	status SeckillStatus,
	startTime time.Time,
	endTime time.Time,
	createdAt domain.Timestamp,
	updatedAt domain.Timestamp,
) *SeckillActivity {
	return &SeckillActivity{
		id:             id,
		name:           name,
		productID:      productID,
		originalPrice:  originalPrice,
		seckillPrice:   seckillPrice,
		totalStock:     totalStock,
		remainingStock: remainingStock,
		soldCount:      soldCount,
		status:         status,
		startTime:      startTime,
		endTime:        endTime,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
		domainEvents:   make([]events.Event, 0),
	}
}
