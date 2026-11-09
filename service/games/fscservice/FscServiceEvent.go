package fscservice

import (
	"fmt"
	"github.com/duanhf2012/origin/v2/event"
	"github.com/duanhf2012/origin/v2/log"
	log2 "log"
	"time"
	"videogame/common/betnode"
	"videogame/common/collect"
	models "videogame/common/model"
	"videogame/common/msg_struct"
	"videogame/common/proto/msg"
	"videogame/common/proto/rpc"
	"videogame/common/util"
	"videogame/service/games/fscservice/gameTableFsc"
)

// 游戏状态通知大厅
func (self *FscService) NotifyLobbyGameStatus(gs *msg_struct.GameStatusMsg) {

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
func (self *FscService) NotifyLobbyGameAnswer(asn *msg_struct.FscOpenAnswer) {

	var asnMs msg.NotifyLobbyGameAnswer

	asnMs.DeskNo = int32(asn.DeskNo)
	asnMs.Set = int32(asn.Set)
	asnMs.Round = int32(asn.Round)
	asnMs.GameId = self.GameTable.GetGameId()
	asnMs.GameType = int32(msg.DataType_FSC)
	asnMs.Winner = int32(asn.Winner)
	asnMs.Val1 = int32(gameTableFsc.FscAnswerInds[asn.Winner][0])
	asnMs.Val2 = int32(gameTableFsc.FscAnswerInds[asn.Winner][1])
	asnMs.Val3 = int32(gameTableFsc.FscAnswerInds[asn.Winner][2])

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

		log.Errorf("--^^^^^^^^^^^^^^^^^^FscService  OnServiceAnswerEvent--->ToLobby RPCOpenGameAnswer   success data:  \n")

	})

	if err != nil {
		log.Errorf("^^^^^^^^^^^^^^^^^^FscService  OnServiceAnswerEvent--->ToLobby RPCOpenGameAnswer  failed!!!  data:%+v  \n", asnMs)
		return
	}

}

// OnServiceEvent 文播游戏状态改变事件，发给各个玩家端，同时通知大厅
func (self *FscService) OnServiceGameStatusEvent(ev event.IEvent) {

	gs, ok := (ev.(*event.Event).Data).(msg_struct.GameStatusMsg)

	if ok == false {
		return
	}

	gs.GameId = self.GameTable.GetGameId()

	//广播通知客户端游戏状态改变了
	self.GameTable.CastToPlayerGameStatus(&gs)

	fmt.Printf("……………………………………………………………FscService  OnServiceGameStatusEvent  data:%+v\n", gs)

	self.NotifyLobbyGameStatus(&gs)

	//通知玩家有没有4局没有下注的
	self.GameTable.NotifyPlayerLeaveTable(self.GameTable.GetNoBetRound())

}

// 服务监听事件回调,游戏开答事件
func (self *FscService) OnServiceOpenAnswerEvent(ev event.IEvent) {

	asn, ok := (ev.(*event.Event).Data).(msg_struct.FscOpenAnswer)

	if ok == false {
		return
	}

	//广播通知客户端开答内容
	self.GameTable.CastToPlayerOpenAnswer(&asn)

	self.NotifyLobbyGameAnswer(&asn)
	//fmt.Printf("^^^^^^^^^^^^^^^^^^FscService  OnServiceAnswerEvent  data:%+v  \n", asn)

}

// 通知发送历史开答表给大厅
func (self *FscService) OnServiceGameHistoryAnswersToLobbyEvent(ev event.IEvent) {

	answers, ok := (ev.(*event.Event).Data).(msg_struct.FscAnswersPeriodic)

	if ok == false {
		return
	}

	var perioMsg msg.NotifyLobbyHistoryAnswers

	var aans msg.WinnersFsc

	//通知大厅，最后20局的路子列表
	end := len(answers.Winners)
	start := end - msg_struct.HistoryAnswersPerioSize //长度定为20

	if start < 0 {
		start = 0
	}

	//for i := 0; i < len(answers.Winners); i++ {
	for i := start; i < end; i++ {
		aans.WinnerItems = append(aans.WinnerItems, uint32(answers.Winners[i]))
	}

	aans.Totalnum = int32(len(answers.Winners))
	aans.Lastnum = int32(end - start)

	perioMsg.GameId = self.GameTable.GetGameId()
	perioMsg.DeskNo = int32(answers.DeskNo)
	perioMsg.Set = int32(answers.Set)
	perioMsg.Round = int32(answers.Round)
	perioMsg.GameStatus = int32(answers.GameStatus)
	perioMsg.DataType = int32(msg.DataType_FSC)

	perioMsg.Answers = util.ProtoToAny(&aans)

	//通知大厅游戏状态改变
	nodeId := util.GetGameServiceName(util.LobbyService)

	if nodeId == "" {
		return
	}

	rpcFun := util.FormatRpcMethod(util.LobbyService, "RPCUpdateGameAnswersPerio")

	//通知大厅服务，游戏状态更新了
	err := self.AsyncCallNode(nodeId, rpcFun, &perioMsg, func(ret *rpc.Int, err error) {

		if err != nil {
			return
		}
	})

	if err != nil {

		fmt.Printf("FscService ----->Call LobbyService RPCUpdateGameAnswersPerio error!!! \n")
	}

	fmt.Printf("FscService ----->Call LobbyService RPCUpdateGameAnswersPerio \n")

}

// 通知控端历史开答列表更新了
func (self *FscService) OnServiceGameHistoryAnswersEvent(ev event.IEvent) {

	answers, ok := (ev.(*event.Event).Data).(msg_struct.FscAnswersPeriodic)

	if ok == false {
		return
	}

	fmt.Printf("FscService---> OnServiceGameHistoryAnswersEvent type data:%+v  \n", answers)

	//先通知玩家历史开答变了
	self.GameTable.CastToPlayerHistoryAnswers(&answers)

	//需要通知控端历史开答内容变化了
	self.adminCtrl.BroadCastMsg(msg_struct.GameFscAnswerPeriodicProtocol, &answers)

}

// 游戏投注事件，需要区分是游戏内的压注，还是游戏外压注(多台投注用)
func (self *FscService) OnServiceTableBetEvent(ev event.IEvent) {

	betMsg, ok := (ev.(*event.Event).Data).(*rpc.BetToGameRPC)

	totalbet := 0

	if ok == false {

		fmt.Printf("FscService OnServiceTableBetEvent  Failed.... \n  ")
		log.Error("FscService OnServiceTableBetEvent  Failed....")

		return
	}

	fmt.Printf("FscService OnServiceTableBetEvent ....%v \n ", *betMsg)

	player := self.PlayersImp.GetPlayer(betMsg.UserId)

	log2.Printf("1111----->clientId = %v ,BetPoint = %v , OrderKey = %v", player.ClientId, betMsg.BetPoint, betMsg.OrderKey)
	log.Errorf("1111----->clientId = %v ,BetPoint = %v , OrderKey = %v", player.ClientId, betMsg.BetPoint, betMsg.OrderKey)

	//若没有找到
	if player == nil {
		//需要重新生成一个玩家
	}

	var betResult msg.STOC_BetGame_Result

	if player.Id != betMsg.UserId {

		//处理不成功代理
		betResult.Success = 1
		betResult.ErrCode = int32(msg.ErrCode_BetErr_ServerClusterErr)

	} else {
		betbar := betnode.BetNode{}

		betbar.UserId = betMsg.UserId
		betbar.ClientId = betMsg.ClientId
		betbar.BetTime = time.Now()
		betbar.OrderKey = betMsg.OrderKey
		betbar.BetAreaId = int(betMsg.BetAreadId)
		betbar.BetPoint = betMsg.BetPoint
		betbar.Digits = betMsg.Digits
		betbar.BetMode = int(betMsg.BetMode)

		betbar.GameId = self.GameTable.GetGameId()
		betbar.DeskNo = self.GameTable.GetDeskNo()

		betbar.Currency = player.GetCurrency()
		betbar.ParentId = player.GetParentId()
		betbar.MerchantId = player.GetMerchantId()
		betbar.MerchantCode = player.GetMerchantCode()
		betbar.RootKey = player.GetRootKey()

		betbar.CreateOrderId(msg_struct.GameTypeFSC)

		tempbet, ret := self.GameTable.PlayerBet(&betbar)

		totalbet = tempbet

		if ret == 0 {
			betResult.Success = 0

			fmt.Printf("*************FscService OnServiceTableBetEvent ...……%v ok !!!!! \n ", *betMsg)

		} else {
			betResult.Success = 1
			betResult.ErrCode = int32(ret)

			fmt.Printf("^^^^^^^^^^^^^^^^^^FscService OnServiceTableBetEvent ...……%v failed !!!!! ret:%d \n ", *betMsg, ret)
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

	log2.Printf("2222----->clientId = %v ,BetPoint = %v , OrderKey = %v", player.ClientId, betMsg.BetPoint, betMsg.OrderKey)
	log.Errorf("2222----->clientId = %v ,BetPoint = %v , OrderKey = %v", player.ClientId, betMsg.BetPoint, betMsg.OrderKey)

	player.SendMsg(msg.MsgType_GameTable_BetResult, &betResult)

}

// 游戏投注事件，需要区分是游戏内的压注，还是游戏外压注(多台投注用)
func (self *FscService) OnServiceTableCancelBetEvent(ev event.IEvent) {

	betMsg, ok := (ev.(*event.Event).Data).(*rpc.BetToGameRPC)

	if ok == false {

		fmt.Printf("FscService OnServiceTableCancelBetEvent  Failed.... \n  ")
		log.Error(" FscService OnServiceTableCancelBetEvent  Failed....")

		return
	}

	fmt.Printf("FscService OnServiceTableCancelBetEvent ....%v \n ", betMsg)

	player := self.PlayersImp.GetPlayer(betMsg.UserId)

	//若没有找到
	if player == nil {
		//需要重新生成一个玩家,需要去大厅服，读取玩家数据

		return
	}
	var betResult msg.STOC_CancelBetGame_Result

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

			fmt.Printf("*************FscService OnServiceTableCancelBetEvent ...……%v Success  !!!!! \n ", *betMsg)

		} else {

			betResult.Success = 1
			betResult.ErrCode = int32(ret)

			fmt.Printf("^^^^^^^^^^^^^^^^^^FscService OnServiceTableCancelBetEvent ...……%v failed !!!!! \n ", *betMsg)
		}
	}

	betResult.GameId = betMsg.GameId
	betResult.BetMode = betMsg.BetMode
	betResult.DeskNo = int32(self.GameTable.GetDeskNo())
	betResult.Balance = int32(player.GetCredit())
	betResult.Digits = int32(player.GetDigits())

	player.SendMsg(msg.MsgType_GameTable_CancelBetResult, &betResult)

}

func (self *FscService) SaveAnswerToDB(asn *msg_struct.FscOpenAnswer) bool {

	//开答结果存数据库
	var dbTable models.SGAnswerFsc

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

// OnServiceTableSettleEvent 结算事件
func (self *FscService) OnServiceAnswerSettleEvent(ev event.IEvent) {

	asn, ok := (ev.(*event.Event).Data).(*msg_struct.FscOpenAnswer)

	if ok == false {
		//出错处理
	}

	fmt.Printf("OnServiceAnswerSettleEvent .....%v \n ", asn)

	//存储开答结果
	self.SaveAnswerToDB(asn)

	//开答
	self.GameTable.SettleWinLost(asn)

	//把有投注，并且有赢的通知对应玩家
	self.GameTable.SendPlayersWin()

	//把玩家余额下发
	self.GameTable.SendPlayersCardInfo()

	//通知商户单钱包服务，游戏开答了
	self.GameTable.NotifyWalletServiceUpdateTrans()

	self.GameTable.ClearMapPlayerBet()

}

// 打赏小费事件
func (self *FscService) OnServiceTipEvent(ev event.IEvent) {

	tipMsg, ok := (ev.(*event.Event).Data).(*rpc.TipToGameRPC)

	if ok == false {

		fmt.Printf("FscService OnServiceTableBetEvent  Failed.... \n  ")
		log.Error(" FscService OnServiceTableBetEvent  Failed....")

		return
	}

	fmt.Printf("FscService OnServiceTableBetEvent ....%v \n ", tipMsg)

	player := self.PlayersImp.GetPlayer(tipMsg.UserId)

	//若没有找到
	if player == nil {
		//需要重新生成一个玩家
	}

	var tipResult msg.STOC_TipMoneyResult

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

}

// 通知更新玩家在线人数
func (self *FscService) OnServiceNotifyPlayerNumEvent(ev event.IEvent) {

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

		log.Debugf("*************LobbyService.RPCUpdateOnlinePlayerNum  :%s  find no game  nodeId !!!", util.LobbyService)

		return
	}

	//通知大厅玩家人数
	self.GoNode(nodeId, rpcFunc, &playerInfo)

}

// 处理玩家聊天事件
func (self *FscService) OnServiceTalkChatEvent(ev event.IEvent) {

	talkMsg, ok := (ev.(*event.Event).Data).(*rpc.GameTalkChat)

	if ok == false {

		log.Error(" BacxService OnServiceTalkChatEvent  Failed....")
		return
	}
	self.TalkChatBroad(talkMsg.UserId, talkMsg.Message)
}

// 广播此聊天记录
func (self *FscService) TalkChatBroad(userId string, message string) {

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
