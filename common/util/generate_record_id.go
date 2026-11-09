package util

import (
	"crypto/rand"
	"fmt"
	"github.com/google/uuid"
	"math/big"
	"strconv"
	"time"
)

const RecordIdLength = 24

// 生成指定长度的高强度随机数
func generateSecureRandom(n int) string {
	maxInt := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n)), nil) // 10^n
	randNum, _ := rand.Int(rand.Reader, maxInt)                           // 生成随机数
	return fmt.Sprintf("%0*d", n, randNum)                                // 补足 n 位
}

// 生成唯一 ID（时间戳10位 + 计数器6位 + 随机数8位）
func GenerateUniqueID(length int, uid string) string {
	// 确保最小长度
	if length < 16 {
		length = 16
	}

	// 获取当前时间戳（纳秒级，取前 10 位）
	timestampNano := strconv.FormatInt(time.Now().UnixNano(), 10)[:10]

	// 生成 UUID
	id := uuid.New()

	// 生成 8 位高强度随机数
	randomNum := generateSecureRandom(8)

	// 组合 UID + 时间戳 + 计数器 + UUID + 随机数
	combined := fmt.Sprintf("%s-%s-%s-%s", uid, timestampNano, id.String(), randomNum)

	// 将组合字符串转换为大整数
	idInt := new(big.Int)
	idInt.SetBytes([]byte(combined))

	// 计算 10^(length - len(timestamp))，确保最终长度
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	randomLength := length - len(timestamp)
	if randomLength < 1 {
		randomLength = 1 // 防止 timestamp 过长导致崩溃
	}

	// 计算 10^randomLength 作为取模范围
	maxValue := new(big.Int)
	maxValue.Exp(big.NewInt(10), big.NewInt(int64(randomLength)), nil)

	// 取模，确保生成的 ID 在指定长度范围内
	randomPart := idInt.Mod(idInt, maxValue)

	// 格式化输出，确保随机数部分补足 `randomLength` 位
	randomStr := fmt.Sprintf("%0*d", randomLength, randomPart)

	// 最终 时间戳(10) + 随机数部分
	return timestamp + randomStr
}
