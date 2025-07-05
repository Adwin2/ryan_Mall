# Ryan Mall 开发环境搭建指南

## 环境要求

- Go 1.19+
- MySQL 8.0+
- Redis 6.0+
- Docker & Docker Compose (推荐)

## 方式一：使用Docker Compose（推荐）

### 1. 安装Docker和Docker Compose

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install docker.io docker-compose

# 启动Docker服务
sudo systemctl start docker
sudo systemctl enable docker

# 将当前用户添加到docker组（避免每次使用sudo）
sudo usermod -aG docker $USER
# 注意：需要重新登录或重启终端才能生效
```

### 2. 启动服务

```bash
# 启动MySQL和Redis
docker-compose up -d mysql redis

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs mysql
```

### 3. 验证服务

```bash
# 测试MySQL连接
docker exec -it ryan-mall-mysql mysql -uroot -p123456 -e "SHOW DATABASES;"

# 测试Redis连接
docker exec -it ryan-mall-redis redis-cli ping
```

## 方式二：本地安装

### 1. 安装MySQL

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install mysql-server

# 启动MySQL服务
sudo systemctl start mysql
sudo systemctl enable mysql

# 安全配置
sudo mysql_secure_installation
```

### 2. 创建数据库和用户

```sql
-- 登录MySQL
sudo mysql -u root -p

-- 创建数据库
CREATE DATABASE ryan_mall CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 创建用户（如果需要）
CREATE USER 'ryan_mall'@'localhost' IDENTIFIED BY '123456';
GRANT ALL PRIVILEGES ON ryan_mall.* TO 'ryan_mall'@'localhost';
FLUSH PRIVILEGES;

-- 退出
EXIT;
```

### 3. 安装Redis

```bash
# Ubuntu/Debian
sudo apt install redis-server

# 启动Redis服务
sudo systemctl start redis-server
sudo systemctl enable redis-server

# 测试Redis
redis-cli ping
```

## 配置环境变量

创建 `.env` 文件（复制 `.env.example`）：

```bash
cp .env.example .env
```

根据你的实际配置修改 `.env` 文件：

```env
# 服务器配置
SERVER_PORT=8080
GIN_MODE=debug

# 数据库配置
DB_HOST=localhost
DB_PORT=3306
DB_USERNAME=root
DB_PASSWORD=123456
DB_NAME=ryan_mall

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT配置
JWT_SECRET=ryan-mall-secret-key
JWT_EXPIRE_HOURS=24
```

## 启动应用

### 1. 安装依赖

```bash
go mod tidy
```

### 2. 启动服务器

```bash
go run cmd/server/main.go
```

### 3. 初始化测试数据（可选）

```bash
go run migrations/seed_data.go
```

## 验证安装

### 1. 健康检查

```bash
curl http://localhost:8080/ping
```

预期响应：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "pong",
    "version": "1.0.0"
  }
}
```

### 2. API测试

```bash
curl http://localhost:8080/api/v1/test
```

预期响应：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "API v1 is working!"
  }
}
```

## 常见问题

### 1. MySQL连接被拒绝

- 检查MySQL服务是否启动：`sudo systemctl status mysql`
- 检查用户名密码是否正确
- 检查数据库是否存在

### 2. 端口被占用

- 检查端口占用：`sudo netstat -tlnp | grep :8080`
- 修改配置文件中的端口号

### 3. 权限问题

- 确保MySQL用户有足够的权限
- 检查文件权限

## 开发工具推荐

- **数据库管理**: phpMyAdmin (http://localhost:8081) 或 MySQL Workbench
- **API测试**: Postman 或 curl
- **代码编辑器**: VS Code + Go扩展
