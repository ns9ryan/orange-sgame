//*******************************************************************
// funcTokenRelated.go里，放的是对一些公共函数，包括redis/mysql的相关查询等。
//*******************************************************************

package apifunc

import (
    "videogame/common/model"
    "fmt"
    "videogame/common/global/variable"
    "time"
    "encoding/json"
    "github.com/golang-jwt/jwt/v5"
    "errors"
    "math/rand"
)

//以下是商户的Token的生成和验证函数
const (
    Charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789" // 基本字符集
    TokenExpiryDuration = 30 * 24 * time.Hour
)

// 生成商户密钥merchantKey，32字节长字符串
func GenerateMerchantKey() string {
    rand.Seed(time.Now().UnixNano())
    saltLength := 32

    salt := make([]byte, saltLength)
    for i := range salt {
        salt[i] = Charset[rand.Intn(len(Charset))]
    }

    return string(salt)
}

// 生成 merchantToken（JWT），有效期 30 天
func GenerateMerchantToken(merchantCode string, merchantKey []byte) (string, error) {
    claims := jwt.MapClaims{
        "merchant_code": merchantCode,
        "ext":           time.Now().Add(TokenExpiryDuration).Unix(), // 自定义过期字段
        "iat":           time.Now().Unix(),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(merchantKey)
}

// 验证 merchantToken 并返回 merchantCode
func ValidateMerchantToken(merchantToken string, merchantKey []byte) (string, error) {
    token, err := jwt.Parse(merchantToken, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("无效的签名方法")
        }
        return merchantKey, nil
    })
    if err != nil {
        return "", err
    }

    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        // 验证过期时间 ext
        if _, ok := claims["ext"].(float64); ok {
            // if time.Now().Unix() > int64(expUnix) {
            //     return "", errors.New("token 已过期")
            // }
        } else {
            return "", errors.New("ext 字段不存在或格式错误")
        }

        // 读取 merchant_code
        if merchantCode, ok := claims["merchant_code"].(string); ok {
            return merchantCode, nil
        }
        return "", errors.New("merchant_code 不存在或格式错误")
    }

    return "", errors.New("token 无效")
}

// 对商户OrToken和OrTokenKey的示例测试
// func main() {
//     merchantCode := "VEZUX77S"
//     merchantKey := GenerateMerchantKey() //("k3H9tN1Qz6VbXyM8LcWdRaEp0jYgUoZm") // 32位商户token key

//     fmt.Println("生成的 merchantKey:", merchantKey)

//     merchantToken, err := GenerateMerchantToken(merchantCode, []byte(merchantKey))
//     if err != nil {
//         fmt.Println("生成 token 错误:", err)
//         return
//     }
//     fmt.Println("生成的 merchantToken:", merchantToken)

//     resultCode, err := ValidateMerchantToken(merchantToken, []byte(merchantKey))
//     if err != nil {
//         fmt.Println("验证 token 错误:", err)
//         return
//     }
//     fmt.Println("验证成功，merchantCode:", resultCode)
// }

// 函数：保存用户的token和相关信息到Redis的函数调用
// 目前先在/fastGame/v2和/forwardGame/v1之后调用，后续移到LobbyService登录成功那里
func saveTokenToRedis(token string, user *models.SUserPlayer, ttl time.Duration) error {
    redisCon := variable.RedisStorage

    // userData, _ := json.Marshal(req.BetRoulette)
    // key := fmt.Sprintf("%d:%s", gm.GSName, gm.Id, gm.Set, gm.Round)
    // err = redisCon.HashSet(key, req.Uid, userData)
    // if err != nil {
    //     return err
    // }

    userDataJSON, err := json.Marshal(user)
    if err != nil {
        return fmt.Errorf("Marshal失败: %v", err)
    }

    err = redisCon.Set("token:"+token, userDataJSON, ttl)          // 存储token对应的用户信息
    if err != nil {
        return err
    }

    err = redisCon.Set("user:"+fmt.Sprint(user.ID), token, ttl)    // 存储用户对应的 token
    if err != nil {
        return err
    }

    return nil
}

// 函数：验证token是否有效，并获取用户信息
func validateTokenFromRedis(token string) (*models.SUserPlayer, error) {
    redisCon := variable.RedisStorage
    userData, err := redisCon.Get("token:"+token)
    if err != nil {
        return nil, err
    }

    // 解析存储的用户信息
    var user *models.SUserPlayer
    if err := json.Unmarshal([]byte(userData), &user); err != nil {
        return nil, fmt.Errorf("invalid user data")
    }

    // 校验玩家是否被强制下线
    currentToken, err := redisCon.Get("user:"+fmt.Sprint(user.ID))
    if err == nil && currentToken != token {
        return nil, fmt.Errorf("token invalid due to logout or overwrite")
    }

    return user, nil
}