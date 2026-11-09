package gameTableRot

import (
	"errors"
	"fmt"
	"github.com/duanhf2012/origin/v2/log"

	//"github.com/duanhf2012/origin/v2/log"
	betOdds "videogame/common/betodds"
	"videogame/common/msg_struct"
	"videogame/common/proto/msg"
	"videogame/common/util"
	betAreadUtil "videogame/service/betarea"
	"videogame/service/games/msgrouter"
	"videogame/service/games/player"
)

// SendToPlayerGameSetting 发送游戏配置给玩家
func (gm *GameTableRot) SendToPlayerGameSetting(player *player.Player) error {

	var gameSetting msg.STOC_GameSetting

	var betAreas []*msg.BetArea

	betAreas = make([]*msg.BetArea, 0)

	digits := float32(player.GetDigits())
	betAreas = make([]*msg.BetArea, 0)

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

		//lowBet := int(1 * float32(digits))
		//highBet := int(1000 * float32(digits))

		//单数区域
		for i := betOdds.Bet_Rot_Single0; i <= betOdds.Bet_Rot_Single36; i++ {
			singleArea := betAreadUtil.GetBetArea(i, int(0.1*digits), 100*digits1)
			betAreas = append(betAreas, singleArea)
		}

		//双数区域
		for i := betOdds.Bet_Rot_Double1or2; i <= betOdds.Bet_Rot_Double0or3; i++ {
			doubleArea := betAreadUtil.GetBetArea(i, int(0.1*digits), 100*digits1)
			betAreas = append(betAreas, doubleArea)
		}

		//三数区域
		for i := betOdds.Bet_Rot_Three1to3; i <= betOdds.Bet_Rot_Three0or2or3; i++ {
			threeArea := betAreadUtil.GetBetArea(i, int(0.1*digits), 100*digits1)
			betAreas = append(betAreas, threeArea)
		}

		//四数区域
		for i := betOdds.Bet_Rot_Four1or2or4or5; i <= betOdds.Bet_Rot_Four0or1or2or3; i++ {
			fourArea := betAreadUtil.GetBetArea(i, int(0.1*digits), 100*digits1)
			betAreas = append(betAreas, fourArea)
		}

		//六数区域
		for i := betOdds.Bet_Rot_Six1to6; i <= betOdds.Bet_Rot_Six31to36; i++ {
			sixArea := betAreadUtil.GetBetArea(i, int(0.1*digits), 100*digits1)
			betAreas = append(betAreas, sixArea)
		}

		//十二数区域
		for i := betOdds.Bet_Rot_Twelve1rd12; i <= betOdds.Bet_Rot_Twelve2to1c; i++ {
			twelveArea := betAreadUtil.GetBetArea(i, int(0.1*digits), 100*digits1)
			betAreas = append(betAreas, twelveArea)
		}

		//十八数区域
		for i := betOdds.Bet_Rot_EighteenRed; i <= betOdds.Bet_Rot_EighteenDouble; i++ {
			eighteenArea := betAreadUtil.GetBetArea(i, int(0.1*digits), 100*digits1)
			betAreas = append(betAreas, eighteenArea)
		}

		//NeighbourTwo邻注区域，目前每个数左右两个邻居(以后可能更多邻居)，筹码要5的倍数
		for i := betOdds.Bet_Rot_NeighbourTwo0; i <= betOdds.Bet_Rot_NeighbourTwo26; i++ {
			neighbourTwoArea := betAreadUtil.GetBetArea(i, int(0.5*digits), 100*digits1)
			betAreas = append(betAreas, neighbourTwoArea)
		}

		//Tiers(轮盘下角注)区域，筹码要6的倍数
		tiersArea := betAreadUtil.GetBetArea(betOdds.Bet_Rot_Tiers, int(0.6*digits), 100*digits1)
		betAreas = append(betAreas, tiersArea)

		//Orphelins(轮盘孤注)区域，筹码要5的倍数
		orphelinsArea := betAreadUtil.GetBetArea(betOdds.Bet_Rot_Voisins, int(0.5*digits), 100*digits1)
		betAreas = append(betAreas, orphelinsArea)

		//Voisins(零旁注上角)区域，筹码要9的倍数
		voisinsArea := betAreadUtil.GetBetArea(betOdds.Bet_Rot_Voisins, int(0.9*digits), 100*digits1)
		betAreas = append(betAreas, voisinsArea)

		//Jeu0(轮盘零旁)区域，筹码要4的倍数
		jeu0Area := betAreadUtil.GetBetArea(betOdds.Bet_Rot_Jeu0, int(0.4*digits), 100*digits1)
		betAreas = append(betAreas, jeu0Area)

	}

	betSetting := msg.BetSetting{}

	betSetting.Digits = int32(digits)
	betSetting.BetAreaNum = int32(len(betAreas))
	betSetting.BetArea = append(betSetting.BetArea, betAreas...)

	//下面这些以后改成从数据库读取，每个玩家是不一样的
	gameSetting.GameId = gm.GetGameId()
	gameSetting.DeskNo = int32(gm.GameInfo.DeskNo)
	gameSetting.Currency = player.GetCurrency()
	gameSetting.DealerName = gm.GameInfo.Dealer.Name //荷管的名字

	gameSetting.BetNum = int32(len(betAreas)) //198

	gameSetting.DataType = int32(msg.DataType_ROT)
	gameSetting.TableSetting = util.ProtoToAny(&betSetting)

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

	//fmt.Printf("*******Send GameSetting ToPlayer ..  \n")

	return nil
}

// 发送高低注信息给玩家
func (gm *GameTableRot) SendToPlayerBetLimit(player *player.Player) error {

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

func (gm *GameTableRot) SendPlayersCardInfo() {

	for _, pp := range gm.MapPlayer {
		//从数据库读取
		gm.SendToPlayerCardInfo(pp)
	}
}

// 若玩家上线时发送之前投注未开答的投注额
func (gm *GameTableRot) SendToPlayerBetted(player *player.Player) error {

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
func (gm *GameTableRot) SendToPlayerBetWin(ppbet *PlayerBet) error {

	var ppwin msg.STOC_GamePlayerWin

	ppwin.Deskno = int32(gm.GetDeskNo())
	ppwin.GameId = gm.GetGameId()
	ppwin.UserId = ppbet.Player.Id
	ppwin.WinPoint = int32(ppbet.CurWin)
	ppwin.Digits = int32(ppbet.Digits)

	ppbet.Player.SendMsg(msg.MsgType_GameTable_PlayerWin, &ppwin)

	//fmt.Printf("轮盘发送给玩家胜额 userid:%s , win:%d , digit:%d  \n\n\n", ppwin.UserId, ppwin.WinPoint, ppwin.Digits)
	//log.Error("轮盘发送给玩家胜额 userid:%s , win:%d , digit:%d  \n", ppwin.UserId, ppwin.WinPoint, ppwin.Digits)

	return nil
}

// 通知 玩家 胜额
func (gm *GameTableRot) SendPlayersWin() {

	//bwin := false

	for _, ppbet := range gm.MapPlayerBet {

		if !ppbet.IsWin() {
			continue
		}

		fmt.Printf("玩家%s有下注中了，通知他...\n", ppbet.Player.Id)
		gm.SendToPlayerBetWin(ppbet)

		//bwin = true
	}

	// if bwin == false {
	//     fmt.Printf("没有人胜！！！！\n")
	// }

}

// 发送历史开答给玩家
func (gm *GameTableRot) SendToPlayerHistoryAnswers(player *player.Player) error {

	var perioMsg msg.STOC_GameHistoryAnswers

	var answers msg_struct.GameRotAnswersPeriodic

	gm.GetAnswerPeriodic(&answers)

	perioMsg.GameId = gm.GetGameId()
	perioMsg.DeskNo = int32(answers.DeskNo)
	perioMsg.Set = int32(answers.Set)
	perioMsg.Round = int32(answers.Round)
	perioMsg.GameStatus = int32(answers.GameStatus)
	perioMsg.DataType = int32(msg.DataType_ROT)

	var aans msg.WinnersROT

	start := len(answers.Winners) - ShowAnswerNum //暂定100

	if start < 0 {
		start = 0
	}

	answerarray := make([]uint32, 0)
	for i := start; i < len(answers.Winners); i++ {
		answerarray = append(answerarray, uint32(answers.Winners[i]))
	}
	aans.Totalnum = uint32(len(answers.Winners))
	aans.WinnerItems = append(aans.WinnerItems, answerarray...)
	aans.Lastnum = uint32(len(aans.WinnerItems))
	perioMsg.Answers = util.ProtoToAny(&aans)

	sendnum := player.SendToPlayer(player.Id, msg.MsgType_GameTable_HistoryAnswers, &perioMsg)

	if sendnum != 1 {
		return errors.New("send to player error")
	}

	return nil
}

// 广播历史开答列表
func (gm *GameTableRot) CastToPlayerHistoryAnswers(sender *msgrouter.MsgSender, answers *msg_struct.GameRotAnswersPeriodic) error {

	var perioMsg msg.STOC_GameHistoryAnswers

	perioMsg.GameId = gm.GetGameId()
	perioMsg.DeskNo = int32(answers.DeskNo)
	perioMsg.Set = int32(answers.Set)
	perioMsg.Round = int32(answers.Round)
	perioMsg.GameStatus = int32(answers.GameStatus)
	perioMsg.DataType = int32(msg.DataType_ROT)

	var aans msg.WinnersROT

	start := len(answers.Winners) - ShowAnswerNum //暂定100

	if start < 0 {
		start = 0
	}

	answerarray := make([]uint32, 0)
	for i := start; i < len(answers.Winners); i++ {
		answerarray = append(answerarray, uint32(answers.Winners[i]))
	}
	aans.Totalnum = uint32(len(answers.Winners))
	aans.WinnerItems = append(aans.WinnerItems, answerarray...)
	aans.Lastnum = uint32(len(aans.WinnerItems))
	perioMsg.Answers = util.ProtoToAny(&aans)

	clientIds := gm.GetAllClientId()

	//发送客户端游戏历史开答
	sendnum := sender.CastToClient(clientIds, msg.MsgType_GameTable_HistoryAnswers, &perioMsg)

	if sendnum != len(clientIds) {
		return errors.New("error send ")
	}

	return nil
}

// 广播开奖的牌
func (gm *GameTableRot) CastToPlayerOpenAnswer(sender *msgrouter.MsgSender, asn *msg_struct.GameOpenAnswerRot) error {

	var asnMs msg.STOC_GameAnswer

	var ansRot msg.AnswerROT
	ansRot.Winner = uint32(asn.Winner)

	asnMs.DeskNo = int32(asn.DeskNo)
	asnMs.Set = int32(asn.Set)
	asnMs.Round = int32(asn.Round)
	asnMs.GameId = gm.GetGameId()
	asnMs.DataType = int32(msg.DataType_ROT)

	betIds := GlobalGetWinAreaIds(asn.Winner)

	ansRot.WinnerIds = make([]uint32, 0)
	ansRot.WinnerIds = append(ansRot.WinnerIds, betIds...)

	fmt.Printf("30103 MsgType_GameTable_Answer, ansRot=%v\n", ansRot)

	asnMs.Answer = util.ProtoToAny(&ansRot)

	clientIds := gm.GetAllClientId()

	//通知客户端游戏状态改变了
	sendnum := sender.CastToClient(clientIds, msg.MsgType_GameTable_Answer, &asnMs)

	if sendnum != len(clientIds) {
		return errors.New("error send ")
	}

	return nil
}

// 发送给玩家游戏状态
func (gm *GameTableRot) CastToPlayerGameStatus(sender *msgrouter.MsgSender, gsStatus *msg_struct.GameStatusMsg) error {

	var gameGs msg.STOC_GameStatus

	gameGs.GameId = gm.GetGameId()
	gameGs.DeskNo = int32(gsStatus.DeskNo)
	gameGs.Set = int32(gsStatus.Set)
	gameGs.Round = int32(gsStatus.Round)
	gameGs.GameSatus = int32(gsStatus.GameStatus)
	gameGs.IsAnswered = int32(gsStatus.IsAnswered)
	gameGs.BetTime = int32(gm.GetBetSeconds())

	clientIds := gm.GetAllClientId()

	sendnum := sender.CastToClient(clientIds, msg.MsgType_GameTable_Status, &gameGs)

	fmt.Printf("*************** Send Player BetTime:%d gameStatus:%d \n", gameGs.BetTime, gameGs.GameSatus)

	fmt.Printf("clientIds=%v***************\n", clientIds)

	if sendnum != len(clientIds) {
		fmt.Printf("***************")
		return errors.New("error send ")
	}

	return nil
}
