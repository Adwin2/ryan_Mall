# Ryan Mall - 前后端电商MVP

一个基于Go + Gin + MySQL + Redis的现代化电商系统，专注于前后端产品MVP，提供完整的电商核心功能。

## 🌟 项目特色

### 技术架构
- **后端**: Go 1.19+ + Gin框架 + GORM
- **数据库**: MySQL 8.0 + Redis 7.0
- **前端**: HTML5 + CSS3 + JavaScript + Bootstrap 5
- **部署**: Docker + Docker Compose
- **认证**: JWT Token认证

### 核心功能
- 🛒 **商品管理**: 商品展示、分类管理、搜索筛选
- 👤 **用户系统**: 注册登录、个人中心、权限管理
- 🛍️ **购物车**: 添加商品、数量调整、实时更新
- 📦 **订单系统**: 订单创建、状态跟踪、支付管理
- 📊 **管理后台**: 数据统计、订单管理、系统监控

## 🚀 快速开始

### 环境要求
- Docker & Docker Compose
- Go 1.19+
- Git

### 一键启动
```bash
# 克隆项目
git clone <repository-url>
cd ryan_Mall

# 启动MVP服务
./start_mvp.sh

# 启动AI服务
cd eino-minimal/ && go run ./main.go

# 在新终端启动后端API
SERVER_PORT=8082 go run cmd/server/main.go
```

### 访问地址
- **前端主页**: http://localhost:8080
- **登录页面**: http://localhost:8080/views/login.html
- **商品页面**: http://localhost:8080/views/products.html
- **购物车**: http://localhost:8080/views/cart.html
- **订单管理**: http://localhost:8080/views/orders.html
- **管理后台**: http://localhost:8080/admin/dashboard.html
- **API测试**: http://localhost:8080/views/test-api.html

### 演示账户
- **管理员**: admin / admin123
- **普通用户**: user1 / password123

## 📱 功能演示

### 用户界面
- **主页**: 项目介绍和功能展示
- **登录注册**: 用户认证和账户管理
- **商品浏览**: 商品列表、搜索、分类筛选
- **购物车**: 商品添加、数量调整、价格计算
- **订单管理**: 订单创建、状态跟踪、支付操作

### 管理后台
- **数据仪表板**: 实时数据统计和图表
- **订单管理**: 最近订单查看和管理
- **系统监控**: 服务状态和性能指标

## 🏗️ 系统架构

### MVP架构
```
┌─────────────────┐
│   前端界面 (Web)  │  ← Bootstrap 5 响应式UI
├─────────────────┤
│   API网关层      │  ← Gin HTTP服务器
├─────────────────┤
│   业务逻辑层      │  ← Go Services
├─────────────────┤
│   数据访问层      │  ← GORM + Redis
├─────────────────┤
│   数据存储层      │  ← MySQL + Redis
└─────────────────┘
```

### 端口配置
- **前端服务**: 8080 (Nginx)
- **后端API**: 8081 (Go Gin)
- **MySQL**: 3306
- **Redis**: 6379

## 📊 核心特性

### 前端特性
- **响应式设计**: 完全适配移动端和桌面端
- **现代化UI**: Bootstrap 5 + 自定义CSS
- **实时交互**: JavaScript原生开发，无框架依赖
- **用户体验**: Toast通知、加载状态、错误处理

### 后端特性
- **RESTful API**: 标准REST API设计
- **JWT认证**: 安全的用户认证机制
- **数据缓存**: Redis缓存提升性能
- **错误处理**: 统一的错误响应格式

### 数据库设计
- **用户表**: users (用户基本信息)
- **商品表**: products (商品信息)
- **分类表**: categories (商品分类)
- **购物车**: cart_items (购物车项)
- **订单表**: orders (订单主表)
- **订单项**: order_items (订单详情)

## 🔧 开发指南

### 项目结构
```
ryan_Mall/
├── cmd/server/          # 应用入口
├── internal/           # 内部业务逻辑
│   ├── handler/        # HTTP处理器
│   ├── service/        # 业务服务
│   ├── model/          # 数据模型
│   └── router/         # 路由配置
├── pkg/               # 公共包
│   ├── database/      # 数据库连接
│   ├── redis/         # Redis客户端
│   ├── auth/          # 认证中间件
│   └── utils/         # 工具函数
├── template/          # 前端模板
│   ├── views/         # 页面文件
│   ├── admin/         # 管理后台
│   └── static/        # 静态资源
├── docker/            # Docker配置
└── tests/             # 测试文件
```

### API端点
```
# 用户认证
POST /api/v1/register   # 用户注册
POST /api/v1/login      # 用户登录
GET  /api/v1/profile    # 获取用户信息

# 商品管理
GET  /api/v1/products   # 获取商品列表
GET  /api/v1/products/:id # 获取商品详情
GET  /api/v1/categories # 获取分类列表

# 购物车
POST /api/v1/cart       # 添加到购物车
GET  /api/v1/cart       # 获取购物车
PUT  /api/v1/cart/:id   # 更新购物车项
DELETE /api/v1/cart/:id # 删除购物车项

# 订单管理
POST /api/v1/orders     # 创建订单
GET  /api/v1/orders     # 获取订单列表
GET  /api/v1/orders/:id # 获取订单详情
```

## 🧪 测试

### API测试
访问 http://localhost:8080/views/test-api.html 进行在线API测试

### 功能测试
1. 用户注册登录
2. 商品浏览和搜索
3. 购物车操作
4. 订单创建和管理
5. 管理后台功能

## 🚀 部署指南

### 开发环境
```bash
# 启动MVP服务
./start_mvp.sh

# 启动后端API （或者自行配置使用Air热加载工具）
go run cmd/server/main.go
```

### 生产环境
```bash
# 构建生产镜像
docker build -t ryan-mall:latest .

# 启动生产环境
docker-compose up -d
```

## 📈 项目价值

这个MVP项目展示了：

1. **全栈开发能力** - 前后端完整集成
2. **现代化技术栈** - 使用最新的Web技术
3. **用户体验设计** - 注重交互和视觉效果
4. **工程化实践** - 模块化、组件化开发
5. **生产级质量** - 错误处理、性能优化
6. **可扩展架构** - 易于维护和扩展

## 🤝 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

---

**Ryan Mall** - 现代化电商MVP解决方案
