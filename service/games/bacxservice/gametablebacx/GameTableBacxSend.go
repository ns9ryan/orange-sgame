package gameTableBacx

import (
	"errors"
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	betOdds "videogame/common/betodds"
	"videogame/common/msg_struct"
	"videogame/common/proto/msg"
	"videogame/common/util"
	betAreadUtil "videogame/service/betarea"
	"videogame/service/games/player"
)

// SendToPlayerGameSetting 发送游戏配置给玩家
func (gm *GameTableBacx) SendToPlayerGameSetting(player *player.Player) error {

	var gameSetting msg.STOC_GameSetting

	var betAreas []*msg.BetArea

	betAreas = make([]*msg.BetArea, 0)

	digits := float32(player.GetDigits())

	//从数据库读取高低注，若读不成功，刚自已初始化
	betAreasTemp, err := player.GetBetLimit(player.BetLimitTemplateId, gm.GetGameId())

	if err == nil {

		for _, item := range betAreasTemp {

			//赔率内部处理了
			areaitem := betAreadUtil.GetBetArea(int(item.AreaId), int(item.LowBet), int(item.HighBet))
			betAreas = append(betAreas, areaitem)
		}

	} else {

		digits1 := int(player.GetDigits())

		//闲家区域
		playerArea := betAreadUtil.GetBetArea(betOdds.Bet_Bacx_Player, int(0.1*digits), 100*digits1)
		betAreas = append(betAreas, playerArea)
		//庄家区域
		bankerArea := betAreadUtil.GetBetArea(betOdds.Bet_Bacx_Banker, int(0.1*digits), 100*digits1)
		betAreas = append(betAreas, bankerArea)
		//和区域
		tieArea := betAreadUtil.GetBetArea(betOdds.Bet_Bacx_Tie, int(0.1*digits), 100*digits1)
		betAreas = append(betAreas, tieArea)
		//闲对子
		ppairArea := betAreadUtil.GetBetArea(betOdds.Bet_Bacx_PPair, int(0.1*digits), 100*digits1)
		betAreas = append(betAreas, ppairArea)
		//庄对子
		bpairArea := betAreadUtil.GetBetArea(betOdds.Bet_Bacx_BPair, int(0.1*digits), 100*digits1)
		betAreas = append(betAreas, bpairArea)
		//免佣的庄
		bankerExtArea := betAreadUtil.GetBetArea(betOdds.Bet_Bacx_BankerNocomm, int(0.1*digits), 100*digits1)
		betAreas = append(betAreas, bankerExtArea)
		//超级6点
		superSixArea := betAreadUtil.GetBetArea(betOdds.Bet_Bacx_SuperSix, int(0.1*digits), 1000*digits1)
		betAreas = append(betAreas, superSixArea)
		//任意对子
		anyPairArea := betAreadUtil.GetBetArea(betOdds.Bet_Bacx_AnyPair, int(0.1*digits), 1000*digits1)
		betAreas = append(betAreas, anyPairArea)
		//完美对子
		perfectArea := betAreadUtil.GetBetArea(betOdds.Bet_Bacx_PerfectPair, int(0.1*digits), 1000*digits1)
		betAreas = append(betAreas, perfectArea)
		//大
		bigArea := betAreadUtil.GetBetArea(betOdds.Bet_Bacx_Big, int(0.1*digits), 1000*digits1)
		betAreas = append(betAreas, bigArea)
		//小
		smallArea := betAreadUtil.GetBetArea(betOdds.Bet_Bacx_Small, int(0.1*digits), 1000*digits1)
		betAreas = append(betAreas, smallArea)
		//闲龙宝
		playerBonusArea := betAreadUtil.GetBetArea(betOdds.Bet_Bacx_PlayerBonus, int(0.1*digits), 100*digits1)
		betAreas = append(betAreas, playerBonusArea)
		//庄龙宝
		bankerBonusArea := betAreadUtil.GetBetArea(betOdds.Bet_Bacx_BankerBonus, int(0.1*digits), 100*digits1)
		betAreas = append(betAreas, bankerBonusArea)
	}

	betSetting := msg.BetSetting{}
	betSetting.Digits = int32(digits)
	betSetting.BetAreaNum = int32(len(betAreas))
	betSetting.BetArea = append(betSetting.BetArea, betAreas...)

	//下面这些以后改成从数据库读取，每个玩家是不一样的
	gameSetting.GameId = gm.GetGameId()
	gameSetting.DeskNo = int32(gm.GameInfo.DeskNo)
	gameSetting.Currency = player.GetCurrency()
	gameSetting.DealerName = gm.GameInfo.Dealer.DealerId //荷管的名字

	gameSetting.BetNum = 13 //百家乐有13个投注区域

	gameSetting.DataType = int32(msg.DataType_BACX)
	gameSetting.TableSetting = util.ProtoToAny(&betSetting)

	//读取抽码
	chiplist := player.GetChipsShow(player.GetDigits())

	gameSetting.ChipNum = int32(len(chiplist))
	gameSetting.ChipDigits = int32(player.GetDigits())

	gameSetting.VideoUrl = gm.GameInfo.VideoURL

	//需要从数据库读取
	gameSetting.ChipPoints = append(gameSetting.ChipPoints, chiplist...)

	tipData, ok := player.GetTipData()

	if ok == true {

		gameSetting.TipMin = int32(tipData.MinTip)
		gameSetting.TipMax = int32(tipData.MaxTip)
		gameSetting.Digits = int32(tipData.TipDigits)

		gameSetting.TipUseNum = int32(len(tipData.TipItems))
		gameSetting.TipItems = append(gameSetting.TipItems, tipData.TipItems...)

	} else {

		//设置打赏最小，最大金额
		gameSetting.TipMin = int32(player.MinTip) //需要修改
		gameSetting.TipMax = int32(player.MaxTip)
		gameSetting.Digits = int32(player.GetDigits())

		tipItems := []int32{1, 2, 5, 10, 20, 50, 100, 200}

		gameSetting.TipUseNum = int32(len(tipItems))
		gameSetting.TipItems = append(gameSetting.TipItems, tipItems...)
	}

	player.SendMsg(msg.MsgType_GameTable_Setting, &gameSetting)

	fmt.Printf("*******Send GameSetting ToPlayer ..  \n")

	return nil
}

// 发送高低注信息给玩家
func (gm *GameTableBacx) SendToPlayerBetLimit(player *player.Player) error {

	var limitMsg msg.STOC_GameBetLimit

	limitMsg.GameId = gm.GetGameId()
	limitMsg.DeskNo = int32(gm.GetDeskNo())

	limitMsg.LowBet = int32(player.SettingData.BetSet.LowBet)
	limitMsg.HighBet = int32(player.SettingData.BetSet.HighBet)
	limitMsg.Digits = int32(player.SettingData.BetSet.Digits)

	//发送高低注投注信息
	player.SendMsg(msg.MsgType_GameTable_BetLimit, &limitMsg)

	return nil
}

func (gm *GameTableBacx) SendPlayersCardInfo() {

	for _, pp := range gm.MapPlayer {
		//从数据库读取
		gm.SendToPlayerCardInfo(pp)
	}
}

// 若玩家上线时发送之前投注未开答的投注额
/*func (gm *GameTableBacx) SendToPlayerBetted(player *player.Player) error {

	ppbet, ok := gm.MapPlayerBet[player.Id]

	if ok == false {
		return nil
	}
	if ppbet.TotalBet() == 0 {
		return nil
	}

	bets := ppbet.GetValidBets()

	var betmsg msg.STOC_GamePlayerBetted

	validbet := make([]*msg.BetAmout, 0)

	for key, bet := range bets {

		var betItem msg.BetAmout

		betItem.BetAreaId = int32(key)
		betItem.BetPoint = int32(bet)

		validbet = append(validbet, &betItem)
	}

	betmsg.BetItems = append(betmsg.BetItems, validbet...)

	betmsg.GameId = gm.GetGameId()
	betmsg.Deskno = int32(gm.GameInfo.DeskNo)
	betmsg.TotalBet = uint32(uint(ppbet.TotalBet()))
	betmsg.Digits = int32(player.GetDigits())

	//发送投注
	player.SendMsg(msg.MsgType_GameTable_PlayerBetted, &betmsg)

	fmt.Printf("发送给玩家之前的投注额 userid:%s betAmout: %d \n ", player.Id, betmsg.TotalBet/uint32(betmsg.Digits))

	return nil
}*/

func (gm *GameTableBacx) SendToPlayerBetted(player *player.Player) error {

	ppbet, ok := gm.MapPlayerBet[player.Id]

	if ok == false {
		return nil
	}
	if ppbet.TotalBet() == 0 {
		return nil
	}

	bets := ppbet.GetBetNode()

	var betmsg msg.STOC_GamePlayerBetted

	validbet := make([]*msg.BetAmout, 0)

	for _, bet := range bets {

		var betItem msg.BetAmout

		betItem.BetAreaId = int32(bet.BetAreaId)
		betItem.BetPoint = int32(bet.BetPoint)
		betItem.OrderKey = uint32(bet.OrderKey)

		validbet = append(validbet, &betItem)
	}

	betmsg.BetItems = append(betmsg.BetItems, validbet...)

	betmsg.GameId = gm.GetGameId()
	betmsg.Deskno = int32(gm.GameInfo.DeskNo)
	betmsg.TotalBet = uint32(uint(ppbet.TotalBet()))
	betmsg.Digits = int32(player.GetDigits())

	//发送投注
	player.SendMsg(msg.MsgType_GameTable_PlayerBetted, &betmsg)

	log.Infof("发送给玩家之前的投注额gameId:%s,  userid:%s betAmout: %d \n ", gm.GetGameId(), player.Id, betmsg.TotalBet/uint32(betmsg.Digits))

	return nil
}

// 发送玩家胜额
func (gm *GameTableBacx) SendToPlayerBetWin(ppbet *PlayerBet) error {

	var ppwin msg.STOC_GamePlayerWin

	ppwin.Deskno = int32(gm.GetDeskNo())
	ppwin.GameId = gm.GetGameId()
	ppwin.UserId = ppbet.Player.Id
	ppwin.WinPoint = int32(ppbet.CurWin)
	ppwin.Digits = int32(ppbet.Digits)

	ppbet.Player.SendMsg(msg.MsgType_GameTable_PlayerWin, &ppwin)

	fmt.Printf("***发送给玩家胜额 userid:%s , win:%d , digit:%d  \n", ppwin.UserId, ppwin.WinPoint, ppwin.Digits)

	return nil
}

// 通知 玩家 胜额
func (gm *GameTableBacx) SendPlayersWin() {

	bwin := false

	for _, ppbet := range gm.MapPlayerBet {

		if !ppbet.IsWin() {
			continue
		}
		gm.SendToPlayerBetWin(ppbet)

		bwin = true
	}

	if bwin == false {
		fmt.Printf("没有人胜！！！！\n")
	}

}

// 发送历史开答给玩家
func (gm *GameTableBacx) SendToPlayerHistoryAnswers(player *player.Player) error {

	var perioMsg msg.STOC_GameHistoryAnswers

	var answers msg_struct.GameBacXAnswersPeriodic

	//先读取历史开答列表
	gm.GetAnswerPeriodic(&answers)

	perioMsg.GameId = gm.GetGameId()
	perioMsg.DeskNo = int32(answers.DeskNo)
	perioMsg.Set = int32(answers.Set)
	perioMsg.Round = int32(answers.Round)
	perioMsg.GameStatus = int32(answers.GameStatus)
	perioMsg.DataType = int32(msg.DataType_BACX)

	var aans msg.WinnersBACX

	for i := 0; i < len(answers.Winners); i++ {

		var winner msg.WinnerBACX

		winner.Winner = uint32(answers.Winners[i].Winner)
		winner.PPair = uint32(answers.Winners[i].PPair)
		winner.BPair = uint32(answers.Winners[i].BPair)

		aans.WinnerItems = append(aans.WinnerItems, &winner)
	}

	perioMsg.Answers = util.ProtoToAny(&aans)

	sendnum := player.SendToPlayer(player.Id, msg.MsgType_GameTable_HistoryAnswers, &perioMsg)

	if sendnum != 1 {
		return errors.New("send to player error")
	}

	return nil
}

// 发送玩家SuperSix的开关状态
func (gm *GameTableBacx) SendToPlayerSuperSix(player *player.Player) error {

	//返回给玩家端修改了这个supersix 状态
	var retmsg msg.STOC_GameSuperSixStatus

	retmsg.GameId = gm.GetGameId()
	retmsg.DeskNo = int32(gm.GetDeskNo())
	retmsg.SuperSix = int32(player.GetSuperSix())
	retmsg.EnableOpt = player.IsSuperSixOpt()

	retmsg.SuperVisable = player.GetSuperVisible()

	player.SendMsg(msg.MsgType_GameTable_SuperSixStatus, &retmsg)

	fmt.Printf("发送给玩家 它的supersix 状态: supersix :%d   ", retmsg.SuperSix)

	return nil
}

// 广播历史开答列表
func (gm *GameTableBacx) CastToPlayerHistoryAnswers(answers *msg_struct.GameBacXAnswersPeriodic) error {

	var perioMsg msg.STOC_GameHistoryAnswers

	var Winners msg.WinnersBACX

	perioMsg.GameId = gm.GetGameId()
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

	clientIds := gm.GetAllClientId()

	//发送客户端游戏历史开答
	sendnum := gm.Gs.GetMsgSender().CastToClient(clientIds, msg.MsgType_GameTable_HistoryAnswers, &perioMsg)

	if sendnum != len(clientIds) {
		return errors.New("error send ")
	}

	return nil
}

// 广播游戏开奖
func (gm *GameTableBacx) CastToPlayerOpenAnswer(asn *msg_struct.GameOpenAnswerBacx) error {

	var asnMs msg.STOC_GameAnswer

	var cardBacx msg.CardBACX
	var answerBacx msg.AnswerBACX

	cardBacx.PCardNum = 0
	cardBacx.BCardNum = 0

	for i := 0; i < 3; i++ {

		//检测填充启用的闲家牌
		if asn.PlayerCard[i].Used == 1 {

			pcard := msg.Card{}
			pcard.Number = uint32(asn.PlayerCard[i].Number)
			pcard.Color = uint32(asn.PlayerCard[i].Color)

			cardBacx.PCardNum++
			cardBacx.PlayerCards = append(cardBacx.PlayerCards, &pcard)
		}

		//检测填充启用的庄家牌
		if asn.BankerCard[i].Used == 1 {
			bcard := msg.Card{}

			bcard.Number = uint32(asn.BankerCard[i].Number)
			bcard.Color = uint32(asn.BankerCard[i].Color)

			cardBacx.BCardNum++
			cardBacx.BankerCards = append(cardBacx.BankerCards, &bcard)
		}
	}

	answerBacx.Cards = &cardBacx
	answerBacx.Winner = uint32(asn.Winner.Winner)
	answerBacx.IsPPair = uint32(asn.Winner.PPair)
	answerBacx.IsBPair = uint32(asn.Winner.BPair)

	//付加的对子检测
	answerBacx.IsAnyPair = uint32(util.ConverBoolToInt(asn.BacxAnswerCard.IsAnyPair()))
	answerBacx.IsBig = uint32(util.ConverBoolToInt(asn.BacxAnswerCard.IsBig()))
	answerBacx.IsPerfectPair = uint32(util.ConverBoolToInt(asn.BacxAnswerCard.IsPerfectPair()))

	ok := asn.BacxAnswerCard.IsBankerBonus()
	answerBacx.IsBankerVBonus = uint32(util.ConverBoolToInt(ok))

	ok = asn.BacxAnswerCard.IsPlayerBonus()
	answerBacx.IsPlayerBonus = uint32(util.ConverBoolToInt(ok))

	asnMs.DeskNo = int32(asn.DeskNo)
	asnMs.Set = int32(asn.Set)
	asnMs.Round = int32(asn.Round)
	asnMs.GameId = gm.GetGameId()
	asnMs.DataType = int32(msg.DataType_BACX)

	asnMs.Answer = util.ProtoToAny(&answerBacx)

	clientIds := gm.GetAllClientId()

	//通知客户端游戏状态改变了
	sendnum := gm.Gs.GetMsgSender().CastToClient(clientIds, msg.MsgType_GameTable_Answer, &asnMs)

	if sendnum != len(clientIds) {
		return errors.New("error send ")
	}

	return nil
}

// 广播游戏状态给玩家
func (gm *GameTableBacx) CastToPlayerGameStatus(gsStatus *msg_struct.GameStatusMsg) error {

	var gameGs msg.STOC_GameStatus

	gameGs.GameId = gm.GetGameId()
	gameGs.DeskNo = int32(gsStatus.DeskNo)
	gameGs.Set = int32(gsStatus.Set)
	gameGs.Round = int32(gsStatus.Round)
	gameGs.GameSatus = int32(gsStatus.GameStatus)
	gameGs.IsAnswered = int32(gsStatus.IsAnswered)
	gameGs.BetTime = int32(gm.GetBetSeconds())

	//fmt.Printf("BacxService  CastToPlayerGameStatus  %d \n", gameGs.BetTime)
	//log.Infof("**************BacxService  CastToPlayerGameStatus   gamestauts:%d, betTime : %d \n", gameGs.GameSatus, gameGs.BetTime)

	clientIds := gm.GetAllClientId()

	sendnum := gm.Gs.GetMsgSender().CastToClient(clientIds, msg.MsgType_GameTable_Status, &gameGs)

	if sendnum != len(clientIds) {
		return errors.New("error send ")
	}

	return nil
}
