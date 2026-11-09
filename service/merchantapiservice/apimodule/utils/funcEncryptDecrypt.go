//********************************************************************
// funcEncryptDecrypte.go里，放的是对数据加密/解密/签名/验证操作的基础处理函数。
//********************************************************************

package apifunc

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/md5"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "errors"
    "fmt"
    "github.com/duanhf2012/origin/v2/log"
    "time"
    "strings"
    "bytes"
    "encoding/base64"
    "sort"
    "math/rand"
)

// 加密 params 字段的 AES/ECB/PKCS5Padding 加密数据，参考OB Live的AES/ECB模式
func EncryptParams(params interface{}, aesKey []byte) (string, error) {
    // 将 params 转为 JSON 字符串
    data, err := json.Marshal(params)
    if err != nil {
        return "", fmt.Errorf("JSON 序列化失败: %v", err)
    }

    // 填充数据至 AES 的块大小
    data = AddPKCS5Padding(data, aes.BlockSize)

    // 确保 aesKey 长度符合 AES 标准（16, 24, 32 字节）
    key := aesKey
    fmt.Printf("key len=%d\n", len(key))
    if len(key) != 16 && len(key) != 24 && len(key) != 32 {
        return "", errors.New("无效的 AES 密钥长度")
    }

    // 创建 AES 加密器
    block, err := aes.NewCipher(key)
    if err != nil {
        return "", fmt.Errorf("AES 加密器创建失败: %v", err)
    }

    // 使用手动实现的 ECB 模式进行加密
    encryptedBytes, err := AESEncryptECB(block, data)
    if err != nil {
        return "", fmt.Errorf("AES 加密失败: %v", err)
    }

    // 使用 Base64 编码将加密字节转换为字符串
    encryptedParams := base64.StdEncoding.EncodeToString(encryptedBytes)

    return encryptedParams, nil
}

// 实现 AES/ECB 加密
func AESEncryptECB(block cipher.Block, data []byte) ([]byte, error) {
    encrypted := make([]byte, len(data))
    blockSize := block.BlockSize()

    for i := 0; i < len(data); i += blockSize {
        block.Encrypt(encrypted[i:i+blockSize], data[i:i+blockSize])
    }

    return encrypted, nil
}

// 添加 PKCS5Padding
func AddPKCS5Padding(data []byte, blockSize int) []byte {
    padding := blockSize - len(data)%blockSize
    padText := bytes.Repeat([]byte{byte(padding)}, padding)
    return append(data, padText...)
}

// 解密 params 字段的 AES/ECB/PKCS5Padding 加密数据
func DecryptParams(encryptedParams string, aesKey []byte) (interface{}, error) {
    // 将 Base64 编码的字符串解码为字节
    encryptedBytes, err := base64.StdEncoding.DecodeString(encryptedParams)
    if err != nil {
        return nil, fmt.Errorf("Base64 解码失败: %v", err)
    }

    // 确保 aesKey 长度符合 AES 标准（16, 24, 32 字节）
    key := aesKey
    if len(key) != 16 && len(key) != 24 && len(key) != 32 {
        return nil, errors.New("无效的 AES 密钥长度")
    }

    // 创建 AES 解密器
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, fmt.Errorf("AES 解密器创建失败: %v", err)
    }

    // 使用手动实现的 ECB 模式进行解密
    decryptedBytes, err := AESDecryptECB(block, encryptedBytes)
    if err != nil {
        return nil, fmt.Errorf("AES 解密失败: %v", err)
    }

    // 去除 PKCS5Padding
    decryptedBytes = RemovePKCS5Padding(decryptedBytes)

    // 解析为 JSON
    var params interface{}
    err = json.Unmarshal(decryptedBytes, &params)
    if err != nil {
        return nil, fmt.Errorf("JSON 解析失败: %v", err)
    }

    return params, nil
}

// 手动实现 AES/ECB 解密
func AESDecryptECB(block cipher.Block, encryptedData []byte) ([]byte, error) {
    decrypted := make([]byte, len(encryptedData))
    blockSize := block.BlockSize()

    for i := 0; i < len(encryptedData); i += blockSize {
        block.Decrypt(decrypted[i:i+blockSize], encryptedData[i:i+blockSize])
    }

    return decrypted, nil
}

// 去除 PKCS5Padding
func RemovePKCS5Padding(data []byte) []byte {
    length := len(data)
    unpadding := int(data[length-1])
    return data[:(length - unpadding)]
}

// GenerateMerchantCode 生成一个不重复的4位商户代码，保存数据库后返回
func GenerateMerchantCode() (string, string, error) {
    rand.Seed(time.Now().UnixNano()) // 使用当前时间的纳秒作为随机数种子
    codeLength := 8 // 商户代码长度为4位

    var code string
    for {
        var sb strings.Builder
        // 生成4位商户代码
        for i := 0; i < codeLength; i++ {
            sb.WriteByte(Charset[rand.Intn(len(Charset))]) // 随机选择一个字符
        }

        // 确保代码唯一，查询数据库确认商户代码是否已存在，默认大写
        code = strings.ToUpper(sb.String())
        // existingMerchant, err := models.GetMerchantByCode(code)
        // if err != nil {
        //     return "", "", fmt.Errorf("查询商户代码失败: %v", err)
        // }

        // if existingMerchant == nil {
        //     // 商户代码不存在，可以返回
        //     break
        // }
    }

    //保存新商户信息到数据库，这部分后续由管理系统处理，这里只做模拟
    md5Salt := GenerateDynamicSalt()
    // newMerchant := &models.Merchant{
    //     MerchantCode: code,
    //     AeskeySalt:   GenerateDynamicSalt(),
    //     Md5Salt:      md5Salt,
    //     CreatedAt:    time.Now(),              // 设置创建时间
    //     UpdatedAt:    time.Now(),              // 设置更新时间
    // }

    // if err := newMerchant.Create(); err != nil {
    //     return "", "", fmt.Errorf("保存商户信息失败: %v", err)
    // }

    return code, md5Salt, nil // 返回生成的商户代码
}

// 生成动态盐值
func GenerateDynamicSalt() string {
    rand.Seed(time.Now().UnixNano())    // 使用当前时间的纳秒作为随机数种子
    saltLength := 16                    // 16字节长, 动态盐的长度可以根据需求调整

    salt := make([]byte, saltLength)
    for i := range salt {
        salt[i] = Charset[rand.Intn(len(Charset))]
    }

    return string(salt)
}

// 定义全局缓存Map来存储 商户code和AESKey及Md5Salt的映射
var aesKeyCache = make(map[string][]byte)
var md5SaltCache = make(map[string]string)

// GetAESKeyByDynamicSalt 获取商户的 AES 密钥
// 该函数会根据商户号获取数据库中的动态盐值，并生成 AES 密钥
func GetAESKeyByDynamicSalt(merchantCode string) ([]byte, error) {
    // 如果缓存中有对应的密钥，则直接返回
    if key, exists := aesKeyCache[merchantCode]; exists {
        return key, nil
    }

    // 从数据库中获取商户的动态Aeskey盐值
    // merchant, err := models.GetMerchantByCode(merchantCode)
    // if err != nil || merchant == nil {
    //     log.Error("GetMerchantByCode错误:", log.ErrorField("err", err))
    //     return nil, err
    // }

    // 使用商户号和动态盐值生成 AES 密钥
    dynamicAesSalt := "4Zfk3KEpwVald6U6" //merchant.AeskeySalt
    hash := sha256.New()
    hash.Write([]byte(merchantCode + dynamicAesSalt))   // 商户号 + 动态盐值
    aesKey := hash.Sum(nil)                             // 直接返回 32 字节的哈希值

    // 使用 Base64 编码将加密字节转换为字符串，仅用于这里打印输出测试
    base64AesKey := base64.StdEncoding.EncodeToString(aesKey)
    fmt.Printf("get merchantCode:%s, aesKey:%s\n", merchantCode, base64AesKey)

    aesKeyCache[merchantCode] = aesKey
    return aesKey, nil
}

func GenerateSignature(merchantCode string, data []byte) (error, string) {
    // 如果缓存中有对应的密钥，则直接返回
    var md5Salt string
    if key, exists := md5SaltCache[merchantCode]; exists {
        md5Salt = key
    }

    // 从数据库中获取商户的动态Aeskey盐值
    // merchant, err := models.GetMerchantByCode(merchantCode)
    // if err != nil || merchant == nil {
    //     log.Error("GetMerchantByCode错误:", log.ErrorField("err", err))
    //     return err, ""
    // }

    md5Salt = "ZhynnyBs3Ug3239R" //merchant.Md5Salt

    // 生成 MD5 签名
    md5Hash := md5.New()
    combinedData := append(data, []byte(md5Salt)...)
    md5Hash.Write(combinedData)  // 直接对 data 进行 MD5 计算
    signature := strings.ToUpper(hex.EncodeToString(md5Hash.Sum(nil)))

    return nil, signature
}

// 验证 MD5 签名
func ValidateSignature(merchantCode string, originJson interface{}, requestSignature string) bool {
    // 如果缓存中有对应的密钥，则直接返回
    var md5Salt string
    if key, exists := md5SaltCache[merchantCode]; exists {
        md5Salt = key
    }

    // 从数据库中获取商户的动态Aeskey盐值
    // merchant, err := models.GetMerchantByCode(merchantCode)
    // if err != nil || merchant == nil {
    //     log.Error("GetMerchantByCode错误:", log.ErrorField("err", err))
    //     return false
    // }

    md5Salt = "ZhynnyBs3Ug3239R" //merchant.Md5Salt

    // 将请求数据转换为 map，并对键进行排序
    rawMap, err := StructToSortedMap(originJson)
    if err != nil {
        log.Error("参数处理失败:", log.ErrorField("err", err))
        return false
    }

    // 序列化请求数据并生成 MD5 签名
    data, err := json.Marshal(rawMap)
    if err != nil {
        log.Error("参数序列化失败:", log.ErrorField("err", err))
        return false
    }

    // 生成 MD5 签名
    md5Hash := md5.New()
    combinedData := append(data, []byte(md5Salt)...)
    md5Hash.Write(combinedData)
    calculatedSignature := strings.ToUpper(hex.EncodeToString(md5Hash.Sum(nil)))
    fmt.Printf("calculatedSignature=%s\n", calculatedSignature)

    // 比较签名是否匹配
    md5SaltCache[merchantCode] = md5Salt
    return calculatedSignature == requestSignature
}

func StructToSortedMap(input interface{}) (map[string]interface{}, error) {
    // 将结构体转换为 JSON 字符串
    jsonData, err := json.Marshal(input)
    if err != nil {
        return nil, err
    }

    // 解析 JSON 字符串为 map
    var rawMap map[string]interface{}
    if err := json.Unmarshal(jsonData, &rawMap); err != nil {
        return nil, err
    }

    // 对 map 的键进行排序
    sortedMap := make(map[string]interface{})
    keys := make([]string, 0, len(rawMap))
    for k := range rawMap {
        keys = append(keys, k)
    }
    sort.Strings(keys)

    for _, k := range keys {
        sortedMap[k] = rawMap[k]
    }

    return sortedMap, nil
}
