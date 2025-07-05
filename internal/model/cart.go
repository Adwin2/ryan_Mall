package model

import (
	"time"

	"gorm.io/gorm"
)

// CartItem 购物车商品模型
// 存储用户购物车中的商品信息
type CartItem struct {
	ID        uint           `json:"id" gorm:"primaryKey"`                                       // 购物车项ID
	UserID    uint           `json:"user_id" gorm:"not null;index"`                             // 用户ID，添加索引
	ProductID uint           `json:"product_id" gorm:"not null;index"`                          // 商品ID，添加索引
	Quantity  int            `json:"quantity" gorm:"not null;default:1"`                        // 商品数量
	CreatedAt time.Time      `json:"created_at"`                                                // 添加时间
	UpdatedAt time.Time      `json:"updated_at"`                                                // 更新时间
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`                                            // 软删除时间
	
	// 关联关系
	User      User           `json:"user,omitempty" gorm:"foreignKey:UserID"`                   // 所属用户
	Product   Product        `json:"product,omitempty" gorm:"foreignKey:ProductID"`             // 关联商品
}

// TableName 指定表名
// GORM默认会将结构体名转换为复数形式作为表名
// 这里显式指定表名为cart_items
func (CartItem) TableName() string {
	return "cart_items"
}

// AddToCartRequest 添加到购物车请求
type AddToCartRequest struct {
	ProductID uint `json:"product_id" binding:"required"`                    // 商品ID，必填
	Quantity  int  `json:"quantity" binding:"required,min=1,max=999"`        // 数量，必填，1-999
}

// UpdateCartRequest 更新购物车请求
type UpdateCartRequest struct {
	Quantity int `json:"quantity" binding:"required,min=1,max=999"`          // 数量，必填，1-999
}

// CartResponse 购物车响应
// 包含商品的详细信息，便于前端展示
type CartResponse struct {
	ID          uint    `json:"id"`                                           // 购物车项ID
	ProductID   uint    `json:"product_id"`                                   // 商品ID
	ProductName string  `json:"product_name"`                                 // 商品名称
	ProductImage *string `json:"product_image"`                               // 商品图片
	Price       float64 `json:"price"`                                        // 商品价格
	Quantity    int     `json:"quantity"`                                     // 数量
	TotalPrice  float64 `json:"total_price"`                                  // 小计金额
	Stock       int     `json:"stock"`                                        // 库存数量
	CreatedAt   time.Time `json:"created_at"`                                 // 添加时间
}

// CartListResponse 购物车列表响应
type CartListResponse struct {
	Items      []*CartResponse `json:"items"`                                 // 购物车商品列表
	TotalCount int             `json:"total_count"`                           // 商品种类数量
	TotalPrice float64         `json:"total_price"`                           // 总金额
}

// CartSummary 购物车汇总信息
type CartSummary struct {
	UserID      uint    `json:"user_id"`      // 用户ID
	ItemCount   int     `json:"item_count"`   // 商品种类数量
	TotalItems  int     `json:"total_items"`  // 商品总数量
	TotalAmount float64 `json:"total_amount"` // 总金额
}
