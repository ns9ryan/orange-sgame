package botservice

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"time"
	"videogame/common/proto/msg"
	"videogame/common/util"
)

func init() {
	mapRegisterMsg = make(map[msg.MsgType]*MsgEvent)
	//regMsg(msg.MsgType_LoadFinish, &msg.MsgLoadFinish{}, loadFinish)
	regMsg(msg.MsgType_Pong, &msg.MsgPong{}, pong)

	//玩家请求

	regMsg(msg.MsgType_LoginResult, &msg.LoginRes{}, loginRes)

	regMsg(msg.MsgType_JoinGameResult, &msg.JoinToGameResult{}, enterGameRes)

	regMsg(msg.MsgType_ExitGameResult, &msg.LeaveFromGameResult{}, leaveGameRes)
	regMsg(msg.MsgType_GameTable_Status, &msg.STOC_GameStatus{}, updateGameStatus)
	regMsg(msg.MsgType_GameTable_Answer, &msg.STOC_GameAnswer{}, openGameAnswer)

	regMsg(msg.MsgType_GameTable_HistoryAnswers, &msg.STOC_GameHistoryAnswers{}, historyGameAnswers)

	regMsg(msg.MsgType_PlayerInfo_Card, &msg.STOC_PlayerInfo{}, playerInfoDTX)
	regMsg(msg.MsgType_GameTable_Setting, &msg.STOC_GameSetting{}, gameConfig)
	regMsg(msg.MsgType_Lobby_GameGroup, &msg.STOC_LobbyGameGroups{}, lobbyGameGroups)
	regMsg(msg.MsgType_Lobby_GameRoom, &msg.STOC_LobbyGameRooms{}, lobbyGameRooms)
	regMsg(msg.MsgType_GameInfo_PlayNum, &msg.STOC_PlayerNum{}, playerInfoNum)
	regMsg(msg.MsgType_GameTable_BetResult, &msg.STOC_BetGame_Result{}, betInfoResult)

	regMsg(msg.MsgType_Lobby_GameStatus, &msg.NotifyLobbyGameStatus{}, lobbyGameStatus)
	regMsg(msg.MsgType_Lobby_GameAnswer, &msg.NotifyLobbyGameAnswer{}, lobbyGameAnswer)
	regMsg(msg.MsgType_Lobby_GameHistoryAnswers, &msg.NotifyLobbyHistoryAnswers{}, lobbyGameHistoryAnswers)

	regMsg(msg.MsgType_GameTable_SuperSixStatus, &msg.STOC_GameSuperSixStatus{}, supersixStatus)
	regMsg(msg.MsgType_GameTable_TipMoney, &msg.STOC_TipMoneyResult{}, tipMoneyResult)

	regMsg(msg.MsgType_GameTable_TalkChatShow, &msg.STOC_TalkChat{}, talkChatMsgResult)

}

func startBet(bot *Bot, msgProto proto.Message) {
	_ = bot
	//	_ = msgProto.(*msg.StartBetToClient)
	fmt.Println("【玩家】可以下注\n")
	//fmt.Printf("【玩家】开始下注 返回=%v\n", util.ProtoMessageToMap(retMsg))

	if bot.status == finish { //玩家进去房间以后，开始调用下注
		//bot.userBet()
	}
}

func enterGameRes(bot *Bot, msgProto proto.Message) {

	_ = bot

	retMsg := msgProto.(*msg.JoinToGameResult)

	fmt.Printf("bot ********JoinGame gameId:%s ,deskNo:%s, success:%d \n", retMsg.GameId, retMsg.DeskNo, retMsg.Success)

	var entergame msg.EnterToGame

	entergame.GameId = bot.gameId
	entergame.UserId = playerId
	entergame.DeskNo = retMsg.DeskNo

	bot.SendMsg(msg.MsgType_EnterGame, &entergame)

}

func leaveGameRes(bot *Bot, msgProto proto.Message) {

	_ = bot

	retMsg := msgProto.(*msg.LeaveFromGameResult)

	fmt.Printf("bot********leaveGame gameId:%s ,deskNo:%s, success:%d \n", retMsg.GameId, retMsg.DeskNo, retMsg.Success)

}

func updateGameStatus(bot *Bot, msgProto proto.Message) {

	_ = bot

	retMsg := msgProto.(*msg.STOC_GameStatus)

	fmt.Printf("bot##############update gamestatus  gameId:%s ,deskNo:%d,  set:%d, round:%d  gamestatus:%d \n",
		retMsg.GameId, retMsg.DeskNo, retMsg.Set, retMsg.Round, retMsg.GameSatus)

	if retMsg.GameSatus == 2 {

		time.AfterFunc(2*time.Second, bot.userBet)

		bot.SendSuperSixSet()

	}

}

func openGameAnswer(bot *Bot, msgProto proto.Message) {

	_ = bot

	retMsg := msgProto.(*msg.STOC_GameAnswer)

	if retMsg.DataType == int32(msg.DataType_DTX) {
		var cardDtx msg.AnswerDTX

		retMsg.Answer.UnmarshalTo(&cardDtx)

		fmt.Printf("bot=================DTX  Answsergame gameId:%s ,deskNo:%d,  set:%d, round:%d winner:%d dcard:%v  tcard:%v\n", retMsg.GameId, retMsg.DeskNo, retMsg.Set, retMsg.Round, cardDtx.Winner, cardDtx.Cards.DragonCard, cardDtx.Cards.TigerCard)

	} else if retMsg.DataType == int32(msg.DataType_BACX) {

		var answerBacx msg.AnswerBACX

		retMsg.Answer.UnmarshalTo(&answerBacx)

		fmt.Printf("bot==================BACX AnswerGame: winner:%d, ppair:%d bpair:%d poker:%v \n", answerBacx.Winner, answerBacx.IsPPair, answerBacx.IsBPair, answerBacx.Cards)

		fmt.Printf("bot==================BACX AnswerGame:big:%d, supersix:%d, anypair:%d, perfectpair:%d,   \n", answerBacx.IsBig, answerBacx.IsSuperSix,
			answerBacx.IsAnyPair, answerBacx.IsPerfectPair)

	} else if retMsg.DataType == int32(msg.DataType_SICBO) {

		var answerSicbo msg.AnswerSicbo

		retMsg.Answer.UnmarshalTo(&answerSicbo)

		fmt.Printf("bot=================Sicbo  Answsergame gameId:%s ,deskNo:%d,  set:%d, round:%d winner:%d SicboA:%d, SicboB:%d, SicboC:%D   v\n",
			retMsg.GameId, retMsg.DeskNo, retMsg.Set, retMsg.Round, answerSicbo.Winner, answerSicbo.SicboA, answerSicbo.SicboB, answerSicbo.SicboC)
	}

}

func historyGameAnswers(bot *Bot, msgProto proto.Message) {

	_ = bot

	retMsg := msgProto.(*msg.STOC_GameHistoryAnswers)

	if retMsg.DataType == int32(msg.DataType_DTX) {

		var winnersDtx msg.WinnersDTX

		retMsg.Answers.UnmarshalTo(&winnersDtx)

		fmt.Printf("bot========historyGameAnswers gameId:%s ,deskNo:%d,  set:%d, round:%d answers:%v\n", retMsg.GameId, retMsg.DeskNo, retMsg.Set, retMsg.Round, winnersDtx.WinnerItems)

	} else if retMsg.DataType == int32(msg.DataType_BACX) {

		var WinnerBacx msg.WinnersBACX

		retMsg.Answers.UnmarshalTo(&WinnerBacx)

		fmt.Printf("bot========historyGameAnswers gameId:%s ,deskNo:%d,  set:%d, round:%d answers:%v\n", retMsg.GameId, retMsg.DeskNo, retMsg.Set, retMsg.Round, WinnerBacx.WinnerItems)

	} else if retMsg.DataType == int32(msg.DataType_SICBO) {

		var Winner msg.WinnersSicbo

		retMsg.Answers.UnmarshalTo(&Winner)

		fmt.Printf("bot========historyGameAnswers gameId:%s ,deskNo:%d,  set:%d, round:%d answers:%v\n", retMsg.GameId, retMsg.DeskNo, retMsg.Set, retMsg.Round, Winner.WinnerItems)

	}

}

func playerInfoDTX(bot *Bot, msgProto proto.Message) {

	_ = bot

	retMsg := msgProto.(*msg.STOC_PlayerInfo)

	fmt.Printf("bot========playerInfoDTX playinf:%v \n", retMsg)
}

func gameConfig(bot *Bot, msgProto proto.Message) {

	_ = bot

	retMsg := msgProto.(*msg.STOC_GameSetting)

	fmt.Printf("bot========gameConfig gameSettin:%v  \n", retMsg)

	if retMsg.DataType == int32(msg.DataType_DTX) {

		var setting msg.BetSetting

		retMsg.TableSetting.UnmarshalTo(&setting)

		//fmt.Printf("       low:%d ,high：%d \n", setting.DragonArea.LowBet, setting.DragonArea.HighBet)

	} else if retMsg.DataType == int32(msg.DataType_BACX) {

		var setting msg.BetSetting

		retMsg.TableSetting.UnmarshalTo(&setting)

		areanum := len(setting.BetArea)

		for i := 0; i < areanum; i++ {
			fmt.Printf(" areaId:%d ,  low:%d ,high：%d \n", setting.BetArea[i].AreaId, setting.BetArea[i].LowBet, setting.BetArea[i].HighBet)
		}

	} else if retMsg.DataType == int32(msg.DataType_SICBO) {

		var setting msg.BetSetting

		retMsg.TableSetting.UnmarshalTo(&setting)

		areanum := len(setting.BetArea)

		for i := 0; i < areanum; i++ {
			fmt.Printf(" areaId[%d]:%d ,  low:%d ,high：%d \n", i, setting.BetArea[i].AreaId, setting.BetArea[i].LowBet, setting.BetArea[i].HighBet)
		}
	}
}

func lobbyGameGroups(bot *Bot, msgProto proto.Message) {

	_ = bot

	//var lobbyGroups msg.STOC_LobbyGameGroups

	retMsg := msgProto.(*msg.STOC_LobbyGameGroups)

	fmt.Printf("bot====lobbyGameGroups:%v  \n", retMsg)

	time.AfterFunc(time.Second*2, bot.work)

}

func lobbyGameRooms(bot *Bot, msgProto proto.Message) {

	_ = bot

	retMsg := msgProto.(*msg.STOC_LobbyGameRooms)

	fmt.Printf("bot====lobbyGameRooms :%v  \n", retMsg)

	for index, v := range retMsg.GameRooms {
		fmt.Printf("bot == gameRooms :idx:%d %v \n", index, v)

		fmt.Printf("      minbet :%d  maxbet:%d digits:%d \n\n", v.MinBet, v.MaxBet, v.Digits)
	}

}

func lobbyGameStatus(bot *Bot, msgProto proto.Message) {

	_ = bot

	retMsg := msgProto.(*msg.NotifyLobbyGameStatus)

	fmt.Printf("bot====lobbyGameStatus :%v  \n", retMsg)
}

func lobbyGameAnswer(bot *Bot, msgProto proto.Message) {

	_ = bot

	retMsg := msgProto.(*msg.NotifyLobbyGameAnswer)

	fmt.Printf("bot*****lobbyAnswser   gameId:%s ,deskNo:%d,  set:%d, round:%d gameType:%d  winner:%d Val1:%d  Val2:%d Val3:%d  \n",
		retMsg.GameId, retMsg.DeskNo, retMsg.Set, retMsg.Round, retMsg.GameType, retMsg.Winner, retMsg.Val1, retMsg.Val2, retMsg.Val3)

}

func lobbyGameHistoryAnswers(bot *Bot, msgProto proto.Message) {

	_ = bot

	retMsg := msgProto.(*msg.NotifyLobbyHistoryAnswers)

	switch msg.DataType(retMsg.DataType) {

	case msg.DataType_DTX:

		var cardDtx msg.WinnersDTX

		retMsg.Answers.UnmarshalTo(&cardDtx)

		fmt.Printf("bot*****DTX Answsergame gameId:%s ,deskNo:%d,  set:%d, round:%d winner:%v  ", retMsg.GameId, retMsg.DeskNo, retMsg.Set, retMsg.Round, cardDtx.WinnerItems)

	case msg.DataType_SICBO:

		var Winner msg.WinnersSicbo

		retMsg.Answers.UnmarshalTo(&Winner)

		fmt.Printf("bot*****sicbo Answsergame gameId:%s ,deskNo:%d,  set:%d, round:%d winner:%v  ", retMsg.GameId, retMsg.DeskNo, retMsg.Set, retMsg.Round, Winner.WinnerItems)

	}

	fmt.Printf("bot====lobbyGameHistoryAnswers :%v  \n", retMsg)

	time.AfterFunc(time.Second*1, bot.enterGame)

}

func playerInfoNum(bot *Bot, msgProto proto.Message) {

	_ = bot

	retMsg := msgProto.(*msg.STOC_PlayerNum)

	fmt.Printf("bot====playerInfoNum :%s playenum:%d  \n", retMsg.GameId, retMsg.OnlineNum)

}

func betInfoResult(bot *Bot, msgProto proto.Message) {

	_ = bot

	retMsg := msgProto.(*msg.STOC_BetGame_Result)

	fmt.Printf("bot））））））——————betInfoResult :%v  \n", retMsg)

}

func supersixStatus(bot *Bot, msgProto proto.Message) {

	_ = bot

	retMsg := msgProto.(*msg.STOC_GameSuperSixStatus)

	fmt.Printf("bot ****************** ——————supersixStatus :%d opt:%t \n", retMsg.SuperSix, retMsg.EnableOpt)

}

func loginRes(bot *Bot, msgProto proto.Message) {

	retMsg := msgProto.(*msg.LoginRes)

	if retMsg.Success == int32(msg.ErrCode_Ok) {

		fmt.Printf(" lobby game server ok!!!! Account:%s \n", retMsg.Msg)
	} else {
		fmt.Printf(" lobby game server failed !!!! \n")
	}

	var enterlobby msg.EnterToLobby

	enterlobby.UserId = playerId

	bot.SendMsg(msg.MsgType_EnterLobby, &enterlobby)

}

func tipMoneyResult(bot *Bot, msgProto proto.Message) {

	_ = bot

	retMsg := msgProto.(*msg.STOC_TipMoneyResult)

	fmt.Printf("bot ****************** ——————tipMoneyResult  tipMoney:%d surccess :%d \n", retMsg.TipMoney, retMsg.Success)

}

func talkChatMsgResult(bot *Bot, msgProto proto.Message) {

	_ = bot

	retMsg := msgProto.(*msg.STOC_TalkChat)

	fmt.Printf("bot ****************** ——————talkChatMsg   playerId:%s  chatMessage :%d \n", retMsg.PlayerId, retMsg.Message)

}

func loadFinish(bot *Bot, msgProto proto.Message) {
	///msgLoadFinish := msgProto.(*msg.MsgLoadFinish)
	//log.Debugf("msgLoadFinish", log.Any("msgLoadFinish", msgLoadFinish))
	bot.setLoginFinish()
}

func pong(bot *Bot, msgProto proto.Message) {

	msgPong := msgProto.(*msg.MsgPong)
	_ = msgPong

	fmt.Printf("bot   【玩家】游戏心跳 返回=%v\n", util.ProtoMessageToMap(msgPong))

	//bot.setLoginFinish()
}
