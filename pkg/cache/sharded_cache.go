package cache

import (
	"encoding/json"
	"errors"
	"hash/fnv"
	"sync"
	"time"
)

// 错误定义
var (
	ErrCacheNotFound = errors.New("cache not found")
)

// ShardedCache 分片缓存，减少锁竞争
// 通过将数据分散到多个分片中，减少并发访问时的锁竞争
type ShardedCache struct {
	shards    []*CacheShard
	shardMask uint32
}

// CacheShard 缓存分片
type CacheShard struct {
	data   map[string]*CacheItem
	mutex  sync.RWMutex
	ticker *time.Ticker
	stop   chan bool
}

// NewShardedCache 创建分片缓存
// shardCount 必须是2的幂，推荐16或32
func NewShardedCache(shardCount int) *ShardedCache {
	// 确保shardCount是2的幂
	if shardCount <= 0 || (shardCount&(shardCount-1)) != 0 {
		shardCount = 16 // 默认16个分片
	}

	cache := &ShardedCache{
		shards:    make([]*CacheShard, shardCount),
		shardMask: uint32(shardCount - 1),
	}

	// 初始化每个分片
	for i := 0; i < shardCount; i++ {
		cache.shards[i] = &CacheShard{
			data:   make(map[string]*CacheItem),
			ticker: time.NewTicker(2 * time.Minute), // 每2分钟清理一次
			stop:   make(chan bool),
		}
		
		// 启动每个分片的清理协程
		go cache.shards[i].cleanup()
	}

	return cache
}

// getShard 根据key获取对应的分片
func (c *ShardedCache) getShard(key string) *CacheShard {
	hash := fnv.New32a()
	hash.Write([]byte(key))
	return c.shards[hash.Sum32()&c.shardMask]
}

// Set 设置缓存
func (c *ShardedCache) Set(key string, value interface{}, duration time.Duration) error {
	shard := c.getShard(key)
	shard.mutex.Lock()
	defer shard.mutex.Unlock()

	expiration := time.Now().Add(duration)
	shard.data[key] = &CacheItem{
		Value:      value,
		Expiration: expiration,
	}

	return nil
}

// Get 获取缓存
func (c *ShardedCache) Get(key string) (interface{}, bool) {
	shard := c.getShard(key)
	shard.mutex.RLock()
	defer shard.mutex.RUnlock()

	item, exists := shard.data[key]
	if !exists {
		return nil, false
	}

	// 检查是否过期
	if time.Now().After(item.Expiration) {
		// 过期了，需要删除（但这里不删除，留给清理协程）
		return nil, false
	}

	return item.Value, true
}

// SetJSON 设置JSON缓存
func (c *ShardedCache) SetJSON(key string, value interface{}, duration time.Duration) error {
	return c.Set(key, value, duration)
}

// GetJSON 获取JSON缓存
func (c *ShardedCache) GetJSON(key string, dest interface{}) error {
	value, exists := c.Get(key)
	if !exists {
		return ErrCacheNotFound
	}

	// 如果存储的就是目标类型，直接赋值
	if v, ok := value.([]byte); ok {
		return json.Unmarshal(v, dest)
	}

	// 否则先序列化再反序列化
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, dest)
}

// Exists 检查缓存是否存在
func (c *ShardedCache) Exists(key string) bool {
	_, exists := c.Get(key)
	return exists
}

// Delete 删除缓存
func (c *ShardedCache) Delete(key string) error {
	shard := c.getShard(key)
	shard.mutex.Lock()
	defer shard.mutex.Unlock()

	delete(shard.data, key)
	return nil
}

// DeletePattern 删除匹配模式的缓存
func (c *ShardedCache) DeletePattern(pattern string) error {
	// 遍历所有分片
	for _, shard := range c.shards {
		shard.mutex.Lock()
		for key := range shard.data {
			if matchPattern(key, pattern) {
				delete(shard.data, key)
			}
		}
		shard.mutex.Unlock()
	}
	return nil
}

// Clear 清空所有缓存
func (c *ShardedCache) Clear() error {
	for _, shard := range c.shards {
		shard.mutex.Lock()
		shard.data = make(map[string]*CacheItem)
		shard.mutex.Unlock()
	}
	return nil
}

// Size 获取缓存大小
func (c *ShardedCache) Size() int {
	total := 0
	for _, shard := range c.shards {
		shard.mutex.RLock()
		total += len(shard.data)
		shard.mutex.RUnlock()
	}
	return total
}

// Stats 获取缓存统计信息
func (c *ShardedCache) Stats() map[string]interface{} {
	stats := map[string]interface{}{
		"cache_type":   "sharded_memory_cache",
		"cache_size":   c.Size(),
		"shard_count":  len(c.shards),
		"shard_stats":  make([]map[string]interface{}, len(c.shards)),
	}

	// 获取每个分片的统计
	for i, shard := range c.shards {
		shard.mutex.RLock()
		shardStats := map[string]interface{}{
			"shard_id": i,
			"size":     len(shard.data),
		}
		stats["shard_stats"].([]map[string]interface{})[i] = shardStats
		shard.mutex.RUnlock()
	}

	return stats
}

// Close 关闭缓存
func (c *ShardedCache) Close() error {
	for _, shard := range c.shards {
		shard.stop <- true
		shard.ticker.Stop()
	}
	return nil
}

// cleanup 清理过期数据
func (s *CacheShard) cleanup() {
	for {
		select {
		case <-s.ticker.C:
			s.mutex.Lock()
			now := time.Now()
			for key, item := range s.data {
				if now.After(item.Expiration) {
					delete(s.data, key)
				}
			}
			s.mutex.Unlock()
		case <-s.stop:
			return
		}
	}
}

// matchPattern 简单的模式匹配
func matchPattern(key, pattern string) bool {
	if pattern == "*" {
		return true
	}
	
	// 支持前缀匹配，如 "product:*"
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(key) >= len(prefix) && key[:len(prefix)] == prefix
	}
	
	return key == pattern
}
