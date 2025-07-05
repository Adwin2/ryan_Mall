package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Product 商品模型
// 使用GORM标签定义数据库映射
type Product struct {
	ID            uint           `json:"id" gorm:"primaryKey"`                                    // 商品ID，主键
	Name          string         `json:"name" gorm:"size:200;not null;index"`                    // 商品名称，添加索引便于搜索
	Description   *string        `json:"description" gorm:"type:text"`                           // 商品描述，使用TEXT类型
	CategoryID    uint           `json:"category_id" gorm:"not null;index"`                      // 分类ID，外键，添加索引
	Price         float64        `json:"price" gorm:"type:decimal(10,2);not null;index"`         // 商品价格，使用DECIMAL类型确保精度
	OriginalPrice *float64       `json:"original_price" gorm:"type:decimal(10,2)"`               // 原价
	Stock         int            `json:"stock" gorm:"not null;default:0"`                        // 库存数量
	SalesCount    int            `json:"sales_count" gorm:"default:0;index"`                     // 销售数量，添加索引便于排序
	MainImage     *string        `json:"main_image" gorm:"size:255"`                             // 主图片URL
	Images        JSONArray `json:"images" gorm:"type:json"`                               // 商品图片列表，JSON格式
	Status        int            `json:"status" gorm:"default:1;index"`                          // 商品状态，添加索引
	CreatedAt     time.Time      `json:"created_at" gorm:"index"`                                // 创建时间，添加索引便于排序
	UpdatedAt     time.Time      `json:"updated_at"`                                             // 更新时间
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`                                         // 软删除时间

	// 关联关系
	Category      Category       `json:"category,omitempty" gorm:"foreignKey:CategoryID"`        // 所属分类
}

// Category 商品分类模型
type Category struct {
	ID        uint           `json:"id" gorm:"primaryKey"`                                       // 分类ID，主键
	Name      string         `json:"name" gorm:"size:100;not null;index"`                       // 分类名称
	ParentID  uint           `json:"parent_id" gorm:"default:0;index"`                          // 父分类ID，0表示顶级分类
	SortOrder int            `json:"sort_order" gorm:"default:0;index"`                         // 排序权重
	Status    int            `json:"status" gorm:"default:1;index"`                             // 状态
	CreatedAt time.Time      `json:"created_at"`                                                // 创建时间
	UpdatedAt time.Time      `json:"updated_at"`                                                // 更新时间
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`                                            // 软删除时间

	// 关联关系（不使用外键约束，避免自引用问题）
	Products  []Product      `json:"products,omitempty" gorm:"foreignKey:CategoryID"`           // 该分类下的商品
	Children  []*Category    `json:"children,omitempty" gorm:"-"`                               // 子分类（不存储到数据库）
}

// JSONArray 自定义类型，用于处理MySQL的JSON字段
// 这个类型可以在Go结构体和MySQL JSON字段之间自动转换
type JSONArray []string

// Scan 实现sql.Scanner接口，用于从数据库读取JSON数据
func (ja *JSONArray) Scan(value interface{}) error {
	if value == nil {
		*ja = nil
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into JSONArray", value)
	}
	
	return json.Unmarshal(bytes, ja)
}

// Value 实现driver.Valuer接口，用于将Go数据写入数据库
func (ja JSONArray) Value() (driver.Value, error) {
	if ja == nil {
		return nil, nil
	}
	return json.Marshal(ja)
}

// ProductListRequest 商品列表查询请求
type ProductListRequest struct {
	Page       int    `form:"page,default=1" binding:"min=1"`           // 页码，默认第1页
	PageSize   int    `form:"page_size,default=10" binding:"min=1,max=100"` // 每页数量，默认10条，最多100条
	CategoryID *uint  `form:"category_id"`                              // 分类ID，可选
	Keyword    string `form:"keyword"`                                  // 搜索关键词，可选
	MinPrice   *float64 `form:"min_price" binding:"omitempty,min=0"`    // 最低价格
	MaxPrice   *float64 `form:"max_price" binding:"omitempty,min=0"`    // 最高价格
	SortBy     string `form:"sort_by,default=created_at"`               // 排序字段
	SortOrder  string `form:"sort_order,default=desc"`                 // 排序方向：asc, desc
}

// ProductListResponse 商品列表响应
type ProductListResponse struct {
	Products   []*Product `json:"products"`    // 商品列表
	Total      int64      `json:"total"`       // 总数量
	Page       int        `json:"page"`        // 当前页码
	PageSize   int        `json:"page_size"`   // 每页数量
	TotalPages int        `json:"total_pages"` // 总页数
}

// ProductCreateRequest 创建商品请求
type ProductCreateRequest struct {
	Name          string   `json:"name" binding:"required,max=200"`
	Description   string   `json:"description"`
	CategoryID    uint     `json:"category_id" binding:"required"`
	Price         float64  `json:"price" binding:"required,min=0"`
	OriginalPrice *float64 `json:"original_price" binding:"omitempty,min=0"`
	Stock         int      `json:"stock" binding:"required,min=0"`
	MainImage     string   `json:"main_image"`
	Images        []string `json:"images"`
}

// ProductUpdateRequest 更新商品请求
type ProductUpdateRequest struct {
	Name          *string  `json:"name" binding:"omitempty,max=200"`
	Description   *string  `json:"description"`
	CategoryID    *uint    `json:"category_id"`
	Price         *float64 `json:"price" binding:"omitempty,min=0"`
	OriginalPrice *float64 `json:"original_price" binding:"omitempty,min=0"`
	Stock         *int     `json:"stock" binding:"omitempty,min=0"`
	MainImage     *string  `json:"main_image"`
	Images        []string `json:"images"`
	Status        *int     `json:"status" binding:"omitempty,oneof=0 1"`
}

// 商品状态常量
const (
	ProductStatusOffline = 0 // 下架
	ProductStatusOnline  = 1 // 上架
)

// 排序字段常量
const (
	SortByCreatedAt  = "created_at"
	SortByPrice      = "price"
	SortBySalesCount = "sales_count"
)

// 排序方向常量
const (
	SortOrderAsc  = "asc"
	SortOrderDesc = "desc"
)

// CategoryCreateRequest 创建分类请求
type CategoryCreateRequest struct {
	Name      string `json:"name" binding:"required,max=100"`      // 分类名称
	ParentID  uint   `json:"parent_id"`                            // 父分类ID，0表示顶级分类
	SortOrder int    `json:"sort_order"`                           // 排序权重
}

// CategoryUpdateRequest 更新分类请求
type CategoryUpdateRequest struct {
	Name      *string `json:"name" binding:"omitempty,max=100"`     // 分类名称
	ParentID  *uint   `json:"parent_id"`                            // 父分类ID
	SortOrder *int    `json:"sort_order"`                           // 排序权重
	Status    *int    `json:"status" binding:"omitempty,oneof=0 1"` // 状态
}
