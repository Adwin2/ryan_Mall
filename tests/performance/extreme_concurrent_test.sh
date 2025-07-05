#!/bin/bash

# 极限并发性能测试
set -e

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

# 配置
BASE_URL="http://localhost:8080"
API_URL="$BASE_URL/api/v1"

echo -e "${BLUE}=== 极限并发性能测试 ===${NC}"

# 检查服务
echo "检查服务状态..."
if ! curl -s "$BASE_URL/ping" > /dev/null; then
    echo "❌ 服务未运行"
    exit 1
fi
echo "✅ 服务运行正常"

echo ""
echo -e "${YELLOW}开始极限并发测试...${NC}"

# 预热缓存
echo "预热缓存..."
for i in {1..5}; do
    curl -s "$API_URL/products/7" > /dev/null
    curl -s "$API_URL/products?page=1&page_size=10" > /dev/null
done

# 测试极限并发级别
for concurrent in 100 200 500 1000; do
    echo ""
    echo "=== ${concurrent}并发测试 ==="
    
    # 执行测试
    echo "执行${concurrent}个请求，${concurrent}并发..."
    start_time=$(date +%s%N)
    
    # 使用xargs控制并发
    seq 1 $concurrent | xargs -n1 -P${concurrent} -I{} bash -c "
        start=\$(date +%s%N)
        response=\$(curl -s -w '%{http_code}' '$API_URL/products/7')
        end=\$(date +%s%N)
        time=\$(( (end - start) / 1000000 ))
        code=\$(echo \"\$response\" | tail -c 4)
        echo \"\$time,\$code\"
    " > /tmp/results_${concurrent}.txt 2>/dev/null
    
    end_time=$(date +%s%N)
    total_time=$(( (end_time - start_time) / 1000000 ))
    
    # 分析结果
    if [ -f /tmp/results_${concurrent}.txt ] && [ -s /tmp/results_${concurrent}.txt ]; then
        # 统计成功请求
        success_count=$(grep ",200" /tmp/results_${concurrent}.txt | wc -l)
        error_count=$(( concurrent - success_count ))
        
        if [ $success_count -gt 0 ]; then
            # 提取响应时间
            grep ",200" /tmp/results_${concurrent}.txt | cut -d',' -f1 > /tmp/times_${concurrent}.txt
            
            min_time=$(sort -n /tmp/times_${concurrent}.txt | head -1)
            max_time=$(sort -n /tmp/times_${concurrent}.txt | tail -1)
            avg_time=$(awk '{sum+=$1} END {print int(sum/NR)}' /tmp/times_${concurrent}.txt)
            
            # 计算百分位数
            p95_time=$(sort -n /tmp/times_${concurrent}.txt | awk 'BEGIN{c=0} {a[c++]=$1} END{print a[int(c*0.95)]}')
            p99_time=$(sort -n /tmp/times_${concurrent}.txt | awk 'BEGIN{c=0} {a[c++]=$1} END{print a[int(c*0.99)]}')
            
            qps=$(( success_count * 1000 / total_time ))
            success_rate=$(( success_count * 100 / concurrent ))
            
            echo "  总耗时: ${total_time}ms"
            echo "  成功请求: ${success_count}/${concurrent} (${success_rate}%)"
            echo "  错误请求: ${error_count}"
            echo "  QPS: ${qps}"
            echo "  响应时间统计:"
            echo "    最小: ${min_time}ms"
            echo "    最大: ${max_time}ms"
            echo "    平均: ${avg_time}ms"
            echo "    P95: ${p95_time}ms"
            echo "    P99: ${p99_time}ms"
            
            # 性能评估
            if [ $qps -gt 2000 ]; then
                echo -e "  ${GREEN}✅ 性能优秀${NC}"
            elif [ $qps -gt 1000 ]; then
                echo -e "  ${YELLOW}⚠️  性能良好${NC}"
            else
                echo -e "  ${RED}❌ 性能需要优化${NC}"
            fi
        else
            echo -e "  ${RED}❌ 所有请求都失败了${NC}"
        fi
        
        # 清理
        rm -f /tmp/results_${concurrent}.txt /tmp/times_${concurrent}.txt
    else
        echo -e "  ${RED}❌ 测试失败，无结果数据${NC}"
    fi
    
    # 检查系统资源
    echo "  系统负载: $(uptime | awk -F'load average:' '{print $2}')"
    
    sleep 2
done

echo ""
echo -e "${GREEN}=== 极限测试完成 ===${NC}"

# 最终状态
echo "最终缓存状态:"
curl -s "$BASE_URL/cache/stats" | jq '.'

echo ""
echo "系统资源使用情况:"
echo "内存使用:"
free -h | grep "内存"
echo "网络连接数:"
netstat -an | grep :8080 | wc -l
