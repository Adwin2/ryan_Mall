package entity

import (
	"strings"

	"ryan-mall-microservices/internal/shared/domain"
	"ryan-mall-microservices/internal/shared/events"
)

// Product 商品聚合根
type Product struct {
	id          domain.ProductID
	name        string
	description string
	categoryID  string
	price       domain.Money
	stock       int
	salesCount  int
	status      ProductStatus
	createdAt   domain.Timestamp
	updatedAt   domain.Timestamp
	domainEvents []events.Event
}

// ProductStatus 商品状态
type ProductStatus int

const (
	ProductStatusAvailable   ProductStatus = 1
	ProductStatusUnavailable ProductStatus = 0
)

// NewProduct 创建新商品
func NewProduct(name, description, categoryID string, price domain.Money, stock int) (*Product, error) {
	// 验证商品名称
	if err := validateProductName(name); err != nil {
		return nil, err
	}

	// 验证分类ID
	if err := validateCategoryID(categoryID); err != nil {
		return nil, err
	}

	// 验证价格
	if err := validatePrice(price); err != nil {
		return nil, err
	}

	// 验证库存
	if err := validateStock(stock); err != nil {
		return nil, err
	}

	// 创建商品
	product := &Product{
		id:          domain.NewProductID(),
		name:        strings.TrimSpace(name),
		description: strings.TrimSpace(description),
		categoryID:  categoryID,
		price:       price,
		stock:       stock,
		salesCount:  0,
		status:      ProductStatusAvailable,
		createdAt:   domain.Now(),
		updatedAt:   domain.Now(),
		domainEvents: make([]events.Event, 0),
	}

	// 添加商品创建事件
	event := events.NewProductCreatedEvent(
		product.id.String(),
		product.name,
		product.categoryID,
		product.price.ToYuan(),
	)
	product.addDomainEvent(event)

	return product, nil
}

// ID 获取商品ID
func (p *Product) ID() domain.ProductID {
	return p.id
}

// Name 获取商品名称
func (p *Product) Name() string {
	return p.name
}

// Description 获取商品描述
func (p *Product) Description() string {
	return p.description
}

// CategoryID 获取分类ID
func (p *Product) CategoryID() string {
	return p.categoryID
}

// Price 获取价格
func (p *Product) Price() domain.Money {
	return p.price
}

// Stock 获取库存
func (p *Product) Stock() int {
	return p.stock
}

// SalesCount 获取销量
func (p *Product) SalesCount() int {
	return p.salesCount
}

// IsAvailable 是否可用
func (p *Product) IsAvailable() bool {
	return p.status == ProductStatusAvailable
}

// CreatedAt 获取创建时间
func (p *Product) CreatedAt() domain.Timestamp {
	return p.createdAt
}

// UpdatedAt 获取更新时间
func (p *Product) UpdatedAt() domain.Timestamp {
	return p.updatedAt
}

// UpdatePrice 更新价格
func (p *Product) UpdatePrice(newPrice domain.Money) error {
	if err := validatePrice(newPrice); err != nil {
		return err
	}

	oldPrice := p.price
	p.price = newPrice
	p.updatedAt = domain.Now()

	// 添加价格更新事件
	event := events.NewPriceUpdatedEvent(
		p.id.String(),
		oldPrice.ToYuan(),
		newPrice.ToYuan(),
	)
	p.addDomainEvent(event)

	return nil
}

// UpdateStock 更新库存
func (p *Product) UpdateStock(newStock int) error {
	if err := validateStock(newStock); err != nil {
		return err
	}

	oldStock := p.stock
	p.stock = newStock
	p.updatedAt = domain.Now()

	// 添加库存更新事件
	event := events.NewStockUpdatedEvent(
		p.id.String(),
		oldStock,
		newStock,
	)
	p.addDomainEvent(event)

	return nil
}

// ReserveStock 预留库存
func (p *Product) ReserveStock(quantity int, orderID string) error {
	if quantity <= 0 {
		return domain.NewValidationError("quantity must be positive")
	}

	if p.stock < quantity {
		return domain.NewInsufficientStockError(p.id.String(), quantity, p.stock)
	}

	p.stock -= quantity
	p.updatedAt = domain.Now()

	// 添加库存预留事件
	event := events.NewStockReservedEvent(
		p.id.String(),
		quantity,
		orderID,
		p.stock,
	)
	p.addDomainEvent(event)

	return nil
}

// ReleaseStock 释放库存
func (p *Product) ReleaseStock(quantity int, orderID string) error {
	if quantity <= 0 {
		return domain.NewValidationError("quantity must be positive")
	}

	p.stock += quantity
	p.updatedAt = domain.Now()

	// 添加库存释放事件
	event := events.NewStockReleasedEvent(
		p.id.String(),
		quantity,
		orderID,
		p.stock,
	)
	p.addDomainEvent(event)

	return nil
}

// IncreaseSales 增加销量
func (p *Product) IncreaseSales(quantity int) error {
	if quantity <= 0 {
		return domain.NewValidationError("quantity must be positive")
	}

	p.salesCount += quantity
	p.updatedAt = domain.Now()

	return nil
}

// SetAvailable 设置为可用
func (p *Product) SetAvailable() {
	p.status = ProductStatusAvailable
	p.updatedAt = domain.Now()
}

// SetUnavailable 设置为不可用
func (p *Product) SetUnavailable() {
	p.status = ProductStatusUnavailable
	p.updatedAt = domain.Now()
}

// UpdateInfo 更新商品信息
func (p *Product) UpdateInfo(name, description string) error {
	if err := validateProductName(name); err != nil {
		return err
	}

	p.name = strings.TrimSpace(name)
	p.description = strings.TrimSpace(description)
	p.updatedAt = domain.Now()

	return nil
}

// DomainEvents 获取领域事件
func (p *Product) DomainEvents() []events.Event {
	return p.domainEvents
}

// ClearDomainEvents 清除领域事件
func (p *Product) ClearDomainEvents() {
	p.domainEvents = make([]events.Event, 0)
}

// addDomainEvent 添加领域事件
func (p *Product) addDomainEvent(event events.Event) {
	p.domainEvents = append(p.domainEvents, event)
}

// validateProductName 验证商品名称
func validateProductName(name string) error {
	name = strings.TrimSpace(name)
	
	if name == "" {
		return domain.NewValidationError("product name cannot be empty")
	}
	
	if len(name) < 2 {
		return domain.NewValidationError("product name must be at least 2 characters long")
	}
	
	if len(name) > 255 {
		return domain.NewValidationError("product name must be at most 255 characters long")
	}
	
	return nil
}

// validateCategoryID 验证分类ID
func validateCategoryID(categoryID string) error {
	if categoryID == "" {
		return domain.NewValidationError("category ID cannot be empty")
	}
	
	return nil
}

// validatePrice 验证价格
func validatePrice(price domain.Money) error {
	if price.IsNegative() {
		return domain.NewValidationError("price cannot be negative")
	}
	
	if price.IsZero() {
		return domain.NewValidationError("price cannot be zero")
	}
	
	return nil
}

// validateStock 验证库存
func validateStock(stock int) error {
	if stock < 0 {
		return domain.NewValidationError("stock cannot be negative")
	}

	return nil
}

// ReconstructProduct 重建商品实体（用于从数据库恢复）
func ReconstructProduct(
	id domain.ProductID,
	name string,
	description string,
	categoryID string,
	price domain.Money,
	stock int,
	salesCount int,
	isAvailable bool,
	createdAt domain.Timestamp,
	updatedAt domain.Timestamp,
) *Product {
	status := ProductStatusUnavailable
	if isAvailable {
		status = ProductStatusAvailable
	}

	return &Product{
		id:          id,
		name:        name,
		description: description,
		categoryID:  categoryID,
		price:       price,
		stock:       stock,
		salesCount:  salesCount,
		status:      status,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
		domainEvents: make([]events.Event, 0),
	}
}
