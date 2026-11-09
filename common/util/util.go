package util

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto" // 新版 proto
	"net/http"
	"strconv"
	"strings"
	"videogame/common/global/variable"
	"videogame/common/proto/msg"
)

// Fail 用于返回失败响应
func Fail(c *gin.Context, message string) {

	if message == "" {
		message = GetErrCodeMsg(msg.ErrCode_Fail) // 默认错误信息
	}

	c.JSON(http.StatusOK, gin.H{
		"Code": msg.ErrCode_Fail,
		"Msg":  message,
	})

	return
}

// Success 用于返回成功响应
func Success(c *gin.Context, message proto.Message) {

	jsonData, err := protojson.MarshalOptions{EmitUnpopulated: true}.Marshal(message)
	if err != nil {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Code": msg.ErrCode_Ok,
		"Msg":  GetErrCodeMsg(msg.ErrCode_Ok),
		"Data": json.RawMessage(jsonData),
	})
}

// 转换bool到int
func ConverBoolToInt(val bool) int {

	if val == true {
		return 1
	}

	return 0
}

// 转换int到bool
func ConvertIntToBool(val int) bool {

	if val == 0 {
		return false
	}
	return true
}

// 转换字符串为数字KEY
func StringToUintKey(value string) uint {

	key := uint(0)

	for val := range value {
		key += uint(val)
	}

	return key
}

// 获取语言对应的投注区域名称
func GetAreaNameByLangAndAreaId(lang string, areaId int) string {

	key := fmt.Sprintf("%s.%d", lang, areaId)
	return variable.ConfigAreaNameYml.GetString(key)
}

// 转换字符串为整型切片
func GetUintSliceByStrings(strnum string, sep string) []uint {

	numstrs := strings.Split(strnum, sep)

	numslice := make([]uint, 0)

	for _, str := range numstrs {

		num, err := strconv.Atoi(str)

		if err != nil {
			continue
		}

		numslice = append(numslice, uint(num))
	}

	return numslice
}
