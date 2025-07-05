#!/bin/bash

# Redis集群启动脚本
set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

print_header() {
    echo -e "${BLUE}"
    echo "=================================================="
    echo "           Redis集群部署脚本"
    echo "=================================================="
    echo -e "${NC}"
}

print_step() {
    echo -e "${CYAN}=== $1 ===${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# 检查Docker和Docker Compose
check_requirements() {
    print_step "检查系统要求"
    
    if ! command -v docker &> /dev/null; then
        print_error "Docker未安装，请先安装Docker"
        exit 1
    fi
    print_success "Docker已安装"
    
    if ! docker compose version &> /dev/null; then
        print_error "Docker Compose未安装，请先安装Docker Compose"
        exit 1
    fi
    print_success "Docker Compose已安装"
}

# 停止现有服务
stop_existing_services() {
    print_step "停止现有服务"
    
    # 停止单机Redis
    if docker ps | grep -q "ryan-mall-redis"; then
        print_warning "停止现有的单机Redis服务"
        docker stop ryan-mall-redis || true
        docker rm ryan-mall-redis || true
    fi
    
    # 停止现有集群
    if docker ps | grep -q "ryan-mall-redis-node"; then
        print_warning "停止现有的Redis集群"
        docker compose -f docker-compose.redis-cluster.yml down || true
    fi
    
    print_success "现有服务已停止"
}

# 清理数据目录
cleanup_data() {
    print_step "清理数据目录"
    
    if [ -d "docker/redis-cluster" ]; then
        print_warning "清理Redis集群数据目录"
        sudo rm -rf docker/redis-cluster/node-*/nodes-*.conf || true
        sudo rm -rf docker/redis-cluster/node-*/appendonly.aof || true
        sudo rm -rf docker/redis-cluster/node-*/dump.rdb || true
        sudo rm -rf docker/redis-cluster/node-*/redis-*.log || true
    fi
    
    print_success "数据目录已清理"
}

# 启动Redis集群
start_cluster() {
    print_step "启动Redis集群"
    
    print_warning "启动Redis集群节点..."
    docker compose -f docker-compose.redis-cluster.yml up -d redis-node-1 redis-node-2 redis-node-3 redis-node-4 redis-node-5 redis-node-6
    
    if [ $? -eq 0 ]; then
        print_success "Redis集群节点启动成功"
    else
        print_error "Redis集群节点启动失败"
        exit 1
    fi
    
    # 等待节点启动
    print_warning "等待Redis节点完全启动..."
    sleep 15
    
    # 启动集群管理器
    print_warning "初始化Redis集群..."
    docker compose -f docker-compose.redis-cluster.yml up -d redis-cluster-manager
    
    if [ $? -eq 0 ]; then
        print_success "Redis集群管理器启动成功"
    else
        print_error "Redis集群管理器启动失败"
        exit 1
    fi
}

# 启动其他服务
start_other_services() {
    print_step "启动其他服务"
    
    print_warning "启动MySQL数据库..."
    docker compose -f docker-compose.redis-cluster.yml up -d mysql

    print_warning "启动Redis Insight管理界面..."
    docker compose -f docker-compose.redis-cluster.yml up -d redis-insight
    
    print_success "其他服务启动完成"
}

# 检查集群状态
check_cluster_status() {
    print_step "检查集群状态"
    
    print_warning "等待集群初始化完成..."
    sleep 30
    
    # 检查集群状态
    if docker exec ryan-mall-redis-node-1 redis-cli cluster info | grep -q "cluster_state:ok"; then
        print_success "Redis集群状态正常"
        
        echo ""
        echo "📊 集群信息："
        docker exec ryan-mall-redis-node-1 redis-cli cluster info
        
        echo ""
        echo "🔗 集群节点："
        docker exec ryan-mall-redis-node-1 redis-cli cluster nodes
        
    else
        print_warning "集群可能还在初始化中，请稍后检查"
        echo "可以使用以下命令检查集群状态："
        echo "docker exec ryan-mall-redis-node-1 redis-cli cluster info"
    fi
}

# 显示访问信息
show_access_info() {
    print_step "访问信息"
    
    echo "🌐 Redis集群节点："
    echo "   - 节点1: localhost:7001"
    echo "   - 节点2: localhost:7002"
    echo "   - 节点3: localhost:7003"
    echo "   - 节点4: localhost:7004"
    echo "   - 节点5: localhost:7005"
    echo "   - 节点6: localhost:7006"
    echo ""
    echo "🎯 管理界面："
    echo "   - Redis Insight: http://localhost:8001"
    echo "   - phpMyAdmin: http://localhost:8081"
    echo ""
    echo "📝 测试连接："
    echo "   docker exec ryan-mall-redis-node-1 redis-cli -c"
    echo "   docker exec ryan-mall-redis-node-1 redis-cli cluster nodes"
    echo ""
    echo "🔧 环境变量配置："
    echo "   export REDIS_CLUSTER_ENABLED=true"
    echo "   export REDIS_CLUSTER_NODES=localhost:7001,localhost:7002,localhost:7003,localhost:7004,localhost:7005,localhost:7006"
}

# 主函数
main() {
    print_header
    
    check_requirements
    echo ""
    
    stop_existing_services
    echo ""
    
    cleanup_data
    echo ""
    
    start_cluster
    echo ""
    
    start_other_services
    echo ""
    
    check_cluster_status
    echo ""
    
    show_access_info
    
    print_success "Redis集群部署完成！"
    print_warning "请等待1-2分钟让集群完全初始化"
}

# 执行主函数
main "$@"
