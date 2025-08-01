package valueobject

import (
	"ryan-mall-microservices/internal/shared/domain"
)

// Gender 性别枚举
type Gender int

const (
	GenderUnknown Gender = 0
	GenderMale    Gender = 1
	GenderFemale  Gender = 2
)

// String 转换为字符串
func (g Gender) String() string {
	switch g {
	case GenderMale:
		return "male"
	case GenderFemale:
		return "female"
	default:
		return "unknown"
	}
}

// UserProfile 用户档案值对象
type UserProfile struct {
	Nickname  string       `json:"nickname"`
	AvatarURL string       `json:"avatar_url"`
	Gender    Gender       `json:"gender"`
	Phone     *domain.Phone `json:"phone"`
	Bio       string       `json:"bio"`
}

// NewUserProfile 创建新的用户档案
func NewUserProfile() *UserProfile {
	return &UserProfile{
		Gender: GenderUnknown,
	}
}

// UpdateProfile 更新档案信息
func (p *UserProfile) UpdateProfile(nickname string, phone *domain.Phone) {
	if nickname != "" {
		p.Nickname = nickname
	}
	if phone != nil {
		p.Phone = phone
	}
}

// UpdateAvatar 更新头像
func (p *UserProfile) UpdateAvatar(avatarURL string) {
	p.AvatarURL = avatarURL
}

// UpdateGender 更新性别
func (p *UserProfile) UpdateGender(gender Gender) {
	p.Gender = gender
}

// UpdateBio 更新个人简介
func (p *UserProfile) UpdateBio(bio string) {
	p.Bio = bio
}

// IsComplete 检查档案是否完整
func (p *UserProfile) IsComplete() bool {
	return p.Nickname != "" && 
		   p.Phone != nil && 
		   p.Gender != GenderUnknown
}

// GetDisplayName 获取显示名称
func (p *UserProfile) GetDisplayName() string {
	if p.Nickname != "" {
		return p.Nickname
	}
	return "用户" // 默认显示名称
}
