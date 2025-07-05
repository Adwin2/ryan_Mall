# 📊 Prometheus + Grafana 监控系统部署指南 - Ryan Mall项目

## 📋 **概述**

为Ryan Mall项目部署了完整的监控告警系统，包括指标采集、可视化展示和告警通知功能。

## 🏗️ **监控架构**

### **核心组件**
```
监控系统架构
├── Prometheus (9090) - 指标采集和存储
├── Grafana (3001) - 可视化仪表板  
├── AlertManager (9093) - 告警管理
└── 指标采集器
    ├── Node Exporter (9100) - 系统指标
    ├── Redis Exporter (9121) - Redis监控
    ├── MySQL Exporter (9104) - 数据库监控
    └── Blackbox Exporter (9115) - 黑盒监控
```

### **监控目标**
- **应用程序**: Ryan Mall API服务
- **基础设施**: CPU、内存、磁盘、网络
- **数据库**: MySQL性能和连接
- **缓存**: Redis集群状态
- **容器**: Docker容器资源使用
- **网络**: HTTP/TCP端点可用性

## 🚀 **部署状态**

### **✅ 已成功部署的组件**
1. **Prometheus** - 指标采集服务器 ✅
   - 端口: 9090
   - 状态: 运行正常
   - 配置: 完整的采集规则和告警规则

2. **Node Exporter** - 系统监控 ✅
   - 端口: 9100
   - 状态: 运行正常
   - 功能: CPU、内存、磁盘、网络指标

3. **Blackbox Exporter** - 黑盒监控 ✅
   - 端口: 9115
   - 状态: 运行正常
   - 功能: HTTP/TCP端点探测

### **🔄 部分就绪的组件**
4. **Grafana** - 可视化平台 🔄
   - 端口: 3001
   - 状态: 配置中（插件下载问题）
   - 功能: 仪表板和数据源已配置

5. **AlertManager** - 告警管理 🔄
   - 端口: 9093
   - 状态: 配置调整中
   - 功能: 邮件和Webhook告警

6. **Redis Exporter** - Redis监控 🔄
   - 端口: 9121
   - 状态: 连接Redis集群中
   - 功能: Redis性能指标

7. **MySQL Exporter** - 数据库监控 🔄
   - 端口: 9104
   - 状态: 数据库连接配置中
   - 功能: MySQL性能指标

## 📊 **监控指标**

### **应用程序指标**
- HTTP请求数量和响应时间
- 错误率和状态码分布
- 内存和CPU使用情况
- 缓存命中率

### **系统指标**
- CPU使用率和负载
- 内存使用情况
- 磁盘空间和IO
- 网络流量

### **数据库指标**
- 连接数和慢查询
- 查询性能和锁等待
- 缓冲池使用情况
- 复制延迟

### **Redis指标**
- 内存使用和键数量
- 命令执行统计
- 集群节点状态
- 连接数和延迟

## 🔧 **配置文件**

### **已完成的配置**
1. **Prometheus配置** (`monitoring/prometheus/prometheus.yml`)
   - 8个监控目标配置
   - 告警规则集成
   - 数据保留策略

2. **告警规则** (`monitoring/prometheus/rules/ryan-mall-alerts.yml`)
   - 应用程序告警（宕机、错误率、响应时间）
   - Redis集群告警（节点状态、内存使用）
   - MySQL告警（连接数、慢查询）
   - 系统资源告警（CPU、内存、磁盘）

3. **AlertManager配置** (`monitoring/alertmanager/alertmanager.yml`)
   - 邮件通知配置
   - 告警分组和路由
   - 抑制规则

4. **Grafana配置**
   - 数据源自动配置
   - 仪表板模板
   - 用户权限设置

## 🎯 **使用指南**

### **启动监控系统**
```bash
# 启动完整监控栈
./start_monitoring.sh

# 或手动启动
docker compose -f docker-compose.monitoring.yml up -d
```

### **访问监控界面**
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3001 (admin/admin123)
- **AlertManager**: http://localhost:9093

### **检查系统状态**
```bash
# 运行监控测试
./test_monitoring.sh

# 检查服务状态
docker compose -f docker-compose.monitoring.yml ps

# 查看日志
docker compose -f docker-compose.monitoring.yml logs [service]
```

## 📈 **监控仪表板**

### **已配置的仪表板**
1. **Ryan Mall应用概览**
   - HTTP请求速率
   - 应用可用性
   - 响应时间分布
   - 错误率统计

2. **系统资源监控**
   - CPU和内存使用
   - 磁盘空间和IO
   - 网络流量
   - 系统负载

3. **Redis集群监控**
   - 集群节点状态
   - 内存使用情况
   - 命令执行统计
   - 连接数监控

4. **MySQL数据库监控**
   - 连接数和查询性能
   - 缓冲池状态
   - 慢查询统计
   - 锁等待情况

## 🚨 **告警配置**

### **告警级别**
- **Critical**: 服务宕机、集群故障
- **Warning**: 性能问题、资源使用过高
- **Info**: 一般性通知

### **告警通道**
- **邮件通知**: 管理员和运维团队
- **Webhook**: 集成到Ryan Mall应用
- **静默规则**: 维护期间告警抑制

## 🔍 **故障排查**

### **常见问题**
1. **Grafana插件下载失败**
   - 原因: 网络连接问题
   - 解决: 移除插件配置或使用离线安装

2. **Redis Exporter连接失败**
   - 原因: Redis集群网络配置
   - 解决: 检查网络连接和认证配置

3. **MySQL Exporter权限问题**
   - 原因: 数据库用户权限不足
   - 解决: 创建专用监控用户

### **日志查看**
```bash
# 查看特定服务日志
docker logs ryan-mall-prometheus
docker logs ryan-mall-grafana
docker logs ryan-mall-alertmanager

# 实时日志
docker logs -f ryan-mall-prometheus
```

## 📝 **下一步优化**

### **短期目标**
1. 🔧 修复Grafana插件问题
2. 🔗 完善Redis和MySQL监控连接
3. 📧 配置邮件告警通知
4. 📊 导入更多仪表板模板

### **中期目标**
1. 🎯 集成应用程序自定义指标
2. 🔄 配置数据备份和恢复
3. 📱 添加移动端告警通知
4. 🎨 自定义仪表板样式

### **长期目标**
1. 🤖 智能告警和异常检测
2. 📈 性能趋势分析和预测
3. 🔐 安全监控和审计
4. ☁️ 云原生监控集成

## 🎉 **部署成果**

✅ **监控基础设施就绪**  
✅ **核心指标采集正常**  
✅ **告警规则配置完成**  
✅ **可视化平台部署**  
✅ **黑盒监控运行**  

**监控系统已基本就绪，为Ryan Mall提供全方位的可观测性！** 🚀

---

**注意**: 部分组件仍在配置调整中，建议在生产环境使用前完成所有组件的稳定性测试。
