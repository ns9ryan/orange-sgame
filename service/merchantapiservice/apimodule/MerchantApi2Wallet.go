//*****************************************************************
// OBLiveApi-2-Wallet.go里，放的是API文档里描述的基础接口(目前一共10个)，包括：
// 创建账号、快速开始游戏、上分下分、查询余额等供商户调用的API接口。
//*****************************************************************

package apimodule

import (
	"encoding/json"
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/gin-gonic/gin"
	//"github.com/shopspring/decimal"
	//"videogame/service/merchantapiservice/apimodule/utils"
	//"videogame/common/proto/rpc"
	"strings"
	models "videogame/common/model"
	"math"
	"strconv"
	"videogame/common/proto/rpc"
	"time"
	//redisId "videogame/common/redis"
	//"videogame/common/global/variable"
)

// 会员余额查询，定义请求结构体
type MWalletBalanceApiReq struct {
	MerchantCode	string  `json:"merchant_code"`
	Username		string  `json:"username"`
}

type MWalletBalanceApiRet struct {
	Balance 		FixedOutput 	`json:"balance"`
	Code 		    string  		`json:"currency"`
}

var (
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

type FixedOutput struct {
	value any
}

func (f FixedOutput) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.value)
}

func WrapFixedFloat(amount float64, code string) FixedOutput {
	digits, ok := codeDigitsMap[code]

	if !ok {
		return FixedOutput{FixedFloat2(amount)}
	}

	switch digits {
	case 1:
		return FixedOutput{FixedFloat0(amount)}
	case 10:
		return FixedOutput{FixedFloat1(amount)}
	case 100:
		return FixedOutput{FixedFloat2(amount)}
	case 1000:
		return FixedOutput{FixedFloat3(amount)}
	case 10000:
		return FixedOutput{FixedFloat4(amount)}
	default:
		return FixedOutput{amount}
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

// 会员余额查询
func (h *MerchantApiModule) GetBalance(c *gin.Context) {
	var req MWalletBalanceApiReq
	var ret MWalletBalanceApiRet

	if err := c.ShouldBind(&req); err != nil {
		log.Error("请求参数错误:", log.ErrorField("err", err))
		ReturnHttpApi(c, 10001, "请求参数错误", "")
		return
	}

	username := strings.ToLower(fmt.Sprintf("%s_%s", req.MerchantCode, req.Username))

	//不能调RPCWalletOpt，改为直接从数据库查，因为WalletOpt需要玩家RPCLogin()
	// fmt.Printf("call WalletService.RPCWalletOpt \n")
	// var walletreq rpc.WalletOptReq
	// var walletret rpc.WalletOptRet
	// walletreq.Opt = int32(2) //(2)为对多钱包服的查询余额动作
	// walletreq.UserId = username
	// walletreq.Point = 0
	// err := h.Call("WalletService.RPCWalletOpt", &walletreq, &walletret)
	// if err != nil {
	// 	fmt.Printf("WalletService.RPCWalletOpt ************************************** Call error: %+v\n", err)
	// }

	accountInfo, err := models.GetUserWalletByUid(username)
	if err != nil {
		ReturnHttpApi(c, 10001, "没有找到用户!", "")
		return
	}

	fmt.Printf("accountInfo.Balance=%d, digits=%d", accountInfo.Balance, accountInfo.Digits)

	balance := float64(accountInfo.Balance) /  float64(accountInfo.Digits)
	ret.Balance = WrapFixedFloat(balance, accountInfo.Currency)
	ret.Code = "USD" //待从or_player表获取玩家币种

	fmt.Printf("GetBalance ret is %v\n", ret)
	ReturnHttpApi(c, 200, "", ret)
}


// 存款上分，定义请求结构体
type MWalletDepositApiReq struct {
	MerchantCode	string  `json:"merchant_code"`
	Username		string  `json:"username"`
	Code 		    string  `json:"currency"`
	Amount	 		float64 `json:"amount"`
	MerchantTransId string  `json:"merchat_trans_id"`
}

type MWalletDepositApiRet struct {
	ActionType		string  		`json:"action_type"`
	Amount	 		any 			`json:"amount"`
	Balance	 		any 			`json:"balance"`
	Code 		    string  		`json:"currency"`
	MerchantTransId string  		`json:"merchat_trans_id"`
	OrangeTransId 	string  		`json:"orange_trans_id"`
	OrangeTransTime string  		`json:"orange_trans_time"`
}

// 转账上分
func (h *MerchantApiModule) Deposit(c *gin.Context) {
	var req MWalletDepositApiReq
	var ret MWalletDepositApiRet

	if err := c.ShouldBind(&req); err != nil {
		log.Error("请求参数错误:", log.ErrorField("err", err))
		ReturnHttpApi(c, 10001, "请求参数错误", "")
		return
	}

	if req.Amount <= 0 {
		log.Error("存款金额需要大于0")
		ReturnHttpApi(c, 10001, "输入金额需要大于0", "")
		return
	}

	var rpcReq rpc.DepositApiReq
	var rpcRet rpc.DepositApiRet

	rpcReq.MerchantCode = req.MerchantCode
	rpcReq.Username = req.Username
	rpcReq.Amount = float32(req.Amount)
	rpcReq.MerchantTransId = req.MerchantTransId

	log.Infof("准备CRM RPC请求: %+v\n", rpcReq)
	err := h.Call("CrmRpcService.RPCDeposit", &rpcReq, &rpcRet)
	if err != nil {
		fmt.Printf("CrmRpcService.RPCDeposit ************************************** Call error: %+v\n", err)
		ReturnHttpApi(c, 10001, err.Error(), "")
		return
	}

	if rpcRet.Code != 200 {
		fmt.Printf("CrmRpcService.RPCDeposit ************************************** Call error: %+v\n", rpcRet.Msg)
		ReturnHttpApi(c, 10001, rpcRet.Msg, "")
		return
	}

	//通知多钱包服，让钱包服更新玩家的内存余额、队列更新数据库balance，以及redis的钱包余额
	var walletreq rpc.WalletOptReq
	var walletret rpc.WalletOptRet

	walletreq.Opt = int32(3) //3是对玩家的WOpt_Init请求
	walletreq.UserId = req.Username //strings.ToLower(req.MerchantCode + "_" + req.Username)

	log.Infof("准备多钱包玩家初始化请求: %+v\n", walletreq)
	err = h.Call("WalletService.RPCWalletOpt", &walletreq, &walletret)
	if err != nil {
		fmt.Printf("WalletService.RPCWalletOpt ************************************** Call error: %+v\n", err)
		ReturnHttpApi(c, 10001, err.Error(), "")
		return
	}

	walletreq.Opt = int32(0) //换成0，表示加点请求
	walletreq.Point = float32(req.Amount)

	log.Infof("准备多钱包加点请求: %+v\n", walletreq)
	err = h.Call("WalletService.RPCWalletOpt", &walletreq, &walletret)
	if err != nil {
		fmt.Printf("WalletService.RPCWalletOpt ************************************** Call error: %+v\n", err)
		ReturnHttpApi(c, 10001, err.Error(), "")
		return
	}

	if walletret.Success == 0 {
		fmt.Printf("WalletService.RPCWalletOpt ************************************** Call error: %+v\n", walletret)
		ReturnHttpApi(c, 10001, "多钱包请求失败", "")
		return
	}


	//准备最终结果返回
	ret.ActionType = "deposit"
	ret.Amount = WrapFixedFloat(req.Amount, req.Code)
	ret.Balance = WrapFixedFloat(float64(rpcRet.Balance), req.Code)
	ret.Code = req.Code
	ret.MerchantTransId = req.MerchantTransId
	ret.OrangeTransId = rpcRet.OrangeTransId
	ret.OrangeTransTime = rpcRet.OrangeTransTime

	//充值后删除Redis的余额缓存
	// wallketKey := redisId.RedisId_PlayerWallet + req.Username
	// _, err = variable.RedisStorage.Del(wallketKey)
	// if err != nil {
	// 	fmt.Printf("RPCDeposit后,清除redis缓存失败,err=%v\n", err)
	// }

	fmt.Printf("RPCDeposit ret is %v\n", ret)
	ReturnHttpApi(c, rpcRet.Code, rpcRet.Msg, ret)
	return
}


// 提现下分，定义请求结构体
type MWalletWithdrawApiReq struct {
	MerchantCode	string  `json:"merchant_code"`
	Username		string  `json:"username"`
	Code 		    string  `json:"currency"`
	Amount	 		float64 `json:"amount"`
	MerchantTransId string  `json:"merchat_trans_id"`
}

type MWalletWithDrawApiRet struct {
	ActionType		string  		`json:"action_type"`
	Amount	 		any 			`json:"amount"`
	Balance	 		any 			`json:"balance"`
	Code 		    string  		`json:"currency"`
	MerchantTransId string  		`json:"merchat_trans_id"`
	OrangeTransId 	string  		`json:"orange_trans_id"`
	OrangeTransTime string  		`json:"orange_trans_time"`
}

// 提现下分
func (h *MerchantApiModule) Withdraw(c *gin.Context) {
	var req MWalletWithdrawApiReq
	var ret MWalletWithDrawApiRet

	if err := c.ShouldBind(&req); err != nil {
		log.Error("请求参数错误:", log.ErrorField("err", err))
		ReturnHttpApi(c, 10001, "请求参数错误", "")
		return
	}

	if req.Amount <= 0 {
		log.Error("取款金额需要大于0")
		ReturnHttpApi(c, 10001, "输入金额需要大于0", "")
		return
	}

	var rpcReq rpc.WithdrawApiReq
	var rpcRet rpc.WithdrawApiRet

	rpcReq.MerchantCode = req.MerchantCode
	rpcReq.Username = req.Username
	rpcReq.Amount = float32(req.Amount)
	rpcReq.MerchantTransId = req.MerchantTransId

	log.Infof("准备RPC请求: %+v\n", rpcReq)
	err := h.Call("CrmRpcService.RPCWithdraw", &rpcReq, &rpcRet)
	if err != nil {
		fmt.Printf("CrmRpcService.RPCWithdraw ************************************** Call error: %+v\n", err)
		ReturnHttpApi(c, 10001, err.Error(), "")
		return
	}

	if rpcRet.Code != 200 {
		fmt.Printf("CrmRpcService.RPCDeposit ************************************** Call error: %+v\n", rpcRet.Msg)
		ReturnHttpApi(c, 10001, rpcRet.Msg, "")
		return
	}

	//通知多钱包服，让钱包服更新玩家的内存余额、队列更新数据库balance，以及redis的钱包余额
	var walletreq rpc.WalletOptReq
	var walletret rpc.WalletOptRet

	walletreq.Opt = int32(3) //3是对玩家的WOpt_Init请求
	walletreq.UserId = req.Username //strings.ToLower(req.MerchantCode + "_" + req.Username)

	log.Infof("准备多钱包玩家初始化请求: %+v\n", walletreq)
	err = h.Call("WalletService.RPCWalletOpt", &walletreq, &walletret)
	if err != nil {
		fmt.Printf("WalletService.RPCWalletOpt ************************************** Call error: %+v\n", err)
		ReturnHttpApi(c, 10001, err.Error(), "")
		return
	}

	walletreq.Opt = int32(1) //换成0，表示减点请求
	walletreq.Point = float32(req.Amount)

	log.Infof("准备多钱包减点请求: %+v\n", walletreq)
	err = h.Call("WalletService.RPCWalletOpt", &walletreq, &walletret)
	if err != nil {
		fmt.Printf("WalletService.RPCWalletOpt ************************************** Call error: %+v\n", err)
		ReturnHttpApi(c, 10001, err.Error(), "")
		return
	}

	if walletret.Success == 0 {
		fmt.Printf("WalletService.RPCWalletOpt ************************************** Call error: %+v\n", walletret)
		ReturnHttpApi(c, 10001, "多钱包请求失败", "")
		return
	}

	//准备最终结果返回
	ret.ActionType = "withdraw"
	ret.Amount = WrapFixedFloat(req.Amount, req.Code)
	ret.Balance = WrapFixedFloat(float64(rpcRet.Balance), req.Code)
	ret.Code = req.Code
	ret.MerchantTransId = req.MerchantTransId
	ret.OrangeTransId = rpcRet.OrangeTransId
	ret.OrangeTransTime = rpcRet.OrangeTransTime

	//充值后删除Redis的余额缓存
	// wallketKey := redisId.RedisId_PlayerWallet + req.Username
	// _, err = variable.RedisStorage.Del(wallketKey)
	// if err != nil {
	// 	fmt.Printf("RPCWithdraw后,清除redis缓存失败,err=%v\n", err)
	// }

	fmt.Printf("RPCWithdraw ret is %v\n", ret)
	ReturnHttpApi(c, rpcRet.Code, rpcRet.Msg, ret)
	return
}

// 转账结果查询，定义请求结构体
type MWalletTransferApiReq struct {
	MerchantCode	string  `json:"merchant_code"`
	Username		string  `json:"username"`
	Code 		    string  `json:"currency"`
	MerchantTransId string  `json:"merchat_trans_id"`
}

type MWalletTransferApiRet struct {
	Balance 		FixedOutput 	`json:"balance"`
	Code 		    string  `json:"currency"`
	MerchantTransId string  `json:"merchat_trans_id"`
	OrangeTransId 	string  `json:"orange_trans_id"`
}

// 转账结果查询
func (h *MerchantApiModule) ValidateTransfer(c *gin.Context) {
	var req MWalletTransferApiReq
	var ret MWalletTransferApiRet

	if err := c.ShouldBind(&req); err != nil {
		log.Error("请求参数错误:", log.ErrorField("err", err))
		ReturnHttpApi(c, 10001, "请求参数错误", "")
		return
	}

	trans, err := models.GetTransByMid(req.MerchantTransId)
	if err != nil {
		ReturnHttpApi(c, 10001, "商户订单号不存在", "")
		return
	}

	fmt.Printf("trans OrangeTransId=%s, AfterBalance=%d", trans.OrangeTransId, trans.After)

	if trans.Digits == 0 {
		trans.Digits = 1
	}

	balance := float64(trans.After / trans.Digits)
	ret.Balance = WrapFixedFloat(balance, req.Code)
	ret.Code = req.Code
	ret.MerchantTransId = req.MerchantTransId
	ret.OrangeTransId = trans.OrangeTransId

	fmt.Printf("ValidateTransfer ret is %v\n", ret)
	ReturnHttpApi(c, 200, "", ret)
}

// 玩家注单量查询，定义请求结构体
type MWalletReportApiReq struct {
	MerchantCode	string  `json:"merchant_code"`
	Username		string  `json:"username"`
	Code 		    string  `json:"currency"`
	StartTime 		string  `json:"start_time"`
	EndTime 		string  `json:"end_time"`
	PageIndex		int		`json:"page_index"`
}

type Summary struct {
	BetCount		int		`json:"bet_count"`
	BetAmount		FixedOutput		`json:"bet_amount"`
	ValidBetAmount	FixedOutput		`json:"valid_bet_amount"`
	NetAmount		FixedOutput		`json:"net_amount"`
}

type BetAccountRecord struct {
	GameTableId		int		`json:"game_table_id"`
	GameRoundId		string	`json:"game_round_id"`
	BetCount		int		`json:"bet_count"`
	BetAmount		FixedOutput		`json:"bet_amount"`
	ValidBetAmount	FixedOutput		`json:"valid_bet_amount"`
	NetAmount		FixedOutput		`json:"net_amount"`
}

type MWalletReportApiRet struct {
	Username		string  `json:"username"`
	Code 		    string  `json:"currency"`

	PageIndex		int		`json:"page_index"`
	PageSize		int		`json:"page_size"`
	TotalRecord		int		`json:"total_record"`
	TotalPage		int		`json:"total_page"`

	Summary			Summary 			`json:"summary"`
	Records			[]BetAccountRecord 	`json:"records"`
}

// 玩家注单量查询
func (h *MerchantApiModule) PlayerReport(c *gin.Context) {
	var req MWalletReportApiReq
	var ret MWalletReportApiRet

	if err := c.ShouldBind(&req); err != nil {
		log.Error("请求参数错误:", log.ErrorField("err", err))
		ReturnHttpApi(c, 10001, "请求参数错误", "")
		return
	}

	username := strings.ToLower(fmt.Sprintf("%s_%s", req.MerchantCode, req.Username))

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		ReturnHttpApi(c, 10001, "解析开始时间失败", "")
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		ReturnHttpApi(c, 10001, "解析开始时间失败", "")
		return
	}

	betAccountRecords, err := models.GetBetAccountRecords(username, req.Code, startTime, endTime, req.PageIndex)
	if err != nil {
		ReturnHttpApi(c, 10001, "商户订单号不存在", "")
		return
	}

	var records = make([]BetAccountRecord, 0)
	for _, bar := range betAccountRecords {
		if bar.Digits == 0 {
			bar.Digits = 1
		}

		var record BetAccountRecord
		record.GameTableId = bar.DeskNo
		record.GameRoundId = bar.RoundId
		record.BetCount = bar.BetCount
		record.BetAmount = WrapFixedFloat(float64(bar.BetAmount) / float64(bar.Digits), req.Code)
		record.ValidBetAmount = WrapFixedFloat(float64(bar.BetAmount) / float64(bar.Digits), req.Code)

		record.NetAmount = WrapFixedFloat(float64(bar.WinLost) / float64(bar.Digits), req.Code)
		if bar.WinLost == 0 {
			fmt.Printf("....bar.Winlost=%d", bar.WinLost)
			record.NetAmount = WrapFixedFloat(-float64(bar.BetAmount) / float64(bar.Digits), req.Code)
		}

		records = append(records, record)
	}

	amountSum, netSum, count, err := models.GetBetAccountStats(username, req.Code, startTime, endTime)

	ret.Username = req.Username
	ret.Code = req.Code
	ret.PageIndex = req.PageIndex
	ret.PageSize = 500
	ret.TotalRecord = len(records)
	ret.TotalPage = (len(records) + 500 - 1) / 500

	var summary Summary
	summary.BetCount = count
	summary.BetAmount = WrapFixedFloat(float64(amountSum), req.Code)
	summary.ValidBetAmount = WrapFixedFloat(float64(amountSum), req.Code)
	summary.NetAmount = WrapFixedFloat(float64(netSum), req.Code)

	ret.Summary = summary
	ret.Records = records

	fmt.Printf("PlayerReport ret is %v\n", ret)
	ReturnHttpApi(c, 200, "", ret)
}