package cache

import (
	"encoding/json"
	"errors"
	"sync"
	"time"
)

// SimpleCache 简单的内存缓存实现
// 在没有Redis的情况下提供基础缓存功能
type SimpleCache struct {
	data   map[string]*CacheItem
	mutex  sync.RWMutex
	ticker *time.Ticker
	stop   chan bool
}

// CacheItem 缓存项
type CacheItem struct {
	Value      interface{}
	Expiration time.Time
}

// NewSimpleCache 创建简单缓存
func NewSimpleCache() *SimpleCache {
	cache := &SimpleCache{
		data:   make(map[string]*CacheItem),
		ticker: time.NewTicker(1 * time.Minute), // 每分钟清理一次过期数据
		stop:   make(chan bool),
	}
	
	// 启动清理协程
	go cache.cleanup()
	
	return cache
}

// Set 设置缓存
func (c *SimpleCache) Set(key string, value interface{}, duration time.Duration) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	expiration := time.Now().Add(duration)
	c.data[key] = &CacheItem{
		Value:      value,
		Expiration: expiration,
	}
	
	return nil
}

// Get 获取缓存
func (c *SimpleCache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	item, exists := c.data[key]
	if !exists {
		return nil, false
	}
	
	// 检查是否过期
	if time.Now().After(item.Expiration) {
		delete(c.data, key)
		return nil, false
	}
	
	return item.Value, true
}

// Delete 删除缓存
func (c *SimpleCache) Delete(key string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	delete(c.data, key)
	return nil
}

// GetString 获取字符串缓存
func (c *SimpleCache) GetString(key string) (string, bool) {
	value, exists := c.Get(key)
	if !exists {
		return "", false
	}
	
	if str, ok := value.(string); ok {
		return str, true
	}
	
	return "", false
}

// SetJSON 设置JSON缓存
func (c *SimpleCache) SetJSON(key string, value interface{}, duration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	
	return c.Set(key, string(data), duration)
}

// GetJSON 获取JSON缓存
func (c *SimpleCache) GetJSON(key string, dest interface{}) error {
	jsonStr, exists := c.GetString(key)
	if !exists {
		return errors.New("cache not found")
	}

	return json.Unmarshal([]byte(jsonStr), dest)
}

// Exists 检查缓存是否存在
func (c *SimpleCache) Exists(key string) bool {
	_, exists := c.Get(key)
	return exists
}

// Clear 清空所有缓存
func (c *SimpleCache) Clear() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data = make(map[string]*CacheItem)
	return nil
}

// Size 获取缓存大小
func (c *SimpleCache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	return len(c.data)
}

// cleanup 清理过期数据
func (c *SimpleCache) cleanup() {
	for {
		select {
		case <-c.ticker.C:
			c.mutex.Lock()
			now := time.Now()
			for key, item := range c.data {
				if now.After(item.Expiration) {
					delete(c.data, key)
				}
			}
			c.mutex.Unlock()
		case <-c.stop:
			return
		}
	}
}

// Stats 获取缓存统计信息
func (c *SimpleCache) Stats() map[string]interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return map[string]interface{}{
		"cache_type": "simple_memory_cache",
		"cache_size": len(c.data),
	}
}

// Close 关闭缓存
func (c *SimpleCache) Close() {
	c.ticker.Stop()
	c.stop <- true
}

// CacheManager 缓存管理器接口
type CacheManager interface {
	Set(key string, value interface{}, duration time.Duration) error
	Get(key string) (interface{}, bool)
	Delete(key string) error
	SetJSON(key string, value interface{}, duration time.Duration) error
	GetJSON(key string, dest interface{}) error // 修改返回类型
	Exists(key string) bool
	Clear() error // 修改返回类型
	Size() int
	Stats() map[string]interface{} // 新增统计接口
}

// 全局缓存实例
var globalCache CacheManager

// InitCache 初始化缓存
// 使用分片缓存以获得更好的并发性能
func InitCache() {
	// 使用32个分片的高性能缓存
	globalCache = NewShardedCache(32)
}

// GetCache 获取全局缓存实例
func GetCache() CacheManager {
	if globalCache == nil {
		// 如果没有设置全局缓存，使用默认的简单缓存
		InitCache()
	}
	return globalCache
}

// SetGlobalCache 设置全局缓存实例
func SetGlobalCache(cache CacheManager) {
	globalCache = cache
}

// 便捷函数
func Set(key string, value interface{}, duration time.Duration) error {
	return GetCache().Set(key, value, duration)
}

func Get(key string) (interface{}, bool) {
	return GetCache().Get(key)
}

func Delete(key string) error {
	return GetCache().Delete(key)
}

func SetJSON(key string, value interface{}, duration time.Duration) error {
	return GetCache().SetJSON(key, value, duration)
}

func GetJSON(key string, dest interface{}) error {
	return GetCache().GetJSON(key, dest)
}

func Exists(key string) bool {
	return GetCache().Exists(key)
}

func Clear() {
	GetCache().Clear()
}

func Size() int {
	return GetCache().Size()
}
