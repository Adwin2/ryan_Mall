package service

import (
	"context"
	"fmt"
	"time"

	"ryan-mall-microservices/internal/seckill/domain/entity"
	"ryan-mall-microservices/internal/seckill/domain/repository"
	"ryan-mall-microservices/internal/shared/domain"
	"ryan-mall-microservices/pkg/ratelimiter"

	"github.com/go-redis/redis/v8"
)

// SeckillDomainService 秒杀领域服务
type SeckillDomainService struct {
	activityRepo repository.SeckillActivityRepository
	orderRepo    repository.SeckillOrderRepository
	redisClient  *redis.Client
	rateLimiter  ratelimiter.RateLimiter
}

// NewSeckillDomainService 创建秒杀领域服务
func NewSeckillDomainService(
	activityRepo repository.SeckillActivityRepository,
	orderRepo repository.SeckillOrderRepository,
	redisClient *redis.Client,
	rateLimiter ratelimiter.RateLimiter,
) *SeckillDomainService {
	return &SeckillDomainService{
		activityRepo: activityRepo,
		orderRepo:    orderRepo,
		redisClient:  redisClient,
		rateLimiter:  rateLimiter,
	}
}

// ValidateSeckillRequest 验证秒杀请求
func (s *SeckillDomainService) ValidateSeckillRequest(
	ctx context.Context,
	userID domain.UserID,
	activityID domain.ID,
	quantity int,
) (*entity.SeckillActivity, error) {
	// 1. 检查用户限流
	rateLimitKey := fmt.Sprintf("seckill:user:%s", userID.String())
	allowed, err := s.rateLimiter.Allow(ctx, rateLimitKey)
	if err != nil {
		return nil, domain.NewInternalError("rate limit check failed", err)
	}
	if !allowed {
		return nil, domain.NewBusinessError(domain.ErrCodeTooManyRequests, "too many requests, please try again later")
	}

	// 2. 查找秒杀活动
	activity, err := s.activityRepo.FindByID(ctx, activityID)
	if err != nil {
		return nil, domain.NewInternalError("failed to find seckill activity", err)
	}
	if activity == nil {
		return nil, domain.NewNotFoundError("seckill activity", activityID.String())
	}

	// 3. 检查活动是否激活
	now := time.Now()
	if !activity.IsActive(now) {
		return nil, domain.NewBusinessError(domain.ErrCodeValidation, "seckill activity is not active")
	}

	// 4. 检查库存（预检查，实际扣减在Redis中原子操作）
	if activity.RemainingStock() < quantity {
		return nil, domain.NewInsufficientStockError(activity.ProductID().String(), quantity, activity.RemainingStock())
	}

	// 5. 检查用户是否已经参与过该活动
	hasParticipated, err := s.HasUserParticipated(ctx, userID, activityID)
	if err != nil {
		return nil, err
	}
	if hasParticipated {
		return nil, domain.NewBusinessError(domain.ErrCodeValidation, "user has already participated in this activity")
	}

	return activity, nil
}

// AtomicDeductStock 原子性扣减库存（基于Redis Lua脚本）
func (s *SeckillDomainService) AtomicDeductStock(
	ctx context.Context,
	activityID domain.ID,
	quantity int,
) (bool, error) {
	// 使用Redis Lua脚本确保原子性
	luaScript := `
		local key = KEYS[1]
		local deduct = tonumber(ARGV[1])
		
		-- 获取当前库存
		local current = redis.call('GET', key)
		if not current then
			return {0, 0}
		end
		
		current = tonumber(current)
		
		-- 检查库存是否足够
		if current < deduct then
			return {0, current}
		end
		
		-- 扣减库存
		local remaining = redis.call('DECRBY', key, deduct)
		return {1, remaining}
	`

	stockKey := fmt.Sprintf("seckill:stock:%s", activityID.String())
	result, err := s.redisClient.Eval(ctx, luaScript, []string{stockKey}, quantity).Result()
	if err != nil {
		return false, fmt.Errorf("failed to execute stock deduction script: %w", err)
	}

	resultSlice, ok := result.([]interface{})
	if !ok || len(resultSlice) != 2 {
		return false, fmt.Errorf("unexpected script result format")
	}

	success, ok := resultSlice[0].(int64)
	if !ok {
		return false, fmt.Errorf("unexpected success result type")
	}

	return success == 1, nil
}

// RestoreStock 恢复库存（订单取消或过期时）
func (s *SeckillDomainService) RestoreStock(
	ctx context.Context,
	activityID domain.ID,
	quantity int,
) error {
	stockKey := fmt.Sprintf("seckill:stock:%s", activityID.String())
	_, err := s.redisClient.IncrBy(ctx, stockKey, int64(quantity)).Result()
	return err
}

// InitializeActivityStock 初始化活动库存到Redis
func (s *SeckillDomainService) InitializeActivityStock(
	ctx context.Context,
	activity *entity.SeckillActivity,
) error {
	stockKey := fmt.Sprintf("seckill:stock:%s", activity.ID().String())
	
	// 设置库存，过期时间为活动结束时间后1小时
	expiration := time.Until(activity.EndTime()) + time.Hour
	err := s.redisClient.Set(ctx, stockKey, activity.TotalStock(), expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to initialize activity stock: %w", err)
	}

	return nil
}

// HasUserParticipated 检查用户是否已参与活动
func (s *SeckillDomainService) HasUserParticipated(
	ctx context.Context,
	userID domain.UserID,
	activityID domain.ID,
) (bool, error) {
	// 先检查Redis缓存
	participationKey := fmt.Sprintf("seckill:participation:%s:%s", activityID.String(), userID.String())
	exists, err := s.redisClient.Exists(ctx, participationKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check participation cache: %w", err)
	}
	if exists > 0 {
		return true, nil
	}

	// 检查数据库
	orders, err := s.orderRepo.FindByUserAndActivity(ctx, userID, activityID)
	if err != nil {
		return false, fmt.Errorf("failed to check user participation: %w", err)
	}

	hasParticipated := len(orders) > 0
	
	// 如果已参与，缓存结果
	if hasParticipated {
		// 缓存到活动结束
		activity, err := s.activityRepo.FindByID(ctx, activityID)
		if err == nil && activity != nil {
			expiration := time.Until(activity.EndTime())
			if expiration > 0 {
				s.redisClient.Set(ctx, participationKey, "1", expiration)
			}
		}
	}

	return hasParticipated, nil
}

// MarkUserParticipation 标记用户参与活动
func (s *SeckillDomainService) MarkUserParticipation(
	ctx context.Context,
	userID domain.UserID,
	activityID domain.ID,
) error {
	participationKey := fmt.Sprintf("seckill:participation:%s:%s", activityID.String(), userID.String())
	
	// 获取活动信息以设置合适的过期时间
	activity, err := s.activityRepo.FindByID(ctx, activityID)
	if err != nil {
		return fmt.Errorf("failed to find activity: %w", err)
	}
	if activity == nil {
		return fmt.Errorf("activity not found")
	}

	// 缓存到活动结束
	expiration := time.Until(activity.EndTime())
	if expiration <= 0 {
		expiration = time.Hour // 如果活动已结束，缓存1小时
	}

	return s.redisClient.Set(ctx, participationKey, "1", expiration).Err()
}

// ProcessSeckillOrder 处理秒杀订单（完整流程）
func (s *SeckillDomainService) ProcessSeckillOrder(
	ctx context.Context,
	userID domain.UserID,
	activityID domain.ID,
	quantity int,
) (*entity.SeckillOrder, error) {
	// 1. 验证秒杀请求
	activity, err := s.ValidateSeckillRequest(ctx, userID, activityID, quantity)
	if err != nil {
		return nil, err
	}

	// 2. 原子性扣减库存
	success, err := s.AtomicDeductStock(ctx, activityID, quantity)
	if err != nil {
		return nil, domain.NewInternalError("failed to deduct stock", err)
	}
	if !success {
		return nil, domain.NewInsufficientStockError(activity.ProductID().String(), quantity, 0)
	}

	// 3. 创建秒杀订单
	order, err := entity.NewSeckillOrder(
		userID,
		activityID,
		activity.ProductID(),
		quantity,
		activity.SeckillPrice(),
	)
	if err != nil {
		// 恢复库存
		s.RestoreStock(ctx, activityID, quantity)
		return nil, err
	}

	// 4. 保存订单
	if err := s.orderRepo.Save(ctx, order); err != nil {
		// 恢复库存
		s.RestoreStock(ctx, activityID, quantity)
		return nil, domain.NewInternalError("failed to save seckill order", err)
	}

	// 5. 标记用户参与
	if err := s.MarkUserParticipation(ctx, userID, activityID); err != nil {
		// 记录日志但不影响主流程
	}

	return order, nil
}

// CalculateActivityStatistics 计算活动统计信息
func (s *SeckillDomainService) CalculateActivityStatistics(
	ctx context.Context,
	activityID domain.ID,
) (*ActivityStatistics, error) {
	// 获取活动信息
	activity, err := s.activityRepo.FindByID(ctx, activityID)
	if err != nil {
		return nil, err
	}
	if activity == nil {
		return nil, domain.NewNotFoundError("activity", activityID.String())
	}

	// 获取Redis中的实时库存
	stockKey := fmt.Sprintf("seckill:stock:%s", activityID.String())
	currentStock, err := s.redisClient.Get(ctx, stockKey).Int()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to get current stock: %w", err)
	}
	if err == redis.Nil {
		currentStock = activity.RemainingStock()
	}

	// 计算统计信息
	soldCount := activity.TotalStock() - currentStock
	successRate := float64(soldCount) / float64(activity.TotalStock()) * 100

	return &ActivityStatistics{
		ActivityID:     activityID.String(),
		TotalStock:     activity.TotalStock(),
		RemainingStock: currentStock,
		SoldCount:      soldCount,
		SuccessRate:    successRate,
		Status:         string(activity.Status()),
	}, nil
}

// ActivityStatistics 活动统计信息
type ActivityStatistics struct {
	ActivityID     string  `json:"activity_id"`
	TotalStock     int     `json:"total_stock"`
	RemainingStock int     `json:"remaining_stock"`
	SoldCount      int     `json:"sold_count"`
	SuccessRate    float64 `json:"success_rate"`
	Status         string  `json:"status"`
}
