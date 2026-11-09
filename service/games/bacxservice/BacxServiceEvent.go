package bacxservice

import (
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
func (self *BacxService) NotifyLobbyGameStatus(gs *msg_struct.GameStatusMsg) {

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
func (self *BacxService) NotifyLobbyGameAnswer(asn *msg_struct.GameOpenAnswerBacx) {

	var asnMs msg.NotifyLobbyGameAnswer

	asnMs.DeskNo = int32(asn.DeskNo)
	asnMs.Set = int32(asn.Set)
	asnMs.Round = int32(asn.Round)
	asnMs.GameId = self.GameTable.GetGameId()
	asnMs.GameType = int32(msg.DataType_BACX)
	asnMs.Winner = int32(asn.Winner.Winner)
	asnMs.Val1 = int32(asn.Winner.PPair)
	asnMs.Val2 = int32(asn.Winner.BPair)

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

		//	log.Info("^^^BacxService  NotifyLobbyGameAnswer--->ToLobby RPCOpenGameAnswer   success data:  \n")
	})

	if err != nil {
		log.Errorf("^^^^^^^^^^^^^^^^^^BacxService  NotifyLobbyGameAnswer--->ToLobby RPCOpenGameAnswer  failed!!!  data:%+v  \n", asnMs)
		return
	}

}

// 通知大厅历史开答记录
func (self *BacxService) NotifyLobbyHistoryGames(answers *msg_struct.GameBacXAnswersPeriodic) {

	var perioMsg msg.NotifyLobbyHistoryAnswers

	var Winners msg.WinnersBACX

	perioMsg.GameId = answers.GameId
	perioMsg.DeskNo = int32(answers.DeskNo)
	perioMsg.Set = int32(answers.Set)
	perioMsg.Round = int32(answers.Round)
	perioMsg.GameStatus = int32(answers.GameStatus)
	perioMsg.DataType = int32(msg.DataType_BACX)

	for i := 0; i < len(answers.Winners); i++ {

		Winner := msg.WinnerBACX{}

		Winner.Winner = uint32(answers.Winners[i].Winner)
		Winner.PPair = uint32(answers.Winners[i].PPair)
		Winner.BPair = uint32(answers.Winners[i].BPair)

		Winners.WinnerItems = append(Winners.WinnerItems, &Winner)
	}

	perioMsg.Answers = util.ProtoToAny(&Winners)

	//通知大厅游戏状态改变
	nodeId := util.GetGameServiceName(util.LobbyService)

	if nodeId == "" {
		return
	}

	rpcFun := util.FormatRpcMethod(util.LobbyService, "RPCUpdateGameAnswersPerio")

	//通知大厅服务，游戏状态更新了
	self.AsyncCallNode(nodeId, rpcFun, &perioMsg, func(ret *rpc.Int, err error) {
		if err != nil {
			return
		}

	})

}

// OnServiceEvent 文播游戏状态改变事件，发给各个玩家端，同时通知大厅
func (self *BacxService) OnServiceGameStatusEvent(ev event.IEvent) {

	gs, ok := (ev.(*event.Event).Data).(msg_struct.GameStatusMsg)

	if ok == false {
		return
	}

	gs.GameId = self.GameTable.GetGameId()

	//广播通知客户端游戏状态改变了
	self.GameTable.CastToPlayerGameStatus(&gs)

	//	fmt.Printf("……………………………………………………………BacxService  OnServiceGameStatusEvent  data:%+v\n", gs)

	//通知大厅游戏状态改变
	self.NotifyLobbyGameStatus(&gs)

	//通知超出未下注局数的玩家退出游戏
	self.GameTable.NotifyPlayerLeaveTable(self.GameTable.GetNoBetRound())

}

// 服务监听事件回调,游戏开答事件
func (self *BacxService) OnServiceOpenAnswerEvent(ev event.IEvent) {

	asn, ok := (ev.(*event.Event).Data).(*msg_struct.GameOpenAnswerBacx)

	if ok == false {
		return
	}
	if asn == nil {
		return
	}

	log.Infof("^^^^BacxService  OnServiceAnswerEvent  gameId:%s ,win:%d  \n", asn.GameId, asn.Winner)

	//fmt.Printf("^^^^^^^^^^^^^^^^^^BacxService  OnServiceAnswerEvent  data:%+v  \n", asn)

	//广播通知客户端开答内容
	self.GameTable.CastToPlayerOpenAnswer(asn)

	//通知大厅开答
	self.NotifyLobbyGameAnswer(asn)

}

// 通知发送历史开答表给大厅
func (self *BacxService) OnServiceGameHistoryAnswersToLobbyEvent(ev event.IEvent) {

	answers, ok := (ev.(*event.Event).Data).(*msg_struct.GameBacXAnswersPeriodic)

	if ok == false {
		return
	}

	self.NotifyLobbyHistoryGames(answers)

	//	fmt.Printf("BacxService ----->Call LobbyService RPCUpdateGameAnswersPerio \n")

}

// 通知控端历史开答列表更新了
func (self *BacxService) OnServiceGameHistoryAnswersEvent(ev event.IEvent) {

	answers, ok := (ev.(*event.Event).Data).(*msg_struct.GameBacXAnswersPeriodic)

	if ok == false {
		return
	}

	//fmt.Printf("BacxService---> OnServiceGameHistoryAnswersEvent type data:%+v  \n", answers)

	//先通知玩家历史开答变了
	self.GameTable.CastToPlayerHistoryAnswers(answers)

	//需要通知控端历史开答内容变化了
	self.adminCtrl.BroadCastMsg(msg_struct.GameBacxAnswerPeriodicProtocol, answers)

}

// 游戏投注事件，需要区分是游戏内的压注，还是游戏外压注(多台投注用)
func (self *BacxService) OnServiceTableBetEvent(ev event.IEvent) {

	betMsg, ok := (ev.(*event.Event).Data).(*rpc.BetToGameRPC)

	totalbet := 0

	if ok == false {

		//fmt.Printf("BacxService OnServiceTableBetEvent  Failed.... \n  ")
		log.Errorf(" BacxService OnServiceTableBetEvent gameId:%s  Failed....", self.GameTable.GetGameId())

		return
	}

	//fmt.Printf("BacxService OnServiceTableBetEvent  gameId:%s \n ",  self.GameTable.GameInfo.GameId  )

	player := self.PlayersImp.GetPlayer(betMsg.UserId)

	//若没有找到
	if player == nil {

		log.Errorf(" BacxService OnServiceTableBetEvent gameId:%s  Failed....内存找不到玩家", self.GameTable.GetGameId())
		//需要重新生成一个玩家

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
	}

	var betResult msg.STOC_BetGame_Result

	if player.Id != betMsg.UserId {

		//处理不成功代理
		betResult.Success = 1
		betResult.ErrCode = int32(msg.ErrCode_BetErr_ServerClusterErr)

	} else {

		//以后需要改成从内存池列表中读取
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

		betbar.CreateOrderId(msg_struct.GameTypeBacx)

		if betbar.Digits == 0 {
			betbar.Digits = 1
		}

		//投注
		tempbet, ret := self.GameTable.PlayerBet(&betbar)

		totalbet = tempbet

		if ret == 0 {
			betResult.Success = 0
			//fmt.Printf("*************BacxService OnServiceTableBetEvent ...……%v ok !!!!! \n ", *betMsg)
		} else {
			betResult.Success = 1
			betResult.ErrCode = int32(ret)
			//fmt.Printf("^^^^^^^^^^^^^^^^^^BacxService OnServiceTableBetEvent ...……%v failed !!!!! ret:%d \n ", *betMsg, ret)
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

	log.Infof("send bet Result gameId:%s , userId:%s credit :%d   orkey:%d ", betMsg.GameId, betMsg.UserId, betResult.Balance, betResult.OrderKey)

}

// 游戏投注事件，需要区分是游戏内的压注，还是游戏外压注(多台投注用)
func (self *BacxService) OnServiceTableCancelBetEvent(ev event.IEvent) {

	betMsg, ok := (ev.(*event.Event).Data).(*rpc.BetToGameRPC)

	if ok == false {

		//	fmt.Printf("BacxService OnServiceTableCancelBetEvent  Failed.... \n  ")
		log.Error(" BacxService OnServiceTableCancelBetEvent  Failed....")

		return
	}

	//fmt.Printf("BacxService   OnServiceTableCancelBetEvent ....%v \n ", betMsg)

	log.Infof("BacxService  gameId:%s  OnServiceTableCancelBetEvent  ", self.GameTable.GetGameId())

	player := self.PlayersImp.GetPlayer(betMsg.UserId)

	var betResult msg.STOC_CancelBetGame_Result

	//若没有找到
	if player == nil {

		log.Errorf("BacxService  gameId:%s  OnServiceTableCancelBetEvent 内存找不到玩家：%s ", self.GameTable.GetGameId(), betMsg.UserId)

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
		//需要重新生成一个玩家
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

			//fmt.Printf("*************BacxService OnServiceTableCancelBetEvent ...……%v Success  !!!!! \n ", *betMsg)

		} else {

			betResult.Success = 1
			betResult.ErrCode = int32(ret)

			//fmt.Printf("^^^^^^^^^^^^^^^^^^BacxService OnServiceTableCancelBetEvent ...……%v failed !!!!! \n ", *betMsg)
		}
	}

	betResult.GameId = betMsg.GameId
	betResult.BetMode = int32(betMsg.BetMode)
	betResult.DeskNo = int32(self.GameTable.GetDeskNo())
	betResult.Balance = int32(player.GetCredit())
	betResult.Digits = int32(player.GetDigits())

	player.SendMsg(msg.MsgType_GameTable_CancelBetResult, &betResult)

}

// 处理打赏小费事件
func (self *BacxService) OnServiceTipEvent(ev event.IEvent) {

	tipMsg, ok := (ev.(*event.Event).Data).(*rpc.TipToGameRPC)

	if ok == false {

		//fmt.Printf("BacxService OnServiceTipEvent  Failed.... \n  ")
		log.Error(" BacxService OnServiceTipEvent  Failed....")
		return
	}

	log.Infof("BacxService-->OnServiceTipEvent gameId:%s 玩家打赏 userId:%s   tipMoney:%d Digits:%d  开始处理！",
		self.GameTable.GetGameId(), tipMsg.UserId, tipMsg.TipPoint, tipMsg.Digits)

	//fmt.Printf("BacxService OnServiceTipEvent ....%v \n ", tipMsg)

	player := self.PlayersImp.GetPlayer(tipMsg.UserId)

	//若没有找到
	if player == nil {

		//需要重新生成一个玩家,去大厅服读取生成
		log.Errorf("BacxService-->OnServiceTipEvent 内存中找不到玩家:%s", tipMsg.UserId)

		var tipResult msg.STOC_TipMoneyResult

		tipResult.TipMoney = tipMsg.TipPoint
		tipResult.OrderKey = uint32(tipMsg.OrderKey)
		tipResult.Digits = tipMsg.Digits

		tipResult.GameId = self.GameTable.GetGameId()
		tipResult.DeskNo = int32(self.GameTable.GetDeskNo())

		tipResult.Success = 1
		tipResult.ErrCode = int32(msg.ErrCode_BetErr_ServerPlayerIdErr)

		player.SendMsg(msg.MsgType_GameTable_TipMoneyResult, &tipResult)

		return
	}

	tipCheck := true

	//检测打赏
	if tipMsg.TipPoint < uint32(player.MinTip) {

		tipCheck = false

	} else if tipMsg.TipPoint > uint32(player.MaxTip) {
		tipCheck = false
	}

	var tipResult msg.STOC_TipMoneyResult

	if tipCheck == false {

		tipResult.TipMoney = tipMsg.TipPoint
		tipResult.OrderKey = uint32(tipMsg.OrderKey)
		tipResult.Digits = tipMsg.Digits

		tipResult.GameId = self.GameTable.GetGameId()
		tipResult.DeskNo = int32(self.GameTable.GetDeskNo())

		tipResult.Success = 1
		tipResult.ErrCode = int32(msg.ErrCode_BetErr_ServerTipErr)

		player.SendMsg(msg.MsgType_GameTable_TipMoneyResult, &tipResult)

		log.Errorf("BacxService-->OnServiceTipEvent 打赏金额不符合要求 userId:%s  error!  ", tipMsg.UserId)

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

		log.Errorf("BacxService-->OnServiceTipEvent 打赏金额不成功 userId:%s  error:%d  ", tipMsg.UserId, err)

		return
	}

	tipResult.Success = 0

	player.SendMsg(msg.MsgType_GameTable_TipMoneyResult, &tipResult)

	//再发送一次玩家余额消息
	self.GameTable.SendToPlayerCardInfo(player)

}

// 开答存入数据表
func (self *BacxService) SaveAnswerToDB(asn *msg_struct.GameOpenAnswerBacx) bool {

	//开答结果存数据库
	var dbTable models.SGAnswerBacx

	asn.DeskNo = self.GameTable.GetDeskNo()
	asn.Set = self.GameTable.GameInfo.Set
	asn.Round = self.GameTable.GameInfo.Round

	err := dbTable.AppendToTable(self.GameTable.GameInfo.RoundId, self.GameTable.GameInfo.GameSubCode, asn, self.GameTable.GameInfo.GetDealerId())

	if err != nil {

		log.Errorf("SaveAnswerToDB AppendToTable Failed. %v ", asn)

		return false
	}
	log.Infof("gameId:%s SaveAnswerToDB AppendToTable Success ", self.GameTable.GetGameId())

	return err == nil

}

// OnServiceTableSettleEvent 结算事件
func (self *BacxService) OnServiceAnswerSettleEvent(ev event.IEvent) {

	asn, ok := (ev.(*event.Event).Data).(*msg_struct.GameOpenAnswerBacx)

	if ok == false {
		//出错处理
	}

	//fmt.Printf("OnServiceAnswerSettleEvent .....%v \n ", asn)

	//开答结果存数据库
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

// 通知大厅,更新玩家在线人数
func (self *BacxService) OnServiceNotifyPlayerNumEvent(ev event.IEvent) {

	var playerInfo rpc.GameOnlinePlayerNum

	playerInfo.GameId = self.GameTable.GetGameId()
	playerInfo.DeskNo = uint32(self.GameTable.GetDeskNo())

	playerInfo.OnlineNum = uint32(self.GetPlayerNum())
	playerInfo.DealerId = self.GameTable.GameInfo.GetDealerId()
	playerInfo.Unixtime = time.Now().Unix()

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
func (self *BacxService) OnServiceTalkChatEvent(ev event.IEvent) {

	talkMsg, ok := (ev.(*event.Event).Data).(*rpc.GameTalkChat)

	if ok == false {

		log.Error(" BacxService OnServiceTalkChatEvent  Failed....")

		return
	}

	//fmt.Printf("BacxService OnServiceTalkChatEvent ....%v \n ", talkMsg)

	//player := self.PlayersImp.GetPlayer(talkMsg.UserId)
	//
	////若没有找到
	//if player == nil {
	//	//需要重新生成一个玩家
	//
	//}

	self.TalkChatBroad(talkMsg.UserId, talkMsg.Message)

}

// 广播此聊天记录
func (self *BacxService) TalkChatBroad(userId string, message string) {

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
			talkChat.Unused1 = 1 //表示自已
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
