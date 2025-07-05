# 🔗 Redis集群应用集成指南 - Ryan Mall项目

## 📋 **概述**

本文档提供Redis集群在Ryan Mall项目中的应用集成方案，包括配置方法、代码示例和最佳实践。

## 🔧 **环境变量配置**

### **启用Redis集群模式**
```bash
# 在 .env 文件中添加以下配置
REDIS_CLUSTER_ENABLED=true
REDIS_CLUSTER_NODES=localhost:7001,localhost:7002,localhost:7003,localhost:7004,localhost:7005,localhost:7006
REDIS_CLUSTER_PASSWORD=
REDIS_CLUSTER_POOL_SIZE=100
REDIS_CLUSTER_MIN_IDLE_CONNS=10
REDIS_CLUSTER_MAX_RETRIES=3
```

### **混合缓存策略配置**
```bash
# 缓存策略配置
CACHE_STRATEGY=hybrid  # hybrid, memory, redis
MEMORY_CACHE_SIZE=1000  # 内存缓存条目数
REDIS_CACHE_TTL=3600   # Redis缓存TTL(秒)
HOT_DATA_THRESHOLD=100 # 热点数据阈值
```

## 💻 **代码集成示例**

### **1. 初始化Redis集群客户端**
```go
// pkg/redis/client.go
package redis

import (
    "context"
    "time"
    "github.com/go-redis/redis/v8"
)

var (
    ClusterClient *redis.ClusterClient
    ctx = context.Background()
)

// InitCluster 初始化Redis集群
func InitCluster(nodes []string, password string) error {
    ClusterClient = redis.NewClusterClient(&redis.ClusterOptions{
        Addrs:    nodes,
        Password: password,
        
        // 连接池配置
        PoolSize:     100,
        MinIdleConns: 10,
        MaxRetries:   3,
        
        // 集群配置
        MaxRedirects:   8,
        ReadOnly:       false,
        RouteByLatency: true,
        RouteRandomly:  true,
        
        // 超时配置
        DialTimeout:  5 * time.Second,
        ReadTimeout:  3 * time.Second,
        WriteTimeout: 3 * time.Second,
    })
    
    return ClusterClient.Ping(ctx).Err()
}
```

### **2. 混合缓存管理器**
```go
// pkg/cache/hybrid_cache.go
package cache

import (
    "encoding/json"
    "fmt"
    "time"
    "ryan-mall/pkg/redis"
)

type HybridCache struct {
    memoryCache *ShardedCache
    redisClient *redis.ClusterClient
    strategy    string
}

func NewHybridCache(strategy string) *HybridCache {
    return &HybridCache{
        memoryCache: NewShardedCache(16),
        redisClient: redis.ClusterClient,
        strategy:    strategy,
    }
}

// Get 获取缓存数据
func (hc *HybridCache) Get(key string) (interface{}, bool) {
    // 1. 先从内存缓存获取
    if value, found := hc.memoryCache.Get(key); found {
        return value, true
    }
    
    // 2. 从Redis集群获取
    if hc.redisClient != nil {
        result, err := hc.redisClient.Get(ctx, key).Result()
        if err == nil {
            var value interface{}
            if json.Unmarshal([]byte(result), &value) == nil {
                // 回写到内存缓存
                hc.memoryCache.Set(key, value, 5*time.Minute)
                return value, true
            }
        }
    }
    
    return nil, false
}

// Set 设置缓存数据
func (hc *HybridCache) Set(key string, value interface{}, ttl time.Duration) {
    // 1. 设置到内存缓存
    hc.memoryCache.Set(key, value, ttl)
    
    // 2. 设置到Redis集群
    if hc.redisClient != nil {
        data, _ := json.Marshal(value)
        hc.redisClient.Set(ctx, key, data, ttl)
    }
}
```

### **3. 商品缓存服务**
```go
// internal/service/product_cache.go
package service

import (
    "fmt"
    "time"
    "ryan-mall/internal/model"
    "ryan-mall/pkg/cache"
)

type ProductCacheService struct {
    cache cache.Cache
}

func NewProductCacheService(cache cache.Cache) *ProductCacheService {
    return &ProductCacheService{cache: cache}
}

// GetProduct 获取商品（带缓存）
func (pcs *ProductCacheService) GetProduct(productID uint) (*model.Product, error) {
    key := fmt.Sprintf("product:%d", productID)
    
    // 从缓存获取
    if cached, found := pcs.cache.Get(key); found {
        if product, ok := cached.(*model.Product); ok {
            return product, nil
        }
    }
    
    // 从数据库获取
    product, err := pcs.getProductFromDB(productID)
    if err != nil {
        return nil, err
    }
    
    // 设置缓存
    pcs.cache.Set(key, product, 1*time.Hour)
    return product, nil
}

// InvalidateProduct 失效商品缓存
func (pcs *ProductCacheService) InvalidateProduct(productID uint) {
    key := fmt.Sprintf("product:%d", productID)
    pcs.cache.Delete(key)
}
```

### **4. 购物车Redis存储**
```go
// internal/service/cart_redis.go
package service

import (
    "encoding/json"
    "fmt"
    "time"
    "ryan-mall/internal/model"
    "ryan-mall/pkg/redis"
)

type CartRedisService struct {
    client *redis.ClusterClient
}

func NewCartRedisService() *CartRedisService {
    return &CartRedisService{
        client: redis.ClusterClient,
    }
}

// SaveCart 保存购物车到Redis
func (crs *CartRedisService) SaveCart(userID uint, items []model.CartItem) error {
    key := fmt.Sprintf("cart:user:%d", userID)
    data, err := json.Marshal(items)
    if err != nil {
        return err
    }
    
    return crs.client.Set(ctx, key, data, 24*time.Hour).Err()
}

// GetCart 从Redis获取购物车
func (crs *CartRedisService) GetCart(userID uint) ([]model.CartItem, error) {
    key := fmt.Sprintf("cart:user:%d", userID)
    result, err := crs.client.Get(ctx, key).Result()
    if err != nil {
        return nil, err
    }
    
    var items []model.CartItem
    err = json.Unmarshal([]byte(result), &items)
    return items, err
}

// AddToCart 添加商品到购物车
func (crs *CartRedisService) AddToCart(userID uint, productID uint, quantity int) error {
    key := fmt.Sprintf("cart:user:%d", userID)
    
    // 使用Redis Hash存储
    field := fmt.Sprintf("product:%d", productID)
    return crs.client.HSet(ctx, key, field, quantity).Err()
}
```

## 🎯 **缓存策略配置**

### **热点数据识别**
```go
// pkg/cache/hotspot.go
package cache

import (
    "sync"
    "time"
)

type HotspotDetector struct {
    counters map[string]*AccessCounter
    mutex    sync.RWMutex
    threshold int
}

type AccessCounter struct {
    count     int
    lastAccess time.Time
}

func (hd *HotspotDetector) RecordAccess(key string) bool {
    hd.mutex.Lock()
    defer hd.mutex.Unlock()
    
    counter, exists := hd.counters[key]
    if !exists {
        counter = &AccessCounter{}
        hd.counters[key] = counter
    }
    
    counter.count++
    counter.lastAccess = time.Now()
    
    // 判断是否为热点数据
    return counter.count >= hd.threshold
}
```

### **缓存预热策略**
```go
// internal/service/cache_warmup.go
package service

import (
    "log"
    "ryan-mall/internal/repository"
    "ryan-mall/pkg/cache"
)

type CacheWarmupService struct {
    productRepo repository.ProductRepository
    cache       cache.Cache
}

// WarmupPopularProducts 预热热门商品
func (cws *CacheWarmupService) WarmupPopularProducts() error {
    log.Println("开始预热热门商品缓存...")
    
    // 获取热门商品
    products, err := cws.productRepo.GetPopularProducts(100)
    if err != nil {
        return err
    }
    
    // 预热到缓存
    for _, product := range products {
        key := fmt.Sprintf("product:%d", product.ID)
        cws.cache.Set(key, product, 2*time.Hour)
    }
    
    log.Printf("预热完成，共预热 %d 个商品", len(products))
    return nil
}
```

## 📊 **监控指标**

### **缓存性能监控**
```go
// pkg/cache/metrics.go
package cache

import (
    "sync/atomic"
    "time"
)

type CacheMetrics struct {
    hits        int64
    misses      int64
    sets        int64
    deletes     int64
    errors      int64
    totalTime   int64
    operations  int64
}

func (cm *CacheMetrics) RecordHit() {
    atomic.AddInt64(&cm.hits, 1)
}

func (cm *CacheMetrics) RecordMiss() {
    atomic.AddInt64(&cm.misses, 1)
}

func (cm *CacheMetrics) GetHitRate() float64 {
    hits := atomic.LoadInt64(&cm.hits)
    misses := atomic.LoadInt64(&cm.misses)
    total := hits + misses
    if total == 0 {
        return 0
    }
    return float64(hits) / float64(total)
}
```

## 🔧 **配置文件更新**

### **config.go 配置扩展**
```go
// internal/config/config.go
type Config struct {
    // ... 现有配置
    
    // Redis集群配置
    Redis RedisConfig `mapstructure:"redis"`
    
    // 缓存策略配置
    Cache CacheConfig `mapstructure:"cache"`
}

type RedisConfig struct {
    ClusterEnabled   bool     `mapstructure:"cluster_enabled"`
    ClusterNodes     []string `mapstructure:"cluster_nodes"`
    Password         string   `mapstructure:"password"`
    PoolSize         int      `mapstructure:"pool_size"`
    MinIdleConns     int      `mapstructure:"min_idle_conns"`
    MaxRetries       int      `mapstructure:"max_retries"`
}

type CacheConfig struct {
    Strategy          string `mapstructure:"strategy"`
    MemoryCacheSize   int    `mapstructure:"memory_cache_size"`
    RedisCacheTTL     int    `mapstructure:"redis_cache_ttl"`
    HotDataThreshold  int    `mapstructure:"hot_data_threshold"`
}
```

## 🚀 **部署和启动**

### **1. 启动Redis集群**
```bash
# 启动Redis集群
./start_redis_cluster.sh

# 验证集群状态
docker exec ryan-mall-redis-node-1 redis-cli cluster info
```

### **2. 配置应用程序**
```bash
# 设置环境变量
export REDIS_CLUSTER_ENABLED=true
export REDIS_CLUSTER_NODES=localhost:7001,localhost:7002,localhost:7003

# 启动应用
go run cmd/server/main.go
```

### **3. 验证集成**
```bash
# 测试缓存功能
curl "http://localhost:8080/api/v1/products/1"

# 查看缓存统计
curl "http://localhost:8080/cache/stats"
```

## 📈 **性能优化建议**

### **1. 缓存分层策略**
- **L1缓存**: 内存缓存 (热点数据，1-5分钟TTL)
- **L2缓存**: Redis集群 (温数据，1-24小时TTL)
- **L3存储**: MySQL数据库 (冷数据，持久存储)

### **2. 键命名规范**
```
product:{id}           # 商品详情
user:{id}:profile      # 用户资料
cart:user:{id}         # 购物车
session:{token}        # 用户会话
category:{id}:products # 分类商品列表
```

### **3. 过期策略**
- **商品信息**: 1小时 (频繁更新)
- **用户资料**: 4小时 (较少更新)
- **购物车**: 24小时 (用户会话)
- **分类列表**: 12小时 (相对稳定)

## 🎯 **最佳实践**

1. **渐进式迁移**: 先在非关键功能测试Redis集群
2. **监控告警**: 部署Prometheus监控集群健康状态
3. **故障恢复**: 准备Redis集群故障时的降级方案
4. **性能调优**: 根据业务特点调整连接池和超时参数
5. **数据一致性**: 确保缓存和数据库的数据一致性

---

**Redis集群应用方案已准备就绪，可根据业务需求逐步集成！** 🚀
