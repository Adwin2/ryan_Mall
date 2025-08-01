package entity

import (
	"testing"

	"ryan-mall-microservices/internal/shared/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPayment_Create(t *testing.T) {
	userID := domain.NewUserID()
	orderID := domain.NewOrderID()
	amount := domain.NewMoneyFromYuan(999.00, "CNY")

	tests := []struct {
		name        string
		userID      domain.UserID
		orderID     domain.OrderID
		amount      domain.Money
		method      PaymentMethod
		expectError bool
		errorType   domain.ErrorCode
	}{
		{
			name:        "valid payment creation",
			userID:      userID,
			orderID:     orderID,
			amount:      amount,
			method:      PaymentMethodAlipay,
			expectError: false,
		},
		{
			name:        "empty user ID",
			userID:      domain.UserID(""),
			orderID:     orderID,
			amount:      amount,
			method:      PaymentMethodAlipay,
			expectError: true,
			errorType:   domain.ErrCodeValidation,
		},
		{
			name:        "empty order ID",
			userID:      userID,
			orderID:     domain.OrderID(""),
			amount:      amount,
			method:      PaymentMethodAlipay,
			expectError: true,
			errorType:   domain.ErrCodeValidation,
		},
		{
			name:        "zero amount",
			userID:      userID,
			orderID:     orderID,
			amount:      domain.NewMoney(0, "CNY"),
			method:      PaymentMethodAlipay,
			expectError: true,
			errorType:   domain.ErrCodeValidation,
		},
		{
			name:        "negative amount",
			userID:      userID,
			orderID:     orderID,
			amount:      domain.NewMoney(-100, "CNY"),
			method:      PaymentMethodAlipay,
			expectError: true,
			errorType:   domain.ErrCodeValidation,
		},
		{
			name:        "invalid payment method",
			userID:      userID,
			orderID:     orderID,
			amount:      amount,
			method:      PaymentMethod("INVALID"),
			expectError: true,
			errorType:   domain.ErrCodeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payment, err := NewPayment(tt.userID, tt.orderID, tt.amount, tt.method)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, payment)
				if tt.errorType != "" {
					assert.Equal(t, tt.errorType, domain.GetErrorCode(err))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, payment)
				assert.Equal(t, tt.userID, payment.UserID())
				assert.Equal(t, tt.orderID, payment.OrderID())
				assert.True(t, payment.Amount().Equal(tt.amount))
				assert.Equal(t, tt.method, payment.Method())
				assert.Equal(t, PaymentStatusPending, payment.Status())
				assert.NotEmpty(t, payment.ID())
			}
		})
	}
}

func TestPayment_Process(t *testing.T) {
	// 创建支付
	payment, err := NewPayment(
		domain.NewUserID(),
		domain.NewOrderID(),
		domain.NewMoneyFromYuan(999.00, "CNY"),
		PaymentMethodAlipay,
	)
	require.NoError(t, err)
	require.NotNil(t, payment)

	// 处理支付
	transactionID := "alipay_txn_123456"
	err = payment.Process(transactionID)
	assert.NoError(t, err)
	assert.Equal(t, PaymentStatusProcessing, payment.Status())
	assert.Equal(t, transactionID, payment.TransactionID())

	// 重复处理应该失败
	err = payment.Process("another_txn")
	assert.Error(t, err)
	assert.Equal(t, domain.ErrCodeValidation, domain.GetErrorCode(err))
}

func TestPayment_Complete(t *testing.T) {
	// 创建并处理支付
	payment, err := NewPayment(
		domain.NewUserID(),
		domain.NewOrderID(),
		domain.NewMoneyFromYuan(999.00, "CNY"),
		PaymentMethodAlipay,
	)
	require.NoError(t, err)
	require.NotNil(t, payment)

	err = payment.Process("alipay_txn_123456")
	require.NoError(t, err)

	// 完成支付
	err = payment.Complete()
	assert.NoError(t, err)
	assert.Equal(t, PaymentStatusCompleted, payment.Status())

	// 重复完成应该失败
	err = payment.Complete()
	assert.Error(t, err)
	assert.Equal(t, domain.ErrCodeValidation, domain.GetErrorCode(err))
}

func TestPayment_Fail(t *testing.T) {
	// 创建并处理支付
	payment, err := NewPayment(
		domain.NewUserID(),
		domain.NewOrderID(),
		domain.NewMoneyFromYuan(999.00, "CNY"),
		PaymentMethodAlipay,
	)
	require.NoError(t, err)
	require.NotNil(t, payment)

	err = payment.Process("alipay_txn_123456")
	require.NoError(t, err)

	// 支付失败
	failureReason := "insufficient balance"
	err = payment.Fail(failureReason)
	assert.NoError(t, err)
	assert.Equal(t, PaymentStatusFailed, payment.Status())
	assert.Equal(t, failureReason, payment.FailureReason())

	// 重复失败应该失败
	err = payment.Fail("another reason")
	assert.Error(t, err)
	assert.Equal(t, domain.ErrCodeValidation, domain.GetErrorCode(err))
}

func TestPayment_Cancel(t *testing.T) {
	// 创建支付
	payment, err := NewPayment(
		domain.NewUserID(),
		domain.NewOrderID(),
		domain.NewMoneyFromYuan(999.00, "CNY"),
		PaymentMethodAlipay,
	)
	require.NoError(t, err)
	require.NotNil(t, payment)

	tests := []struct {
		name        string
		setupStatus PaymentStatus
		expectError bool
		errorType   domain.ErrorCode
	}{
		{
			name:        "cancel pending payment",
			setupStatus: PaymentStatusPending,
			expectError: false,
		},
		{
			name:        "cancel processing payment",
			setupStatus: PaymentStatusProcessing,
			expectError: false,
		},
		{
			name:        "cancel completed payment should fail",
			setupStatus: PaymentStatusCompleted,
			expectError: true,
			errorType:   domain.ErrCodeValidation,
		},
		{
			name:        "cancel failed payment should fail",
			setupStatus: PaymentStatusFailed,
			expectError: true,
			errorType:   domain.ErrCodeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重新创建支付以避免状态污染
			testPayment, _ := NewPayment(
				domain.NewUserID(),
				domain.NewOrderID(),
				domain.NewMoneyFromYuan(999.00, "CNY"),
				PaymentMethodAlipay,
			)

			// 设置初始状态
			switch tt.setupStatus {
			case PaymentStatusProcessing:
				testPayment.Process("txn_123")
			case PaymentStatusCompleted:
				testPayment.Process("txn_123")
				testPayment.Complete()
			case PaymentStatusFailed:
				testPayment.Process("txn_123")
				testPayment.Fail("test failure")
			}

			err := testPayment.Cancel()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType != "" {
					assert.Equal(t, tt.errorType, domain.GetErrorCode(err))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, PaymentStatusCancelled, testPayment.Status())
			}
		})
	}
}

func TestPayment_Refund(t *testing.T) {
	// 创建并完成支付
	payment, err := NewPayment(
		domain.NewUserID(),
		domain.NewOrderID(),
		domain.NewMoneyFromYuan(999.00, "CNY"),
		PaymentMethodAlipay,
	)
	require.NoError(t, err)
	require.NotNil(t, payment)

	err = payment.Process("alipay_txn_123456")
	require.NoError(t, err)

	err = payment.Complete()
	require.NoError(t, err)

	tests := []struct {
		name         string
		refundAmount domain.Money
		refundReason string
		expectError  bool
		errorType    domain.ErrorCode
	}{
		{
			name:         "valid full refund",
			refundAmount: domain.NewMoneyFromYuan(999.00, "CNY"),
			refundReason: "customer request",
			expectError:  false,
		},
		{
			name:         "valid partial refund",
			refundAmount: domain.NewMoneyFromYuan(500.00, "CNY"),
			refundReason: "partial return",
			expectError:  false,
		},
		{
			name:         "refund amount exceeds payment amount",
			refundAmount: domain.NewMoneyFromYuan(1500.00, "CNY"),
			refundReason: "test",
			expectError:  true,
			errorType:    domain.ErrCodeValidation,
		},
		{
			name:         "zero refund amount",
			refundAmount: domain.NewMoney(0, "CNY"),
			refundReason: "test",
			expectError:  true,
			errorType:    domain.ErrCodeValidation,
		},
		{
			name:         "empty refund reason",
			refundAmount: domain.NewMoneyFromYuan(100.00, "CNY"),
			refundReason: "",
			expectError:  true,
			errorType:    domain.ErrCodeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重新创建并完成支付
			testPayment, _ := NewPayment(
				domain.NewUserID(),
				domain.NewOrderID(),
				domain.NewMoneyFromYuan(999.00, "CNY"),
				PaymentMethodAlipay,
			)
			testPayment.Process("txn_123")
			testPayment.Complete()

			refundID := "refund_123"
			err := testPayment.Refund(refundID, tt.refundAmount, tt.refundReason)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType != "" {
					assert.Equal(t, tt.errorType, domain.GetErrorCode(err))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, PaymentStatusRefunded, testPayment.Status())
				assert.Equal(t, refundID, testPayment.RefundID())
				assert.True(t, testPayment.RefundAmount().Equal(tt.refundAmount))
				assert.Equal(t, tt.refundReason, testPayment.RefundReason())
			}
		})
	}
}

func TestPayment_DomainEvents(t *testing.T) {
	// 创建支付
	payment, err := NewPayment(
		domain.NewUserID(),
		domain.NewOrderID(),
		domain.NewMoneyFromYuan(999.00, "CNY"),
		PaymentMethodAlipay,
	)
	require.NoError(t, err)
	require.NotNil(t, payment)

	// 检查是否有支付创建事件
	events := payment.DomainEvents()
	assert.Len(t, events, 1)
	assert.Equal(t, "payment.created", events[0].EventType())

	// 处理支付会产生新事件
	payment.Process("txn_123")
	events = payment.DomainEvents()
	assert.Len(t, events, 2)
	assert.Equal(t, "payment.processing", events[1].EventType())

	// 完成支付会产生新事件
	payment.Complete()
	events = payment.DomainEvents()
	assert.Len(t, events, 3)
	assert.Equal(t, "payment.completed", events[2].EventType())

	// 清除事件
	payment.ClearDomainEvents()
	events = payment.DomainEvents()
	assert.Len(t, events, 0)
}
