package command

import (
	"context"
	"time"

	"ryan-mall-microservices/internal/shared/domain"
	"ryan-mall-microservices/internal/user/domain/repository"
	"ryan-mall-microservices/internal/user/domain/service"

	"github.com/golang-jwt/jwt/v4"
)

// LoginUserCommand 用户登录命令
type LoginUserCommand struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// LoginUserResult 用户登录结果
type LoginUserResult struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	AccessToken string `json:"access_token"`
	ExpiresAt   int64  `json:"expires_at"`
}

// LoginUserHandler 用户登录命令处理器
type LoginUserHandler struct {
	userRepo      repository.UserRepository
	userDomainSvc *service.UserDomainService
	jwtSecret     string
	jwtExpireTime time.Duration
}

// NewLoginUserHandler 创建用户登录命令处理器
func NewLoginUserHandler(
	userRepo repository.UserRepository,
	userDomainSvc *service.UserDomainService,
	jwtSecret string,
	jwtExpireTime time.Duration,
) *LoginUserHandler {
	return &LoginUserHandler{
		userRepo:      userRepo,
		userDomainSvc: userDomainSvc,
		jwtSecret:     jwtSecret,
		jwtExpireTime: jwtExpireTime,
	}
}

// Handle 处理用户登录命令
func (h *LoginUserHandler) Handle(ctx context.Context, cmd *LoginUserCommand) (*LoginUserResult, error) {
	// 查找用户
	user, err := h.userRepo.FindByUsername(ctx, cmd.Username)
	if err != nil {
		return nil, domain.NewInternalError("failed to find user", err)
	}
	if user == nil {
		return nil, domain.NewUnauthorizedError("invalid username or password")
	}

	// 验证密码
	if !user.CheckPassword(cmd.Password) {
		return nil, domain.NewUnauthorizedError("invalid username or password")
	}

	// 检查用户是否可以登录
	if err := h.userDomainSvc.CanUserLogin(ctx, user); err != nil {
		return nil, err
	}

	// 生成JWT令牌
	token, expiresAt, err := h.generateJWT(user.ID().String(), user.Username())
	if err != nil {
		return nil, domain.NewInternalError("failed to generate token", err)
	}

	return &LoginUserResult{
		UserID:      user.ID().String(),
		Username:    user.Username(),
		Email:       user.Email().String(),
		AccessToken: token,
		ExpiresAt:   expiresAt,
	}, nil
}

// generateJWT 生成JWT令牌
func (h *LoginUserHandler) generateJWT(userID, username string) (string, int64, error) {
	now := time.Now()
	expiresAt := now.Add(h.jwtExpireTime)

	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"iat":      now.Unix(),
		"exp":      expiresAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		return "", 0, err
	}

	return tokenString, expiresAt.Unix(), nil
}
