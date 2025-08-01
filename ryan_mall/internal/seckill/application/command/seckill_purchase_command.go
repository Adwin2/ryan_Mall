package command

import (
	"context"
	"fmt"
	"time"

	"ryan-mall-microservices/internal/seckill/domain/repository"
	"ryan-mall-microservices/internal/shared/domain"
	"ryan-mall-microservices/internal/shared/events"
	"ryan-mall-microservices/internal/shared/infrastructure"
)

// SeckillPurchaseCommand 秒杀购买命令
type SeckillPurchaseCommand struct {
	ActivityID string `json:"activity_id" validate:"required"`
	UserID     string `json:"user_id" validate:"required"`
	Quantity   int    `json:"quantity" validate:"required,gt=0"`
}

// SeckillPurchaseResult 秒杀购买结果
type SeckillPurchaseResult struct {
	ActivityID     string  `json:"activity_id"`
	UserID         string  `json:"user_id"`
	Quantity       int     `json:"quantity"`
	Price          float64 `json:"price"`
	TotalAmount    float64 `json:"total_amount"`
	RemainingStock int     `json:"remaining_stock"`
	PurchaseTime   string  `json:"purchase_time"`
	Success        bool    `json:"success"`
	Message        string  `json:"message"`
}

// SeckillPurchaseHandler 秒杀购买命令处理器
type SeckillPurchaseHandler struct {
	seckillRepo    repository.SeckillActivityRepository
	eventPublisher *events.EventPublisher
	lockManager    *infrastructure.LockManager
}

// NewSeckillPurchaseHandler 创建秒杀购买命令处理器
func NewSeckillPurchaseHandler(
	seckillRepo repository.SeckillActivityRepository,
	eventPublisher *events.EventPublisher,
	lockManager *infrastructure.LockManager,
) *SeckillPurchaseHandler {
	return &SeckillPurchaseHandler{
		seckillRepo:    seckillRepo,
		eventPublisher: eventPublisher,
		lockManager:    lockManager,
	}
}

// Handle 处理秒杀购买命令
func (h *SeckillPurchaseHandler) Handle(ctx context.Context, cmd *SeckillPurchaseCommand) (*SeckillPurchaseResult, error) {
	// 使用分布式锁防止超卖
	lockKey := infrastructure.SeckillLockKey(cmd.ActivityID)
	lockExpiration := 5 * time.Second

	var result *SeckillPurchaseResult
	var err error

	lockErr := h.lockManager.WithLock(ctx, lockKey, lockExpiration, func() error {
		result, err = h.handlePurchaseWithLock(ctx, cmd)
		return err
	})

	if lockErr != nil {
		return &SeckillPurchaseResult{
			ActivityID: cmd.ActivityID,
			UserID:     cmd.UserID,
			Quantity:   cmd.Quantity,
			Success:    false,
			Message:    "系统繁忙，请稍后重试",
		}, lockErr
	}

	return result, err
}

// handlePurchaseWithLock 在锁保护下处理购买
func (h *SeckillPurchaseHandler) handlePurchaseWithLock(ctx context.Context, cmd *SeckillPurchaseCommand) (*SeckillPurchaseResult, error) {
	// 查找秒杀活动
	activity, err := h.seckillRepo.FindByID(ctx, domain.ID(cmd.ActivityID))
	if err != nil {
		return nil, domain.NewInternalError("failed to find seckill activity", err)
	}
	if activity == nil {
		return &SeckillPurchaseResult{
			ActivityID: cmd.ActivityID,
			UserID:     cmd.UserID,
			Quantity:   cmd.Quantity,
			Success:    false,
			Message:    "秒杀活动不存在",
		}, nil
	}

	// 检查用户是否已经购买过（防重复购买）
	hasPurchased, err := h.seckillRepo.HasUserPurchased(ctx, domain.ID(cmd.ActivityID), cmd.UserID)
	if err != nil {
		return nil, domain.NewInternalError("failed to check user purchase history", err)
	}
	if hasPurchased {
		return &SeckillPurchaseResult{
			ActivityID: cmd.ActivityID,
			UserID:     cmd.UserID,
			Quantity:   cmd.Quantity,
			Success:    false,
			Message:    "您已经参与过此次秒杀活动",
		}, nil
	}

	// 尝试购买
	err = activity.Purchase(cmd.UserID, cmd.Quantity)
	if err != nil {
		// 根据错误类型返回不同的消息
		var message string
		switch {
		case err.Error() == "seckill activity is not active":
			message = "秒杀活动未开始或已结束"
		case err.Error() == "quantity exceeds max per user limit":
			message = "购买数量超过限制"
		case domain.IsInsufficientStockError(err):
			message = "库存不足，秒杀失败"
		default:
			message = "秒杀失败，请重试"
		}

		return &SeckillPurchaseResult{
			ActivityID: cmd.ActivityID,
			UserID:     cmd.UserID,
			Quantity:   cmd.Quantity,
			Success:    false,
			Message:    message,
		}, nil
	}

	// 保存活动状态
	if err := h.seckillRepo.Update(ctx, activity); err != nil {
		return nil, domain.NewInternalError("failed to update seckill activity", err)
	}

	// 记录用户购买记录
	if err := h.seckillRepo.RecordUserPurchase(ctx, domain.ID(cmd.ActivityID), cmd.UserID, cmd.Quantity); err != nil {
		return nil, domain.NewInternalError("failed to record user purchase", err)
	}

	// 发布领域事件
	if h.eventPublisher != nil {
		if err := h.eventPublisher.PublishEvents(ctx, activity.DomainEvents()...); err != nil {
			// 事件发布失败不应该影响主流程，但应该记录日志
		}
	}

	// 清除领域事件
	activity.ClearDomainEvents()

	// 计算总金额
	price := activity.SeckillPrice().ToYuan()
	totalAmount := price * float64(cmd.Quantity)

	return &SeckillPurchaseResult{
		ActivityID:     cmd.ActivityID,
		UserID:         cmd.UserID,
		Quantity:       cmd.Quantity,
		Price:          price,
		TotalAmount:    totalAmount,
		RemainingStock: activity.RemainingStock(),
		PurchaseTime:   time.Now().Format("2006-01-02T15:04:05Z07:00"),
		Success:        true,
		Message:        "秒杀成功",
	}, nil
}

// FastSeckillPurchaseHandler 快速秒杀购买处理器（使用Redis预扣库存）
type FastSeckillPurchaseHandler struct {
	seckillRepo    repository.SeckillActivityRepository
	eventPublisher *events.EventPublisher
	redisClient    infrastructure.RedisClient
}

// NewFastSeckillPurchaseHandler 创建快速秒杀购买处理器
func NewFastSeckillPurchaseHandler(
	seckillRepo repository.SeckillActivityRepository,
	eventPublisher *events.EventPublisher,
	redisClient infrastructure.RedisClient,
) *FastSeckillPurchaseHandler {
	return &FastSeckillPurchaseHandler{
		seckillRepo:    seckillRepo,
		eventPublisher: eventPublisher,
		redisClient:    redisClient,
	}
}

// Handle 处理快速秒杀购买
func (h *FastSeckillPurchaseHandler) Handle(ctx context.Context, cmd *SeckillPurchaseCommand) (*SeckillPurchaseResult, error) {
	// 1. 先在Redis中预扣库存（原子操作）
	stockKey := fmt.Sprintf("seckill:stock:%s", cmd.ActivityID)
	userKey := fmt.Sprintf("seckill:user:%s:%s", cmd.ActivityID, cmd.UserID)

	// 检查用户是否已购买
	exists, err := h.redisClient.Exists(ctx, userKey)
	if err != nil {
		return nil, err
	}
	if exists {
		return &SeckillPurchaseResult{
			ActivityID: cmd.ActivityID,
			UserID:     cmd.UserID,
			Quantity:   cmd.Quantity,
			Success:    false,
			Message:    "您已经参与过此次秒杀活动",
		}, nil
	}

	// 使用Lua脚本原子性地扣减库存和记录用户购买
	luaScript := `
		local stock_key = KEYS[1]
		local user_key = KEYS[2]
		local quantity = tonumber(ARGV[1])
		local user_id = ARGV[2]
		
		-- 检查库存
		local current_stock = redis.call('GET', stock_key)
		if not current_stock then
			return {-1, "activity not found"}
		end
		
		current_stock = tonumber(current_stock)
		if current_stock < quantity then
			return {-2, "insufficient stock"}
		end
		
		-- 扣减库存
		local remaining = redis.call('DECRBY', stock_key, quantity)
		
		-- 记录用户购买
		redis.call('SETEX', user_key, 86400, quantity)
		
		return {0, remaining}
	`

	result, err := h.redisClient.Eval(ctx, luaScript, []string{stockKey, userKey}, cmd.Quantity, cmd.UserID)
	if err != nil {
		return nil, err
	}

	resultArray := result.([]interface{})
	code := resultArray[0].(int64)
	
	if code == -1 {
		return &SeckillPurchaseResult{
			ActivityID: cmd.ActivityID,
			UserID:     cmd.UserID,
			Quantity:   cmd.Quantity,
			Success:    false,
			Message:    "秒杀活动不存在",
		}, nil
	}
	
	if code == -2 {
		return &SeckillPurchaseResult{
			ActivityID: cmd.ActivityID,
			UserID:     cmd.UserID,
			Quantity:   cmd.Quantity,
			Success:    false,
			Message:    "库存不足，秒杀失败",
		}, nil
	}

	remainingStock := resultArray[1].(int64)

	// 2. 异步处理数据库更新和事件发布
	go h.asyncProcessPurchase(context.Background(), cmd, int(remainingStock))

	// 3. 立即返回成功结果
	return &SeckillPurchaseResult{
		ActivityID:     cmd.ActivityID,
		UserID:         cmd.UserID,
		Quantity:       cmd.Quantity,
		RemainingStock: int(remainingStock),
		PurchaseTime:   time.Now().Format("2006-01-02T15:04:05Z07:00"),
		Success:        true,
		Message:        "秒杀成功",
	}, nil
}

// asyncProcessPurchase 异步处理购买后续流程
func (h *FastSeckillPurchaseHandler) asyncProcessPurchase(ctx context.Context, cmd *SeckillPurchaseCommand, remainingStock int) {
	// 更新数据库
	activity, err := h.seckillRepo.FindByID(ctx, domain.ID(cmd.ActivityID))
	if err != nil {
		// 记录日志
		return
	}

	if activity != nil {
		// 更新活动状态
		activity.Purchase(cmd.UserID, cmd.Quantity)
		h.seckillRepo.Update(ctx, activity)
		h.seckillRepo.RecordUserPurchase(ctx, domain.ID(cmd.ActivityID), cmd.UserID, cmd.Quantity)

		// 发布事件
		if h.eventPublisher != nil {
			h.eventPublisher.PublishEvents(ctx, activity.DomainEvents()...)
		}
	}
}
