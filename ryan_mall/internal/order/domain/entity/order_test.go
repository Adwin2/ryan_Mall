package entity

import (
	"testing"

	"ryan-mall-microservices/internal/shared/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrder_Create(t *testing.T) {
	userID := domain.NewUserID()
	
	// 创建订单项
	items := []OrderItemData{
		{
			ProductID: domain.NewProductID(),
			Quantity:  2,
			Price:     domain.NewMoneyFromYuan(999.00, "CNY"),
		},
		{
			ProductID: domain.NewProductID(),
			Quantity:  1,
			Price:     domain.NewMoneyFromYuan(1999.00, "CNY"),
		},
	}

	tests := []struct {
		name        string
		userID      domain.UserID
		items       []OrderItemData
		expectError bool
		errorType   domain.ErrorCode
	}{
		{
			name:        "valid order creation",
			userID:      userID,
			items:       items,
			expectError: false,
		},
		{
			name:        "empty user ID",
			userID:      domain.UserID(""),
			items:       items,
			expectError: true,
			errorType:   domain.ErrCodeValidation,
		},
		{
			name:        "empty items",
			userID:      userID,
			items:       []OrderItemData{},
			expectError: true,
			errorType:   domain.ErrCodeValidation,
		},
		{
			name:   "invalid item quantity",
			userID: userID,
			items: []OrderItemData{
				{
					ProductID: domain.NewProductID(),
					Quantity:  0,
					Price:     domain.NewMoneyFromYuan(999.00, "CNY"),
				},
			},
			expectError: true,
			errorType:   domain.ErrCodeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order, err := NewOrder(tt.userID, tt.items)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, order)
				if tt.errorType != "" {
					assert.Equal(t, tt.errorType, domain.GetErrorCode(err))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, order)
				assert.Equal(t, tt.userID, order.UserID())
				assert.Equal(t, len(tt.items), len(order.Items()))
				assert.Equal(t, OrderStatusPending, order.Status())
				assert.NotEmpty(t, order.ID())
				
				// 验证总金额计算
				expectedTotal := domain.NewMoney(0, "CNY")
				for _, item := range tt.items {
					itemTotal := item.Price.Multiply(int64(item.Quantity))
					expectedTotal = expectedTotal.Add(itemTotal)
				}
				assert.True(t, order.TotalAmount().Equal(expectedTotal))
			}
		})
	}
}

func TestOrder_Confirm(t *testing.T) {
	// 创建订单
	userID := domain.NewUserID()
	items := []OrderItemData{
		{
			ProductID: domain.NewProductID(),
			Quantity:  2,
			Price:     domain.NewMoneyFromYuan(999.00, "CNY"),
		},
	}
	
	order, err := NewOrder(userID, items)
	require.NoError(t, err)
	require.NotNil(t, order)

	// 确认订单
	err = order.Confirm()
	assert.NoError(t, err)
	assert.Equal(t, OrderStatusConfirmed, order.Status())

	// 重复确认应该失败
	err = order.Confirm()
	assert.Error(t, err)
	assert.Equal(t, domain.ErrCodeValidation, domain.GetErrorCode(err))
}

func TestOrder_Cancel(t *testing.T) {
	// 创建订单
	userID := domain.NewUserID()
	items := []OrderItemData{
		{
			ProductID: domain.NewProductID(),
			Quantity:  2,
			Price:     domain.NewMoneyFromYuan(999.00, "CNY"),
		},
	}
	
	order, err := NewOrder(userID, items)
	require.NoError(t, err)
	require.NotNil(t, order)

	tests := []struct {
		name        string
		setupStatus OrderStatus
		reason      string
		expectError bool
		errorType   domain.ErrorCode
	}{
		{
			name:        "cancel pending order",
			setupStatus: OrderStatusPending,
			reason:      "用户取消",
			expectError: false,
		},
		{
			name:        "cancel confirmed order",
			setupStatus: OrderStatusConfirmed,
			reason:      "库存不足",
			expectError: false,
		},
		{
			name:        "cancel completed order should fail",
			setupStatus: OrderStatusCompleted,
			reason:      "测试",
			expectError: true,
			errorType:   domain.ErrCodeValidation,
		},
		{
			name:        "cancel already cancelled order should fail",
			setupStatus: OrderStatusCancelled,
			reason:      "测试",
			expectError: true,
			errorType:   domain.ErrCodeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重新创建订单以避免状态污染
			testOrder, _ := NewOrder(userID, items)
			
			// 设置初始状态
			switch tt.setupStatus {
			case OrderStatusConfirmed:
				testOrder.Confirm()
			case OrderStatusCompleted:
				testOrder.Confirm()
				testOrder.Complete()
			case OrderStatusCancelled:
				testOrder.Cancel("预设取消")
			}

			err := testOrder.Cancel(tt.reason)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType != "" {
					assert.Equal(t, tt.errorType, domain.GetErrorCode(err))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, OrderStatusCancelled, testOrder.Status())
				assert.Equal(t, tt.reason, testOrder.CancelReason())
			}
		})
	}
}

func TestOrder_Complete(t *testing.T) {
	// 创建订单
	userID := domain.NewUserID()
	items := []OrderItemData{
		{
			ProductID: domain.NewProductID(),
			Quantity:  2,
			Price:     domain.NewMoneyFromYuan(999.00, "CNY"),
		},
	}
	
	order, err := NewOrder(userID, items)
	require.NoError(t, err)
	require.NotNil(t, order)

	// 未确认的订单不能完成
	err = order.Complete()
	assert.Error(t, err)
	assert.Equal(t, domain.ErrCodeValidation, domain.GetErrorCode(err))

	// 确认订单后可以完成
	err = order.Confirm()
	require.NoError(t, err)

	err = order.Complete()
	assert.NoError(t, err)
	assert.Equal(t, OrderStatusCompleted, order.Status())

	// 重复完成应该失败
	err = order.Complete()
	assert.Error(t, err)
	assert.Equal(t, domain.ErrCodeValidation, domain.GetErrorCode(err))
}

func TestOrder_UpdateShippingAddress(t *testing.T) {
	// 创建订单
	userID := domain.NewUserID()
	items := []OrderItemData{
		{
			ProductID: domain.NewProductID(),
			Quantity:  2,
			Price:     domain.NewMoneyFromYuan(999.00, "CNY"),
		},
	}
	
	order, err := NewOrder(userID, items)
	require.NoError(t, err)
	require.NotNil(t, order)

	address := "北京市朝阳区xxx街道xxx号"

	// 待处理状态可以更新地址
	err = order.UpdateShippingAddress(address)
	assert.NoError(t, err)
	assert.Equal(t, address, order.ShippingAddress())

	// 已完成的订单不能更新地址
	order.Confirm()
	order.Complete()
	
	err = order.UpdateShippingAddress("新地址")
	assert.Error(t, err)
	assert.Equal(t, domain.ErrCodeValidation, domain.GetErrorCode(err))
}

func TestOrder_DomainEvents(t *testing.T) {
	// 创建订单
	userID := domain.NewUserID()
	items := []OrderItemData{
		{
			ProductID: domain.NewProductID(),
			Quantity:  2,
			Price:     domain.NewMoneyFromYuan(999.00, "CNY"),
		},
	}
	
	order, err := NewOrder(userID, items)
	require.NoError(t, err)
	require.NotNil(t, order)

	// 检查是否有订单创建事件
	events := order.DomainEvents()
	assert.Len(t, events, 1)

	// 检查事件类型
	event := events[0]
	assert.Equal(t, "order.created", event.EventType())
	assert.Equal(t, order.ID().String(), event.AggregateID())
	assert.Equal(t, "Order", event.AggregateType())

	// 确认订单会产生新事件
	order.Confirm()
	events = order.DomainEvents()
	assert.Len(t, events, 2)
	assert.Equal(t, "order.confirmed", events[1].EventType())

	// 清除事件
	order.ClearDomainEvents()
	events = order.DomainEvents()
	assert.Len(t, events, 0)
}
