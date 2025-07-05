#!/bin/bash

# 简化的Redis集群测试脚本
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
    echo "           Redis集群功能测试"
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
    
    if docker exec ryan-mall-redis-node-1 redis-cli cluster info | grep -q "cluster_state:ok"; then
        print_success "集群状态正常"
        
        echo ""
        echo "📊 集群信息："
        docker exec ryan-mall-redis-node-1 redis-cli cluster info | head -10
        
        echo ""
        echo "🔗 集群节点："
        docker exec ryan-mall-redis-node-1 redis-cli cluster nodes
        
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
    
    # 检查总数据量
    echo ""
    echo "📊 集群数据统计："
    total_keys=$(docker exec ryan-mall-redis-node-1 redis-cli -c dbsize)
    echo "  总键数量: $total_keys"
    
    # 检查各节点数据分布
    echo ""
    echo "📊 各节点数据分布："
    node1_keys=$(docker exec ryan-mall-redis-node-1 redis-cli dbsize)
    node2_keys=$(docker exec ryan-mall-redis-node-2 redis-cli dbsize)
    node3_keys=$(docker exec ryan-mall-redis-node-3 redis-cli dbsize)
    
    echo "  主节点1 (172.20.0.11): $node1_keys 个键"
    echo "  主节点2 (172.20.0.12): $node2_keys 个键"
    echo "  主节点3 (172.20.0.13): $node3_keys 个键"
    
    total_distributed=$((node1_keys + node2_keys + node3_keys))
    echo "  分布式总计: $total_distributed 个键"
}

# 性能测试
test_performance() {
    print_step "性能测试"
    
    echo "执行Redis集群性能测试..."
    
    # 写入性能测试
    echo ""
    echo "📈 写入性能测试 (1000次SET操作):"
    start_time=$(date +%s%N)
    for i in {1..1000}; do
        docker exec ryan-mall-redis-node-1 redis-cli -c set "perf:test:$i" "value$i" > /dev/null
    done
    end_time=$(date +%s%N)
    write_time=$(( (end_time - start_time) / 1000000 ))
    write_qps=$(( 1000 * 1000 / write_time ))
    echo "  写入耗时: ${write_time}ms"
    echo "  写入QPS: $write_qps"
    
    # 读取性能测试
    echo ""
    echo "📈 读取性能测试 (1000次GET操作):"
    start_time=$(date +%s%N)
    for i in {1..1000}; do
        docker exec ryan-mall-redis-node-1 redis-cli -c get "perf:test:$i" > /dev/null
    done
    end_time=$(date +%s%N)
    read_time=$(( (end_time - start_time) / 1000000 ))
    read_qps=$(( 1000 * 1000 / read_time ))
    echo "  读取耗时: ${read_time}ms"
    echo "  读取QPS: $read_qps"
}

# 清理测试数据
cleanup_test_data() {
    print_step "清理测试数据"
    
    echo "删除测试数据..."
    
    # 删除测试键
    docker exec ryan-mall-redis-node-1 redis-cli -c del test:key1 > /dev/null 2>&1 || true
    docker exec ryan-mall-redis-node-1 redis-cli -c del test:hash > /dev/null 2>&1 || true
    docker exec ryan-mall-redis-node-1 redis-cli -c del test:list > /dev/null 2>&1 || true
    docker exec ryan-mall-redis-node-1 redis-cli -c del test:set > /dev/null 2>&1 || true
    
    # 删除分布测试数据
    for i in {1..100}; do
        docker exec ryan-mall-redis-node-1 redis-cli -c del "test:distribution:$i" > /dev/null 2>&1 || true
    done
    
    # 删除性能测试数据
    for i in {1..1000}; do
        docker exec ryan-mall-redis-node-1 redis-cli -c del "perf:test:$i" > /dev/null 2>&1 || true
    done
    
    print_success "测试数据已清理"
}

# 主函数
main() {
    print_header
    
    # 检查集群状态
    if ! check_cluster_status; then
        print_error "集群状态异常，请先启动Redis集群"
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
    
    # 清理测试数据
    cleanup_test_data
    
    print_success "Redis集群测试完成！"
    echo ""
    echo "🎯 集群访问方式："
    echo "   - 容器内访问: docker exec ryan-mall-redis-node-1 redis-cli -c"
    echo "   - 外部访问: redis-cli -h localhost -p 7001 -c"
    echo "   - 集群节点: localhost:7001-7006"
}

# 执行主函数
main "$@"
