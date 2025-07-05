#!/bin/bash

# Go运行时环境变量优化
export GOMAXPROCS=$(nproc)
export GOGC=100
export GODEBUG=gctrace=0

# 内存相关
export GOMEMLIMIT=8GiB

# 网络相关
export GODEBUG=netdns=go

echo "Go运行时参数已设置:"
echo "  GOMAXPROCS=$GOMAXPROCS"
echo "  GOGC=$GOGC"
echo "  GOMEMLIMIT=$GOMEMLIMIT"
echo "  文件描述符限制: $(ulimit -n)"
