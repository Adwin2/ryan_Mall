package entity

import (
	"testing"

	"ryan-mall-microservices/internal/shared/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProduct_Create(t *testing.T) {
	tests := []struct {
		name        string
		productName string
		description string
		categoryID  string
		price       domain.Money
		stock       int
		expectError bool
		errorType   domain.ErrorCode
	}{
		{
			name:        "valid product creation",
			productName: "iPhone 15",
			description: "Latest iPhone model",
			categoryID:  "cat-123",
			price:       domain.NewMoneyFromYuan(6999.00, "CNY"),
			stock:       100,
			expectError: false,
		},
		{
			name:        "empty product name",
			productName: "",
			description: "Description",
			categoryID:  "cat-123",
			price:       domain.NewMoneyFromYuan(100.00, "CNY"),
			stock:       10,
			expectError: true,
			errorType:   domain.ErrCodeValidation,
		},
		{
			name:        "negative price",
			productName: "Product",
			description: "Description",
			categoryID:  "cat-123",
			price:       domain.NewMoney(-100, "CNY"),
			stock:       10,
			expectError: true,
			errorType:   domain.ErrCodeValidation,
		},
		{
			name:        "negative stock",
			productName: "Product",
			description: "Description",
			categoryID:  "cat-123",
			price:       domain.NewMoneyFromYuan(100.00, "CNY"),
			stock:       -1,
			expectError: true,
			errorType:   domain.ErrCodeValidation,
		},
		{
			name:        "empty category ID",
			productName: "Product",
			description: "Description",
			categoryID:  "",
			price:       domain.NewMoneyFromYuan(100.00, "CNY"),
			stock:       10,
			expectError: true,
			errorType:   domain.ErrCodeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product, err := NewProduct(tt.productName, tt.description, tt.categoryID, tt.price, tt.stock)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, product)
				if tt.errorType != "" {
					assert.Equal(t, tt.errorType, domain.GetErrorCode(err))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, product)
				assert.Equal(t, tt.productName, product.Name())
				assert.Equal(t, tt.description, product.Description())
				assert.Equal(t, tt.categoryID, product.CategoryID())
				assert.True(t, product.Price().Equal(tt.price))
				assert.Equal(t, tt.stock, product.Stock())
				assert.NotEmpty(t, product.ID())
				assert.True(t, product.IsAvailable())
			}
		})
	}
}

func TestProduct_UpdatePrice(t *testing.T) {
	// 创建商品
	product, err := NewProduct("iPhone 15", "Latest iPhone", "cat-123", domain.NewMoneyFromYuan(6999.00, "CNY"), 100)
	require.NoError(t, err)
	require.NotNil(t, product)

	tests := []struct {
		name        string
		newPrice    domain.Money
		expectError bool
		errorType   domain.ErrorCode
	}{
		{
			name:        "valid price update",
			newPrice:    domain.NewMoneyFromYuan(7999.00, "CNY"),
			expectError: false,
		},
		{
			name:        "negative price",
			newPrice:    domain.NewMoney(-100, "CNY"),
			expectError: true,
			errorType:   domain.ErrCodeValidation,
		},
		{
			name:        "zero price",
			newPrice:    domain.NewMoney(0, "CNY"),
			expectError: true,
			errorType:   domain.ErrCodeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := product.UpdatePrice(tt.newPrice)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType != "" {
					assert.Equal(t, tt.errorType, domain.GetErrorCode(err))
				}
			} else {
				assert.NoError(t, err)
				assert.True(t, product.Price().Equal(tt.newPrice))
			}
		})
	}
}

func TestProduct_UpdateStock(t *testing.T) {
	// 创建商品
	product, err := NewProduct("iPhone 15", "Latest iPhone", "cat-123", domain.NewMoneyFromYuan(6999.00, "CNY"), 100)
	require.NoError(t, err)
	require.NotNil(t, product)

	tests := []struct {
		name        string
		newStock    int
		expectError bool
		errorType   domain.ErrorCode
	}{
		{
			name:        "valid stock update",
			newStock:    200,
			expectError: false,
		},
		{
			name:        "zero stock",
			newStock:    0,
			expectError: false,
		},
		{
			name:        "negative stock",
			newStock:    -1,
			expectError: true,
			errorType:   domain.ErrCodeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := product.UpdateStock(tt.newStock)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType != "" {
					assert.Equal(t, tt.errorType, domain.GetErrorCode(err))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.newStock, product.Stock())
			}
		})
	}
}

func TestProduct_ReserveStock(t *testing.T) {
	// 创建商品
	product, err := NewProduct("iPhone 15", "Latest iPhone", "cat-123", domain.NewMoneyFromYuan(6999.00, "CNY"), 100)
	require.NoError(t, err)
	require.NotNil(t, product)

	tests := []struct {
		name        string
		quantity    int
		expectError bool
		errorType   domain.ErrorCode
	}{
		{
			name:        "valid stock reservation",
			quantity:    10,
			expectError: false,
		},
		{
			name:        "insufficient stock",
			quantity:    200,
			expectError: true,
			errorType:   domain.ErrCodeInsufficientStock,
		},
		{
			name:        "zero quantity",
			quantity:    0,
			expectError: true,
			errorType:   domain.ErrCodeValidation,
		},
		{
			name:        "negative quantity",
			quantity:    -1,
			expectError: true,
			errorType:   domain.ErrCodeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalStock := product.Stock()
			err := product.ReserveStock(tt.quantity)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType != "" {
					assert.Equal(t, tt.errorType, domain.GetErrorCode(err))
				}
				// 库存不应该改变
				assert.Equal(t, originalStock, product.Stock())
			} else {
				assert.NoError(t, err)
				// 库存应该减少
				assert.Equal(t, originalStock-tt.quantity, product.Stock())
			}
		})
	}
}

func TestProduct_ReleaseStock(t *testing.T) {
	// 创建商品
	product, err := NewProduct("iPhone 15", "Latest iPhone", "cat-123", domain.NewMoneyFromYuan(6999.00, "CNY"), 100)
	require.NoError(t, err)
	require.NotNil(t, product)

	// 先预留一些库存
	err = product.ReserveStock(20)
	require.NoError(t, err)

	tests := []struct {
		name        string
		quantity    int
		expectError bool
		errorType   domain.ErrorCode
	}{
		{
			name:        "valid stock release",
			quantity:    10,
			expectError: false,
		},
		{
			name:        "zero quantity",
			quantity:    0,
			expectError: true,
			errorType:   domain.ErrCodeValidation,
		},
		{
			name:        "negative quantity",
			quantity:    -1,
			expectError: true,
			errorType:   domain.ErrCodeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalStock := product.Stock()
			err := product.ReleaseStock(tt.quantity)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType != "" {
					assert.Equal(t, tt.errorType, domain.GetErrorCode(err))
				}
				// 库存不应该改变
				assert.Equal(t, originalStock, product.Stock())
			} else {
				assert.NoError(t, err)
				// 库存应该增加
				assert.Equal(t, originalStock+tt.quantity, product.Stock())
			}
		})
	}
}

func TestProduct_SetAvailability(t *testing.T) {
	// 创建商品
	product, err := NewProduct("iPhone 15", "Latest iPhone", "cat-123", domain.NewMoneyFromYuan(6999.00, "CNY"), 100)
	require.NoError(t, err)
	require.NotNil(t, product)

	// 商品应该是可用的
	assert.True(t, product.IsAvailable())

	// 设置为不可用
	product.SetUnavailable()
	assert.False(t, product.IsAvailable())

	// 设置为可用
	product.SetAvailable()
	assert.True(t, product.IsAvailable())
}

func TestProduct_DomainEvents(t *testing.T) {
	// 创建商品
	product, err := NewProduct("iPhone 15", "Latest iPhone", "cat-123", domain.NewMoneyFromYuan(6999.00, "CNY"), 100)
	require.NoError(t, err)
	require.NotNil(t, product)

	// 检查是否有商品创建事件
	events := product.DomainEvents()
	assert.Len(t, events, 1)

	// 检查事件类型
	event := events[0]
	assert.Equal(t, "product.created", event.EventType())
	assert.Equal(t, product.ID().String(), event.AggregateID())
	assert.Equal(t, "Product", event.AggregateType())

	// 清除事件
	product.ClearDomainEvents()
	events = product.DomainEvents()
	assert.Len(t, events, 0)
}
