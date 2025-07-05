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
