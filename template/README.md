# Ryan Mall 前端模板

这是Ryan Mall电商系统的前端模板，提供了完整的用户界面来展示和使用后端API功能。

## 📁 目录结构

```
template/
├── views/                  # HTML页面
│   ├── index.html         # 首页 - 项目介绍和功能展示
│   ├── login.html         # 登录页面 - 用户认证
│   ├── products.html      # 商品列表 - 商品浏览和搜索
│   ├── cart.html          # 购物车页面
│   └── orders.html        # 订单管理页面
├── static/                # 静态资源
│   ├── css/
│   │   └── style.css      # 自定义样式
│   ├── js/
│   │   └── app.js         # 应用工具函数
│   └── images/            # 图片资源
└── admin/                 # 管理后台（待开发）
    ├── dashboard.html     # 管理仪表板
    ├── product-manage.html # 商品管理
    └── order-manage.html  # 订单管理
```

## 🌟 功能特性

### 1. 响应式设计
- 基于Bootstrap 5构建
- 支持桌面端和移动端
- 现代化的UI设计

### 2. 完整的用户流程
- **首页**: 项目介绍、功能展示、技术栈说明
- **登录页面**: 用户认证、演示账户、注册功能
- **商品页面**: 商品浏览、搜索筛选、分页显示
- **购物车**: 商品管理、数量调整、结算功能
- **订单管理**: 订单创建、状态跟踪、历史查询

### 3. 实时API交互
- 与后端API实时通信
- JWT认证集成
- 错误处理和用户反馈
- 购物车状态同步

### 4. 用户体验优化
- 加载状态提示
- 错误信息展示
- 成功操作反馈
- 平滑的页面过渡

## 🚀 快速开始

### 方式一：使用完整启动脚本（推荐）

```bash
# 启动完整服务（前端 + 后端 + 数据库）
./start_with_frontend.sh

# 访问前端界面
open http://localhost:8080
```

### 方式二：使用Docker Compose

```bash
# 启动前端服务
docker compose up -d frontend

# 启动后端服务
go run cmd/server/main.go

# 访问前端界面
open http://localhost:8080
```

### 方式三：独立启动

```bash
# 启动nginx服务器
docker run -d -p 8080:80 -v $(pwd)/template:/usr/share/nginx/html nginx:alpine

# 启动后端API
go run cmd/server/main.go
```

## 📱 页面说明

### 首页 (index.html)
- **功能**: 项目介绍和功能展示
- **特色**: 
  - 英雄区域展示项目亮点
  - 功能模块卡片展示
  - 技术栈介绍
  - 在线演示入口

### 登录页面 (login.html)
- **功能**: 用户认证和注册
- **特色**:
  - 演示账户快速体验
  - 用户注册功能
  - JWT认证集成
  - 响应式设计

### 商品页面 (products.html)
- **功能**: 商品浏览和管理
- **特色**:
  - 商品列表展示
  - 搜索和筛选功能
  - 分页显示
  - 商品详情模态框
  - 购物车集成

## 🎨 设计特色

### 1. 现代化UI
- 渐变色彩搭配
- 卡片式布局
- 图标字体集成
- 悬停动画效果

### 2. 交互体验
- 平滑过渡动画
- 加载状态指示
- 错误提示优化
- 操作反馈及时

### 3. 响应式布局
- 移动端适配
- 弹性网格系统
- 自适应组件
- 触摸友好

## 🔧 技术栈

### 前端技术
- **HTML5**: 语义化标签
- **CSS3**: 现代样式特性
- **JavaScript ES6+**: 原生JS开发
- **Bootstrap 5**: UI框架
- **Bootstrap Icons**: 图标库

### 工具和库
- **Fetch API**: HTTP请求
- **LocalStorage**: 本地存储
- **CSS Grid/Flexbox**: 布局系统
- **CSS Variables**: 主题定制

## 🧪 演示账户

### 管理员账户
- **用户名**: admin
- **密码**: admin123
- **权限**: 商品管理、订单管理

### 普通用户
- **用户名**: user1
- **密码**: password123
- **权限**: 购物、下单

## 📊 API集成

### 认证相关
- `POST /api/v1/login` - 用户登录
- `POST /api/v1/register` - 用户注册
- `GET /api/v1/profile` - 获取用户信息

### 商品相关
- `GET /api/v1/products` - 获取商品列表
- `GET /api/v1/products/:id` - 获取商品详情
- `GET /api/v1/categories` - 获取分类列表

### 购物车相关
- `POST /api/v1/cart` - 添加到购物车
- `GET /api/v1/cart` - 获取购物车
- `PUT /api/v1/cart/:id` - 更新购物车
- `DELETE /api/v1/cart/:id` - 删除购物车项

### 订单相关
- `POST /api/v1/orders` - 创建订单
- `GET /api/v1/orders` - 获取订单列表
- `GET /api/v1/orders/:id` - 获取订单详情

## 🔮 后续计划

### 短期优化
- [ ] 购物车页面完善
- [ ] 订单页面开发
- [ ] 用户个人中心
- [ ] 商品图片上传

### 中期扩展
- [ ] 管理后台界面
- [ ] 数据可视化图表
- [ ] 实时通知系统
- [ ] 移动端PWA

### 长期规划
- [ ] 微前端架构
- [ ] 组件库开发
- [ ] 主题定制系统
- [ ] 国际化支持

## 🤝 贡献指南

1. Fork项目
2. 创建功能分支
3. 提交更改
4. 推送到分支
5. 创建Pull Request

## 📄 许可证

MIT License - 详见 [LICENSE](../LICENSE) 文件
