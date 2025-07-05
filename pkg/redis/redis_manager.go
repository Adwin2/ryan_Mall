package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisManager Redis管理器 - 支持单机和集群模式
type RedisManager struct {
	client        redis.Cmdable // 使用接口，支持单机和集群
	clusterClient *redis.ClusterClient
	singleClient  *redis.Client
	ctx           context.Context
	isCluster     bool
}

// NewRedisManager 创建单机Redis管理器
func NewRedisManager(addr, password string, db int) *RedisManager {
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		PoolSize:     100,
		MinIdleConns: 10,
		MaxRetries:   3,
	})

	return &RedisManager{
		client:       rdb,
		singleClient: rdb,
		ctx:          context.Background(),
		isCluster:    false,
	}
}

// NewRedisClusterManager 创建Redis集群管理器
func NewRedisClusterManager(addrs []string, password string) *RedisManager {
	rdb := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        addrs,
		Password:     password,
		PoolSize:     100,
		MinIdleConns: 10,
		MaxRetries:   3,
		// 集群特定配置
		MaxRedirects:   8,
		ReadOnly:       false,
		RouteByLatency: true,
		RouteRandomly:  true,
	})

	return &RedisManager{
		client:        rdb,
		clusterClient: rdb,
		ctx:           context.Background(),
		isCluster:     true,
	}
}

// GetClient 获取Redis客户端
func (rm *RedisManager) GetClient() redis.Cmdable {
	return rm.client
}

// IsCluster 检查是否为集群模式
func (rm *RedisManager) IsCluster() bool {
	return rm.isCluster
}

// GetClusterInfo 获取集群信息（仅集群模式）
func (rm *RedisManager) GetClusterInfo() (map[string]interface{}, error) {
	if !rm.isCluster {
		return nil, fmt.Errorf("not in cluster mode")
	}

	info := make(map[string]interface{})

	// 获取集群节点信息
	nodes, err := rm.clusterClient.ClusterNodes(rm.ctx).Result()
	if err != nil {
		return nil, err
	}
	info["nodes"] = nodes

	// 获取集群状态
	clusterInfo, err := rm.clusterClient.ClusterInfo(rm.ctx).Result()
	if err != nil {
		return nil, err
	}
	info["cluster_info"] = clusterInfo

	// 获取集群槽位信息
	slots, err := rm.clusterClient.ClusterSlots(rm.ctx).Result()
	if err != nil {
		return nil, err
	}
	info["slots"] = len(slots)

	return info, nil
}

// InventoryManager 库存管理器
type InventoryManager struct {
	redis *RedisManager
}

// NewInventoryManager 创建库存管理器
func NewInventoryManager(redis *RedisManager) *InventoryManager {
	return &InventoryManager{redis: redis}
}

// 库存扣减Lua脚本 - 防止超卖
const decreaseStockScript = `
local key = KEYS[1]
local quantity = tonumber(ARGV[1])
local current = redis.call('GET', key)

if current == false then
    return -1  -- 商品不存在
end

current = tonumber(current)
if current < quantity then
    return -2  -- 库存不足
end

local new_stock = current - quantity
redis.call('SET', key, new_stock)
return new_stock  -- 返回剩余库存
`

// DecreaseStock 原子性库存扣减
func (im *InventoryManager) DecreaseStock(productID uint, quantity int) (int, error) {
	key := fmt.Sprintf("product:stock:%d", productID)
	
	result, err := im.redis.client.Eval(im.redis.ctx, decreaseStockScript, []string{key}, quantity).Result()
	if err != nil {
		return 0, err
	}
	
	stock := result.(int64)
	switch stock {
	case -1:
		return 0, fmt.Errorf("product %d not found", productID)
	case -2:
		return 0, fmt.Errorf("insufficient stock for product %d", productID)
	default:
		return int(stock), nil
	}
}

// 分布式锁Lua脚本
const distributedLockScript = `
if redis.call('GET', KEYS[1]) == ARGV[1] then
    return redis.call('DEL', KEYS[1])
else
    return 0
end
`

// DistributedLock 分布式锁
type DistributedLock struct {
	redis    *RedisManager
	key      string
	value    string
	expiry   time.Duration
	acquired bool
}

// NewDistributedLock 创建分布式锁
func (rm *RedisManager) NewDistributedLock(key, value string, expiry time.Duration) *DistributedLock {
	return &DistributedLock{
		redis:  rm,
		key:    key,
		value:  value,
		expiry: expiry,
	}
}

// TryLock 尝试获取锁
func (dl *DistributedLock) TryLock() (bool, error) {
	result, err := dl.redis.client.SetNX(dl.redis.ctx, dl.key, dl.value, dl.expiry).Result()
	if err != nil {
		return false, err
	}
	
	dl.acquired = result
	return result, nil
}

// Unlock 释放锁
func (dl *DistributedLock) Unlock() error {
	if !dl.acquired {
		return nil
	}
	
	_, err := dl.redis.client.Eval(dl.redis.ctx, distributedLockScript, []string{dl.key}, dl.value).Result()
	if err != nil {
		return err
	}
	
	dl.acquired = false
	return nil
}

// 限流器Lua脚本 - 滑动窗口
const rateLimiterScript = `
local key = KEYS[1]
local window = tonumber(ARGV[1])
local limit = tonumber(ARGV[2])
local current_time = tonumber(ARGV[3])

-- 清理过期的记录
redis.call('ZREMRANGEBYSCORE', key, 0, current_time - window)

-- 获取当前窗口内的请求数
local current_requests = redis.call('ZCARD', key)

if current_requests < limit then
    -- 添加当前请求
    redis.call('ZADD', key, current_time, current_time)
    redis.call('EXPIRE', key, window)
    return {1, limit - current_requests - 1}  -- {允许, 剩余次数}
else
    return {0, 0}  -- {拒绝, 剩余次数}
end
`

// RateLimiter 限流器
type RateLimiter struct {
	redis *RedisManager
}

// NewRateLimiter 创建限流器
func NewRateLimiter(redis *RedisManager) *RateLimiter {
	return &RateLimiter{redis: redis}
}

// IsAllowed 检查是否允许请求
func (rl *RateLimiter) IsAllowed(userID uint, window time.Duration, limit int) (bool, int, error) {
	key := fmt.Sprintf("rate_limit:user:%d", userID)
	currentTime := time.Now().Unix()
	
	result, err := rl.redis.client.Eval(
		rl.redis.ctx,
		rateLimiterScript,
		[]string{key},
		int(window.Seconds()),
		limit,
		currentTime,
	).Result()
	
	if err != nil {
		return false, 0, err
	}
	
	res := result.([]interface{})
	allowed := res[0].(int64) == 1
	remaining := int(res[1].(int64))
	
	return allowed, remaining, nil
}

// CacheManager 缓存管理器
type CacheManager struct {
	redis *RedisManager
}

// NewCacheManager 创建缓存管理器
func NewCacheManager(redis *RedisManager) *CacheManager {
	return &CacheManager{redis: redis}
}

// 缓存预热Lua脚本
const cacheWarmupScript = `
local keys = KEYS
local values = ARGV
local ttl = tonumber(ARGV[#ARGV])

for i = 1, #keys do
    if values[i] ~= nil then
        redis.call('SETEX', keys[i], ttl, values[i])
    end
end

return #keys
`

// BatchSet 批量设置缓存
func (cm *CacheManager) BatchSet(data map[string]interface{}, ttl time.Duration) error {
	if len(data) == 0 {
		return nil
	}
	
	keys := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data)+1)
	
	for k, v := range data {
		keys = append(keys, k)
		values = append(values, v)
	}
	values = append(values, int(ttl.Seconds()))
	
	_, err := cm.redis.client.Eval(cm.redis.ctx, cacheWarmupScript, keys, values...).Result()
	return err
}

// GetMulti 批量获取缓存
func (cm *CacheManager) GetMulti(keys []string) (map[string]string, error) {
	if len(keys) == 0 {
		return make(map[string]string), nil
	}
	
	values, err := cm.redis.client.MGet(cm.redis.ctx, keys...).Result()
	if err != nil {
		return nil, err
	}
	
	result := make(map[string]string)
	for i, key := range keys {
		if values[i] != nil {
			result[key] = values[i].(string)
		}
	}
	
	return result, nil
}

// 购物车缓存管理
type CartCacheManager struct {
	redis *RedisManager
}

// NewCartCacheManager 创建购物车缓存管理器
func NewCartCacheManager(redis *RedisManager) *CartCacheManager {
	return &CartCacheManager{redis: redis}
}

// AddToCart 添加商品到购物车缓存
func (ccm *CartCacheManager) AddToCart(userID, productID uint, quantity int) error {
	key := fmt.Sprintf("cart:user:%d", userID)
	field := fmt.Sprintf("product:%d", productID)
	
	// 使用HINCRBY原子性增加数量
	_, err := ccm.redis.client.HIncrBy(ccm.redis.ctx, key, field, int64(quantity)).Result()
	if err != nil {
		return err
	}
	
	// 设置过期时间
	ccm.redis.client.Expire(ccm.redis.ctx, key, 24*time.Hour)
	return nil
}

// GetCart 获取购物车
func (ccm *CartCacheManager) GetCart(userID uint) (map[string]string, error) {
	key := fmt.Sprintf("cart:user:%d", userID)
	return ccm.redis.client.HGetAll(ccm.redis.ctx, key).Result()
}

// RemoveFromCart 从购物车移除商品
func (ccm *CartCacheManager) RemoveFromCart(userID, productID uint) error {
	key := fmt.Sprintf("cart:user:%d", userID)
	field := fmt.Sprintf("product:%d", productID)
	
	return ccm.redis.client.HDel(ccm.redis.ctx, key, field).Err()
}

// ClearCart 清空购物车
func (ccm *CartCacheManager) ClearCart(userID uint) error {
	key := fmt.Sprintf("cart:user:%d", userID)
	return ccm.redis.client.Del(ccm.redis.ctx, key).Err()
}

// 热点数据管理
type HotDataManager struct {
	redis *RedisManager
}

// NewHotDataManager 创建热点数据管理器
func NewHotDataManager(redis *RedisManager) *HotDataManager {
	return &HotDataManager{redis: redis}
}

// IncrementViewCount 增加商品浏览次数
func (hdm *HotDataManager) IncrementViewCount(productID uint) (int64, error) {
	key := fmt.Sprintf("hot:product:view:%d", productID)
	count, err := hdm.redis.client.Incr(hdm.redis.ctx, key).Result()
	if err != nil {
		return 0, err
	}
	
	// 设置过期时间为1天
	hdm.redis.client.Expire(hdm.redis.ctx, key, 24*time.Hour)
	return count, nil
}

// GetHotProducts 获取热门商品
func (hdm *HotDataManager) GetHotProducts(limit int) ([]string, error) {
	key := "hot:products:ranking"
	
	// 使用ZREVRANGE获取热门商品（按浏览量倒序）
	return hdm.redis.client.ZRevRange(hdm.redis.ctx, key, 0, int64(limit-1)).Result()
}

// UpdateHotProductsRanking 更新热门商品排行
func (hdm *HotDataManager) UpdateHotProductsRanking(productID uint, score float64) error {
	key := "hot:products:ranking"
	member := fmt.Sprintf("product:%d", productID)
	
	return hdm.redis.client.ZAdd(hdm.redis.ctx, key, &redis.Z{
		Score:  score,
		Member: member,
	}).Err()
}
