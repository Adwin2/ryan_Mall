package repository

import (
	"context"
	"fmt"
	"time"

	"ryan-mall-microservices/internal/product/domain/repository"
	"ryan-mall-microservices/internal/shared/domain"
	"ryan-mall-microservices/internal/shared/infrastructure"

	"gorm.io/gorm"
)

// MySQLProductRepositoryWithLock 带分布式锁的MySQL商品仓储实现
type MySQLProductRepositoryWithLock struct {
	*MySQLProductRepository
	lockManager *infrastructure.LockManager
}

// NewMySQLProductRepositoryWithLock 创建带分布式锁的MySQL商品仓储
func NewMySQLProductRepositoryWithLock(
	db *gorm.DB,
	lockManager *infrastructure.LockManager,
) repository.ProductRepository {
	baseRepo := NewMySQLProductRepository(db).(*MySQLProductRepository)
	return &MySQLProductRepositoryWithLock{
		MySQLProductRepository: baseRepo,
		lockManager:            lockManager,
	}
}

// ReserveStock 预留库存（使用分布式锁防止超卖）
func (r *MySQLProductRepositoryWithLock) ReserveStock(ctx context.Context, id domain.ProductID, quantity int) error {
	fmt.Printf("[DEBUG] 带锁仓储的ReserveStock被调用: productID=%s, quantity=%d\n", id.String(), quantity)

	lockKey := infrastructure.StockLockKey(id.String())
	lockExpiration := 10 * time.Second

	return r.lockManager.WithLock(ctx, lockKey, lockExpiration, func() error {
		return r.reserveStockWithLock(ctx, id, quantity)
	})
}

// reserveStockWithLock 在锁保护下预留库存
func (r *MySQLProductRepositoryWithLock) reserveStockWithLock(ctx context.Context, id domain.ProductID, quantity int) error {
	// 开启事务
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 查询当前库存（加行锁）
		var currentStock int
		err := tx.Model(&ProductPO{}).
			Select("stock").
			Where("product_id = ?", id.String()).
			Set("gorm:query_option", "FOR UPDATE").
			Row().Scan(&currentStock)
		if err != nil {
			return err
		}

		// 检查库存是否足够
		if currentStock < quantity {
			return domain.NewInsufficientStockError(id.String(), quantity, currentStock)
		}

		// 扣减库存
		result := tx.Model(&ProductPO{}).
			Where("product_id = ? AND stock >= ?", id.String(), quantity).
			Update("stock", gorm.Expr("stock - ?", quantity))

		if result.Error != nil {
			return result.Error
		}

		// 再次检查是否成功扣减（防止并发问题）
		if result.RowsAffected == 0 {
			return domain.NewInsufficientStockError(id.String(), quantity, currentStock)
		}

		return nil
	})
}

// ReleaseStock 释放库存（使用分布式锁）
func (r *MySQLProductRepositoryWithLock) ReleaseStock(ctx context.Context, id domain.ProductID, quantity int) error {
	lockKey := infrastructure.StockLockKey(id.String())
	lockExpiration := 10 * time.Second

	return r.lockManager.WithLock(ctx, lockKey, lockExpiration, func() error {
		return r.MySQLProductRepository.ReleaseStock(ctx, id, quantity)
	})
}

// UpdateStock 更新库存（使用分布式锁）
func (r *MySQLProductRepositoryWithLock) UpdateStock(ctx context.Context, id domain.ProductID, quantity int) error {
	lockKey := infrastructure.StockLockKey(id.String())
	lockExpiration := 10 * time.Second

	return r.lockManager.WithLock(ctx, lockKey, lockExpiration, func() error {
		return r.MySQLProductRepository.UpdateStock(ctx, id, quantity)
	})
}

// BatchReserveStock 批量预留库存（用于订单中的多个商品）
func (r *MySQLProductRepositoryWithLock) BatchReserveStock(ctx context.Context, items []StockReservationItem) error {
	// 按商品ID排序，避免死锁
	sortedItems := make([]StockReservationItem, len(items))
	copy(sortedItems, items)
	
	// 简单排序（实际项目中可以使用更高效的排序算法）
	for i := 0; i < len(sortedItems)-1; i++ {
		for j := i + 1; j < len(sortedItems); j++ {
			if sortedItems[i].ProductID > sortedItems[j].ProductID {
				sortedItems[i], sortedItems[j] = sortedItems[j], sortedItems[i]
			}
		}
	}

	// 开启事务
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range sortedItems {
			lockKey := infrastructure.StockLockKey(item.ProductID)
			lockExpiration := 30 * time.Second

			err := r.lockManager.WithLock(ctx, lockKey, lockExpiration, func() error {
				return r.reserveStockWithLock(ctx, domain.ProductID(item.ProductID), item.Quantity)
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// StockReservationItem 库存预留项
type StockReservationItem struct {
	ProductID string
	Quantity  int
}

// OptimisticLockProductRepository 乐观锁商品仓储实现
type OptimisticLockProductRepository struct {
	*MySQLProductRepository
}

// NewOptimisticLockProductRepository 创建乐观锁商品仓储
func NewOptimisticLockProductRepository(db *gorm.DB) repository.ProductRepository {
	baseRepo := NewMySQLProductRepository(db).(*MySQLProductRepository)
	return &OptimisticLockProductRepository{
		MySQLProductRepository: baseRepo,
	}
}

// ReserveStock 使用乐观锁预留库存
func (r *OptimisticLockProductRepository) ReserveStock(ctx context.Context, id domain.ProductID, quantity int) error {
	maxRetries := 3
	
	for i := 0; i < maxRetries; i++ {
		// 查询当前商品信息（包括版本号）
		var po ProductPO
		err := r.db.WithContext(ctx).Where("product_id = ?", id.String()).First(&po).Error
		if err != nil {
			return err
		}

		// 检查库存
		if po.Stock < quantity {
			return domain.NewInsufficientStockError(id.String(), quantity, po.Stock)
		}

		// 使用乐观锁更新（基于版本号或更新时间）
		result := r.db.WithContext(ctx).Model(&ProductPO{}).
			Where("product_id = ? AND updated_at = ? AND stock >= ?", 
				id.String(), po.UpdatedAt, quantity).
			Updates(map[string]interface{}{
				"stock":      gorm.Expr("stock - ?", quantity),
				"updated_at": time.Now(),
			})

		if result.Error != nil {
			return result.Error
		}

		// 如果更新成功，返回
		if result.RowsAffected > 0 {
			return nil
		}

		// 如果更新失败（版本冲突），重试
		if i == maxRetries-1 {
			return fmt.Errorf("failed to reserve stock after %d retries due to concurrent updates", maxRetries)
		}

		// 短暂等待后重试
		time.Sleep(time.Millisecond * 10)
	}

	return nil
}

// PessimisticLockProductRepository 悲观锁商品仓储实现
type PessimisticLockProductRepository struct {
	*MySQLProductRepository
}

// NewPessimisticLockProductRepository 创建悲观锁商品仓储
func NewPessimisticLockProductRepository(db *gorm.DB) repository.ProductRepository {
	baseRepo := NewMySQLProductRepository(db).(*MySQLProductRepository)
	return &PessimisticLockProductRepository{
		MySQLProductRepository: baseRepo,
	}
}

// ReserveStock 使用悲观锁预留库存
func (r *PessimisticLockProductRepository) ReserveStock(ctx context.Context, id domain.ProductID, quantity int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 使用SELECT FOR UPDATE加行锁
		var po ProductPO
		err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("product_id = ?", id.String()).
			First(&po).Error
		if err != nil {
			return err
		}

		// 检查库存
		if po.Stock < quantity {
			return domain.NewInsufficientStockError(id.String(), quantity, po.Stock)
		}

		// 更新库存
		return tx.Model(&po).Update("stock", gorm.Expr("stock - ?", quantity)).Error
	})
}
