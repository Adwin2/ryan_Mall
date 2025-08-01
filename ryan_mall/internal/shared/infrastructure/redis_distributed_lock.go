package infrastructure

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// DistributedLock 分布式锁接口
type DistributedLock interface {
	Lock(ctx context.Context, key string, expiration time.Duration) (string, error)
	Unlock(ctx context.Context, key, token string) error
	Extend(ctx context.Context, key, token string, expiration time.Duration) error
}

// RedisDistributedLock Redis分布式锁实现
type RedisDistributedLock struct {
	client *redis.Client
}

// NewRedisDistributedLock 创建Redis分布式锁
func NewRedisDistributedLock(client *redis.Client) DistributedLock {
	return &RedisDistributedLock{
		client: client,
	}
}

// Lock 获取分布式锁
func (r *RedisDistributedLock) Lock(ctx context.Context, key string, expiration time.Duration) (string, error) {
	fmt.Printf("[DEBUG] 尝试获取分布式锁: key=%s, expiration=%v\n", key, expiration)

	// 生成唯一token
	token, err := generateToken()
	if err != nil {
		fmt.Printf("[DEBUG] 生成token失败: %v\n", err)
		return "", err
	}

	// 使用SET命令的NX和EX选项实现原子性加锁
	result := r.client.SetNX(ctx, key, token, expiration)
	if err := result.Err(); err != nil {
		fmt.Printf("[DEBUG] Redis SetNX失败: %v\n", err)
		return "", err
	}

	if !result.Val() {
		fmt.Printf("[DEBUG] 获取锁失败，锁已被占用: key=%s\n", key)
		return "", errors.New("failed to acquire lock")
	}

	fmt.Printf("[DEBUG] 成功获取分布式锁: key=%s, token=%s\n", key, token)
	return token, nil
}

// Unlock 释放分布式锁
func (r *RedisDistributedLock) Unlock(ctx context.Context, key, token string) error {
	fmt.Printf("[DEBUG] 尝试释放分布式锁: key=%s, token=%s\n", key, token)

	// 使用Lua脚本确保原子性：只有token匹配才能删除
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`

	result := r.client.Eval(ctx, script, []string{key}, token)
	if err := result.Err(); err != nil {
		fmt.Printf("[DEBUG] Redis Eval失败: %v\n", err)
		return err
	}

	if result.Val().(int64) == 0 {
		fmt.Printf("[DEBUG] 释放锁失败，token不匹配: key=%s\n", key)
		return errors.New("lock not found or token mismatch")
	}

	fmt.Printf("[DEBUG] 成功释放分布式锁: key=%s\n", key)
	return nil
}

// Extend 延长锁的过期时间
func (r *RedisDistributedLock) Extend(ctx context.Context, key, token string, expiration time.Duration) error {
	// 使用Lua脚本确保原子性：只有token匹配才能延长过期时间
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("EXPIRE", KEYS[1], ARGV[2])
		else
			return 0
		end
	`
	
	result := r.client.Eval(ctx, script, []string{key}, token, int(expiration.Seconds()))
	if err := result.Err(); err != nil {
		return err
	}

	if result.Val().(int64) == 0 {
		return errors.New("lock not found or token mismatch")
	}

	return nil
}

// generateToken 生成唯一token
func generateToken() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// LockManager 锁管理器
type LockManager struct {
	lock DistributedLock
}

// NewLockManager 创建锁管理器
func NewLockManager(lock DistributedLock) *LockManager {
	return &LockManager{
		lock: lock,
	}
}

// WithLock 使用分布式锁执行函数
func (m *LockManager) WithLock(ctx context.Context, key string, expiration time.Duration, fn func() error) error {
	// 获取锁
	token, err := m.lock.Lock(ctx, key, expiration)
	if err != nil {
		return err
	}

	// 确保释放锁
	defer func() {
		if unlockErr := m.lock.Unlock(ctx, key, token); unlockErr != nil {
			// 记录日志，但不影响主流程
		}
	}()

	// 执行业务逻辑
	return fn()
}

// TryWithLock 尝试获取锁并执行函数，如果获取失败立即返回
func (m *LockManager) TryWithLock(ctx context.Context, key string, expiration time.Duration, fn func() error) error {
	token, err := m.lock.Lock(ctx, key, expiration)
	if err != nil {
		return err
	}

	defer func() {
		if unlockErr := m.lock.Unlock(ctx, key, token); unlockErr != nil {
			// 记录日志，但不影响主流程
		}
	}()

	return fn()
}

// RetryWithLock 重试获取锁并执行函数
func (m *LockManager) RetryWithLock(ctx context.Context, key string, expiration time.Duration, maxRetries int, retryInterval time.Duration, fn func() error) error {
	var lastErr error
	
	for i := 0; i <= maxRetries; i++ {
		token, err := m.lock.Lock(ctx, key, expiration)
		if err != nil {
			lastErr = err
			if i < maxRetries {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(retryInterval):
					continue
				}
			}
			continue
		}

		// 获取锁成功，执行业务逻辑
		defer func() {
			if unlockErr := m.lock.Unlock(ctx, key, token); unlockErr != nil {
				// 记录日志，但不影响主流程
			}
		}()

		return fn()
	}

	return lastErr
}

// StockLockKey 生成库存锁的key
func StockLockKey(productID string) string {
	return "stock:lock:" + productID
}

// OrderLockKey 生成订单锁的key
func OrderLockKey(orderID string) string {
	return "order:lock:" + orderID
}

// SeckillLockKey 生成秒杀锁的key
func SeckillLockKey(activityID string) string {
	return "seckill:lock:" + activityID
}

// RedisClient Redis客户端接口
type RedisClient interface {
	Exists(ctx context.Context, key string) (bool, error)
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error)
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Del(ctx context.Context, keys ...string) error
	Incr(ctx context.Context, key string) (int64, error)
	Decr(ctx context.Context, key string) (int64, error)
}
