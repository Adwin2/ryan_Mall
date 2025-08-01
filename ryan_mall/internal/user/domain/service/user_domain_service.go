package service

import (
	"context"

	"ryan-mall-microservices/internal/shared/domain"
	"ryan-mall-microservices/internal/user/domain/entity"
	"ryan-mall-microservices/internal/user/domain/repository"
)

// UserDomainService 用户领域服务
type UserDomainService struct {
	userRepo repository.UserRepository
}

// NewUserDomainService 创建用户领域服务
func NewUserDomainService(userRepo repository.UserRepository) *UserDomainService {
	return &UserDomainService{
		userRepo: userRepo,
	}
}

// CheckUsernameUniqueness 检查用户名唯一性
func (s *UserDomainService) CheckUsernameUniqueness(ctx context.Context, username string) error {
	exists, err := s.userRepo.ExistsByUsername(ctx, username)
	if err != nil {
		return domain.NewInternalError("failed to check username uniqueness", err)
	}
	
	if exists {
		return domain.NewAlreadyExistsError("user", "username", username)
	}
	
	return nil
}

// CheckEmailUniqueness 检查邮箱唯一性
func (s *UserDomainService) CheckEmailUniqueness(ctx context.Context, email string) error {
	exists, err := s.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return domain.NewInternalError("failed to check email uniqueness", err)
	}
	
	if exists {
		return domain.NewAlreadyExistsError("user", "email", email)
	}
	
	return nil
}

// ValidateUserForRegistration 验证用户注册
func (s *UserDomainService) ValidateUserForRegistration(ctx context.Context, username, email string) error {
	// 检查用户名唯一性
	if err := s.CheckUsernameUniqueness(ctx, username); err != nil {
		return err
	}
	
	// 检查邮箱唯一性
	if err := s.CheckEmailUniqueness(ctx, email); err != nil {
		return err
	}
	
	return nil
}

// CanUserLogin 检查用户是否可以登录
func (s *UserDomainService) CanUserLogin(ctx context.Context, user *entity.User) error {
	if user == nil {
		return domain.NewNotFoundError("user", "")
	}
	
	if !user.IsActive() {
		return domain.NewBusinessError(domain.ErrCodeForbidden, "user account is inactive")
	}
	
	return nil
}
