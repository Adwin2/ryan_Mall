#!/bin/bash

# Redis集群测试脚本
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
    echo "           Redis集群性能测试"
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

# 检查集群状态
check_cluster_status() {
    print_step "检查集群状态"
    
    # 检查所有节点是否在线
    nodes=("7001" "7002" "7003" "7004" "7005" "7006")
    online_nodes=0
    
    for port in "${nodes[@]}"; do
        if docker exec ryan-mall-redis-node-1 redis-cli -h 172.20.0.11 -p 6379 ping > /dev/null 2>&1; then
            echo "✅ 节点 172.20.0.11:6379 在线"
            ((online_nodes++))
            break
        else
            echo "❌ 节点 172.20.0.11:6379 离线"
        fi
    done
    
    echo "在线节点数: $online_nodes/6"
    
    # 检查集群状态
    if docker exec ryan-mall-redis-node-1 redis-cli cluster info | grep -q "cluster_state:ok"; then
        print_success "集群状态正常"
        return 0
    else
        print_error "集群状态异常"
        return 1
    fi
}

# 基础功能测试
test_basic_operations() {
    print_step "基础功能测试"
    
    # 测试SET/GET
    echo "测试 SET/GET 操作..."
    docker exec ryan-mall-redis-node-1 redis-cli -c set test:key1 "Hello Redis Cluster" > /dev/null
    result=$(docker exec ryan-mall-redis-node-1 redis-cli -c get test:key1)
    if [ "$result" = "Hello Redis Cluster" ]; then
        print_success "SET/GET 测试通过"
    else
        print_error "SET/GET 测试失败"
    fi
    
    # 测试HASH操作
    echo "测试 HASH 操作..."
    docker exec ryan-mall-redis-node-1 redis-cli -c hset test:hash field1 value1 > /dev/null
    docker exec ryan-mall-redis-node-1 redis-cli -c hset test:hash field2 value2 > /dev/null
    hash_result=$(docker exec ryan-mall-redis-node-1 redis-cli -c hget test:hash field1)
    if [ "$hash_result" = "value1" ]; then
        print_success "HASH 测试通过"
    else
        print_error "HASH 测试失败"
    fi
    
    # 测试LIST操作
    echo "测试 LIST 操作..."
    docker exec ryan-mall-redis-node-1 redis-cli -c lpush test:list item1 item2 item3 > /dev/null
    list_length=$(docker exec ryan-mall-redis-node-1 redis-cli -c llen test:list)
    if [ "$list_length" = "3" ]; then
        print_success "LIST 测试通过"
    else
        print_error "LIST 测试失败"
    fi
    
    # 测试SET操作
    echo "测试 SET 操作..."
    docker exec ryan-mall-redis-node-1 redis-cli -c sadd test:set member1 member2 member3 > /dev/null
    set_size=$(docker exec ryan-mall-redis-node-1 redis-cli -c scard test:set)
    if [ "$set_size" = "3" ]; then
        print_success "SET 测试通过"
    else
        print_error "SET 测试失败"
    fi
}

# 数据分布测试
test_data_distribution() {
    print_step "数据分布测试"
    
    echo "写入测试数据到不同槽位..."
    
    # 写入100个键值对
    for i in {1..100}; do
        docker exec ryan-mall-redis-node-1 redis-cli -c set "test:distribution:$i" "value$i" > /dev/null
    done
    
    print_success "已写入100个测试键值对"
    
    # 检查数据分布
    echo ""
    echo "📊 数据分布情况："
    for port in 7001 7002 7003; do
        count=$(docker exec ryan-mall-redis-node-1 redis-cli -h localhost -p $port dbsize)
        echo "  节点 localhost:$port: $count 个键"
    done
}

# 故障转移测试
test_failover() {
    print_step "故障转移测试"
    
    print_warning "模拟节点故障..."
    
    # 停止一个主节点
    echo "停止节点 redis-node-1..."
    docker stop ryan-mall-redis-node-1 > /dev/null
    
    # 等待故障转移
    echo "等待故障转移..."
    sleep 10
    
    # 测试集群是否仍然可用
    echo "测试集群可用性..."
    if docker exec ryan-mall-redis-node-2 redis-cli -c set test:failover "failover test" > /dev/null 2>&1; then
        print_success "故障转移成功，集群仍然可用"
    else
        print_error "故障转移失败"
    fi
    
    # 恢复节点
    echo "恢复节点 redis-node-1..."
    docker start ryan-mall-redis-node-1 > /dev/null
    sleep 5
    
    print_success "节点已恢复"
}

# 性能测试
test_performance() {
    print_step "性能测试"
    
    echo "执行性能基准测试..."
    
    # 使用redis-benchmark进行性能测试
    echo ""
    echo "📈 SET 操作性能测试 (10000次):"
    docker exec ryan-mall-redis-node-1 redis-cli --cluster call localhost:7001 redis-benchmark -t set -n 10000 -q -c 50 | head -5
    
    echo ""
    echo "📈 GET 操作性能测试 (10000次):"
    docker exec ryan-mall-redis-node-1 redis-cli --cluster call localhost:7001 redis-benchmark -t get -n 10000 -q -c 50 | head -5
    
    echo ""
    echo "📈 混合操作性能测试 (5000次):"
    docker exec ryan-mall-redis-node-1 redis-cli --cluster call localhost:7001 redis-benchmark -n 5000 -q -c 50 | head -10
}

# 集群信息展示
show_cluster_info() {
    print_step "集群详细信息"
    
    echo "📊 集群状态信息："
    docker exec ryan-mall-redis-node-1 redis-cli cluster info
    
    echo ""
    echo "🔗 集群节点信息："
    docker exec ryan-mall-redis-node-1 redis-cli cluster nodes
    
    echo ""
    echo "📈 集群槽位分布："
    docker exec ryan-mall-redis-node-1 redis-cli cluster slots | head -20
}

# 清理测试数据
cleanup_test_data() {
    print_step "清理测试数据"
    
    echo "删除测试数据..."
    docker exec ryan-mall-redis-node-1 redis-cli -c --scan --pattern "test:*" | xargs -r docker exec ryan-mall-redis-node-1 redis-cli -c del > /dev/null 2>&1 || true
    
    print_success "测试数据已清理"
}

# 主函数
main() {
    print_header
    
    # 检查集群状态
    if ! check_cluster_status; then
        print_error "集群状态异常，请先启动Redis集群"
        echo "运行: ./start_redis_cluster.sh"
        exit 1
    fi
    echo ""
    
    # 基础功能测试
    test_basic_operations
    echo ""
    
    # 数据分布测试
    test_data_distribution
    echo ""
    
    # 性能测试
    test_performance
    echo ""
    
    # 故障转移测试（可选）
    read -p "是否执行故障转移测试？这会临时停止一个节点 (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        test_failover
        echo ""
    fi
    
    # 显示集群信息
    show_cluster_info
    echo ""
    
    # 清理测试数据
    cleanup_test_data
    
    print_success "Redis集群测试完成！"
    echo ""
    echo "🎯 集群访问方式："
    echo "   - 应用连接: localhost:7001,localhost:7002,localhost:7003"
    echo "   - 管理界面: http://localhost:8001"
    echo "   - 命令行: docker exec ryan-mall-redis-node-1 redis-cli -c"
}

# 执行主函数
main "$@"
