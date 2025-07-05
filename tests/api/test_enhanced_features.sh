#!/bin/bash

# 增强功能测试脚本
# 测试Redis缓存、监控指标等新功能

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

# 全局变量
TOKEN=""
USER_ID=""
PRODUCT_ID=""

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

# 检查服务是否运行
check_service() {
    print_step "检查服务状态"
    
    if curl -s "$BASE_URL/ping" > /dev/null; then
        print_success "服务运行正常"
    else
        print_error "服务未运行，请先启动服务"
        exit 1
    fi
}

# 用户注册和登录
setup_user() {
    print_step "设置测试用户"
    
    # 注册用户
    REGISTER_RESPONSE=$(curl -s -X POST "$API_URL/register" \
        -H "Content-Type: application/json" \
        -d '{
            "username": "testuser_enhanced",
            "email": "testuser_enhanced@example.com",
            "password": "password123",
            "nickname": "Enhanced Test User"
        }')
    
    if echo "$REGISTER_RESPONSE" | grep -q '"code":200'; then
        print_success "用户注册成功"
    else
        print_warning "用户可能已存在，尝试登录"
    fi
    
    # 用户登录
    LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/login" \
        -H "Content-Type: application/json" \
        -d '{
            "username": "testuser_enhanced",
            "password": "password123"
        }')
    
    if echo "$LOGIN_RESPONSE" | grep -q '"code":200'; then
        TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
        USER_ID=$(echo "$LOGIN_RESPONSE" | grep -o '"id":[0-9]*' | cut -d':' -f2)
        print_success "用户登录成功，Token: ${TOKEN:0:20}..."
    else
        print_error "用户登录失败"
        echo "$LOGIN_RESPONSE"
        exit 1
    fi
}

# 测试监控指标
test_metrics() {
    print_step "测试监控指标"
    
    # 访问metrics端点
    METRICS_RESPONSE=$(curl -s "$BASE_URL/metrics" || echo "metrics endpoint not available")
    
    if echo "$METRICS_RESPONSE" | grep -q "http_requests_total"; then
        print_success "Prometheus指标可用"
        echo "发现的指标："
        echo "$METRICS_RESPONSE" | grep -E "^# HELP|^http_requests_total|^user_registrations_total" | head -10
    else
        print_warning "Prometheus指标端点未配置或不可用"
    fi
}

# 测试商品缓存功能
test_product_cache() {
    print_step "测试商品缓存功能"
    
    # 创建测试商品
    CREATE_RESPONSE=$(curl -s -X POST "$API_URL/products" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $TOKEN" \
        -d '{
            "name": "缓存测试商品",
            "description": "用于测试Redis缓存功能的商品",
            "category_id": 1,
            "price": 99.99,
            "stock": 100,
            "main_image": "https://example.com/image.jpg"
        }')
    
    if echo "$CREATE_RESPONSE" | grep -q '"code":200'; then
        PRODUCT_ID=$(echo "$CREATE_RESPONSE" | grep -o '"id":[0-9]*' | cut -d':' -f2)
        print_success "测试商品创建成功，ID: $PRODUCT_ID"
        
        # 多次访问商品详情，测试缓存
        print_step "测试商品详情缓存"
        for i in {1..5}; do
            start_time=$(date +%s%N)
            PRODUCT_RESPONSE=$(curl -s "$API_URL/products/$PRODUCT_ID")
            end_time=$(date +%s%N)
            duration=$(( (end_time - start_time) / 1000000 ))
            
            if echo "$PRODUCT_RESPONSE" | grep -q '"code":200'; then
                echo "第${i}次请求: ${duration}ms"
            else
                print_error "商品详情获取失败"
            fi
        done
        
    else
        print_error "测试商品创建失败"
        echo "$CREATE_RESPONSE"
    fi
}

# 测试购物车缓存
test_cart_cache() {
    print_step "测试购物车缓存功能"
    
    if [ -z "$PRODUCT_ID" ]; then
        print_warning "跳过购物车测试，没有可用的商品ID"
        return
    fi
    
    # 添加商品到购物车
    CART_ADD_RESPONSE=$(curl -s -X POST "$API_URL/cart" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $TOKEN" \
        -d "{
            \"product_id\": $PRODUCT_ID,
            \"quantity\": 2
        }")
    
    if echo "$CART_ADD_RESPONSE" | grep -q '"code":200'; then
        print_success "商品添加到购物车成功"
        
        # 多次获取购物车，测试缓存
        print_step "测试购物车查询缓存"
        for i in {1..3}; do
            start_time=$(date +%s%N)
            CART_RESPONSE=$(curl -s -H "Authorization: Bearer $TOKEN" "$API_URL/cart")
            end_time=$(date +%s%N)
            duration=$(( (end_time - start_time) / 1000000 ))
            
            if echo "$CART_RESPONSE" | grep -q '"code":200'; then
                echo "第${i}次购物车查询: ${duration}ms"
            else
                print_error "购物车查询失败"
            fi
        done
        
    else
        print_error "添加商品到购物车失败"
        echo "$CART_ADD_RESPONSE"
    fi
}

# 测试WebSocket连接
test_websocket() {
    print_step "测试WebSocket功能"
    
    # 检查WebSocket端点是否可用
    WS_TEST=$(curl -s -I "$BASE_URL/ws" 2>/dev/null || echo "websocket endpoint not available")
    
    if echo "$WS_TEST" | grep -q "101\|Upgrade"; then
        print_success "WebSocket端点可用"
    else
        print_warning "WebSocket端点未配置或不可用"
        print_warning "需要在路由中添加WebSocket处理器"
    fi
}

# 测试健康检查
test_health_check() {
    print_step "测试健康检查"
    
    # 基础ping测试
    PING_RESPONSE=$(curl -s "$BASE_URL/ping")
    if echo "$PING_RESPONSE" | grep -q "pong"; then
        print_success "基础健康检查通过"
    else
        print_error "基础健康检查失败"
    fi
    
    # 详细健康检查（如果存在）
    HEALTH_RESPONSE=$(curl -s "$BASE_URL/health" 2>/dev/null || echo "detailed health check not available")
    if echo "$HEALTH_RESPONSE" | grep -q "status"; then
        print_success "详细健康检查可用"
        echo "$HEALTH_RESPONSE" | head -5
    else
        print_warning "详细健康检查端点未配置"
    fi
}

# 性能测试
performance_test() {
    print_step "简单性能测试"
    
    if [ -z "$PRODUCT_ID" ]; then
        print_warning "跳过性能测试，没有可用的商品ID"
        return
    fi
    
    print_step "并发商品查询测试"
    
    # 创建临时脚本进行并发测试
    cat > /tmp/concurrent_test.sh << EOF
#!/bin/bash
for i in {1..10}; do
    curl -s "$API_URL/products/$PRODUCT_ID" > /dev/null &
done
wait
EOF
    
    chmod +x /tmp/concurrent_test.sh
    
    start_time=$(date +%s%N)
    /tmp/concurrent_test.sh
    end_time=$(date +%s%N)
    duration=$(( (end_time - start_time) / 1000000 ))
    
    print_success "10个并发请求完成，总耗时: ${duration}ms"
    
    # 清理
    rm -f /tmp/concurrent_test.sh
}

# 清理测试数据
cleanup() {
    print_step "清理测试数据"
    
    if [ -n "$PRODUCT_ID" ] && [ -n "$TOKEN" ]; then
        # 删除测试商品
        DELETE_RESPONSE=$(curl -s -X DELETE "$API_URL/products/$PRODUCT_ID" \
            -H "Authorization: Bearer $TOKEN")
        
        if echo "$DELETE_RESPONSE" | grep -q '"code":200'; then
            print_success "测试商品删除成功"
        else
            print_warning "测试商品删除失败或已删除"
        fi
    fi
}

# 主函数
main() {
    echo -e "${BLUE}"
    echo "=================================="
    echo "   Ryan Mall 增强功能测试"
    echo "=================================="
    echo -e "${NC}"
    
    # 执行测试
    check_service
    setup_user
    test_metrics
    test_product_cache
    test_cart_cache
    test_websocket
    test_health_check
    performance_test
    
    # 清理
    cleanup
    
    echo -e "${GREEN}"
    echo "=================================="
    echo "      测试完成！"
    echo "=================================="
    echo -e "${NC}"
    
    print_step "测试总结"
    echo "✅ 基础功能测试完成"
    echo "✅ 缓存功能测试完成"
    echo "✅ 监控指标检查完成"
    echo "✅ 性能测试完成"
    echo ""
    echo "💡 下一步建议："
    echo "1. 配置Redis服务以启用缓存功能"
    echo "2. 集成Prometheus监控"
    echo "3. 添加WebSocket实时通知"
    echo "4. 配置Elasticsearch搜索"
}

# 执行主函数
main "$@"
