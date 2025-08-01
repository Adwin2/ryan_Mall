package entity

import (
	"testing"
	"time"

	"ryan-mall-microservices/internal/shared/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSeckillActivity_Create(t *testing.T) {
	now := time.Now()
	startTime := now.Add(time.Hour)
	endTime := now.Add(2 * time.Hour)
	
	tests := []struct {
		name        string
		activityName string
		productID   domain.ProductID
		originalPrice domain.Money
		seckillPrice  domain.Money
		totalStock    int
		startTime     time.Time
		endTime       time.Time
		expectError   bool
		errorType     domain.ErrorCode
	}{
		{
			name:          "valid seckill activity creation",
			activityName:  "iPhone 15 秒杀",
			productID:     domain.NewProductID(),
			originalPrice: domain.NewMoneyFromYuan(6999.00, "CNY"),
			seckillPrice:  domain.NewMoneyFromYuan(5999.00, "CNY"),
			totalStock:    100,
			startTime:     startTime,
			endTime:       endTime,
			expectError:   false,
		},
		{
			name:          "empty activity name",
			activityName:  "",
			productID:     domain.NewProductID(),
			originalPrice: domain.NewMoneyFromYuan(6999.00, "CNY"),
			seckillPrice:  domain.NewMoneyFromYuan(5999.00, "CNY"),
			totalStock:    100,
			startTime:     startTime,
			endTime:       endTime,
			expectError:   true,
			errorType:     domain.ErrCodeValidation,
		},
		{
			name:          "seckill price higher than original price",
			activityName:  "iPhone 15 秒杀",
			productID:     domain.NewProductID(),
			originalPrice: domain.NewMoneyFromYuan(5999.00, "CNY"),
			seckillPrice:  domain.NewMoneyFromYuan(6999.00, "CNY"),
			totalStock:    100,
			startTime:     startTime,
			endTime:       endTime,
			expectError:   true,
			errorType:     domain.ErrCodeValidation,
		},
		{
			name:          "zero stock",
			activityName:  "iPhone 15 秒杀",
			productID:     domain.NewProductID(),
			originalPrice: domain.NewMoneyFromYuan(6999.00, "CNY"),
			seckillPrice:  domain.NewMoneyFromYuan(5999.00, "CNY"),
			totalStock:    0,
			startTime:     startTime,
			endTime:       endTime,
			expectError:   true,
			errorType:     domain.ErrCodeValidation,
		},
		{
			name:          "end time before start time",
			activityName:  "iPhone 15 秒杀",
			productID:     domain.NewProductID(),
			originalPrice: domain.NewMoneyFromYuan(6999.00, "CNY"),
			seckillPrice:  domain.NewMoneyFromYuan(5999.00, "CNY"),
			totalStock:    100,
			startTime:     endTime,
			endTime:       startTime,
			expectError:   true,
			errorType:     domain.ErrCodeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			activity, err := NewSeckillActivity(
				tt.activityName,
				tt.productID,
				tt.originalPrice,
				tt.seckillPrice,
				tt.totalStock,
				tt.startTime,
				tt.endTime,
			)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, activity)
				if tt.errorType != "" {
					assert.Equal(t, tt.errorType, domain.GetErrorCode(err))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, activity)
				assert.Equal(t, tt.activityName, activity.Name())
				assert.Equal(t, tt.productID, activity.ProductID())
				assert.True(t, activity.OriginalPrice().Equal(tt.originalPrice))
				assert.True(t, activity.SeckillPrice().Equal(tt.seckillPrice))
				assert.Equal(t, tt.totalStock, activity.TotalStock())
				assert.Equal(t, tt.totalStock, activity.RemainingStock())
				assert.Equal(t, SeckillStatusPending, activity.Status())
				assert.NotEmpty(t, activity.ID())
			}
		})
	}
}

func TestSeckillActivity_Start(t *testing.T) {
	// 创建秒杀活动
	now := time.Now()
	activity, err := NewSeckillActivity(
		"iPhone 15 秒杀",
		domain.NewProductID(),
		domain.NewMoneyFromYuan(6999.00, "CNY"),
		domain.NewMoneyFromYuan(5999.00, "CNY"),
		100,
		now.Add(time.Hour),
		now.Add(2*time.Hour),
	)
	require.NoError(t, err)
	require.NotNil(t, activity)

	// 启动秒杀活动
	err = activity.Start()
	assert.NoError(t, err)
	assert.Equal(t, SeckillStatusActive, activity.Status())

	// 重复启动应该失败
	err = activity.Start()
	assert.Error(t, err)
	assert.Equal(t, domain.ErrCodeValidation, domain.GetErrorCode(err))
}

func TestSeckillActivity_End(t *testing.T) {
	// 创建并启动秒杀活动
	now := time.Now()
	activity, err := NewSeckillActivity(
		"iPhone 15 秒杀",
		domain.NewProductID(),
		domain.NewMoneyFromYuan(6999.00, "CNY"),
		domain.NewMoneyFromYuan(5999.00, "CNY"),
		100,
		now.Add(time.Hour),
		now.Add(2*time.Hour),
	)
	require.NoError(t, err)
	require.NotNil(t, activity)

	err = activity.Start()
	require.NoError(t, err)

	// 结束秒杀活动
	err = activity.End()
	assert.NoError(t, err)
	assert.Equal(t, SeckillStatusEnded, activity.Status())

	// 重复结束应该失败
	err = activity.End()
	assert.Error(t, err)
	assert.Equal(t, domain.ErrCodeValidation, domain.GetErrorCode(err))
}

func TestSeckillActivity_ReserveStock(t *testing.T) {
	// 创建并启动秒杀活动
	now := time.Now()
	activity, err := NewSeckillActivity(
		"iPhone 15 秒杀",
		domain.NewProductID(),
		domain.NewMoneyFromYuan(6999.00, "CNY"),
		domain.NewMoneyFromYuan(5999.00, "CNY"),
		10,
		now.Add(time.Hour),
		now.Add(2*time.Hour),
	)
	require.NoError(t, err)
	require.NotNil(t, activity)

	err = activity.Start()
	require.NoError(t, err)

	tests := []struct {
		name        string
		quantity    int
		expectError bool
		errorType   domain.ErrorCode
	}{
		{
			name:        "valid stock reservation",
			quantity:    5,
			expectError: false,
		},
		{
			name:        "insufficient stock",
			quantity:    10, // 剩余5个，请求10个
			expectError: true,
			errorType:   domain.ErrCodeInsufficientStock,
		},
		{
			name:        "zero quantity",
			quantity:    0,
			expectError: true,
			errorType:   domain.ErrCodeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalStock := activity.RemainingStock()
			err := activity.ReserveStock(tt.quantity)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType != "" {
					assert.Equal(t, tt.errorType, domain.GetErrorCode(err))
				}
				// 库存不应该改变
				assert.Equal(t, originalStock, activity.RemainingStock())
			} else {
				assert.NoError(t, err)
				// 库存应该减少
				assert.Equal(t, originalStock-tt.quantity, activity.RemainingStock())
			}
		})
	}
}

func TestSeckillActivity_IsActive(t *testing.T) {
	now := time.Now()
	
	tests := []struct {
		name       string
		status     SeckillStatus
		startTime  time.Time
		endTime    time.Time
		checkTime  time.Time
		expected   bool
	}{
		{
			name:      "active status and within time range",
			status:    SeckillStatusActive,
			startTime: now.Add(-time.Hour),
			endTime:   now.Add(time.Hour),
			checkTime: now,
			expected:  true,
		},
		{
			name:      "active status but before start time",
			status:    SeckillStatusActive,
			startTime: now.Add(time.Hour),
			endTime:   now.Add(2*time.Hour),
			checkTime: now,
			expected:  false,
		},
		{
			name:      "active status but after end time",
			status:    SeckillStatusActive,
			startTime: now.Add(-2*time.Hour),
			endTime:   now.Add(-time.Hour),
			checkTime: now,
			expected:  false,
		},
		{
			name:      "pending status",
			status:    SeckillStatusPending,
			startTime: now.Add(-time.Hour),
			endTime:   now.Add(time.Hour),
			checkTime: now,
			expected:  false,
		},
		{
			name:      "ended status",
			status:    SeckillStatusEnded,
			startTime: now.Add(-time.Hour),
			endTime:   now.Add(time.Hour),
			checkTime: now,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			activity, err := NewSeckillActivity(
				"Test Activity",
				domain.NewProductID(),
				domain.NewMoneyFromYuan(100.00, "CNY"),
				domain.NewMoneyFromYuan(80.00, "CNY"),
				10,
				tt.startTime,
				tt.endTime,
			)
			require.NoError(t, err)

			// 设置状态
			if tt.status == SeckillStatusActive {
				activity.Start()
			} else if tt.status == SeckillStatusEnded {
				activity.Start()
				activity.End()
			}

			result := activity.IsActive(tt.checkTime)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSeckillActivity_DomainEvents(t *testing.T) {
	// 创建秒杀活动
	now := time.Now()
	activity, err := NewSeckillActivity(
		"iPhone 15 秒杀",
		domain.NewProductID(),
		domain.NewMoneyFromYuan(6999.00, "CNY"),
		domain.NewMoneyFromYuan(5999.00, "CNY"),
		100,
		now.Add(time.Hour),
		now.Add(2*time.Hour),
	)
	require.NoError(t, err)
	require.NotNil(t, activity)

	// 检查是否有秒杀活动创建事件
	events := activity.DomainEvents()
	assert.Len(t, events, 1)
	assert.Equal(t, "seckill.created", events[0].EventType())

	// 启动活动会产生新事件
	activity.Start()
	events = activity.DomainEvents()
	assert.Len(t, events, 2)
	assert.Equal(t, "seckill.started", events[1].EventType())

	// 清除事件
	activity.ClearDomainEvents()
	events = activity.DomainEvents()
	assert.Len(t, events, 0)
}
