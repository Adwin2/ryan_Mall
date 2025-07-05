#!/bin/bash

# Redis集群初始化脚本
echo "🚀 开始初始化Redis集群..."

# 等待所有Redis节点启动
echo "⏳ 等待Redis节点启动..."
sleep 30

# 检查节点是否可用
check_node() {
    local ip=$1
    local port=$2
    echo "检查节点 $ip:$port..."
    
    for i in {1..30}; do
        if redis-cli -h $ip -p $port ping > /dev/null 2>&1; then
            echo "✅ 节点 $ip:$port 已就绪"
            return 0
        fi
        echo "⏳ 等待节点 $ip:$port 启动... ($i/30)"
        sleep 2
    done
    
    echo "❌ 节点 $ip:$port 启动失败"
    return 1
}

# 检查所有节点
echo "🔍 检查所有Redis节点状态..."

if ! check_node "172.20.0.11" "6379"; then
    echo "❌ 集群初始化失败：节点 172.20.0.11:6379 不可用"
    exit 1
fi

if ! check_node "172.20.0.12" "6379"; then
    echo "❌ 集群初始化失败：节点 172.20.0.12:6379 不可用"
    exit 1
fi

if ! check_node "172.20.0.13" "6379"; then
    echo "❌ 集群初始化失败：节点 172.20.0.13:6379 不可用"
    exit 1
fi

if ! check_node "172.20.0.14" "6379"; then
    echo "❌ 集群初始化失败：节点 172.20.0.14:6379 不可用"
    exit 1
fi

if ! check_node "172.20.0.15" "6379"; then
    echo "❌ 集群初始化失败：节点 172.20.0.15:6379 不可用"
    exit 1
fi

if ! check_node "172.20.0.16" "6379"; then
    echo "❌ 集群初始化失败：节点 172.20.0.16:6379 不可用"
    exit 1
fi

echo "✅ 所有节点已就绪，开始创建集群..."

# 创建Redis集群
# 3个主节点，3个从节点，每个主节点有1个从节点
redis-cli --cluster create \
    172.20.0.11:6379 \
    172.20.0.12:6379 \
    172.20.0.13:6379 \
    172.20.0.14:6379 \
    172.20.0.15:6379 \
    172.20.0.16:6379 \
    --cluster-replicas 1 \
    --cluster-yes

if [ $? -eq 0 ]; then
    echo "🎉 Redis集群创建成功！"
    
    # 显示集群信息
    echo ""
    echo "📊 集群信息："
    redis-cli -h 172.20.0.11 -p 6379 cluster info
    
    echo ""
    echo "🔗 集群节点："
    redis-cli -h 172.20.0.11 -p 6379 cluster nodes
    
    echo ""
    echo "✅ Redis集群部署完成！"
    echo "🌐 集群访问地址："
    echo "   - 节点1: 172.20.0.11:6379"
    echo "   - 节点2: 172.20.0.12:6379"
    echo "   - 节点3: 172.20.0.13:6379"
    echo "   - 节点4: 172.20.0.14:6379"
    echo "   - 节点5: 172.20.0.15:6379"
    echo "   - 节点6: 172.20.0.16:6379"
    echo ""
    echo "🎯 Redis Insight管理界面: http://localhost:8001"
    
else
    echo "❌ Redis集群创建失败"
    exit 1
fi

# 保持容器运行
echo "🔄 集群初始化完成，保持监控状态..."
while true; do
    sleep 60
    # 检查集群健康状态
    if ! redis-cli -h 172.20.0.11 -p 6379 cluster info | grep -q "cluster_state:ok"; then
        echo "⚠️  检测到集群状态异常"
        redis-cli -h 172.20.0.11 -p 6379 cluster info
    fi
done
