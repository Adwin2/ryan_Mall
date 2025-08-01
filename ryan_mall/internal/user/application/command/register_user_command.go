package command

import (
	"context"

	"ryan-mall-microservices/internal/shared/domain"
	"ryan-mall-microservices/internal/shared/events"
	"ryan-mall-microservices/internal/user/domain/entity"
	"ryan-mall-microservices/internal/user/domain/repository"
	"ryan-mall-microservices/internal/user/domain/service"
)

// RegisterUserCommand 用户注册命令
type RegisterUserCommand struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// RegisterUserResult 用户注册结果
type RegisterUserResult struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// RegisterUserHandler 用户注册命令处理器
type RegisterUserHandler struct {
	userRepo        repository.UserRepository
	userDomainSvc   *service.UserDomainService
	eventPublisher  *events.EventPublisher
}

// NewRegisterUserHandler 创建用户注册命令处理器
func NewRegisterUserHandler(
	userRepo repository.UserRepository,
	userDomainSvc *service.UserDomainService,
	eventPublisher *events.EventPublisher,
) *RegisterUserHandler {
	return &RegisterUserHandler{
		userRepo:       userRepo,
		userDomainSvc:  userDomainSvc,
		eventPublisher: eventPublisher,
	}
}

// Handle 处理用户注册命令
func (h *RegisterUserHandler) Handle(ctx context.Context, cmd *RegisterUserCommand) (*RegisterUserResult, error) {
	// 验证用户注册信息
	if err := h.userDomainSvc.ValidateUserForRegistration(ctx, cmd.Username, cmd.Email); err != nil {
		return nil, err
	}

	// 创建用户实体
	user, err := entity.NewUser(cmd.Username, cmd.Email, cmd.Password)
	if err != nil {
		return nil, err
	}

	// 保存用户
	if err := h.userRepo.Save(ctx, user); err != nil {
		return nil, domain.NewInternalError("failed to save user", err)
	}

	// 发布领域事件
	if h.eventPublisher != nil {
		if err := h.eventPublisher.PublishEvents(ctx, user.DomainEvents()...); err != nil {
			// 事件发布失败不应该影响主流程，但应该记录日志
			// 在实际项目中，这里应该记录日志
		}
	}

	// 清除领域事件
	user.ClearDomainEvents()

	return &RegisterUserResult{
		UserID:   user.ID().String(),
		Username: user.Username(),
		Email:    user.Email().String(),
	}, nil
}
