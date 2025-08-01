package repository

import (
	"context"

	"ryan-mall-microservices/internal/shared/domain"
	"ryan-mall-microservices/internal/user/domain/entity"
)

// UserRepository 用户仓储接口
type UserRepository interface {
	// Save 保存用户
	Save(ctx context.Context, user *entity.User) error
	
	// FindByID 根据ID查找用户
	FindByID(ctx context.Context, id domain.UserID) (*entity.User, error)
	
	// FindByUsername 根据用户名查找用户
	FindByUsername(ctx context.Context, username string) (*entity.User, error)
	
	// FindByEmail 根据邮箱查找用户
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	
	// ExistsByUsername 检查用户名是否存在
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	
	// ExistsByEmail 检查邮箱是否存在
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	
	// Update 更新用户
	Update(ctx context.Context, user *entity.User) error
	
	// Delete 删除用户
	Delete(ctx context.Context, id domain.UserID) error
	
	// List 分页查询用户列表
	List(ctx context.Context, offset, limit int) ([]*entity.User, int64, error)
}
