package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构体
// 这样设计的好处是前端可以统一处理响应格式
type Response struct {
	Code    int         `json:"code"`    // 业务状态码
	Message string      `json:"message"` // 响应消息
	Data    interface{} `json:"data"`    // 响应数据
}

// 定义常用的业务状态码
const (
	SUCCESS = 200  // 成功
	ERROR   = 500  // 服务器内部错误
	INVALID_PARAMS = 400  // 请求参数错误
	UNAUTHORIZED = 401    // 未授权
	FORBIDDEN = 403       // 禁止访问
	NOT_FOUND = 404       // 资源不存在
)

// Success 成功响应
// 当业务逻辑执行成功时调用
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    SUCCESS,
		Message: "success",
		Data:    data,
	})
}

// SuccessWithMessage 带自定义消息的成功响应
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    SUCCESS,
		Message: message,
		Data:    data,
	})
}

// Error 错误响应
// 当业务逻辑执行失败时调用
func Error(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// ErrorWithData 带数据的错误响应
func ErrorWithData(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

// BadRequest 请求参数错误
func BadRequest(c *gin.Context, message string) {
	Error(c, INVALID_PARAMS, message)
}

// Unauthorized 未授权
func Unauthorized(c *gin.Context, message string) {
	Error(c, UNAUTHORIZED, message)
}

// Forbidden 禁止访问
func Forbidden(c *gin.Context, message string) {
	Error(c, FORBIDDEN, message)
}

// NotFound 资源不存在
func NotFound(c *gin.Context, message string) {
	Error(c, NOT_FOUND, message)
}

// InternalServerError 服务器内部错误
func InternalServerError(c *gin.Context, message string) {
	Error(c, ERROR, message)
}
