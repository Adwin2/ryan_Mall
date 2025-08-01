package domain

import (
	"fmt"
)

// ErrorCode 错误码类型
type ErrorCode string

const (
	// 通用错误码
	ErrCodeValidation     ErrorCode = "VALIDATION_ERROR"
	ErrCodeNotFound       ErrorCode = "NOT_FOUND"
	ErrCodeAlreadyExists  ErrorCode = "ALREADY_EXISTS"
	ErrCodeUnauthorized   ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden      ErrorCode = "FORBIDDEN"
	ErrCodeInternalError  ErrorCode = "INTERNAL_ERROR"
	ErrCodeBadRequest     ErrorCode = "BAD_REQUEST"
	ErrCodeConflict       ErrorCode = "CONFLICT"

	// 业务错误码
	ErrCodeInsufficientStock ErrorCode = "INSUFFICIENT_STOCK"
	ErrCodeOrderNotFound     ErrorCode = "ORDER_NOT_FOUND"
	ErrCodeUserNotFound      ErrorCode = "USER_NOT_FOUND"
	ErrCodeProductNotFound   ErrorCode = "PRODUCT_NOT_FOUND"
	ErrCodePaymentFailed     ErrorCode = "PAYMENT_FAILED"
	ErrCodeSeckillEnded      ErrorCode = "SECKILL_ENDED"
	ErrCodeSeckillNotStarted ErrorCode = "SECKILL_NOT_STARTED"
	ErrCodeLimitExceeded     ErrorCode = "LIMIT_EXCEEDED"
	ErrCodeTooManyRequests   ErrorCode = "TOO_MANY_REQUESTS"
)

// DomainError 领域错误接口
type DomainError interface {
	error
	Code() ErrorCode
	Message() string
	Details() map[string]interface{}
}

// BaseError 基础错误实现
type BaseError struct {
	code    ErrorCode
	message string
	details map[string]interface{}
	cause   error
}

// NewBaseError 创建基础错误
func NewBaseError(code ErrorCode, message string) *BaseError {
	return &BaseError{
		code:    code,
		message: message,
		details: make(map[string]interface{}),
	}
}

// Error 实现error接口
func (e *BaseError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.code, e.message, e.cause)
	}
	return fmt.Sprintf("%s: %s", e.code, e.message)
}

// Code 获取错误码
func (e *BaseError) Code() ErrorCode {
	return e.code
}

// Message 获取错误消息
func (e *BaseError) Message() string {
	return e.message
}

// Details 获取错误详情
func (e *BaseError) Details() map[string]interface{} {
	return e.details
}

// WithDetail 添加错误详情
func (e *BaseError) WithDetail(key string, value interface{}) *BaseError {
	e.details[key] = value
	return e
}

// WithCause 添加原因错误
func (e *BaseError) WithCause(cause error) *BaseError {
	e.cause = cause
	return e
}

// Unwrap 解包错误
func (e *BaseError) Unwrap() error {
	return e.cause
}

// ValidationError 验证错误
type ValidationError struct {
	*BaseError
	field string
	value interface{}
}

// NewValidationError 创建验证错误
func NewValidationError(message string) *ValidationError {
	return &ValidationError{
		BaseError: NewBaseError(ErrCodeValidation, message),
	}
}

// NewFieldValidationError 创建字段验证错误
func NewFieldValidationError(field string, value interface{}, message string) *ValidationError {
	err := &ValidationError{
		BaseError: NewBaseError(ErrCodeValidation, message),
		field:     field,
		value:     value,
	}
	err.WithDetail("field", field).WithDetail("value", value)
	return err
}

// Field 获取字段名
func (e *ValidationError) Field() string {
	return e.field
}

// Value 获取字段值
func (e *ValidationError) Value() interface{} {
	return e.value
}

// NotFoundError 未找到错误
type NotFoundError struct {
	*BaseError
	resource string
	id       string
}

// NewNotFoundError 创建未找到错误
func NewNotFoundError(resource, id string) *NotFoundError {
	message := fmt.Sprintf("%s with id '%s' not found", resource, id)
	err := &NotFoundError{
		BaseError: NewBaseError(ErrCodeNotFound, message),
		resource:  resource,
		id:        id,
	}
	err.WithDetail("resource", resource).WithDetail("id", id)
	return err
}

// Resource 获取资源类型
func (e *NotFoundError) Resource() string {
	return e.resource
}

// ID 获取资源ID
func (e *NotFoundError) ID() string {
	return e.id
}

// AlreadyExistsError 已存在错误
type AlreadyExistsError struct {
	*BaseError
	resource string
	field    string
	value    string
}

// NewAlreadyExistsError 创建已存在错误
func NewAlreadyExistsError(resource, field, value string) *AlreadyExistsError {
	message := fmt.Sprintf("%s with %s '%s' already exists", resource, field, value)
	err := &AlreadyExistsError{
		BaseError: NewBaseError(ErrCodeAlreadyExists, message),
		resource:  resource,
		field:     field,
		value:     value,
	}
	err.WithDetail("resource", resource).WithDetail("field", field).WithDetail("value", value)
	return err
}

// Resource 获取资源类型
func (e *AlreadyExistsError) Resource() string {
	return e.resource
}

// Field 获取字段名
func (e *AlreadyExistsError) Field() string {
	return e.field
}

// Value 获取字段值
func (e *AlreadyExistsError) Value() string {
	return e.value
}

// BusinessError 业务错误
type BusinessError struct {
	*BaseError
}

// NewBusinessError 创建业务错误
func NewBusinessError(code ErrorCode, message string) *BusinessError {
	return &BusinessError{
		BaseError: NewBaseError(code, message),
	}
}

// InsufficientStockError 库存不足错误
func NewInsufficientStockError(productID string, requested, available int) *BusinessError {
	message := fmt.Sprintf("insufficient stock for product %s: requested %d, available %d", 
		productID, requested, available)
	err := NewBusinessError(ErrCodeInsufficientStock, message)
	err.WithDetail("product_id", productID)
	err.WithDetail("requested", requested)
	err.WithDetail("available", available)
	return err
}

// UnauthorizedError 未授权错误
func NewUnauthorizedError(message string) *BaseError {
	if message == "" {
		message = "unauthorized access"
	}
	return NewBaseError(ErrCodeUnauthorized, message)
}

// ForbiddenError 禁止访问错误
func NewForbiddenError(message string) *BaseError {
	if message == "" {
		message = "access forbidden"
	}
	return NewBaseError(ErrCodeForbidden, message)
}

// InternalError 内部错误
func NewInternalError(message string, cause error) *BaseError {
	if message == "" {
		message = "internal server error"
	}
	err := NewBaseError(ErrCodeInternalError, message)
	if cause != nil {
		err.WithCause(cause)
	}
	return err
}

// IsErrorCode 检查错误是否为指定错误码
func IsErrorCode(err error, code ErrorCode) bool {
	if domainErr, ok := err.(DomainError); ok {
		return domainErr.Code() == code
	}
	return false
}

// GetErrorCode 获取错误码
func GetErrorCode(err error) ErrorCode {
	if domainErr, ok := err.(DomainError); ok {
		return domainErr.Code()
	}
	return ErrCodeInternalError
}
