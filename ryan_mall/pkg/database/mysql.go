package database

import (
	"fmt"
	"time"

	"ryan-mall-microservices/pkg/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// MySQLConnection MySQL连接管理器
type MySQLConnection struct {
	db     *gorm.DB
	config *config.DatabaseConfig
}

// NewMySQLConnection 创建MySQL连接
func NewMySQLConnection(cfg *config.DatabaseConfig) (*MySQLConnection, error) {
	dsn := cfg.GetDSN()
	
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 获取底层的sql.DB对象来设置连接池参数
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &MySQLConnection{
		db:     db,
		config: cfg,
	}, nil
}

// GetDB 获取数据库连接
func (m *MySQLConnection) GetDB() *gorm.DB {
	return m.db
}

// Close 关闭数据库连接
func (m *MySQLConnection) Close() error {
	sqlDB, err := m.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Health 检查数据库健康状态
func (m *MySQLConnection) Health() error {
	sqlDB, err := m.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// GetStats 获取连接池统计信息
func (m *MySQLConnection) GetStats() map[string]interface{} {
	sqlDB, err := m.db.DB()
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":              stats.InUse,
		"idle":                stats.Idle,
		"wait_count":          stats.WaitCount,
		"wait_duration":       stats.WaitDuration,
		"max_idle_closed":     stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}
}

// Transaction 执行事务
func (m *MySQLConnection) Transaction(fn func(*gorm.DB) error) error {
	return m.db.Transaction(fn)
}

// AutoMigrate 自动迁移表结构
func (m *MySQLConnection) AutoMigrate(models ...interface{}) error {
	return m.db.AutoMigrate(models...)
}
