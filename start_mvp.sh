#!/bin/bash

# Ryan Mall MVP 启动脚本
# 专注于前后端产品MVP，移除了集群和监控系统

set -e

echo "🚀 Ryan Mall MVP 启动脚本"
echo "================================"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 检查Docker是否运行
check_docker() {
    echo -e "${BLUE}🔍 检查Docker服务...${NC}"
    
    if ! docker info > /dev/null 2>&1; then
        echo -e "${RED}❌ Docker未运行，请先启动Docker${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✅ Docker服务正常${NC}"
}

# 停止现有服务
stop_existing_services() {
    echo -e "${BLUE}🛑 停止现有服务...${NC}"
    
    # 停止基础服务
    docker compose down 2>/dev/null || true
    
    echo -e "${GREEN}✅ 现有服务已停止${NC}"
}

# 启动MVP核心服务
start_mvp_services() {
    echo -e "${BLUE}🚀 启动MVP核心服务...${NC}"
    
    # 启动MySQL、Redis和前端
    echo -e "${YELLOW}启动数据库和前端服务...${NC}"
    docker compose up -d mysql redis 
    #frontend
    
    # 等待MySQL启动
    echo -e "${YELLOW}等待MySQL启动...${NC}"
    sleep 10
    
    # 检查MySQL是否就绪
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if docker exec ryan-mall-mysql mysqladmin ping -h localhost --silent; then
            echo -e "${GREEN}✅ MySQL已就绪${NC}"
            break
        fi
        
        echo -e "${YELLOW}等待MySQL启动... (${attempt}/${max_attempts})${NC}"
        sleep 2
        ((attempt++))
    done
    
    if [ $attempt -gt $max_attempts ]; then
        echo -e "${RED}❌ MySQL启动超时${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✅ MVP核心服务启动成功${NC}"
}

# 显示服务状态
show_status() {
    echo ""
    echo -e "${GREEN}🎉 Ryan Mall MVP 启动完成！${NC}"
    echo "================================"
    echo -e "${BLUE}📊 服务状态:${NC}"
    
    # 检查各个服务
    if docker ps | grep -q ryan-mall-mysql; then
        echo -e "   ✅ MySQL: 运行中 (端口 3306)"
    else
        echo -e "   ❌ MySQL: 未运行"
    fi
    
    if docker ps | grep -q ryan-mall-redis; then
        echo -e "   ✅ Redis: 运行中 (端口 6379)"
    else
        echo -e "   ❌ Redis: 未运行"
    fi
    
    # if docker ps | grep -q ryan-mall-frontend; then
    #     echo -e "   ✅ 前端: 运行中 (端口 8080)"
    # else
    #     echo -e "   ❌ 前端: 未运行"
    # fi
    
    echo -e "${YELLOW}⚠️ 下一步: 启动后端API服务${NC}"
    echo -e "   启动命令: ${BLUE}go run ./cmd/server/main.go${NC}"
    echo -e "   后端端口: ${BLUE}8081${NC}"
    echo ""
    echo -e "${GREEN}🧪 演示账户:${NC}"
    echo -e "   管理员: admin / admin123"
    echo -e "   用户: user1 / password123"
    echo ""
}

# 主函数
main() {
    check_docker
    stop_existing_services
    start_mvp_services
    show_status
}

# 显示帮助信息
show_help() {
    echo "Ryan Mall MVP 启动脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help     显示帮助信息"
    echo "  --stop         停止所有服务"
    echo "  --restart      重启所有服务"
    echo "  --status       显示服务状态"
    echo ""
    echo "示例:"
    echo "  $0              # 启动服务"
    echo "  $0 --stop       # 停止所有服务"
    echo "  $0 --restart    # 重启所有服务"
}

# 停止所有服务
stop_all() {
    echo -e "${BLUE}🛑 停止所有服务...${NC}"
    docker compose down
    echo -e "${GREEN}✅ 所有服务已停止${NC}"
}

# 显示服务状态
show_service_status() {
    echo -e "${BLUE}📊 当前服务状态:${NC}"
    docker compose ps
}

# 解析命令行参数
case "${1:-}" in
    -h|--help)
        show_help
        exit 0
        ;;
    --stop)
        stop_all
        ;;
    --restart)
        stop_all
        sleep 2
        main
        ;;
    --status)
        show_service_status
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
