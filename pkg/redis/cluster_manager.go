package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// ClusterManager Redis集群管理器
type ClusterManager struct {
	client *redis.ClusterClient
	ctx    context.Context
}

// NewClusterManager 创建Redis集群管理器
func NewClusterManager(addrs []string, password string) *ClusterManager {
	rdb := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    addrs,
		Password: password,
		
		// 连接池配置
		PoolSize:     200,  // 增加连接池大小
		MinIdleConns: 20,   // 最小空闲连接
		MaxRetries:   3,    // 最大重试次数
		
		// 集群特定配置
		MaxRedirects:   8,     // 最大重定向次数
		ReadOnly:       false, // 允许写操作
		RouteByLatency: true,  // 根据延迟路由
		RouteRandomly:  true,  // 随机路由
		
		// 超时配置
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
		
		// 空闲连接检查
		IdleTimeout:        300 * time.Second,
		IdleCheckFrequency: 60 * time.Second,
	})

	return &ClusterManager{
		client: rdb,
		ctx:    context.Background(),
	}
}

// GetClient 获取集群客户端
func (cm *ClusterManager) GetClient() *redis.ClusterClient {
	return cm.client
}

// Ping 检查集群连接
func (cm *ClusterManager) Ping() error {
	return cm.client.Ping(cm.ctx).Err()
}

// GetClusterInfo 获取集群信息
func (cm *ClusterManager) GetClusterInfo() (map[string]interface{}, error) {
	info := make(map[string]interface{})
	
	// 获取集群节点信息
	nodes, err := cm.client.ClusterNodes(cm.ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster nodes: %w", err)
	}
	info["nodes"] = nodes
	
	// 获取集群状态
	clusterInfo, err := cm.client.ClusterInfo(cm.ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster info: %w", err)
	}
	info["cluster_info"] = clusterInfo
	
	// 获取集群槽位信息
	slots, err := cm.client.ClusterSlots(cm.ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster slots: %w", err)
	}
	info["slots_count"] = len(slots)
	info["slots_detail"] = slots
	
	return info, nil
}

// GetClusterStats 获取集群统计信息
func (cm *ClusterManager) GetClusterStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// 获取连接池统计
	poolStats := cm.client.PoolStats()
	stats["pool_stats"] = map[string]interface{}{
		"hits":         poolStats.Hits,
		"misses":       poolStats.Misses,
		"timeouts":     poolStats.Timeouts,
		"total_conns":  poolStats.TotalConns,
		"idle_conns":   poolStats.IdleConns,
		"stale_conns":  poolStats.StaleConns,
	}
	
	// 获取集群节点数量
	nodes, err := cm.client.ClusterNodes(cm.ctx).Result()
	if err != nil {
		return nil, err
	}
	
	// 解析节点信息
	masterCount := 0
	slaveCount := 0
	for _, line := range parseClusterNodes(nodes) {
		if containsString(line, "master") {
			masterCount++
		} else if containsString(line, "slave") {
			slaveCount++
		}
	}
	
	stats["cluster_stats"] = map[string]interface{}{
		"master_nodes": masterCount,
		"slave_nodes":  slaveCount,
		"total_nodes":  masterCount + slaveCount,
	}
	
	return stats, nil
}

// parseClusterNodes 解析集群节点信息
func parseClusterNodes(nodes string) []string {
	var lines []string
	current := ""
	for _, char := range nodes {
		if char == '\n' {
			if current != "" {
				lines = append(lines, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

// containsString 检查字符串是否包含子字符串
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr) >= 0
}

// findSubstring 查找子字符串位置
func findSubstring(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	if len(s) < len(substr) {
		return -1
	}
	
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

// Close 关闭集群连接
func (cm *ClusterManager) Close() error {
	return cm.client.Close()
}

// DistributedCacheManager 分布式缓存管理器
type DistributedCacheManager struct {
	cluster *ClusterManager
}

// NewDistributedCacheManager 创建分布式缓存管理器
func NewDistributedCacheManager(cluster *ClusterManager) *DistributedCacheManager {
	return &DistributedCacheManager{
		cluster: cluster,
	}
}

// Set 设置缓存
func (dcm *DistributedCacheManager) Set(key string, value interface{}, expiration time.Duration) error {
	return dcm.cluster.client.Set(dcm.cluster.ctx, key, value, expiration).Err()
}

// Get 获取缓存
func (dcm *DistributedCacheManager) Get(key string) (string, error) {
	return dcm.cluster.client.Get(dcm.cluster.ctx, key).Result()
}

// Del 删除缓存
func (dcm *DistributedCacheManager) Del(keys ...string) error {
	return dcm.cluster.client.Del(dcm.cluster.ctx, keys...).Err()
}

// Exists 检查键是否存在
func (dcm *DistributedCacheManager) Exists(keys ...string) (int64, error) {
	return dcm.cluster.client.Exists(dcm.cluster.ctx, keys...).Result()
}

// Expire 设置过期时间
func (dcm *DistributedCacheManager) Expire(key string, expiration time.Duration) error {
	return dcm.cluster.client.Expire(dcm.cluster.ctx, key, expiration).Err()
}

// MGet 批量获取
func (dcm *DistributedCacheManager) MGet(keys ...string) ([]interface{}, error) {
	return dcm.cluster.client.MGet(dcm.cluster.ctx, keys...).Result()
}

// MSet 批量设置
func (dcm *DistributedCacheManager) MSet(pairs ...interface{}) error {
	return dcm.cluster.client.MSet(dcm.cluster.ctx, pairs...).Err()
}

// Incr 原子递增
func (dcm *DistributedCacheManager) Incr(key string) (int64, error) {
	return dcm.cluster.client.Incr(dcm.cluster.ctx, key).Result()
}

// Decr 原子递减
func (dcm *DistributedCacheManager) Decr(key string) (int64, error) {
	return dcm.cluster.client.Decr(dcm.cluster.ctx, key).Result()
}

// HSet 设置哈希字段
func (dcm *DistributedCacheManager) HSet(key string, values ...interface{}) error {
	return dcm.cluster.client.HSet(dcm.cluster.ctx, key, values...).Err()
}

// HGet 获取哈希字段
func (dcm *DistributedCacheManager) HGet(key, field string) (string, error) {
	return dcm.cluster.client.HGet(dcm.cluster.ctx, key, field).Result()
}

// HGetAll 获取所有哈希字段
func (dcm *DistributedCacheManager) HGetAll(key string) (map[string]string, error) {
	return dcm.cluster.client.HGetAll(dcm.cluster.ctx, key).Result()
}

// HDel 删除哈希字段
func (dcm *DistributedCacheManager) HDel(key string, fields ...string) error {
	return dcm.cluster.client.HDel(dcm.cluster.ctx, key, fields...).Err()
}
