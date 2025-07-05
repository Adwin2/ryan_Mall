#!/bin/bash

# Ryan Mall 快速启动脚本
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
    echo "           Ryan Mall 快速启动菜单"
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

# 显示主菜单
show_main_menu() {
    echo ""
    echo "🚀 请选择操作："
    echo ""
    echo "=== 🔧 服务管理 ==="
    echo "1) 启动Ryan Mall应用 (优化版)"
    echo "2) 启动Redis集群"
    echo "3) 启动监控系统"
    echo "4) 停止所有服务"
    echo ""
    echo "=== 🧪 测试工具 ==="
    echo "5) 运行API测试"
    echo "6) 运行性能测试"
    echo "7) 运行完整测试套件"
    echo ""
    echo "=== 📊 监控查看 ==="
    echo "8) 查看服务状态"
    echo "9) 打开监控面板"
    echo ""
    echo "=== 📚 文档帮助 ==="
    echo "10) 查看项目文档"
    echo "11) 查看测试指南"
    echo ""
    echo "0) 退出"
    echo ""
    read -p "请输入选择 (0-11): " choice
}

# 启动Ryan Mall应用
start_ryan_mall() {
    print_step "启动Ryan Mall应用"
    
    if [ -f "tests/deployment/start_optimized.sh" ]; then
        cd tests/deployment
        ./start_optimized.sh
        cd ../..
        print_success "Ryan Mall应用启动完成"
    else
        print_error "启动脚本不存在"
    fi
}

# 启动Redis集群
start_redis_cluster() {
    print_step "启动Redis集群"
    
    if [ -f "tests/deployment/start_redis_cluster.sh" ]; then
        cd tests/deployment
        ./start_redis_cluster.sh
        cd ../..
        print_success "Redis集群启动完成"
    else
        print_error "Redis集群启动脚本不存在"
    fi
}

# 启动监控系统
start_monitoring() {
    print_step "启动监控系统"
    
    if [ -f "tests/deployment/start_monitoring.sh" ]; then
        cd tests/deployment
        ./start_monitoring.sh
        cd ../..
        print_success "监控系统启动完成"
    else
        print_error "监控启动脚本不存在"
    fi
}

# 停止所有服务
stop_all_services() {
    print_step "停止所有服务"
    
    print_warning "停止Docker Compose服务..."
    
    # 停止主应用
    if [ -f "docker-compose.yml" ]; then
        docker compose -f docker-compose.yml down || true
    fi
    
    # 停止Redis集群
    if [ -f "docker-compose.redis-cluster.yml" ]; then
        docker compose -f docker-compose.redis-cluster.yml down || true
    fi
    
    # 停止监控系统
    if [ -f "docker-compose.monitoring.yml" ]; then
        docker compose -f docker-compose.monitoring.yml down || true
    fi
    
    # 停止Go应用进程
    pkill -f "go run cmd/server/main.go" || true
    
    print_success "所有服务已停止"
}

# 运行API测试
run_api_tests() {
    print_step "运行API测试"
    
    if [ -f "tests/api/test_api.sh" ]; then
        cd tests/api
        ./test_api.sh
        cd ../..
        print_success "API测试完成"
    else
        print_error "API测试脚本不存在"
    fi
}

# 运行性能测试
run_performance_tests() {
    print_step "运行性能测试"
    
    if [ -f "tests/performance/test_performance.sh" ]; then
        cd tests/performance
        ./test_performance.sh
        cd ../..
        print_success "性能测试完成"
    else
        print_error "性能测试脚本不存在"
    fi
}

# 运行完整测试套件
run_full_tests() {
    print_step "运行完整测试套件"
    
    if [ -f "tests/run_all_tests.sh" ]; then
        cd tests
        ./run_all_tests.sh
        cd ..
        print_success "完整测试套件完成"
    else
        print_error "测试套件脚本不存在"
    fi
}

# 查看服务状态
show_service_status() {
    print_step "服务状态检查"
    
    echo "🌐 Web服务状态："
    if curl -s http://localhost:8080/ping > /dev/null 2>&1; then
        print_success "Ryan Mall API (端口:8080) - 运行正常"
    else
        print_warning "Ryan Mall API (端口:8080) - 未运行"
    fi
    
    if curl -s http://localhost:9090/-/healthy > /dev/null 2>&1; then
        print_success "Prometheus (端口:9090) - 运行正常"
    else
        print_warning "Prometheus (端口:9090) - 未运行"
    fi
    
    if curl -s http://localhost:3001/api/health > /dev/null 2>&1; then
        print_success "Grafana (端口:3001) - 运行正常"
    else
        print_warning "Grafana (端口:3001) - 未运行"
    fi
    
    echo ""
    echo "🔴 Redis集群状态："
    if docker exec ryan-mall-redis-node-1 redis-cli cluster info 2>/dev/null | grep -q "cluster_state:ok"; then
        print_success "Redis集群 - 运行正常"
    else
        print_warning "Redis集群 - 未运行或状态异常"
    fi
    
    echo ""
    echo "🐳 Docker容器状态："
    docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep ryan-mall || print_warning "没有运行的Ryan Mall容器"
}

# 打开监控面板
open_monitoring_panels() {
    print_step "打开监控面板"
    
    echo "🌐 监控面板地址："
    echo "  - Prometheus: http://localhost:9090"
    echo "  - Grafana: http://localhost:3001 (admin/admin123)"
    echo "  - AlertManager: http://localhost:9093"
    echo ""
    echo "🔗 应用地址："
    echo "  - Ryan Mall API: http://localhost:8080"
    echo "  - API文档: http://localhost:8080/swagger/index.html"
    echo ""
    
    read -p "是否在浏览器中打开Grafana？(y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        if command -v xdg-open > /dev/null; then
            xdg-open http://localhost:3001
        elif command -v open > /dev/null; then
            open http://localhost:3001
        else
            print_warning "无法自动打开浏览器，请手动访问 http://localhost:3001"
        fi
    fi
}

# 查看项目文档
show_project_docs() {
    print_step "项目文档"
    
    echo "📚 可用文档："
    echo ""
    echo "=== 📋 主要文档 ==="
    echo "  - README.md - 项目概述"
    echo "  - Plan.md - 项目规划"
    echo "  - SETUP.md - 环境搭建"
    echo ""
    echo "=== 🏗️ 架构文档 ==="
    echo "  - docs/ARCHITECTURE.md - 系统架构"
    echo "  - docs/DEPLOYMENT.md - 部署指南"
    echo "  - docs/PROJECT_SUMMARY.md - 项目总结"
    echo ""
    echo "=== ⚡ 性能优化 ==="
    echo "  - PERFORMANCE_OPTIMIZATION_REPORT.md - 性能优化报告"
    echo "  - performance_optimization_guide.md - 优化指南"
    echo ""
    echo "=== 🔴 Redis集群 ==="
    echo "  - REDIS_CLUSTER_DEPLOYMENT_GUIDE.md - 部署指南"
    echo "  - REDIS_CLUSTER_APPLICATION_GUIDE.md - 应用指南"
    echo ""
    echo "=== 📊 监控系统 ==="
    echo "  - MONITORING_DEPLOYMENT_GUIDE.md - 监控部署指南"
    echo ""
    
    read -p "输入文档名称查看内容 (或按回车返回): " doc_name
    if [ -n "$doc_name" ] && [ -f "$doc_name" ]; then
        echo ""
        echo "=== $doc_name 内容 ==="
        head -50 "$doc_name"
        echo ""
        echo "... (显示前50行，完整内容请直接查看文件)"
    fi
}

# 查看测试指南
show_test_guide() {
    print_step "测试指南"
    
    if [ -f "tests/README.md" ]; then
        echo "📖 测试指南内容："
        echo ""
        head -30 tests/README.md
        echo ""
        echo "... (完整内容请查看 tests/README.md)"
        echo ""
        echo "🧪 快速测试命令："
        echo "  cd tests && ./run_all_tests.sh"
    else
        print_error "测试指南文件不存在"
    fi
}

# 主函数
main() {
    print_header
    
    while true; do
        show_main_menu
        
        case $choice in
            1) start_ryan_mall ;;
            2) start_redis_cluster ;;
            3) start_monitoring ;;
            4) stop_all_services ;;
            5) run_api_tests ;;
            6) run_performance_tests ;;
            7) run_full_tests ;;
            8) show_service_status ;;
            9) open_monitoring_panels ;;
            10) show_project_docs ;;
            11) show_test_guide ;;
            0) 
                print_success "再见！"
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
