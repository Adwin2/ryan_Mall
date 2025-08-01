#!/bin/bash

# Ryan Mall 微服务部署脚本
# 支持Docker Compose和Kubernetes部署

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
Ryan Mall 微服务部署脚本

用法: $0 [选项] [命令]

选项:
    -h, --help          显示帮助信息
    -e, --env ENV       指定环境 (dev|test|prod) [默认: dev]
    -m, --mode MODE     部署模式 (docker|k8s) [默认: docker]
    -v, --verbose       详细输出

命令:
    build               构建所有服务镜像
    deploy              部署所有服务
    start               启动所有服务
    stop                停止所有服务
    restart             重启所有服务
    status              查看服务状态
    logs                查看服务日志
    clean               清理资源
    test                运行测试

示例:
    $0 -e dev -m docker deploy    # 使用Docker在开发环境部署
    $0 -e prod -m k8s deploy      # 使用Kubernetes在生产环境部署
    $0 build                      # 构建所有镜像
    $0 logs gateway               # 查看网关日志

EOF
}

# 默认配置
ENVIRONMENT="dev"
MODE="docker"
VERBOSE=false
COMMAND=""
SERVICE=""

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -e|--env)
            ENVIRONMENT="$2"
            shift 2
            ;;
        -m|--mode)
            MODE="$2"
            shift 2
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        build|deploy|start|stop|restart|status|logs|clean|test)
            COMMAND="$1"
            shift
            ;;
        *)
            if [[ -z "$SERVICE" ]]; then
                SERVICE="$1"
            fi
            shift
            ;;
    esac
done

# 检查必要的工具
check_dependencies() {
    log_info "检查依赖工具..."
    
    if [[ "$MODE" == "docker" ]]; then
        if ! command -v docker &> /dev/null; then
            log_error "Docker 未安装"
            exit 1
        fi
        
        if ! command -v docker-compose &> /dev/null; then
            log_error "Docker Compose 未安装"
            exit 1
        fi
    elif [[ "$MODE" == "k8s" ]]; then
        if ! command -v kubectl &> /dev/null; then
            log_error "kubectl 未安装"
            exit 1
        fi
    fi
    
    log_success "依赖检查完成"
}

# 构建镜像
build_images() {
    log_info "构建服务镜像..."
    
    # 构建网关
    log_info "构建网关镜像..."
    docker build -f deployments/docker/gateway.Dockerfile -t ryan-mall/gateway:latest .
    
    # 构建用户服务
    log_info "构建用户服务镜像..."
    docker build -f deployments/docker/user-service.Dockerfile -t ryan-mall/user-service:latest .
    
    log_success "镜像构建完成"
}

# Docker部署
deploy_docker() {
    log_info "使用Docker Compose部署..."
    
    cd deployments/docker
    
    # 设置环境变量
    export ENVIRONMENT=$ENVIRONMENT
    
    # 启动服务
    docker-compose up -d
    
    log_success "Docker部署完成"
    
    # 显示服务状态
    docker-compose ps
}

# Kubernetes部署
deploy_k8s() {
    log_info "使用Kubernetes部署..."
    
    # 创建命名空间
    kubectl apply -f deployments/k8s/namespace.yaml
    
    # 应用配置
    kubectl apply -f deployments/k8s/configmap.yaml
    kubectl apply -f deployments/k8s/secret.yaml
    
    # 部署基础设施
    kubectl apply -f deployments/k8s/mysql.yaml
    kubectl apply -f deployments/k8s/redis.yaml
    kubectl apply -f deployments/k8s/monitoring.yaml
    
    # 等待基础设施就绪
    log_info "等待基础设施就绪..."
    kubectl wait --for=condition=ready pod -l app=mysql -n ryan-mall --timeout=300s
    kubectl wait --for=condition=ready pod -l app=redis -n ryan-mall --timeout=300s
    
    # 部署应用服务
    kubectl apply -f deployments/k8s/gateway.yaml
    
    log_success "Kubernetes部署完成"
    
    # 显示服务状态
    kubectl get pods -n ryan-mall
}

# 启动服务
start_services() {
    if [[ "$MODE" == "docker" ]]; then
        cd deployments/docker
        docker-compose start
    elif [[ "$MODE" == "k8s" ]]; then
        kubectl scale deployment --replicas=1 --all -n ryan-mall
    fi
    
    log_success "服务启动完成"
}

# 停止服务
stop_services() {
    if [[ "$MODE" == "docker" ]]; then
        cd deployments/docker
        docker-compose stop
    elif [[ "$MODE" == "k8s" ]]; then
        kubectl scale deployment --replicas=0 --all -n ryan-mall
    fi
    
    log_success "服务停止完成"
}

# 重启服务
restart_services() {
    log_info "重启服务..."
    stop_services
    sleep 5
    start_services
}

# 查看服务状态
show_status() {
    if [[ "$MODE" == "docker" ]]; then
        cd deployments/docker
        docker-compose ps
    elif [[ "$MODE" == "k8s" ]]; then
        kubectl get pods -n ryan-mall
        kubectl get services -n ryan-mall
    fi
}

# 查看日志
show_logs() {
    if [[ "$MODE" == "docker" ]]; then
        cd deployments/docker
        if [[ -n "$SERVICE" ]]; then
            docker-compose logs -f "$SERVICE"
        else
            docker-compose logs -f
        fi
    elif [[ "$MODE" == "k8s" ]]; then
        if [[ -n "$SERVICE" ]]; then
            kubectl logs -f deployment/"$SERVICE" -n ryan-mall
        else
            kubectl logs -f --all-containers=true -n ryan-mall
        fi
    fi
}

# 清理资源
clean_resources() {
    log_warning "清理资源..."
    
    if [[ "$MODE" == "docker" ]]; then
        cd deployments/docker
        docker-compose down -v
        docker system prune -f
    elif [[ "$MODE" == "k8s" ]]; then
        kubectl delete namespace ryan-mall
    fi
    
    log_success "资源清理完成"
}

# 运行测试
run_tests() {
    log_info "运行测试..."
    
    # 单元测试
    go test ./... -v
    
    # 集成测试（如果服务正在运行）
    if [[ "$MODE" == "docker" ]]; then
        # 等待服务启动
        sleep 30
        
        # 健康检查
        curl -f http://localhost:8080/health || log_error "网关健康检查失败"
    fi
    
    log_success "测试完成"
}

# 主函数
main() {
    log_info "Ryan Mall 微服务部署脚本"
    log_info "环境: $ENVIRONMENT, 模式: $MODE"
    
    check_dependencies
    
    case $COMMAND in
        build)
            build_images
            ;;
        deploy)
            build_images
            if [[ "$MODE" == "docker" ]]; then
                deploy_docker
            elif [[ "$MODE" == "k8s" ]]; then
                deploy_k8s
            fi
            ;;
        start)
            start_services
            ;;
        stop)
            stop_services
            ;;
        restart)
            restart_services
            ;;
        status)
            show_status
            ;;
        logs)
            show_logs
            ;;
        clean)
            clean_resources
            ;;
        test)
            run_tests
            ;;
        *)
            log_error "未知命令: $COMMAND"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main
