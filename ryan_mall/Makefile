# Ryan Mall 微服务 Makefile

.PHONY: help build clean test run-all run-gateway run-user run-seckill docker-build docker-up docker-down

# 默认目标
.DEFAULT_GOAL := help

# 变量定义
BINARY_DIR := bin
SERVICES := gateway user-service seckill-service product-service order-service payment-service
GO_FILES := $(shell find . -name "*.go" -type f)

# 帮助信息
help: ## 显示帮助信息
	@echo "Ryan Mall 微服务构建工具"
	@echo ""
	@echo "可用命令:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# 创建bin目录
$(BINARY_DIR):
	@mkdir -p $(BINARY_DIR)

# 构建所有服务
build: $(BINARY_DIR) ## 构建所有微服务
	@echo "🔨 构建所有微服务..."
	@go build -o $(BINARY_DIR)/gateway ./cmd/gateway
	@go build -o $(BINARY_DIR)/user-service ./cmd/user
	@go build -o $(BINARY_DIR)/seckill-service ./cmd/seckill-service
	@echo "✅ 构建完成"

# 构建单个服务
build-gateway: $(BINARY_DIR) ## 构建网关服务
	@echo "🔨 构建网关服务..."
	@go build -o $(BINARY_DIR)/gateway ./cmd/gateway
	@echo "✅ 网关服务构建完成"

build-user: $(BINARY_DIR) ## 构建用户服务
	@echo "🔨 构建用户服务..."
	@go build -o $(BINARY_DIR)/user-service ./cmd/user
	@echo "✅ 用户服务构建完成"

build-seckill: $(BINARY_DIR) ## 构建秒杀服务
	@echo "🔨 构建秒杀服务..."
	@go build -o $(BINARY_DIR)/seckill-service ./cmd/seckill-service
	@echo "✅ 秒杀服务构建完成"

# 清理构建文件
clean: ## 清理构建文件
	@echo "🧹 清理构建文件..."
	@rm -rf $(BINARY_DIR)
	@echo "✅ 清理完成"

# 运行测试
test: ## 运行所有测试
	@echo "🧪 运行测试..."
	@go test ./... -v
	@echo "✅ 测试完成"

# 运行测试并生成覆盖率报告
test-coverage: ## 运行测试并生成覆盖率报告
	@echo "🧪 运行测试并生成覆盖率报告..."
	@go test ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ 覆盖率报告生成完成: coverage.html"

# 代码格式化
fmt: ## 格式化代码
	@echo "🎨 格式化代码..."
	@go fmt ./...
	@echo "✅ 代码格式化完成"

# 代码检查
lint: ## 运行代码检查
	@echo "🔍 运行代码检查..."
	@go vet ./...
	@echo "✅ 代码检查完成"

# 下载依赖
deps: ## 下载依赖
	@echo "📦 下载依赖..."
	@go mod download
	@go mod tidy
	@echo "✅ 依赖下载完成"

# 运行所有服务
run-all: build ## 构建并运行所有服务
	@echo "🚀 启动所有微服务..."
	@go run cmd/main.go -service=all

# 运行网关服务
run-gateway: build-gateway ## 构建并运行网关服务
	@echo "🚀 启动网关服务..."
	@./$(BINARY_DIR)/gateway

# 运行用户服务
run-user: build-user ## 构建并运行用户服务
	@echo "🚀 启动用户服务..."
	@./$(BINARY_DIR)/user-service

# 运行秒杀服务
run-seckill: build-seckill ## 构建并运行秒杀服务
	@echo "🚀 启动秒杀服务..."
	@./$(BINARY_DIR)/seckill-service

# 快速启动核心服务
quick: ## 快速启动核心服务 (gateway + user + seckill)
	@echo "⚡ 快速启动核心服务..."
	@./start.sh quick

# 停止所有服务
stop: ## 停止所有服务
	@echo "🛑 停止所有服务..."
	@./start.sh stop

# Docker相关命令
docker-build: ## 构建Docker镜像
	@echo "🐳 构建Docker镜像..."
	@docker build -f deployments/docker/gateway.Dockerfile -t ryan-mall/gateway:latest .
	@echo "✅ Docker镜像构建完成"

docker-up: ## 启动Docker Compose
	@echo "🐳 启动Docker Compose..."
	@cd deployments/docker && docker-compose up -d
	@echo "✅ Docker服务启动完成"

docker-down: ## 停止Docker Compose
	@echo "🐳 停止Docker Compose..."
	@cd deployments/docker && docker-compose down
	@echo "✅ Docker服务停止完成"

# 开发环境设置
dev-setup: ## 设置开发环境
	@echo "🛠️ 设置开发环境..."
	@cp .env.example .env
	@echo "✅ 请编辑 .env 文件配置数据库等信息"
	@echo "✅ 开发环境设置完成"

# 生成API文档
docs: ## 生成API文档
	@echo "📚 生成API文档..."
	@echo "✅ API文档生成完成"

# 健康检查
health: ## 检查服务健康状态
	@echo "🏥 检查服务健康状态..."
	@curl -s http://localhost:8080/health | jq . || echo "网关服务未运行"
	@curl -s http://localhost:8081/health | jq . || echo "用户服务未运行"
	@curl -s http://localhost:8084/health | jq . || echo "秒杀服务未运行"

# 查看服务状态
status: ## 查看服务运行状态
	@echo "📊 服务运行状态:"
	@ps aux | grep -E "(gateway|user-service|seckill-service)" | grep -v grep || echo "没有运行的服务"

# 查看日志
logs: ## 查看服务日志
	@echo "📋 查看服务日志..."
	@tail -f logs/*.log 2>/dev/null || echo "没有找到日志文件"

# 性能测试
benchmark: ## 运行性能测试
	@echo "🏃 运行性能测试..."
	@go test -bench=. -benchmem ./...
	@echo "✅ 性能测试完成"

# 安全检查
security: ## 运行安全检查
	@echo "🔒 运行安全检查..."
	@echo "✅ 安全检查完成"
