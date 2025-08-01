#!/bin/bash

# Ryan Mall 数据初始化脚本
# 创建测试用户、分类和商品数据

set -e

echo "🚀 Ryan Mall 数据初始化脚本"
echo "================================"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# API基础URL
API_BASE="http://localhost:8080"

# 检查后端服务是否运行
check_backend() {
    echo -e "${BLUE}🔍 检查后端服务...${NC}"
    
    if ! curl -s "${API_BASE}/ping" > /dev/null; then
        echo -e "${RED}❌ 后端服务未运行，请先启动后端服务${NC}"
        echo -e "${YELLOW}启动命令: SERVER_PORT=8081 go run ./cmd/server/main.go${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✅ 后端服务运行正常${NC}"
}

# 创建用户
create_users() {
    echo -e "${BLUE}👤 创建测试用户...${NC}"
    
    # 创建管理员用户
    echo -e "${YELLOW}创建管理员用户...${NC}"
    ADMIN_RESPONSE=$(curl -s -X POST "${API_BASE}/api/v1/register" \
        -H "Content-Type: application/json" \
        -d '{"username":"admin","email":"admin@example.com","password":"admin123"}')
    
    if echo "$ADMIN_RESPONSE" | grep -q '"code":200'; then
        echo -e "${GREEN}✅ 管理员用户创建成功${NC}"
        # 提取token
        ADMIN_TOKEN=$(echo "$ADMIN_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    else
        echo -e "${YELLOW}⚠️ 管理员用户可能已存在${NC}"
        # 尝试登录获取token
        LOGIN_RESPONSE=$(curl -s -X POST "${API_BASE}/api/v1/login" \
            -H "Content-Type: application/json" \
            -d '{"username":"admin","password":"admin123"}')
        ADMIN_TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    fi
    
    # 创建普通用户
    echo -e "${YELLOW}创建普通用户...${NC}"
    USER_RESPONSE=$(curl -s -X POST "${API_BASE}/api/v1/register" \
        -H "Content-Type: application/json" \
        -d '{"username":"user1","email":"user1@example.com","password":"password123"}')
    
    if echo "$USER_RESPONSE" | grep -q '"code":200'; then
        echo -e "${GREEN}✅ 普通用户创建成功${NC}"
    else
        echo -e "${YELLOW}⚠️ 普通用户可能已存在${NC}"
    fi
    
    echo -e "${GREEN}✅ 用户创建完成${NC}"
}

# 创建分类
create_categories() {
    echo -e "${BLUE}📂 创建商品分类...${NC}"
    
    if [ -z "$ADMIN_TOKEN" ]; then
        echo -e "${RED}❌ 管理员Token为空，无法创建分类${NC}"
        return 1
    fi
    
    # 创建主要分类
    categories=(
        '{"name":"电子产品","parent_id":0,"sort_order":1,"status":1}'
        '{"name":"服装鞋帽","parent_id":0,"sort_order":2,"status":1}'
        '{"name":"家居用品","parent_id":0,"sort_order":3,"status":1}'
        '{"name":"图书音像","parent_id":0,"sort_order":4,"status":1}'
        '{"name":"运动户外","parent_id":0,"sort_order":5,"status":1}'
    )
    
    for category in "${categories[@]}"; do
        RESPONSE=$(curl -s -X POST "${API_BASE}/api/v1/categories" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer ${ADMIN_TOKEN}" \
            -d "$category")
        
        if echo "$RESPONSE" | grep -q '"code":200'; then
            CATEGORY_NAME=$(echo "$category" | grep -o '"name":"[^"]*"' | cut -d'"' -f4)
            echo -e "${GREEN}✅ 分类创建成功: ${CATEGORY_NAME}${NC}"
        else
            echo -e "${YELLOW}⚠️ 分类可能已存在或创建失败${NC}"
        fi
    done
    
    echo -e "${GREEN}✅ 分类创建完成${NC}"
}

# 创建商品
create_products() {
    echo -e "${BLUE}🛍️ 创建测试商品...${NC}"
    
    if [ -z "$ADMIN_TOKEN" ]; then
        echo -e "${RED}❌ 管理员Token为空，无法创建商品${NC}"
        return 1
    fi
    
    # 创建商品数据
    products=(
        '{"name":"iPhone 15 Pro","description":"苹果最新旗舰手机，搭载A17 Pro芯片，支持钛金属机身","category_id":1,"price":7999.00,"original_price":8999.00,"stock":50,"status":1}'
        '{"name":"MacBook Pro 14寸","description":"专业级笔记本电脑，M3 Pro芯片，32GB内存，1TB存储","category_id":1,"price":14999.00,"original_price":16999.00,"stock":30,"status":1}'
        '{"name":"iPad Air","description":"轻薄便携的平板电脑，M2芯片，10.9英寸液晶显示屏","category_id":1,"price":4399.00,"original_price":4999.00,"stock":80,"status":1}'
        '{"name":"AirPods Pro","description":"主动降噪无线耳机，空间音频，自适应透明模式","category_id":1,"price":1899.00,"original_price":2199.00,"stock":100,"status":1}'
        '{"name":"Nike Air Max","description":"经典运动鞋，舒适透气，适合日常运动和休闲","category_id":2,"price":899.00,"original_price":1299.00,"stock":60,"status":1}'
        '{"name":"Adidas T恤","description":"纯棉运动T恤，吸湿排汗，多色可选","category_id":2,"price":199.00,"original_price":299.00,"stock":120,"status":1}'
        '{"name":"智能扫地机器人","description":"全自动清扫，智能规划路径，支持APP控制","category_id":3,"price":1599.00,"original_price":1999.00,"stock":40,"status":1}'
        '{"name":"咖啡机","description":"全自动意式咖啡机，一键制作多种咖啡","category_id":3,"price":2999.00,"original_price":3499.00,"stock":25,"status":1}'
        '{"name":"编程入门指南","description":"零基础学编程，从入门到精通，配套视频教程","category_id":4,"price":89.00,"original_price":129.00,"stock":200,"status":1}'
        '{"name":"瑜伽垫","description":"环保TPE材质，防滑耐用，适合各种运动","category_id":5,"price":159.00,"original_price":199.00,"stock":80,"status":1}'
    )
    
    for product in "${products[@]}"; do
        RESPONSE=$(curl -s -X POST "${API_BASE}/api/v1/products" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer ${ADMIN_TOKEN}" \
            -d "$product")
        
        if echo "$RESPONSE" | grep -q '"code":200'; then
            PRODUCT_NAME=$(echo "$product" | grep -o '"name":"[^"]*"' | cut -d'"' -f4)
            echo -e "${GREEN}✅ 商品创建成功: ${PRODUCT_NAME}${NC}"
        else
            echo -e "${YELLOW}⚠️ 商品可能已存在或创建失败${NC}"
        fi
        
        # 添加小延迟避免请求过快
        sleep 0.1
    done
    
    echo -e "${GREEN}✅ 商品创建完成${NC}"
}

# 显示总结信息
show_summary() {
    echo ""
    echo -e "${GREEN}🎉 数据初始化完成！${NC}"
    echo "================================"
    echo -e "${BLUE}📊 创建的数据:${NC}"
    echo -e "   👤 用户: admin, user1"
    echo -e "   📂 分类: 5个主要分类"
    echo -e "   🛍️ 商品: 10个测试商品"
    echo ""
    echo -e "${BLUE}🧪 演示账户:${NC}"
    echo -e "   管理员: admin / admin123"
    echo -e "   用户: user1 / password123"
    echo ""
    echo -e "${BLUE}🌐 访问地址:${NC}"
    echo -e "   前端: http://localhost:8080"
    echo -e "   API测试: http://localhost:8080/views/test-api.html"
    echo -e "   登录页面: http://localhost:8080/views/login.html"
    echo -e "   商品页面: http://localhost:8080/views/products.html"
    echo ""
    echo -e "${YELLOW}💡 提示: 现在可以正常使用前端界面进行登录和购物了！${NC}"
}

# 主函数
main() {
    check_backend
    create_users
    create_categories
    create_products
    show_summary
}

# 显示帮助信息
show_help() {
    echo "Ryan Mall 数据初始化脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help     显示帮助信息"
    echo "  --users-only   仅创建用户"
    echo "  --reset        重置所有数据（危险操作）"
    echo ""
    echo "示例:"
    echo "  $0              # 初始化所有数据"
    echo "  $0 --users-only # 仅创建用户"
}

# 解析命令行参数
case "${1:-}" in
    -h|--help)
        show_help
        exit 0
        ;;
    --users-only)
        check_backend
        create_users
        echo -e "${GREEN}✅ 用户创建完成${NC}"
        ;;
    --reset)
        echo -e "${RED}⚠️ 重置功能暂未实现${NC}"
        exit 1
        ;;
    "")
        main
        ;;
    *)
        echo -e "${RED}❌ 未知选项: $1${NC}"
        show_help
        exit 1
        ;;
esac
