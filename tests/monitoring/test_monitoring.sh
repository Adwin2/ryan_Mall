#!/bin/bash

# 监控系统功能测试脚本
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
    echo "           监控系统功能测试"
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

# 检查服务状态
check_services() {
    print_step "检查监控服务状态"
    
    local services=(
        "prometheus:9090"
        "grafana:3001"
        "alertmanager:9093"
        "node-exporter:9100"
        "redis-exporter:9121"
        "mysql-exporter:9104"
        "cadvisor:8080"
        "blackbox-exporter:9115"
    )
    
    local failed_services=()
    
    for service_port in "${services[@]}"; do
        local service=$(echo $service_port | cut -d: -f1)
        local port=$(echo $service_port | cut -d: -f2)
        
        if curl -s --max-time 5 "http://localhost:$port" > /dev/null 2>&1; then
            print_success "$service (端口:$port) 运行正常"
        else
            print_error "$service (端口:$port) 无法访问"
            failed_services+=("$service")
        fi
    done
    
    if [ ${#failed_services[@]} -gt 0 ]; then
        print_error "以下服务无法访问: ${failed_services[*]}"
        return 1
    fi
    
    return 0
}

# 测试Prometheus指标采集
test_prometheus_metrics() {
    print_step "测试Prometheus指标采集"
    
    # 检查Prometheus健康状态
    if curl -s "http://localhost:9090/-/healthy" | grep -q "Prometheus is Healthy"; then
        print_success "Prometheus健康状态正常"
    else
        print_error "Prometheus健康状态异常"
    fi
    
    # 检查目标状态
    local targets_response=$(curl -s "http://localhost:9090/api/v1/targets")
    if echo "$targets_response" | grep -q '"status":"success"'; then
        print_success "Prometheus目标状态查询正常"
        
        # 统计活跃目标数量
        local active_targets=$(echo "$targets_response" | grep -o '"health":"up"' | wc -l)
        local total_targets=$(echo "$targets_response" | grep -o '"health":"' | wc -l)
        echo "  活跃目标: $active_targets/$total_targets"
    else
        print_error "Prometheus目标状态查询失败"
    fi
    
    # 测试基本查询
    local query_response=$(curl -s "http://localhost:9090/api/v1/query?query=up")
    if echo "$query_response" | grep -q '"status":"success"'; then
        print_success "Prometheus查询功能正常"
    else
        print_error "Prometheus查询功能异常"
    fi
}

# 测试Grafana功能
test_grafana() {
    print_step "测试Grafana功能"
    
    # 检查Grafana健康状态
    local health_response=$(curl -s "http://localhost:3001/api/health")
    if echo "$health_response" | grep -q '"database":"ok"'; then
        print_success "Grafana健康状态正常"
    else
        print_error "Grafana健康状态异常"
    fi
    
    # 检查数据源
    local datasources_response=$(curl -s -u admin:admin123 "http://localhost:3001/api/datasources")
    if echo "$datasources_response" | grep -q '"name":"Prometheus"'; then
        print_success "Prometheus数据源配置正常"
    else
        print_error "Prometheus数据源配置异常"
    fi
    
    # 检查仪表板
    local dashboards_response=$(curl -s -u admin:admin123 "http://localhost:3001/api/search")
    if echo "$dashboards_response" | grep -q '"title"'; then
        local dashboard_count=$(echo "$dashboards_response" | grep -o '"title"' | wc -l)
        print_success "仪表板加载正常 (共$dashboard_count个)"
    else
        print_warning "未找到仪表板或加载异常"
    fi
}

# 测试AlertManager
test_alertmanager() {
    print_step "测试AlertManager功能"
    
    # 检查AlertManager健康状态
    if curl -s "http://localhost:9093/-/healthy" | grep -q "OK"; then
        print_success "AlertManager健康状态正常"
    else
        print_error "AlertManager健康状态异常"
    fi
    
    # 检查告警规则
    local alerts_response=$(curl -s "http://localhost:9093/api/v1/alerts")
    if echo "$alerts_response" | grep -q '"status":"success"'; then
        print_success "AlertManager告警查询正常"
        
        # 统计告警数量
        local alert_count=$(echo "$alerts_response" | grep -o '"alertname"' | wc -l)
        echo "  当前告警数量: $alert_count"
    else
        print_error "AlertManager告警查询失败"
    fi
    
    # 检查配置
    local config_response=$(curl -s "http://localhost:9093/api/v1/status")
    if echo "$config_response" | grep -q '"configYAML"'; then
        print_success "AlertManager配置加载正常"
    else
        print_error "AlertManager配置加载异常"
    fi
}

# 测试指标采集器
test_exporters() {
    print_step "测试指标采集器"
    
    # 测试Node Exporter
    local node_metrics=$(curl -s "http://localhost:9100/metrics" | grep "node_" | wc -l)
    if [ $node_metrics -gt 0 ]; then
        print_success "Node Exporter指标采集正常 ($node_metrics个指标)"
    else
        print_error "Node Exporter指标采集异常"
    fi
    
    # 测试Redis Exporter
    local redis_metrics=$(curl -s "http://localhost:9121/metrics" | grep "redis_" | wc -l)
    if [ $redis_metrics -gt 0 ]; then
        print_success "Redis Exporter指标采集正常 ($redis_metrics个指标)"
    else
        print_error "Redis Exporter指标采集异常"
    fi
    
    # 测试MySQL Exporter
    local mysql_metrics=$(curl -s "http://localhost:9104/metrics" | grep "mysql_" | wc -l)
    if [ $mysql_metrics -gt 0 ]; then
        print_success "MySQL Exporter指标采集正常 ($mysql_metrics个指标)"
    else
        print_error "MySQL Exporter指标采集异常"
    fi
    
    # 测试cAdvisor
    local cadvisor_metrics=$(curl -s "http://localhost:8080/metrics" | grep "container_" | wc -l)
    if [ $cadvisor_metrics -gt 0 ]; then
        print_success "cAdvisor指标采集正常 ($cadvisor_metrics个指标)"
    else
        print_error "cAdvisor指标采集异常"
    fi
}

# 测试黑盒监控
test_blackbox_monitoring() {
    print_step "测试黑盒监控"
    
    # 测试HTTP探测
    local http_probe=$(curl -s "http://localhost:9115/probe?target=http://localhost:8080/ping&module=http_2xx")
    if echo "$http_probe" | grep -q "probe_success 1"; then
        print_success "HTTP探测功能正常"
    else
        print_error "HTTP探测功能异常"
    fi
    
    # 测试TCP探测
    local tcp_probe=$(curl -s "http://localhost:9115/probe?target=localhost:3306&module=tcp_connect")
    if echo "$tcp_probe" | grep -q "probe_success"; then
        print_success "TCP探测功能正常"
    else
        print_error "TCP探测功能异常"
    fi
}

# 生成负载进行测试
generate_test_load() {
    print_step "生成测试负载"
    
    print_warning "生成HTTP请求负载..."
    
    # 生成一些HTTP请求
    for i in {1..20}; do
        curl -s "http://localhost:8080/ping" > /dev/null &
        curl -s "http://localhost:8080/health" > /dev/null &
        curl -s "http://localhost:8080/api/v1/products" > /dev/null &
    done
    
    wait
    print_success "测试负载生成完成"
    
    # 等待指标更新
    print_warning "等待指标更新..."
    sleep 10
}

# 验证监控数据
verify_monitoring_data() {
    print_step "验证监控数据"
    
    # 检查HTTP请求指标
    local http_requests=$(curl -s "http://localhost:9090/api/v1/query?query=http_requests_total" | grep -o '"value":\[[^]]*\]' | wc -l)
    if [ $http_requests -gt 0 ]; then
        print_success "HTTP请求指标采集正常"
    else
        print_warning "HTTP请求指标未检测到（可能需要更多时间）"
    fi
    
    # 检查系统指标
    local cpu_usage=$(curl -s "http://localhost:9090/api/v1/query?query=node_cpu_seconds_total")
    if echo "$cpu_usage" | grep -q '"status":"success"'; then
        print_success "系统CPU指标采集正常"
    else
        print_error "系统CPU指标采集异常"
    fi
    
    # 检查Redis指标
    local redis_info=$(curl -s "http://localhost:9090/api/v1/query?query=redis_up")
    if echo "$redis_info" | grep -q '"status":"success"'; then
        print_success "Redis监控指标采集正常"
    else
        print_error "Redis监控指标采集异常"
    fi
}

# 显示监控概览
show_monitoring_overview() {
    print_step "监控系统概览"
    
    echo "📊 监控服务状态："
    docker compose -f docker-compose.monitoring.yml ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}"
    
    echo ""
    echo "📈 关键指标："
    
    # 获取一些关键指标
    local up_targets=$(curl -s "http://localhost:9090/api/v1/query?query=up" | grep -o '"value":\["[^"]*","1"\]' | wc -l)
    echo "  在线目标数: $up_targets"
    
    local total_metrics=$(curl -s "http://localhost:9090/api/v1/label/__name__/values" | grep -o '\"[^\"]*\"' | wc -l)
    echo "  总指标数: $total_metrics"
    
    echo ""
    echo "🔗 访问地址："
    echo "  Grafana: http://localhost:3000 (admin/admin123)"
    echo "  Prometheus: http://localhost:9090"
    echo "  AlertManager: http://localhost:9093"
}

# 主函数
main() {
    print_header
    
    # 检查服务状态
    if ! check_services; then
        print_error "监控服务状态检查失败，请先启动监控系统"
        echo "运行: ./start_monitoring.sh"
        exit 1
    fi
    echo ""
    
    # 测试各个组件
    test_prometheus_metrics
    echo ""
    
    test_grafana
    echo ""
    
    test_alertmanager
    echo ""
    
    test_exporters
    echo ""
    
    test_blackbox_monitoring
    echo ""
    
    # 生成测试负载
    generate_test_load
    echo ""
    
    # 验证监控数据
    verify_monitoring_data
    echo ""
    
    # 显示概览
    show_monitoring_overview
    
    print_success "监控系统功能测试完成！"
    echo ""
    print_warning "建议访问 Grafana 仪表板查看详细监控数据"
}

# 执行主函数
main "$@"
