# Ryan Mall - Go电商系统

一个基于Go语言开发的现代化电商系统，采用分层架构设计，提供完整的用户管理、商品管理、购物车和订单功能。

## 🚀 项目特性

### 核心功能
- **用户系统**：注册、登录、JWT认证、个人信息管理
- **商品系统**：商品CRUD、分类管理、搜索筛选、库存管理
- **购物车系统**：添加商品、数量管理、智能合并、实时验证
- **订单系统**：完整下单流程、支付模拟、状态管理、库存扣减

### 技术特性
- **分层架构**：Handler → Service → Repository 清晰分层
- **RESTful API**：统一的API设计规范
- **JWT认证**：安全的用户认证机制
- **事务支持**：保证数据一致性
- **关联查询**：优化的数据库查询性能
- **参数验证**：完整的请求参数验证

## 🛠️ 技术栈

- **后端框架**：[Gin](https://gin-gonic.com/) - 高性能Go Web框架
- **数据库**：MySQL 8.0+
- **ORM**：[GORM](https://gorm.io/) - Go语言ORM库
- **认证**：JWT (JSON Web Token)
- **配置管理**：Viper
- **日志**：内置日志系统
- **API文档**：RESTful API设计

## 📁 项目结构

```
ryan-mall/
├── cmd/
│   └── server/
│       └── main.go              # 应用程序入口
├── internal/
│   ├── handler/                 # HTTP处理器层
│   │   ├── user_handler.go      # 用户相关API
│   │   ├── product_handler.go   # 商品相关API
│   │   ├── category_handler.go  # 分类相关API
│   │   ├── cart_handler.go      # 购物车相关API
│   │   └── order_handler.go     # 订单相关API
│   ├── service/                 # 业务逻辑层
│   │   ├── user_service.go      # 用户业务逻辑
│   │   ├── product_service.go   # 商品业务逻辑
│   │   ├── category_service.go  # 分类业务逻辑
│   │   ├── cart_service.go      # 购物车业务逻辑
│   │   └── order_service.go     # 订单业务逻辑
│   ├── repository/              # 数据访问层
│   │   ├── user_repository.go   # 用户数据访问
│   │   ├── product_repository.go# 商品数据访问
│   │   ├── category_repository.go# 分类数据访问
│   │   ├── cart_repository.go   # 购物车数据访问
│   │   └── order_repository.go  # 订单数据访问
│   ├── model/                   # 数据模型
│   │   ├── user.go              # 用户模型
│   │   ├── product.go           # 商品模型
│   │   ├── cart.go              # 购物车模型
│   │   └── order.go             # 订单模型
│   └── middleware/              # 中间件
│       └── auth.go              # 认证中间件
├── pkg/                         # 公共包
│   ├── config/                  # 配置管理
│   ├── database/                # 数据库连接
│   ├── jwt/                     # JWT工具
│   └── response/                # 统一响应格式
├── configs/                     # 配置文件
│   └── config.yaml              # 应用配置
├── docs/                        # 文档目录
├── test_*.sh                    # API测试脚本
└── README.md                    # 项目说明
```

## ✅ 开发进度
- [x] 项目初始化
- [x] 数据库设计
- [x] 用户模块（注册、登录、认证、个人信息）
- [x] 商品模块（CRUD、分类、搜索、库存）
- [x] 购物车模块（添加、修改、删除、汇总）
- [x] 订单模块（创建、支付、取消、查询）
- [x] API测试（完整测试脚本）

## 🚦 快速开始

### 环境要求
- Go 1.19+
- MySQL 8.0+
- Docker & Docker Compose
- Git

### 🚀 一键启动（推荐）
```bash
# 使用快速启动菜单
./quick_start.sh

# 或直接启动优化版应用
cd tests/deployment
./start_optimized.sh
```

### 📋 手动安装步骤

1. **克隆项目**
```bash
git clone <repository-url>
cd ryan-mall
```

2. **安装依赖**
```bash
go mod tidy
```

3. **配置数据库**
```bash
# 使用Docker启动MySQL
docker compose up -d mysql

# 或手动创建数据库
mysql -u root -p
CREATE DATABASE ryan_mall CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

4. **运行应用**
```bash
go run cmd/server/main.go
```

6. **验证安装**
```bash
curl http://localhost:8080/ping
```

## 📖 API文档

### 基础信息
- **Base URL**: `http://localhost:8080/api/v1`
- **认证方式**: Bearer Token (JWT)
- **响应格式**: JSON

### 用户相关API
| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| POST | `/register` | 用户注册 | ❌ |
| POST | `/login` | 用户登录 | ❌ |
| GET | `/profile` | 获取个人信息 | ✅ |
| PUT | `/profile` | 更新个人信息 | ✅ |
| POST | `/change-password` | 修改密码 | ✅ |

### 商品相关API
| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | `/products` | 获取商品列表 | ❌ |
| GET | `/products/:id` | 获取商品详情 | ❌ |
| POST | `/products` | 创建商品 | ✅ |
| PUT | `/products/:id` | 更新商品 | ✅ |
| DELETE | `/products/:id` | 删除商品 | ✅ |

### 购物车相关API
| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| POST | `/cart` | 添加商品到购物车 | ✅ |
| GET | `/cart` | 获取购物车 | ✅ |
| PUT | `/cart/:id` | 更新购物车商品数量 | ✅ |
| DELETE | `/cart/:id` | 移除购物车商品 | ✅ |

### 订单相关API
| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| POST | `/orders` | 创建订单 | ✅ |
| GET | `/orders` | 获取订单列表 | ✅ |
| GET | `/orders/:id` | 获取订单详情 | ✅ |
| POST | `/orders/:id/pay` | 支付订单 | ✅ |
| PUT | `/orders/:id/cancel` | 取消订单 | ✅ |

## 🧪 测试和工具

### 测试脚本目录
项目包含完整的测试脚本，按功能分类组织：

```
tests/
├── api/                    # API功能测试
├── performance/            # 性能压力测试
├── monitoring/             # 监控系统测试
├── redis/                  # Redis集群测试
├── deployment/             # 部署启动脚本
├── optimization/           # 系统优化脚本
└── run_all_tests.sh       # 一键测试脚本
```

### 快速测试
```bash
# 运行完整测试套件
cd tests && ./run_all_tests.sh

# 运行API测试
cd tests/api && ./test_api.sh

# 运行性能测试
cd tests/performance && ./test_performance.sh
```

### 监控和Redis集群
```bash
# 启动Redis集群
cd tests/deployment && ./start_redis_cluster.sh

# 启动监控系统
cd tests/deployment && ./start_monitoring.sh

# 测试Redis集群
cd tests/redis && ./simple_redis_cluster_test.sh
```

## 📊 监控面板
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3001 (admin/admin123)
- **AlertManager**: http://localhost:9093

## 🎯 项目亮点
- ✅ **完整的电商功能** - 用户、商品、购物车、订单全流程
- ✅ **高性能优化** - 分片缓存、连接池、并发优化
- ✅ **Redis集群支持** - 分布式缓存和高可用性
- ✅ **监控告警系统** - Prometheus + Grafana 完整监控
- ✅ **完善的测试** - API测试、性能测试、集成测试
- ✅ **一键部署** - Docker Compose 容器化部署
- ✅ **详细文档** - 完整的部署和使用文档
