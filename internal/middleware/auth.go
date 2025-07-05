package middleware

import (
	"net/http"
	"ryan-mall/internal/service"
	"ryan-mall/pkg/response"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 认证中间件
// 用于验证JWT令牌并提取用户信息
type AuthMiddleware struct {
	userService service.UserService // 用户服务，用于验证令牌
}

// NewAuthMiddleware 创建认证中间件实例
func NewAuthMiddleware(userService service.UserService) *AuthMiddleware {
	return &AuthMiddleware{
		userService: userService,
	}
}

// RequireAuth 需要认证的中间件
// 验证请求头中的JWT令牌，如果有效则继续处理，否则返回401错误
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 从请求头获取Authorization字段
		// 标准格式：Authorization: Bearer <token>
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "缺少认证令牌")
			c.Abort() // 终止请求处理
			return
		}
		
		// 2. 解析Bearer令牌
		// 检查是否以"Bearer "开头
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			response.Unauthorized(c, "认证令牌格式错误")
			c.Abort()
			return
		}
		
		// 提取令牌字符串（去掉"Bearer "前缀）
		tokenString := authHeader[len(bearerPrefix):]
		if tokenString == "" {
			response.Unauthorized(c, "认证令牌为空")
			c.Abort()
			return
		}
		
		// 3. 验证令牌
		claims, err := m.userService.ValidateToken(tokenString)
		if err != nil {
			response.Unauthorized(c, "认证令牌无效: "+err.Error())
			c.Abort()
			return
		}
		
		// 4. 将用户信息存储到上下文中
		// 后续的处理器可以通过c.Get()获取用户信息
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("claims", claims)
		
		// 5. 继续处理请求
		c.Next()
	}
}

// OptionalAuth 可选认证的中间件
// 如果有令牌则验证并设置用户信息，没有令牌也允许继续处理
// 适用于某些接口既支持游客访问也支持用户访问的场景
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 从请求头获取Authorization字段
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// 没有令牌，继续处理（作为游客）
			c.Next()
			return
		}
		
		// 2. 解析Bearer令牌
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			// 令牌格式错误，继续处理（作为游客）
			c.Next()
			return
		}
		
		tokenString := authHeader[len(bearerPrefix):]
		if tokenString == "" {
			// 令牌为空，继续处理（作为游客）
			c.Next()
			return
		}
		
		// 3. 验证令牌
		claims, err := m.userService.ValidateToken(tokenString)
		if err != nil {
			// 令牌无效，继续处理（作为游客）
			c.Next()
			return
		}
		
		// 4. 令牌有效，设置用户信息
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("claims", claims)
		
		// 5. 继续处理请求
		c.Next()
	}
}

// GetCurrentUserID 从上下文中获取当前用户ID
// 这是一个辅助函数，用于在处理器中获取用户ID
func GetCurrentUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	
	id, ok := userID.(uint)
	return id, ok
}

// GetCurrentUsername 从上下文中获取当前用户名
func GetCurrentUsername(c *gin.Context) (string, bool) {
	username, exists := c.Get("username")
	if !exists {
		return "", false
	}
	
	name, ok := username.(string)
	return name, ok
}

// GetCurrentUserEmail 从上下文中获取当前用户邮箱
func GetCurrentUserEmail(c *gin.Context) (string, bool) {
	email, exists := c.Get("email")
	if !exists {
		return "", false
	}
	
	emailStr, ok := email.(string)
	return emailStr, ok
}

// RequireRole 需要特定角色的中间件（扩展功能）
// 这是一个示例，展示如何实现基于角色的访问控制
// 在实际项目中，可能需要在用户模型中添加角色字段
func (m *AuthMiddleware) RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 首先需要认证
		m.RequireAuth()(c)
		
		// 如果认证失败，上面的中间件已经处理了
		if c.IsAborted() {
			return
		}
		
		// TODO: 实现角色验证逻辑
		// 1. 从数据库获取用户角色
		// 2. 检查用户是否具有所需角色
		// 3. 如果没有权限，返回403错误
		
		// 示例代码（需要根据实际角色系统实现）:
		/*
		userID, _ := GetCurrentUserID(c)
		userRoles := getUserRoles(userID) // 需要实现这个函数
		
		hasPermission := false
		for _, role := range roles {
			if contains(userRoles, role) {
				hasPermission = true
				break
			}
		}
		
		if !hasPermission {
			response.Forbidden(c, "权限不足")
			c.Abort()
			return
		}
		*/
		
		c.Next()
	}
}

// CORS 跨域中间件
// 处理跨域请求，允许前端应用访问API
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")
		
		// 设置CORS响应头
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization, Cache-Control, X-File-Name")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")
		
		// 处理预检请求
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		
		c.Next()
	}
}
