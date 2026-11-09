package player

import (
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"videogame/common/proto/rpc"
	"videogame/common/util"
)

// 读取本地集群内的钱包服的余额
func (p *Player) getClustorWalletBalance() bool {

	fmt.Printf("call WallService.GetWalletBalance \n")

	nodeId := util.GetGameServiceName(util.WalletService)
	rpcFunc := util.FormatRpcMethod(util.WalletService, "RPCWalletOpt")

	var walletreq rpc.WalletOptReq
	var walletret rpc.WalletOptRet

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

		fmt.Printf("&&&&读取玩家钱包服的余额成功: userId:%s , balance:%v \n", walletret.UserId, walletret.Balance)

		return true
	}

	return false
}

// 调用集群内钱包余额
func (p *Player) optClusterWalletBalance(data *OptData) bool {

	fmt.Printf("call WallService.optClusterWalletBalance  \n")

	nodeId := util.GetGameServiceName(util.WalletService)
	rpcFunc := util.FormatRpcMethod(util.WalletService, "RPCWalletOpt")

	if nodeId == "" {

		log.Errorf("optClusterWalletBalance nodeId is Empty !!!")

		return false
	}

	var walletreq rpc.WalletOptReq
	var walletret rpc.WalletOptRet

	//校验一下
	if data.Digits == 0 {
		data.Digits = 1
	}

	walletreq.Opt = int32(data.Opt)
	walletreq.UserId = p.Id
	walletreq.Point = float32(data.Point) / float32(data.Digits)
	walletreq.GameId = data.GameId

	//钱包服的余额增加
	err := p.IGSService.CallNode(nodeId, rpcFunc, &walletreq, &walletret)

	if err != nil {

		//检测是否余额不够
		if err.Error() == "balance is insufficient" {

		}

		log.Errorf("call WallService.WalletOpt err:%v \n", err)

		return false
	}

	if walletret.Success == 1 {

		credit := uint(walletret.Balance * float64(p.WalletData.GetDigits()))

		p.WalletData.SetBalance(walletret.Balance)
		p.WalletData.SetCredit(credit)

		log.Infof("存入玩家钱包服的余额成功: userId:%s , balance:%v \n", walletret.UserId, walletret.Balance)

		return true
	}

	return false
}

// 增加玩家的投注帐目数据
func (p *Player) AddBetAccount(bet uint, beteffect uint, water uint, winlost int) bool {

	fmt.Printf("111111 AddBetAccount\n")

	nodeId := util.GetGameServiceName(util.WalletService)
	rpcFunc := util.FormatRpcMethod(util.WalletService, "RPCWalletAddAccount")

	var walletAccount rpc.WalletAccount

	walletAccount.UserId = p.UserId
	walletAccount.Bet = uint32(bet)
	walletAccount.BetEffective = uint32(beteffect)
	walletAccount.Water = uint32(water)
	walletAccount.WinLost = int32(winlost)

	//钱包服的余额增加
	err := p.IGSService.GoNode(nodeId, rpcFunc, &walletAccount)

	if err != nil {
		fmt.Printf("call WallService.WalletOpt err:%v \n", err)
		return false
	}

	return true
}

// 玩家打赏，从钱包服扣款
func (p *Player) AddClustorWalletTip(data *OptData) bool {

	nodeId := util.GetGameServiceName(util.WalletService)
	rpcFunc := util.FormatRpcMethod(util.WalletService, "RPCWalletAddTipAmount")

	if nodeId == util.IsEmpty {
		return false
	}

	var walletTip rpc.WalletTipAmout

	walletTip.UserId = p.UserId
	walletTip.TipAmout = uint32(data.Point)
	walletTip.Digits = uint32(data.Digits)

	//钱包服的余额增加
	err := p.IGSService.GoNode(nodeId, rpcFunc, &walletTip)

	if err != nil {
		fmt.Printf("call WallService.WalletAddTip err:%v \n", err)
		return false
	}

	return true
}
