package repository

import (
	"errors"
	"ryan-mall/internal/model"

	"gorm.io/gorm"
)

// UserRepository 用户数据访问层接口
// 定义用户相关的数据库操作方法
type UserRepository interface {
	Create(user *model.User) error                    // 创建用户
	GetByID(id uint) (*model.User, error)            // 根据ID获取用户
	GetByUsername(username string) (*model.User, error) // 根据用户名获取用户
	GetByEmail(email string) (*model.User, error)    // 根据邮箱获取用户
	Update(user *model.User) error                    // 更新用户信息
	Delete(id uint) error                             // 删除用户（软删除）
	ExistsByUsername(username string) (bool, error)  // 检查用户名是否存在
	ExistsByEmail(email string) (bool, error)        // 检查邮箱是否存在
}

// userRepository 用户数据访问层实现
type userRepository struct {
	db *gorm.DB // GORM数据库连接
}

// NewUserRepository 创建用户数据访问层实例
// 使用依赖注入的方式传入数据库连接
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

// Create 创建用户
// 在数据库中插入新用户记录
func (r *userRepository) Create(user *model.User) error {
	// GORM的Create方法会自动设置CreatedAt和UpdatedAt
	// 同时会将生成的ID赋值给user.ID
	if err := r.db.Create(user).Error; err != nil {
		return err
	}
	return nil
}

// GetByID 根据ID获取用户
// 使用GORM的First方法查询单条记录
func (r *userRepository) GetByID(id uint) (*model.User, error) {
	var user model.User
	
	// First方法会查询第一条匹配的记录
	// 如果没有找到记录，会返回gorm.ErrRecordNotFound错误
	err := r.db.First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 用户不存在，返回nil而不是错误
		}
		return nil, err
	}
	
	return &user, nil
}

// GetByUsername 根据用户名获取用户
// 使用Where条件查询
func (r *userRepository) GetByUsername(username string) (*model.User, error) {
	var user model.User
	
	// Where方法添加查询条件
	// First方法执行查询并获取第一条记录
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 用户不存在
		}
		return nil, err
	}
	
	return &user, nil
}

// GetByEmail 根据邮箱获取用户
func (r *userRepository) GetByEmail(email string) (*model.User, error) {
	var user model.User
	
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 用户不存在
		}
		return nil, err
	}
	
	return &user, nil
}

// Update 更新用户信息
// 使用GORM的Save方法更新记录
func (r *userRepository) Update(user *model.User) error {
	// Save方法会更新所有字段
	// GORM会自动更新UpdatedAt字段
	return r.db.Save(user).Error
}

// Delete 删除用户（软删除）
// GORM的Delete方法会执行软删除，设置deleted_at字段
func (r *userRepository) Delete(id uint) error {
	// 软删除：只是设置deleted_at字段，不会真正删除记录
	// 这样可以保留数据用于审计和恢复
	return r.db.Delete(&model.User{}, id).Error
}

// ExistsByUsername 检查用户名是否存在
// 使用Count方法统计匹配的记录数
func (r *userRepository) ExistsByUsername(username string) (bool, error) {
	var count int64
	
	// Count方法统计匹配条件的记录数
	// 只查询ID字段以提高性能
	err := r.db.Model(&model.User{}).Where("username = ?", username).Count(&count).Error
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}

// ExistsByEmail 检查邮箱是否存在
func (r *userRepository) ExistsByEmail(email string) (bool, error) {
	var count int64
	
	err := r.db.Model(&model.User{}).Where("email = ?", email).Count(&count).Error
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}
