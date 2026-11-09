//*****************************************************************
// OBLiveApi-5-Debug.go里，放的是内部对接和调试用途的接口(目前一共3个)，包括：
// 获取新商户code，获取商户code对应的aeskey，对原始JSON字符串签名加密。
//*****************************************************************

package apimodule

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/gin-gonic/gin"
	"videogame/service/merchantapiservice/apimodule/utils"
)

// 自测：生成新商户merchantCode和对应的Aeskey
func (h *MerchantApiModule) GenNewCode(c *gin.Context) {
	merchantCode, md5Salt, err := apifunc.GenerateMerchantCode()
	if err != nil {
		log.Error("生成商户Code失败:", log.ErrorField("err", err))
		ReturnHttpError(c, 0, err.Error(), "")
		return
	}

	aesKey, err := apifunc.GetAESKeyByDynamicSalt(merchantCode)
	if err != nil {
		log.Error("获取商户Aeskey失败:", log.ErrorField("err", err))
		ReturnHttpError(c, 0, err.Error(), "")
		return
	}

	// 构造返回的数据
	resultData := make(map[string]interface{})
	resultData["new_merchant_code"] = merchantCode
	resultData["aeskey_base64"] = base64.StdEncoding.EncodeToString(aesKey)
	resultData["md5salt_base64"] = base64.StdEncoding.EncodeToString([]byte(md5Salt))

	ReturnHttpOK(c, resultData, "")
	return
}

// 自测：获取某个商户Code对应的Aeskey
func (h *MerchantApiModule) GetCodeAeskey(c *gin.Context) {
	code := c.DefaultQuery("code", "")
	if code == "" {
		ReturnHttpError(c, 0, "未读到code值", "")
		return
	}

	fmt.Printf("get code=%s\n", code)
	aesKey, err := apifunc.GetAESKeyByDynamicSalt(code)
	if err != nil {
		ReturnHttpError(c, 0, err.Error(), "")
		return
	}

	// merchant, err := models.GetMerchantByCode(code)
	// if err != nil || merchant == nil {
	// 	log.Error("GetMerchantByCode:", log.ErrorField("err", err))
	// 	ReturnHttpError(c, 0, "未查询到merchant信息", "")
	// 	return
	// }

	// 构造返回的数据
	md5Salt := "ZhynnyBs3Ug3239RZhynnyBs3Ug3239R" //merchant.Md5Salt
	resultData := make(map[string]interface{})
	//resultData["aeskey_hex"] = fmt.Sprintf("%x", aesKey)
	resultData["aeskey_base64"] = base64.StdEncoding.EncodeToString(aesKey)
	resultData["md5salt_base64"] = base64.StdEncoding.EncodeToString([]byte(md5Salt))

	ReturnHttpOK(c, resultData, "")
	return
}

var requestMap = map[string]interface{}{}

// 获取原始JSON的签名和加密串
func (h *MerchantApiModule) GenerateSignatureEncrypt(c *gin.Context) {
	// 测试检查rawJson和Marshal的MD5是否一致
	// md5HashRaw := md5.New()
	// rawJson := `{"loginName":"j7l6testuser","timestamp":"1578662499442"}`
	// md5HashRaw.Write([]byte(rawJson))  // 直接对 data 进行 MD5 计算
	// signatureRaw := strings.ToUpper(hex.EncodeToString(md5HashRaw.Sum(nil)))
	// fmt.Printf("get signatureRaw=%s\n", signatureRaw)

	code := c.DefaultQuery("code", "")
	if code == "" {
		ReturnHttpError(c, 0, "未读到code值", "")
		return
	}

	op := c.DefaultQuery("op", "")
	if op == "" {
		ReturnHttpError(c, 0, "未读到op值", "")
		return
	}

	fmt.Printf("get code=%s, op=%s\n", code, op)
	apiRequest, ok := requestMap[op]
	if !ok {
		ReturnHttpError(c, 0, "无效的op操作类型", "")
		return
	}

	// 绑定请求体数据到 ApiRequest 结构体
	// var apiRequest ForwardGameRequest
	if err := c.ShouldBindJSON(&apiRequest); err != nil {
		log.Error("请求参数错误:", log.ErrorField("err", err))
		ReturnHttpError(c, 0, "请求参数错误", "")
		return
	}

	// 将请求数据转换为 map，并对键进行排序
	rawMap, err := apifunc.StructToSortedMap(apiRequest)
	if err != nil {
		log.Error("参数处理失败:", log.ErrorField("err", err))
		ReturnHttpError(c, 0, "参数处理失败", "")
		return
	}

	// 序列化请求数据并生成 MD5 签名
	data, err := json.Marshal(rawMap)
	if err != nil {
		log.Error("参数序列化失败:", log.ErrorField("err", err))
		ReturnHttpError(c, 0, "参数序列化失败", "")
		return
	}

	// 生成data的签名
	err, signature := apifunc.GenerateSignature(code, data)
	if err != nil {
		log.Error("生成签名失败:", log.ErrorField("err", err))
		ReturnHttpError(c, 0, "生成签名失败", "")
		return
	}

	fmt.Printf("signature=%s\n", signature)

	// 获取 AES 密钥
	aesKey, err := apifunc.GetAESKeyByDynamicSalt(code)
	if err != nil {
		ReturnHttpError(c, 0, err.Error(), "")
		return
	}

	encryptParams, err := apifunc.EncryptParams(apiRequest, aesKey)
	if err != nil {
		ReturnHttpError(c, 0, err.Error(), "")
		return
	}

	// 构造返回的数据
	resultData := make(map[string]string)
	resultData["signature"] = signature
	resultData["params"] = encryptParams

	// 返回成功响应
	ReturnHttpOK(c, resultData, apiRequest)
	return
}
