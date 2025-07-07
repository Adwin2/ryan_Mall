# Ryan Mall

一个基于Go + Gin + MySQL + Redis的现代化电商系统，专注于前后端产品MVP，提供完整的电商核心功能。

## 🌟 项目特色

### 🤖 AI购物助手 (Feature)

- **功能**: 智能购物咨询、商品推荐、平台帮助
- **使用**: 登录后点击右下角AI助手按钮
- **服务**: 基于豆包模型的专业电商AI助手

### 技术架构

- **后端**: Go + Gin框架 + GORM
- **数据库**: MySQL 8.0 + Redis 7.0
- **前端**: HTML5 + CSS3 + JavaScript + Bootstrap 5
- **部署**: Docker + Docker Compose
- **认证**: JWT Token认证

### 核心功能

-  **商品管理**: 商品展示、分类管理、搜索筛选
-  **用户系统**: 注册登录、个人中心、权限管理
-  **购物车**: 添加商品、数量调整、实时更新
-  **订单系统**: 订单创建、状态跟踪、支付管理

## 🚀 快速开始

### 环境要求

- Docker & Docker Compose
- Go 1.19+
- Git

### 启动指南

```bash
# 克隆项目
git clone <repository-url>
cd ryan_Mall

# 启动MVP服务
./start_mvp.sh

# 启动AI服务 (端口8083)
cd eino-minimal/ && go run ./main.go

# 在新终端启动后端API (端口8080，包含前端)
go run cmd/server/main.go
```

### 端口配置

- **8080**: Go后端服务 + 前端静态文件
- **8083**: AI服务 (eino-minimal)
- **3306**: MySQL数据库
- **6379**: Redis缓存

## 🏗️ 系统架构

```txt
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

## 🔧 开发指南

### 项目结构

```text
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

### API测试

访问 `http://localhost:8080/views/test-api.html`进行在线API测试

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
