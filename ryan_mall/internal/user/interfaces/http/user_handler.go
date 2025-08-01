package http

import (
	"net/http"
	"strconv"

	"ryan-mall-microservices/internal/shared/domain"
	"ryan-mall-microservices/internal/shared/infrastructure"
	"ryan-mall-microservices/internal/user/application/command"
	"ryan-mall-microservices/internal/user/application/query"
	"ryan-mall-microservices/internal/user/application/service"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户HTTP处理器
type UserHandler struct {
	userAppSvc *service.UserApplicationService
	logger     infrastructure.Logger
}

// NewUserHandler 创建用户HTTP处理器
func NewUserHandler(userAppSvc *service.UserApplicationService, logger infrastructure.Logger) *UserHandler {
	return &UserHandler{
		userAppSvc: userAppSvc,
		logger:     logger,
	}
}

// RegisterRoutes 注册路由
func (h *UserHandler) RegisterRoutes(router *gin.Engine) {
	userGroup := router.Group("/api/v1/users")
	{
		userGroup.POST("/register", h.RegisterUser)
		userGroup.POST("/login", h.LoginUser)
		userGroup.GET("/:id", h.GetUser)
		userGroup.GET("", h.ListUsers)
	}
}

// RegisterUser 用户注册
func (h *UserHandler) RegisterUser(c *gin.Context) {
	var req command.RegisterUserCommand
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	result, err := h.userAppSvc.RegisterUser(c.Request.Context(), &req)
	if err != nil {
		h.handleDomainError(c, err)
		return
	}

	h.respondSuccess(c, http.StatusCreated, "user registered successfully", result)
}

// LoginUser 用户登录
func (h *UserHandler) LoginUser(c *gin.Context) {
	var req command.LoginUserCommand
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	result, err := h.userAppSvc.LoginUser(c.Request.Context(), &req)
	if err != nil {
		h.handleDomainError(c, err)
		return
	}

	h.respondSuccess(c, http.StatusOK, "login successful", result)
}

// GetUser 获取用户信息
func (h *UserHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		h.respondError(c, http.StatusBadRequest, "user id is required", nil)
		return
	}

	req := &query.GetUserQuery{UserID: userID}
	result, err := h.userAppSvc.GetUser(c.Request.Context(), req)
	if err != nil {
		h.handleDomainError(c, err)
		return
	}

	h.respondSuccess(c, http.StatusOK, "user retrieved successfully", result)
}

// ListUsers 获取用户列表
func (h *UserHandler) ListUsers(c *gin.Context) {
	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	req := &query.ListUsersQuery{
		Page:     page,
		PageSize: pageSize,
	}

	result, err := h.userAppSvc.ListUsers(c.Request.Context(), req)
	if err != nil {
		h.handleDomainError(c, err)
		return
	}

	h.respondSuccess(c, http.StatusOK, "users retrieved successfully", result)
}

// APIResponse API响应结构
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// respondSuccess 成功响应
func (h *UserHandler) respondSuccess(c *gin.Context, statusCode int, message string, data interface{}) {
	response := APIResponse{
		Code:    statusCode,
		Message: message,
		Data:    data,
	}
	c.JSON(statusCode, response)
}

// respondError 错误响应
func (h *UserHandler) respondError(c *gin.Context, statusCode int, message string, err error) {
	response := APIResponse{
		Code:    statusCode,
		Message: message,
	}

	if err != nil {
		response.Error = err.Error()
		h.logger.Error("HTTP request error",
			infrastructure.String("method", c.Request.Method),
			infrastructure.String("path", c.Request.URL.Path),
			infrastructure.String("error", err.Error()),
		)
	}

	c.JSON(statusCode, response)
}

// handleDomainError 处理领域错误
func (h *UserHandler) handleDomainError(c *gin.Context, err error) {
	if domainErr, ok := err.(domain.DomainError); ok {
		switch domainErr.Code() {
		case domain.ErrCodeValidation:
			h.respondError(c, http.StatusBadRequest, "validation error", err)
		case domain.ErrCodeNotFound:
			h.respondError(c, http.StatusNotFound, "resource not found", err)
		case domain.ErrCodeAlreadyExists:
			h.respondError(c, http.StatusConflict, "resource already exists", err)
		case domain.ErrCodeUnauthorized:
			h.respondError(c, http.StatusUnauthorized, "unauthorized", err)
		case domain.ErrCodeForbidden:
			h.respondError(c, http.StatusForbidden, "forbidden", err)
		case domain.ErrCodeInternalError:
			h.respondError(c, http.StatusInternalServerError, "internal server error", err)
		default:
			h.respondError(c, http.StatusInternalServerError, "unknown error", err)
		}
	} else {
		h.respondError(c, http.StatusInternalServerError, "internal server error", err)
	}
}
