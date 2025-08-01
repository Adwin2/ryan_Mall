package http

import (
	"net/http"
	"strconv"

	"ryan-mall-microservices/internal/product/application/command"
	"ryan-mall-microservices/internal/product/application/query"
	"ryan-mall-microservices/internal/product/application/service"
	"ryan-mall-microservices/internal/shared/infrastructure"

	"github.com/gin-gonic/gin"
)

// ProductHandler 商品HTTP处理器
type ProductHandler struct {
	productAppSvc *service.ProductApplicationService
	logger        infrastructure.Logger
}

// NewProductHandler 创建商品HTTP处理器
func NewProductHandler(
	productAppSvc *service.ProductApplicationService,
	logger infrastructure.Logger,
) *ProductHandler {
	return &ProductHandler{
		productAppSvc: productAppSvc,
		logger:        logger,
	}
}

// RegisterRoutes 注册路由
func (h *ProductHandler) RegisterRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1")
	{
		products := v1.Group("/products")
		{
			products.POST("", h.CreateProduct)
			products.GET("/:id", h.GetProduct)
			products.PUT("/:id", h.UpdateProduct)
			products.DELETE("/:id", h.DeleteProduct)
			products.GET("", h.ListProducts)
			
			// 库存管理
			products.PUT("/:id/stock", h.UpdateStock)
			products.POST("/:id/stock/reserve", h.ReserveStock)
			products.POST("/:id/stock/release", h.ReleaseStock)
			products.GET("/:id/stock", h.CheckStock)
			
			// 价格管理
			products.PUT("/:id/price", h.UpdatePrice)
		}
	}
}

// CreateProduct 创建商品 - 暂未实现
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	h.respondError(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Create product not implemented yet", nil)
}

// GetProduct 获取商品
func (h *ProductHandler) GetProduct(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		h.respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Product ID is required", nil)
		return
	}

	qry := &query.GetProductQuery{
		ProductID: productID,
	}

	result, err := h.productAppSvc.GetProduct(c.Request.Context(), qry)
	if err != nil {
		h.handleDomainError(c, err)
		return
	}

	h.respondSuccess(c, http.StatusOK, "product retrieved successfully", result)
}

// UpdateProduct 更新商品 - 暂未实现
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	h.respondError(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Update product not implemented yet", nil)
}

// DeleteProduct 删除商品 - 暂未实现
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	h.respondError(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Delete product not implemented yet", nil)
}

// ListProducts 商品列表
func (h *ProductHandler) ListProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	categoryID := c.Query("category_id")
	keyword := c.Query("keyword")

	qry := &query.ListProductsQuery{
		CategoryID: categoryID,
		Keyword:    keyword,
		Page:       page,
		PageSize:   pageSize,
	}

	result, err := h.productAppSvc.ListProducts(c.Request.Context(), qry)
	if err != nil {
		h.handleDomainError(c, err)
		return
	}

	h.respondSuccess(c, http.StatusOK, "products retrieved successfully", result)
}

// UpdateStock 更新库存
func (h *ProductHandler) UpdateStock(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		h.respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Product ID is required", nil)
		return
	}

	var cmd command.UpdateStockCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		h.respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
		return
	}

	cmd.ProductID = productID

	if err := h.productAppSvc.UpdateStock(c.Request.Context(), &cmd); err != nil {
		h.handleDomainError(c, err)
		return
	}

	h.respondSuccess(c, http.StatusOK, "stock updated successfully", nil)
}

// ReserveStock 预留库存
func (h *ProductHandler) ReserveStock(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		h.respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Product ID is required", nil)
		return
	}

	var cmd command.ReserveStockCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		h.respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
		return
	}

	cmd.ProductID = productID

	if err := h.productAppSvc.ReserveStock(c.Request.Context(), &cmd); err != nil {
		h.handleDomainError(c, err)
		return
	}

	h.respondSuccess(c, http.StatusOK, "stock reserved successfully", nil)
}

// ReleaseStock 释放库存
func (h *ProductHandler) ReleaseStock(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		h.respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Product ID is required", nil)
		return
	}

	var cmd command.ReleaseStockCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		h.respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
		return
	}

	cmd.ProductID = productID

	if err := h.productAppSvc.ReleaseStock(c.Request.Context(), &cmd); err != nil {
		h.handleDomainError(c, err)
		return
	}

	h.respondSuccess(c, http.StatusOK, "stock released successfully", nil)
}

// CheckStock 检查库存
func (h *ProductHandler) CheckStock(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		h.respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Product ID is required", nil)
		return
	}

	qry := &query.CheckStockQuery{
		ProductID: productID,
	}

	result, err := h.productAppSvc.CheckStock(c.Request.Context(), qry)
	if err != nil {
		h.handleDomainError(c, err)
		return
	}

	h.respondSuccess(c, http.StatusOK, "stock checked successfully", result)
}

// UpdatePrice 更新价格
func (h *ProductHandler) UpdatePrice(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		h.respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Product ID is required", nil)
		return
	}

	var cmd command.UpdatePriceCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		h.respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
		return
	}

	cmd.ProductID = productID

	if err := h.productAppSvc.UpdatePrice(c.Request.Context(), &cmd); err != nil {
		h.handleDomainError(c, err)
		return
	}

	h.respondSuccess(c, http.StatusOK, "price updated successfully", nil)
}

// respondSuccess 成功响应
func (h *ProductHandler) respondSuccess(c *gin.Context, code int, message string, data interface{}) {
	response := gin.H{
		"code":    code,
		"message": message,
	}
	if data != nil {
		response["data"] = data
	}
	c.JSON(code, response)
}

// respondError 错误响应
func (h *ProductHandler) respondError(c *gin.Context, code int, errorCode, message string, err error) {
	response := gin.H{
		"code":    code,
		"message": message,
		"error":   errorCode,
	}

	if err != nil {
		h.logger.Error("HTTP request error",
			infrastructure.String("method", c.Request.Method),
			infrastructure.String("path", c.Request.URL.Path),
			infrastructure.String("error", err.Error()),
		)
	}

	c.JSON(code, response)
}

// handleDomainError 处理领域错误
func (h *ProductHandler) handleDomainError(c *gin.Context, err error) {
	// 这里需要根据实际的错误类型进行处理
	// 暂时使用简单的错误处理
	if err.Error() == "record not found" {
		h.respondError(c, http.StatusNotFound, "NOT_FOUND", "Product not found", err)
		return
	}

	// 检查是否是库存不足错误
	if err.Error() == "insufficient stock" {
		h.respondError(c, http.StatusBadRequest, "INSUFFICIENT_STOCK", err.Error(), err)
		return
	}

	// 默认内部错误
	h.respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", err)
}
