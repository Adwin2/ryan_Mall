# 启动指南

```bash
# 1. 启动数据库服务
docker compose up -d mysql redis

# 2. 启动前端服务
docker compose up -d frontend

# 3. 启动后端API
SERVER_PORT=8081 go run ./cmd/server/main.go
```
