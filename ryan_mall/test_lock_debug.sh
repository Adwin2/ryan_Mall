#!/bin/bash

# 测试分布式锁调试脚本

echo "=== 分布式锁调试测试 ==="

# 创建测试商品
echo "创建测试商品..."
PRODUCT_ID=$(uuidgen)
docker exec -i ryan-mall-mysql mysql -u root -p123456 ryan_mall << EOF
INSERT INTO products (product_id, name, description, price, stock, category_id, created_at, updated_at) 
VALUES ('$PRODUCT_ID', '测试商品', '用于测试分布式锁', 99.99, 5, 'test-category', NOW(), NOW());
EOF

echo "商品ID: $PRODUCT_ID"
echo "初始库存: 5"

# 检查Redis中是否有锁的痕迹
echo ""
echo "=== 发送5个并发请求 ==="

# 并发发送5个库存预留请求
for i in {1..5}; do
    (
        ORDER_ID="order-$i-$(date +%s)"
        echo "发送请求 $i (订单: $ORDER_ID)"
        
        RESPONSE=$(curl -s -w "%{http_code}" -X POST \
            "http://localhost:8082/api/v1/products/$PRODUCT_ID/stock/reserve" \
            -H "Content-Type: application/json" \
            -d "{\"quantity\": 1, \"order_id\": \"$ORDER_ID\"}")
        
        HTTP_CODE="${RESPONSE: -3}"
        BODY="${RESPONSE%???}"
        
        echo "请求 $i 完成: HTTP $HTTP_CODE"
        if [ "$HTTP_CODE" != "200" ]; then
            echo "响应体: $BODY"
        fi
    ) &
done

# 等待所有请求完成
wait

echo ""
echo "=== 检查最终库存 ==="
FINAL_STOCK=$(curl -s "http://localhost:8082/api/v1/products/$PRODUCT_ID/stock" | jq -r '.stock')
echo "最终库存: $FINAL_STOCK"

echo ""
echo "=== 检查数据库中的库存 ==="
DB_STOCK=$(docker exec -i ryan-mall-mysql mysql -u root -p123456 ryan_mall -e "SELECT stock FROM products WHERE product_id='$PRODUCT_ID';" | tail -n 1)
echo "数据库库存: $DB_STOCK"

echo ""
echo "=== 分析结果 ==="
if [ "$FINAL_STOCK" = "0" ]; then
    echo "✓ 分布式锁工作正常：5个请求成功，库存从5减少到0"
else
    echo "✗ 分布式锁有问题：期望库存0，实际库存$FINAL_STOCK"
fi

# 清理测试数据
echo ""
echo "清理测试数据..."
docker exec -i ryan-mall-mysql mysql -u root -p123456 ryan_mall -e "DELETE FROM products WHERE product_id='$PRODUCT_ID';"
echo "✓ 清理完成"
