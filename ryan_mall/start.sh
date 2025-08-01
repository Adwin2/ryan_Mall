#!/bin/bash

# Ryan Mall 微服务启动脚本

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

# 显示帮助信息
show_help() {
    cat << EOF
Ryan Mall 微服务启动脚本

用法: $0 [选项] [服务名]

选项:
    -h, --help          显示帮助信息
    -b, --build         构建后启动
    -d, --dev           开发模式启动
    -p, --prod          生产模式启动

服务名:
    gateway             API网关 (端口: 8080)
    user                用户服务 (端口: 8081)
    product             商品服务 (端口: 8082)
    order               订单服务 (端口: 8083)
    seckill             秒杀服务 (端口: 8084)
    payment             支付服务 (端口: 8085)
    all                 所有服务

示例:
    $0 gateway                    # 启动网关服务
    $0 --build all               # 构建并启动所有服务
    $0 --dev user                # 开发模式启动用户服务
    $0 --prod gateway            # 生产模式启动网关

快速启动:
    $0 quick                     # 快速启动核心服务 (gateway + user + seckill)

EOF
}

# 检查Go环境
check_go_env() {
    if ! command -v go &> /dev/null; then
        log_error "Go 未安装或未在PATH中"
        exit 1
    fi
    
    log_info "Go版本: $(go version)"
}

# 构建服务
build_services() {
    local service=$1
    log_info "构建服务: $service"
    
    if ! go run cmd/main.go -service="$service" -build; then
        log_error "构建失败"
        exit 1
    fi
    
    log_success "构建完成"
}

# 启动服务
start_services() {
    local service=$1
    local mode=${2:-"development"}
    
    log_info "启动服务: $service (模式: $mode)"
    
    # 设置环境变量
    export ENVIRONMENT=$mode
    export LOG_LEVEL="info"
    
    if [ "$mode" = "production" ]; then
        export GIN_MODE="release"
        export LOG_LEVEL="warn"
    fi
    
    # 启动服务
    go run cmd/main.go -service="$service"
}

# 快速启动核心服务
quick_start() {
    log_info "快速启动核心服务..."
    
    # 构建核心服务
    build_services "gateway"
    build_services "user"
    build_services "seckill"
    
    log_info "启动核心服务 (gateway + user + seckill)..."
    
    # 后台启动用户服务
    export ENVIRONMENT="development"
    nohup ./bin/user-service > logs/user-service.log 2>&1 &
    USER_PID=$!
    log_success "用户服务启动 (PID: $USER_PID)"
    
    # 等待用户服务启动
    sleep 3
    
    # 后台启动秒杀服务
    nohup ./bin/seckill-service > logs/seckill-service.log 2>&1 &
    SECKILL_PID=$!
    log_success "秒杀服务启动 (PID: $SECKILL_PID)"
    
    # 等待秒杀服务启动
    sleep 3
    
    # 前台启动网关服务
    log_info "启动网关服务..."
    ./bin/gateway
}

# 停止所有服务
stop_all() {
    log_info "停止所有服务..."
    
    # 查找并停止所有相关进程
    pkill -f "ryan-mall" || true
    pkill -f "gateway" || true
    pkill -f "user-service" || true
    pkill -f "seckill-service" || true
    
    log_success "所有服务已停止"
}

# 创建日志目录
mkdir -p logs

# 解析命令行参数
BUILD=false
MODE="development"
SERVICE=""

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -b|--build)
            BUILD=true
            shift
            ;;
        -d|--dev)
            MODE="development"
            shift
            ;;
        -p|--prod)
            MODE="production"
            shift
            ;;
        quick)
            check_go_env
            quick_start
            exit 0
            ;;
        stop)
            stop_all
            exit 0
            ;;
        gateway|user|product|order|seckill|payment|all)
            SERVICE="$1"
            shift
            ;;
        *)
            log_error "未知参数: $1"
            show_help
            exit 1
            ;;
    esac
done

# 检查是否指定了服务
if [ -z "$SERVICE" ]; then
    log_error "请指定要启动的服务"
    show_help
    exit 1
fi

# 检查Go环境
check_go_env

# 构建服务（如果需要）
if [ "$BUILD" = true ]; then
    build_services "$SERVICE"
fi

# 启动服务
start_services "$SERVICE" "$MODE"
