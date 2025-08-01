#!/bin/bash

# 构建脚本
set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 项目根目录
PROJECT_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "$PROJECT_ROOT"

# 输出函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查Go环境
check_go() {
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed"
        exit 1
    fi
    log_info "Go version: $(go version)"
}

# 下载依赖
download_deps() {
    log_info "Downloading dependencies..."
    go mod download
    go mod tidy
}

# 生成protobuf代码
generate_proto() {
    log_info "Generating protobuf code..."
    if [ -d "api/proto" ]; then
        find api/proto -name "*.proto" -exec protoc --go_out=. --go-grpc_out=. {} \;
    fi
}

# 构建服务
build_service() {
    local service=$1
    local output_dir="bin"
    
    log_info "Building $service..."
    
    mkdir -p "$output_dir"
    
    CGO_ENABLED=0 GOOS=linux go build \
        -ldflags="-w -s" \
        -o "$output_dir/$service" \
        "./cmd/$service"
    
    if [ $? -eq 0 ]; then
        log_info "$service built successfully"
    else
        log_error "Failed to build $service"
        exit 1
    fi
}

# 构建所有服务
build_all() {
    log_info "Building all services..."
    
    services=(
        "user-service"
        "product-service"
        "order-service"
        "seckill-service"
        "payment-service"
        "api-gateway"
    )
    
    for service in "${services[@]}"; do
        if [ -d "cmd/$service" ]; then
            build_service "$service"
        else
            log_warn "Service directory cmd/$service not found, skipping..."
        fi
    done
}

# 清理构建产物
clean() {
    log_info "Cleaning build artifacts..."
    rm -rf bin/
    rm -rf vendor/
}

# 运行测试
test() {
    log_info "Running tests..."
    go test -v -race -coverprofile=coverage.out ./...
    
    if [ $? -eq 0 ]; then
        log_info "All tests passed"
        go tool cover -html=coverage.out -o coverage.html
        log_info "Coverage report generated: coverage.html"
    else
        log_error "Tests failed"
        exit 1
    fi
}

# 代码格式化
format() {
    log_info "Formatting code..."
    go fmt ./...
    
    if command -v goimports &> /dev/null; then
        goimports -w .
    fi
}

# 代码检查
lint() {
    log_info "Running linters..."
    
    if command -v golangci-lint &> /dev/null; then
        golangci-lint run
    else
        log_warn "golangci-lint not found, running go vet instead"
        go vet ./...
    fi
}

# Docker构建
docker_build() {
    local service=$1
    
    if [ -z "$service" ]; then
        log_error "Service name is required for docker build"
        exit 1
    fi
    
    log_info "Building Docker image for $service..."
    
    docker build -t "ryan-mall/$service:latest" \
        -f "deployments/docker/Dockerfile.$service" .
    
    if [ $? -eq 0 ]; then
        log_info "Docker image for $service built successfully"
    else
        log_error "Failed to build Docker image for $service"
        exit 1
    fi
}

# 显示帮助信息
show_help() {
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  build [service]    Build service (or all services if no service specified)"
    echo "  clean             Clean build artifacts"
    echo "  test              Run tests"
    echo "  format            Format code"
    echo "  lint              Run linters"
    echo "  docker [service]  Build Docker image for service"
    echo "  deps              Download dependencies"
    echo "  proto             Generate protobuf code"
    echo "  help              Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 build                    # Build all services"
    echo "  $0 build user-service       # Build user service only"
    echo "  $0 docker user-service      # Build Docker image for user service"
    echo "  $0 test                     # Run all tests"
}

# 主函数
main() {
    check_go
    
    case "${1:-build}" in
        "build")
            download_deps
            if [ -n "$2" ]; then
                build_service "$2"
            else
                build_all
            fi
            ;;
        "clean")
            clean
            ;;
        "test")
            download_deps
            test
            ;;
        "format")
            format
            ;;
        "lint")
            lint
            ;;
        "docker")
            if [ -n "$2" ]; then
                docker_build "$2"
            else
                log_error "Service name is required for docker build"
                show_help
                exit 1
            fi
            ;;
        "deps")
            download_deps
            ;;
        "proto")
            generate_proto
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        *)
            log_error "Unknown command: $1"
            show_help
            exit 1
            ;;
    esac
}

main "$@"
