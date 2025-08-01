#!/bin/bash

# Ryan Mall 微服务集成测试脚本

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

# 测试计数器
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 测试结果记录
test_result() {
    local test_name=$1
    local result=$2
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    if [ "$result" = "PASS" ]; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
        log_success "✓ $test_name"
    else
        FAILED_TESTS=$((FAILED_TESTS + 1))
        log_error "✗ $test_name"
    fi
}

# 检查服务是否运行
check_service() {
    local service_name=$1
    local port=$2
    
    if curl -s -f "http://localhost:$port/health" > /dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

# 等待服务启动
wait_for_service() {
    local service_name=$1
    local port=$2
    local max_wait=30
    local count=0
    
    log_info "等待 $service_name 启动..."
    
    while [ $count -lt $max_wait ]; do
        if check_service "$service_name" $port; then
            return 0
        fi
        sleep 1
        count=$((count + 1))
    done
    
    return 1
}

# 测试用户注册
test_user_registration() {
    log_info "测试用户注册..."
    
    local response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d '{
            "username": "integration_test_user",
            "email": "integration@test.com",
            "password": "Test123!@#"
        }' \
        http://localhost:8081/api/v1/users/register)
    
    if echo "$response" | grep -q "user_id"; then
        test_result "用户注册" "PASS"
        # 提取用户ID供后续测试使用
        USER_ID=$(echo "$response" | grep -o '"user_id":"[^"]*"' | cut -d'"' -f4)
        export USER_ID
        return 0
    else
        test_result "用户注册" "FAIL"
        log_error "注册响应: $response"
        return 1
    fi
}

# 测试用户登录
test_user_login() {
    log_info "测试用户登录..."
    
    local response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d '{
            "username": "integration_test_user",
            "password": "Test123!@#"
        }' \
        http://localhost:8081/api/v1/users/login)
    
    if echo "$response" | grep -q "access_token"; then
        test_result "用户登录" "PASS"
        # 提取访问令牌供后续测试使用
        ACCESS_TOKEN=$(echo "$response" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
        export ACCESS_TOKEN
        return 0
    else
        test_result "用户登录" "FAIL"
        log_error "登录响应: $response"
        return 1
    fi
}

# 测试获取用户信息
test_get_user() {
    log_info "测试获取用户信息..."
    
    if [ -z "$USER_ID" ] || [ -z "$ACCESS_TOKEN" ]; then
        test_result "获取用户信息" "FAIL"
        log_error "缺少用户ID或访问令牌"
        return 1
    fi
    
    local response=$(curl -s -X GET \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        http://localhost:8081/api/v1/users/$USER_ID)
    
    if echo "$response" | grep -q "integration_test_user"; then
        test_result "获取用户信息" "PASS"
        return 0
    else
        test_result "获取用户信息" "FAIL"
        log_error "获取用户信息响应: $response"
        return 1
    fi
}

# 测试用户列表
test_user_list() {
    log_info "测试用户列表..."
    
    local response=$(curl -s -X GET \
        "http://localhost:8081/api/v1/users?page=1&page_size=10")
    
    if echo "$response" | grep -q "users"; then
        test_result "用户列表" "PASS"
        return 0
    else
        test_result "用户列表" "FAIL"
        log_error "用户列表响应: $response"
        return 1
    fi
}

# 测试网关健康检查
test_gateway_health() {
    log_info "测试网关健康检查..."
    
    local response=$(curl -s http://localhost:8080/health)
    
    if echo "$response" | grep -q "healthy"; then
        test_result "网关健康检查" "PASS"
        return 0
    else
        test_result "网关健康检查" "FAIL"
        log_error "网关健康检查响应: $response"
        return 1
    fi
}

# 测试通过网关访问用户服务
test_gateway_proxy() {
    log_info "测试通过网关访问用户服务..."
    
    local response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d '{
            "username": "gateway_test_user",
            "email": "gateway@test.com",
            "password": "Test123!@#"
        }' \
        http://localhost:8080/api/v1/users/register)
    
    if echo "$response" | grep -q "user_id"; then
        test_result "网关代理" "PASS"
        return 0
    else
        test_result "网关代理" "FAIL"
        log_error "网关代理响应: $response"
        return 1
    fi
}

# 测试服务监控指标
test_metrics() {
    log_info "测试服务监控指标..."
    
    # 测试用户服务指标
    local user_metrics=$(curl -s http://localhost:8081/metrics)
    if echo "$user_metrics" | grep -q "go_"; then
        test_result "用户服务指标" "PASS"
    else
        test_result "用户服务指标" "FAIL"
    fi
    
    # 测试网关服务指标
    local gateway_metrics=$(curl -s http://localhost:8080/metrics)
    if echo "$gateway_metrics" | grep -q "go_"; then
        test_result "网关服务指标" "PASS"
    else
        test_result "网关服务指标" "FAIL"
    fi
}

# 测试错误处理
test_error_handling() {
    log_info "测试错误处理..."
    
    # 测试无效的用户注册
    local response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d '{
            "username": "",
            "email": "invalid-email",
            "password": "123"
        }' \
        http://localhost:8081/api/v1/users/register)
    
    if echo "$response" | grep -q "error"; then
        test_result "错误处理" "PASS"
    else
        test_result "错误处理" "FAIL"
        log_error "错误处理响应: $response"
    fi
}

# 性能测试
test_performance() {
    log_info "测试基本性能..."
    
    # 测试用户服务响应时间
    local start_time=$(date +%s%N)
    curl -s http://localhost:8081/health > /dev/null
    local end_time=$(date +%s%N)
    local duration=$(( (end_time - start_time) / 1000000 )) # 转换为毫秒
    
    if [ $duration -lt 1000 ]; then # 小于1秒
        test_result "用户服务响应时间 (${duration}ms)" "PASS"
    else
        test_result "用户服务响应时间 (${duration}ms)" "FAIL"
    fi
    
    # 测试网关服务响应时间
    start_time=$(date +%s%N)
    curl -s http://localhost:8080/health > /dev/null
    end_time=$(date +%s%N)
    duration=$(( (end_time - start_time) / 1000000 ))
    
    if [ $duration -lt 1000 ]; then
        test_result "网关服务响应时间 (${duration}ms)" "PASS"
    else
        test_result "网关服务响应时间 (${duration}ms)" "FAIL"
    fi
}

# 清理测试数据
cleanup_test_data() {
    log_info "清理测试数据..."
    
    # 这里可以添加清理逻辑，比如删除测试用户
    # 由于我们使用的是内存数据库或测试数据库，重启服务即可清理
    
    log_info "测试数据清理完成"
}

# 显示测试结果摘要
show_test_summary() {
    echo ""
    echo "========================================"
    echo "           集成测试结果摘要"
    echo "========================================"
    echo -e "总测试数: ${BLUE}$TOTAL_TESTS${NC}"
    echo -e "通过测试: ${GREEN}$PASSED_TESTS${NC}"
    echo -e "失败测试: ${RED}$FAILED_TESTS${NC}"
    
    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "测试状态: ${GREEN}全部通过 ✓${NC}"
        echo "========================================"
        return 0
    else
        echo -e "测试状态: ${RED}有失败测试 ✗${NC}"
        echo "========================================"
        return 1
    fi
}

# 主测试流程
main() {
    log_info "开始Ryan Mall微服务集成测试"
    echo ""
    
    # 检查服务是否运行
    log_info "检查服务状态..."
    
    if ! wait_for_service "用户服务" 8081; then
        log_error "用户服务未运行，请先启动服务"
        exit 1
    fi
    
    if ! wait_for_service "网关服务" 8080; then
        log_error "网关服务未运行，请先启动服务"
        exit 1
    fi
    
    log_success "所有服务已就绪"
    echo ""
    
    # 执行测试
    log_info "执行集成测试..."
    echo ""
    
    # 基础功能测试
    test_user_registration
    test_user_login
    test_get_user
    test_user_list
    
    # 网关测试
    test_gateway_health
    test_gateway_proxy
    
    # 监控测试
    test_metrics
    
    # 错误处理测试
    test_error_handling
    
    # 性能测试
    test_performance
    
    # 清理测试数据
    cleanup_test_data
    
    # 显示测试结果
    echo ""
    show_test_summary
}

# 执行主函数
main "$@"
