#!/bin/bash

# 高级并发性能测试脚本
# 测试不同并发级别下的性能表现

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 服务器地址
BASE_URL="http://localhost:8080"
API_URL="$BASE_URL/api/v1"

# 测试配置
CONCURRENT_LEVELS=(1 5 10 20 50 100)
REQUEST_COUNT=100
WARMUP_REQUESTS=10

# 打印函数
print_header() {
    echo -e "${BLUE}"
    echo "=================================================="
    echo "           高级并发性能测试"
    echo "=================================================="
    echo -e "${NC}"
}

print_step() {
    echo -e "${CYAN}=== $1 ===${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_info() {
    echo -e "${PURPLE}ℹ️  $1${NC}"
}

# 检查服务状态
check_service() {
    print_step "检查服务状态"
    
    if curl -s "$BASE_URL/ping" > /dev/null; then
        print_success "服务运行正常"
        
        # 获取缓存统计
        cache_stats=$(curl -s "$BASE_URL/cache/stats")
        echo "当前缓存状态: $cache_stats"
    else
        print_error "服务未运行，请先启动服务"
        exit 1
    fi
    echo ""
}

# 预热缓存
warmup_cache() {
    print_step "预热缓存"
    
    echo "执行 $WARMUP_REQUESTS 次预热请求..."
    for i in $(seq 1 $WARMUP_REQUESTS); do
        curl -s "$API_URL/products/7" > /dev/null
        curl -s "$API_URL/products?page=1&page_size=10" > /dev/null
        echo -n "."
    done
    echo ""
    
    # 显示预热后的缓存状态
    cache_stats=$(curl -s "$BASE_URL/cache/stats")
    echo "预热后缓存状态: $cache_stats"
    print_success "缓存预热完成"
    echo ""
}

# 单个并发级别测试
test_concurrent_level() {
    local concurrent=$1
    local url=$2
    local test_name=$3
    
    echo -e "${YELLOW}测试 $test_name - ${concurrent}并发${NC}"
    
    # 创建临时测试脚本
    local test_script="/tmp/concurrent_test_${concurrent}.sh"
    cat > "$test_script" << EOF
#!/bin/bash
for i in \$(seq 1 $REQUEST_COUNT); do
    (
        start_time=\$(date +%s%N)
        response=\$(curl -s -w "%{http_code},%{time_total}" "$url")
        end_time=\$(date +%s%N)

        # 解析响应
        http_code=\$(echo "\$response" | tail -c 8 | cut -d',' -f1)
        curl_time=\$(echo "\$response" | tail -c 8 | cut -d',' -f2)

        # 计算总时间（毫秒）
        total_time=\$(( (end_time - start_time) / 1000000 ))

        echo "\$total_time,\$http_code,\$curl_time"
    ) &

    # 控制并发数
    if (( \$(jobs -r | wc -l) >= $concurrent )); then
        wait -n  # 等待任意一个后台任务完成
    fi
done
wait  # 等待所有任务完成
EOF
    
    chmod +x "$test_script"
    
    # 执行测试
    echo "执行 ${REQUEST_COUNT} 个请求，${concurrent} 并发..."
    start_time=$(date +%s%N)
    results=$("$test_script")
    end_time=$(date +%s%N)
    
    # 计算总耗时
    total_duration=$(( (end_time - start_time) / 1000000 ))
    
    # 分析结果
    if [ -n "$results" ]; then
        # 解析结果
        times=$(echo "$results" | cut -d',' -f1)
        codes=$(echo "$results" | cut -d',' -f2)
        
        # 统计成功请求
        success_count=$(echo "$codes" | grep -c "200" || echo "0")
        error_count=$(( REQUEST_COUNT - success_count ))
        
        # 计算时间统计
        min_time=$(echo "$times" | sort -n | head -1)
        max_time=$(echo "$times" | sort -n | tail -1)
        avg_time=$(echo "$times" | awk '{sum+=$1} END {print int(sum/NR)}')
        
        # 计算百分位数
        p95_time=$(echo "$times" | sort -n | awk 'BEGIN{c=0} {a[c++]=$1} END{print a[int(c*0.95)]}')
        p99_time=$(echo "$times" | sort -n | awk 'BEGIN{c=0} {a[c++]=$1} END{print a[int(c*0.99)]}')
        
        # 计算QPS
        qps=$(( success_count * 1000 / total_duration ))
        
        # 输出结果
        echo "  总耗时: ${total_duration}ms"
        echo "  成功请求: ${success_count}/${REQUEST_COUNT}"
        echo "  错误请求: ${error_count}"
        echo "  QPS: ${qps}"
        echo "  响应时间统计:"
        echo "    最小: ${min_time}ms"
        echo "    最大: ${max_time}ms"
        echo "    平均: ${avg_time}ms"
        echo "    P95: ${p95_time}ms"
        echo "    P99: ${p99_time}ms"
        
        # 返回QPS用于汇总
        echo "$concurrent,$qps,$avg_time,$p95_time,$success_count,$error_count" >> /tmp/performance_results.csv
    else
        print_error "测试失败，无结果数据"
        echo "$concurrent,0,0,0,0,$REQUEST_COUNT" >> /tmp/performance_results.csv
    fi
    
    # 清理
    rm -f "$test_script"
    echo ""
}

# 压力测试
stress_test() {
    print_step "压力测试 - 寻找性能极限"
    
    # 初始化结果文件
    echo "concurrent,qps,avg_time,p95_time,success,errors" > /tmp/performance_results.csv
    
    # 测试商品详情接口
    echo -e "${PURPLE}测试商品详情接口 (/api/v1/products/7)${NC}"
    for concurrent in "${CONCURRENT_LEVELS[@]}"; do
        test_concurrent_level "$concurrent" "$API_URL/products/7" "商品详情"
        sleep 2  # 间隔2秒
    done
    
    echo ""
    echo -e "${PURPLE}测试商品列表接口 (/api/v1/products)${NC}"
    for concurrent in "${CONCURRENT_LEVELS[@]}"; do
        test_concurrent_level "$concurrent" "$API_URL/products?page=1&page_size=10" "商品列表"
        sleep 2  # 间隔2秒
    done
}

# 生成性能报告
generate_report() {
    print_step "生成性能报告"
    
    if [ -f /tmp/performance_results.csv ]; then
        echo -e "${GREEN}性能测试汇总报告${NC}"
        echo "=================================================="
        printf "%-12s %-8s %-12s %-12s %-10s %-8s\n" "并发数" "QPS" "平均响应" "P95响应" "成功率" "错误数"
        echo "=================================================="
        
        while IFS=',' read -r concurrent qps avg_time p95_time success errors; do
            if [ "$concurrent" != "concurrent" ]; then  # 跳过标题行
                success_rate=$(( success * 100 / REQUEST_COUNT ))
                printf "%-12s %-8s %-12s %-12s %-10s %-8s\n" \
                    "${concurrent}" "${qps}" "${avg_time}ms" "${p95_time}ms" "${success_rate}%" "${errors}"
            fi
        done < /tmp/performance_results.csv
        
        echo "=================================================="
        
        # 找出最佳性能点
        best_qps=$(tail -n +2 /tmp/performance_results.csv | cut -d',' -f2 | sort -nr | head -1)
        best_concurrent=$(tail -n +2 /tmp/performance_results.csv | awk -F',' -v max="$best_qps" '$2==max {print $1}' | head -1)
        
        echo ""
        print_success "最佳性能: ${best_concurrent}并发 → ${best_qps} QPS"
        
        # 性能建议
        echo ""
        print_info "性能优化建议:"
        if [ "$best_qps" -lt 1000 ]; then
            echo "  • 考虑优化数据库连接池配置"
            echo "  • 检查是否有慢查询"
            echo "  • 考虑增加缓存层"
        elif [ "$best_qps" -lt 5000 ]; then
            echo "  • 性能良好，可考虑进一步优化缓存策略"
            echo "  • 监控系统资源使用情况"
        else
            echo "  • 性能优秀！"
            echo "  • 可以考虑增加更多功能特性"
        fi
    else
        print_error "未找到性能测试结果文件"
    fi
    
    echo ""
}

# 资源监控
monitor_resources() {
    print_step "系统资源监控"
    
    echo "CPU使用率:"
    top -bn1 | grep "Cpu(s)" | awk '{print $2}' | cut -d'%' -f1
    
    echo "内存使用情况:"
    free -h
    
    echo "网络连接数:"
    netstat -an | grep :8080 | wc -l
    
    echo ""
}

# 缓存性能分析
analyze_cache_performance() {
    print_step "缓存性能分析"
    
    # 获取缓存统计
    cache_stats=$(curl -s "$BASE_URL/cache/stats")
    echo "当前缓存统计: $cache_stats"
    
    # 测试缓存命中率
    echo ""
    echo "测试缓存命中率..."
    
    # 清空缓存
    echo "清空缓存..."
    # 这里可以添加清空缓存的API调用
    
    # 第一次请求（缓存未命中）
    echo -n "首次请求（缓存未命中）: "
    start_time=$(date +%s%N)
    curl -s "$API_URL/products/7" > /dev/null
    end_time=$(date +%s%N)
    miss_time=$(( (end_time - start_time) / 1000000 ))
    echo "${miss_time}ms"
    
    # 第二次请求（缓存命中）
    echo -n "缓存命中请求: "
    start_time=$(date +%s%N)
    curl -s "$API_URL/products/7" > /dev/null
    end_time=$(date +%s%N)
    hit_time=$(( (end_time - start_time) / 1000000 ))
    echo "${hit_time}ms"
    
    # 计算缓存效果
    if [ "$miss_time" -gt 0 ]; then
        improvement=$(( (miss_time - hit_time) * 100 / miss_time ))
        echo "缓存性能提升: ${improvement}%"
    fi
    
    echo ""
}

# 主函数
main() {
    print_header
    
    # 检查依赖
    if ! command -v curl &> /dev/null; then
        print_error "curl 未安装"
        exit 1
    fi
    
    # 执行测试
    check_service
    monitor_resources
    warmup_cache
    analyze_cache_performance
    stress_test
    generate_report
    
    # 清理临时文件
    rm -f /tmp/performance_results.csv
    
    echo -e "${GREEN}"
    echo "=================================================="
    echo "           并发性能测试完成！"
    echo "=================================================="
    echo -e "${NC}"
}

# 执行主函数
main "$@"
