#!/bin/bash

# 性能测试脚本
# 测试API响应时间和并发性能

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

# 测试单个API响应时间
test_api_response_time() {
    local url=$1
    local name=$2
    local headers=$3
    
    echo -n "测试 $name: "
    
    # 预热请求
    curl -s $headers "$url" > /dev/null
    
    # 测试5次取平均值
    total_time=0
    for i in {1..5}; do
        start_time=$(date +%s%N)
        curl -s $headers "$url" > /dev/null
        end_time=$(date +%s%N)
        duration=$(( (end_time - start_time) / 1000000 ))
        total_time=$((total_time + duration))
    done
    
    avg_time=$((total_time / 5))
    
    if [ $avg_time -lt 100 ]; then
        echo -e "${GREEN}${avg_time}ms (优秀)${NC}"
    elif [ $avg_time -lt 500 ]; then
        echo -e "${YELLOW}${avg_time}ms (良好)${NC}"
    else
        echo -e "${RED}${avg_time}ms (需要优化)${NC}"
    fi
}

# 并发测试
test_concurrent_requests() {
    local url=$1
    local name=$2
    local concurrent_count=$3
    local headers=$4
    
    echo -n "并发测试 $name ($concurrent_count 并发): "
    
    # 创建临时脚本
    cat > /tmp/concurrent_test.sh << EOF
#!/bin/bash
for i in {1..$concurrent_count}; do
    curl -s $headers "$url" > /dev/null &
done
wait
EOF
    
    chmod +x /tmp/concurrent_test.sh
    
    start_time=$(date +%s%N)
    /tmp/concurrent_test.sh
    end_time=$(date +%s%N)
    duration=$(( (end_time - start_time) / 1000000 ))
    
    echo -e "${GREEN}${duration}ms${NC}"
    
    # 清理
    rm -f /tmp/concurrent_test.sh
}

# 主测试函数
main() {
    echo -e "${BLUE}"
    echo "=================================="
    echo "     Ryan Mall 性能测试"
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
    
    # API响应时间测试
    print_step "API响应时间测试"
    
    test_api_response_time "$BASE_URL/ping" "健康检查" ""
    test_api_response_time "$API_URL/products" "商品列表" ""
    test_api_response_time "$API_URL/products/7" "商品详情" ""
    test_api_response_time "$API_URL/categories" "分类列表" ""
    
    # 并发测试
    print_step "并发性能测试"
    
    test_concurrent_requests "$BASE_URL/ping" "健康检查" 10 ""
    test_concurrent_requests "$API_URL/products" "商品列表" 10 ""
    test_concurrent_requests "$API_URL/products/7" "商品详情" 20 ""
    
    # 压力测试（如果有ab工具）
    if command -v ab > /dev/null; then
        print_step "Apache Bench 压力测试"
        
        echo "商品列表 - 100请求，10并发:"
        ab -n 100 -c 10 -q "$API_URL/products" | grep -E "Requests per second|Time per request"
        
        echo "商品详情 - 100请求，10并发:"
        ab -n 100 -c 10 -q "$API_URL/products/7" | grep -E "Requests per second|Time per request"
    else
        print_warning "Apache Bench (ab) 未安装，跳过压力测试"
        echo "安装命令: sudo apt-get install apache2-utils"
    fi
    
    # 内存和CPU使用情况
    print_step "系统资源使用情况"
    
    echo "内存使用情况:"
    free -h | head -2
    
    echo "CPU使用情况:"
    top -bn1 | grep "Cpu(s)" | head -1
    
    echo "Go进程资源使用:"
    ps aux | grep "go run\|ryan-mall" | grep -v grep | head -5
    
    echo -e "${GREEN}"
    echo "=================================="
    echo "      性能测试完成！"
    echo "=================================="
    echo -e "${NC}"
    
    print_step "性能总结"
    echo "✅ 基础API响应时间测试完成"
    echo "✅ 并发性能测试完成"
    echo "✅ 系统资源监控完成"
    echo ""
    echo "💡 性能优化建议："
    echo "1. 添加Redis缓存减少数据库查询"
    echo "2. 使用连接池优化数据库连接"
    echo "3. 添加CDN加速静态资源"
    echo "4. 实现API响应压缩"
    echo "5. 添加数据库查询索引优化"
}

# 执行主函数
main "$@"
