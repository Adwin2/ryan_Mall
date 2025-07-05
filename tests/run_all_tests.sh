#!/bin/bash

# Ryan Mall 一键测试脚本
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
    echo "           Ryan Mall 一键测试套件"
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

# 检查脚本权限
check_permissions() {
    print_step "检查脚本执行权限"
    
    find . -name "*.sh" -type f ! -executable -exec chmod +x {} \;
    print_success "脚本权限检查完成"
}

# 显示测试菜单
show_menu() {
    echo ""
    echo "🧪 请选择要运行的测试类型："
    echo ""
    echo "1) 🔧 API功能测试"
    echo "2) ⚡ 性能压力测试"
    echo "3) 📊 监控系统测试"
    echo "4) 🔴 Redis集群测试"
    echo "5) 🚀 部署启动测试"
    echo "6) ⚙️  系统优化测试"
    echo "7) 🎯 完整测试套件"
    echo "8) 📋 查看测试状态"
    echo "9) 🔍 故障排查工具"
    echo "0) 退出"
    echo ""
    read -p "请输入选择 (0-9): " choice
}

# API功能测试
run_api_tests() {
    print_step "运行API功能测试"
    
    cd api
    
    echo "1. 基础API测试..."
    if ./test_api.sh; then
        print_success "基础API测试通过"
    else
        print_error "基础API测试失败"
    fi
    
    echo ""
    echo "2. 商品API测试..."
    if ./test_product_api.sh; then
        print_success "商品API测试通过"
    else
        print_error "商品API测试失败"
    fi
    
    echo ""
    echo "3. 购物车API测试..."
    if ./test_cart_api.sh; then
        print_success "购物车API测试通过"
    else
        print_error "购物车API测试失败"
    fi
    
    echo ""
    echo "4. 订单API测试..."
    if ./test_order_api.sh; then
        print_success "订单API测试通过"
    else
        print_error "订单API测试失败"
    fi
    
    echo ""
    echo "5. 增强功能测试..."
    if ./test_enhanced_features.sh; then
        print_success "增强功能测试通过"
    else
        print_error "增强功能测试失败"
    fi
    
    cd ..
}

# 性能压力测试
run_performance_tests() {
    print_step "运行性能压力测试"
    
    cd performance
    
    echo "1. 基础性能测试..."
    if ./test_performance.sh; then
        print_success "基础性能测试通过"
    else
        print_error "基础性能测试失败"
    fi
    
    echo ""
    echo "2. 并发性能测试..."
    if ./test_concurrent_performance.sh; then
        print_success "并发性能测试通过"
    else
        print_error "并发性能测试失败"
    fi
    
    echo ""
    echo "3. 缓存性能测试..."
    if ./test_cache_performance.sh; then
        print_success "缓存性能测试通过"
    else
        print_error "缓存性能测试失败"
    fi
    
    echo ""
    read -p "是否运行压力测试？(可能消耗大量资源) (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "4. 增强压力测试..."
        if ./enhanced_stress_test.sh; then
            print_success "增强压力测试通过"
        else
            print_error "增强压力测试失败"
        fi
    fi
    
    cd ..
}

# 监控系统测试
run_monitoring_tests() {
    print_step "运行监控系统测试"
    
    cd monitoring
    
    if ./test_monitoring.sh; then
        print_success "监控系统测试通过"
    else
        print_error "监控系统测试失败"
    fi
    
    cd ..
}

# Redis集群测试
run_redis_tests() {
    print_step "运行Redis集群测试"
    
    cd redis
    
    echo "1. 简单集群测试..."
    if ./simple_redis_cluster_test.sh; then
        print_success "Redis集群测试通过"
    else
        print_error "Redis集群测试失败"
    fi
    
    echo ""
    echo "2. 性能对比测试..."
    if ./redis_vs_memory_performance.sh; then
        print_success "Redis性能对比测试通过"
    else
        print_error "Redis性能对比测试失败"
    fi
    
    cd ..
}

# 部署启动测试
run_deployment_tests() {
    print_step "运行部署启动测试"
    
    cd deployment
    
    echo "测试服务启动脚本..."
    
    echo "1. 检查优化版启动脚本..."
    if [ -f "./start_optimized.sh" ]; then
        print_success "优化版启动脚本存在"
    else
        print_error "优化版启动脚本不存在"
    fi
    
    echo "2. 检查监控启动脚本..."
    if [ -f "./start_monitoring.sh" ]; then
        print_success "监控启动脚本存在"
    else
        print_error "监控启动脚本不存在"
    fi
    
    echo "3. 检查Redis集群启动脚本..."
    if [ -f "./start_redis_cluster.sh" ]; then
        print_success "Redis集群启动脚本存在"
    else
        print_error "Redis集群启动脚本不存在"
    fi
    
    cd ..
}

# 系统优化测试
run_optimization_tests() {
    print_step "运行系统优化测试"
    
    cd optimization
    
    echo "1. 检查系统网络优化脚本..."
    if [ -f "./system_network_optimization.sh" ]; then
        print_success "系统网络优化脚本存在"
    else
        print_error "系统网络优化脚本不存在"
    fi
    
    echo "2. 检查用户级优化脚本..."
    if [ -f "./user_level_optimization.sh" ]; then
        print_success "用户级优化脚本存在"
    else
        print_error "用户级优化脚本不存在"
    fi
    
    echo "3. 检查Go运行时优化脚本..."
    if [ -f "./go_runtime_env.sh" ]; then
        print_success "Go运行时优化脚本存在"
    else
        print_error "Go运行时优化脚本不存在"
    fi
    
    cd ..
}

# 完整测试套件
run_full_tests() {
    print_step "运行完整测试套件"
    
    print_warning "这将运行所有测试，可能需要较长时间..."
    read -p "确认继续？(y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        return
    fi
    
    run_api_tests
    echo ""
    run_performance_tests
    echo ""
    run_monitoring_tests
    echo ""
    run_redis_tests
    echo ""
    run_deployment_tests
    echo ""
    run_optimization_tests
    
    print_success "完整测试套件执行完成"
}

# 查看测试状态
show_test_status() {
    print_step "测试环境状态检查"
    
    echo "📊 服务状态："
    if curl -s http://localhost:8080/ping > /dev/null; then
        print_success "Ryan Mall服务运行正常 (端口:8080)"
    else
        print_warning "Ryan Mall服务未运行 (端口:8080)"
    fi
    
    if curl -s http://localhost:9090/-/healthy > /dev/null; then
        print_success "Prometheus运行正常 (端口:9090)"
    else
        print_warning "Prometheus未运行 (端口:9090)"
    fi
    
    if curl -s http://localhost:3001/api/health > /dev/null; then
        print_success "Grafana运行正常 (端口:3001)"
    else
        print_warning "Grafana未运行 (端口:3001)"
    fi
    
    echo ""
    echo "🔴 Redis集群状态："
    if docker exec ryan-mall-redis-node-1 redis-cli cluster info 2>/dev/null | grep -q "cluster_state:ok"; then
        print_success "Redis集群运行正常"
    else
        print_warning "Redis集群未运行或状态异常"
    fi
    
    echo ""
    echo "🐳 Docker容器状态："
    docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep ryan-mall || print_warning "没有运行的Ryan Mall容器"
}

# 故障排查工具
troubleshooting_tools() {
    print_step "故障排查工具"
    
    echo "🔍 选择排查工具："
    echo "1) 检查端口占用"
    echo "2) 查看Docker日志"
    echo "3) 检查网络连接"
    echo "4) 查看系统资源"
    echo "5) 返回主菜单"
    
    read -p "请选择 (1-5): " tool_choice
    
    case $tool_choice in
        1)
            echo "检查常用端口占用："
            for port in 8080 9090 3001 9093 3306 7001 7002 7003; do
                if lsof -i :$port > /dev/null 2>&1; then
                    echo "端口 $port: 已占用"
                    lsof -i :$port
                else
                    echo "端口 $port: 空闲"
                fi
            done
            ;;
        2)
            echo "Docker容器日志："
            docker ps --format "{{.Names}}" | grep ryan-mall | head -5 | while read container; do
                echo "=== $container 日志 ==="
                docker logs --tail 10 $container
                echo ""
            done
            ;;
        3)
            echo "网络连接测试："
            curl -s http://localhost:8080/ping && echo "✅ Ryan Mall API可达" || echo "❌ Ryan Mall API不可达"
            curl -s http://localhost:9090/-/healthy && echo "✅ Prometheus可达" || echo "❌ Prometheus不可达"
            ;;
        4)
            echo "系统资源使用："
            echo "CPU和内存："
            top -bn1 | head -5
            echo ""
            echo "磁盘使用："
            df -h | head -5
            ;;
        5)
            return
            ;;
    esac
}

# 主函数
main() {
    print_header
    check_permissions
    
    while true; do
        show_menu
        
        case $choice in
            1) run_api_tests ;;
            2) run_performance_tests ;;
            3) run_monitoring_tests ;;
            4) run_redis_tests ;;
            5) run_deployment_tests ;;
            6) run_optimization_tests ;;
            7) run_full_tests ;;
            8) show_test_status ;;
            9) troubleshooting_tools ;;
            0) 
                print_success "测试完成，再见！"
                exit 0
                ;;
            *)
                print_error "无效选择，请重新输入"
                ;;
        esac
        
        echo ""
        read -p "按回车键继续..." -r
    done
}

# 执行主函数
main "$@"
