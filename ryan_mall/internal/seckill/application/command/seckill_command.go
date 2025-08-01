package command

import (
	"context"
	"time"

	"ryan-mall-microservices/internal/seckill/domain/entity"
	"ryan-mall-microservices/internal/seckill/domain/repository"
	"ryan-mall-microservices/internal/seckill/domain/service"
	"ryan-mall-microservices/internal/shared/domain"
	"ryan-mall-microservices/internal/shared/events"
)

// CreateSeckillActivityCommand 创建秒杀活动命令
type CreateSeckillActivityCommand struct {
	Name          string    `json:"name" validate:"required,min=2,max=100"`
	ProductID     string    `json:"product_id" validate:"required"`
	OriginalPrice float64   `json:"original_price" validate:"required,gt=0"`
	SeckillPrice  float64   `json:"seckill_price" validate:"required,gt=0"`
	TotalStock    int       `json:"total_stock" validate:"required,gt=0"`
	StartTime     time.Time `json:"start_time" validate:"required"`
	EndTime       time.Time `json:"end_time" validate:"required"`
}

// CreateSeckillActivityResult 创建秒杀活动结果
type CreateSeckillActivityResult struct {
	ActivityID    string  `json:"activity_id"`
	Name          string  `json:"name"`
	ProductID     string  `json:"product_id"`
	OriginalPrice float64 `json:"original_price"`
	SeckillPrice  float64 `json:"seckill_price"`
	TotalStock    int     `json:"total_stock"`
	StartTime     int64   `json:"start_time"`
	EndTime       int64   `json:"end_time"`
	Status        string  `json:"status"`
}

// CreateSeckillActivityHandler 创建秒杀活动命令处理器
type CreateSeckillActivityHandler struct {
	activityRepo   repository.SeckillActivityRepository
	domainService  *service.SeckillDomainService
	eventPublisher *events.EventPublisher
}

// NewCreateSeckillActivityHandler 创建秒杀活动命令处理器
func NewCreateSeckillActivityHandler(
	activityRepo repository.SeckillActivityRepository,
	domainService *service.SeckillDomainService,
	eventPublisher *events.EventPublisher,
) *CreateSeckillActivityHandler {
	return &CreateSeckillActivityHandler{
		activityRepo:   activityRepo,
		domainService:  domainService,
		eventPublisher: eventPublisher,
	}
}

// Handle 处理创建秒杀活动命令
func (h *CreateSeckillActivityHandler) Handle(ctx context.Context, cmd *CreateSeckillActivityCommand) (*CreateSeckillActivityResult, error) {
	// 创建价格值对象
	originalPrice := domain.NewMoneyFromYuan(cmd.OriginalPrice, "CNY")
	seckillPrice := domain.NewMoneyFromYuan(cmd.SeckillPrice, "CNY")

	// 创建秒杀活动实体
	activity, err := entity.NewSeckillActivity(
		cmd.Name,
		domain.ProductID(cmd.ProductID),
		originalPrice,
		seckillPrice,
		cmd.TotalStock,
		cmd.StartTime,
		cmd.EndTime,
	)
	if err != nil {
		return nil, err
	}

	// 保存秒杀活动
	if err := h.activityRepo.Save(ctx, activity); err != nil {
		return nil, domain.NewInternalError("failed to save seckill activity", err)
	}

	// 初始化Redis库存
	if err := h.domainService.InitializeActivityStock(ctx, activity); err != nil {
		return nil, domain.NewInternalError("failed to initialize activity stock", err)
	}

	// 发布领域事件
	if h.eventPublisher != nil {
		if err := h.eventPublisher.PublishEvents(ctx, activity.DomainEvents()...); err != nil {
			// 事件发布失败不应该影响主流程，但应该记录日志
		}
	}

	// 清除领域事件
	activity.ClearDomainEvents()

	return &CreateSeckillActivityResult{
		ActivityID:    activity.ID().String(),
		Name:          activity.Name(),
		ProductID:     activity.ProductID().String(),
		OriginalPrice: activity.OriginalPrice().ToYuan(),
		SeckillPrice:  activity.SeckillPrice().ToYuan(),
		TotalStock:    activity.TotalStock(),
		StartTime:     activity.StartTime().Unix(),
		EndTime:       activity.EndTime().Unix(),
		Status:        string(activity.Status()),
	}, nil
}

// StartSeckillActivityCommand 启动秒杀活动命令
type StartSeckillActivityCommand struct {
	ActivityID string `json:"activity_id" validate:"required"`
}

// StartSeckillActivityHandler 启动秒杀活动命令处理器
type StartSeckillActivityHandler struct {
	activityRepo   repository.SeckillActivityRepository
	eventPublisher *events.EventPublisher
}

// NewStartSeckillActivityHandler 创建启动秒杀活动命令处理器
func NewStartSeckillActivityHandler(
	activityRepo repository.SeckillActivityRepository,
	eventPublisher *events.EventPublisher,
) *StartSeckillActivityHandler {
	return &StartSeckillActivityHandler{
		activityRepo:   activityRepo,
		eventPublisher: eventPublisher,
	}
}

// Handle 处理启动秒杀活动命令
func (h *StartSeckillActivityHandler) Handle(ctx context.Context, cmd *StartSeckillActivityCommand) error {
	// 查找秒杀活动
	activity, err := h.activityRepo.FindByID(ctx, domain.ID(cmd.ActivityID))
	if err != nil {
		return domain.NewInternalError("failed to find seckill activity", err)
	}
	if activity == nil {
		return domain.NewNotFoundError("seckill activity", cmd.ActivityID)
	}

	// 启动活动
	if err := activity.Start(); err != nil {
		return err
	}

	// 更新活动
	if err := h.activityRepo.Update(ctx, activity); err != nil {
		return domain.NewInternalError("failed to update seckill activity", err)
	}

	// 发布领域事件
	if h.eventPublisher != nil {
		if err := h.eventPublisher.PublishEvents(ctx, activity.DomainEvents()...); err != nil {
			// 事件发布失败不应该影响主流程，但应该记录日志
		}
	}

	// 清除领域事件
	activity.ClearDomainEvents()

	return nil
}

// ParticipateInSeckillCommand 参与秒杀命令
type ParticipateInSeckillCommand struct {
	UserID     string `json:"user_id" validate:"required"`
	ActivityID string `json:"activity_id" validate:"required"`
	Quantity   int    `json:"quantity" validate:"required,min=1"`
}

// ParticipateInSeckillResult 参与秒杀结果
type ParticipateInSeckillResult struct {
	OrderID     string  `json:"order_id"`
	ActivityID  string  `json:"activity_id"`
	ProductID   string  `json:"product_id"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
	TotalAmount float64 `json:"total_amount"`
	Status      string  `json:"status"`
}

// ParticipateInSeckillHandler 参与秒杀命令处理器
type ParticipateInSeckillHandler struct {
	domainService  *service.SeckillDomainService
	eventPublisher *events.EventPublisher
}

// NewParticipateInSeckillHandler 创建参与秒杀命令处理器
func NewParticipateInSeckillHandler(
	domainService *service.SeckillDomainService,
	eventPublisher *events.EventPublisher,
) *ParticipateInSeckillHandler {
	return &ParticipateInSeckillHandler{
		domainService:  domainService,
		eventPublisher: eventPublisher,
	}
}

// Handle 处理参与秒杀命令
func (h *ParticipateInSeckillHandler) Handle(ctx context.Context, cmd *ParticipateInSeckillCommand) (*ParticipateInSeckillResult, error) {
	// 处理秒杀订单（包含完整的验证和库存扣减逻辑）
	order, err := h.domainService.ProcessSeckillOrder(
		ctx,
		domain.UserID(cmd.UserID),
		domain.ID(cmd.ActivityID),
		cmd.Quantity,
	)
	if err != nil {
		return nil, err
	}

	// 发布领域事件
	if h.eventPublisher != nil {
		if err := h.eventPublisher.PublishEvents(ctx, order.DomainEvents()...); err != nil {
			// 事件发布失败不应该影响主流程，但应该记录日志
		}
	}

	// 清除领域事件
	order.ClearDomainEvents()

	return &ParticipateInSeckillResult{
		OrderID:     order.ID().String(),
		ActivityID:  order.ActivityID().String(),
		ProductID:   order.ProductID().String(),
		Quantity:    order.Quantity(),
		Price:       order.Price().ToYuan(),
		TotalAmount: order.TotalAmount().ToYuan(),
		Status:      string(order.Status()),
	}, nil
}

// PaySeckillOrderCommand 支付秒杀订单命令
type PaySeckillOrderCommand struct {
	OrderID string `json:"order_id" validate:"required"`
	UserID  string `json:"user_id" validate:"required"`
}

// PaySeckillOrderHandler 支付秒杀订单命令处理器
type PaySeckillOrderHandler struct {
	orderRepo      repository.SeckillOrderRepository
	eventPublisher *events.EventPublisher
}

// NewPaySeckillOrderHandler 创建支付秒杀订单命令处理器
func NewPaySeckillOrderHandler(
	orderRepo repository.SeckillOrderRepository,
	eventPublisher *events.EventPublisher,
) *PaySeckillOrderHandler {
	return &PaySeckillOrderHandler{
		orderRepo:      orderRepo,
		eventPublisher: eventPublisher,
	}
}

// Handle 处理支付秒杀订单命令
func (h *PaySeckillOrderHandler) Handle(ctx context.Context, cmd *PaySeckillOrderCommand) error {
	// 查找订单
	order, err := h.orderRepo.FindByID(ctx, domain.OrderID(cmd.OrderID))
	if err != nil {
		return domain.NewInternalError("failed to find seckill order", err)
	}
	if order == nil {
		return domain.NewNotFoundError("seckill order", cmd.OrderID)
	}

	// 验证订单所有者
	if order.UserID().String() != cmd.UserID {
		return domain.NewForbiddenError("order does not belong to user")
	}

	// 支付订单
	if err := order.Pay(); err != nil {
		return err
	}

	// 更新订单
	if err := h.orderRepo.Update(ctx, order); err != nil {
		return domain.NewInternalError("failed to update seckill order", err)
	}

	// 发布领域事件
	if h.eventPublisher != nil {
		if err := h.eventPublisher.PublishEvents(ctx, order.DomainEvents()...); err != nil {
			// 事件发布失败不应该影响主流程，但应该记录日志
		}
	}

	// 清除领域事件
	order.ClearDomainEvents()

	return nil
}

// CancelSeckillOrderCommand 取消秒杀订单命令
type CancelSeckillOrderCommand struct {
	OrderID string `json:"order_id" validate:"required"`
	UserID  string `json:"user_id" validate:"required"`
}

// CancelSeckillOrderHandler 取消秒杀订单命令处理器
type CancelSeckillOrderHandler struct {
	orderRepo     repository.SeckillOrderRepository
	domainService *service.SeckillDomainService
	eventPublisher *events.EventPublisher
}

// NewCancelSeckillOrderHandler 创建取消秒杀订单命令处理器
func NewCancelSeckillOrderHandler(
	orderRepo repository.SeckillOrderRepository,
	domainService *service.SeckillDomainService,
	eventPublisher *events.EventPublisher,
) *CancelSeckillOrderHandler {
	return &CancelSeckillOrderHandler{
		orderRepo:      orderRepo,
		domainService:  domainService,
		eventPublisher: eventPublisher,
	}
}

// Handle 处理取消秒杀订单命令
func (h *CancelSeckillOrderHandler) Handle(ctx context.Context, cmd *CancelSeckillOrderCommand) error {
	// 查找订单
	order, err := h.orderRepo.FindByID(ctx, domain.OrderID(cmd.OrderID))
	if err != nil {
		return domain.NewInternalError("failed to find seckill order", err)
	}
	if order == nil {
		return domain.NewNotFoundError("seckill order", cmd.OrderID)
	}

	// 验证订单所有者
	if order.UserID().String() != cmd.UserID {
		return domain.NewForbiddenError("order does not belong to user")
	}

	// 取消订单
	if err := order.Cancel(); err != nil {
		return err
	}

	// 恢复库存
	if err := h.domainService.RestoreStock(ctx, order.ActivityID(), order.Quantity()); err != nil {
		// 记录日志但不影响主流程
	}

	// 更新订单
	if err := h.orderRepo.Update(ctx, order); err != nil {
		return domain.NewInternalError("failed to update seckill order", err)
	}

	// 发布领域事件
	if h.eventPublisher != nil {
		if err := h.eventPublisher.PublishEvents(ctx, order.DomainEvents()...); err != nil {
			// 事件发布失败不应该影响主流程，但应该记录日志
		}
	}

	// 清除领域事件
	order.ClearDomainEvents()

	return nil
}
