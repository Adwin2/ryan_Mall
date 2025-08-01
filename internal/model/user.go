package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
// 使用GORM的标签来定义数据库映射和验证规则
type User struct {
	ID           uint      `json:"id" gorm:"primaryKey;autoIncrement;not null"`                                    // 用户ID，主键
	Username     string    `json:"username" gorm:"uniqueIndex;size:50;not null"`           // 用户名，唯一索引，最大50字符
	Email        string    `json:"email" gorm:"uniqueIndex;size:100;not null"`             // 邮箱，唯一索引
	PasswordHash string    `json:"-" gorm:"column:password_hash;size:255;not null"`        // 密码哈希（json:"-"表示不序列化到JSON）
	Phone        *string   `json:"phone" gorm:"size:20"`                                   // 手机号（指针类型表示可为空）
	Avatar       *string   `json:"avatar" gorm:"size:255"`                                 // 头像URL
	Status       int       `json:"status" gorm:"default:1;index"`                          // 用户状态，默认1，添加索引
	CreatedAt    time.Time `json:"created_at"`                                             // 创建时间，GORM自动管理
	UpdatedAt    time.Time `json:"updated_at"`                                             // 更新时间，GORM自动管理
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`                                    // 软删除时间，GORM自动管理
}

// UserRegisterRequest 用户注册请求结构体
// 用于接收前端传来的注册数据
type UserRegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"` // 用户名，必填，3-50字符
	Email    string `json:"email" binding:"required,email"`           // 邮箱，必填，需要是有效邮箱格式
	Password string `json:"password" binding:"required,min=6"`        // 密码，必填，最少6位
	Phone    string `json:"phone" binding:"omitempty,len=11"`         // 手机号，可选，如果填写必须是11位
}

// UserLoginRequest 用户登录请求结构体
type UserLoginRequest struct {
	Username string `json:"username" binding:"required"` // 用户名或邮箱
	Password string `json:"password" binding:"required"` // 密码
}

// UserLoginResponse 用户登录响应结构体
// 登录成功后返回给前端的数据
type UserLoginResponse struct {
	User  *User  `json:"user"`  // 用户信息
	Token string `json:"token"` // JWT令牌
}

// UserProfileResponse 用户资料响应结构体
// 获取用户信息时返回的数据（不包含敏感信息）
type UserProfileResponse struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Phone     *string   `json:"phone"`
	Avatar    *string   `json:"avatar"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// ToProfileResponse 将User转换为UserProfileResponse
// 这个方法用于过滤敏感信息，只返回可以公开的用户数据
func (u *User) ToProfileResponse() *UserProfileResponse {
	return &UserProfileResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Phone:     u.Phone,
		Avatar:    u.Avatar,
		Status:    u.Status,
		CreatedAt: u.CreatedAt,
	}
}

// 用户状态常量
const (
	UserStatusDisabled = 0 // 禁用
	UserStatusActive   = 1 // 正常
)
