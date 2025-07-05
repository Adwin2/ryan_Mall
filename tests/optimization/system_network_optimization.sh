#!/bin/bash

# 系统级网络优化脚本
# 针对高并发Web服务进行系统参数调优

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
    echo "           系统级网络性能优化"
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

# 检查权限
check_permissions() {
    print_step "检查系统权限"
    
    if [[ $EUID -eq 0 ]]; then
        print_success "以root权限运行，可以进行系统级优化"
        return 0
    else
        print_warning "当前非root权限，将尝试使用sudo"
        if sudo -n true 2>/dev/null; then
            print_success "sudo权限可用"
            return 0
        else
            print_error "需要root权限进行系统级优化"
            echo "请使用以下命令之一："
            echo "1. sudo ./system_network_optimization.sh"
            echo "2. su - root 然后运行脚本"
            exit 1
        fi
    fi
}

# 备份当前配置
backup_configs() {
    print_step "备份当前系统配置"
    
    BACKUP_DIR="/tmp/ryan_mall_backup_$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$BACKUP_DIR"
    
    # 备份网络参数
    if [ -f /etc/sysctl.conf ]; then
        cp /etc/sysctl.conf "$BACKUP_DIR/sysctl.conf.bak"
        print_success "已备份 /etc/sysctl.conf"
    fi
    
    # 备份limits配置
    if [ -f /etc/security/limits.conf ]; then
        cp /etc/security/limits.conf "$BACKUP_DIR/limits.conf.bak"
        print_success "已备份 /etc/security/limits.conf"
    fi
    
    # 保存当前网络参数
    sysctl -a > "$BACKUP_DIR/current_sysctl.txt" 2>/dev/null
    ulimit -a > "$BACKUP_DIR/current_limits.txt"
    
    echo "备份目录: $BACKUP_DIR"
    print_success "配置备份完成"
}

# 显示当前网络参数
show_current_params() {
    print_step "当前网络参数"
    
    echo "TCP连接相关:"
    echo "  net.core.somaxconn = $(sysctl -n net.core.somaxconn 2>/dev/null || echo 'N/A')"
    echo "  net.ipv4.tcp_max_syn_backlog = $(sysctl -n net.ipv4.tcp_max_syn_backlog 2>/dev/null || echo 'N/A')"
    echo "  net.core.netdev_max_backlog = $(sysctl -n net.core.netdev_max_backlog 2>/dev/null || echo 'N/A')"
    echo "  net.ipv4.tcp_fin_timeout = $(sysctl -n net.ipv4.tcp_fin_timeout 2>/dev/null || echo 'N/A')"
    echo "  net.ipv4.tcp_keepalive_time = $(sysctl -n net.ipv4.tcp_keepalive_time 2>/dev/null || echo 'N/A')"
    
    echo ""
    echo "文件描述符限制:"
    echo "  当前进程: $(ulimit -n)"
    echo "  系统最大: $(cat /proc/sys/fs/file-max 2>/dev/null || echo 'N/A')"
    
    echo ""
    echo "当前网络连接数:"
    echo "  总连接数: $(ss -tuln | wc -l)"
    echo "  监听端口: $(ss -tuln | grep LISTEN | wc -l)"
    echo "  8080端口: $(ss -tuln | grep :8080 | wc -l)"
}

# 优化网络参数
optimize_network_params() {
    print_step "优化网络参数"
    
    # 创建临时配置文件
    TEMP_SYSCTL="/tmp/ryan_mall_sysctl.conf"
    
    cat > "$TEMP_SYSCTL" << 'EOF'
# Ryan Mall 网络性能优化配置
# 生成时间: $(date)

# TCP连接队列优化
net.core.somaxconn = 65535
net.ipv4.tcp_max_syn_backlog = 65535
net.core.netdev_max_backlog = 5000

# TCP连接优化
net.ipv4.tcp_fin_timeout = 30
net.ipv4.tcp_keepalive_time = 1200
net.ipv4.tcp_keepalive_probes = 3
net.ipv4.tcp_keepalive_intvl = 15

# TCP窗口缩放
net.ipv4.tcp_window_scaling = 1
net.ipv4.tcp_timestamps = 1
net.ipv4.tcp_sack = 1

# TCP拥塞控制
net.ipv4.tcp_congestion_control = bbr
net.core.default_qdisc = fq

# 内存优化
net.core.rmem_default = 262144
net.core.rmem_max = 16777216
net.core.wmem_default = 262144
net.core.wmem_max = 16777216
net.ipv4.tcp_rmem = 4096 65536 16777216
net.ipv4.tcp_wmem = 4096 65536 16777216

# 连接跟踪优化
net.netfilter.nf_conntrack_max = 1000000
net.netfilter.nf_conntrack_tcp_timeout_established = 1200

# 端口范围
net.ipv4.ip_local_port_range = 1024 65535

# TIME_WAIT优化
net.ipv4.tcp_tw_reuse = 1
net.ipv4.tcp_max_tw_buckets = 6000

# 文件系统优化
fs.file-max = 1000000
fs.nr_open = 1000000
EOF

    # 应用网络参数
    if [[ $EUID -eq 0 ]]; then
        sysctl -p "$TEMP_SYSCTL"
    else
        sudo sysctl -p "$TEMP_SYSCTL"
    fi
    
    print_success "网络参数优化完成"
}

# 优化文件描述符限制
optimize_file_limits() {
    print_step "优化文件描述符限制"
    
    # 创建临时limits配置
    TEMP_LIMITS="/tmp/ryan_mall_limits.conf"
    
    cat > "$TEMP_LIMITS" << 'EOF'
# Ryan Mall 文件描述符限制优化
# 生成时间: $(date)

# 所有用户的软限制和硬限制
* soft nofile 65535
* hard nofile 65535
* soft nproc 65535
* hard nproc 65535

# root用户的限制
root soft nofile 65535
root hard nofile 65535
root soft nproc 65535
root hard nproc 65535
EOF

    # 应用文件限制
    if [[ $EUID -eq 0 ]]; then
        cat "$TEMP_LIMITS" >> /etc/security/limits.conf
    else
        sudo bash -c "cat '$TEMP_LIMITS' >> /etc/security/limits.conf"
    fi
    
    # 设置当前会话的限制
    ulimit -n 65535 2>/dev/null || print_warning "无法设置当前会话的文件描述符限制"
    
    print_success "文件描述符限制优化完成"
}

# 优化Go运行时参数
optimize_go_runtime() {
    print_step "优化Go运行时参数"
    
    # 创建Go运行时优化脚本
    cat > "/tmp/go_runtime_optimization.sh" << 'EOF'
#!/bin/bash

# Go运行时环境变量优化
export GOMAXPROCS=$(nproc)
export GOGC=100
export GODEBUG=gctrace=0

# 内存相关
export GOMEMLIMIT=8GiB

echo "Go运行时参数已优化:"
echo "  GOMAXPROCS=$GOMAXPROCS"
echo "  GOGC=$GOGC"
echo "  GOMEMLIMIT=$GOMEMLIMIT"
EOF

    chmod +x "/tmp/go_runtime_optimization.sh"
    
    print_success "Go运行时优化脚本已创建: /tmp/go_runtime_optimization.sh"
    print_warning "请在启动服务前执行: source /tmp/go_runtime_optimization.sh"
}

# 创建服务启动脚本
create_optimized_startup() {
    print_step "创建优化的服务启动脚本"
    
    cat > "start_optimized_server.sh" << 'EOF'
#!/bin/bash

# Ryan Mall 优化启动脚本
set -e

echo "🚀 启动优化的Ryan Mall服务..."

# 设置Go运行时参数
export GOMAXPROCS=$(nproc)
export GOGC=100
export GOMEMLIMIT=8GiB

# 显示优化参数
echo "Go运行时参数:"
echo "  GOMAXPROCS=$GOMAXPROCS"
echo "  GOGC=$GOGC"
echo "  GOMEMLIMIT=$GOMEMLIMIT"

# 显示系统参数
echo ""
echo "系统网络参数:"
echo "  somaxconn: $(sysctl -n net.core.somaxconn)"
echo "  max_syn_backlog: $(sysctl -n net.ipv4.tcp_max_syn_backlog)"
echo "  文件描述符限制: $(ulimit -n)"

echo ""
echo "启动服务..."
go run cmd/server/main.go
EOF

    chmod +x "start_optimized_server.sh"
    print_success "优化启动脚本已创建: start_optimized_server.sh"
}

# 验证优化效果
verify_optimization() {
    print_step "验证优化效果"
    
    echo "优化后的网络参数:"
    echo "  net.core.somaxconn = $(sysctl -n net.core.somaxconn)"
    echo "  net.ipv4.tcp_max_syn_backlog = $(sysctl -n net.ipv4.tcp_max_syn_backlog)"
    echo "  net.core.netdev_max_backlog = $(sysctl -n net.core.netdev_max_backlog)"
    echo "  文件描述符限制 = $(ulimit -n)"
    
    echo ""
    echo "TCP拥塞控制算法:"
    echo "  当前算法: $(sysctl -n net.ipv4.tcp_congestion_control)"
    echo "  可用算法: $(sysctl -n net.ipv4.tcp_available_congestion_control)"
}

# 性能测试建议
performance_test_suggestions() {
    print_step "性能测试建议"
    
    echo "建议的测试步骤:"
    echo ""
    echo "1. 重启Ryan Mall服务:"
    echo "   ./start_optimized_server.sh"
    echo ""
    echo "2. 运行基准测试:"
    echo "   ./simple_concurrent_test.sh"
    echo ""
    echo "3. 运行极限测试:"
    echo "   ./extreme_concurrent_test.sh"
    echo ""
    echo "4. 监控系统资源:"
    echo "   watch -n 1 'ss -tuln | grep :8080 | wc -l'"
    echo "   watch -n 1 'cat /proc/loadavg'"
    echo ""
    echo "预期改善:"
    echo "  - QPS提升: 20-50%"
    echo "  - 响应时间降低: 10-30%"
    echo "  - 支持更高并发: 2000-5000"
    echo "  - 系统稳定性提升"
}

# 回滚说明
rollback_instructions() {
    print_step "回滚说明"
    
    echo "如需回滚优化，请执行:"
    echo ""
    echo "1. 恢复sysctl配置:"
    echo "   sudo cp $BACKUP_DIR/sysctl.conf.bak /etc/sysctl.conf"
    echo "   sudo sysctl -p"
    echo ""
    echo "2. 恢复limits配置:"
    echo "   sudo cp $BACKUP_DIR/limits.conf.bak /etc/security/limits.conf"
    echo ""
    echo "3. 重启系统以完全恢复:"
    echo "   sudo reboot"
}

# 主函数
main() {
    print_header
    
    # 检查权限
    check_permissions
    
    # 显示当前状态
    show_current_params
    echo ""
    
    # 确认执行
    read -p "是否继续进行系统级网络优化? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "优化已取消"
        exit 0
    fi
    
    # 执行优化
    backup_configs
    echo ""
    
    optimize_network_params
    echo ""
    
    optimize_file_limits
    echo ""
    
    optimize_go_runtime
    echo ""
    
    create_optimized_startup
    echo ""
    
    verify_optimization
    echo ""
    
    performance_test_suggestions
    echo ""
    
    rollback_instructions
    
    print_success "系统级网络优化完成！"
    print_warning "建议重启系统以确保所有优化生效"
}

# 执行主函数
main "$@"
