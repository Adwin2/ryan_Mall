package entity

import (
	"strings"

	"ryan-mall-microservices/internal/shared/domain"
	"ryan-mall-microservices/internal/shared/events"
	"ryan-mall-microservices/internal/shared/infrastructure"
	"ryan-mall-microservices/internal/user/domain/valueobject"
)

// User 用户聚合根
type User struct {
	id           domain.UserID
	username     string
	email        domain.Email
	passwordHash string
	profile      *valueobject.UserProfile
	status       UserStatus
	createdAt    domain.Timestamp
	updatedAt    domain.Timestamp
	domainEvents []events.Event
}

// UserStatus 用户状态
type UserStatus int

const (
	UserStatusActive   UserStatus = 1
	UserStatusInactive UserStatus = 0
)

// NewUser 创建新用户
func NewUser(username, email, password string) (*User, error) {
	// 验证用户名
	if err := validateUsername(username); err != nil {
		return nil, err
	}

	// 验证邮箱
	emailVO, err := domain.NewEmail(email)
	if err != nil {
		return nil, err
	}

	// 验证密码
	if err := validatePassword(password); err != nil {
		return nil, err
	}

	// 哈希密码
	passwordHash, err := infrastructure.HashPassword(password)
	if err != nil {
		return nil, domain.NewInternalError("failed to hash password", err)
	}

	// 创建用户
	user := &User{
		id:           domain.NewUserID(),
		username:     username,
		email:        emailVO,
		passwordHash: passwordHash,
		profile:      valueobject.NewUserProfile(),
		status:       UserStatusActive,
		createdAt:    domain.Now(),
		updatedAt:    domain.Now(),
		domainEvents: make([]events.Event, 0),
	}

	// 添加用户注册事件
	event := events.NewUserRegisteredEvent(
		user.id.String(),
		user.username,
		user.email.String(),
	)
	user.addDomainEvent(event)

	return user, nil
}

// ID 获取用户ID
func (u *User) ID() domain.UserID {
	return u.id
}

// Username 获取用户名
func (u *User) Username() string {
	return u.username
}

// Email 获取邮箱
func (u *User) Email() domain.Email {
	return u.email
}

// Profile 获取用户档案
func (u *User) Profile() *valueobject.UserProfile {
	return u.profile
}

// IsActive 是否激活
func (u *User) IsActive() bool {
	return u.status == UserStatusActive
}

// CreatedAt 获取创建时间
func (u *User) CreatedAt() domain.Timestamp {
	return u.createdAt
}

// UpdatedAt 获取更新时间
func (u *User) UpdatedAt() domain.Timestamp {
	return u.updatedAt
}

// PasswordHash 获取密码哈希（仅供基础设施层使用）
func (u *User) PasswordHash() string {
	return u.passwordHash
}

// CheckPassword 验证密码
func (u *User) CheckPassword(password string) bool {
	return infrastructure.CheckPassword(password, u.passwordHash)
}

// ChangePassword 修改密码
func (u *User) ChangePassword(oldPassword, newPassword string) error {
	// 验证旧密码
	if !u.CheckPassword(oldPassword) {
		return domain.NewValidationError("old password is incorrect")
	}

	// 验证新密码
	if err := validatePassword(newPassword); err != nil {
		return err
	}

	// 检查新密码是否与旧密码相同
	if oldPassword == newPassword {
		return domain.NewValidationError("new password must be different from old password")
	}

	// 哈希新密码
	passwordHash, err := infrastructure.HashPassword(newPassword)
	if err != nil {
		return domain.NewInternalError("failed to hash password", err)
	}

	u.passwordHash = passwordHash
	u.updatedAt = domain.Now()

	return nil
}

// UpdateProfile 更新用户档案
func (u *User) UpdateProfile(nickname, phone string) error {
	// 验证手机号（如果提供）
	var phoneVO *domain.Phone
	if phone != "" {
		p, err := domain.NewPhone(phone)
		if err != nil {
			return err
		}
		phoneVO = &p
	}

	// 更新档案
	u.profile.UpdateProfile(nickname, phoneVO)
	u.updatedAt = domain.Now()

	return nil
}

// Activate 激活用户
func (u *User) Activate() {
	u.status = UserStatusActive
	u.updatedAt = domain.Now()
}

// Deactivate 停用用户
func (u *User) Deactivate() {
	u.status = UserStatusInactive
	u.updatedAt = domain.Now()
}

// DomainEvents 获取领域事件
func (u *User) DomainEvents() []events.Event {
	return u.domainEvents
}

// ClearDomainEvents 清除领域事件
func (u *User) ClearDomainEvents() {
	u.domainEvents = make([]events.Event, 0)
}

// addDomainEvent 添加领域事件
func (u *User) addDomainEvent(event events.Event) {
	u.domainEvents = append(u.domainEvents, event)
}

// validateUsername 验证用户名
func validateUsername(username string) error {
	username = strings.TrimSpace(username)
	
	if username == "" {
		return domain.NewValidationError("username cannot be empty")
	}
	
	if len(username) < 3 {
		return domain.NewValidationError("username must be at least 3 characters long")
	}
	
	if len(username) > 50 {
		return domain.NewValidationError("username must be at most 50 characters long")
	}
	
	// 检查用户名格式（只允许字母、数字、下划线）
	for _, char := range username {
		if !((char >= 'a' && char <= 'z') || 
			 (char >= 'A' && char <= 'Z') || 
			 (char >= '0' && char <= '9') || 
			 char == '_') {
			return domain.NewValidationError("username can only contain letters, numbers, and underscores")
		}
	}
	
	return nil
}

// validatePassword 验证密码
func validatePassword(password string) error {
	if password == "" {
		return domain.NewValidationError("password cannot be empty")
	}
	
	if len(password) < 8 {
		return domain.NewValidationError("password must be at least 8 characters long")
	}
	
	if len(password) > 128 {
		return domain.NewValidationError("password must be at most 128 characters long")
	}
	
	// 检查密码强度
	if !infrastructure.IsValidPassword(password) {
		return domain.NewValidationError("password must contain at least one uppercase letter, one lowercase letter, one number, and one special character")
	}
	
	return nil
}

// ReconstructUser 重建用户实体（用于从持久化存储重建）
func ReconstructUser(
	id domain.UserID,
	username string,
	email domain.Email,
	passwordHash string,
	profile *valueobject.UserProfile,
	isActive bool,
	createdAt domain.Timestamp,
	updatedAt domain.Timestamp,
) *User {
	status := UserStatusInactive
	if isActive {
		status = UserStatusActive
	}

	return &User{
		id:           id,
		username:     username,
		email:        email,
		passwordHash: passwordHash,
		profile:      profile,
		status:       status,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
		domainEvents: make([]events.Event, 0),
	}
}
