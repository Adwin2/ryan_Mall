package handler

import (
	"ryan-mall/internal/middleware"
	"ryan-mall/internal/model"
	"ryan-mall/internal/service"
	"ryan-mall/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ProductHandler 商品HTTP处理器
type ProductHandler struct {
	productService  service.ProductService
	categoryService service.CategoryService
}

// NewProductHandler 创建商品处理器实例
func NewProductHandler(productService service.ProductService, categoryService service.CategoryService) *ProductHandler {
	return &ProductHandler{
		productService:  productService,
		categoryService: categoryService,
	}
}

// CreateProduct 创建商品
// POST /api/v1/products
// 需要认证（管理员权限）
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	// 1. 绑定请求参数
	var req model.ProductCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	
	// 2. 调用业务逻辑
	product, err := h.productService.CreateProduct(&req)
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}
	
	// 3. 返回成功响应
	response.SuccessWithMessage(c, "商品创建成功", product)
}

// GetProduct 获取商品详情
// GET /api/v1/products/:id
// 公开接口
func (h *ProductHandler) GetProduct(c *gin.Context) {
	// 1. 获取路径参数
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "商品ID格式错误")
		return
	}
	
	// 2. 调用业务逻辑
	product, err := h.productService.GetProduct(uint(id))
	if err != nil {
		response.Error(c, response.NOT_FOUND, err.Error())
		return
	}
	
	// 3. 返回成功响应
	response.Success(c, product)
}

// UpdateProduct 更新商品
// PUT /api/v1/products/:id
// 需要认证（管理员权限）
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	// 1. 获取路径参数
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "商品ID格式错误")
		return
	}
	
	// 2. 绑定请求参数
	var req model.ProductUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	
	// 3. 调用业务逻辑
	err = h.productService.UpdateProduct(uint(id), &req)
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}
	
	// 4. 返回成功响应
	response.SuccessWithMessage(c, "商品更新成功", nil)
}

// DeleteProduct 删除商品
// DELETE /api/v1/products/:id
// 需要认证（管理员权限）
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	// 1. 获取路径参数
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "商品ID格式错误")
		return
	}
	
	// 2. 调用业务逻辑
	err = h.productService.DeleteProduct(uint(id))
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}
	
	// 3. 返回成功响应
	response.SuccessWithMessage(c, "商品删除成功", nil)
}

// GetProductList 获取商品列表
// GET /api/v1/products
// 公开接口，支持搜索和筛选
func (h *ProductHandler) GetProductList(c *gin.Context) {
	// 1. 绑定查询参数
	var req model.ProductListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "查询参数错误: "+err.Error())
		return
	}
	
	// 2. 调用业务逻辑
	result, err := h.productService.GetProductList(&req)
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}
	
	// 3. 返回成功响应
	response.Success(c, result)
}

// GetProductsByCategory 根据分类获取商品
// GET /api/v1/categories/:id/products
// 公开接口
func (h *ProductHandler) GetProductsByCategory(c *gin.Context) {
	// 1. 获取路径参数
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "分类ID格式错误")
		return
	}
	
	// 2. 调用业务逻辑
	products, err := h.productService.GetProductsByCategory(uint(id))
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}
	
	// 3. 返回成功响应
	response.Success(c, products)
}

// UpdateStock 更新商品库存
// PUT /api/v1/products/:id/stock
// 需要认证（管理员权限）
func (h *ProductHandler) UpdateStock(c *gin.Context) {
	// 1. 获取路径参数
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "商品ID格式错误")
		return
	}
	
	// 2. 绑定请求参数
	var req struct {
		Stock int `json:"stock" binding:"required,min=0"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	
	// 3. 调用业务逻辑
	err = h.productService.UpdateStock(uint(id), req.Stock)
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}
	
	// 4. 返回成功响应
	response.SuccessWithMessage(c, "库存更新成功", nil)
}

// RegisterRoutes 注册商品相关路由
func (h *ProductHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	// 公开路由（不需要认证）
	r.GET("/products", h.GetProductList)           // 获取商品列表
	r.GET("/products/:id", h.GetProduct)           // 获取商品详情
	r.GET("/categories/:id/products", h.GetProductsByCategory) // 根据分类获取商品
	
	// 需要认证的路由（管理员功能）
	admin := r.Group("")
	admin.Use(authMiddleware.RequireAuth())
	{
		admin.POST("/products", h.CreateProduct)           // 创建商品
		admin.PUT("/products/:id", h.UpdateProduct)        // 更新商品
		admin.DELETE("/products/:id", h.DeleteProduct)     // 删除商品
		admin.PUT("/products/:id/stock", h.UpdateStock)    // 更新库存
	}
}
