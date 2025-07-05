#!/bin/bash

# Redis集群 vs 内存缓存性能对比测试
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
    echo "     Redis集群 vs 内存缓存 性能对比"
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

# 测试内存缓存性能
test_memory_cache() {
    print_step "测试内存缓存性能"
    
    # 确保Ryan Mall服务正在运行
    if ! curl -s http://localhost:8080/ping > /dev/null; then
        print_error "Ryan Mall服务未运行，请先启动服务"
        return 1
    fi
    
    print_warning "预热内存缓存..."
    for i in {1..5}; do
        curl -s "http://localhost:8080/api/v1/products/7" > /dev/null
    done
    
    echo "开始内存缓存性能测试..."
    
    # 测试不同并发级别
    for concurrent in 10 50 100; do
        echo ""
        echo "=== ${concurrent}并发测试 (内存缓存) ==="
        
        start_time=$(date +%s%N)
        
        # 执行并发请求
        seq 1 $concurrent | xargs -n1 -P${concurrent} -I{} bash -c "
            start=\$(date +%s%N)
            if curl -s --max-time 5 'http://localhost:8080/api/v1/products/7' > /dev/null 2>&1; then
                end=\$(date +%s%N)
                echo \$(( (end - start) / 1000000 ))
            else
                echo 'ERROR'
            fi
        " > /tmp/memory_cache_results.txt
        
        end_time=$(date +%s%N)
        total_time=$(( (end_time - start_time) / 1000000 ))
        
        # 统计结果
        success_count=$(grep -v ERROR /tmp/memory_cache_results.txt | wc -l)
        error_count=$(grep ERROR /tmp/memory_cache_results.txt | wc -l || echo 0)
        
        if [ $success_count -gt 0 ]; then
            qps=$(( concurrent * 1000 / total_time ))
            avg_time=$(grep -v ERROR /tmp/memory_cache_results.txt | awk '{sum+=$1} END {print int(sum/NR)}')
            
            echo "  内存缓存 - 总耗时: ${total_time}ms, QPS: $qps, 平均响应: ${avg_time}ms, 成功率: $(( success_count * 100 / concurrent ))%"
            
            # 保存结果
            echo "$concurrent,$qps,$avg_time,memory" >> /tmp/performance_comparison.csv
        fi
        
        sleep 1
    done
    
    rm -f /tmp/memory_cache_results.txt
}

# 测试Redis集群性能
test_redis_cluster() {
    print_step "测试Redis集群性能"
    
    # 检查Redis集群是否运行
    if ! docker exec ryan-mall-redis-node-1 redis-cli cluster info | grep -q "cluster_state:ok" 2>/dev/null; then
        print_error "Redis集群未运行，请先启动集群"
        return 1
    fi
    
    print_warning "预热Redis集群..."
    
    # 预热Redis集群
    for i in {1..5}; do
        docker exec ryan-mall-redis-node-1 redis-cli -c set "product:7" '{"id":7,"name":"测试商品","price":99.99}' > /dev/null
        docker exec ryan-mall-redis-node-1 redis-cli -c get "product:7" > /dev/null
    done
    
    echo "开始Redis集群性能测试..."
    
    # 测试不同并发级别
    for concurrent in 10 50 100; do
        echo ""
        echo "=== ${concurrent}并发测试 (Redis集群) ==="
        
        start_time=$(date +%s%N)
        
        # 执行并发Redis操作
        seq 1 $concurrent | xargs -n1 -P${concurrent} -I{} bash -c "
            start=\$(date +%s%N)
            if docker exec ryan-mall-redis-node-1 redis-cli -c get 'product:7' > /dev/null 2>&1; then
                end=\$(date +%s%N)
                echo \$(( (end - start) / 1000000 ))
            else
                echo 'ERROR'
            fi
        " > /tmp/redis_cluster_results.txt
        
        end_time=$(date +%s%N)
        total_time=$(( (end_time - start_time) / 1000000 ))
        
        # 统计结果
        success_count=$(grep -v ERROR /tmp/redis_cluster_results.txt | wc -l)
        error_count=$(grep ERROR /tmp/redis_cluster_results.txt | wc -l || echo 0)
        
        if [ $success_count -gt 0 ]; then
            qps=$(( concurrent * 1000 / total_time ))
            avg_time=$(grep -v ERROR /tmp/redis_cluster_results.txt | awk '{sum+=$1} END {print int(sum/NR)}')
            
            echo "  Redis集群 - 总耗时: ${total_time}ms, QPS: $qps, 平均响应: ${avg_time}ms, 成功率: $(( success_count * 100 / concurrent ))%"
            
            # 保存结果
            echo "$concurrent,$qps,$avg_time,redis" >> /tmp/performance_comparison.csv
        fi
        
        sleep 1
    done
    
    rm -f /tmp/redis_cluster_results.txt
}

# 生成性能对比报告
generate_report() {
    print_step "生成性能对比报告"
    
    if [ ! -f /tmp/performance_comparison.csv ]; then
        print_error "没有找到测试结果文件"
        return 1
    fi
    
    echo "📊 性能对比报告"
    echo "==============================================="
    echo ""
    
    echo "| 并发级别 | 缓存类型 | QPS | 平均响应时间(ms) |"
    echo "|---------|---------|-----|-----------------|"
    
    # 读取并格式化结果
    while IFS=',' read -r concurrent qps avg_time cache_type; do
        cache_name=""
        if [ "$cache_type" = "memory" ]; then
            cache_name="内存缓存"
        else
            cache_name="Redis集群"
        fi
        printf "| %-7s | %-7s | %-3s | %-13s |\n" "$concurrent" "$cache_name" "$qps" "$avg_time"
    done < /tmp/performance_comparison.csv
    
    echo ""
    echo "📈 性能分析："
    
    # 计算平均性能
    if grep -q "memory" /tmp/performance_comparison.csv; then
        memory_avg_qps=$(grep "memory" /tmp/performance_comparison.csv | awk -F',' '{sum+=$2} END {print int(sum/NR)}')
        memory_avg_time=$(grep "memory" /tmp/performance_comparison.csv | awk -F',' '{sum+=$3} END {print int(sum/NR)}')
        echo "  内存缓存平均QPS: $memory_avg_qps"
        echo "  内存缓存平均响应时间: ${memory_avg_time}ms"
    fi
    
    if grep -q "redis" /tmp/performance_comparison.csv; then
        redis_avg_qps=$(grep "redis" /tmp/performance_comparison.csv | awk -F',' '{sum+=$2} END {print int(sum/NR)}')
        redis_avg_time=$(grep "redis" /tmp/performance_comparison.csv | awk -F',' '{sum+=$3} END {print int(sum/NR)}')
        echo "  Redis集群平均QPS: $redis_avg_qps"
        echo "  Redis集群平均响应时间: ${redis_avg_time}ms"
    fi
    
    echo ""
    echo "🎯 优化建议："
    echo "  1. 内存缓存适合单机高性能场景"
    echo "  2. Redis集群适合分布式、高可用场景"
    echo "  3. 可以结合使用：热点数据用内存缓存，持久化数据用Redis集群"
    echo "  4. Redis集群提供数据分片和故障转移能力"
}

# 显示Redis集群特性
show_cluster_features() {
    print_step "Redis集群特性展示"
    
    echo "🔧 集群配置信息："
    echo "  - 节点数量: 6个 (3主3从)"
    echo "  - 数据分片: 16384个槽位"
    echo "  - 高可用: 自动故障转移"
    echo "  - 负载均衡: 读写分离"
    
    echo ""
    echo "📊 当前集群状态："
    docker exec ryan-mall-redis-node-1 redis-cli cluster info | grep -E "(cluster_state|cluster_size|cluster_known_nodes)"
    
    echo ""
    echo "🔗 节点分布："
    docker exec ryan-mall-redis-node-1 redis-cli cluster nodes | awk '{print $2, $3, $4}' | head -6
}

# 主函数
main() {
    print_header
    
    # 初始化结果文件
    echo "concurrent,qps,avg_time,cache_type" > /tmp/performance_comparison.csv
    
    # 测试内存缓存
    test_memory_cache
    echo ""
    
    # 测试Redis集群
    test_redis_cluster
    echo ""
    
    # 显示集群特性
    show_cluster_features
    echo ""
    
    # 生成对比报告
    generate_report
    echo ""
    
    print_success "性能对比测试完成！"
    
    # 清理临时文件
    rm -f /tmp/performance_comparison.csv
}

# 执行主函数
main "$@"
