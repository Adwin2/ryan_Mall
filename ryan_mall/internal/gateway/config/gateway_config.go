package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// GatewayConfig API网关配置
type GatewayConfig struct {
	Port        int           `json:"port"`
	Environment string        `json:"environment"`
	LogLevel    string        `json:"log_level"`
	JWT         JWTConfig     `json:"jwt"`
	Redis       RedisConfig   `json:"redis"`
	Etcd        EtcdConfig    `json:"etcd"`
	Services    ServicesConfig `json:"services"`
	RateLimit   RateLimitConfig `json:"rate_limit"`
	Timeout     TimeoutConfig `json:"timeout"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret             string `json:"secret"`
	ExpiryHours        int    `json:"expiry_hours"`
	RefreshExpiryHours int    `json:"refresh_expiry_hours"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Address  string `json:"address"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

// EtcdConfig Etcd配置
type EtcdConfig struct {
	Endpoints []string `json:"endpoints"`
	KeyPrefix string   `json:"key_prefix"`
	TTL       int64    `json:"ttl"`
}

// ServicesConfig 服务配置
type ServicesConfig struct {
	UserService    ServiceConfig `json:"user_service"`
	ProductService ServiceConfig `json:"product_service"`
	OrderService   ServiceConfig `json:"order_service"`
	SeckillService ServiceConfig `json:"seckill_service"`
	PaymentService ServiceConfig `json:"payment_service"`
}

// ServiceConfig 单个服务配置
type ServiceConfig struct {
	Name         string        `json:"name"`
	PathPrefix   string        `json:"path_prefix"`
	Timeout      time.Duration `json:"timeout"`
	RetryCount   int           `json:"retry_count"`
	CircuitBreaker CircuitBreakerConfig `json:"circuit_breaker"`
}

// CircuitBreakerConfig 熔断器配置
type CircuitBreakerConfig struct {
	Enabled           bool          `json:"enabled"`
	FailureThreshold  int           `json:"failure_threshold"`
	RecoveryTimeout   time.Duration `json:"recovery_timeout"`
	SuccessThreshold  int           `json:"success_threshold"`
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Enabled     bool `json:"enabled"`
	GlobalLimit int  `json:"global_limit"`
	UserLimit   int  `json:"user_limit"`
	IPLimit     int  `json:"ip_limit"`
}

// TimeoutConfig 超时配置
type TimeoutConfig struct {
	Read  time.Duration `json:"read"`
	Write time.Duration `json:"write"`
	Idle  time.Duration `json:"idle"`
}

// LoadGatewayConfig 加载网关配置
func LoadGatewayConfig() (*GatewayConfig, error) {
	config := &GatewayConfig{
		Port:        getEnvAsInt("GATEWAY_PORT", 8080),
		Environment: getEnv("ENVIRONMENT", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		JWT: JWTConfig{
			Secret:             getEnv("JWT_SECRET", "your-secret-key"),
			ExpiryHours:        getEnvAsInt("JWT_EXPIRY_HOURS", 24),
			RefreshExpiryHours: getEnvAsInt("JWT_REFRESH_EXPIRY_HOURS", 168), // 7 days
		},
		Redis: RedisConfig{
			Address:  getEnv("REDIS_ADDRESS", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		Etcd: EtcdConfig{
			Endpoints: getEnvAsSlice("ETCD_ENDPOINTS", []string{"localhost:2379"}),
			KeyPrefix: getEnv("ETCD_KEY_PREFIX", "/ryan-mall"),
			TTL:       int64(getEnvAsInt("ETCD_TTL", 30)),
		},
		Services: ServicesConfig{
			UserService: ServiceConfig{
				Name:       "user-service",
				PathPrefix: "/api/v1/users",
				Timeout:    time.Duration(getEnvAsInt("USER_SERVICE_TIMEOUT", 30)) * time.Second,
				RetryCount: getEnvAsInt("USER_SERVICE_RETRY", 3),
				CircuitBreaker: CircuitBreakerConfig{
					Enabled:          getEnvAsBool("USER_SERVICE_CB_ENABLED", true),
					FailureThreshold: getEnvAsInt("USER_SERVICE_CB_FAILURE_THRESHOLD", 5),
					RecoveryTimeout:  time.Duration(getEnvAsInt("USER_SERVICE_CB_RECOVERY_TIMEOUT", 60)) * time.Second,
					SuccessThreshold: getEnvAsInt("USER_SERVICE_CB_SUCCESS_THRESHOLD", 3),
				},
			},
			ProductService: ServiceConfig{
				Name:       "product-service",
				PathPrefix: "/api/v1/products",
				Timeout:    time.Duration(getEnvAsInt("PRODUCT_SERVICE_TIMEOUT", 30)) * time.Second,
				RetryCount: getEnvAsInt("PRODUCT_SERVICE_RETRY", 3),
				CircuitBreaker: CircuitBreakerConfig{
					Enabled:          getEnvAsBool("PRODUCT_SERVICE_CB_ENABLED", true),
					FailureThreshold: getEnvAsInt("PRODUCT_SERVICE_CB_FAILURE_THRESHOLD", 5),
					RecoveryTimeout:  time.Duration(getEnvAsInt("PRODUCT_SERVICE_CB_RECOVERY_TIMEOUT", 60)) * time.Second,
					SuccessThreshold: getEnvAsInt("PRODUCT_SERVICE_CB_SUCCESS_THRESHOLD", 3),
				},
			},
			OrderService: ServiceConfig{
				Name:       "order-service",
				PathPrefix: "/api/v1/orders",
				Timeout:    time.Duration(getEnvAsInt("ORDER_SERVICE_TIMEOUT", 30)) * time.Second,
				RetryCount: getEnvAsInt("ORDER_SERVICE_RETRY", 3),
				CircuitBreaker: CircuitBreakerConfig{
					Enabled:          getEnvAsBool("ORDER_SERVICE_CB_ENABLED", true),
					FailureThreshold: getEnvAsInt("ORDER_SERVICE_CB_FAILURE_THRESHOLD", 5),
					RecoveryTimeout:  time.Duration(getEnvAsInt("ORDER_SERVICE_CB_RECOVERY_TIMEOUT", 60)) * time.Second,
					SuccessThreshold: getEnvAsInt("ORDER_SERVICE_CB_SUCCESS_THRESHOLD", 3),
				},
			},
			SeckillService: ServiceConfig{
				Name:       "seckill-service",
				PathPrefix: "/api/v1/seckill",
				Timeout:    time.Duration(getEnvAsInt("SECKILL_SERVICE_TIMEOUT", 10)) * time.Second,
				RetryCount: getEnvAsInt("SECKILL_SERVICE_RETRY", 1), // 秒杀服务重试次数较少
				CircuitBreaker: CircuitBreakerConfig{
					Enabled:          getEnvAsBool("SECKILL_SERVICE_CB_ENABLED", true),
					FailureThreshold: getEnvAsInt("SECKILL_SERVICE_CB_FAILURE_THRESHOLD", 3),
					RecoveryTimeout:  time.Duration(getEnvAsInt("SECKILL_SERVICE_CB_RECOVERY_TIMEOUT", 30)) * time.Second,
					SuccessThreshold: getEnvAsInt("SECKILL_SERVICE_CB_SUCCESS_THRESHOLD", 2),
				},
			},
			PaymentService: ServiceConfig{
				Name:       "payment-service",
				PathPrefix: "/api/v1/payments",
				Timeout:    time.Duration(getEnvAsInt("PAYMENT_SERVICE_TIMEOUT", 60)) * time.Second,
				RetryCount: getEnvAsInt("PAYMENT_SERVICE_RETRY", 2),
				CircuitBreaker: CircuitBreakerConfig{
					Enabled:          getEnvAsBool("PAYMENT_SERVICE_CB_ENABLED", true),
					FailureThreshold: getEnvAsInt("PAYMENT_SERVICE_CB_FAILURE_THRESHOLD", 3),
					RecoveryTimeout:  time.Duration(getEnvAsInt("PAYMENT_SERVICE_CB_RECOVERY_TIMEOUT", 120)) * time.Second,
					SuccessThreshold: getEnvAsInt("PAYMENT_SERVICE_CB_SUCCESS_THRESHOLD", 2),
				},
			},
		},
		RateLimit: RateLimitConfig{
			Enabled:     getEnvAsBool("RATE_LIMIT_ENABLED", true),
			GlobalLimit: getEnvAsInt("RATE_LIMIT_GLOBAL", 10000),
			UserLimit:   getEnvAsInt("RATE_LIMIT_USER", 100),
			IPLimit:     getEnvAsInt("RATE_LIMIT_IP", 1000),
		},
		Timeout: TimeoutConfig{
			Read:  time.Duration(getEnvAsInt("TIMEOUT_READ", 30)) * time.Second,
			Write: time.Duration(getEnvAsInt("TIMEOUT_WRITE", 30)) * time.Second,
			Idle:  time.Duration(getEnvAsInt("TIMEOUT_IDLE", 120)) * time.Second,
		},
	}

	return config, nil
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt 获取环境变量并转换为int
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsBool 获取环境变量并转换为bool
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// getEnvAsSlice 获取环境变量并转换为字符串切片
func getEnvAsSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}
