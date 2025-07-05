# Ryan Mall 部署指南

## 📋 目录
- [环境要求](#环境要求)
- [本地开发环境](#本地开发环境)
- [生产环境部署](#生产环境部署)
- [Docker 部署](#docker-部署)
- [配置说明](#配置说明)
- [监控运维](#监控运维)

## 环境要求

### 基础环境
- **Go**: 1.19 或更高版本
- **MySQL**: 8.0 或更高版本
- **Git**: 版本控制工具
- **操作系统**: Linux/macOS/Windows

### 推荐配置
- **CPU**: 2核心或以上
- **内存**: 4GB 或以上
- **存储**: 20GB 可用空间
- **网络**: 稳定的网络连接

## 本地开发环境

### 1. 克隆项目
```bash
# 克隆代码仓库
git clone https://github.com/your-username/ryan-mall.git
cd ryan-mall

# 查看项目结构
tree -L 2
```

### 2. 安装 Go 依赖
```bash
# 初始化 Go 模块
go mod tidy

# 验证依赖安装
go mod verify
```

### 3. 配置数据库
```bash
# 登录 MySQL
mysql -u root -p

# 创建数据库
CREATE DATABASE ryan_mall CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

# 创建用户（可选）
CREATE USER 'ryan_mall'@'localhost' IDENTIFIED BY 'your_password';
GRANT ALL PRIVILEGES ON ryan_mall.* TO 'ryan_mall'@'localhost';
FLUSH PRIVILEGES;
```

### 4. 配置应用
```bash
# 复制配置文件模板
cp configs/config.yaml.example configs/config.yaml

# 编辑配置文件
vim configs/config.yaml
```

配置示例：
```yaml
server:
  port: 8080
  mode: debug

database:
  host: localhost
  port: 3306
  username: root
  password: your_password
  database: ryan_mall
  charset: utf8mb4
  parse_time: true
  loc: Local

jwt:
  secret_key: your_jwt_secret_key_here
  expire_hours: 24

log:
  level: debug
  format: json
  output: stdout
```

### 5. 启动应用
```bash
# 启动开发服务器
go run cmd/server/main.go

# 或者编译后运行
go build -o bin/ryan-mall cmd/server/main.go
./bin/ryan-mall
```

### 6. 验证安装
```bash
# 检查服务状态
curl http://localhost:8080/ping

# 运行测试脚本
chmod +x test_*.sh
./test_user_api.sh
```

## 生产环境部署

### 1. 服务器准备
```bash
# 更新系统包
sudo apt update && sudo apt upgrade -y

# 安装必要工具
sudo apt install -y git curl wget vim

# 安装 Go
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

### 2. 安装 MySQL
```bash
# 安装 MySQL 8.0
sudo apt install -y mysql-server

# 安全配置
sudo mysql_secure_installation

# 配置 MySQL
sudo vim /etc/mysql/mysql.conf.d/mysqld.cnf
```

MySQL 配置优化：
```ini
[mysqld]
# 基础配置
bind-address = 0.0.0.0
port = 3306
max_connections = 200
max_connect_errors = 10

# 性能优化
innodb_buffer_pool_size = 1G
innodb_log_file_size = 256M
innodb_flush_log_at_trx_commit = 2
innodb_flush_method = O_DIRECT

# 字符集
character-set-server = utf8mb4
collation-server = utf8mb4_unicode_ci
```

### 3. 部署应用
```bash
# 创建应用目录
sudo mkdir -p /opt/ryan-mall
sudo chown $USER:$USER /opt/ryan-mall

# 克隆代码
cd /opt/ryan-mall
git clone https://github.com/your-username/ryan-mall.git .

# 编译应用
go build -o bin/ryan-mall cmd/server/main.go

# 创建配置文件
cp configs/config.yaml.example configs/config.yaml
vim configs/config.yaml
```

生产环境配置：
```yaml
server:
  port: 8080
  mode: release

database:
  host: localhost
  port: 3306
  username: ryan_mall
  password: your_secure_password
  database: ryan_mall
  max_idle_conns: 10
  max_open_conns: 100
  conn_max_lifetime: 3600

jwt:
  secret_key: your_very_secure_jwt_secret_key
  expire_hours: 24

log:
  level: info
  format: json
  output: /var/log/ryan-mall/app.log
```

### 4. 创建系统服务
```bash
# 创建服务文件
sudo vim /etc/systemd/system/ryan-mall.service
```

服务配置：
```ini
[Unit]
Description=Ryan Mall E-commerce API Server
After=network.target mysql.service

[Service]
Type=simple
User=ryan-mall
Group=ryan-mall
WorkingDirectory=/opt/ryan-mall
ExecStart=/opt/ryan-mall/bin/ryan-mall
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

# 环境变量
Environment=GIN_MODE=release

# 安全设置
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ReadWritePaths=/opt/ryan-mall /var/log/ryan-mall

[Install]
WantedBy=multi-user.target
```

### 5. 启动服务
```bash
# 创建用户
sudo useradd -r -s /bin/false ryan-mall

# 设置权限
sudo chown -R ryan-mall:ryan-mall /opt/ryan-mall
sudo mkdir -p /var/log/ryan-mall
sudo chown ryan-mall:ryan-mall /var/log/ryan-mall

# 启动服务
sudo systemctl daemon-reload
sudo systemctl enable ryan-mall
sudo systemctl start ryan-mall

# 检查状态
sudo systemctl status ryan-mall
```

### 6. 配置反向代理 (Nginx)
```bash
# 安装 Nginx
sudo apt install -y nginx

# 创建配置文件
sudo vim /etc/nginx/sites-available/ryan-mall
```

Nginx 配置：
```nginx
server {
    listen 80;
    server_name your-domain.com;

    # 重定向到 HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;

    # SSL 配置
    ssl_certificate /path/to/your/certificate.crt;
    ssl_certificate_key /path/to/your/private.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512;

    # 日志配置
    access_log /var/log/nginx/ryan-mall.access.log;
    error_log /var/log/nginx/ryan-mall.error.log;

    # 反向代理配置
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # 超时设置
        proxy_connect_timeout 30s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;
    }

    # 静态文件缓存
    location ~* \.(jpg|jpeg|png|gif|ico|css|js)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    # 安全头
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header Referrer-Policy "no-referrer-when-downgrade" always;
    add_header Content-Security-Policy "default-src 'self'" always;
}
```

启用配置：
```bash
# 启用站点
sudo ln -s /etc/nginx/sites-available/ryan-mall /etc/nginx/sites-enabled/

# 测试配置
sudo nginx -t

# 重启 Nginx
sudo systemctl restart nginx
```

## Docker 部署

### 1. 创建 Dockerfile
```dockerfile
# 多阶段构建
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 编译应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/server/main.go

# 运行阶段
FROM alpine:latest

# 安装必要工具
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

# 复制编译好的应用
COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs

# 暴露端口
EXPOSE 8080

# 启动应用
CMD ["./main"]
```

### 2. 创建 docker-compose.yml
```yaml
version: '3.8'

services:
  # MySQL 数据库
  mysql:
    image: mysql:8.0
    container_name: ryan-mall-mysql
    restart: always
    environment:
      MYSQL_ROOT_PASSWORD: rootpassword
      MYSQL_DATABASE: ryan_mall
      MYSQL_USER: ryan_mall
      MYSQL_PASSWORD: password
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
      - ./configs/mysql.cnf:/etc/mysql/conf.d/mysql.cnf
    command: --default-authentication-plugin=mysql_native_password

  # Redis 缓存
  redis:
    image: redis:7-alpine
    container_name: ryan-mall-redis
    restart: always
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

  # 应用服务
  app:
    build: .
    container_name: ryan-mall-app
    restart: always
    ports:
      - "8080:8080"
    depends_on:
      - mysql
      - redis
    environment:
      - GIN_MODE=release
    volumes:
      - ./configs:/root/configs
      - ./logs:/root/logs

  # Nginx 反向代理
  nginx:
    image: nginx:alpine
    container_name: ryan-mall-nginx
    restart: always
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./configs/nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/nginx/ssl
    depends_on:
      - app

volumes:
  mysql_data:
  redis_data:
```

### 3. 启动容器
```bash
# 构建并启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f app

# 停止服务
docker-compose down
```

## 配置说明

### 环境变量配置
```bash
# 创建环境变量文件
vim .env
```

```env
# 服务配置
SERVER_PORT=8080
GIN_MODE=release

# 数据库配置
DB_HOST=localhost
DB_PORT=3306
DB_USERNAME=ryan_mall
DB_PASSWORD=your_password
DB_DATABASE=ryan_mall

# JWT 配置
JWT_SECRET_KEY=your_jwt_secret_key
JWT_EXPIRE_HOURS=24

# 日志配置
LOG_LEVEL=info
LOG_FORMAT=json
```

### 配置文件模板
```yaml
# configs/config.yaml.example
server:
  port: ${SERVER_PORT:8080}
  mode: ${GIN_MODE:debug}

database:
  host: ${DB_HOST:localhost}
  port: ${DB_PORT:3306}
  username: ${DB_USERNAME:root}
  password: ${DB_PASSWORD:password}
  database: ${DB_DATABASE:ryan_mall}
  charset: utf8mb4
  parse_time: true
  loc: Local
  max_idle_conns: 10
  max_open_conns: 100
  conn_max_lifetime: 3600

jwt:
  secret_key: ${JWT_SECRET_KEY:default_secret}
  expire_hours: ${JWT_EXPIRE_HOURS:24}

log:
  level: ${LOG_LEVEL:debug}
  format: ${LOG_FORMAT:text}
  output: ${LOG_OUTPUT:stdout}
```

## 监控运维

### 1. 健康检查
```bash
# 创建健康检查脚本
vim scripts/health_check.sh
```

```bash
#!/bin/bash
# 健康检查脚本

API_URL="http://localhost:8080"
LOG_FILE="/var/log/ryan-mall/health.log"

# 检查 API 服务
check_api() {
    response=$(curl -s -o /dev/null -w "%{http_code}" $API_URL/ping)
    if [ $response -eq 200 ]; then
        echo "$(date): API service is healthy" >> $LOG_FILE
        return 0
    else
        echo "$(date): API service is unhealthy (HTTP $response)" >> $LOG_FILE
        return 1
    fi
}

# 检查数据库连接
check_database() {
    mysql -h localhost -u ryan_mall -p$DB_PASSWORD -e "SELECT 1" ryan_mall > /dev/null 2>&1
    if [ $? -eq 0 ]; then
        echo "$(date): Database is healthy" >> $LOG_FILE
        return 0
    else
        echo "$(date): Database is unhealthy" >> $LOG_FILE
        return 1
    fi
}

# 执行检查
check_api && check_database
```

### 2. 日志轮转
```bash
# 配置 logrotate
sudo vim /etc/logrotate.d/ryan-mall
```

```
/var/log/ryan-mall/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 644 ryan-mall ryan-mall
    postrotate
        systemctl reload ryan-mall
    endscript
}
```

### 3. 监控脚本
```bash
# 创建监控脚本
vim scripts/monitor.sh
```

```bash
#!/bin/bash
# 系统监控脚本

# 检查服务状态
check_service() {
    if systemctl is-active --quiet ryan-mall; then
        echo "✅ Ryan Mall service is running"
    else
        echo "❌ Ryan Mall service is not running"
        # 尝试重启服务
        sudo systemctl restart ryan-mall
    fi
}

# 检查磁盘空间
check_disk() {
    usage=$(df / | awk 'NR==2 {print $5}' | sed 's/%//')
    if [ $usage -gt 80 ]; then
        echo "⚠️  Disk usage is high: ${usage}%"
    fi
}

# 检查内存使用
check_memory() {
    usage=$(free | awk 'NR==2{printf "%.0f", $3*100/$2}')
    if [ $usage -gt 80 ]; then
        echo "⚠️  Memory usage is high: ${usage}%"
    fi
}

# 执行所有检查
check_service
check_disk
check_memory
```

### 4. 备份脚本
```bash
# 创建备份脚本
vim scripts/backup.sh
```

```bash
#!/bin/bash
# 数据库备份脚本

BACKUP_DIR="/opt/backups/ryan-mall"
DATE=$(date +%Y%m%d_%H%M%S)
DB_NAME="ryan_mall"

# 创建备份目录
mkdir -p $BACKUP_DIR

# 备份数据库
mysqldump -u ryan_mall -p$DB_PASSWORD $DB_NAME > $BACKUP_DIR/db_backup_$DATE.sql

# 压缩备份文件
gzip $BACKUP_DIR/db_backup_$DATE.sql

# 删除7天前的备份
find $BACKUP_DIR -name "*.sql.gz" -mtime +7 -delete

echo "Database backup completed: db_backup_$DATE.sql.gz"
```

## 故障排除

### 常见问题

1. **服务启动失败**
```bash
# 查看服务日志
sudo journalctl -u ryan-mall -f

# 检查配置文件
go run cmd/server/main.go --config-check
```

2. **数据库连接失败**
```bash
# 测试数据库连接
mysql -h localhost -u ryan_mall -p ryan_mall

# 检查数据库状态
sudo systemctl status mysql
```

3. **端口占用**
```bash
# 查看端口占用
sudo netstat -tlnp | grep :8080

# 杀死占用进程
sudo kill -9 <PID>
```

4. **权限问题**
```bash
# 检查文件权限
ls -la /opt/ryan-mall/

# 修复权限
sudo chown -R ryan-mall:ryan-mall /opt/ryan-mall/
```

通过以上部署指南，你可以在不同环境中成功部署 Ryan Mall 电商系统，并建立完善的监控运维机制。
