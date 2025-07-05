#!/bin/bash

# 性能分析工具 - 找出QPS下降的原因
set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

BASE_URL="http://localhost:8080"
API_URL="$BASE_URL/api/v1"

print_header() {
    echo -e "${BLUE}"
    echo "=================================================="
    echo "           性能下降原因分析"
    echo "=================================================="
    echo -e "${NC}"
}

print_step() {
    echo -e "${CYAN}=== $1 ===${NC}"
}

# 基准测试函数
benchmark_test() {
    local test_name=$1
    local concurrent=$2
    local requests=$3
    
    echo -e "${YELLOW}测试: $test_name${NC}"
    echo "并发: $concurrent, 请求数: $requests"
    
    # 预热
    curl -s "$API_URL/products/7" > /dev/null
    
    # 执行测试
    start_time=$(date +%s%N)
    seq 1 $requests | xargs -n1 -P${concurrent} -I{} bash -c "
        start=\$(date +%s%N)
        curl -s '$API_URL/products/7' > /dev/null
        end=\$(date +%s%N)
        echo \$(( (end - start) / 1000000 ))
    " > /tmp/times.txt
    end_time=$(date +%s%N)
    
    total_time=$(( (end_time - start_time) / 1000000 ))
    qps=$(( requests * 1000 / total_time ))
    avg_time=$(awk '{sum+=$1} END {print int(sum/NR)}' /tmp/times.txt)
    min_time=$(sort -n /tmp/times.txt | head -1)
    max_time=$(sort -n /tmp/times.txt | tail -1)
    
    echo "  QPS: $qps"
    echo "  总耗时: ${total_time}ms"
    echo "  平均响应: ${avg_time}ms"
    echo "  最小响应: ${min_time}ms"
    echo "  最大响应: ${max_time}ms"
    echo ""
    
    rm -f /tmp/times.txt
    echo "$test_name,$qps,$avg_time,$min_time,$max_time" >> /tmp/performance_comparison.csv
}

# 测试缓存性能影响
test_cache_performance() {
    print_step "测试缓存系统性能影响"
    
    echo "测试当前分片缓存性能..."
    
    # 测试缓存命中性能
    echo "1. 缓存命中性能测试"
    for i in {1..5}; do
        curl -s "$API_URL/products/7" > /dev/null  # 预热
    done
    
    # 测试纯缓存访问速度
    echo "测试100次缓存访问..."
    start_time=$(date +%s%N)
    for i in {1..100}; do
        curl -s "$API_URL/products/7" > /dev/null
    done
    end_time=$(date +%s%N)
    
    cache_total_time=$(( (end_time - start_time) / 1000000 ))
    cache_avg_time=$(( cache_total_time / 100 ))
    cache_qps=$(( 100 * 1000 / cache_total_time ))
    
    echo "  缓存QPS: $cache_qps"
    echo "  缓存平均响应: ${cache_avg_time}ms"
    echo ""
}

# 测试数据库连接池影响
test_db_pool_impact() {
    print_step "测试数据库连接池影响"
    
    echo "当前数据库连接池状态:"
    curl -s "$BASE_URL/db/stats" | jq '.data'
    echo ""
    
    # 测试数据库未命中性能（清空缓存后的第一次请求）
    echo "测试数据库查询性能（缓存未命中）..."
    
    # 这里我们无法直接清空缓存，但可以访问不同的商品ID
    total_db_time=0
    for product_id in {1..10}; do
        start_time=$(date +%s%N)
        curl -s "$API_URL/products/$product_id" > /dev/null
        end_time=$(date +%s%N)
        db_time=$(( (end_time - start_time) / 1000000 ))
        total_db_time=$((total_db_time + db_time))
        echo "  商品$product_id: ${db_time}ms"
    done
    
    avg_db_time=$((total_db_time / 10))
    echo "  数据库平均响应: ${avg_db_time}ms"
    echo ""
}

# 测试系统资源使用
test_system_resources() {
    print_step "测试系统资源使用情况"
    
    echo "CPU使用率:"
    top -bn1 | grep "Cpu(s)" | awk '{print $2}' | cut -d'%' -f1
    
    echo "内存使用:"
    free -h | grep "内存"
    
    echo "网络连接数:"
    netstat -an | grep :8080 | wc -l
    
    echo "Go进程信息:"
    ps aux | grep "go run" | grep -v grep
    echo ""
}

# 对比测试不同配置
comparative_test() {
    print_step "对比测试不同并发级别"
    
    echo "concurrent,qps,avg_time,min_time,max_time" > /tmp/performance_comparison.csv
    
    # 测试不同并发级别
    benchmark_test "10并发" 10 50
    benchmark_test "20并发" 20 50  
    benchmark_test "50并发" 50 50
    benchmark_test "100并发" 100 100
    
    echo "性能对比结果:"
    echo "=================================================="
    printf "%-12s %-8s %-12s %-12s %-12s\n" "测试场景" "QPS" "平均响应" "最小响应" "最大响应"
    echo "=================================================="
    
    while IFS=',' read -r test_name qps avg_time min_time max_time; do
        if [ "$test_name" != "concurrent" ]; then
            printf "%-12s %-8s %-12s %-12s %-12s\n" "$test_name" "$qps" "${avg_time}ms" "${min_time}ms" "${max_time}ms"
        fi
    done < /tmp/performance_comparison.csv
    
    echo "=================================================="
}

# 分析性能瓶颈
analyze_bottlenecks() {
    print_step "分析性能瓶颈"
    
    echo "1. 缓存系统分析:"
    cache_stats=$(curl -s "$BASE_URL/cache/stats")
    echo "$cache_stats" | jq '.'
    
    echo ""
    echo "2. 数据库连接池分析:"
    db_stats=$(curl -s "$BASE_URL/db/stats")
    echo "$db_stats" | jq '.'
    
    echo ""
    echo "3. 系统负载分析:"
    uptime
    
    echo ""
    echo "4. 内存使用分析:"
    free -m
    
    echo ""
    echo "5. 网络连接分析:"
    ss -tuln | grep :8080
}

# 提供优化建议
provide_recommendations() {
    print_step "性能优化建议"
    
    echo "基于测试结果，可能的优化方向:"
    echo ""
    echo "1. 缓存系统优化:"
    echo "   - 考虑减少分片数量（从32减少到8或16）"
    echo "   - 或者回退到简单缓存进行对比测试"
    echo ""
    echo "2. 数据库连接池优化:"
    echo "   - 增加ConnMaxIdleTime到5-10分钟"
    echo "   - 监控连接池使用情况"
    echo ""
    echo "3. HTTP服务器配置优化:"
    echo "   - 调整超时参数"
    echo "   - 测试默认配置的性能"
    echo ""
    echo "4. 系统级优化:"
    echo "   - 检查系统资源限制"
    echo "   - 优化网络参数"
}

# 主函数
main() {
    print_header
    
    # 检查服务状态
    if ! curl -s "$BASE_URL/ping" > /dev/null; then
        echo -e "${RED}❌ 服务未运行${NC}"
        exit 1
    fi
    echo -e "${GREEN}✅ 服务运行正常${NC}"
    echo ""
    
    # 执行各项测试
    test_system_resources
    test_cache_performance
    test_db_pool_impact
    comparative_test
    analyze_bottlenecks
    provide_recommendations
    
    # 清理临时文件
    rm -f /tmp/performance_comparison.csv /tmp/times.txt
    
    echo -e "${GREEN}"
    echo "=================================================="
    echo "           性能分析完成！"
    echo "=================================================="
    echo -e "${NC}"
}

# 执行主函数
main "$@"
