package domain

import (
	"time"

	"github.com/google/uuid"
)

// ID 通用ID类型
type ID string

// NewID 生成新的ID
func NewID() ID {
	return ID(uuid.New().String())
}

// String 转换为字符串
func (id ID) String() string {
	return string(id)
}

// IsEmpty 检查ID是否为空
func (id ID) IsEmpty() bool {
	return string(id) == ""
}

// UserID 用户ID
type UserID ID

// NewUserID 生成新的用户ID
func NewUserID() UserID {
	return UserID(NewID())
}

// String 转换为字符串
func (id UserID) String() string {
	return string(id)
}

// ProductID 商品ID
type ProductID ID

// NewProductID 生成新的商品ID
func NewProductID() ProductID {
	return ProductID(NewID())
}

// String 转换为字符串
func (id ProductID) String() string {
	return string(id)
}

// OrderID 订单ID
type OrderID ID

// NewOrderID 生成新的订单ID
func NewOrderID() OrderID {
	return OrderID(NewID())
}

// String 转换为字符串
func (id OrderID) String() string {
	return string(id)
}

// Money 金额值对象
type Money struct {
	Amount   int64  `json:"amount"`   // 以分为单位存储
	Currency string `json:"currency"` // 货币类型
}

// NewMoney 创建金额对象
func NewMoney(amount int64, currency string) Money {
	if currency == "" {
		currency = "CNY" // 默认人民币
	}
	return Money{
		Amount:   amount,
		Currency: currency,
	}
}

// NewMoneyFromYuan 从元创建金额对象
func NewMoneyFromYuan(yuan float64, currency string) Money {
	return NewMoney(int64(yuan*100), currency)
}

// ToYuan 转换为元
func (m Money) ToYuan() float64 {
	return float64(m.Amount) / 100
}

// Add 加法
func (m Money) Add(other Money) Money {
	if m.Currency != other.Currency {
		panic("cannot add money with different currencies")
	}
	return Money{
		Amount:   m.Amount + other.Amount,
		Currency: m.Currency,
	}
}

// Subtract 减法
func (m Money) Subtract(other Money) Money {
	if m.Currency != other.Currency {
		panic("cannot subtract money with different currencies")
	}
	return Money{
		Amount:   m.Amount - other.Amount,
		Currency: m.Currency,
	}
}

// Multiply 乘法
func (m Money) Multiply(factor int64) Money {
	return Money{
		Amount:   m.Amount * factor,
		Currency: m.Currency,
	}
}

// IsZero 是否为零
func (m Money) IsZero() bool {
	return m.Amount == 0
}

// IsPositive 是否为正数
func (m Money) IsPositive() bool {
	return m.Amount > 0
}

// IsNegative 是否为负数
func (m Money) IsNegative() bool {
	return m.Amount < 0
}

// Equal 比较相等
func (m Money) Equal(other Money) bool {
	return m.Amount == other.Amount && m.Currency == other.Currency
}

// GreaterThan 大于比较
func (m Money) GreaterThan(other Money) bool {
	if m.Currency != other.Currency {
		panic("cannot compare money with different currencies")
	}
	return m.Amount > other.Amount
}

// LessThan 小于比较
func (m Money) LessThan(other Money) bool {
	if m.Currency != other.Currency {
		panic("cannot compare money with different currencies")
	}
	return m.Amount < other.Amount
}

// Email 邮箱值对象
type Email struct {
	value string
}

// NewEmail 创建邮箱对象
func NewEmail(email string) (Email, error) {
	if err := validateEmail(email); err != nil {
		return Email{}, err
	}
	return Email{value: email}, nil
}

// String 转换为字符串
func (e Email) String() string {
	return e.value
}

// validateEmail 验证邮箱格式
func validateEmail(email string) error {
	// 简单的邮箱验证，实际项目中应该使用更严格的验证
	if len(email) < 5 || !contains(email, "@") || !contains(email, ".") {
		return NewValidationError("invalid email format")
	}
	return nil
}

// Phone 手机号值对象
type Phone struct {
	value string
}

// NewPhone 创建手机号对象
func NewPhone(phone string) (Phone, error) {
	if err := validatePhone(phone); err != nil {
		return Phone{}, err
	}
	return Phone{value: phone}, nil
}

// String 转换为字符串
func (p Phone) String() string {
	return p.value
}

// validatePhone 验证手机号格式
func validatePhone(phone string) error {
	// 简单的手机号验证
	if len(phone) != 11 {
		return NewValidationError("phone number must be 11 digits")
	}
	return nil
}

// Timestamp 时间戳值对象
type Timestamp struct {
	value time.Time
}

// NewTimestamp 创建时间戳对象
func NewTimestamp(t time.Time) Timestamp {
	return Timestamp{value: t}
}

// Now 创建当前时间戳
func Now() Timestamp {
	return Timestamp{value: time.Now()}
}

// Time 获取时间
func (t Timestamp) Time() time.Time {
	return t.value
}

// Unix 获取Unix时间戳
func (t Timestamp) Unix() int64 {
	return t.value.Unix()
}

// Format 格式化时间
func (t Timestamp) Format(layout string) string {
	return t.value.Format(layout)
}

// Before 是否在指定时间之前
func (t Timestamp) Before(other Timestamp) bool {
	return t.value.Before(other.value)
}

// After 是否在指定时间之后
func (t Timestamp) After(other Timestamp) bool {
	return t.value.After(other.value)
}

// Equal 时间是否相等
func (t Timestamp) Equal(other Timestamp) bool {
	return t.value.Equal(other.value)
}

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
