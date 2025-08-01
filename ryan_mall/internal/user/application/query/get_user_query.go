package query

import (
	"context"

	"ryan-mall-microservices/internal/shared/domain"
	"ryan-mall-microservices/internal/user/domain/repository"
)

// GetUserQuery 获取用户查询
type GetUserQuery struct {
	UserID string `json:"user_id" validate:"required"`
}

// UserDTO 用户数据传输对象
type UserDTO struct {
	ID        string      `json:"id"`
	Username  string      `json:"username"`
	Email     string      `json:"email"`
	Profile   ProfileDTO  `json:"profile"`
	IsActive  bool        `json:"is_active"`
	CreatedAt int64       `json:"created_at"`
	UpdatedAt int64       `json:"updated_at"`
}

// ProfileDTO 用户档案数据传输对象
type ProfileDTO struct {
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
	Gender    string `json:"gender"`
	Phone     string `json:"phone"`
	Bio       string `json:"bio"`
}

// GetUserHandler 获取用户查询处理器
type GetUserHandler struct {
	userRepo repository.UserRepository
}

// NewGetUserHandler 创建获取用户查询处理器
func NewGetUserHandler(userRepo repository.UserRepository) *GetUserHandler {
	return &GetUserHandler{
		userRepo: userRepo,
	}
}

// Handle 处理获取用户查询
func (h *GetUserHandler) Handle(ctx context.Context, query *GetUserQuery) (*UserDTO, error) {
	// 查找用户
	user, err := h.userRepo.FindByID(ctx, domain.UserID(query.UserID))
	if err != nil {
		return nil, domain.NewInternalError("failed to find user", err)
	}
	if user == nil {
		return nil, domain.NewNotFoundError("user", query.UserID)
	}

	// 转换为DTO
	dto := &UserDTO{
		ID:        user.ID().String(),
		Username:  user.Username(),
		Email:     user.Email().String(),
		IsActive:  user.IsActive(),
		CreatedAt: user.CreatedAt().Unix(),
		UpdatedAt: user.UpdatedAt().Unix(),
		Profile: ProfileDTO{
			Nickname:  user.Profile().Nickname,
			AvatarURL: user.Profile().AvatarURL,
			Gender:    user.Profile().Gender.String(),
			Bio:       user.Profile().Bio,
		},
	}

	// 处理手机号（可能为空）
	if user.Profile().Phone != nil {
		dto.Profile.Phone = user.Profile().Phone.String()
	}

	return dto, nil
}

// ListUsersQuery 用户列表查询
type ListUsersQuery struct {
	Page     int `json:"page" validate:"min=1"`
	PageSize int `json:"page_size" validate:"min=1,max=100"`
}

// ListUsersResult 用户列表结果
type ListUsersResult struct {
	Users      []*UserDTO `json:"users"`
	Total      int64      `json:"total"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
	TotalPages int        `json:"total_pages"`
}

// ListUsersHandler 用户列表查询处理器
type ListUsersHandler struct {
	userRepo repository.UserRepository
}

// NewListUsersHandler 创建用户列表查询处理器
func NewListUsersHandler(userRepo repository.UserRepository) *ListUsersHandler {
	return &ListUsersHandler{
		userRepo: userRepo,
	}
}

// Handle 处理用户列表查询
func (h *ListUsersHandler) Handle(ctx context.Context, query *ListUsersQuery) (*ListUsersResult, error) {
	// 计算偏移量
	offset := (query.Page - 1) * query.PageSize

	// 查询用户列表
	users, total, err := h.userRepo.List(ctx, offset, query.PageSize)
	if err != nil {
		return nil, domain.NewInternalError("failed to list users", err)
	}

	// 转换为DTO
	userDTOs := make([]*UserDTO, len(users))
	for i, user := range users {
		userDTOs[i] = &UserDTO{
			ID:        user.ID().String(),
			Username:  user.Username(),
			Email:     user.Email().String(),
			IsActive:  user.IsActive(),
			CreatedAt: user.CreatedAt().Unix(),
			UpdatedAt: user.UpdatedAt().Unix(),
			Profile: ProfileDTO{
				Nickname:  user.Profile().Nickname,
				AvatarURL: user.Profile().AvatarURL,
				Gender:    user.Profile().Gender.String(),
				Bio:       user.Profile().Bio,
			},
		}

		// 处理手机号（可能为空）
		if user.Profile().Phone != nil {
			userDTOs[i].Profile.Phone = user.Profile().Phone.String()
		}
	}

	// 计算总页数
	totalPages := int((total + int64(query.PageSize) - 1) / int64(query.PageSize))

	return &ListUsersResult{
		Users:      userDTOs,
		Total:      total,
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalPages: totalPages,
	}, nil
}
