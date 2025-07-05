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
