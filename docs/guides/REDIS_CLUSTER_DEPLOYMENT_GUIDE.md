# 🔗 Redis集群部署指南 - Ryan Mall项目

## 📋 **概述**

Redis集群为Ryan Mall项目提供分布式缓存解决方案，具备以下特性：
- **高可用性**: 3主3从架构，自动故障转移
- **数据分片**: 16384个槽位自动分片存储
- **水平扩展**: 支持动态添加/删除节点
- **性能优化**: 分布式读写，负载均衡

## 🏗️ **架构设计**

### **集群拓扑**
```
Redis集群 (3主3从)
├── 主节点1 (7001) ←→ 从节点4 (7004)
├── 主节点2 (7002) ←→ 从节点5 (7005)  
└── 主节点3 (7003) ←→ 从节点6 (7006)
```

### **端口分配**
| 节点 | Redis端口 | 集群总线端口 | 角色 |
|------|----------|-------------|------|
| node-1 | 7001 | 17001 | 主节点 |
| node-2 | 7002 | 17002 | 主节点 |
| node-3 | 7003 | 17003 | 主节点 |
| node-4 | 7004 | 17004 | 从节点 |
| node-5 | 7005 | 17005 | 从节点 |
| node-6 | 7006 | 17006 | 从节点 |

### **数据分片策略**
- **槽位范围**: 0-16383 (共16384个槽位)
- **分片算法**: CRC16(key) % 16384
- **主节点1**: 槽位 0-5460
- **主节点2**: 槽位 5461-10922  
- **主节点3**: 槽位 10923-16383

## 🚀 **快速部署**

### **1. 启动Redis集群**
```bash
# 一键启动Redis集群
./start_redis_cluster.sh
```

### **2. 验证集群状态**
```bash
# 检查集群状态
docker exec ryan-mall-redis-node-1 redis-cli cluster info

# 查看集群节点
docker exec ryan-mall-redis-node-1 redis-cli cluster nodes
```

### **3. 测试集群功能**
```bash
# 运行集群测试
./test_redis_cluster.sh

# 性能对比测试
./redis_cluster_performance_test.sh
```

## 🔧 **配置详解**

### **Redis节点配置**
```conf
# 基础配置
port 6379
bind 0.0.0.0
protected-mode no

# 集群配置
cluster-enabled yes
cluster-config-file nodes-6379.conf
cluster-node-timeout 15000
cluster-announce-ip 172.20.0.11
cluster-announce-port 6379
cluster-announce-bus-port 16379

# 性能优化
maxmemory 512mb
maxmemory-policy allkeys-lru
tcp-keepalive 300
tcp-backlog 511
maxclients 10000

# 持久化配置
appendonly yes
appendfilename "appendonly.aof"
appendfsync everysec
```

### **应用程序配置**
```bash
# 环境变量配置
export REDIS_CLUSTER_ENABLED=true
export REDIS_CLUSTER_NODES=localhost:7001,localhost:7002,localhost:7003,localhost:7004,localhost:7005,localhost:7006
```

## 📊 **性能特性**

### **集群优势**
1. **数据分片**: 数据自动分布到多个节点
2. **高可用**: 主节点故障时从节点自动接管
3. **负载均衡**: 读写请求分布到不同节点
4. **水平扩展**: 支持动态添加节点

### **性能指标**
- **QPS**: 单节点 50,000+ ops/sec
- **延迟**: 平均 < 1ms (局域网)
- **内存**: 每节点 512MB (可调整)
- **连接数**: 每节点 10,000 连接

### **vs 单机Redis对比**
| 特性 | 单机Redis | Redis集群 | 优势 |
|------|----------|----------|------|
| **容量** | 单机内存限制 | 多节点扩展 | 🟢 无限扩展 |
| **可用性** | 单点故障 | 自动故障转移 | 🟢 99.9%+ |
| **性能** | 单机性能 | 分布式性能 | 🟢 线性扩展 |
| **复杂度** | 简单 | 相对复杂 | 🟡 需要运维 |

## 🛠️ **运维管理**

### **集群监控**
```bash
# 查看集群信息
redis-cli -h localhost -p 7001 cluster info

# 查看节点状态
redis-cli -h localhost -p 7001 cluster nodes

# 查看槽位分布
redis-cli -h localhost -p 7001 cluster slots
```

### **故障处理**
```bash
# 检查节点健康
redis-cli -h localhost -p 7001 ping

# 手动故障转移
redis-cli -h localhost -p 7004 cluster failover

# 重新分片
redis-cli --cluster reshard localhost:7001
```

### **数据备份**
```bash
# 备份所有节点数据
for port in 7001 7002 7003; do
    redis-cli -h localhost -p $port bgsave
done

# 备份配置文件
cp docker/redis-cluster/*.conf /backup/redis-cluster/
```

## 🎯 **应用集成**

### **Go客户端配置**
```go
// 创建Redis集群客户端
clusterClient := redis.NewClusterClient(&redis.ClusterOptions{
    Addrs: []string{
        "localhost:7001",
        "localhost:7002", 
        "localhost:7003",
        "localhost:7004",
        "localhost:7005",
        "localhost:7006",
    },
    Password: "",
    PoolSize: 100,
    MinIdleConns: 10,
    MaxRetries: 3,
    MaxRedirects: 8,
    RouteByLatency: true,
    RouteRandomly: true,
})
```

### **缓存策略**
```go
// 商品缓存
func CacheProduct(productID uint, product *Product) error {
    key := fmt.Sprintf("product:%d", productID)
    data, _ := json.Marshal(product)
    return clusterClient.Set(ctx, key, data, 1*time.Hour).Err()
}

// 用户会话缓存
func CacheUserSession(userID uint, session *Session) error {
    key := fmt.Sprintf("session:%d", userID)
    data, _ := json.Marshal(session)
    return clusterClient.Set(ctx, key, data, 24*time.Hour).Err()
}
```

## 🔍 **故障排查**

### **常见问题**

#### **1. 集群初始化失败**
```bash
# 检查节点是否启动
docker ps | grep redis-node

# 检查网络连接
docker exec ryan-mall-redis-node-1 redis-cli -h 172.20.0.12 -p 6379 ping

# 重新初始化集群
docker-compose -f docker-compose.redis-cluster.yml restart redis-cluster-manager
```

#### **2. 节点无法加入集群**
```bash
# 清理节点数据
docker exec ryan-mall-redis-node-1 redis-cli flushall
docker exec ryan-mall-redis-node-1 redis-cli cluster reset

# 重新创建集群
redis-cli --cluster create 172.20.0.11:6379 172.20.0.12:6379 172.20.0.13:6379 172.20.0.14:6379 172.20.0.15:6379 172.20.0.16:6379 --cluster-replicas 1
```

#### **3. 数据不一致**
```bash
# 检查主从同步状态
redis-cli -h localhost -p 7001 info replication

# 手动同步
redis-cli -h localhost -p 7004 cluster replicate <master-node-id>
```

## 📈 **性能优化**

### **内存优化**
```conf
# 内存策略
maxmemory-policy allkeys-lru
maxmemory-samples 5

# 压缩配置
hash-max-ziplist-entries 512
hash-max-ziplist-value 64
```

### **网络优化**
```conf
# TCP优化
tcp-keepalive 300
tcp-backlog 511
timeout 0

# 客户端优化
maxclients 10000
```

### **持久化优化**
```conf
# AOF优化
appendonly yes
appendfsync everysec
no-appendfsync-on-rewrite yes
auto-aof-rewrite-percentage 100
auto-aof-rewrite-min-size 64mb
```

## 🎉 **部署完成**

Redis集群部署完成后，你将获得：

✅ **高可用分布式缓存系统**  
✅ **自动故障转移能力**  
✅ **数据分片和负载均衡**  
✅ **水平扩展能力**  
✅ **完整的监控和管理工具**  

### **下一步优化方向**
1. 🔴 **应用层集成** - 在Ryan Mall中集成Redis集群
2. 🟡 **监控告警** - 部署Prometheus监控
3. 🔵 **自动化运维** - K8s部署和自动扩缩容
4. 🟢 **性能调优** - 根据业务场景优化配置

**Redis集群已就绪，为Ryan Mall提供企业级分布式缓存能力！** 🚀
