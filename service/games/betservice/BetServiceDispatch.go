package betService

import (
	"errors"
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"videogame/common/proto/msg"
	"videogame/common/proto/rpc"
	"videogame/common/util"
	"videogame/service/games/player"
)

// 玩家投注
func (self *BetService) OnBetToGame(cliId string, betMsg *msg.CTOS_BetGame) error {

	//fmt.Println("BetService   RPCBetGame---------------------")

	pp := self.GetClientPlayer(cliId)

	pplayer, ok := pp.(*player.Player)

	//if ok == false {
	//	return errors.New("player is not player")
	//}

	//若没有找到玩家，需要去大厅服读取相关数据
	if pplayer == nil || ok == false {

		log.Errorf("BetService-->RPCBetGame 内存找不到玩家:clienId: %s  orderkey:%s", cliId, betMsg.OrderKey)

		nodeId := util.GetGameServiceName(util.LobbyService)
		rpcFunc := util.FormatRpcMethod(util.LobbyService, "RPCGetPlayer")

		if nodeId == "" {
			return errors.New("lobby clustor is empty！")
		}

		var rpcclid rpc.String

		rpcclid.Value = cliId

		err := self.AsyncCallNode(nodeId, rpcFunc, &rpcclid, func(cliInfo *rpc.ClientInfoRPC, err error) {

			if err != nil {
				//发送投注不成功的消息，返回玩家端
				log.Errorf("BetService-->RPCBetGame-->RPCGetPlayer  发生错误 ！！")

				return
			}

			var betInfo rpc.BetToGameRPC

			betInfo.UserId = cliInfo.UserId
			betInfo.ClientId = cliInfo.ClientId
			betInfo.GateNodeId = cliInfo.GateNodeId

			betInfo.OrderKey = int64(betMsg.OrderKey)
			betInfo.GameId = betMsg.GameId
			betInfo.DeskNo = int32(betMsg.DeskNo)
			betInfo.BetAreadId = betMsg.BetAreadId
			betInfo.BetPoint = betMsg.BetPoint
			betInfo.Digits = betMsg.Digits

			gameName, ok := util.GetServerNameFromGameId(betMsg.GameId)

			if ok == false {

				log.Errorf("BetService-->RPCBetGame-->查找:%s 没有找到！", betMsg.GameId)
				return
			}

			nodeId = util.GetGameServiceName(gameName)
			rpcFunc = util.FormatRpcMethod(gameName, "RPCBetGame")

			if nodeId == "" {
				log.Errorf("BetService-->RPCBetGame-->查找nodeId:%s 没有找到！", betMsg.GameId)
				return
			}

			//投递到实际的游戏服
			err = self.GoNode(nodeId, rpcFunc, &betInfo)

			if err != nil {

				log.Errorf("BetService-->RPCBetGame--> 调用nodeId:%s 不成功！ ", nodeId)

				return
			}

		})

		if err != nil {
			return errors.New("call rpc fail")
		}
	}
	//这边不再过多的处理

	//发往对应的游戏服，进一步处理,rpc调用
	gameName, ok := util.GetServerNameFromGameId(betMsg.GameId)

	if ok == false {

		log.Errorf(" 投注服找不到对应的游戏id的服务名 gameId:%s  userId:%s", betMsg.GameId, pplayer.UserId)

		var betResult msg.STOC_BetGame_Result

		betResult.Success = 1
		betResult.GameId = betMsg.GameId
		betResult.BetMode = int32(betMsg.BetMode)
		betResult.DeskNo = betMsg.DeskNo
		betResult.OrderKey = uint32(betMsg.OrderKey)

		betResult.Balance = int32(pplayer.GetCredit())
		betResult.Digits = int32(pplayer.GetDigits())
		betResult.ErrCode = int32(msg.ErrCode_BetErr_ServerClusterErr)

		pplayer.SendMsg(msg.MsgType_GameTable_BetResult, &betResult)

		return errors.New("RPCBetGame  GetServerNameFromGameId find no gameName  failed")
	}

	nodeId := util.GetGameServiceName(gameName)
	rpcFunc := util.FormatRpcMethod(gameName, "RPCBetGame")

	if nodeId == "" {

		log.Errorf(" 投注服找不到对应的游戏服node gameName:%s  userId:%s   ", gameName, pplayer.UserId)

		var betResult msg.STOC_BetGame_Result

		betResult.Success = 1
		betResult.GameId = betMsg.GameId
		betResult.BetMode = int32(betMsg.BetMode)
		betResult.DeskNo = betMsg.DeskNo
		betResult.OrderKey = uint32(betMsg.OrderKey)

		betResult.Balance = int32(pplayer.GetCredit())
		betResult.Digits = int32(pplayer.GetDigits())
		betResult.ErrCode = int32(msg.ErrCode_BetErr_ServerClusterErr)

		pplayer.SendMsg(msg.MsgType_GameTable_BetResult, &betResult)

		return errors.New("Cluster Error!")
	}

	var betInfo rpc.BetToGameRPC

	betInfo.UserId = pplayer.Id
	betInfo.ClientId = pplayer.ClientId
	betInfo.GateNodeId = pplayer.GetGateNodeId()

	betInfo.OrderKey = int64(betMsg.OrderKey)
	betInfo.GameId = betMsg.GameId
	betInfo.DeskNo = int32(betMsg.DeskNo)
	betInfo.BetAreadId = betMsg.BetAreadId
	betInfo.BetPoint = betMsg.BetPoint
	betInfo.Digits = betMsg.Digits

	//调用实际的游戏服，投注过去
	err := self.GoNode(nodeId, rpcFunc, &betInfo)

	//	fmt.Printf("*******BetService  --> OnBetToGame -->gameName:%s ,betInfo:%+v   \n  ", gameName, betInfo)

	if err != nil {

		//fmt.Printf("*******BetService  --> OnBetToGame -->gameName failed!!!!!!!!!!!!!!!!!!!!\n")
		log.Errorf(" 投注服调用游戏服出错！！！ nodeId:%s  userId:%s   ", nodeId, pplayer.UserId)

		var betResult msg.STOC_BetGame_Result

		betResult.Success = 1
		betResult.GameId = betMsg.GameId
		betResult.BetMode = int32(betMsg.BetMode)
		betResult.DeskNo = betMsg.DeskNo
		betResult.OrderKey = uint32(betMsg.OrderKey)

		betResult.Balance = int32(pplayer.GetCredit())
		betResult.Digits = int32(pplayer.GetDigits())
		betResult.ErrCode = int32(msg.ErrCode_BetErr_ServerClusterErr)

		pplayer.SendMsg(msg.MsgType_GameTable_BetResult, &betResult)

		return errors.New("Cluster Error!")
	}

	return nil
}

// 撤消投注消息
func (self *BetService) OnCancelBetToGame(cliId string, betMsg *msg.CTOS_CancelBetGame) error {

	fmt.Println("BetService   OnCancelBetToGame---------------------")

	pp := self.GetClientPlayer(cliId)

	pplayer, ok := pp.(*player.Player)

	if ok == false {
		return errors.New("player is not player")
	}

	//

	//若没有找到玩家，需要去大厅服读取相关数据
	if pplayer == nil {

		nodeId := util.GetGameServiceName(util.LobbyService)
		rpcFunc := util.FormatRpcMethod(util.LobbyService, "RPCGetPlayer")

		var rpcclid rpc.String

		rpcclid.Value = cliId

		err := self.AsyncCallNode(nodeId, rpcFunc, &rpcclid, func(cliInfo *rpc.ClientInfoRPC, err error) {

			if err != nil {
				//发送投注不成功的消息
				return
			}

			var betInfo rpc.BetToGameRPC

			betInfo.UserId = cliInfo.UserId
			betInfo.ClientId = cliInfo.ClientId
			betInfo.GateNodeId = cliInfo.GateNodeId

			betInfo.OrderKey = int64(betMsg.OrderKey)
			betInfo.GameId = betMsg.GameId
			betInfo.DeskNo = int32(betMsg.DeskNo)

			gameName, ok := util.GetServerNameFromGameId(betMsg.GameId)

			if ok == false {
				return
			}

			nodeId = util.GetGameServiceName(gameName)
			rpcFunc = util.FormatRpcMethod(gameName, "RPCCancelBetGame")

			//投递到实际的游戏服
			err = self.GoNode(nodeId, rpcFunc, &betInfo)

			if err != nil {
				log.Errorf("投注取消 投注实际游戏服不成功！！！")
				return
			}
		})

		if err != nil {

			log.Errorf("投注取消 投注实际游戏服不成功！！！")

			return errors.New("call rpc fail")
		}
	}

	//发往对应的游戏服，进一步处理,rpc调用
	gameName, ok := util.GetServerNameFromGameId(betMsg.GameId)

	if ok == false {

		return errors.New("RPCBetGame  GetServerNameFromGameId find no gameName  failed")
	}

	nodeId := util.GetGameServiceName(gameName)
	rpcFunc := util.FormatRpcMethod(gameName, "RPCCancelBetGame")

	var betInfo rpc.BetToGameRPC

	betInfo.UserId = pplayer.Id
	betInfo.ClientId = pplayer.ClientId
	betInfo.GateNodeId = pplayer.GetGateNodeId()

	betInfo.OrderKey = int64(betMsg.OrderKey)
	betInfo.GameId = betMsg.GameId
	betInfo.DeskNo = int32(betMsg.DeskNo)

	err := self.GoNode(nodeId, rpcFunc, &betInfo)

	if err != nil {

		log.Errorf("投注取消 投注实际游戏服不成功！！！")

		return errors.New("Cluster Error!")
	}

	return nil
}

// 处理打赏功能
func (self *BetService) OnTipToGame(cliId string, tipMsg *msg.CTOS_TipMoney) error {

	fmt.Println("BetService   OnTipToGame---------------------")

	pp := self.GetClientPlayer(cliId)

	pplayer, ok := pp.(*player.Player)

	if ok == false {
		return errors.New("player is not player")
	}

	//若没有找到玩家，需要去大厅服读取相关数据
	if pplayer == nil {

		nodeId := util.GetGameServiceName(util.LobbyService)
		rpcFunc := util.FormatRpcMethod(util.LobbyService, "RPCGetPlayer")

		var rpcclid rpc.String

		rpcclid.Value = cliId

		err := self.AsyncCallNode(nodeId, rpcFunc, &rpcclid, func(cliInfo *rpc.ClientInfoRPC, err error) {

			if err != nil {

				//发送投注不成功的消息
				return
			}

			var tipInfo rpc.TipToGameRPC

			tipInfo.UserId = cliInfo.UserId
			tipInfo.ClientId = cliInfo.ClientId
			tipInfo.GateNodeId = cliInfo.GateNodeId

			tipInfo.OrderKey = int64(tipMsg.OrderKey)
			tipInfo.GameId = tipMsg.GameId
			tipInfo.DeskNo = int32(tipMsg.DeskNo)
			tipInfo.Digits = tipMsg.Digits
			tipInfo.TipPoint = tipMsg.TipMoney
			tipInfo.DealerId = tipMsg.DealerId

			gameName, ok := util.GetServerNameFromGameId(tipMsg.GameId)

			if ok == false {
				return
			}

			nodeId = util.GetGameServiceName(gameName)
			rpcFunc = util.FormatRpcMethod(gameName, "RPCTipGame")

			//投递到实际的游戏服
			err = self.GoNode(nodeId, rpcFunc, &tipInfo)

			if err != nil {

				return
			}

		})

		if err != nil {
			return errors.New("call rpc fail")
		}
	}
	//这边不再过多的处理

	//发往对应的游戏服，进一步处理,rpc调用
	gameName, ok := util.GetServerNameFromGameId(tipMsg.GameId)

	if ok == false {

		log.Errorf("打赏时 投注服找不到对应的游戏id的服务名 gameId:%s  userId:%s", tipMsg.GameId, pplayer.UserId)

		return errors.New("RPCBetGame  GetServerNameFromGameId find no gameName  failed")
	}

	nodeId := util.GetGameServiceName(gameName)
	rpcFunc := util.FormatRpcMethod(gameName, "RPCTipGame")

	var tipInfo rpc.TipToGameRPC

	tipInfo.UserId = pplayer.Id
	tipInfo.ClientId = pplayer.ClientId
	tipInfo.GateNodeId = pplayer.GetGateNodeId()

	tipInfo.OrderKey = int64(tipMsg.OrderKey)
	tipInfo.GameId = tipMsg.GameId
	tipInfo.DeskNo = int32(tipMsg.DeskNo)
	tipInfo.TipPoint = tipMsg.TipMoney
	tipInfo.Digits = tipMsg.Digits

	tipInfo.DealerId = tipMsg.DealerId

	err := self.GoNode(nodeId, rpcFunc, &tipInfo)

	fmt.Printf("*******BetService  --> OnTipToGame -->gameName:%s ,betInfo:%+v   \n  ", gameName, tipInfo)

	if err != nil {

		fmt.Printf("*******BetService  --> OnBetToGame -->gameName failed!!!!!!!!!!!!!!!!!!!!\n")

		log.Errorf(" 打赏 投注服找不到对应的游戏服 gameId:%s  userId:%s", tipMsg.GameId, pplayer.UserId)

		return errors.New("Cluster Error!")
	}

	return nil
}

// 处理打赏功能
func (self *BetService) OnTalkToGame(cliId string, talkMsg *msg.CTOS_TalkChat) error {

	fmt.Println("BetService   OnTalkChat---------------------")

	pp := self.GetClientPlayer(cliId)

	pplayer, ok := pp.(*player.Player)

	if ok == false {
		return errors.New("player is not player")
	}

	//若没有找到玩家，需要去大厅服读取相关数据
	if pplayer == nil {

		nodeId := util.GetGameServiceName(util.LobbyService)
		rpcFunc := util.FormatRpcMethod(util.LobbyService, "RPCGetPlayer")

		var rpcclid rpc.String

		rpcclid.Value = cliId

		err := self.AsyncCallNode(nodeId, rpcFunc, &rpcclid, func(cliInfo *rpc.ClientInfoRPC, err error) {

			if err != nil {

				//发送投注不成功的消息
				return
			}

			//检测帐号ID是否一致
			if cliInfo.UserId != talkMsg.PlayerId {

				return
			}

			var talkInfo rpc.GameTalkChat

			talkInfo.GameId = talkMsg.GameId
			talkInfo.Deskno = talkMsg.Deskno
			talkInfo.Message = talkMsg.Message
			talkInfo.UserId = talkMsg.PlayerId

			gameName, ok := util.GetServerNameFromGameId(talkMsg.GameId)

			if ok == false {
				return
			}

			nodeId = util.GetGameServiceName(gameName)
			rpcFunc = util.FormatRpcMethod(gameName, "RPCTalkGame")

			//投递到实际的游戏服
			err = self.GoNode(nodeId, rpcFunc, &talkInfo)

			if err != nil {

				log.Errorf("call rpc  talk fail")
				return
			}

		})

		if err != nil {
			return errors.New("call rpc fail")
		}
	}
	//这边不再过多的处理

	//发往对应的游戏服，进一步处理,rpc调用
	gameName, ok1 := util.GetServerNameFromGameId(talkMsg.GameId)

	if ok1 == false {

		log.Errorf("打赏时 投注服找不到对应的游戏id的服务名 gameId:%s  userId:%s", talkMsg.GameId, pplayer.UserId)

		return errors.New("RPCBetGame  GetServerNameFromGameId find no gameName  failed")
	}

	nodeId := util.GetGameServiceName(gameName)
	rpcFunc := util.FormatRpcMethod(gameName, "RPCTalkGame")

	var talkInfo rpc.GameTalkChat

	talkInfo.GameId = talkMsg.GameId
	talkInfo.Deskno = talkMsg.Deskno
	talkInfo.Message = talkMsg.Message
	talkInfo.UserId = talkMsg.PlayerId

	//投递到实际的游戏服
	err := self.GoNode(nodeId, rpcFunc, &talkInfo)

	if err != nil {
		log.Errorf("call rpc  talk fail")
	}

	return nil
}
