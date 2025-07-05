#!/bin/bash

# 简化版并发性能测试
set -e

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 配置
BASE_URL="http://localhost:8080"
API_URL="$BASE_URL/api/v1"

echo -e "${BLUE}=== 并发性能测试 ===${NC}"

# 检查服务
echo "检查服务状态..."
if ! curl -s "$BASE_URL/ping" > /dev/null; then
    echo "❌ 服务未运行"
    exit 1
fi
echo "✅ 服务运行正常"

# 获取缓存状态
echo "当前缓存状态:"
curl -s "$BASE_URL/cache/stats" | jq '.'

echo ""
echo -e "${YELLOW}开始并发测试...${NC}"

# 测试不同并发级别
for concurrent in 1 5 10 20 50; do
    echo ""
    echo "=== ${concurrent}并发测试 ==="
    
    # 预热
    curl -s "$API_URL/products/7" > /dev/null
    
    # 执行测试
    echo "执行50个请求，${concurrent}并发..."
    start_time=$(date +%s%N)
    
    # 使用xargs控制并发
    seq 1 50 | xargs -n1 -P${concurrent} -I{} bash -c "
        start=\$(date +%s%N)
        curl -s '$API_URL/products/7' > /dev/null
        end=\$(date +%s%N)
        echo \$(( (end - start) / 1000000 ))
    " > /tmp/times_${concurrent}.txt
    
    end_time=$(date +%s%N)
    total_time=$(( (end_time - start_time) / 1000000 ))
    
    # 分析结果
    if [ -f /tmp/times_${concurrent}.txt ]; then
        min_time=$(sort -n /tmp/times_${concurrent}.txt | head -1)
        max_time=$(sort -n /tmp/times_${concurrent}.txt | tail -1)
        avg_time=$(awk '{sum+=$1} END {print int(sum/NR)}' /tmp/times_${concurrent}.txt)
        qps=$(( 50 * 1000 / total_time ))
        
        echo "  总耗时: ${total_time}ms"
        echo "  QPS: ${qps}"
        echo "  响应时间: 最小${min_time}ms, 最大${max_time}ms, 平均${avg_time}ms"
        
        # 清理
        rm -f /tmp/times_${concurrent}.txt
    fi
    
    sleep 1
done

echo ""
echo -e "${GREEN}=== 测试完成 ===${NC}"

# 最终缓存状态
echo "最终缓存状态:"
curl -s "$BASE_URL/cache/stats" | jq '.'
