package ratelimiter

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// RateLimiter 分布式限流器接口
type RateLimiter interface {
	// Allow 检查是否允许请求
	Allow(ctx context.Context, key string) (bool, error)
	
	// AllowN 检查是否允许N个请求
	AllowN(ctx context.Context, key string, n int) (bool, error)
	
	// Reset 重置限流器
	Reset(ctx context.Context, key string) error
	
	// GetRemaining 获取剩余请求数
	GetRemaining(ctx context.Context, key string) (int, error)
}

// RedisRateLimiter 基于Redis的分布式限流器（类似Redisson RateLimiter）
type RedisRateLimiter struct {
	client   *redis.Client
	rate     int           // 每个时间窗口允许的请求数
	interval time.Duration // 时间窗口大小
}

// NewRedisRateLimiter 创建Redis限流器
func NewRedisRateLimiter(client *redis.Client, rate int, interval time.Duration) *RedisRateLimiter {
	return &RedisRateLimiter{
		client:   client,
		rate:     rate,
		interval: interval,
	}
}

// Allow 检查是否允许单个请求
func (r *RedisRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	return r.AllowN(ctx, key, 1)
}

// AllowN 检查是否允许N个请求（滑动窗口算法）
func (r *RedisRateLimiter) AllowN(ctx context.Context, key string, n int) (bool, error) {
	if n <= 0 {
		return false, fmt.Errorf("n must be positive")
	}

	// 使用Lua脚本确保原子性
	luaScript := `
		local key = KEYS[1]
		local window = tonumber(ARGV[1])
		local limit = tonumber(ARGV[2])
		local requests = tonumber(ARGV[3])
		local now = tonumber(ARGV[4])
		
		-- 清理过期的请求记录
		redis.call('ZREMRANGEBYSCORE', key, 0, now - window * 1000)
		
		-- 获取当前窗口内的请求数
		local current = redis.call('ZCARD', key)
		
		-- 检查是否超过限制
		if current + requests > limit then
			return {0, current, limit - current}
		end
		
		-- 添加新的请求记录
		for i = 1, requests do
			redis.call('ZADD', key, now, now .. ':' .. i)
		end
		
		-- 设置过期时间
		redis.call('EXPIRE', key, math.ceil(window))
		
		return {1, current + requests, limit - current - requests}
	`

	now := time.Now().UnixMilli()
	windowMs := r.interval.Milliseconds()

	result, err := r.client.Eval(ctx, luaScript, []string{key}, windowMs, r.rate, n, now).Result()
	if err != nil {
		return false, fmt.Errorf("failed to execute rate limit script: %w", err)
	}

	resultSlice, ok := result.([]interface{})
	if !ok || len(resultSlice) != 3 {
		return false, fmt.Errorf("unexpected script result format")
	}

	allowed, ok := resultSlice[0].(int64)
	if !ok {
		return false, fmt.Errorf("unexpected allowed result type")
	}

	return allowed == 1, nil
}

// Reset 重置限流器
func (r *RedisRateLimiter) Reset(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// GetRemaining 获取剩余请求数
func (r *RedisRateLimiter) GetRemaining(ctx context.Context, key string) (int, error) {
	luaScript := `
		local key = KEYS[1]
		local window = tonumber(ARGV[1])
		local limit = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])
		
		-- 清理过期的请求记录
		redis.call('ZREMRANGEBYSCORE', key, 0, now - window * 1000)
		
		-- 获取当前窗口内的请求数
		local current = redis.call('ZCARD', key)
		
		return limit - current
	`

	now := time.Now().UnixMilli()
	windowMs := r.interval.Milliseconds()

	result, err := r.client.Eval(ctx, luaScript, []string{key}, windowMs, r.rate, now).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get remaining requests: %w", err)
	}

	remaining, ok := result.(int64)
	if !ok {
		return 0, fmt.Errorf("unexpected result type")
	}

	if remaining < 0 {
		remaining = 0
	}

	return int(remaining), nil
}

// TokenBucketRateLimiter 令牌桶限流器
type TokenBucketRateLimiter struct {
	client   *redis.Client
	capacity int           // 桶容量
	refillRate int         // 每秒补充的令牌数
	interval time.Duration // 补充间隔
}

// NewTokenBucketRateLimiter 创建令牌桶限流器
func NewTokenBucketRateLimiter(client *redis.Client, capacity, refillRate int, interval time.Duration) *TokenBucketRateLimiter {
	return &TokenBucketRateLimiter{
		client:     client,
		capacity:   capacity,
		refillRate: refillRate,
		interval:   interval,
	}
}

// Allow 检查是否允许单个请求
func (t *TokenBucketRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	return t.AllowN(ctx, key, 1)
}

// AllowN 检查是否允许N个请求（令牌桶算法）
func (t *TokenBucketRateLimiter) AllowN(ctx context.Context, key string, n int) (bool, error) {
	if n <= 0 {
		return false, fmt.Errorf("n must be positive")
	}

	luaScript := `
		local key = KEYS[1]
		local capacity = tonumber(ARGV[1])
		local refill_rate = tonumber(ARGV[2])
		local interval_ms = tonumber(ARGV[3])
		local requests = tonumber(ARGV[4])
		local now = tonumber(ARGV[5])
		
		-- 获取当前状态
		local bucket_data = redis.call('HMGET', key, 'tokens', 'last_refill')
		local tokens = tonumber(bucket_data[1]) or capacity
		local last_refill = tonumber(bucket_data[2]) or now
		
		-- 计算需要补充的令牌数
		local time_passed = now - last_refill
		local tokens_to_add = math.floor(time_passed / interval_ms * refill_rate)
		
		-- 补充令牌，但不超过容量
		tokens = math.min(capacity, tokens + tokens_to_add)
		
		-- 检查是否有足够的令牌
		if tokens < requests then
			-- 更新状态（即使请求被拒绝也要更新时间）
			redis.call('HMSET', key, 'tokens', tokens, 'last_refill', now)
			redis.call('EXPIRE', key, math.ceil(interval_ms / 1000 * 2))
			return {0, tokens}
		end
		
		-- 消费令牌
		tokens = tokens - requests
		
		-- 更新状态
		redis.call('HMSET', key, 'tokens', tokens, 'last_refill', now)
		redis.call('EXPIRE', key, math.ceil(interval_ms / 1000 * 2))
		
		return {1, tokens}
	`

	now := time.Now().UnixMilli()
	intervalMs := t.interval.Milliseconds()

	result, err := t.client.Eval(ctx, luaScript, []string{key}, t.capacity, t.refillRate, intervalMs, n, now).Result()
	if err != nil {
		return false, fmt.Errorf("failed to execute token bucket script: %w", err)
	}

	resultSlice, ok := result.([]interface{})
	if !ok || len(resultSlice) != 2 {
		return false, fmt.Errorf("unexpected script result format")
	}

	allowed, ok := resultSlice[0].(int64)
	if !ok {
		return false, fmt.Errorf("unexpected allowed result type")
	}

	return allowed == 1, nil
}

// Reset 重置令牌桶
func (t *TokenBucketRateLimiter) Reset(ctx context.Context, key string) error {
	return t.client.Del(ctx, key).Err()
}

// GetRemaining 获取剩余令牌数
func (t *TokenBucketRateLimiter) GetRemaining(ctx context.Context, key string) (int, error) {
	luaScript := `
		local key = KEYS[1]
		local capacity = tonumber(ARGV[1])
		local refill_rate = tonumber(ARGV[2])
		local interval_ms = tonumber(ARGV[3])
		local now = tonumber(ARGV[4])
		
		-- 获取当前状态
		local bucket_data = redis.call('HMGET', key, 'tokens', 'last_refill')
		local tokens = tonumber(bucket_data[1]) or capacity
		local last_refill = tonumber(bucket_data[2]) or now
		
		-- 计算需要补充的令牌数
		local time_passed = now - last_refill
		local tokens_to_add = math.floor(time_passed / interval_ms * refill_rate)
		
		-- 补充令牌，但不超过容量
		tokens = math.min(capacity, tokens + tokens_to_add)
		
		return tokens
	`

	now := time.Now().UnixMilli()
	intervalMs := t.interval.Milliseconds()

	result, err := t.client.Eval(ctx, luaScript, []string{key}, t.capacity, t.refillRate, intervalMs, now).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get remaining tokens: %w", err)
	}

	remaining, ok := result.(int64)
	if !ok {
		return 0, fmt.Errorf("unexpected result type")
	}

	return int(remaining), nil
}

// RateLimiterManager 限流器管理器
type RateLimiterManager struct {
	client *redis.Client
	limiters map[string]RateLimiter
}

// NewRateLimiterManager 创建限流器管理器
func NewRateLimiterManager(client *redis.Client) *RateLimiterManager {
	return &RateLimiterManager{
		client:   client,
		limiters: make(map[string]RateLimiter),
	}
}

// GetSlidingWindowLimiter 获取滑动窗口限流器
func (m *RateLimiterManager) GetSlidingWindowLimiter(name string, rate int, interval time.Duration) RateLimiter {
	key := fmt.Sprintf("sliding_window_%s_%d_%s", name, rate, interval.String())
	if limiter, exists := m.limiters[key]; exists {
		return limiter
	}
	
	limiter := NewRedisRateLimiter(m.client, rate, interval)
	m.limiters[key] = limiter
	return limiter
}

// GetTokenBucketLimiter 获取令牌桶限流器
func (m *RateLimiterManager) GetTokenBucketLimiter(name string, capacity, refillRate int, interval time.Duration) RateLimiter {
	key := fmt.Sprintf("token_bucket_%s_%d_%d_%s", name, capacity, refillRate, interval.String())
	if limiter, exists := m.limiters[key]; exists {
		return limiter
	}
	
	limiter := NewTokenBucketRateLimiter(m.client, capacity, refillRate, interval)
	m.limiters[key] = limiter
	return limiter
}

// GenerateUserKey 生成用户限流键
func GenerateUserKey(userID, action string) string {
	return fmt.Sprintf("rate_limit:user:%s:%s", userID, action)
}

// GenerateIPKey 生成IP限流键
func GenerateIPKey(ip, action string) string {
	return fmt.Sprintf("rate_limit:ip:%s:%s", ip, action)
}

// GenerateAPIKey 生成API限流键
func GenerateAPIKey(api, method string) string {
	return fmt.Sprintf("rate_limit:api:%s:%s", api, method)
}
