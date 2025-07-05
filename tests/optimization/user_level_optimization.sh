#!/bin/bash

# 用户级性能优化脚本
# 不需要root权限的优化措施

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
    echo "           用户级性能优化"
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

# 显示当前系统状态
show_current_status() {
    print_step "当前系统状态"
    
    echo "CPU信息:"
    echo "  核心数: $(nproc)"
    echo "  负载: $(uptime | awk -F'load average:' '{print $2}')"
    
    echo ""
    echo "内存信息:"
    free -h | head -2
    
    echo ""
    echo "当前用户限制:"
    echo "  文件描述符: $(ulimit -n)"
    echo "  进程数: $(ulimit -u)"
    echo "  内存: $(ulimit -v)"
    
    echo ""
    echo "网络连接:"
    echo "  总连接数: $(ss -tuln | wc -l)"
    echo "  8080端口: $(ss -tuln | grep :8080 | wc -l || echo 0)"
}

# 优化用户级限制
optimize_user_limits() {
    print_step "优化用户级限制"
    
    # 尝试设置更高的文件描述符限制
    current_limit=$(ulimit -n)
    echo "当前文件描述符限制: $current_limit"
    
    # 尝试设置更高的限制
    for limit in 65535 32768 16384 8192; do
        if ulimit -n $limit 2>/dev/null; then
            print_success "文件描述符限制设置为: $limit"
            break
        fi
    done
    
    # 设置进程数限制
    if ulimit -u 32768 2>/dev/null; then
        print_success "进程数限制设置为: 32768"
    else
        print_warning "无法设置进程数限制"
    fi
}

# 创建Go运行时优化配置
create_go_optimization() {
    print_step "创建Go运行时优化配置"
    
    cat > "go_runtime_env.sh" << EOF
#!/bin/bash

# Go运行时环境变量优化
export GOMAXPROCS=\$(nproc)
export GOGC=100
export GODEBUG=gctrace=0

# 内存相关
export GOMEMLIMIT=8GiB

# 网络相关
export GODEBUG=netdns=go

echo "Go运行时参数已设置:"
echo "  GOMAXPROCS=\$GOMAXPROCS"
echo "  GOGC=\$GOGC"
echo "  GOMEMLIMIT=\$GOMEMLIMIT"
echo "  文件描述符限制: \$(ulimit -n)"
EOF

    chmod +x "go_runtime_env.sh"
    print_success "Go运行时优化配置已创建: go_runtime_env.sh"
}

# 创建优化的HTTP服务器配置
create_http_optimization() {
    print_step "创建HTTP服务器优化配置"
    
    # 创建优化的main.go配置建议
    cat > "http_server_optimization.md" << 'EOF'
# HTTP服务器优化建议

## 1. 服务器配置优化

在 `cmd/server/main.go` 中优化HTTP服务器配置：

```go
server := &http.Server{
    Addr:           ":" + cfg.Server.Port,
    Handler:        r,
    ReadTimeout:    5 * time.Second,   // 减少读取超时
    WriteTimeout:   5 * time.Second,   // 减少写入超时
    IdleTimeout:    30 * time.Second,  // 减少空闲超时
    MaxHeaderBytes: 1 << 16,           // 减少最大请求头大小 64KB
    
    // 启用HTTP/2
    TLSConfig: &tls.Config{
        NextProtos: []string{"h2", "http/1.1"},
    },
}
```

## 2. 连接池优化

优化数据库连接池配置：

```go
// 根据并发需求调整
sqlDB.SetMaxOpenConns(200)        // 减少到200
sqlDB.SetMaxIdleConns(50)         // 减少到50
sqlDB.SetConnMaxLifetime(5 * time.Minute)   // 减少生命周期
sqlDB.SetConnMaxIdleTime(2 * time.Minute)   // 减少空闲时间
```

## 3. 缓存优化

减少分片数量以降低开销：

```go
// 从32分片减少到16分片
cache.SetGlobalCache(cache.NewShardedCache(16))
```
EOF

    print_success "HTTP服务器优化建议已创建: http_server_optimization.md"
}

# 创建优化的启动脚本
create_optimized_startup() {
    print_step "创建优化的启动脚本"
    
    cat > "start_optimized.sh" << 'EOF'
#!/bin/bash

# Ryan Mall 用户级优化启动脚本
set -e

echo "🚀 启动优化的Ryan Mall服务..."

# 设置用户级限制
ulimit -n 65535 2>/dev/null || ulimit -n 32768 2>/dev/null || echo "无法设置文件描述符限制"
ulimit -u 32768 2>/dev/null || echo "无法设置进程数限制"

# 加载Go运行时优化
source ./go_runtime_env.sh

echo ""
echo "当前优化参数:"
echo "  文件描述符限制: $(ulimit -n)"
echo "  进程数限制: $(ulimit -u)"
echo "  CPU核心数: $(nproc)"

echo ""
echo "启动服务..."
go run cmd/server/main.go
EOF

    chmod +x "start_optimized.sh"
    print_success "优化启动脚本已创建: start_optimized.sh"
}

# 创建性能监控脚本
create_monitoring_script() {
    print_step "创建性能监控脚本"
    
    cat > "monitor_performance.sh" << 'EOF'
#!/bin/bash

# 性能监控脚本
echo "Ryan Mall 性能监控"
echo "按 Ctrl+C 退出"
echo ""

while true; do
    clear
    echo "=== Ryan Mall 性能监控 $(date) ==="
    echo ""
    
    echo "系统负载:"
    uptime
    echo ""
    
    echo "内存使用:"
    free -h | head -2
    echo ""
    
    echo "网络连接:"
    echo "  总连接数: $(ss -tuln | wc -l)"
    echo "  8080端口连接: $(ss -tuln | grep :8080 | wc -l || echo 0)"
    echo "  ESTABLISHED: $(ss -tun | grep ESTAB | wc -l)"
    echo ""
    
    echo "Go进程信息:"
    ps aux | grep "go run" | grep -v grep | head -3
    echo ""
    
    echo "最近的HTTP请求 (最后5个):"
    tail -5 /tmp/ryan_mall_requests.log 2>/dev/null || echo "无请求日志"
    
    sleep 2
done
EOF

    chmod +x "monitor_performance.sh"
    print_success "性能监控脚本已创建: monitor_performance.sh"
}

# 创建压力测试脚本
create_stress_test() {
    print_step "创建增强压力测试脚本"
    
    cat > "enhanced_stress_test.sh" << 'EOF'
#!/bin/bash

# 增强压力测试脚本
set -e

BASE_URL="http://localhost:8080"
API_URL="$BASE_URL/api/v1"

echo "=== 增强压力测试 ==="

# 预热
echo "预热服务..."
for i in {1..10}; do
    curl -s "$API_URL/products/7" > /dev/null
done

echo "开始压力测试..."

# 测试不同并发级别
for concurrent in 50 100 200 500 1000 1500 2000; do
    echo ""
    echo "=== ${concurrent}并发测试 ==="
    
    start_time=$(date +%s%N)
    
    # 执行并发请求
    seq 1 $concurrent | xargs -n1 -P${concurrent} -I{} bash -c "
        start=\$(date +%s%N)
        if curl -s --max-time 10 '$API_URL/products/7' > /dev/null 2>&1; then
            end=\$(date +%s%N)
            echo \$(( (end - start) / 1000000 ))
        else
            echo 'ERROR'
        fi
    " > /tmp/test_results.txt
    
    end_time=$(date +%s%N)
    total_time=$(( (end_time - start_time) / 1000000 ))
    
    # 统计结果
    success_count=$(grep -v ERROR /tmp/test_results.txt | wc -l)
    error_count=$(grep ERROR /tmp/test_results.txt | wc -l || echo 0)
    
    if [ $success_count -gt 0 ]; then
        qps=$(( concurrent * 1000 / total_time ))
        avg_time=$(grep -v ERROR /tmp/test_results.txt | awk '{sum+=$1} END {print int(sum/NR)}')
        min_time=$(grep -v ERROR /tmp/test_results.txt | sort -n | head -1)
        max_time=$(grep -v ERROR /tmp/test_results.txt | sort -n | tail -1)
        
        echo "  总耗时: ${total_time}ms"
        echo "  成功请求: $success_count/$concurrent ($(( success_count * 100 / concurrent ))%)"
        echo "  失败请求: $error_count"
        echo "  QPS: $qps"
        echo "  响应时间: 最小${min_time}ms, 最大${max_time}ms, 平均${avg_time}ms"
        
        # 如果错误率超过5%，停止测试
        if [ $error_count -gt $(( concurrent / 20 )) ]; then
            echo "  ⚠️  错误率过高，停止测试"
            break
        fi
    else
        echo "  ❌ 所有请求都失败了"
        break
    fi
    
    # 等待系统恢复
    sleep 2
done

rm -f /tmp/test_results.txt
echo ""
echo "压力测试完成"
EOF

    chmod +x "enhanced_stress_test.sh"
    print_success "增强压力测试脚本已创建: enhanced_stress_test.sh"
}

# 显示优化建议
show_optimization_tips() {
    print_step "优化建议"
    
    echo "1. 应用代码优化:"
    echo "   - 查看 http_server_optimization.md 中的建议"
    echo "   - 减少HTTP超时时间"
    echo "   - 优化数据库连接池配置"
    echo "   - 减少缓存分片数量"
    echo ""
    
    echo "2. 启动优化服务:"
    echo "   ./start_optimized.sh"
    echo ""
    
    echo "3. 运行性能测试:"
    echo "   ./enhanced_stress_test.sh"
    echo ""
    
    echo "4. 监控性能:"
    echo "   ./monitor_performance.sh"
    echo ""
    
    echo "5. 系统级优化 (需要sudo权限):"
    echo "   sudo ./system_network_optimization.sh"
}

# 主函数
main() {
    print_header
    
    show_current_status
    echo ""
    
    optimize_user_limits
    echo ""
    
    create_go_optimization
    echo ""
    
    create_http_optimization
    echo ""
    
    create_optimized_startup
    echo ""
    
    create_monitoring_script
    echo ""
    
    create_stress_test
    echo ""
    
    show_optimization_tips
    
    print_success "用户级优化完成！"
}

# 执行主函数
main "$@"
