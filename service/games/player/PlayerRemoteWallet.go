package player

import (
	"fmt"
	"videogame/common/proto/rpc"
	"videogame/common/util"
)

// 读取远程API钱包的数据
func (p *Player) getApiWalletBalance() bool {

	fmt.Printf("call SingleWalletService.RPCWalletOpt \n")

	nodeId := util.GetGameServiceName(util.SingleWalletService)
	rpcFunc := util.FormatRpcMethod(util.SingleWalletService, "RPCWalletOpt")

	var walletreq rpc.WalletOptReq
	var walletret rpc.WalletOptRet

	//获取余额只要三个参数
	walletreq.Opt = int32(2)
	walletreq.UserId = p.Id
	walletreq.Point = 0

	//读取钱包服的余额
	err := p.IGSService.CallNode(nodeId, rpcFunc, &walletreq, &walletret)

	if err != nil {

		fmt.Printf("call WallService.WalletOpt err:%v \n", err)
		return false
	}

	if walletret.Success == 1 {

		credit := uint(walletret.Balance * float64(p.WalletData.GetDigits()))

		p.WalletData.SetBalance(walletret.Balance)
		p.WalletData.SetCredit(credit)

		fmt.Printf("读取玩家单钱包服的余额成功: userId:%s , balance:%v \n", walletret.UserId, walletret.Balance)

		return true
	}

	return true

}

// 调用单钱包加减点
func (p *Player) optApiWalletBalance(data *OptData) bool {

	fmt.Printf("call SingleWalletService.RPCWalletOpt \n")

	nodeId := util.GetGameServiceName(util.SingleWalletService)
	rpcFunc := util.FormatRpcMethod(util.SingleWalletService, "RPCWalletOpt")

	if nodeId == util.IsEmpty {
	       return false
	}

	var walletreq rpc.WalletOptReq
	var walletret rpc.WalletOptRet

	//校验一下
	if data.Digits == 0 {
		data.Digits = 1
	}

	//加减点需要7个参数
	walletreq.Opt = int32(data.Opt)
	walletreq.UserId = p.Id
	walletreq.Point = float32(data.Point) / float32(data.Digits)
	walletreq.TranId = data.TransId
	walletreq.RoundId = data.RoundId
	walletreq.DataType = int32(data.TransType) //用于tranId的订单动作类型, 0下注订单/1取消下注订单/2小费订单/3结算订单/4其他订单
	walletreq.GameId = data.GameId
	walletreq.DeskNo = int32(data.DeskNo)

	//钱包服的余额增加
	err := p.IGSService.CallNode(nodeId, rpcFunc, &walletreq, &walletret)

	if err != nil {

		fmt.Printf("call WallService.WalletOpt err:%v \n", err)
		return false
	}

	//更新本地内存缓存
	if walletret.Success == 1 {

		credit := uint(walletret.Balance * float64(p.WalletData.GetDigits()))

		p.WalletData.SetBalance(walletret.Balance)
		p.WalletData.SetCredit(credit)

		fmt.Printf("更新玩家单钱包服的余额成功: userId:%s , balance:%v \n", walletret.UserId, walletret.Balance)

		return true
	}

	return true

}

// 主要用于通知商户单钱包服务，游戏开答了，让他们更新相关玩家订单的状态
func (p *Player) optApiWalletOpenAnswer(data *OpenAnswerData) bool {

	fmt.Printf("call SingleWalletService.RPCWalletOpt \n")

	nodeId := util.GetGameServiceName(util.SingleWalletService)
	rpcFunc := util.FormatRpcMethod(util.SingleWalletService, "RPCWalletOpt")

	if nodeId == util.IsEmpty {
	       return false
	}

	var walletreq rpc.WalletOptReq
	var walletret rpc.WalletOptRet

	//通知商户单钱包服务，游戏开答了
	walletreq.Opt = int32(4) //4类Opt是WOpt_Answer动作
	walletreq.UserId = p.Id
	walletreq.DeskNo = int32(data.DeskNo)
	walletreq.RoundId = data.RoundId

	err := p.IGSService.CallNode(nodeId, rpcFunc, &walletreq, &walletret)

	if err != nil {

		fmt.Printf("call WallService.WalletOpt err:%v \n", err)
		return false
	}

	if walletret.Success == 1 {

		fmt.Println("通知商户单钱包服务游戏开答了, 成功:")

		return true
	}

	return true

}