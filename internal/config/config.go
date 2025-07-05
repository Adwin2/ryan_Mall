package config

import (
	"log"
	"os"
	"strconv"
)

// Config 应用程序配置结构体
// 这里集中管理所有的配置项，方便维护和修改
type Config struct {
	// 服务器配置
	Server ServerConfig
	// 数据库配置
	Database DatabaseConfig
	// Redis配置
	Redis RedisConfig
	// JWT配置
	JWT JWTConfig
}

// ServerConfig 服务器相关配置
type ServerConfig struct {
	Port string // 服务器端口
	Mode string // 运行模式：debug, release, test
}

// DatabaseConfig 数据库相关配置
type DatabaseConfig struct {
	Host     string // 数据库主机地址
	Port     string // 数据库端口
	Username string // 数据库用户名
	Password string // 数据库密码
	DBName   string // 数据库名称
}

// RedisConfig Redis相关配置
type RedisConfig struct {
	// 单机模式配置
	Host     string // Redis主机地址
	Port     string // Redis端口
	Password string // Redis密码
	DB       int    // Redis数据库编号

	// 集群模式配置
	ClusterEnabled bool     // 是否启用集群模式
	ClusterNodes   []string // 集群节点地址列表
}

// JWTConfig JWT相关配置
type JWTConfig struct {
	SecretKey string // JWT签名密钥
	ExpireHours int  // Token过期时间（小时）
}

// LoadConfig 加载配置
// 这个函数从环境变量中读取配置，如果没有设置则使用默认值
// 在生产环境中，建议通过环境变量来配置这些敏感信息
func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "3306"),
			Username: getEnv("DB_USERNAME", "ryan_mall"),
			Password: getEnv("DB_PASSWORD", "RyanMall123!"),
			DBName:   getEnv("DB_NAME", "ryan_mall"),
		},
		Redis: RedisConfig{
			Host:           getEnv("REDIS_HOST", "localhost"),
			Port:           getEnv("REDIS_PORT", "6379"),
			Password:       getEnv("REDIS_PASSWORD", ""),
			DB:             getEnvAsInt("REDIS_DB", 0),
			ClusterEnabled: getEnvAsBool("REDIS_CLUSTER_ENABLED", false),
			ClusterNodes:   getEnvAsStringSlice("REDIS_CLUSTER_NODES", []string{}),
		},
		JWT: JWTConfig{
			SecretKey:   getEnv("JWT_SECRET", "ryan-mall-secret-key"),
			ExpireHours: getEnvAsInt("JWT_EXPIRE_HOURS", 24),
		},
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt 获取环境变量并转换为整数，如果不存在或转换失败则返回默认值
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		log.Printf("Warning: Invalid integer value for %s: %s, using default: %d", key, value, defaultValue)
	}
	return defaultValue
}

// getEnvAsBool 获取环境变量并转换为布尔值，如果不存在或转换失败则返回默认值
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
		log.Printf("Warning: Invalid boolean value for %s: %s, using default: %t", key, value, defaultValue)
	}
	return defaultValue
}

// getEnvAsStringSlice 获取环境变量并转换为字符串切片（逗号分隔），如果不存在则返回默认值
func getEnvAsStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// 简单的字符串分割实现
		var parts []string
		current := ""
		for _, char := range value {
			if char == ',' {
				if current != "" {
					parts = append(parts, current)
					current = ""
				}
			} else if char != ' ' && char != '\t' {
				current += string(char)
			}
		}
		if current != "" {
			parts = append(parts, current)
		}
		if len(parts) > 0 {
			return parts
		}
	}
	return defaultValue
}
