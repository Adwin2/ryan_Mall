package service

import (
	"context"
	"time"

	"ryan-mall-microservices/internal/shared/events"
	"ryan-mall-microservices/internal/user/application/command"
	"ryan-mall-microservices/internal/user/application/query"
	"ryan-mall-microservices/internal/user/domain/repository"
	"ryan-mall-microservices/internal/user/domain/service"
)

// UserApplicationService 用户应用服务
type UserApplicationService struct {
	// 命令处理器
	registerUserHandler *command.RegisterUserHandler
	loginUserHandler    *command.LoginUserHandler

	// 查询处理器
	getUserHandler   *query.GetUserHandler
	listUsersHandler *query.ListUsersHandler
}

// NewUserApplicationService 创建用户应用服务
func NewUserApplicationService(
	userRepo repository.UserRepository,
	eventPublisher *events.EventPublisher,
	jwtSecret string,
	jwtExpireTime time.Duration,
) *UserApplicationService {
	// 创建领域服务
	userDomainSvc := service.NewUserDomainService(userRepo)

	return &UserApplicationService{
		// 初始化命令处理器
		registerUserHandler: command.NewRegisterUserHandler(userRepo, userDomainSvc, eventPublisher),
		loginUserHandler:    command.NewLoginUserHandler(userRepo, userDomainSvc, jwtSecret, jwtExpireTime),

		// 初始化查询处理器
		getUserHandler:   query.NewGetUserHandler(userRepo),
		listUsersHandler: query.NewListUsersHandler(userRepo),
	}
}

// RegisterUser 注册用户
func (s *UserApplicationService) RegisterUser(ctx context.Context, cmd *command.RegisterUserCommand) (*command.RegisterUserResult, error) {
	return s.registerUserHandler.Handle(ctx, cmd)
}

// LoginUser 用户登录
func (s *UserApplicationService) LoginUser(ctx context.Context, cmd *command.LoginUserCommand) (*command.LoginUserResult, error) {
	return s.loginUserHandler.Handle(ctx, cmd)
}

// GetUser 获取用户
func (s *UserApplicationService) GetUser(ctx context.Context, query *query.GetUserQuery) (*query.UserDTO, error) {
	return s.getUserHandler.Handle(ctx, query)
}

// ListUsers 获取用户列表
func (s *UserApplicationService) ListUsers(ctx context.Context, query *query.ListUsersQuery) (*query.ListUsersResult, error) {
	return s.listUsersHandler.Handle(ctx, query)
}
