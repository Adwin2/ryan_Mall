package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims JWT声明结构体
// 包含用户信息和标准JWT声明
type Claims struct {
	UserID   uint   `json:"user_id"`   // 用户ID
	Username string `json:"username"`  // 用户名
	Email    string `json:"email"`     // 邮箱
	jwt.RegisteredClaims                // 标准JWT声明（过期时间、签发者等）
}

// JWTManager JWT管理器
// 负责生成和验证JWT令牌
type JWTManager struct {
	secretKey   string        // 签名密钥
	expireHours int           // 过期时间（小时）
}

// NewJWTManager 创建JWT管理器
// secretKey: 用于签名的密钥，生产环境中应该使用强随机字符串
// expireHours: 令牌过期时间（小时）
func NewJWTManager(secretKey string, expireHours int) *JWTManager {
	return &JWTManager{
		secretKey:   secretKey,
		expireHours: expireHours,
	}
}

// GenerateToken 生成JWT令牌
// 根据用户信息生成包含用户身份的JWT令牌
func (j *JWTManager) GenerateToken(userID uint, username, email string) (string, error) {
	// 1. 设置过期时间
	// 从当前时间开始计算，添加指定的小时数
	expirationTime := time.Now().Add(time.Duration(j.expireHours) * time.Hour)
	
	// 2. 创建声明
	// 包含用户信息和标准JWT声明
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Email:    email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime), // 过期时间
			IssuedAt:  jwt.NewNumericDate(time.Now()),      // 签发时间
			NotBefore: jwt.NewNumericDate(time.Now()),      // 生效时间
			Issuer:    "ryan-mall",                         // 签发者
			Subject:   username,                            // 主题（通常是用户标识）
		},
	}
	
	// 3. 创建令牌
	// 使用HS256算法（HMAC with SHA-256）
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	
	// 4. 签名令牌
	// 使用密钥对令牌进行签名，生成最终的JWT字符串
	tokenString, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", err
	}
	
	return tokenString, nil
}

// ValidateToken 验证JWT令牌
// 解析并验证JWT令牌，返回用户声明信息
func (j *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	// 1. 解析令牌
	// 使用密钥验证签名并解析声明
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		// 确保使用的是我们期望的HMAC算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(j.secretKey), nil
	})
	
	if err != nil {
		return nil, err
	}
	
	// 2. 验证令牌有效性
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	
	// 3. 提取声明
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}
	
	return claims, nil
}

// RefreshToken 刷新JWT令牌
// 基于现有的有效令牌生成新的令牌（延长过期时间）
func (j *JWTManager) RefreshToken(tokenString string) (string, error) {
	// 1. 验证现有令牌
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}
	
	// 2. 检查令牌是否即将过期
	// 只有在令牌还有效但即将过期时才允许刷新
	if time.Until(claims.ExpiresAt.Time) > time.Hour {
		return "", errors.New("token is not eligible for refresh")
	}
	
	// 3. 生成新令牌
	// 使用相同的用户信息生成新的令牌
	return j.GenerateToken(claims.UserID, claims.Username, claims.Email)
}

// ExtractUserID 从令牌中提取用户ID
// 这是一个便捷方法，用于快速获取用户ID
func (j *JWTManager) ExtractUserID(tokenString string) (uint, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}
