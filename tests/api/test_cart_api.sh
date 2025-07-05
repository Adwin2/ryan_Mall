#!/bin/bash

# Ryan Mall 购物车模块API测试脚本

BASE_URL="http://localhost:8080/api/v1"

echo "=== Ryan Mall 购物车模块API测试 ==="
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

# 2. 获取商品列表，选择要添加到购物车的商品
echo "2. 获取商品列表..."
PRODUCTS_RESPONSE=$(curl -s -X GET "$BASE_URL/products?page=1&page_size=5")
echo "商品列表响应: $PRODUCTS_RESPONSE"

# 提取第一个商品的ID（假设存在）
PRODUCT1_ID=$(echo $PRODUCTS_RESPONSE | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
PRODUCT2_ID=$(echo $PRODUCTS_RESPONSE | grep -o '"id":[0-9]*' | head -2 | tail -1 | cut -d':' -f2)
echo "商品1 ID: $PRODUCT1_ID"
echo "商品2 ID: $PRODUCT2_ID"
echo

# 3. 测试添加商品到购物车
echo "3. 测试添加商品到购物车..."
ADD_CART_RESPONSE1=$(curl -s -X POST $BASE_URL/cart \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"product_id\": $PRODUCT1_ID,
    \"quantity\": 2
  }")

echo "添加商品1到购物车响应: $ADD_CART_RESPONSE1"
echo

ADD_CART_RESPONSE2=$(curl -s -X POST $BASE_URL/cart \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"product_id\": $PRODUCT2_ID,
    \"quantity\": 1
  }")

echo "添加商品2到购物车响应: $ADD_CART_RESPONSE2"
echo

# 4. 测试获取购物车
echo "4. 测试获取购物车..."
GET_CART_RESPONSE=$(curl -s -X GET $BASE_URL/cart \
  -H "Authorization: Bearer $TOKEN")

echo "购物车响应: $GET_CART_RESPONSE"
echo

# 提取购物车项ID
CART_ITEM_ID=$(echo $GET_CART_RESPONSE | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
echo "购物车项ID: $CART_ITEM_ID"
echo

# 5. 测试获取购物车汇总
echo "5. 测试获取购物车汇总..."
CART_SUMMARY_RESPONSE=$(curl -s -X GET $BASE_URL/cart/summary \
  -H "Authorization: Bearer $TOKEN")

echo "购物车汇总响应: $CART_SUMMARY_RESPONSE"
echo

# 6. 测试获取购物车商品数量
echo "6. 测试获取购物车商品数量..."
CART_COUNT_RESPONSE=$(curl -s -X GET $BASE_URL/cart/count \
  -H "Authorization: Bearer $TOKEN")

echo "购物车数量响应: $CART_COUNT_RESPONSE"
echo

# 7. 测试更新购物车商品数量
echo "7. 测试更新购物车商品数量..."
UPDATE_CART_RESPONSE=$(curl -s -X PUT $BASE_URL/cart/$CART_ITEM_ID \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "quantity": 3
  }')

echo "更新购物车响应: $UPDATE_CART_RESPONSE"
echo

# 8. 再次获取购物车验证更新
echo "8. 验证购物车更新..."
GET_CART_RESPONSE2=$(curl -s -X GET $BASE_URL/cart \
  -H "Authorization: Bearer $TOKEN")

echo "更新后的购物车: $GET_CART_RESPONSE2"
echo

# 9. 测试重复添加同一商品（应该增加数量）
echo "9. 测试重复添加同一商品..."
ADD_SAME_PRODUCT_RESPONSE=$(curl -s -X POST $BASE_URL/cart \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"product_id\": $PRODUCT1_ID,
    \"quantity\": 1
  }")

echo "重复添加商品响应: $ADD_SAME_PRODUCT_RESPONSE"
echo

# 10. 测试批量添加商品
echo "10. 测试批量添加商品..."
BATCH_ADD_RESPONSE=$(curl -s -X POST $BASE_URL/cart/batch \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "[
    {
      \"product_id\": $PRODUCT1_ID,
      \"quantity\": 1
    },
    {
      \"product_id\": $PRODUCT2_ID,
      \"quantity\": 2
    }
  ]")

echo "批量添加响应: $BATCH_ADD_RESPONSE"
echo

# 11. 测试移除购物车商品
echo "11. 测试移除购物车商品..."
REMOVE_CART_RESPONSE=$(curl -s -X DELETE $BASE_URL/cart/$CART_ITEM_ID \
  -H "Authorization: Bearer $TOKEN")

echo "移除商品响应: $REMOVE_CART_RESPONSE"
echo

# 12. 测试按商品ID移除
echo "12. 测试按商品ID移除..."
REMOVE_PRODUCT_RESPONSE=$(curl -s -X DELETE $BASE_URL/cart/product/$PRODUCT2_ID \
  -H "Authorization: Bearer $TOKEN")

echo "按商品ID移除响应: $REMOVE_PRODUCT_RESPONSE"
echo

# 13. 查看移除后的购物车
echo "13. 查看移除后的购物车..."
GET_CART_RESPONSE3=$(curl -s -X GET $BASE_URL/cart \
  -H "Authorization: Bearer $TOKEN")

echo "移除后的购物车: $GET_CART_RESPONSE3"
echo

# 14. 测试添加不存在的商品
echo "14. 测试添加不存在的商品..."
ADD_INVALID_PRODUCT_RESPONSE=$(curl -s -X POST $BASE_URL/cart \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": 99999,
    "quantity": 1
  }')

echo "添加不存在商品响应: $ADD_INVALID_PRODUCT_RESPONSE"
echo

# 15. 测试添加超过库存的数量
echo "15. 测试添加超过库存的数量..."
ADD_EXCEED_STOCK_RESPONSE=$(curl -s -X POST $BASE_URL/cart \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"product_id\": $PRODUCT1_ID,
    \"quantity\": 999999
  }")

echo "超过库存响应: $ADD_EXCEED_STOCK_RESPONSE"
echo

# 16. 测试清空购物车
echo "16. 测试清空购物车..."
CLEAR_CART_RESPONSE=$(curl -s -X DELETE $BASE_URL/cart \
  -H "Authorization: Bearer $TOKEN")

echo "清空购物车响应: $CLEAR_CART_RESPONSE"
echo

# 17. 验证购物车已清空
echo "17. 验证购物车已清空..."
GET_EMPTY_CART_RESPONSE=$(curl -s -X GET $BASE_URL/cart \
  -H "Authorization: Bearer $TOKEN")

echo "清空后的购物车: $GET_EMPTY_CART_RESPONSE"
echo

# 18. 测试无权限访问
echo "18. 测试无权限访问..."
NO_AUTH_RESPONSE=$(curl -s -X GET $BASE_URL/cart)

echo "无权限响应: $NO_AUTH_RESPONSE"
echo

echo "=== 购物车模块API测试完成 ==="
