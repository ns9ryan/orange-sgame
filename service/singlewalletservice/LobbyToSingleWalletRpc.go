package singleWalletService

import (
	"errors"
	"fmt"
	"videogame/common/proto/rpc"
)

// 大厅会发送这个调用,当有玩家登陆游戏时，大厅会调用此RPC
func (self *SingleWalletService) RPCWalletInit(req *rpc.LoginToLobbyReq) error {

	p := self.playerFactory.GetPlayer(req.UserId)

	fmt.Printf("**************Lobby-->SingleWalletService \n")

	//创建用户
	if p == nil {

		p = self.playerFactory.NewPlayer(req.UserId, &self.msgSender, self, &self.msgReceiver)

		fmt.Printf("钱包服 NewPlayer userID:%v  \n", p.Id)

		if p == nil {
			return errors.New("player is nil")
		}

		//初始化单钱包玩家配置信息，包括币种/商户Id等
		p.InitBalance()

	} else {

		fmt.Printf("钱包服 已存在此玩家 userID:%v  \n", p.Id)
	}

	self.attachConn(req.ClientId, req.GateNodeId, p)

	//var walletreq rpc.WalletOptReq
	////var walletret rpc.WalletOptRet
	//
	////以下代码需要修改
	//walletreq.Opt = int32(WOpt_Init)
	//walletreq.UserId = req.UserId
	//walletreq.Point = 0
	//
	//fmt.Printf("*************call WallService.RPCWalletInit   start:\n")
	//
	////	nodeId := util.GetGameServiceName(util.WalletService)
	//rpcFunc := util.FormatRpcMethod(util.WalletService, "RPCWalletOpt")
	//
	//err := self.Call(rpcFunc, &walletreq, nil)
	//
	//if err != nil {
	//	fmt.Printf("call WallService.WalletOpt err:%v \n", err)
	//
	//	return errors.New("call walletservice RPCWalletOpt  error ")
	//}
	//
	//fmt.Printf("****************call WallService.RPCWalletInit   end:\n")

	return nil
}

// 大厅会发送这个调用,当有玩家退出游戏时，大厅会调用此RPC
func (self *SingleWalletService) RPCLeaveWallet(req *rpc.LoginToLobbyReq, ret *rpc.Int) error {

	//p := self.playerFactory.GetPlayer(req.UserId)
	//
	////创建用户
	//if p == nil {
	//	fmt.Printf("玩家要求玩家游戏房间，但之前已经离开! \n")
	//	return nil
	//}
	//
	//fmt.Printf("BetService RPCLeaveBet userID:%v\n", p.Id)
	//
	//self.delClientPlayer(req.ClientId)
	//self.playerFactory.ReleasePlayer(p)

	return nil
}
