#!/bin/bash

# Ryan Mall 商品模块API测试脚本

BASE_URL="http://localhost:8080/api/v1"

echo "=== Ryan Mall 商品模块API测试 ==="
echo

# 首先登录获取管理员token
echo "1. 管理员登录获取Token..."
LOGIN_RESPONSE=$(curl -s -X POST $BASE_URL/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser2",
    "password": "newpassword123"
  }')

echo "登录响应: $LOGIN_RESPONSE"
TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
echo "管理员Token: $TOKEN"
echo

# 2. 测试创建分类
echo "2. 测试创建分类..."
CATEGORY1_RESPONSE=$(curl -s -X POST $BASE_URL/categories \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "智能手机",
    "parent_id": 0,
    "sort_order": 1
  }')

echo "创建分类1响应: $CATEGORY1_RESPONSE"
CATEGORY1_ID=$(echo $CATEGORY1_RESPONSE | grep -o '"id":[0-9]*' | cut -d':' -f2)
echo "分类1 ID: $CATEGORY1_ID"
echo

CATEGORY2_RESPONSE=$(curl -s -X POST $BASE_URL/categories \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "笔记本电脑",
    "parent_id": 0,
    "sort_order": 2
  }')

echo "创建分类2响应: $CATEGORY2_RESPONSE"
CATEGORY2_ID=$(echo $CATEGORY2_RESPONSE | grep -o '"id":[0-9]*' | cut -d':' -f2)
echo "分类2 ID: $CATEGORY2_ID"
echo

# 3. 测试获取分类列表
echo "3. 测试获取分类列表..."
CATEGORIES_RESPONSE=$(curl -s -X GET $BASE_URL/categories)
echo "分类列表响应: $CATEGORIES_RESPONSE"
echo

# 4. 测试创建商品
echo "4. 测试创建商品..."
PRODUCT1_RESPONSE=$(curl -s -X POST $BASE_URL/products \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"iPhone 15 Pro Max\",
    \"description\": \"苹果最新旗舰手机，性能强劲，拍照出色\",
    \"category_id\": $CATEGORY1_ID,
    \"price\": 9999.00,
    \"original_price\": 10999.00,
    \"stock\": 50,
    \"main_image\": \"https://example.com/iphone15.jpg\",
    \"images\": [\"https://example.com/iphone15-1.jpg\", \"https://example.com/iphone15-2.jpg\"]
  }")

echo "创建商品1响应: $PRODUCT1_RESPONSE"
PRODUCT1_ID=$(echo $PRODUCT1_RESPONSE | grep -o '"id":[0-9]*' | cut -d':' -f2)
echo "商品1 ID: $PRODUCT1_ID"
echo

PRODUCT2_RESPONSE=$(curl -s -X POST $BASE_URL/products \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"MacBook Pro 16\",
    \"description\": \"专业级笔记本电脑，适合开发和设计\",
    \"category_id\": $CATEGORY2_ID,
    \"price\": 19999.00,
    \"original_price\": 21999.00,
    \"stock\": 30,
    \"main_image\": \"https://example.com/macbook.jpg\",
    \"images\": [\"https://example.com/macbook-1.jpg\", \"https://example.com/macbook-2.jpg\"]
  }")

echo "创建商品2响应: $PRODUCT2_RESPONSE"
PRODUCT2_ID=$(echo $PRODUCT2_RESPONSE | grep -o '"id":[0-9]*' | cut -d':' -f2)
echo "商品2 ID: $PRODUCT2_ID"
echo

# 5. 测试获取商品列表
echo "5. 测试获取商品列表..."
PRODUCTS_RESPONSE=$(curl -s -X GET "$BASE_URL/products?page=1&page_size=10")
echo "商品列表响应: $PRODUCTS_RESPONSE"
echo

# 6. 测试获取商品详情
echo "6. 测试获取商品详情..."
PRODUCT_DETAIL_RESPONSE=$(curl -s -X GET $BASE_URL/products/$PRODUCT1_ID)
echo "商品详情响应: $PRODUCT_DETAIL_RESPONSE"
echo

# 7. 测试商品搜索
echo "7. 测试商品搜索..."
SEARCH_RESPONSE=$(curl -s -X GET "$BASE_URL/products?keyword=iPhone&page=1&page_size=5")
echo "搜索响应: $SEARCH_RESPONSE"
echo

# 8. 测试按分类获取商品
echo "8. 测试按分类获取商品..."
CATEGORY_PRODUCTS_RESPONSE=$(curl -s -X GET $BASE_URL/categories/$CATEGORY1_ID/products)
echo "分类商品响应: $CATEGORY_PRODUCTS_RESPONSE"
echo

# 9. 测试价格筛选
echo "9. 测试价格筛选..."
PRICE_FILTER_RESPONSE=$(curl -s -X GET "$BASE_URL/products?min_price=5000&max_price=15000")
echo "价格筛选响应: $PRICE_FILTER_RESPONSE"
echo

# 10. 测试更新商品
echo "10. 测试更新商品..."
UPDATE_PRODUCT_RESPONSE=$(curl -s -X PUT $BASE_URL/products/$PRODUCT1_ID \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "iPhone 15 Pro Max (更新版)",
    "price": 9599.00,
    "stock": 45
  }')

echo "更新商品响应: $UPDATE_PRODUCT_RESPONSE"
echo

# 11. 测试更新库存
echo "11. 测试更新库存..."
UPDATE_STOCK_RESPONSE=$(curl -s -X PUT $BASE_URL/products/$PRODUCT1_ID/stock \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "stock": 100
  }')

echo "更新库存响应: $UPDATE_STOCK_RESPONSE"
echo

# 12. 测试获取分类树
echo "12. 测试获取分类树..."
CATEGORY_TREE_RESPONSE=$(curl -s -X GET $BASE_URL/categories/tree)
echo "分类树响应: $CATEGORY_TREE_RESPONSE"
echo

# 13. 测试排序
echo "13. 测试按价格排序..."
SORT_RESPONSE=$(curl -s -X GET "$BASE_URL/products?sort_by=price&sort_order=asc")
echo "价格排序响应: $SORT_RESPONSE"
echo

# 14. 测试无权限操作
echo "14. 测试无权限创建商品..."
NO_AUTH_RESPONSE=$(curl -s -X POST $BASE_URL/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试商品",
    "category_id": 1,
    "price": 100.00,
    "stock": 10
  }')

echo "无权限响应: $NO_AUTH_RESPONSE"
echo

echo "=== 商品模块API测试完成 ==="
