#!/bin/bash

# Ryan Mall 数据库设置脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查MySQL是否运行
check_mysql() {
    if ! command -v mysql &> /dev/null; then
        log_error "MySQL客户端未安装"
        return 1
    fi
    
    if ! netstat -tln | grep -q :3306; then
        log_error "MySQL服务未运行"
        return 1
    fi
    
    return 0
}

# 尝试连接MySQL
connect_mysql() {
    local user=$1
    local password=$2
    local command=$3
    
    if [ -z "$password" ]; then
        mysql -h localhost -u "$user" -e "$command" 2>/dev/null
    else
        mysql -h localhost -u "$user" -p"$password" -e "$command" 2>/dev/null
    fi
}

# 创建数据库和用户
setup_database() {
    log_info "开始设置数据库..."
    
    # 尝试不同的连接方式
    local mysql_cmd=""
    
    # 方式1: 尝试sudo mysql
    if sudo mysql -u root -e "SELECT 1;" &>/dev/null; then
        mysql_cmd="sudo mysql -u root"
        log_info "使用sudo连接MySQL"
    # 方式2: 尝试无密码root
    elif mysql -h localhost -u root -e "SELECT 1;" &>/dev/null; then
        mysql_cmd="mysql -h localhost -u root"
        log_info "使用无密码root连接MySQL"
    # 方式3: 尝试常见密码
    elif mysql -h localhost -u root -proot -e "SELECT 1;" &>/dev/null; then
        mysql_cmd="mysql -h localhost -u root -proot"
        log_info "使用密码'root'连接MySQL"
    elif mysql -h localhost -u root -proot123 -e "SELECT 1;" &>/dev/null; then
        mysql_cmd="mysql -h localhost -u root -proot123"
        log_info "使用密码'root123'连接MySQL"
    elif mysql -h localhost -u root -p123456 -e "SELECT 1;" &>/dev/null; then
        mysql_cmd="mysql -h localhost -u root -p123456"
        log_info "使用密码'123456'连接MySQL"
    else
        log_error "无法连接到MySQL，请手动设置数据库"
        log_info "请执行以下SQL语句："
        echo ""
        cat << 'EOF'
-- 创建数据库
CREATE DATABASE IF NOT EXISTS ryan_mall CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 创建用户表
USE ryan_mall;
CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id VARCHAR(36) UNIQUE NOT NULL COMMENT '用户UUID',
    username VARCHAR(50) UNIQUE NOT NULL COMMENT '用户名',
    email VARCHAR(100) UNIQUE NOT NULL COMMENT '邮箱',
    password_hash VARCHAR(255) NOT NULL COMMENT '密码哈希',
    phone VARCHAR(20) COMMENT '手机号',
    status TINYINT DEFAULT 1 COMMENT '状态：1-正常，0-禁用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_username (username),
    INDEX idx_email (email),
    INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 创建用户档案表
CREATE TABLE IF NOT EXISTS user_profiles (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id VARCHAR(36) NOT NULL COMMENT '用户UUID',
    nickname VARCHAR(50) COMMENT '昵称',
    avatar_url VARCHAR(255) COMMENT '头像URL',
    gender TINYINT COMMENT '性别：1-男，2-女，0-未知',
    birthday DATE COMMENT '生日',
    bio TEXT COMMENT '个人简介',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户档案表';
EOF
        echo ""
        return 1
    fi
    
    # 执行数据库设置
    log_info "创建数据库和表..."
    
    $mysql_cmd << 'EOF'
-- 创建数据库
CREATE DATABASE IF NOT EXISTS ryan_mall CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 使用数据库
USE ryan_mall;

-- 创建用户表
CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id VARCHAR(36) UNIQUE NOT NULL COMMENT '用户UUID',
    username VARCHAR(50) UNIQUE NOT NULL COMMENT '用户名',
    email VARCHAR(100) UNIQUE NOT NULL COMMENT '邮箱',
    password_hash VARCHAR(255) NOT NULL COMMENT '密码哈希',
    phone VARCHAR(20) COMMENT '手机号',
    status TINYINT DEFAULT 1 COMMENT '状态：1-正常，0-禁用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_username (username),
    INDEX idx_email (email),
    INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 创建用户档案表
CREATE TABLE IF NOT EXISTS user_profiles (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id VARCHAR(36) NOT NULL COMMENT '用户UUID',
    nickname VARCHAR(50) COMMENT '昵称',
    avatar_url VARCHAR(255) COMMENT '头像URL',
    gender TINYINT COMMENT '性别：1-男，2-女，0-未知',
    birthday DATE COMMENT '生日',
    bio TEXT COMMENT '个人简介',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户档案表';
EOF

    if [ $? -eq 0 ]; then
        log_success "数据库设置完成"
        return 0
    else
        log_error "数据库设置失败"
        return 1
    fi
}

# 测试数据库连接
test_connection() {
    log_info "测试数据库连接..."
    
    # 从环境变量读取配置
    source .env 2>/dev/null || true
    
    local db_host=${DB_HOST:-localhost}
    local db_port=${DB_PORT:-3306}
    local db_user=${DB_USER:-root}
    local db_password=${DB_PASSWORD:-root123}
    local db_name=${DB_NAME:-ryan_mall}
    
    if mysql -h "$db_host" -P "$db_port" -u "$db_user" -p"$db_password" -e "USE $db_name; SHOW TABLES;" &>/dev/null; then
        log_success "数据库连接测试成功"
        log_info "数据库配置："
        echo "  主机: $db_host:$db_port"
        echo "  用户: $db_user"
        echo "  数据库: $db_name"
        return 0
    else
        log_error "数据库连接测试失败"
        log_warning "请检查.env文件中的数据库配置"
        return 1
    fi
}

# 显示帮助信息
show_help() {
    echo "Ryan Mall 数据库设置脚本"
    echo ""
    echo "用法: $0 [命令]"
    echo ""
    echo "命令:"
    echo "  setup     设置数据库和表"
    echo "  test      测试数据库连接"
    echo "  help      显示帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 setup     # 设置数据库"
    echo "  $0 test      # 测试连接"
}

# 主函数
main() {
    case "${1:-setup}" in
        "setup")
            if ! check_mysql; then
                exit 1
            fi
            setup_database
            ;;
        "test")
            test_connection
            ;;
        "help"|*)
            show_help
            ;;
    esac
}

# 执行主函数
main "$@"
