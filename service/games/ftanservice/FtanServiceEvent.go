package ftanservice

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
func (self *FtanService) NotifyLobbyGameStatus(gs *msg_struct.GameStatusMsg) {
	nodeId := util.GetGameServiceName(util.LobbyService)
	if nodeId == "" {
		return
	}

	rpcFun := util.FormatRpcMethod(util.LobbyService, "RPCUpdateGameStatus")
	if rpcFun == "" {
		return
	}

	notifyMsg := msg.NotifyLobbyGameStatus{
		GameId:     gs.GameId,
		GameStatus: int32(gs.GameStatus),
		Round:      int32(gs.Round),
		DeskNo:     int32(gs.DeskNo),
		BetTime:    int32(self.GameTable.GetBetSeconds()),
		Set:        int32(gs.Set),
		IsAnswered: int32(gs.IsAnswered),
		PlayerNum:  int32(self.GameTable.GetPlayerNum()),
	}

	// 异步通知大厅服务
	self.AsyncCallNode(nodeId, rpcFun, &notifyMsg, func(ret *rpc.Int, err error) {
		if err != nil {
			return
		}
	})

	//fmt.Printf("------>游戏状态---->大厅------>data=%v\n", notifyMsg)
}

// NotifyLobbyGameAnswer 通知大厅游戏开奖消息
func (self *FtanService) NotifyLobbyGameAnswer(asn *msg_struct.FtanOpenAnswer) {
	// 构造通知消息
	asnMsg := msg.NotifyLobbyGameAnswer{
		DeskNo:   int32(asn.DeskNo),
		Set:      int32(asn.Set),
		Round:    int32(asn.Round),
		GameId:   self.GameTable.GetGameId(),
		GameType: int32(msg.DataType_FTAN),
		Winner:   int32(asn.Winner),
	}

	// 获取大厅服务节点
	nodeId := util.GetGameServiceName(util.LobbyService)
	if nodeId == "" {
		return
	}

	rpcFun := util.FormatRpcMethod(util.LobbyService, "RPCOpenGameAnswer")
	if rpcFun == "" {
		return
	}

	// 异步通知大厅游戏开答
	err := self.AsyncCallNode(nodeId, rpcFun, &asnMsg, func(ret *rpc.Int, err error) {
		if err != nil {
			log.Errorf("FtanService NotifyLobbyGameAnswer -> RPCOpenGameAnswer failed: %v, data: %v", err, asnMsg)
		} else {
			log.Infof("FtanService NotifyLobbyGameAnswer -> RPCOpenGameAnswer success: %v", asnMsg)
		}
	})

	if err != nil {
		log.Errorf("FtanService NotifyLobbyGameAnswer -> AsyncCallNode failed: %v, data: %+v", err, asnMsg)
	}

	//fmt.Printf("------>游戏开达---->大厅------>data=%v\n", asnMsg)
}

// OnServiceEvent 广播游戏状态改变事件，发给各个玩家端，同时通知大厅
func (self *FtanService) OnServiceGameStatusEvent(ev event.IEvent) {

	gs, ok := (ev.(*event.Event).Data).(msg_struct.GameStatusMsg)

	if ok == false {
		return
	}

	gs.GameId = self.GameTable.GetGameId()
	gs.DeskNo = self.GameTable.GetDeskNo()

	//广播通知客户端游戏状态改变了
	self.GameTable.CastToPlayerGameStatus(&gs)

	self.NotifyLobbyGameStatus(&gs)

	//通知玩家有没有4局没有下注的
	self.GameTable.NotifyPlayerLeaveTable(self.GameTable.GetNoBetRound())
}

// 服务监听事件回调,游戏开答事件
func (self *FtanService) OnServiceOpenAnswerEvent(ev event.IEvent) {

	asn, ok := (ev.(*event.Event).Data).(msg_struct.FtanOpenAnswer)

	if ok == false {
		return
	}

	//广播通知客户端开答内容
	self.GameTable.CastToPlayerOpenAnswer(&asn)

	self.NotifyLobbyGameAnswer(&asn)

	//fmt.Printf("------>游戏开达---->玩家/大厅------>data=%v\n", asn)

}

// OnServiceGameHistoryAnswersToLobbyEvent 通知发送历史开答表给大厅
func (self *FtanService) OnServiceGameHistoryAnswersToLobbyEvent(ev event.IEvent) {
	answers, ok := (ev.(*event.Event).Data).(msg_struct.FtanAnswersPeriodic)
	if !ok {
		return
	}

	// 构造 WinnersFtan
	answerArray := make([]uint32, 0)
	//for i := 0; i < len(answers.Winners); i++ {
	//	answerArray = append(answerArray, uint32(answers.Winners[i]))
	//}

	end := len(answers.Winners)
	start := end - msg_struct.HistoryAnswersPerioSize //长度定为20

	if start < 0 {
		start = 0
	}
	for i := start; i < end; i++ {
		answerArray = append(answerArray, uint32(answers.Winners[i]))
	}

	aans := msg.WinnersFtan{
		Totalnum: uint32(len(answers.Winners)),
		Lastnum:  uint32(len(answerArray)),
	}
	aans.WinnerItems = append(aans.WinnerItems, answerArray...)

	// 构造 STOC_GameHistoryAnswers
	perioMsg := msg.NotifyLobbyHistoryAnswers{
		GameId:     self.GameTable.GetGameId(),
		DeskNo:     int32(answers.DeskNo),
		Set:        int32(answers.Set),
		Round:      int32(answers.Round),
		GameStatus: int32(answers.GameStatus),
		DataType:   int32(msg.DataType_FTAN),
		Answers:    util.ProtoToAny(&aans),
	}

	// 获取大厅服务节点
	nodeId := util.GetGameServiceName(util.LobbyService)
	if nodeId == "" {
		return
	}

	rpcFun := util.FormatRpcMethod(util.LobbyService, "RPCUpdateGameAnswersPerio")
	if rpcFun == "" {
		return
	}

	// 异步调用大厅服务
	err := self.AsyncCallNode(nodeId, rpcFun, &perioMsg, func(ret *rpc.Int, err error) {
		if err != nil {
			log.Errorf("RPCUpdateGameAnswersPerio failed: %v, data: %v", err, perioMsg)
		} else {
			//log.Infof("RPCUpdateGameAnswersPerio success: %v", perioMsg)

			log.Infof("RPCUpdateGameAnswersPerio success: ")
		}
	})

	if err != nil {
		log.Errorf("AsyncCallNode failed: %v, data: %v", err, perioMsg)
	}

	//fmt.Printf("------>游戏历史开达---->大厅------>data=%v\n", perioMsg.Answers)
}

// 通知控端历史开答列表更新了
func (self *FtanService) OnServiceGameHistoryAnswersEvent(ev event.IEvent) {

	answers, ok := (ev.(*event.Event).Data).(msg_struct.FtanAnswersPeriodic)

	if ok == false {
		return
	}

	//先通知玩家历史开答变了
	self.GameTable.CastToPlayerHistoryAnswers(&answers)

	//需要通知控端历史开答内容变化了
	self.adminCtrl.BroadCastMsg(msg_struct.GameFtanAnswerPeriodicProtocol, &answers)

}

// 游戏投注事件，需要区分是游戏内的压注，还是游戏外压注(多台投注用)
func (self *FtanService) OnServiceTableBetEvent(ev event.IEvent) {

	betMsg, ok := (ev.(*event.Event).Data).(*rpc.BetToGameRPC)

	totalbet := 0

	if ok == false {

		//fmt.Printf("FtanService OnServiceTableBetEvent  Failed.... \n  ")
		log.Errorf("FtanService-->OnServiceTableBetEvent failed")

		return
	}

	//	fmt.Printf("FtanService OnServiceTableBetEvent ....%v \n ", betMsg)

	player := self.PlayersImp.GetPlayer(betMsg.UserId)

	//若没有找到
	if player == nil {

		log.Errorf("FtanService OnServiceTableBetEvent  内存中无法找到玩家userId:%s", betMsg.UserId)

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

		betbar.Currency = player.GetCurrency()

		betbar.CreateOrderId(msg_struct.GameTypeFtan)

		tempbet, ret := self.GameTable.PlayerBet(&betbar)

		totalbet = tempbet

		if ret == 0 {

			betResult.Success = 0

			//fmt.Printf("*************FtanService OnServiceTableBetEvent ...……%v Success  !!!!! \n ", *betMsg)

		} else {
			betResult.Success = 1
			betResult.ErrCode = int32(ret)

			log.Infof("玩家压ftangame游戏: userId:%s  betAreaId: %d , betMoney:%d Digits:%d 不成功！", betbar.UserId, betbar.BetAreaId, betbar.BetPoint, betbar.Digits)
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

// 处理游戏撤销投注事件
func (self *FtanService) OnServiceTableCancelBetEvent(ev event.IEvent) {

	betMsg, ok := (ev.(*event.Event).Data).(*rpc.BetToGameRPC)

	if !ok {

		log.Error("OnServiceTableCancelBetEvent Failed: Invalid event data")

		return
	}
	player := self.PlayersImp.GetPlayer(betMsg.UserId)

	var betRet msg.STOC_CancelBetGame_Result

	if player == nil {

		log.Warnf("FtanService  gameId:%s  OnServiceTableCancelBetEvent 内存找不到玩家：%s ", self.GameTable.GetGameId(), betMsg.UserId)

		//处理不成功代理
		betRet.Success = 1
		betRet.GameId = betMsg.GameId
		betRet.BetMode = int32(betMsg.BetMode)
		betRet.DeskNo = int32(self.GameTable.GetDeskNo())
		betRet.Balance = int32(player.GetCredit())
		betRet.Digits = int32(player.GetDigits())

		betRet.OrderKey = uint32(betMsg.OrderKey)
		betRet.BetAreadId = betMsg.BetAreadId
		betRet.BetPoint = betMsg.BetPoint
		betRet.TotalBet = 0
		betRet.ErrCode = int32(msg.ErrCode_BetErr_ServerPlayerIdErr)

		player.SendMsg(msg.MsgType_GameTable_CancelBetResult, &betRet)

		return

	}

	orderKey := int64(0)
	betAreaId := int32(0)
	betPoint := uint32(0)
	totalBet := int32(0)
	success := int32(1)
	errCode := int32(msg.ErrCode_BetErr_ServerPlayerIdErr)

	if player.Id == betMsg.UserId {

		betNode, tempBet, ret := self.GameTable.PlayerCancelBet(betMsg.UserId)

		if ret == 0 {
			success = 0
			errCode = 0
			orderKey = betNode.OrderKey
			betAreaId = int32(betNode.BetAreaId)
			betPoint = betNode.BetPoint
			totalBet = int32(tempBet)

			//log.Infof("Bet cancellation successful: %v", *betMsg)

		} else {
			errCode = int32(ret)
			//log.Errorf("Bet cancellation failed: UserId=%d, OrderKey=%s, Ret=%d", betMsg.UserId, betNode.OrderKey, ret)
		}
	}

	betResult := msg.STOC_CancelBetGame_Result{
		Success:    success,
		ErrCode:    errCode,
		OrderKey:   uint32(orderKey),
		BetAreadId: betAreaId,
		BetPoint:   betPoint,
		TotalBet:   totalBet,
		GameId:     betMsg.GameId,
		BetMode:    betMsg.BetMode,
		DeskNo:     int32(self.GameTable.GetDeskNo()),
		Balance:    int32(player.GetCredit()),
		Digits:     int32(player.GetDigits()),
	}

	player.SendMsg(msg.MsgType_GameTable_CancelBetResult, &betResult)

	//	fmt.Printf("------>撤销投注---->玩家------>data=%v\n", betResult)
}

// 打赏小费事件
func (self *FtanService) OnServiceTipEvent(ev event.IEvent) {

	tipMsg, ok := (ev.(*event.Event).Data).(*rpc.TipToGameRPC)

	if ok == false {

		//	fmt.Printf("FtanService OnServiceTableBetEvent  Failed.... \n  ")
		log.Error(" FtanService OnServiceTableBetEvent  Failed....")

		return
	}

	log.Infof("gameId:%s 玩家打赏 userId:%s   tipMoney:%d Digits:%d  开始处理！",
		self.GameTable.GetGameId(), tipMsg.UserId, tipMsg.TipPoint, tipMsg.Digits)

	//fmt.Printf("FtanService OnServiceTableBetEvent ....%v \n ", tipMsg)

	player := self.PlayersImp.GetPlayer(tipMsg.UserId)

	//若没有找到
	if player == nil {
		//需要重新生成一个玩家
		//需要重新生成一个玩家,去大厅服读取生成
		log.Errorf("FtanService-->OnServiceTipEvent 内存中找不到玩家:%s", tipMsg.UserId)

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

		log.Errorf("FtanService-->OnServiceTipEvent 打赏金额不成功 userId:%s  error:%d  ", tipMsg.UserId, err)

		return
	}
	tipResult.Success = 0

	player.SendMsg(msg.MsgType_GameTable_TipMoneyResult, &tipResult)

	//再发送一次玩家余额消息
	self.GameTable.SendToPlayerCardInfo(player)

}

func (self *FtanService) SaveAnswerToDB(asn *msg_struct.FtanOpenAnswer) bool {

	//开答结果存数据库
	var dbTable models.SGAnswerFtan

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
func (self *FtanService) OnServiceAnswerSettleEvent(ev event.IEvent) {

	answer, ok := (ev.(*event.Event).Data).(*msg_struct.FtanOpenAnswer)

	if ok == false {
		//出错处理
	}

	//存储开答
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

	//fmt.Printf("------->游戏结算-------->玩家------->data=%v\n ", *asn)

}

// 通知更新玩家在线人数
func (self *FtanService) OnServiceNotifyPlayerNumEvent(ev event.IEvent) {

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
func (self *FtanService) OnServiceTalkChatEvent(ev event.IEvent) {

	talkMsg, ok := (ev.(*event.Event).Data).(*rpc.GameTalkChat)

	if ok == false {

		log.Error(" FtanService OnServiceTalkChatEvent  Failed....")
		return
	}
	self.TalkChatBroad(talkMsg.UserId, talkMsg.Message)
}

// 广播此聊天记录
func (self *FtanService) TalkChatBroad(userId string, message string) {

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
