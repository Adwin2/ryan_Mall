#!/bin/bash

# Ryan Mall 微服务测试脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查端口是否被占用
check_port() {
    local port=$1
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null ; then
        return 0
    else
        return 1
    fi
}

# 等待服务启动
wait_for_service() {
    local name=$1
    local port=$2
    local max_wait=30
    local count=0
    
    log_info "等待 $name 启动 (端口: $port)..."
    
    while [ $count -lt $max_wait ]; do
        if check_port $port; then
            log_success "$name 启动成功"
            return 0
        fi
        sleep 1
        count=$((count + 1))
    done
    
    log_error "$name 启动超时"
    return 1
}

# 测试服务健康检查
test_health() {
    local name=$1
    local port=$2
    local endpoint=${3:-"/health"}
    
    log_info "测试 $name 健康检查..."
    
    if curl -s -f "http://localhost:$port$endpoint" > /dev/null; then
        log_success "$name 健康检查通过"
        return 0
    else
        log_error "$name 健康检查失败"
        return 1
    fi
}

# 停止所有服务
stop_services() {
    log_info "停止所有服务..."
    
    # 查找并停止相关进程
    pkill -f "user-service" || true
    pkill -f "gateway" || true
    
    sleep 2
    log_success "所有服务已停止"
}

# 构建服务
build_services() {
    log_info "构建服务..."
    
    # 创建bin目录
    mkdir -p bin
    
    # 构建用户服务
    log_info "构建用户服务..."
    go build -o bin/user-service ./cmd/user
    
    # 构建网关服务
    log_info "构建网关服务..."
    go build -o bin/gateway ./cmd/gateway
    
    log_success "所有服务构建完成"
}

# 启动服务
start_services() {
    log_info "启动服务..."
    
    # 设置环境变量
    export ENVIRONMENT=development
    export LOG_LEVEL=info
    export DB_HOST=localhost
    export DB_PORT=3306
    export DB_USERNAME=root
    export DB_PASSWORD=root123
    export DB_DATABASE=ryan_mall
    export REDIS_ADDRESS=localhost:6379
    export JWT_SECRET=test-secret-key
    
    # 启动用户服务
    log_info "启动用户服务..."
    nohup ./bin/user-service > logs/user-service.log 2>&1 &
    USER_PID=$!
    echo $USER_PID > logs/user-service.pid
    
    # 等待用户服务启动
    if wait_for_service "用户服务" 8081; then
        # 测试用户服务健康检查
        test_health "用户服务" 8081
    fi
    
    # 启动网关服务
    log_info "启动网关服务..."
    nohup ./bin/gateway > logs/gateway.log 2>&1 &
    GATEWAY_PID=$!
    echo $GATEWAY_PID > logs/gateway.pid
    
    # 等待网关服务启动
    if wait_for_service "网关服务" 8080; then
        # 测试网关服务健康检查
        test_health "网关服务" 8080
    fi
}

# 测试API
test_apis() {
    log_info "测试API..."
    
    # 测试用户注册
    log_info "测试用户注册..."
    REGISTER_RESPONSE=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d '{"username":"testuser","email":"test@example.com","password":"Test123!@#"}' \
        http://localhost:8081/api/v1/users/register)
    
    if echo "$REGISTER_RESPONSE" | grep -q "user_id"; then
        log_success "用户注册测试通过"
        
        # 提取用户ID
        USER_ID=$(echo "$REGISTER_RESPONSE" | grep -o '"user_id":"[^"]*"' | cut -d'"' -f4)
        log_info "注册用户ID: $USER_ID"
        
        # 测试用户登录
        log_info "测试用户登录..."
        LOGIN_RESPONSE=$(curl -s -X POST \
            -H "Content-Type: application/json" \
            -d '{"username":"testuser","password":"Test123!@#"}' \
            http://localhost:8081/api/v1/users/login)
        
        if echo "$LOGIN_RESPONSE" | grep -q "access_token"; then
            log_success "用户登录测试通过"
            
            # 提取访问令牌
            ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
            
            # 测试获取用户信息
            log_info "测试获取用户信息..."
            USER_INFO_RESPONSE=$(curl -s -X GET \
                -H "Authorization: Bearer $ACCESS_TOKEN" \
                http://localhost:8081/api/v1/users/$USER_ID)
            
            if echo "$USER_INFO_RESPONSE" | grep -q "testuser"; then
                log_success "获取用户信息测试通过"
            else
                log_error "获取用户信息测试失败"
            fi
        else
            log_error "用户登录测试失败"
        fi
    else
        log_error "用户注册测试失败"
        log_error "响应: $REGISTER_RESPONSE"
    fi
    
    # 测试通过网关访问
    log_info "测试通过网关访问..."
    GATEWAY_HEALTH=$(curl -s http://localhost:8080/health)
    if echo "$GATEWAY_HEALTH" | grep -q "healthy"; then
        log_success "网关健康检查通过"
    else
        log_error "网关健康检查失败"
    fi
}

# 显示服务状态
show_status() {
    log_info "服务状态:"
    echo "----------------------------------------"
    
    if check_port 8081; then
        echo -e "用户服务 (8081): ${GREEN}运行中${NC}"
    else
        echo -e "用户服务 (8081): ${RED}未运行${NC}"
    fi
    
    if check_port 8080; then
        echo -e "网关服务 (8080): ${GREEN}运行中${NC}"
    else
        echo -e "网关服务 (8080): ${RED}未运行${NC}"
    fi
    
    echo "----------------------------------------"
}

# 显示日志
show_logs() {
    local service=$1
    if [ -z "$service" ]; then
        echo "可用的日志文件:"
        ls -la logs/*.log 2>/dev/null || echo "没有找到日志文件"
        return
    fi
    
    case $service in
        "user"|"user-service")
            if [ -f "logs/user-service.log" ]; then
                tail -f logs/user-service.log
            else
                log_error "用户服务日志文件不存在"
            fi
            ;;
        "gateway")
            if [ -f "logs/gateway.log" ]; then
                tail -f logs/gateway.log
            else
                log_error "网关服务日志文件不存在"
            fi
            ;;
        *)
            log_error "未知服务: $service"
            ;;
    esac
}

# 主函数
main() {
    # 创建日志目录
    mkdir -p logs
    
    case "${1:-help}" in
        "build")
            build_services
            ;;
        "start")
            build_services
            start_services
            show_status
            ;;
        "stop")
            stop_services
            ;;
        "restart")
            stop_services
            sleep 2
            build_services
            start_services
            show_status
            ;;
        "test")
            test_apis
            ;;
        "status")
            show_status
            ;;
        "logs")
            show_logs $2
            ;;
        "help"|*)
            echo "Ryan Mall 微服务测试脚本"
            echo ""
            echo "用法: $0 [命令] [参数]"
            echo ""
            echo "命令:"
            echo "  build     构建所有服务"
            echo "  start     构建并启动所有服务"
            echo "  stop      停止所有服务"
            echo "  restart   重启所有服务"
            echo "  test      测试API"
            echo "  status    显示服务状态"
            echo "  logs      显示日志 [user|gateway]"
            echo "  help      显示帮助信息"
            echo ""
            echo "示例:"
            echo "  $0 start          # 启动所有服务"
            echo "  $0 test           # 测试API"
            echo "  $0 logs user      # 查看用户服务日志"
            echo "  $0 stop           # 停止所有服务"
            ;;
    esac
}

# 执行主函数
main "$@"
