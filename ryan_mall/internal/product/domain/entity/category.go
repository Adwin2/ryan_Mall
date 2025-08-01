package entity

import (
	"strings"

	"ryan-mall-microservices/internal/shared/domain"
	"ryan-mall-microservices/internal/shared/events"
)

// Category 商品分类聚合根
type Category struct {
	id          domain.ID
	name        string
	parentID    *domain.ID
	level       int
	sortOrder   int
	status      CategoryStatus
	createdAt   domain.Timestamp
	updatedAt   domain.Timestamp
	domainEvents []events.Event
}

// CategoryStatus 分类状态
type CategoryStatus int

const (
	CategoryStatusActive   CategoryStatus = 1
	CategoryStatusInactive CategoryStatus = 0
)

// NewCategory 创建新分类
func NewCategory(name string, parentID *domain.ID) (*Category, error) {
	// 验证分类名称
	if err := validateCategoryName(name); err != nil {
		return nil, err
	}

	// 计算层级
	level := 1
	if parentID != nil {
		level = 2 // 简化处理，实际项目中需要查询父分类的层级
	}

	// 创建分类
	category := &Category{
		id:          domain.NewID(),
		name:        strings.TrimSpace(name),
		parentID:    parentID,
		level:       level,
		sortOrder:   0,
		status:      CategoryStatusActive,
		createdAt:   domain.Now(),
		updatedAt:   domain.Now(),
		domainEvents: make([]events.Event, 0),
	}

	// 添加分类创建事件
	event := events.NewCategoryCreatedEvent(
		category.id.String(),
		category.name,
		category.level,
	)
	category.addDomainEvent(event)

	return category, nil
}

// ID 获取分类ID
func (c *Category) ID() domain.ID {
	return c.id
}

// Name 获取分类名称
func (c *Category) Name() string {
	return c.name
}

// ParentID 获取父分类ID
func (c *Category) ParentID() *domain.ID {
	return c.parentID
}

// Level 获取层级
func (c *Category) Level() int {
	return c.level
}

// SortOrder 获取排序
func (c *Category) SortOrder() int {
	return c.sortOrder
}

// IsActive 是否激活
func (c *Category) IsActive() bool {
	return c.status == CategoryStatusActive
}

// CreatedAt 获取创建时间
func (c *Category) CreatedAt() domain.Timestamp {
	return c.createdAt
}

// UpdatedAt 获取更新时间
func (c *Category) UpdatedAt() domain.Timestamp {
	return c.updatedAt
}

// UpdateName 更新分类名称
func (c *Category) UpdateName(name string) error {
	if err := validateCategoryName(name); err != nil {
		return err
	}

	c.name = strings.TrimSpace(name)
	c.updatedAt = domain.Now()

	return nil
}

// UpdateSortOrder 更新排序
func (c *Category) UpdateSortOrder(sortOrder int) {
	c.sortOrder = sortOrder
	c.updatedAt = domain.Now()
}

// Activate 激活分类
func (c *Category) Activate() {
	c.status = CategoryStatusActive
	c.updatedAt = domain.Now()
}

// Deactivate 停用分类
func (c *Category) Deactivate() {
	c.status = CategoryStatusInactive
	c.updatedAt = domain.Now()
}

// IsRootCategory 是否为根分类
func (c *Category) IsRootCategory() bool {
	return c.parentID == nil
}

// DomainEvents 获取领域事件
func (c *Category) DomainEvents() []events.Event {
	return c.domainEvents
}

// ClearDomainEvents 清除领域事件
func (c *Category) ClearDomainEvents() {
	c.domainEvents = make([]events.Event, 0)
}

// addDomainEvent 添加领域事件
func (c *Category) addDomainEvent(event events.Event) {
	c.domainEvents = append(c.domainEvents, event)
}

// validateCategoryName 验证分类名称
func validateCategoryName(name string) error {
	name = strings.TrimSpace(name)
	
	if name == "" {
		return domain.NewValidationError("category name cannot be empty")
	}
	
	if len(name) < 2 {
		return domain.NewValidationError("category name must be at least 2 characters long")
	}
	
	if len(name) > 100 {
		return domain.NewValidationError("category name must be at most 100 characters long")
	}
	
	return nil
}

// ReconstructCategory 重建分类实体（用于从持久化存储重建）
func ReconstructCategory(
	id domain.ID,
	name string,
	parentID *domain.ID,
	level int,
	sortOrder int,
	isActive bool,
	createdAt domain.Timestamp,
	updatedAt domain.Timestamp,
) *Category {
	status := CategoryStatusInactive
	if isActive {
		status = CategoryStatusActive
	}

	return &Category{
		id:          id,
		name:        name,
		parentID:    parentID,
		level:       level,
		sortOrder:   sortOrder,
		status:      status,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
		domainEvents: make([]events.Event, 0),
	}
}
