#!/bin/bash

# Ryan Mall 订单模块API测试脚本

BASE_URL="http://localhost:8080/api/v1"

echo "=== Ryan Mall 订单模块API测试 ==="
echo

# 首先登录获取用户token
echo "1. 用户登录获取Token..."
LOGIN_RESPONSE=$(curl -s -X POST $BASE_URL/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser2",
    "password": "newpassword123"
  }')

echo "登录响应: $LOGIN_RESPONSE"
TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
echo "用户Token: $TOKEN"
echo

# 2. 先添加一些商品到购物车
echo "2. 添加商品到购物车..."
ADD_CART_RESPONSE1=$(curl -s -X POST $BASE_URL/cart \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": 5,
    "quantity": 2
  }')

echo "添加商品到购物车响应: $ADD_CART_RESPONSE1"
echo

# 3. 获取购物车查看商品
echo "3. 获取购物车..."
GET_CART_RESPONSE=$(curl -s -X GET $BASE_URL/cart \
  -H "Authorization: Bearer $TOKEN")

echo "购物车响应: $GET_CART_RESPONSE"

# 提取购物车项ID
CART_ITEM_ID=$(echo $GET_CART_RESPONSE | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
echo "购物车项ID: $CART_ITEM_ID"
echo

# 4. 测试创建订单
echo "4. 测试创建订单..."
CREATE_ORDER_RESPONSE=$(curl -s -X POST $BASE_URL/orders \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"cart_item_ids\": [$CART_ITEM_ID],
    \"shipping_address\": {
      \"name\": \"张三\",
      \"phone\": \"13800138000\",
      \"province\": \"广东省\",
      \"city\": \"深圳市\",
      \"district\": \"南山区\",
      \"address\": \"科技园南区深南大道10000号\",
      \"zipcode\": \"518000\"
    },
    \"payment_method\": \"alipay\",
    \"contact_phone\": \"13800138000\",
    \"remark\": \"请尽快发货\"
  }")

echo "创建订单响应: $CREATE_ORDER_RESPONSE"

# 提取订单ID和订单号
ORDER_ID=$(echo $CREATE_ORDER_RESPONSE | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
ORDER_NO=$(echo $CREATE_ORDER_RESPONSE | grep -o '"order_no":"[^"]*"' | cut -d'"' -f4)
echo "订单ID: $ORDER_ID"
echo "订单号: $ORDER_NO"
echo

# 5. 测试获取订单详情
echo "5. 测试获取订单详情..."
GET_ORDER_RESPONSE=$(curl -s -X GET $BASE_URL/orders/$ORDER_ID \
  -H "Authorization: Bearer $TOKEN")

echo "订单详情响应: $GET_ORDER_RESPONSE"
echo

# 6. 测试根据订单号获取订单
echo "6. 测试根据订单号获取订单..."
GET_ORDER_BY_NO_RESPONSE=$(curl -s -X GET $BASE_URL/orders/no/$ORDER_NO \
  -H "Authorization: Bearer $TOKEN")

echo "根据订单号获取订单响应: $GET_ORDER_BY_NO_RESPONSE"
echo

# 7. 测试获取订单列表
echo "7. 测试获取订单列表..."
GET_ORDER_LIST_RESPONSE=$(curl -s -X GET "$BASE_URL/orders?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN")

echo "订单列表响应: $GET_ORDER_LIST_RESPONSE"
echo

# 8. 测试获取订单统计
echo "8. 测试获取订单统计..."
GET_ORDER_STATS_RESPONSE=$(curl -s -X GET $BASE_URL/orders/statistics \
  -H "Authorization: Bearer $TOKEN")

echo "订单统计响应: $GET_ORDER_STATS_RESPONSE"
echo

# 9. 测试支付订单
echo "9. 测试支付订单..."
PAY_ORDER_RESPONSE=$(curl -s -X POST $BASE_URL/orders/$ORDER_ID/pay \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "payment_method": "alipay"
  }')

echo "支付订单响应: $PAY_ORDER_RESPONSE"
echo

# 10. 验证订单状态已更新
echo "10. 验证订单状态已更新..."
GET_ORDER_AFTER_PAY_RESPONSE=$(curl -s -X GET $BASE_URL/orders/$ORDER_ID \
  -H "Authorization: Bearer $TOKEN")

echo "支付后订单详情: $GET_ORDER_AFTER_PAY_RESPONSE"
echo

# 11. 测试按状态筛选订单
echo "11. 测试按状态筛选订单（已支付）..."
GET_PAID_ORDERS_RESPONSE=$(curl -s -X GET "$BASE_URL/orders?status=2&page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN")

echo "已支付订单列表: $GET_PAID_ORDERS_RESPONSE"
echo

# 12. 创建另一个订单用于取消测试
echo "12. 创建另一个订单用于取消测试..."
# 先添加商品到购物车
ADD_CART_RESPONSE2=$(curl -s -X POST $BASE_URL/cart \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": 5,
    "quantity": 1
  }')

# 获取新的购物车项ID
GET_CART_RESPONSE2=$(curl -s -X GET $BASE_URL/cart \
  -H "Authorization: Bearer $TOKEN")

CART_ITEM_ID2=$(echo $GET_CART_RESPONSE2 | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)

# 创建第二个订单
CREATE_ORDER_RESPONSE2=$(curl -s -X POST $BASE_URL/orders \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"cart_item_ids\": [$CART_ITEM_ID2],
    \"shipping_address\": {
      \"name\": \"李四\",
      \"phone\": \"13900139000\",
      \"province\": \"北京市\",
      \"city\": \"北京市\",
      \"district\": \"朝阳区\",
      \"address\": \"建国门外大街1号\",
      \"zipcode\": \"100000\"
    },
    \"payment_method\": \"wechat\",
    \"contact_phone\": \"13900139000\",
    \"remark\": \"测试订单\"
  }")

ORDER_ID2=$(echo $CREATE_ORDER_RESPONSE2 | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
echo "第二个订单ID: $ORDER_ID2"
echo

# 13. 测试取消订单
echo "13. 测试取消订单..."
CANCEL_ORDER_RESPONSE=$(curl -s -X PUT $BASE_URL/orders/$ORDER_ID2/cancel \
  -H "Authorization: Bearer $TOKEN")

echo "取消订单响应: $CANCEL_ORDER_RESPONSE"
echo

# 14. 验证订单已取消
echo "14. 验证订单已取消..."
GET_CANCELLED_ORDER_RESPONSE=$(curl -s -X GET $BASE_URL/orders/$ORDER_ID2 \
  -H "Authorization: Bearer $TOKEN")

echo "取消后订单详情: $GET_CANCELLED_ORDER_RESPONSE"
echo

# 15. 测试确认收货（需要先模拟发货状态）
echo "15. 测试确认收货..."
# 注意：在实际应用中，发货状态应该由管理员或系统设置
# 这里我们直接测试确认收货，可能会失败因为订单状态不是已发货
CONFIRM_ORDER_RESPONSE=$(curl -s -X PUT $BASE_URL/orders/$ORDER_ID/confirm \
  -H "Authorization: Bearer $TOKEN")

echo "确认收货响应: $CONFIRM_ORDER_RESPONSE"
echo

# 16. 测试无权限访问他人订单
echo "16. 测试无权限访问他人订单..."
NO_AUTH_ORDER_RESPONSE=$(curl -s -X GET $BASE_URL/orders/999999 \
  -H "Authorization: Bearer $TOKEN")

echo "无权限访问响应: $NO_AUTH_ORDER_RESPONSE"
echo

# 17. 测试无认证访问
echo "17. 测试无认证访问..."
NO_TOKEN_RESPONSE=$(curl -s -X GET $BASE_URL/orders)

echo "无认证访问响应: $NO_TOKEN_RESPONSE"
echo

# 18. 测试创建订单时购物车为空
echo "18. 测试创建订单时购物车为空..."
EMPTY_CART_ORDER_RESPONSE=$(curl -s -X POST $BASE_URL/orders \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "cart_item_ids": [],
    "shipping_address": {
      "name": "测试",
      "phone": "13800138000",
      "province": "广东省",
      "city": "深圳市",
      "district": "南山区",
      "address": "测试地址",
      "zipcode": "518000"
    },
    "payment_method": "alipay",
    "contact_phone": "13800138000"
  }')

echo "空购物车创建订单响应: $EMPTY_CART_ORDER_RESPONSE"
echo

# 19. 测试最终订单统计
echo "19. 测试最终订单统计..."
FINAL_STATS_RESPONSE=$(curl -s -X GET $BASE_URL/orders/statistics \
  -H "Authorization: Bearer $TOKEN")

echo "最终订单统计: $FINAL_STATS_RESPONSE"
echo

echo "=== 订单模块API测试完成 ==="
