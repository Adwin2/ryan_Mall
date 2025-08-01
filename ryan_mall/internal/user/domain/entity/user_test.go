package entity

import (
	"testing"

	"ryan-mall-microservices/internal/shared/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUser_Create(t *testing.T) {
	tests := []struct {
		name        string
		username    string
		email       string
		password    string
		expectError bool
		errorType   domain.ErrorCode
	}{
		{
			name:        "valid user creation",
			username:    "testuser",
			email:       "test@example.com",
			password:    "Password123!",
			expectError: false,
		},
		{
			name:        "empty username",
			username:    "",
			email:       "test@example.com",
			password:    "Password123!",
			expectError: true,
			errorType:   domain.ErrCodeValidation,
		},
		{
			name:        "invalid email",
			username:    "testuser",
			email:       "invalid-email",
			password:    "Password123!",
			expectError: true,
			errorType:   domain.ErrCodeValidation,
		},
		{
			name:        "weak password",
			username:    "testuser",
			email:       "test@example.com",
			password:    "123",
			expectError: true,
			errorType:   domain.ErrCodeValidation,
		},
		{
			name:        "username too short",
			username:    "ab",
			email:       "test@example.com",
			password:    "Password123!",
			expectError: true,
			errorType:   domain.ErrCodeValidation,
		},
		{
			name:        "username too long",
			username:    "this_is_a_very_long_username_that_exceeds_the_maximum_allowed_length",
			email:       "test@example.com",
			password:    "Password123!",
			expectError: true,
			errorType:   domain.ErrCodeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := NewUser(tt.username, tt.email, tt.password)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, user)
				if tt.errorType != "" {
					assert.Equal(t, tt.errorType, domain.GetErrorCode(err))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.username, user.Username())
				assert.Equal(t, tt.email, user.Email().String())
				assert.NotEmpty(t, user.ID())
				assert.True(t, user.IsActive())
				assert.NotEmpty(t, user.CreatedAt())
			}
		})
	}
}

func TestUser_ChangePassword(t *testing.T) {
	// 创建用户
	user, err := NewUser("testuser", "test@example.com", "OldPassword123!")
	require.NoError(t, err)
	require.NotNil(t, user)

	tests := []struct {
		name        string
		oldPassword string
		newPassword string
		expectError bool
		errorType   domain.ErrorCode
	}{
		{
			name:        "valid password change",
			oldPassword: "OldPassword123!",
			newPassword: "NewPassword123!",
			expectError: false,
		},
		{
			name:        "wrong old password",
			oldPassword: "WrongPassword123!",
			newPassword: "NewPassword123!",
			expectError: true,
			errorType:   domain.ErrCodeValidation,
		},
		{
			name:        "weak new password",
			oldPassword: "OldPassword123!",
			newPassword: "123",
			expectError: true,
			errorType:   domain.ErrCodeValidation,
		},
		{
			name:        "same password",
			oldPassword: "OldPassword123!",
			newPassword: "OldPassword123!",
			expectError: true,
			errorType:   domain.ErrCodeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := user.ChangePassword(tt.oldPassword, tt.newPassword)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType != "" {
					assert.Equal(t, tt.errorType, domain.GetErrorCode(err))
				}
			} else {
				assert.NoError(t, err)
				// 验证新密码
				assert.True(t, user.CheckPassword(tt.newPassword))
				assert.False(t, user.CheckPassword(tt.oldPassword))
			}
		})
	}
}

func TestUser_UpdateProfile(t *testing.T) {
	// 创建用户
	user, err := NewUser("testuser", "test@example.com", "Password123!")
	require.NoError(t, err)
	require.NotNil(t, user)

	tests := []struct {
		name        string
		nickname    string
		phone       string
		expectError bool
		errorType   domain.ErrorCode
	}{
		{
			name:        "valid profile update",
			nickname:    "Test User",
			phone:       "13800138000",
			expectError: false,
		},
		{
			name:        "invalid phone",
			nickname:    "Test User",
			phone:       "invalid-phone",
			expectError: true,
			errorType:   domain.ErrCodeValidation,
		},
		{
			name:        "empty nickname",
			nickname:    "",
			phone:       "13800138000",
			expectError: false, // nickname可以为空
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := user.UpdateProfile(tt.nickname, tt.phone)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType != "" {
					assert.Equal(t, tt.errorType, domain.GetErrorCode(err))
				}
			} else {
				assert.NoError(t, err)
				if tt.nickname != "" {
					assert.Equal(t, tt.nickname, user.Profile().Nickname)
				}
				if tt.phone != "" {
					assert.Equal(t, tt.phone, user.Profile().Phone.String())
				}
			}
		})
	}
}

func TestUser_Deactivate(t *testing.T) {
	// 创建用户
	user, err := NewUser("testuser", "test@example.com", "Password123!")
	require.NoError(t, err)
	require.NotNil(t, user)

	// 用户应该是激活状态
	assert.True(t, user.IsActive())

	// 停用用户
	user.Deactivate()

	// 用户应该是停用状态
	assert.False(t, user.IsActive())
}

func TestUser_Activate(t *testing.T) {
	// 创建用户
	user, err := NewUser("testuser", "test@example.com", "Password123!")
	require.NoError(t, err)
	require.NotNil(t, user)

	// 先停用用户
	user.Deactivate()
	assert.False(t, user.IsActive())

	// 激活用户
	user.Activate()

	// 用户应该是激活状态
	assert.True(t, user.IsActive())
}

func TestUser_CheckPassword(t *testing.T) {
	// 创建用户
	user, err := NewUser("testuser", "test@example.com", "Password123!")
	require.NoError(t, err)
	require.NotNil(t, user)

	tests := []struct {
		name     string
		password string
		expected bool
	}{
		{
			name:     "correct password",
			password: "Password123!",
			expected: true,
		},
		{
			name:     "wrong password",
			password: "WrongPassword123!",
			expected: false,
		},
		{
			name:     "empty password",
			password: "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := user.CheckPassword(tt.password)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUser_DomainEvents(t *testing.T) {
	// 创建用户
	user, err := NewUser("testuser", "test@example.com", "Password123!")
	require.NoError(t, err)
	require.NotNil(t, user)

	// 检查是否有用户注册事件
	events := user.DomainEvents()
	assert.Len(t, events, 1)

	// 检查事件类型
	event := events[0]
	assert.Equal(t, "user.registered", event.EventType())
	assert.Equal(t, user.ID().String(), event.AggregateID())
	assert.Equal(t, "User", event.AggregateType())

	// 清除事件
	user.ClearDomainEvents()
	events = user.DomainEvents()
	assert.Len(t, events, 0)
}
