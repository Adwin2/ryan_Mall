# 🧪 Ryan Mall 测试脚本目录

本目录包含Ryan Mall项目的所有测试脚本，按功能分类组织。

## 📁 **目录结构**

```
tests/
├── api/                    # API功能测试
├── performance/            # 性能压力测试
├── monitoring/             # 监控系统测试
├── redis/                  # Redis集群测试
├── deployment/             # 部署启动脚本
├── optimization/           # 系统优化脚本
└── README.md              # 本文件
```

## 🔧 **API测试脚本** (`api/`)

### **基础API测试**
- `test_api.sh` - 完整API功能测试套件
- `test_product_api.sh` - 商品API专项测试
- `test_cart_api.sh` - 购物车API专项测试
- `test_order_api.sh` - 订单API专项测试

### **增强功能测试**
- `test_enhanced_features.sh` - 高级功能测试（搜索、WebSocket等）

**使用方法：**
```bash
cd tests/api
./test_api.sh              # 运行完整API测试
./test_product_api.sh      # 测试商品相关API
```

## ⚡ **性能测试脚本** (`performance/`)

### **并发性能测试**
- `test_performance.sh` - 基础性能测试
- `test_concurrent_performance.sh` - 并发性能测试
- `simple_concurrent_test.sh` - 简单并发测试
- `enhanced_stress_test.sh` - 增强压力测试
- `extreme_concurrent_test.sh` - 极限并发测试

### **缓存性能测试**
- `test_cache_performance.sh` - 缓存性能专项测试

### **性能分析工具**
- `performance_analysis.sh` - 性能分析脚本
- `monitor_performance.sh` - 性能监控脚本

**使用方法：**
```bash
cd tests/performance
./test_performance.sh      # 基础性能测试
./enhanced_stress_test.sh  # 压力测试
./performance_analysis.sh  # 性能分析
```

## 📊 **监控测试脚本** (`monitoring/`)

- `test_monitoring.sh` - Prometheus + Grafana监控系统测试

**使用方法：**
```bash
cd tests/monitoring
./test_monitoring.sh       # 测试监控系统功能
```

## 🔴 **Redis测试脚本** (`redis/`)

### **集群功能测试**
- `test_redis_cluster.sh` - 完整Redis集群测试
- `simple_redis_cluster_test.sh` - 简化集群测试

### **性能对比测试**
- `redis_cluster_performance_test.sh` - Redis集群性能测试
- `redis_vs_memory_performance.sh` - Redis vs 内存缓存性能对比

**使用方法：**
```bash
cd tests/redis
./simple_redis_cluster_test.sh     # 快速集群测试
./redis_vs_memory_performance.sh   # 性能对比测试
```

## 🚀 **部署脚本** (`deployment/`)

### **服务启动脚本**
- `start_optimized.sh` - 启动优化版Ryan Mall服务
- `start_monitoring.sh` - 启动监控系统
- `start_redis_cluster.sh` - 启动Redis集群

**使用方法：**
```bash
cd tests/deployment
./start_optimized.sh       # 启动优化版应用
./start_monitoring.sh      # 启动监控系统
./start_redis_cluster.sh   # 启动Redis集群
```

## ⚙️ **优化脚本** (`optimization/`)

### **系统优化**
- `system_network_optimization.sh` - 系统网络优化
- `user_level_optimization.sh` - 用户级优化
- `go_runtime_env.sh` - Go运行时环境优化

**使用方法：**
```bash
cd tests/optimization
./system_network_optimization.sh   # 系统优化
./user_level_optimization.sh       # 用户优化
```

## 🎯 **快速测试指南**

### **完整功能测试**
```bash
# 1. 启动服务
cd tests/deployment
./start_optimized.sh

# 2. 运行API测试
cd ../api
./test_api.sh

# 3. 运行性能测试
cd ../performance
./test_performance.sh
```

### **Redis集群测试**
```bash
# 1. 启动Redis集群
cd tests/deployment
./start_redis_cluster.sh

# 2. 测试集群功能
cd ../redis
./simple_redis_cluster_test.sh
```

### **监控系统测试**
```bash
# 1. 启动监控系统
cd tests/deployment
./start_monitoring.sh

# 2. 测试监控功能
cd ../monitoring
./test_monitoring.sh
```

## 📋 **测试检查清单**

### **开发环境测试**
- [ ] API功能测试通过
- [ ] 基础性能测试通过
- [ ] 缓存功能正常
- [ ] 数据库连接正常

### **集成测试**
- [ ] Redis集群部署成功
- [ ] 监控系统运行正常
- [ ] 负载均衡配置正确
- [ ] 告警规则生效

### **性能测试**
- [ ] 并发测试达标
- [ ] 响应时间符合要求
- [ ] 资源使用合理
- [ ] 缓存命中率正常

## 🔧 **故障排查**

### **测试失败处理**
1. 检查服务是否正常启动
2. 确认端口没有被占用
3. 查看相关服务日志
4. 检查网络连接状态

### **常用调试命令**
```bash
# 检查服务状态
docker ps
docker compose ps

# 查看日志
docker logs [container_name]
docker compose logs [service_name]

# 检查端口占用
lsof -i :8080
netstat -tulpn | grep :8080
```

## 📝 **注意事项**

1. **执行权限**: 确保所有脚本都有执行权限
2. **依赖检查**: 运行前确保相关服务已启动
3. **端口冲突**: 注意避免端口冲突
4. **资源限制**: 性能测试可能消耗大量资源
5. **网络环境**: 某些测试需要网络连接

## 🎯 **最佳实践**

1. **测试顺序**: 先功能测试，再性能测试
2. **环境隔离**: 不同测试使用不同环境
3. **结果记录**: 保存测试结果用于对比
4. **定期执行**: 建立定期测试机制
5. **持续改进**: 根据测试结果优化系统

---

**测试脚本已分类整理，项目结构更加清晰！** 🚀
