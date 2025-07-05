package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// OrderStatus 订单状态类型
type OrderStatus int

// Order 订单模型
type Order struct {
	ID              uint           `json:"id" gorm:"primaryKey"`                                   // 订单ID
	OrderNo         string         `json:"order_no" gorm:"uniqueIndex;size:32;not null"`          // 订单号，唯一
	UserID          uint           `json:"user_id" gorm:"not null;index"`                         // 用户ID
	TotalAmount     float64        `json:"total_amount" gorm:"type:decimal(10,2);not null"`       // 订单总金额
	Status          OrderStatus    `json:"status" gorm:"default:1;index"`                         // 订单状态
	PaymentMethod   string         `json:"payment_method" gorm:"size:20;index"`                   // 支付方式
	ContactPhone    string         `json:"contact_phone" gorm:"size:20"`                          // 联系电话
	PaymentTime     *time.Time     `json:"payment_time"`                                          // 支付时间
	ShippingAddress JSONAddress    `json:"shipping_address" gorm:"type:json"`                     // 收货地址
	Remark          *string        `json:"remark" gorm:"size:500"`                                // 订单备注
	CreatedAt       time.Time      `json:"created_at" gorm:"index"`                               // 创建时间
	UpdatedAt       time.Time      `json:"updated_at"`                                            // 更新时间
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`                                        // 软删除时间
	
	// 关联关系
	User            User           `json:"user,omitempty" gorm:"foreignKey:UserID"`               // 所属用户
	OrderItems      []OrderItem    `json:"order_items,omitempty" gorm:"foreignKey:OrderID"`       // 订单商品
}

// OrderItem 订单商品模型
type OrderItem struct {
	ID           uint      `json:"id" gorm:"primaryKey"`                                         // 订单商品ID
	OrderID      uint      `json:"order_id" gorm:"not null;index"`                              // 订单ID
	ProductID    uint      `json:"product_id" gorm:"not null;index"`                            // 商品ID
	ProductName  string    `json:"product_name" gorm:"size:200;not null"`                       // 商品名称（冗余存储）
	ProductImage *string   `json:"product_image" gorm:"size:255"`                               // 商品图片（冗余存储）
	Price        float64   `json:"price" gorm:"type:decimal(10,2);not null"`                    // 商品单价
	Quantity     int       `json:"quantity" gorm:"not null"`                                    // 购买数量
	TotalPrice   float64   `json:"total_price" gorm:"type:decimal(10,2);not null"`              // 小计金额
	CreatedAt    time.Time `json:"created_at"`                                                  // 创建时间
	
	// 关联关系
	Order        Order     `json:"order,omitempty" gorm:"foreignKey:OrderID"`                   // 所属订单
	Product      Product   `json:"product,omitempty" gorm:"foreignKey:ProductID"`               // 关联商品
}

// JSONAddress 地址信息的JSON类型
// 用于存储收货地址的详细信息
type JSONAddress struct {
	Name     string `json:"name"`                                                              // 收货人姓名
	Phone    string `json:"phone"`                                                             // 收货人电话
	Province string `json:"province"`                                                          // 省份
	City     string `json:"city"`                                                              // 城市
	District string `json:"district"`                                                          // 区县
	Address  string `json:"address"`                                                           // 详细地址
	Zipcode  string `json:"zipcode"`                                                           // 邮政编码
}

// Scan 实现sql.Scanner接口，用于从数据库读取JSON数据
func (ja *JSONAddress) Scan(value interface{}) error {
	if value == nil {
		*ja = JSONAddress{}
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into JSONAddress", value)
	}
	
	return json.Unmarshal(bytes, ja)
}

// Value 实现driver.Valuer接口，用于将Go数据写入数据库
func (ja JSONAddress) Value() (driver.Value, error) {
	return json.Marshal(ja)
}

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	CartItemIDs     []uint      `json:"cart_item_ids" binding:"required,min=1"`                 // 购物车商品ID列表
	ShippingAddress JSONAddress `json:"shipping_address" binding:"required"`                    // 收货地址
	PaymentMethod   string      `json:"payment_method" binding:"required"`                      // 支付方式
	ContactPhone    string      `json:"contact_phone" binding:"required"`                       // 联系电话
	Remark          string      `json:"remark"`                                                 // 订单备注
}

// OrderListRequest 订单列表查询请求
type OrderListRequest struct {
	Page      int        `form:"page,default=1" binding:"min=1"`                             // 页码
	PageSize  int        `form:"page_size,default=10" binding:"min=1,max=100"`              // 每页数量
	Status    *OrderStatus `form:"status"`                                                   // 订单状态筛选
	StartDate *time.Time `form:"start_date"`                                                 // 开始日期
	EndDate   *time.Time `form:"end_date"`                                                   // 结束日期
}

// OrderResponse 订单响应
type OrderResponse struct {
	ID              uint           `json:"id"`                                                  // 订单ID
	OrderNo         string         `json:"order_no"`                                            // 订单号
	TotalAmount     float64        `json:"total_amount"`                                        // 订单总金额
	Status          int            `json:"status"`                                              // 订单状态
	StatusText      string         `json:"status_text"`                                         // 订单状态文本
	PaymentMethod   *int           `json:"payment_method"`                                      // 支付方式
	PaymentTime     *time.Time     `json:"payment_time"`                                        // 支付时间
	ShippingAddress JSONAddress    `json:"shipping_address"`                                    // 收货地址
	Remark          *string        `json:"remark"`                                              // 订单备注
	CreatedAt       time.Time      `json:"created_at"`                                          // 创建时间
	OrderItems      []OrderItem    `json:"order_items,omitempty"`                               // 订单商品
}

// OrderListResponse 订单列表响应
type OrderListResponse struct {
	Orders     []*Order `json:"orders"`                                                    // 订单列表
	Total      int      `json:"total"`                                                     // 总数量
	Page       int      `json:"page"`                                                      // 当前页码
	PageSize   int      `json:"page_size"`                                                 // 每页数量
	TotalPages int      `json:"total_pages"`                                               // 总页数
}

// PayOrderRequest 支付订单请求
type PayOrderRequest struct {
	PaymentMethod string `json:"payment_method" binding:"required"`                        // 支付方式
}

// OrderStatistics 订单统计
type OrderStatistics struct {
	UserID          uint    `json:"user_id"`          // 用户ID
	TotalOrders     int     `json:"total_orders"`     // 总订单数
	TotalAmount     float64 `json:"total_amount"`     // 总金额
	PendingCount    int     `json:"pending_count"`    // 待支付订单数
	PaidCount       int     `json:"paid_count"`       // 已支付订单数
	ShippedCount    int     `json:"shipped_count"`    // 已发货订单数
	DeliveredCount  int     `json:"delivered_count"`  // 已送达订单数
	CancelledCount  int     `json:"cancelled_count"`  // 已取消订单数
}

// 订单状态常量
const (
	OrderStatusPending   OrderStatus = 1 // 待支付
	OrderStatusPaid      OrderStatus = 2 // 已支付
	OrderStatusShipped   OrderStatus = 3 // 已发货
	OrderStatusDelivered OrderStatus = 4 // 已送达
	OrderStatusCancelled OrderStatus = 5 // 已取消
)

// GetStatusText 获取订单状态文本
func GetOrderStatusText(status OrderStatus) string {
	switch status {
	case OrderStatusPending:
		return "待支付"
	case OrderStatusPaid:
		return "已支付"
	case OrderStatusShipped:
		return "已发货"
	case OrderStatusDelivered:
		return "已送达"
	case OrderStatusCancelled:
		return "已取消"
	default:
		return "未知状态"
	}
}
