#!/bin/bash

# Proto代码生成脚本
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

# 检查protoc是否安装
check_protoc() {
    if ! command -v protoc &> /dev/null; then
        log_error "protoc is not installed. Please install Protocol Buffers compiler."
        log_info "Installation guide: https://grpc.io/docs/protoc-installation/"
        exit 1
    fi
    log_info "protoc version: $(protoc --version)"
}

# 检查Go插件是否安装
check_go_plugins() {
    if ! command -v protoc-gen-go &> /dev/null; then
        log_error "protoc-gen-go is not installed."
        log_info "Installing protoc-gen-go..."
        go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    fi

    if ! command -v protoc-gen-go-grpc &> /dev/null; then
        log_error "protoc-gen-go-grpc is not installed."
        log_info "Installing protoc-gen-go-grpc..."
        go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    fi

    log_info "Go protobuf plugins are ready"
}

# 生成Go代码
generate_go_code() {
    log_info "Generating Go code from proto files..."
    
    # 创建输出目录
    mkdir -p api/proto/order
    
    # 生成订单服务代码
    if [ -f "api/proto/order/order_service.proto" ]; then
        log_info "Generating order service code..."
        protoc \
            --go_out=. \
            --go_opt=paths=source_relative \
            --go-grpc_out=. \
            --go-grpc_opt=paths=source_relative \
            api/proto/order/order_service.proto
        
        if [ $? -eq 0 ]; then
            log_info "Order service code generated successfully"
        else
            log_error "Failed to generate order service code"
            exit 1
        fi
    else
        log_warn "Order service proto file not found"
    fi
    
    # 生成用户服务代码（如果存在）
    if [ -f "api/proto/user/user_service.proto" ]; then
        log_info "Generating user service code..."
        mkdir -p api/proto/user
        protoc \
            --go_out=. \
            --go_opt=paths=source_relative \
            --go-grpc_out=. \
            --go-grpc_opt=paths=source_relative \
            api/proto/user/user_service.proto
        
        if [ $? -eq 0 ]; then
            log_info "User service code generated successfully"
        else
            log_error "Failed to generate user service code"
            exit 1
        fi
    fi
    
    # 生成商品服务代码（如果存在）
    if [ -f "api/proto/product/product_service.proto" ]; then
        log_info "Generating product service code..."
        mkdir -p api/proto/product
        protoc \
            --go_out=. \
            --go_opt=paths=source_relative \
            --go-grpc_out=. \
            --go-grpc_opt=paths=source_relative \
            api/proto/product/product_service.proto
        
        if [ $? -eq 0 ]; then
            log_info "Product service code generated successfully"
        else
            log_error "Failed to generate product service code"
            exit 1
        fi
    fi
}

# 清理生成的代码
clean_generated_code() {
    log_info "Cleaning generated code..."
    find api/proto -name "*.pb.go" -delete
    find api/proto -name "*_grpc.pb.go" -delete
    log_info "Generated code cleaned"
}

# 验证生成的代码
validate_generated_code() {
    log_info "Validating generated code..."
    
    # 检查是否有Go语法错误
    if go build ./api/proto/... > /dev/null 2>&1; then
        log_info "Generated code validation passed"
    else
        log_error "Generated code has syntax errors"
        go build ./api/proto/...
        exit 1
    fi
}

# 显示帮助信息
show_help() {
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  generate    Generate Go code from proto files"
    echo "  clean       Clean generated code"
    echo "  validate    Validate generated code"
    echo "  help        Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 generate    # Generate all proto code"
    echo "  $0 clean       # Clean generated code"
    echo "  $0 validate    # Validate generated code"
}

# 主函数
main() {
    case "${1:-generate}" in
        "generate")
            check_protoc
            check_go_plugins
            generate_go_code
            validate_generated_code
            log_info "Proto code generation completed successfully!"
            ;;
        "clean")
            clean_generated_code
            ;;
        "validate")
            validate_generated_code
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
