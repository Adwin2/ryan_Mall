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
