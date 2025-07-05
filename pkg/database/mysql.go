package database

import (
	"fmt"
	"log"
	"ryan-mall/internal/config"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 全局GORM数据库连接实例
// 使用全局变量便于在整个应用中访问数据库
var DB *gorm.DB

// InitMySQL 初始化MySQL数据库连接
// 使用GORM作为ORM框架，简化数据库操作
func InitMySQL(cfg *config.Config) error {
	// 1. 构建数据库连接字符串 (DSN - Data Source Name)
	// parseTime=true: 自动解析MySQL的时间类型为Go的time.Time
	// charset=utf8mb4: 使用utf8mb4字符集，支持emoji等4字节字符
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName,
	)

	// 2. 配置GORM日志级别
	// 根据运行模式设置不同的日志级别
	var logLevel logger.LogLevel
	if cfg.Server.Mode == "debug" {
		logLevel = logger.Info // 开发模式显示详细SQL日志
	} else {
		logLevel = logger.Error // 生产模式只显示错误日志
	}

	// 3. 打开数据库连接
	// GORM会自动处理连接池和连接管理
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		// NowFunc: 自定义时间函数，确保时区一致性
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	// 4. 获取底层的sql.DB对象来配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// 5. 配置连接池参数 - 高性能优化
	// 这些参数对于生产环境的性能很重要

	// SetMaxOpenConns: 设置最大打开连接数
	// 优化为200以减少连接池开销，针对高并发场景优化
	sqlDB.SetMaxOpenConns(200)

	// SetMaxIdleConns: 设置最大空闲连接数
	// 优化为50，平衡性能和资源使用
	sqlDB.SetMaxIdleConns(50)

	// SetConnMaxLifetime: 设置连接的最大生存时间
	// 优化为5分钟，更快的连接回收
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// SetConnMaxIdleTime: 设置连接的最大空闲时间
	// 优化为2分钟，减少资源占用
	sqlDB.SetConnMaxIdleTime(2 * time.Minute)

	// 6. 测试数据库连接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// 7. 将连接赋值给全局变量
	DB = db

	log.Println("MySQL database connected successfully with GORM")
	return nil
}

// Close 关闭数据库连接
// 应用程序退出时调用，释放数据库资源
func Close() error {
	if DB != nil {
		log.Println("Closing database connection...")
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// GetDB 获取GORM数据库连接实例
// 提供一个函数来获取数据库连接，便于测试和依赖注入
func GetDB() *gorm.DB {
	return DB
}

// AutoMigrate 自动迁移数据库表结构
// GORM可以根据结构体自动创建和更新表结构
func AutoMigrate(models ...interface{}) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	log.Println("Starting database auto migration...")

	// GORM的AutoMigrate会：
	// 1. 创建不存在的表
	// 2. 添加缺失的字段
	// 3. 添加缺失的索引
	// 注意：不会删除未使用的字段，确保数据安全
	if err := DB.AutoMigrate(models...); err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}

	log.Println("Database auto migration completed successfully")
	return nil
}

// CheckConnection 检查数据库连接状态
// 用于健康检查，确保数据库连接正常
func CheckConnection() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Ping()
}

// GetDBStats 获取数据库连接池统计信息
// 用于监控数据库连接池的使用情况
func GetDBStats() (map[string]interface{}, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %v", err)
	}

	stats := sqlDB.Stats()

	return map[string]interface{}{
		"max_open_connections":     stats.MaxOpenConnections,
		"open_connections":         stats.OpenConnections,
		"in_use":                  stats.InUse,
		"idle":                    stats.Idle,
		"wait_count":              stats.WaitCount,
		"wait_duration":           stats.WaitDuration.String(),
		"max_idle_closed":         stats.MaxIdleClosed,
		"max_idle_time_closed":    stats.MaxIdleTimeClosed,
		"max_lifetime_closed":     stats.MaxLifetimeClosed,
	}, nil
}
