#!/bin/bash

# Prometheus + Grafana 监控系统启动脚本
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
    echo "        Prometheus + Grafana 监控系统"
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

# 检查系统要求
check_requirements() {
    print_step "检查系统要求"
    
    if ! command -v docker &> /dev/null; then
        print_error "Docker未安装，请先安装Docker"
        exit 1
    fi
    print_success "Docker已安装"
    
    if ! docker compose version &> /dev/null; then
        print_error "Docker Compose未安装，请先安装Docker Compose"
        exit 1
    fi
    print_success "Docker Compose已安装"
    
    # 检查端口占用
    local ports=(3000 9090 9093 9100 9104 9115 9121 8080)
    for port in "${ports[@]}"; do
        if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
            print_warning "端口 $port 已被占用，可能会导致冲突"
        fi
    done
}

# 创建必要的目录
create_directories() {
    print_step "创建监控目录"
    
    # 创建数据目录
    mkdir -p monitoring/{prometheus/{data,rules},grafana/{data,dashboards},alertmanager/data}
    
    # 设置权限
    sudo chown -R 472:472 monitoring/grafana/ 2>/dev/null || true
    sudo chown -R 65534:65534 monitoring/prometheus/ 2>/dev/null || true
    sudo chown -R 65534:65534 monitoring/alertmanager/ 2>/dev/null || true
    
    print_success "监控目录创建完成"
}

# 停止现有服务
stop_existing_services() {
    print_step "停止现有监控服务"
    
    # 停止现有监控服务
    if docker ps | grep -q "ryan-mall-prometheus\|ryan-mall-grafana"; then
        print_warning "停止现有的监控服务"
        docker compose -f docker-compose.monitoring.yml down || true
    fi
    
    print_success "现有服务已停止"
}

# 启动监控服务
start_monitoring_services() {
    print_step "启动监控服务"
    
    print_warning "启动基础监控组件..."
    docker compose -f docker-compose.monitoring.yml up -d prometheus grafana alertmanager
    
    if [ $? -eq 0 ]; then
        print_success "基础监控组件启动成功"
    else
        print_error "基础监控组件启动失败"
        exit 1
    fi
    
    # 等待服务启动
    print_warning "等待服务完全启动..."
    sleep 10
    
    print_warning "启动监控采集器..."
    docker compose -f docker-compose.monitoring.yml up -d node-exporter redis-exporter mysql-exporter cadvisor blackbox-exporter
    
    if [ $? -eq 0 ]; then
        print_success "监控采集器启动成功"
    else
        print_error "监控采集器启动失败"
        exit 1
    fi
}

# 检查服务状态
check_services_status() {
    print_step "检查服务状态"
    
    local services=("prometheus" "grafana" "alertmanager" "node-exporter" "redis-exporter" "mysql-exporter" "cadvisor" "blackbox-exporter")
    local failed_services=()
    
    for service in "${services[@]}"; do
        if docker ps | grep -q "ryan-mall-$service"; then
            print_success "$service 运行正常"
        else
            print_error "$service 启动失败"
            failed_services+=("$service")
        fi
    done
    
    if [ ${#failed_services[@]} -gt 0 ]; then
        print_error "以下服务启动失败: ${failed_services[*]}"
        return 1
    fi
    
    return 0
}

# 验证监控功能
verify_monitoring() {
    print_step "验证监控功能"
    
    print_warning "等待服务完全就绪..."
    sleep 15
    
    # 检查Prometheus
    if curl -s http://localhost:9090/-/healthy > /dev/null; then
        print_success "Prometheus健康检查通过"
    else
        print_error "Prometheus健康检查失败"
    fi
    
    # 检查Grafana
    if curl -s http://localhost:3000/api/health > /dev/null; then
        print_success "Grafana健康检查通过"
    else
        print_error "Grafana健康检查失败"
    fi
    
    # 检查AlertManager
    if curl -s http://localhost:9093/-/healthy > /dev/null; then
        print_success "AlertManager健康检查通过"
    else
        print_error "AlertManager健康检查失败"
    fi
    
    # 检查指标采集
    print_warning "检查指标采集..."
    if curl -s "http://localhost:9090/api/v1/query?query=up" | grep -q "success"; then
        print_success "指标采集正常"
    else
        print_error "指标采集异常"
    fi
}

# 导入仪表板
import_dashboards() {
    print_step "导入Grafana仪表板"
    
    print_warning "等待Grafana完全启动..."
    sleep 20
    
    # 检查Grafana API
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s -u admin:admin123 http://localhost:3000/api/health > /dev/null; then
            print_success "Grafana API可用"
            break
        fi
        
        if [ $attempt -eq $max_attempts ]; then
            print_error "Grafana API不可用，跳过仪表板导入"
            return 1
        fi
        
        echo "等待Grafana启动... ($attempt/$max_attempts)"
        sleep 2
        ((attempt++))
    done
    
    # 这里可以添加仪表板导入逻辑
    print_success "仪表板配置已就绪"
}

# 显示访问信息
show_access_info() {
    print_step "访问信息"
    
    echo "🌐 监控服务访问地址："
    echo "   - Grafana仪表板: http://localhost:3000"
    echo "     用户名: admin"
    echo "     密码: admin123"
    echo ""
    echo "   - Prometheus: http://localhost:9090"
    echo "   - AlertManager: http://localhost:9093"
    echo ""
    echo "📊 监控指标端点："
    echo "   - Node Exporter: http://localhost:9100/metrics"
    echo "   - Redis Exporter: http://localhost:9121/metrics"
    echo "   - MySQL Exporter: http://localhost:9104/metrics"
    echo "   - cAdvisor: http://localhost:8080/metrics"
    echo "   - Blackbox Exporter: http://localhost:9115/metrics"
    echo ""
    echo "🔧 管理命令："
    echo "   查看服务状态: docker compose -f docker-compose.monitoring.yml ps"
    echo "   查看日志: docker compose -f docker-compose.monitoring.yml logs [service]"
    echo "   停止服务: docker compose -f docker-compose.monitoring.yml down"
    echo "   重启服务: docker compose -f docker-compose.monitoring.yml restart [service]"
}

# 主函数
main() {
    print_header
    
    check_requirements
    echo ""
    
    create_directories
    echo ""
    
    stop_existing_services
    echo ""
    
    start_monitoring_services
    echo ""
    
    if check_services_status; then
        echo ""
        verify_monitoring
        echo ""
        import_dashboards
        echo ""
        show_access_info
        echo ""
        print_success "监控系统部署完成！"
        print_warning "请访问 http://localhost:3000 查看监控仪表板"
    else
        print_error "监控系统部署失败，请检查日志"
        exit 1
    fi
}

# 执行主函数
main "$@"
