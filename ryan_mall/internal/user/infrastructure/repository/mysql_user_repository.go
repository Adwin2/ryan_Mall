package repository

import (
	"context"
	"errors"
	"time"

	"ryan-mall-microservices/internal/shared/domain"
	"ryan-mall-microservices/internal/user/domain/entity"
	"ryan-mall-microservices/internal/user/domain/repository"
	"ryan-mall-microservices/internal/user/domain/valueobject"

	"gorm.io/gorm"
)

// UserPO 用户持久化对象
type UserPO struct {
	ID           uint      `gorm:"primaryKey;autoIncrement"`
	UserID       string    `gorm:"uniqueIndex;size:36;not null"`
	Username     string    `gorm:"uniqueIndex;size:50;not null"`
	Email        string    `gorm:"uniqueIndex;size:100;not null"`
	PasswordHash string    `gorm:"size:255;not null"`
	Phone        string    `gorm:"size:20"`
	Status       int       `gorm:"default:1"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

// TableName 表名
func (UserPO) TableName() string {
	return "users"
}

// UserProfilePO 用户档案持久化对象
type UserProfilePO struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	UserID    string    `gorm:"uniqueIndex;size:36;not null"`
	Nickname  string    `gorm:"size:50"`
	AvatarURL string    `gorm:"size:255"`
	Gender    int       `gorm:"default:0"`
	Bio       string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// TableName 表名
func (UserProfilePO) TableName() string {
	return "user_profiles"
}

// MySQLUserRepository MySQL用户仓储实现
type MySQLUserRepository struct {
	db *gorm.DB
}

// NewMySQLUserRepository 创建MySQL用户仓储
func NewMySQLUserRepository(db *gorm.DB) repository.UserRepository {
	return &MySQLUserRepository{db: db}
}

// Save 保存用户
func (r *MySQLUserRepository) Save(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 保存用户基本信息
		userPO := r.entityToPO(user)
		if err := tx.Create(&userPO).Error; err != nil {
			return err
		}

		// 保存用户档案
		profilePO := r.profileToPO(user.ID().String(), user.Profile())
		if err := tx.Create(&profilePO).Error; err != nil {
			return err
		}

		return nil
	})
}

// FindByID 根据ID查找用户
func (r *MySQLUserRepository) FindByID(ctx context.Context, id domain.UserID) (*entity.User, error) {
	var userPO UserPO
	var profilePO UserProfilePO

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 查找用户基本信息
		if err := tx.Where("user_id = ?", id.String()).First(&userPO).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil // 用户不存在
			}
			return err
		}

		// 查找用户档案
		if err := tx.Where("user_id = ?", id.String()).First(&profilePO).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			// 档案不存在时创建默认档案
			profilePO = UserProfilePO{
				UserID: id.String(),
				Gender: 0,
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 如果用户不存在
	if userPO.UserID == "" {
		return nil, nil
	}

	return r.poToEntity(&userPO, &profilePO)
}

// FindByUsername 根据用户名查找用户
func (r *MySQLUserRepository) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	var userPO UserPO
	var profilePO UserProfilePO

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 查找用户基本信息
		if err := tx.Where("username = ?", username).First(&userPO).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil // 用户不存在
			}
			return err
		}

		// 查找用户档案
		if err := tx.Where("user_id = ?", userPO.UserID).First(&profilePO).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			// 档案不存在时创建默认档案
			profilePO = UserProfilePO{
				UserID: userPO.UserID,
				Gender: 0,
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 如果用户不存在
	if userPO.UserID == "" {
		return nil, nil
	}

	return r.poToEntity(&userPO, &profilePO)
}

// FindByEmail 根据邮箱查找用户
func (r *MySQLUserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var userPO UserPO
	var profilePO UserProfilePO

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 查找用户基本信息
		if err := tx.Where("email = ?", email).First(&userPO).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil // 用户不存在
			}
			return err
		}

		// 查找用户档案
		if err := tx.Where("user_id = ?", userPO.UserID).First(&profilePO).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			// 档案不存在时创建默认档案
			profilePO = UserProfilePO{
				UserID: userPO.UserID,
				Gender: 0,
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 如果用户不存在
	if userPO.UserID == "" {
		return nil, nil
	}

	return r.poToEntity(&userPO, &profilePO)
}

// ExistsByUsername 检查用户名是否存在
func (r *MySQLUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&UserPO{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}

// ExistsByEmail 检查邮箱是否存在
func (r *MySQLUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&UserPO{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

// Update 更新用户
func (r *MySQLUserRepository) Update(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 更新用户基本信息
		userPO := r.entityToPO(user)
		if err := tx.Model(&UserPO{}).Where("user_id = ?", user.ID().String()).Updates(&userPO).Error; err != nil {
			return err
		}

		// 更新用户档案
		profilePO := r.profileToPO(user.ID().String(), user.Profile())
		if err := tx.Model(&UserProfilePO{}).Where("user_id = ?", user.ID().String()).Updates(&profilePO).Error; err != nil {
			return err
		}

		return nil
	})
}

// Delete 删除用户
func (r *MySQLUserRepository) Delete(ctx context.Context, id domain.UserID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除用户档案
		if err := tx.Where("user_id = ?", id.String()).Delete(&UserProfilePO{}).Error; err != nil {
			return err
		}

		// 删除用户基本信息
		if err := tx.Where("user_id = ?", id.String()).Delete(&UserPO{}).Error; err != nil {
			return err
		}

		return nil
	})
}

// List 分页查询用户列表
func (r *MySQLUserRepository) List(ctx context.Context, offset, limit int) ([]*entity.User, int64, error) {
	var userPOs []UserPO
	var total int64

	// 查询总数
	if err := r.db.WithContext(ctx).Model(&UserPO{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询用户列表
	if err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&userPOs).Error; err != nil {
		return nil, 0, err
	}

	// 转换为实体
	users := make([]*entity.User, len(userPOs))
	for i, userPO := range userPOs {
		// 查找对应的用户档案
		var profilePO UserProfilePO
		if err := r.db.WithContext(ctx).Where("user_id = ?", userPO.UserID).First(&profilePO).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, 0, err
			}
			// 档案不存在时创建默认档案
			profilePO = UserProfilePO{
				UserID: userPO.UserID,
				Gender: 0,
			}
		}

		user, err := r.poToEntity(&userPO, &profilePO)
		if err != nil {
			return nil, 0, err
		}
		users[i] = user
	}

	return users, total, nil
}

// entityToPO 实体转持久化对象
func (r *MySQLUserRepository) entityToPO(user *entity.User) UserPO {
	status := 1
	if !user.IsActive() {
		status = 0
	}

	// 获取手机号（从用户档案中）
	phone := ""
	if user.Profile() != nil && user.Profile().Phone != nil {
		phone = user.Profile().Phone.String()
	}

	return UserPO{
		UserID:       user.ID().String(),
		Username:     user.Username(),
		Email:        user.Email().String(),
		PasswordHash: user.PasswordHash(),
		Phone:        phone,
		Status:       status,
		CreatedAt:    user.CreatedAt().Time(),
		UpdatedAt:    user.UpdatedAt().Time(),
	}
}

// profileToPO 档案转持久化对象
func (r *MySQLUserRepository) profileToPO(userID string, profile *valueobject.UserProfile) UserProfilePO {
	po := UserProfilePO{
		UserID:    userID,
		Nickname:  profile.Nickname,
		AvatarURL: profile.AvatarURL,
		Gender:    int(profile.Gender),
		Bio:       profile.Bio,
	}

	// Phone字段现在在users表中，不在user_profiles表中

	return po
}

// poToEntity 持久化对象转实体
func (r *MySQLUserRepository) poToEntity(userPO *UserPO, profilePO *UserProfilePO) (*entity.User, error) {
	// 创建邮箱值对象
	email, err := domain.NewEmail(userPO.Email)
	if err != nil {
		return nil, err
	}

	// 创建用户档案
	profile := &valueobject.UserProfile{
		Nickname:  profilePO.Nickname,
		AvatarURL: profilePO.AvatarURL,
		Gender:    valueobject.Gender(profilePO.Gender),
		Bio:       profilePO.Bio,
	}

	// 处理手机号（从users表中获取）
	if userPO.Phone != "" {
		phone, err := domain.NewPhone(userPO.Phone)
		if err != nil {
			return nil, err
		}
		profile.Phone = &phone
	}

	// 重建用户实体（这里需要添加一个重建方法）
	return entity.ReconstructUser(
		domain.UserID(userPO.UserID),
		userPO.Username,
		email,
		userPO.PasswordHash,
		profile,
		userPO.Status == 1,
		domain.NewTimestamp(userPO.CreatedAt),
		domain.NewTimestamp(userPO.UpdatedAt),
	), nil
}
