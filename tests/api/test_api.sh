#!/bin/bash

# Ryan Mall API 测试脚本
# 用于测试用户相关的API接口

BASE_URL="http://localhost:8080/api/v1"

echo "=== Ryan Mall API 测试 ==="
echo

# 1. 测试用户注册
echo "1. 测试用户注册..."
REGISTER_RESPONSE=$(curl -s -X POST $BASE_URL/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser2",
    "email": "test2@example.com",
    "password": "123456",
    "phone": "13800138001"
  }')

echo "注册响应: $REGISTER_RESPONSE"
echo

# 从注册响应中提取token
TOKEN=$(echo $REGISTER_RESPONSE | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
echo "提取的Token: $TOKEN"
echo

# 2. 测试用户登录
echo "2. 测试用户登录..."
LOGIN_RESPONSE=$(curl -s -X POST $BASE_URL/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser2",
    "password": "123456"
  }')

echo "登录响应: $LOGIN_RESPONSE"
echo

# 从登录响应中提取token
TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
echo "登录Token: $TOKEN"
echo

# 3. 测试获取用户资料
echo "3. 测试获取用户资料..."
PROFILE_RESPONSE=$(curl -s -X GET $BASE_URL/profile \
  -H "Authorization: Bearer $TOKEN")

echo "用户资料响应: $PROFILE_RESPONSE"
echo

# 4. 测试更新用户资料
echo "4. 测试更新用户资料..."
UPDATE_RESPONSE=$(curl -s -X PUT $BASE_URL/profile \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "13900139000",
    "avatar": "https://example.com/avatar.jpg"
  }')

echo "更新资料响应: $UPDATE_RESPONSE"
echo

# 5. 再次获取用户资料验证更新
echo "5. 验证资料更新..."
PROFILE_RESPONSE2=$(curl -s -X GET $BASE_URL/profile \
  -H "Authorization: Bearer $TOKEN")

echo "更新后的用户资料: $PROFILE_RESPONSE2"
echo

# 6. 测试修改密码
echo "6. 测试修改密码..."
CHANGE_PASSWORD_RESPONSE=$(curl -s -X POST $BASE_URL/change-password \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "old_password": "123456",
    "new_password": "newpassword123"
  }')

echo "修改密码响应: $CHANGE_PASSWORD_RESPONSE"
echo

# 7. 测试用新密码登录
echo "7. 测试用新密码登录..."
NEW_LOGIN_RESPONSE=$(curl -s -X POST $BASE_URL/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser2",
    "password": "newpassword123"
  }')

echo "新密码登录响应: $NEW_LOGIN_RESPONSE"
echo

# 8. 测试无效令牌
echo "8. 测试无效令牌..."
INVALID_TOKEN_RESPONSE=$(curl -s -X GET $BASE_URL/profile \
  -H "Authorization: Bearer invalid_token")

echo "无效令牌响应: $INVALID_TOKEN_RESPONSE"
echo

# 9. 测试缺少令牌
echo "9. 测试缺少令牌..."
NO_TOKEN_RESPONSE=$(curl -s -X GET $BASE_URL/profile)

echo "缺少令牌响应: $NO_TOKEN_RESPONSE"
echo

# 10. 测试重复注册
echo "10. 测试重复注册..."
DUPLICATE_RESPONSE=$(curl -s -X POST $BASE_URL/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser2",
    "email": "test3@example.com",
    "password": "123456"
  }')

echo "重复用户名注册响应: $DUPLICATE_RESPONSE"
echo

echo "=== API 测试完成 ==="
