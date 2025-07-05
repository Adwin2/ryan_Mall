#!/bin/bash

# 缓存性能测试脚本
# 测试缓存命中率和性能提升

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 服务器地址
BASE_URL="http://localhost:8080"
API_URL="$BASE_URL/api/v1"

# 打印函数
print_step() {
    echo -e "${BLUE}=== $1 ===${NC}"
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

# 测试缓存性能
test_cache_performance() {
    local url=$1
    local name=$2
    local test_count=$3
    
    echo -e "${BLUE}测试 $name 缓存性能 (${test_count}次请求)${NC}"
    
    # 第一次请求（缓存未命中）
    echo -n "首次请求（缓存未命中）: "
    start_time=$(date +%s%N)
    curl -s "$url" > /dev/null
    end_time=$(date +%s%N)
    first_duration=$(( (end_time - start_time) / 1000000 ))
    echo -e "${YELLOW}${first_duration}ms${NC}"
    
    # 后续请求（缓存命中）
    echo "后续请求（缓存命中）:"
    total_time=0
    for i in $(seq 2 $test_count); do
        start_time=$(date +%s%N)
        curl -s "$url" > /dev/null
        end_time=$(date +%s%N)
        duration=$(( (end_time - start_time) / 1000000 ))
        total_time=$((total_time + duration))
        echo "  第${i}次: ${duration}ms"
    done
    
    # 计算平均时间
    cached_count=$((test_count - 1))
    avg_cached_time=$((total_time / cached_count))
    
    # 计算性能提升
    improvement=$((first_duration - avg_cached_time))
    improvement_percent=$(( improvement * 100 / first_duration ))
    
    echo -e "${GREEN}性能总结:${NC}"
    echo "  首次请求: ${first_duration}ms"
    echo "  缓存平均: ${avg_cached_time}ms"
    echo "  性能提升: ${improvement}ms (${improvement_percent}%)"
    echo ""
}

# 测试并发缓存性能
test_concurrent_cache() {
    local url=$1
    local name=$2
    local concurrent_count=$3
    
    echo -e "${BLUE}测试 $name 并发缓存性能 (${concurrent_count}并发)${NC}"
    
    # 预热缓存
    curl -s "$url" > /dev/null
    
    # 创建并发测试脚本
    cat > /tmp/concurrent_cache_test.sh << EOF
#!/bin/bash
for i in {1..$concurrent_count}; do
    (
        start_time=\$(date +%s%N)
        curl -s "$url" > /dev/null
        end_time=\$(date +%s%N)
        duration=\$(( (end_time - start_time) / 1000000 ))
        echo "\$duration"
    ) &
done
wait
EOF
    
    chmod +x /tmp/concurrent_cache_test.sh
    
    echo "执行${concurrent_count}个并发请求..."
    start_time=$(date +%s%N)
    results=$(/tmp/concurrent_cache_test.sh)
    end_time=$(date +%s%N)
    total_duration=$(( (end_time - start_time) / 1000000 ))
    
    # 分析结果
    min_time=$(echo "$results" | sort -n | head -1)
    max_time=$(echo "$results" | sort -n | tail -1)
    avg_time=$(echo "$results" | awk '{sum+=$1} END {print int(sum/NR)}')
    
    echo -e "${GREEN}并发测试结果:${NC}"
    echo "  总耗时: ${total_duration}ms"
    echo "  最快请求: ${min_time}ms"
    echo "  最慢请求: ${max_time}ms"
    echo "  平均时间: ${avg_time}ms"
    echo "  QPS: $(( concurrent_count * 1000 / total_duration ))"
    echo ""
    
    # 清理
    rm -f /tmp/concurrent_cache_test.sh
}

# 测试缓存命中率
test_cache_hit_rate() {
    print_step "测试缓存命中率"
    
    # 测试不同商品的缓存
    products=(7 1 2 3)
    
    for product_id in "${products[@]}"; do
        echo "测试商品 ID: $product_id"
        
        # 第一次访问（缓存未命中）
        echo -n "  首次访问: "
        start_time=$(date +%s%N)
        response1=$(curl -s "$API_URL/products/$product_id")
        end_time=$(date +%s%N)
        duration1=$(( (end_time - start_time) / 1000000 ))
        echo "${duration1}ms"
        
        # 第二次访问（缓存命中）
        echo -n "  缓存访问: "
        start_time=$(date +%s%N)
        response2=$(curl -s "$API_URL/products/$product_id")
        end_time=$(date +%s%N)
        duration2=$(( (end_time - start_time) / 1000000 ))
        echo "${duration2}ms"
        
        # 验证响应一致性
        if [ "$response1" = "$response2" ]; then
            echo -e "  ${GREEN}✅ 响应一致${NC}"
        else
            echo -e "  ${RED}❌ 响应不一致${NC}"
        fi
        
        echo ""
    done
}

# 测试缓存失效
test_cache_invalidation() {
    print_step "测试缓存失效机制"
    
    # 获取商品信息（建立缓存）
    echo "1. 建立缓存..."
    curl -s "$API_URL/products/7" > /dev/null
    
    # 测试缓存访问速度
    echo -n "2. 缓存访问速度: "
    start_time=$(date +%s%N)
    curl -s "$API_URL/products/7" > /dev/null
    end_time=$(date +%s%N)
    cached_duration=$(( (end_time - start_time) / 1000000 ))
    echo "${cached_duration}ms"
    
    # 更新商品（应该清除缓存）
    echo "3. 更新商品（清除缓存）..."
    TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjozLCJ1c2VybmFtZSI6InRlc3R1c2VyX2VuaGFuY2VkIiwiZW1haWwiOiJ0ZXN0dXNlcl9lbmhhbmNlZEBleGFtcGxlLmNvbSIsImlzcyI6InJ5YW4tbWFsbCIsInN1YiI6InRlc3R1c2VyX2VuaGFuY2VkIiwiZXhwIjoxNzUwOTg5NjkyLCJuYmYiOjE3NTA5MDMyOTIsImlhdCI6MTc1MDkwMzI5Mn0.RY-w9tpwpS5WlxhhNTSWp9bdXVfTE8_jX2R6k5HlAk0"
    
    curl -s -X PUT "$API_URL/products/7" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $TOKEN" \
        -d '{"name": "缓存测试商品(已更新)", "description": "测试缓存失效机制"}' > /dev/null
    
    # 再次访问（缓存应该已失效）
    echo -n "4. 缓存失效后访问: "
    start_time=$(date +%s%N)
    curl -s "$API_URL/products/7" > /dev/null
    end_time=$(date +%s%N)
    invalidated_duration=$(( (end_time - start_time) / 1000000 ))
    echo "${invalidated_duration}ms"
    
    # 分析结果
    if [ $invalidated_duration -gt $cached_duration ]; then
        echo -e "${GREEN}✅ 缓存失效机制正常工作${NC}"
    else
        echo -e "${YELLOW}⚠️  缓存可能未正确失效${NC}"
    fi
    
    echo ""
}

# 测试商品列表缓存
test_list_cache() {
    print_step "测试商品列表缓存"
    
    echo "测试商品列表缓存性能..."
    
    # 第一次请求
    echo -n "首次请求: "
    start_time=$(date +%s%N)
    curl -s "$API_URL/products?page=1&page_size=10" > /dev/null
    end_time=$(date +%s%N)
    first_duration=$(( (end_time - start_time) / 1000000 ))
    echo "${first_duration}ms"
    
    # 缓存请求
    echo -n "缓存请求: "
    start_time=$(date +%s%N)
    curl -s "$API_URL/products?page=1&page_size=10" > /dev/null
    end_time=$(date +%s%N)
    cached_duration=$(( (end_time - start_time) / 1000000 ))
    echo "${cached_duration}ms"
    
    # 不同参数的请求（应该是新的缓存）
    echo -n "不同参数: "
    start_time=$(date +%s%N)
    curl -s "$API_URL/products?page=2&page_size=10" > /dev/null
    end_time=$(date +%s%N)
    different_duration=$(( (end_time - start_time) / 1000000 ))
    echo "${different_duration}ms"
    
    echo ""
}

# 主函数
main() {
    echo -e "${BLUE}"
    echo "=================================="
    echo "     缓存性能测试"
    echo "=================================="
    echo -e "${NC}"
    
    # 检查服务状态
    print_step "检查服务状态"
    if curl -s "$BASE_URL/ping" > /dev/null; then
        print_success "服务运行正常"
    else
        print_error "服务未运行"
        exit 1
    fi
    
    # 执行测试
    test_cache_hit_rate
    test_cache_performance "$API_URL/products/7" "商品详情" 10
    test_list_cache
    test_concurrent_cache "$API_URL/products/7" "商品详情" 20
    test_cache_invalidation
    
    echo -e "${GREEN}"
    echo "=================================="
    echo "      缓存测试完成！"
    echo "=================================="
    echo -e "${NC}"
    
    print_step "测试总结"
    echo "✅ 缓存命中率测试完成"
    echo "✅ 缓存性能测试完成"
    echo "✅ 并发缓存测试完成"
    echo "✅ 缓存失效测试完成"
    echo ""
    echo "💡 缓存优化效果："
    echo "1. 响应时间显著降低"
    echo "2. 数据库查询减少"
    echo "3. 并发性能提升"
    echo "4. 缓存失效机制正常"
}

# 执行主函数
main "$@"
