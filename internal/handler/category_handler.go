package handler

import (
	"ryan-mall/internal/middleware"
	"ryan-mall/internal/model"
	"ryan-mall/internal/service"
	"ryan-mall/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CategoryHandler 分类HTTP处理器
type CategoryHandler struct {
	categoryService service.CategoryService
}

// NewCategoryHandler 创建分类处理器实例
func NewCategoryHandler(categoryService service.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
	}
}

// CreateCategory 创建分类
// POST /api/v1/categories
// 需要认证（管理员权限）
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	// 1. 绑定请求参数
	var req model.CategoryCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	
	// 2. 调用业务逻辑
	category, err := h.categoryService.CreateCategory(&req)
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}
	
	// 3. 返回成功响应
	response.SuccessWithMessage(c, "分类创建成功", category)
}

// GetCategory 获取分类详情
// GET /api/v1/categories/:id
// 公开接口
func (h *CategoryHandler) GetCategory(c *gin.Context) {
	// 1. 获取路径参数
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "分类ID格式错误")
		return
	}
	
	// 2. 调用业务逻辑
	category, err := h.categoryService.GetCategory(uint(id))
	if err != nil {
		response.Error(c, response.NOT_FOUND, err.Error())
		return
	}
	
	// 3. 返回成功响应
	response.Success(c, category)
}

// UpdateCategory 更新分类
// PUT /api/v1/categories/:id
// 需要认证（管理员权限）
func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	// 1. 获取路径参数
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "分类ID格式错误")
		return
	}
	
	// 2. 绑定请求参数
	var req model.CategoryUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	
	// 3. 调用业务逻辑
	err = h.categoryService.UpdateCategory(uint(id), &req)
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}
	
	// 4. 返回成功响应
	response.SuccessWithMessage(c, "分类更新成功", nil)
}

// DeleteCategory 删除分类
// DELETE /api/v1/categories/:id
// 需要认证（管理员权限）
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	// 1. 获取路径参数
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "分类ID格式错误")
		return
	}
	
	// 2. 调用业务逻辑
	err = h.categoryService.DeleteCategory(uint(id))
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}
	
	// 3. 返回成功响应
	response.SuccessWithMessage(c, "分类删除成功", nil)
}

// GetAllCategories 获取所有分类
// GET /api/v1/categories
// 公开接口
func (h *CategoryHandler) GetAllCategories(c *gin.Context) {
	// 调用业务逻辑
	categories, err := h.categoryService.GetAllCategories()
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}
	
	// 返回成功响应
	response.Success(c, categories)
}

// GetCategoryTree 获取分类树
// GET /api/v1/categories/tree
// 公开接口，返回层级结构的分类树
func (h *CategoryHandler) GetCategoryTree(c *gin.Context) {
	// 调用业务逻辑
	tree, err := h.categoryService.GetCategoryTree()
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}
	
	// 返回成功响应
	response.Success(c, tree)
}

// GetTopCategories 获取顶级分类
// GET /api/v1/categories/top
// 公开接口
func (h *CategoryHandler) GetTopCategories(c *gin.Context) {
	// 调用业务逻辑
	categories, err := h.categoryService.GetTopCategories()
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}
	
	// 返回成功响应
	response.Success(c, categories)
}

// GetSubCategories 获取子分类
// GET /api/v1/categories/:id/children
// 公开接口
func (h *CategoryHandler) GetSubCategories(c *gin.Context) {
	// 1. 获取路径参数
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "分类ID格式错误")
		return
	}
	
	// 2. 调用业务逻辑
	categories, err := h.categoryService.GetSubCategories(uint(id))
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}
	
	// 3. 返回成功响应
	response.Success(c, categories)
}

// RegisterRoutes 注册分类相关路由
func (h *CategoryHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	// 公开路由（不需要认证）
	r.GET("/categories", h.GetAllCategories)           // 获取所有分类
	r.GET("/categories/tree", h.GetCategoryTree)       // 获取分类树
	r.GET("/categories/top", h.GetTopCategories)       // 获取顶级分类
	r.GET("/categories/:id", h.GetCategory)            // 获取分类详情
	r.GET("/categories/:id/children", h.GetSubCategories) // 获取子分类
	
	// 需要认证的路由（管理员功能）
	admin := r.Group("")
	admin.Use(authMiddleware.RequireAuth())
	{
		admin.POST("/categories", h.CreateCategory)        // 创建分类
		admin.PUT("/categories/:id", h.UpdateCategory)     // 更新分类
		admin.DELETE("/categories/:id", h.DeleteCategory)  // 删除分类
	}
}
