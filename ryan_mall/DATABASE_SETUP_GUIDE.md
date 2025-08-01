# 数据库设置指南

## 问题描述

您遇到的错误是：
```
Error 1698 (28000): Access denied for user 'root'@'localhost'
```

这是因为MySQL的root用户权限配置问题。以下是几种解决方案：

## 解决方案1：重置MySQL root密码（推荐）

### 步骤1：停止MySQL服务
```bash
sudo systemctl stop mysql
```

### 步骤2：以安全模式启动MySQL
```bash
sudo mysqld_safe --skip-grant-tables &
```

### 步骤3：连接MySQL并重置密码
```bash
mysql -u root

# 在MySQL命令行中执行：
USE mysql;
UPDATE user SET authentication_string=PASSWORD('root123') WHERE User='root';
UPDATE user SET plugin='mysql_native_password' WHERE User='root';
FLUSH PRIVILEGES;
EXIT;
```

### 步骤4：重启MySQL服务
```bash
sudo systemctl restart mysql
```

### 步骤5：测试连接
```bash
mysql -u root -proot123 -e "SHOW DATABASES;"
```

## 解决方案2：创建新的数据库用户

### 步骤1：使用sudo连接MySQL
```bash
sudo mysql -u root
```

### 步骤2：创建新用户和数据库
```sql
-- 创建数据库
CREATE DATABASE IF NOT EXISTS ryan_mall CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 创建用户
CREATE USER 'ryan_mall'@'localhost' IDENTIFIED BY 'ryan_mall123';

-- 授权
GRANT ALL PRIVILEGES ON ryan_mall.* TO 'ryan_mall'@'localhost';
FLUSH PRIVILEGES;

-- 退出
EXIT;
```

### 步骤3：更新环境配置
编辑 `.env` 文件：
```bash
# 数据库配置
DB_DRIVER=mysql
DB_HOST=localhost
DB_PORT=3306
DB_NAME=ryan_mall
DB_USER=ryan_mall
DB_PASSWORD=ryan_mall123
```

## 解决方案3：使用Docker MySQL（最简单）

### 步骤1：停止本地MySQL
```bash
sudo systemctl stop mysql
```

### 步骤2：启动Docker MySQL
```bash
docker run -d \
  --name ryan-mall-mysql \
  -e MYSQL_ROOT_PASSWORD=root123 \
  -e MYSQL_DATABASE=ryan_mall \
  -p 3306:3306 \
  mysql:8.0
```

### 步骤3：等待MySQL启动
```bash
sleep 30
```

### 步骤4：创建数据库表
```bash
docker exec -i ryan-mall-mysql mysql -uroot -proot123 ryan_mall < create_database.sql
```

## 解决方案4：手动创建数据库和表

如果您能够连接到MySQL，请执行以下SQL语句：

```sql
-- 创建数据库
CREATE DATABASE IF NOT EXISTS ryan_mall CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 使用数据库
USE ryan_mall;

-- 创建用户表
CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id VARCHAR(36) UNIQUE NOT NULL COMMENT '用户UUID',
    username VARCHAR(50) UNIQUE NOT NULL COMMENT '用户名',
    email VARCHAR(100) UNIQUE NOT NULL COMMENT '邮箱',
    password_hash VARCHAR(255) NOT NULL COMMENT '密码哈希',
    phone VARCHAR(20) COMMENT '手机号',
    status TINYINT DEFAULT 1 COMMENT '状态：1-正常，0-禁用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_username (username),
    INDEX idx_email (email),
    INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 创建用户档案表
CREATE TABLE IF NOT EXISTS user_profiles (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id VARCHAR(36) NOT NULL COMMENT '用户UUID',
    nickname VARCHAR(50) COMMENT '昵称',
    avatar_url VARCHAR(255) COMMENT '头像URL',
    gender TINYINT COMMENT '性别：1-男，2-女，0-未知',
    birthday DATE COMMENT '生日',
    bio TEXT COMMENT '个人简介',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户档案表';
```

## 验证设置

### 步骤1：测试数据库连接
```bash
cd ryan_mall
./setup_database.sh test
```

### 步骤2：启动用户服务
```bash
./bin/user-service
```

### 步骤3：测试API
```bash
# 健康检查
curl http://localhost:8081/health

# 用户注册
curl -X POST http://localhost:8081/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "Test123!@#"
  }'
```

## 常见问题

### Q1: 忘记MySQL root密码
A: 使用解决方案1重置密码

### Q2: MySQL服务无法启动
A: 检查端口占用：`sudo netstat -tlnp | grep :3306`

### Q3: 权限被拒绝
A: 使用sudo或创建新用户（解决方案2）

### Q4: 想要快速测试
A: 使用Docker MySQL（解决方案3）

## 推荐配置

对于开发环境，推荐使用以下配置：

```bash
# .env 文件
DB_DRIVER=mysql
DB_HOST=localhost
DB_PORT=3306
DB_NAME=ryan_mall
DB_USER=root
DB_PASSWORD=root123
```

## 下一步

数据库设置完成后，您可以：

1. 启动用户服务：`./bin/user-service`
2. 启动网关服务：`go run cmd/gateway/main.go`
3. 运行集成测试：`./integration_test.sh`
4. 使用测试脚本：`./test-services.sh start`

## 技术支持

如果您仍然遇到问题，请：

1. 检查MySQL服务状态：`sudo systemctl status mysql`
2. 查看MySQL错误日志：`sudo tail -f /var/log/mysql/error.log`
3. 确认端口可用：`netstat -tln | grep :3306`
4. 尝试不同的连接方式

---

**注意**：在生产环境中，请使用强密码并正确配置用户权限。
