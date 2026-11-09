package longhuservice

import (
	"fmt"
	"github.com/duanhf2012/origin/v2/event"
	"github.com/duanhf2012/origin/v2/log"
	"time"
	"videogame/common/betnode"
	"videogame/common/collect"
	models "videogame/common/model"
	"videogame/common/msg_struct"
	"videogame/common/proto/msg"
	"videogame/common/proto/rpc"
	"videogame/common/util"
)

// 游戏状态通知大厅
func (self *LongHuService) NotifyLobbyGameStatus(gs *msg_struct.GameStatusMsg) {

	//通知大厅游戏状态改变
	nodeId := util.GetGameServiceName(util.LobbyService)

	if nodeId == "" {
		return
	}

	rpcFun := util.FormatRpcMethod(util.LobbyService, "RPCUpdateGameStatus")

	var notifymsg msg.NotifyLobbyGameStatus

	notifymsg.GameId = gs.GameId
	notifymsg.GameStatus = int32(gs.GameStatus)
	notifymsg.Round = int32(gs.Round)
	notifymsg.DeskNo = int32(gs.DeskNo)
	notifymsg.BetTime = int32(self.GameTable.GetBetSeconds())
	notifymsg.Set = int32(gs.Set)
	notifymsg.IsAnswered = int32(gs.IsAnswered)
	notifymsg.PlayerNum = int32(self.GameTable.GetPlayerNum())

	//通知大厅服务，游戏状态更新了
	self.AsyncCallNode(nodeId, rpcFun, &notifymsg, func(ret *rpc.Int, err error) {
		if err != nil {
			return
		}
	})
}

// 通知大厅游戏开奖消息
func (self *LongHuService) NotifyLobbyGameAnswer(asn *msg_struct.GameOpenAnswerDTX) {

	var asnMs msg.NotifyLobbyGameAnswer

	asnMs.DeskNo = int32(asn.DeskNo)
	asnMs.Set = int32(asn.Set)
	asnMs.Round = int32(asn.Round)
	asnMs.GameId = self.GameTable.GetGameId()
	asnMs.GameType = int32(msg.DataType_DTX)
	asnMs.Winner = int32(asn.Winner)

	//通知大厅游戏开答了
	nodeId := util.GetGameServiceName(util.LobbyService)

	if nodeId == "" {
		return
	}

	////通知大厅，开答了，这个开答方法，每个游戏对大厅调用都不一样，因为开答结构体不一样，后面优化可以改用proto,就可以统一下
	rpcFun := util.FormatRpcMethod(util.LobbyService, "RPCOpenGameAnswer")

	//通知大厅服务，游戏开答结果
	err := self.AsyncCallNode(nodeId, rpcFun, &asnMs, func(ret *rpc.Int, err error) {

		if err != nil {
			return
		}

		log.Infof("LongHu1Service  OnServiceAnswerEvent--->ToLobby RPCOpenGameAnswer   success data:  \n")

	})

	if err != nil {
		log.Errorf("^^^^^^^^^^^^^^^^^^LongHu1Service  OnServiceAnswerEvent--->ToLobby RPCOpenGameAnswer  failed!!!  data:%+v  \n", asnMs)
		return
	}

}

func (self *LongHuService) NotifyLobbyGameHistoryAnswers(answers *msg_struct.GameDtxAnswersPeriodic) {

	var perioMsg msg.NotifyLobbyHistoryAnswers

	perioMsg.GameId = self.GameTable.GetGameId()
	perioMsg.DeskNo = int32(answers.DeskNo)
	perioMsg.Set = int32(answers.Set)
	perioMsg.Round = int32(answers.Round)
	perioMsg.GameStatus = int32(answers.GameStatus)
	perioMsg.DataType = int32(msg.DataType_DTX)

	var aans msg.WinnersDTX

	for i := 0; i < len(answers.Winners); i++ {
		aans.WinnerItems = append(aans.WinnerItems, int32(answers.Winners[i]))
	}

	perioMsg.Answers = util.ProtoToAny(&aans)

	//通知大厅游戏状态改变
	nodeId := util.GetGameServiceName(util.LobbyService)

	if nodeId == "" {
		return
	}

	//这个调用每个游戏对大厅RPC名都不一样
	rpcFun := util.FormatRpcMethod(util.LobbyService, "RPCUpdateGameAnswersPerio")

	//通知大厅服务，游戏状态更新了
	err := self.AsyncCallNode(nodeId, rpcFun, &perioMsg, func(ret *rpc.Int, err error) {
		if err != nil {
			return
		}

		fmt.Printf("Call LobbyService RPCUpdateGameAnswersPerio Success   \n")
	})

	if err != nil {

		//调用不成功的处理代码
		fmt.Printf("Call LobbyService RPCUpdateGameAnswersPerio  Failed!!! \n")

		return
	}
}

// OnServiceEvent 游戏状态改变事件
func (self *LongHuService) OnServiceGameStatusEvent(ev event.IEvent) {

	gs, ok := (ev.(*event.Event).Data).(msg_struct.GameStatusMsg)

	if ok == false {
		return
	}

	//广播通知客户端游戏状态改变了
	self.GameTable.CastToPlayerGameStatus(&self.msgSender, &gs)

	//通知大厅
	self.NotifyLobbyGameStatus(&gs)

	//通知玩家有没有4局没有下注的
	self.GameTable.NotifyPlayerLeaveTable(self.GameTable.GetNoBetRound())

	//	fmt.Printf("……………………………LongHu1Service  OnServiceGameStatusEvent  data:%+v\n", gs)

}

// 广播通知玩家，大厅游戏开答事件
func (self *LongHuService) OnServiceOpenAnswerEvent(ev event.IEvent) {

	asn, ok := (ev.(*event.Event).Data).(msg_struct.GameOpenAnswerDTX)

	if ok == false {
		return
	}

	//广播通知客户端开答内容
	self.GameTable.CastToPlayerOpenAnswer(&self.msgSender, &asn)

	//通知大厅游戏开答
	self.NotifyLobbyGameAnswer(&asn)

}

// 通知发送历史开答表给大厅, 每次开奖后，会触发此事件，当通知大厅，每个场次的第一局也是会触发这个
func (self *LongHuService) OnServiceGameHistoryAnswersToLobbyEvent(ev event.IEvent) {

	answers, ok := (ev.(*event.Event).Data).(msg_struct.GameDtxAnswersPeriodic)

	if ok == false {
		return
	}

	self.NotifyLobbyGameHistoryAnswers(&answers)

	fmt.Printf("Call LobbyService RPCUpdateGameAnswersPerio okokokok  \n")
}

// 通知控端历史开答列表更新了
func (self *LongHuService) OnServiceGameHistoryAnswersEvent(ev event.IEvent) {

	answers, ok := (ev.(*event.Event).Data).(msg_struct.GameDtxAnswersPeriodic)

	if ok == false {
		return
	}

	fmt.Printf("@@@@@@@@@@@@@@@@@@@@@ LongHu1Service  OnServiceGameHistoryAnswersEvent type data:%+v  \n", answers)

	//先通知玩家历史开答变了
	self.GameTable.CastToPlayerHistoryAnswers(&self.msgSender, &answers)

	//需要通知控端历史开答内容变化了
	self.adminCtrl.BroadCastMsg(msg_struct.GameDtxAnswerPeriodicProtocol, &answers)

}

// 游戏投注事件，需要区分是游戏内的压注，还是游戏外压注(多台投注用)
func (self *LongHuService) OnServiceTableBetEvent(ev event.IEvent) {

	betMsg, ok := (ev.(*event.Event).Data).(*rpc.BetToGameRPC)

	totalbet := 0

	if ok == false {

		//fmt.Printf("LongHuService OnServiceTableBetEvent  Failed.... \n  ")
		log.Errorf(" LongHuService OnServiceTableBetEvent  Failed....")

		return
	}

	//fmt.Printf("LongHu1Service OnServiceTableBetEvent ....%v \n ", betMsg)

	player := self.PlayersImp.GetPlayer(betMsg.UserId)

	//若没有找到
	if player == nil {

		log.Errorf(" LongHuService OnServiceTableBetEvent  内存中无法找到玩家userId:%s", betMsg.UserId)

		var betResult msg.STOC_BetGame_Result

		betResult.Success = 1
		betResult.GameId = betMsg.GameId
		betResult.BetMode = int32(betMsg.BetMode)
		betResult.DeskNo = int32(self.GameTable.GetDeskNo())
		betResult.OrderKey = uint32(betMsg.OrderKey)

		betResult.Balance = int32(player.GetCredit())
		betResult.Digits = int32(player.GetDigits())
		betResult.ErrCode = int32(msg.ErrCode_BetErr_ServerPlayerIdErr)

		player.SendMsg(msg.MsgType_GameTable_BetResult, &betResult)

		return
		//需要重新生成一个玩家
	}
	var betResult msg.STOC_BetGame_Result

	if player.Id != betMsg.UserId {

		//处理不成功代理
		betResult.Success = 1
		betResult.ErrCode = int32(msg.ErrCode_BetErr_ServerClusterErr)

	} else {

		if betMsg.Digits != int32(player.GetDigits()) {
			betResult.Success = 0
		}

		betbar := betnode.BetNode{}

		betbar.UserId = betMsg.UserId
		betbar.ClientId = betMsg.ClientId
		betbar.GameId = betMsg.GameId
		betbar.BetTime = time.Now()
		betbar.OrderKey = betMsg.OrderKey
		betbar.BetAreaId = int(betMsg.BetAreadId)
		betbar.BetPoint = betMsg.BetPoint
		betbar.Digits = betMsg.Digits
		betbar.DeskNo = self.GameTable.GetDeskNo()
		betbar.BetMode = int(betMsg.BetMode)
		betbar.Currency = player.GetCurrency()
		betbar.ParentId = player.GetParentId()
		betbar.MerchantId = player.GetMerchantId()
		betbar.MerchantCode = player.GetMerchantCode()
		betbar.RootKey = player.GetRootKey()

		betbar.CreateOrderId(msg_struct.GameTypeDtx)

		tempbet, ret := self.GameTable.PlayerBet(&betbar)

		totalbet = tempbet

		if ret == 0 {

			betResult.Success = 0

			//fmt.Printf("*************LongHu1Service OnServiceTableBetEvent ...……%v Success  !!!!! \n ", *betMsg)
			//log.Infof("玩家压注龙虎游戏: userId:%s  betAreaId: %d , betMoney:%d Digits:%d 成功！", betbar.UserId, betbar.BetAreaId, betbar.BetPoint, betbar.Digits)

		} else {
			betResult.Success = 1
			betResult.ErrCode = int32(ret)

			//fmt.Printf("^^^^^^^^^^^^^^^^^^LongHu1Service OnServiceTableBetEvent ...……%v failed !!!!! \n ", *betMsg)
			log.Infof("玩家压注龙虎游戏: userId:%s  betAreaId: %d , betMoney:%d Digits:%d 不成功！", betbar.UserId, betbar.BetAreaId, betbar.BetPoint, betbar.Digits)
		}
	}

	betResult.BetAreadId = betMsg.BetAreadId
	betResult.BetPoint = betMsg.BetPoint
	betResult.GameId = betMsg.GameId
	betResult.BetMode = int32(betMsg.BetMode)
	betResult.DeskNo = int32(self.GameTable.GetDeskNo())
	betResult.OrderKey = uint32(betMsg.OrderKey)

	betResult.Balance = int32(player.GetCredit())
	betResult.TotalBet = int32(totalbet)
	betResult.Digits = int32(player.GetDigits())

	player.SendMsg(msg.MsgType_GameTable_BetResult, &betResult)

}

// 游戏撤消投注事件
func (self *LongHuService) OnServiceTableCancelBetEvent(ev event.IEvent) {

	betMsg, ok := (ev.(*event.Event).Data).(*rpc.BetToGameRPC)

	if ok == false {

		//fmt.Printf("LongHuService OnServiceTableCancelBetEvent  Failed.... \n  ")
		log.Error(" LongHuService OnServiceTableCancelBetEvent  Failed....")

		return
	}

	//fmt.Printf("LongHu1Service OnServiceTableCancelBetEvent ....%v \n ", betMsg)
	player := self.PlayersImp.GetPlayer(betMsg.UserId)

	var betResult msg.STOC_CancelBetGame_Result

	//若没有找到
	if player == nil {

		log.Errorf("LongHuService  gameId:%s  OnServiceTableCancelBetEvent 内存找不到玩家userId:%s ", self.GameTable.GetGameId(), betMsg.UserId)

		//处理不成功代理
		betResult.Success = 1
		betResult.GameId = betMsg.GameId
		betResult.BetMode = int32(betMsg.BetMode)
		betResult.DeskNo = int32(self.GameTable.GetDeskNo())
		betResult.Balance = int32(player.GetCredit())
		betResult.Digits = int32(player.GetDigits())

		betResult.OrderKey = uint32(betMsg.OrderKey)
		betResult.BetAreadId = betMsg.BetAreadId
		betResult.BetPoint = betMsg.BetPoint
		betResult.TotalBet = 0
		betResult.ErrCode = int32(msg.ErrCode_BetErr_ServerPlayerIdErr)

		player.SendMsg(msg.MsgType_GameTable_CancelBetResult, &betResult)

		return
	}

	if player.Id != betMsg.UserId {

		//处理不成功代理
		betResult.Success = 1
		betResult.ErrCode = int32(msg.ErrCode_BetErr_ServerPlayerIdErr)

	} else {

		betNode, tempbet, ret := self.GameTable.PlayerCancelBet(betMsg.UserId)

		if ret == 0 {
			//撤消成功

			betResult.Success = 0
			betResult.OrderKey = uint32(betNode.OrderKey)
			betResult.BetAreadId = int32(betNode.BetAreaId)
			betResult.BetPoint = betNode.BetPoint
			betResult.TotalBet = int32(tempbet)

			//fmt.Printf("*************LongHu1Service OnServiceTableCancelBetEvent ...……%v Success  !!!!! \n ", *betMsg)

		} else {

			betResult.Success = 1
			betResult.ErrCode = int32(ret)

			//fmt.Printf("^^^^^^^^^^^^^^^^^^LongHu1Service OnServiceTableCancelBetEvent ...……%v failed !!!!! \n ", *betMsg)
		}
	}

	betResult.GameId = betMsg.GameId
	betResult.BetMode = int32(betMsg.BetMode)
	betResult.DeskNo = int32(self.GameTable.GetDeskNo())
	betResult.Balance = int32(player.GetCredit())
	betResult.Digits = int32(player.GetDigits())

	player.SendMsg(msg.MsgType_GameTable_CancelBetResult, &betResult)

}

// 打赏小费事件
func (self *LongHuService) OnServiceTipEvent(ev event.IEvent) {

	tipMsg, ok := (ev.(*event.Event).Data).(*rpc.TipToGameRPC)

	if ok == false {

		//fmt.Printf("LongHuService OnServiceTableBetEvent  Failed.... \n  ")
		log.Error(" LongHuService OnServiceTableBetEvent  Failed....")
		return
	}

	log.Infof("gameId:%s 玩家打赏 userId:%s   tipMoney:%d Digits:%d  开始处理！",
		self.GameTable.GetGameId(), tipMsg.UserId, tipMsg.TipPoint, tipMsg.Digits)

	//fmt.Printf("LongHu1Service OnServiceTableBetEvent ....%v \n ", tipMsg)

	player := self.PlayersImp.GetPlayer(tipMsg.UserId)

	//若没有找到
	if player == nil {
		//需要重新生成一个玩家
	}

	var tipResult msg.STOC_TipMoneyResult

	tipCheck := true

	//检测打赏
	if tipMsg.TipPoint < uint32(player.MinTip) {

		tipCheck = false

	} else if tipMsg.TipPoint > uint32(player.MaxTip) {
		tipCheck = false
	}

	if tipCheck == false {

		tipResult.TipMoney = tipMsg.TipPoint
		tipResult.OrderKey = uint32(tipMsg.OrderKey)
		tipResult.Digits = tipMsg.Digits

		tipResult.GameId = self.GameTable.GetGameId()
		tipResult.DeskNo = int32(self.GameTable.GetDeskNo())

		tipResult.Success = 1
		tipResult.ErrCode = int32(msg.ErrCode_BetErr_ServerTipErr)

		player.SendMsg(msg.MsgType_GameTable_TipMoneyResult, &tipResult)

		log.Errorf("打赏金额不符合要求 userId:%s  error!  ", tipMsg.UserId)

		return
	}

	tipBar := betnode.TipNode{}

	tipBar.DealerId = tipMsg.DealerId
	tipBar.UserId = tipMsg.UserId
	tipBar.OrderKey = tipMsg.OrderKey

	tipBar.TipPoint = tipMsg.TipPoint
	tipBar.Digits = tipMsg.Digits
	tipBar.DeskNo = self.GameTable.GetDeskNo()

	tipBar.Currency = player.GetCurrency()
	tipBar.ParentId = player.GetParentId()
	tipBar.MerchatnCode = player.GetMerchantCode()
	tipBar.MerchantId = player.GetMerchantId()
	tipBar.RootKey = player.GetRootKey()

	tipBar.GameId = self.GameTable.GetGameId()
	tipBar.DealerId = self.GameTable.GameInfo.GetDealerId() //打赏给哪个荷官

	tipBar.TipTime = time.Now()

	ok, err := self.GameTable.PlayerTip(player, &tipBar)

	tipResult.TipMoney = tipBar.TipPoint
	tipResult.OrderKey = uint32(tipBar.OrderKey)
	tipResult.Digits = tipBar.Digits

	tipResult.GameId = self.GameTable.GetGameId()
	tipResult.DeskNo = int32(self.GameTable.GetDeskNo())

	if ok == false {

		tipResult.Success = 1
		tipResult.ErrCode = int32(err)

		player.SendMsg(msg.MsgType_GameTable_TipMoneyResult, &tipResult)

		log.Error(" LongHuService OnServiceTipEvent Failed....")

		return
	}
	tipResult.Success = 0

	player.SendMsg(msg.MsgType_GameTable_TipMoneyResult, &tipResult)

	//再发送一次玩家余额消息
	self.GameTable.SendToPlayerCardInfo(player)

}

func (self *LongHuService) SaveAnswerToDB(asn *msg_struct.GameOpenAnswerDTX) bool {

	//开答结果存数据库
	var dbTable models.SGAnswerDTX

	asn.DeskNo = self.GameTable.GetDeskNo()
	asn.Set = self.GameTable.GameInfo.Set
	asn.Round = self.GameTable.GameInfo.Round

	err := dbTable.AppendToTable(self.GameTable.GameInfo.RoundId, self.GameTable.GameInfo.GameSubCode, asn, self.GameTable.GameInfo.GetDealerId())

	if err != nil {

		log.Errorf("SaveAnswerToDB AppendToTable Failed. %v ", asn)

		return false
	}

	return err == nil

}

// OnServiceTableSettleEvent  开奖结算事件
func (self *LongHuService) OnServiceAnswerSettleEvent(ev event.IEvent) {

	answer, ok := (ev.(*event.Event).Data).(*msg_struct.GameOpenAnswerDTX)

	if ok == false {
		//出错处理
	}

	//fmt.Printf("OnServiceAnswerSettleEvent .....%v \n ", answer)
	log.Infof("open answer gameId :%s  winner:%d answer:%v ", answer.GameId, answer.Winner, *answer)

	//开答结果存数据库
	self.SaveAnswerToDB(answer)

	//计算胜负
	self.GameTable.SettleWinLost(answer)

	//把有投注，并且有赢的通知对应玩家
	self.GameTable.SendPlayersWin()

	//把玩家余额下发
	self.GameTable.SendPlayersCardInfo()

	//通知商户单钱包服务，游戏开答了
	self.GameTable.NotifyWalletServiceUpdateTrans()

	//清除玩家投注
	self.GameTable.ClearMapPlayerBet()

}

// 广播玩家钱包
func (self *LongHuService) OnBroadPlayerWalletEvent(ev event.IEvent) {

	self.GameTable.SendPlayersCardInfo()
}

// 通知更新玩家在线人数
func (self *LongHuService) OnServiceNotifyPlayerNumEvent(ev event.IEvent) {

	var playerInfo rpc.GameOnlinePlayerNum

	playerInfo.GameId = self.GameTable.GetGameId()
	playerInfo.DeskNo = uint32(self.GameTable.GetDeskNo())

	playerInfo.OnlineNum = uint32(self.GetPlayerNum())
	playerInfo.DealerId = self.GameTable.GameInfo.GetDealerId()
	playerInfo.Unixtime = time.Now().Unix()

	//通知大厅的RPC

	//通知实际游戏服务器更新游戏状态
	nodeId := util.GetGameServiceName(util.LobbyService) //请求目的游戏服务的nodeId
	rpcFunc := util.FormatRpcMethod(util.LobbyService, "RPCUpdateOnlinePlayerNum")

	if nodeId == "" {

		log.Errorf("********LobbyService.RPCUpdateOnlinePlayerNum  :%s  find no game  nodeId !!!", util.LobbyService)

		return
	}

	//通知大厅玩家人数
	self.GoNode(nodeId, rpcFunc, &playerInfo)

}

// 处理玩家聊天事件
func (self *LongHuService) OnServiceTalkChatEvent(ev event.IEvent) {

	talkMsg, ok := (ev.(*event.Event).Data).(*rpc.GameTalkChat)

	if ok == false {

		log.Error(" BacxService OnServiceTalkChatEvent  Failed....")
		return
	}
	self.TalkChatBroad(talkMsg.UserId, talkMsg.Message)
}

// 广播此聊天记录
func (self *LongHuService) TalkChatBroad(userId string, message string) {

	var talkChat msg.STOC_TalkChat

	bytes := []byte(userId)
	size := len(userId)

	//把玩家字符串尾数填”*“
	if size < 5 {
		for i := size - 1; i > size-2; i-- {
			bytes[i] = '*'
		}
	} else {
		for i := size - 1; i > size-5; i-- {
			bytes[i] = '*'
		}
	}

	newId := string(bytes)

	talkChat.Message = message
	talkChat.PlayerId = newId
	talkChat.GameId = self.GameTable.GetGameId()
	talkChat.Deskno = int32(self.GameTable.GetDeskNo())
	talkChat.MsgType = 0

	allPlayer := self.PlayersImp.GetAllPlayer()

	//把聊天转发给对应的玩家
	for _, player := range allPlayer {

		if player.UserId == userId {
			//continue
			talkChat.Unused1 = 1
		} else {
			talkChat.Unused1 = 0
		}

		player.SendMsg(msg.MsgType_GameTable_TalkChatShow, &talkChat)
	}

	//聊天记录存Mongo数据库
	talkChatCollect := &collect.TalkChat{
		Message:  message,
		PlayerId: userId, //存储没有隐藏的用户Id（玩家在获取聊天列表的时候，需要判断该条消息是否是本人的聊天内容）
		GameId:   self.GameTable.GetGameId(),
		Deskno:   int32(self.GameTable.GetDeskNo()),
		MsgType:  0,
		CreateAt: util.GetNowDatetime(),
	}

	_, err := talkChatCollect.InsertOne()
	if err != nil {
		log.Error(" TalkChatBroad InsertOne Failed....")
	}

}
