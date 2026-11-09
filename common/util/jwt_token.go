package util

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"time"
	"videogame/common/global/variable"
)

// 以下是用户Token的生成和验证函数
const (
	JwtKey              = "k2DgPaOyMwZUZocvz5UmLKIWn3gemPWx" // 32位密钥字符串
	TokenExpiryDuration = 24 * time.Hour                     // 1天失效

	RedisKeyUserToken = "Or_User_Token:" // token保存到redis里的key
)

// 生成 userToken（JWT），有效期 1 天
func GenerateUserToken(username string) (string, error) {
	claims := jwt.MapClaims{
		"username": username,
		"ext":      time.Now().Add(TokenExpiryDuration).Unix(), // 自定义过期字段
		"iat":      time.Now().Unix(),                          // 签发时间
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(JwtKey))
}

// 验证 userToken 并返回 username
func ValidateUserToken(userToken string) (string, error) {

	token, err := jwt.Parse(userToken, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("无效的签名方法")
		}
		return []byte(JwtKey), nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {

		// 验证过期时间 ext
		if expUnix, ok := claims["ext"].(float64); ok {
			if time.Now().Unix() > int64(expUnix) {
				return "", errors.New("token 已过期")
			}
		} else {
			return "", errors.New("ext 字段不存在或格式错误")
		}

		// 读取 username
		if username, ok := claims["username"].(string); ok {
			return username, nil
		}

		return "", errors.New("username 不存在或格式错误")
	}

	return "", errors.New("token 无效")
}

// 暂时没用，需考虑多设备登录/踢出用户等场景
// SaveTokenToRedis 保存用户token和username到Redis
func SaveTokenToRedis(token string, username string) error {
	err := variable.RedisStorage.HSet(RedisKeyUserToken, username, token)
	if err != nil {
		return fmt.Errorf("保存token到redis失败: %v", err)
	}
	return nil
}

// 暂时没用，需考虑多设备登录/踢出用户等场景
// ValidateTokenFromRedis 验证token是否存在，并返回对应的username
func ValidateTokenFromRedis(username string) (string, error) {
	token, err := variable.RedisStorage.HGet(RedisKeyUserToken, username)
	if err != nil {
		return "", fmt.Errorf("验证用户失败，Redis查询失败: %v", err)
	}

	if token == "" {
		return "", fmt.Errorf("token无效或已过期")
	}
	return token, nil
}
