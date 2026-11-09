//*****************************************************************
// OBLiveApi-5-Debug.go里，放的是内部对接和调试用途的接口(目前一共3个)，包括：
// 获取新商户code，获取商户code对应的aeskey，对原始JSON字符串签名加密。
//*****************************************************************

package apimodule

import (
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/gin-gonic/gin"
	"sync"
	"videogame/common/global/variable"
	"strconv"
	"strings"
	models "videogame/common/model"
	"math"
)

const (
	DefaultValue = 10000
	RedisKeySingleWallet = "SimSingleWallet"
)

var (
	userLocks  sync.Map
	codeDigitsMap map[string]uint
)

//目前digit基本是1/10/100/1000/10000，固定以下5个对应的格式化输出的类型
type FixedFloat0 float64
type FixedFloat1 float64
type FixedFloat2 float64
type FixedFloat3 float64
type FixedFloat4 float64

func (f FixedFloat0) MarshalJSON() ([]byte, error) {
	str := strconv.FormatFloat(math.Round(float64(f)), 'f', 0, 64)
	return []byte(str), nil
}

func (f FixedFloat1) MarshalJSON() ([]byte, error) {
	val := math.Round(float64(f)*1e1) / 1e1
	str := strconv.FormatFloat(val, 'f', 1, 64)
	return []byte(str), nil
}

func (f FixedFloat2) MarshalJSON() ([]byte, error) {
	val := math.Round(float64(f)*1e2) / 1e2
	str := strconv.FormatFloat(val, 'f', 2, 64)
	return []byte(str), nil
}

func (f FixedFloat3) MarshalJSON() ([]byte, error) {
	val := math.Round(float64(f)*1e2) / 1e2
	str := strconv.FormatFloat(val, 'f', 3, 64)
	return []byte(str), nil
}

func (f FixedFloat4) MarshalJSON() ([]byte, error) {
	val := math.Round(float64(f)*1e4) / 1e4
	str := strconv.FormatFloat(val, 'f', 4, 64)
	return []byte(str), nil
}

func WrapFixedFloat(amount float64, code string) any {
	digits, ok := codeDigitsMap[code]

	//如果找不到币种，默认返回两位小数
	if !ok {
		return FixedFloat2(amount)
	}

	switch digits {
	case 1:
		return FixedFloat0(amount)
	case 10:
		return FixedFloat1(amount)
	case 100:
		return FixedFloat2(amount)
	case 1000:
		return FixedFloat3(amount)
	case 10000:
		return FixedFloat4(amount)
	default:
		return amount // fallback
	}
}

func init() {
	if codeDigitsMap == nil {
		codeDigitsMap = make(map[string]uint)
		currs, err := models.GetSystemCurrency()
		if err != nil {
			fmt.Printf("获取系统币种列表失败")
			return
		}else{
			for _, curr := range currs {
				codeDigitsMap[curr.Code] = curr.Digits
			}		
		}

		fmt.Printf("GetSystemCurrency, codeDigitsMap=%v\n", codeDigitsMap)
	}
}

// MerchantBalanceRequest 商户余额请求参数
type CreditDebitRequest struct {
    Username          string        `json:"username"`
    Code              string        `json:"code"` //币种，如 USD、CNY 等

    GameTableId       int32         `json:"game_table_id"`    //4
    GameRoundId       string        `json:"game_round_id"`    //"jPNX07jf_xU7VV6IQPZrr"
    OrangeTransType   string        `json:"orange_trans_type"`//动作备注：bet或者tip或者win
    OrangeTrasnTime   string        `json:"orange_trans_time"`//"2025-03-18T14:23:54+08:00"
    OrangeTransId     string        `json:"orange_trans_id"`  //"DTX9m_edKSfsTDLo_SXCxaVz"
    Amount            float64       `json:"amount"`           //0.5000
}

type CreditDebitDataReturn struct {
    ActionType        string        `json:"action_type"`      //动作类型，debit或者credit
    Amount            any       	`json:"amount"`           //0.5000
    Balance           any       	`json:"balance"`          //130.5000
    Code              string        `json:"code"`             //币种，如 USD、CNY 等

    MerchantTrasnTime string        `json:"merchant_trans_time"`//"2025-03-18T14:23:54+08:00"
    MerchantTransId   string        `json:"merchant_trans_id"`  //"DTX9m_edKSfsTDLo_SXCxaVz"

    OrangeTrasnTime   string        `json:"orange_trans_time"`//"2025-03-18T14:23:54+08:00"
    OrangeTransId     string        `json:"orange_trans_id"`  //"DTX9m_edKSfsTDLo_SXCxaVz"
}

type CancelDebitDataReturn struct {
    Balance           any       	`json:"balance"`          //130.5000
    Code              string        `json:"code"`             //币种，如 USD、CNY 等
}

// 自测：模拟作为单钱包的商户端扣款
func (h *MerchantApiModuleImitate) DebitSingleWallet(c *gin.Context) {
	var req CreditDebitRequest
	var ret CreditDebitDataReturn

	OrToken := c.GetHeader("orToken")
	fmt.Printf("DebitSingleWallet Get Header OrToken=%s\n", OrToken)

	if err := c.ShouldBind(&req); err != nil {
		log.Error("请求参数错误:", log.ErrorField("err", err))
		ReturnHttpApi(c, 10001, "请求参数错误", "")
		return
	}

	if req.Username == ""{
		//log.Error("请求参数错误:", log.ErrorField("err", err))
		ReturnHttpApi(c, 10001, "用户为空", "")
		return		
	}

    lock, _ := userLocks.LoadOrStore(req.Username, &sync.Mutex{})
    userLock := lock.(*sync.Mutex)
    userLock.Lock()
    defer userLock.Unlock()

	var floatValue float64 = 0
	strValue, err := variable.RedisStorage.HGet(RedisKeySingleWallet, req.Username)
	if err != nil {
		fmt.Printf("Redis HGet失败, err=%v\n", err)
		ReturnHttpApi(c, 10001, "连接Redis失败", "")
		return
	}else{
		// 存在则解析 float
		var parseErr error
		floatValue, parseErr = strconv.ParseFloat(strValue, 64)
		if parseErr != nil {
			fmt.Printf("解析钱包值失败, err=%v\n", err)
			ReturnHttpApi(c, 10001, "解析钱包值失败", "")
			return
		}
	}

	if req.Amount > floatValue {
		fmt.Printf("11111111111 余额不足!!!\n")
		ReturnHttpApi(c, 10001, "余额不足", "")
		return
	}

	newFloatValue := floatValue - req.Amount
	valueStr := fmt.Sprintf("%.4f", newFloatValue)
	setErr := variable.RedisStorage.HSet(RedisKeySingleWallet, req.Username, valueStr)
	if setErr != nil {
		fmt.Printf("Redis HSet失败, err=%v\n", setErr)
		ReturnHttpApi(c, 10001, "连接Redis失败", "")
		return
	}

	ret.Balance = WrapFixedFloat(newFloatValue, req.Code)
	ret.ActionType = "debit"
	ret.Amount = WrapFixedFloat(req.Amount, req.Code)
	ret.Code = req.Code
	ret.MerchantTrasnTime = "2025-03-18T14:23:54+08:00"
	ret.MerchantTransId = "DTX9m_edKSfsTDLo_SXCxaVz"
	ret.OrangeTrasnTime = req.OrangeTrasnTime
	ret.OrangeTransId = req.OrangeTransId

	fmt.Printf("DebitSingleWallet ret is %v\n", ret)
	ReturnHttpApi(c, 200, "", ret)
}

// 自测：模拟作为单钱包的商户端加款
func (h *MerchantApiModuleImitate) CreditSingleWallet(c *gin.Context) {
	var req CreditDebitRequest
	var ret CreditDebitDataReturn

	OrToken := c.GetHeader("orToken")
	fmt.Printf("CreditSingleWallet Get Header OrToken=%s\n", OrToken)

	if err := c.ShouldBind(&req); err != nil {
		log.Error("请求参数错误:", log.ErrorField("err", err))
		ReturnHttpApi(c, 10001, "请求参数错误", "")
		return
	}

	if req.Username == ""{
		//log.Error("请求参数错误:", log.ErrorField("err", err))
		ReturnHttpApi(c, 10001, "用户为空", "")
		return		
	}

    lock, _ := userLocks.LoadOrStore(req.Username, &sync.Mutex{})
    userLock := lock.(*sync.Mutex)
    userLock.Lock()
    defer userLock.Unlock()

	var floatValue float64 = 0
	strValue, err := variable.RedisStorage.HGet(RedisKeySingleWallet, req.Username)
	if err != nil {
		fmt.Printf("Redis HGet失败, err=%v\n", err)
		ReturnHttpApi(c, 10001, "连接Redis失败", "")
		return
	}else{
		// 存在则解析 float
		var parseErr error
		floatValue, parseErr = strconv.ParseFloat(strValue, 64)
		if parseErr != nil {
			fmt.Printf("解析钱包值失败, err=%v\n", err)
			ReturnHttpApi(c, 10001, "解析钱包值失败", "")
			return
		}
	}

	newFloatValue := floatValue + req.Amount
	valueStr := fmt.Sprintf("%.4f", newFloatValue)
	setErr := variable.RedisStorage.HSet(RedisKeySingleWallet, req.Username, valueStr)
	if setErr != nil {
		fmt.Printf("Redis HSet失败, err=%v\n", setErr)
		ReturnHttpApi(c, 10001, "连接Redis失败", "")
		return
	}

	ret.Balance = WrapFixedFloat(newFloatValue, req.Code)
	ret.ActionType = "credit"
	ret.Amount =  WrapFixedFloat(req.Amount, req.Code)
	ret.Code = req.Code
	ret.MerchantTrasnTime = "2025-03-18T14:23:54+08:00"
	ret.MerchantTransId = "DTX9m_edKSfsTDLo_SXCxaVz"
	ret.OrangeTrasnTime = req.OrangeTrasnTime
	ret.OrangeTransId = req.OrangeTransId

	fmt.Printf("CreditSingleWallet ret is %v\n", ret)
	ReturnHttpApi(c, 200, "", ret)
}

// 自测：模拟作为单钱包的商户端取消下注
func (h *MerchantApiModuleImitate) CancelDebitSingleWallet(c *gin.Context) {
	var req CreditDebitRequest
	var ret CancelDebitDataReturn

	OrToken := c.GetHeader("orToken")
	fmt.Printf("CreditSingleWallet Get Header OrToken=%s\n", OrToken)

	if err := c.ShouldBind(&req); err != nil {
		log.Error("请求参数错误:", log.ErrorField("err", err))
		ReturnHttpApi(c, 10001, "请求参数错误", "")
		return
	}

	if req.Username == ""{
		//log.Error("请求参数错误:", log.ErrorField("err", err))
		ReturnHttpApi(c, 10001, "用户为空", "")
		return		
	}

    lock, _ := userLocks.LoadOrStore(req.Username, &sync.Mutex{})
    userLock := lock.(*sync.Mutex)
    userLock.Lock()
    defer userLock.Unlock()

	var floatValue float64 = 0
	strValue, err := variable.RedisStorage.HGet(RedisKeySingleWallet, req.Username)
	if err != nil {
		fmt.Printf("Redis HGet失败, err=%v\n", err)
		ReturnHttpApi(c, 10001, "连接Redis失败", "")
		return
	}else{
		// 存在则解析 float
		var parseErr error
		floatValue, parseErr = strconv.ParseFloat(strValue, 64)
		if parseErr != nil {
			fmt.Printf("解析钱包值失败, err=%v\n", err)
			ReturnHttpApi(c, 10001, "解析钱包值失败", "")
			return
		}
	}

	newFloatValue := floatValue + req.Amount
	valueStr := fmt.Sprintf("%.4f", newFloatValue)
	setErr := variable.RedisStorage.HSet(RedisKeySingleWallet, req.Username, valueStr)
	if setErr != nil {
		fmt.Printf("Redis HSet失败, err=%v\n", setErr)
		ReturnHttpApi(c, 10001, "连接Redis失败", "")
		return
	}

	ret.Balance = WrapFixedFloat(newFloatValue, req.Code)
	ret.Code = req.Code

	fmt.Printf("CreditSingleWallet ret is %v\n", ret)
	ReturnHttpApi(c, 200, "", ret)
}

// 会员余额查询，定义请求结构体
type SWalletBalanceApiReq struct {
	Username        string  `json:"username"`
	Code            string  `json:"code"`
}

type SWalletBalanceApiRet struct {
	Balance 		any 		 `json:"balance"`
	Code 		    string  	 `json:"code"`
}

// 自测：模拟作为单钱包的商户端
func (h *MerchantApiModuleImitate) GetBalanceSingleWallet(c *gin.Context) {
	var req SWalletBalanceApiReq
	var ret SWalletBalanceApiRet

	OrToken := c.GetHeader("orToken")
	fmt.Printf("Get Header OrToken=%s\n", OrToken)

	if err := c.ShouldBind(&req); err != nil {
		log.Error("请求参数错误:", log.ErrorField("err", err))
		ReturnHttpApi(c, 10001, "请求参数错误", "")
		return
	}

	var floatValue float64 = 0
	strValue, err := variable.RedisStorage.HGet(RedisKeySingleWallet, req.Username)
	if err != nil {
		fmt.Printf("RedisKeySingleWallet err=%v\n", err)
		if strings.Contains(err.Error(), "redis: nil") {
			// 写入默认值
			fmt.Println("Redis字段不存在，写入默认值")
			floatValue = DefaultValue
			valueStr := fmt.Sprintf("%d", int32(floatValue))
			setErr := variable.RedisStorage.HSet(RedisKeySingleWallet, req.Username, valueStr)
			if setErr != nil {
				fmt.Printf("Redis HSet失败, err=%v\n", setErr)
				ReturnHttpApi(c, 10001, "连接Redis失败", "")
				return
			}
		}else{
			fmt.Printf("Redis HGet失败, err=%v\n", err)
			ReturnHttpApi(c, 10001, "连接Redis失败", "")
			return
		}
	}else{
		// 存在则解析 float
		var parseErr error
		floatValue, parseErr = strconv.ParseFloat(strValue, 64)
		if parseErr != nil {
			fmt.Printf("解析钱包值失败, err=%v\n", err)
			ReturnHttpApi(c, 10001, "解析钱包值失败", "")
			return
		}		
	}

	ret.Balance = WrapFixedFloat(floatValue, req.Code)
	ret.Code = req.Code

	fmt.Printf("GetBalance ret is %v\n", ret)
	ReturnHttpApi(c, 200, "", ret)
}

// 校验API服务器接口是否正常，可用于定期心跳检测
func (h *MerchantApiModuleImitate) CheckSingleWalletOk(c *gin.Context) {
	//判断api服务或者游戏服务处于维护状态
	//ReturnHttpApi(c, 10001, "服务维护中", "")

	ReturnHttpApi(c, 200, "成功", "")
}

// 会员余额查询，定义请求结构体
type SRoundSettleApiReq struct {
    GameTableId       int32         `json:"game_table_id"`    //4
    GameRoundId       string        `json:"game_round_id"`    //"jPNX07jf_xU7VV6IQPZrr"
}

// 自测：模拟作为单钱包的商户端
func (h *MerchantApiModuleImitate) RoundSettle(c *gin.Context) {
	var req SRoundSettleApiReq

	OrToken := c.GetHeader("orToken")
	fmt.Printf("Get Header OrToken=%s\n", OrToken)

	if err := c.ShouldBind(&req); err != nil {
		log.Error("请求参数错误:", log.ErrorField("err", err))
		ReturnHttpApi(c, 10001, "请求参数错误", "")
		return
	}

	fmt.Printf("11111 RoundSettle req is %v\n", req)
	ReturnHttpApi(c, 200, "success", "")
}
