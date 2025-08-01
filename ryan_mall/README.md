# 🏗️ Ryan Mall 微服务架构 (面试重点MVP)

## 🎯 项目概述

基于微服务架构的电商平台，专注于**面试常考技术点**的深度实现，包括秒杀系统、分布式事务、缓存策略、限流熔断等核心功能。

## 🔥 面试亮点技术栈

### 核心技术 (必考点)
- **微服务**: Go + Gin + gRPC
- **数据库**: MySQL读写分离 + 分库分表
- **缓存**: Redis集群 + 本地缓存
- **消息队列**: Kafka + 异步处理
- **搜索**: Elasticsearch
- **监控**: Prometheus + Jaeger链路追踪

## 🏛️ DDD架构设计

### 领域划分（Domain Boundaries）
```
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│   User Domain   │  │ Product Domain  │  │  Order Domain   │
│   (用户王国)     │  │  (商品王国)     │  │  (订单王国)     │
├─────────────────┤  ├─────────────────┤  ├─────────────────┤
│ • User Entity   │  │ • Product Entity│  │ • Order Entity  │
│ • Profile VO    │  │ • Price VO      │  │ • OrderItem VO  │
│ • UserRepo      │  │ • ProductRepo   │  │ • OrderRepo     │
│ • AuthService   │  │ • InventoryServ │  │ • OrderService  │
└─────────────────┘  └─────────────────┘  └─────────────────┘
         │                     │                     │
         └─────────────────────┼─────────────────────┘
                               │
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│ Seckill Domain  │  │ Payment Domain  │  │   Shared Kernel │
│  (秒杀王国)     │  │  (支付王国)     │  │   (共享内核)     │
├─────────────────┤  ├─────────────────┤  ├─────────────────┤
│ • Activity Entity│ │ • Payment Entity│  │ • Events        │
│ • Inventory VO  │  │ • Amount VO     │  │ • Common VOs    │
│ • SeckillRepo   │  │ • PaymentRepo   │  │ • Base Types    │
│ • LimitService  │  │ • PaymentServ   │  │ • Utils         │
└─────────────────┘  └─────────────────┘  └─────────────────┘
```

## 🚀 技术栈

- **语言**: Go 1.21+
- **框架**: Gin + gRPC
- **数据库**: MySQL + Redis
- **消息队列**: Kafka
- **流处理**: Flink
- **服务发现**: Consul
- **监控**: Prometheus + Grafana + Jaeger
- **测试**: TDD + Testify + Docker Test Containers

## 📁 项目结构

```
ryan_mall/
├── cmd/                          # 应用入口
│   ├── user-service/            # 用户服务
│   ├── product-service/         # 商品服务
│   ├── order-service/           # 订单服务
│   ├── seckill-service/         # 秒杀服务
│   ├── payment-service/         # 支付服务
│   └── api-gateway/             # API网关
├── internal/                     # 内部代码
│   ├── user/                    # 用户领域
│   │   ├── domain/              # 领域层
│   │   │   ├── entity/          # 实体
│   │   │   ├── valueobject/     # 值对象
│   │   │   ├── repository/      # 仓储接口
│   │   │   └── service/         # 领域服务
│   │   ├── application/         # 应用层
│   │   │   ├── command/         # 命令
│   │   │   ├── query/           # 查询
│   │   │   └── service/         # 应用服务
│   │   ├── infrastructure/      # 基础设施层
│   │   │   ├── repository/      # 仓储实现
│   │   │   ├── grpc/            # gRPC实现
│   │   │   └── http/            # HTTP实现
│   │   └── interfaces/          # 接口层
│   │       ├── grpc/            # gRPC接口
│   │       └── http/            # HTTP接口
│   ├── product/                 # 商品领域 (同上结构)
│   ├── order/                   # 订单领域 (同上结构)
│   ├── seckill/                 # 秒杀领域 (同上结构)
│   ├── payment/                 # 支付领域 (同上结构)
│   └── shared/                  # 共享内核
│       ├── domain/              # 共享领域对象
│       ├── infrastructure/      # 共享基础设施
│       └── events/              # 领域事件
├── pkg/                         # 公共包
│   ├── config/                  # 配置管理
│   ├── database/                # 数据库连接
│   ├── kafka/                   # Kafka客户端
│   ├── redis/                   # Redis客户端
│   ├── consul/                  # 服务发现
│   ├── monitoring/              # 监控组件
│   └── testing/                 # 测试工具
├── api/                         # API定义
│   ├── proto/                   # gRPC定义
│   └── openapi/                 # OpenAPI文档
├── deployments/                 # 部署配置
│   ├── docker/                  # Docker配置
│   ├── k8s/                     # Kubernetes配置
│   └── docker-compose.yml       # 本地开发环境
├── scripts/                     # 脚本
│   ├── build.sh                 # 构建脚本
│   ├── test.sh                  # 测试脚本
│   └── proto-gen.sh             # 代码生成
├── tests/                       # 测试
│   ├── unit/                    # 单元测试
│   ├── integration/             # 集成测试
│   └── e2e/                     # 端到端测试
└── docs/                        # 文档
    ├── architecture.md          # 架构文档
    ├── api.md                   # API文档
    └── deployment.md            # 部署文档
```

## 🔄 TDD开发流程

### Red-Green-Refactor循环
1. **🔴 Red**: 写一个失败的测试
2. **🟢 Green**: 写最少的代码让测试通过
3. **🔵 Refactor**: 重构代码，保持测试通过

### 测试金字塔
```
        /\
       /  \
      / E2E \     ← 少量端到端测试
     /______\
    /        \
   /Integration\ ← 适量集成测试
  /__________\
 /            \
/  Unit Tests  \   ← 大量单元测试
/______________\
```

## 🎯 第一个TDD循环：用户注册

我们从最简单的用户注册功能开始，体验完整的TDD + DDD流程。

### 业务需求
- 用户可以通过用户名、邮箱、密码注册
- 用户名和邮箱必须唯一
- 密码需要加密存储
- 注册成功返回用户ID

### 下一步
运行以下命令开始第一个TDD循环：
```bash
cd ryan_mall
go mod init ryan-mall-microservices
```

## 🚀 快速启动

### 前置条件

1. **Go 1.21+** - 确保已安装Go语言环境
2. **MySQL 8.0+** - 数据库服务
3. **Redis 6.0+** - 缓存服务
4. **Git** - 版本控制

### 环境准备

```bash
# 1. 克隆项目
git clone <repository-url>
cd ryan_mall

# 2. 复制环境配置
cp .env.example .env

# 3. 编辑配置文件（根据你的环境调整）
vim .env

# 4. 下载依赖
go mod tidy
```

### 数据库初始化

```bash
# 1. 启动MySQL和Redis（使用Docker）
docker-compose -f deployments/docker-compose.yml up -d mysql redis

# 2. 等待数据库启动完成
sleep 10

# 3. 数据库会自动初始化（通过init.sql脚本）
```

### 方式一：使用测试脚本（推荐）

```bash
# 1. 构建并启动核心服务
./test-services.sh start

# 2. 查看服务状态
./test-services.sh status

# 3. 测试API
./test-services.sh test

# 4. 查看日志
./test-services.sh logs user     # 用户服务日志
./test-services.sh logs gateway  # 网关服务日志

# 5. 停止所有服务
./test-services.sh stop
```

### 方式二：使用启动脚本

```bash
# 1. 快速启动核心服务
./start.sh quick

# 2. 启动单个服务
./start.sh --build gateway    # 构建并启动网关
./start.sh user              # 启动用户服务
./start.sh seckill           # 启动秒杀服务

# 3. 启动所有服务
./start.sh --build all

# 4. 停止所有服务
./start.sh stop
```

### 方式二：使用Makefile

```bash
# 构建所有服务
make build

# 快速启动核心服务
make quick

# 启动单个服务
make run-gateway
make run-user
make run-seckill

# 查看服务状态
make status
make health
```

### 方式三：使用统一启动器

```bash
# 构建并启动所有服务
go run cmd/main.go -service=all -build

# 启动单个服务
go run cmd/main.go -service=gateway -build
go run cmd/main.go -service=user
go run cmd/main.go -service=seckill
```

### 方式四：手动启动

```bash
# 1. 构建服务
go build -o bin/gateway ./cmd/gateway
go build -o bin/user-service ./cmd/user
go build -o bin/seckill-service ./cmd/seckill-service

# 2. 启动服务
./bin/gateway &
./bin/user-service &
./bin/seckill-service &
```

## 📋 服务端口

| 服务 | 端口 | 健康检查 | 状态 |
|------|------|----------|------|
| API Gateway | 8080 | http://localhost:8080/health | ✅ 已实现 |
| User Service | 8081 | http://localhost:8081/health | ✅ 已实现 |
| Product Service | 8082 | http://localhost:8082/health | 🚧 开发中 |
| Order Service | 8083 | http://localhost:8083/health | 🚧 开发中 |
| Seckill Service | 8084 | http://localhost:8084/health | 🚧 开发中 |
| Payment Service | 8085 | http://localhost:8085/health | 🚧 开发中 |

## 🧪 API测试示例

### 用户服务API

```bash
# 1. 用户注册
curl -X POST http://localhost:8081/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "Test123!@#"
  }'

# 2. 用户登录
curl -X POST http://localhost:8081/api/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "Test123!@#"
  }'

# 3. 获取用户信息（需要JWT令牌）
curl -X GET http://localhost:8081/api/v1/users/{user_id} \
  -H "Authorization: Bearer {access_token}"

# 4. 获取用户列表
curl -X GET "http://localhost:8081/api/v1/users?page=1&page_size=10"
```

### 网关服务API

```bash
# 1. 健康检查
curl http://localhost:8080/health

# 2. 通过网关访问用户服务（代理）
curl -X POST http://localhost:8080/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "gatewayuser",
    "email": "gateway@example.com",
    "password": "Test123!@#"
  }'
```

### 监控和指标

```bash
# 1. Prometheus指标
curl http://localhost:8081/metrics  # 用户服务指标
curl http://localhost:8080/metrics  # 网关服务指标

# 2. 服务发现
curl http://localhost:8080/gateway/services
```

## 🔧 开发环境设置

```bash
# 1. 复制环境配置
cp .env.example .env

# 2. 编辑配置文件
vim .env

# 3. 设置开发环境
make dev-setup

# 4. 下载依赖
make deps
```

让我们开始编写第一个测试！
