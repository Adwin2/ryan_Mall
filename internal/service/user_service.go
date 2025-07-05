package service

import (
	"errors"
	"ryan-mall/internal/model"
	"ryan-mall/internal/repository"
	"ryan-mall/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)

// UserService 用户业务逻辑层接口
// 定义用户相关的业务操作方法
type UserService interface {
	Register(req *model.UserRegisterRequest) (*model.UserLoginResponse, error) // 用户注册
	Login(req *model.UserLoginRequest) (*model.UserLoginResponse, error)       // 用户登录
	GetProfile(userID uint) (*model.UserProfileResponse, error)                // 获取用户资料
	UpdateProfile(userID uint, updates map[string]interface{}) error           // 更新用户资料
	ChangePassword(userID uint, oldPassword, newPassword string) error         // 修改密码
	ValidateToken(tokenString string) (*jwt.Claims, error)                     // 验证令牌
}

// userService 用户业务逻辑层实现
type userService struct {
	userRepo   repository.UserRepository // 用户数据访问层
	jwtManager *jwt.JWTManager          // JWT管理器
}

// NewUserService 创建用户业务逻辑层实例
// 使用依赖注入的方式传入所需的依赖
func NewUserService(userRepo repository.UserRepository, jwtManager *jwt.JWTManager) UserService {
	return &userService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

// Register 用户注册
// 处理用户注册的完整业务流程
func (s *userService) Register(req *model.UserRegisterRequest) (*model.UserLoginResponse, error) {
	// 1. 验证用户名是否已存在
	// 在创建用户之前检查用户名唯一性
	exists, err := s.userRepo.ExistsByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("用户名已存在")
	}
	
	// 2. 验证邮箱是否已存在
	exists, err = s.userRepo.ExistsByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("邮箱已被注册")
	}
	
	// 3. 加密密码
	// 使用bcrypt算法加密密码，cost=12提供较高的安全性
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	
	// 4. 创建用户对象
	user := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Phone:        nil, // 注册时手机号可选
		Status:       model.UserStatusActive, // 默认激活状态
	}
	
	// 处理可选的手机号
	if req.Phone != "" {
		user.Phone = &req.Phone
	}
	
	// 5. 保存用户到数据库
	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}
	
	// 6. 生成JWT令牌
	token, err := s.jwtManager.GenerateToken(user.ID, user.Username, user.Email)
	if err != nil {
		return nil, err
	}
	
	// 7. 返回登录响应
	return &model.UserLoginResponse{
		User:  user,
		Token: token,
	}, nil
}

// Login 用户登录
// 处理用户登录的完整业务流程
func (s *userService) Login(req *model.UserLoginRequest) (*model.UserLoginResponse, error) {
	// 1. 查找用户
	// 支持使用用户名或邮箱登录
	var user *model.User
	var err error
	
	// 判断输入的是邮箱还是用户名（简单判断：包含@符号）
	if contains(req.Username, "@") {
		user, err = s.userRepo.GetByEmail(req.Username)
	} else {
		user, err = s.userRepo.GetByUsername(req.Username)
	}
	
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("用户名或密码错误")
	}
	
	// 2. 检查用户状态
	if user.Status != model.UserStatusActive {
		return nil, errors.New("账户已被禁用")
	}
	
	// 3. 验证密码
	// 使用bcrypt比较明文密码和哈希密码
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}
	
	// 4. 生成JWT令牌
	token, err := s.jwtManager.GenerateToken(user.ID, user.Username, user.Email)
	if err != nil {
		return nil, err
	}
	
	// 5. 返回登录响应
	return &model.UserLoginResponse{
		User:  user,
		Token: token,
	}, nil
}

// GetProfile 获取用户资料
func (s *userService) GetProfile(userID uint) (*model.UserProfileResponse, error) {
	// 1. 查找用户
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("用户不存在")
	}
	
	// 2. 转换为资料响应格式
	return user.ToProfileResponse(), nil
}

// UpdateProfile 更新用户资料
// 支持部分字段更新
func (s *userService) UpdateProfile(userID uint, updates map[string]interface{}) error {
	// 1. 查找用户
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("用户不存在")
	}
	
	// 2. 验证邮箱唯一性（如果要更新邮箱）
	if email, ok := updates["email"]; ok {
		emailStr := email.(string)
		if emailStr != user.Email {
			exists, err := s.userRepo.ExistsByEmail(emailStr)
			if err != nil {
				return err
			}
			if exists {
				return errors.New("邮箱已被使用")
			}
		}
	}
	
	// 3. 更新字段
	// 这里可以添加更多的字段验证逻辑
	for field, value := range updates {
		switch field {
		case "phone":
			if value == nil {
				user.Phone = nil
			} else {
				phoneStr := value.(string)
				user.Phone = &phoneStr
			}
		case "avatar":
			if value == nil {
				user.Avatar = nil
			} else {
				avatarStr := value.(string)
				user.Avatar = &avatarStr
			}
		case "email":
			user.Email = value.(string)
		}
	}
	
	// 4. 保存更新
	return s.userRepo.Update(user)
}

// ChangePassword 修改密码
func (s *userService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	// 1. 查找用户
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("用户不存在")
	}
	
	// 2. 验证旧密码
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword))
	if err != nil {
		return errors.New("原密码错误")
	}
	
	// 3. 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	
	// 4. 更新密码
	user.PasswordHash = string(hashedPassword)
	return s.userRepo.Update(user)
}

// ValidateToken 验证令牌
func (s *userService) ValidateToken(tokenString string) (*jwt.Claims, error) {
	return s.jwtManager.ValidateToken(tokenString)
}

// contains 检查字符串是否包含子字符串
// 这是一个辅助函数，用于判断登录输入是邮箱还是用户名
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
