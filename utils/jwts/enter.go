package jwts

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

// 定义jwt的token结构体

type JwtPayLoad struct {
	UserID   uint   `json:"userid"`
	NickName string `json:"nickname"`
	Role     int8   `json:"role"`
}

// RegisteredClaims 是一个结构体，用于定义JWT（JSON Web Token）中注册的声明。
// 它是jwt标准中预定义的一组声明，可以被包含在JWT中，并且可以被用来作为验证JWT有效性的依据。
// 这个结构体扩展了jwt.StandardClaims，增加了额外的预注册声明字段。

type CustomClaims struct {
	JwtPayLoad
	jwt.RegisteredClaims
}

// 创建token

// GenerateToken 生成一个JWT Token
// @param payload JwtPayLoad 结构体，包含Token的负载信息
// @param accessSecret string，用于签名Token的密钥
// @param expires int，Token的过期时间（小时）
// @return string，生成的JWT Token字符串
// @return error，生成过程中遇到的任何错误
func GenerateToken(payload JwtPayLoad, accessSecret string, expires int) (string, error) {
	// 创建自定义声明，结合了JwtPayLoad和标准的jwt.RegisteredClaims
	claim := CustomClaims{
		JwtPayLoad: payload,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(expires))), // 设置Token过期时间
		},
	}
	// 使用HS256算法和声明创建一个新的Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	// 签名Token并返回
	return token.SignedString([]byte(accessSecret))
}

// 解析token

// ParseToken 函数用于解析JWT令牌并验证其有效性。
// 参数 tokenString 是待解析的JWT令牌字符串。
// 参数 accessSecret 是用于验证令牌签名的密钥。
// 返回值 *CustomClaims 是解析后的自定义声明对象，如果解析失败或令牌无效，则返回error。
func ParseToken(tokenString string, accessSecret string) (*CustomClaims, error) {
	// 使用指定的密钥和自定义声明类型解析令牌
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(accessSecret), nil // 返回验证签名的密钥
	})
	if err != nil {
		return nil, err // 解析失败，返回错误
	}

	// 验证令牌的有效性，并断言claims的类型为*CustomClaims
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil // 令牌有效，返回自定义声明
	}

	// 令牌无效，返回错误
	return nil, errors.New("token is invalid")
}
