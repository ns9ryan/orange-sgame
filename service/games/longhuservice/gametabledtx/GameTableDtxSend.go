package gameTableDTX

import (
	"errors"
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	betOdds "videogame/common/betodds"
	"videogame/common/msg_struct"
	"videogame/common/proto/msg"
	"videogame/common/util"
	betAreadUtil "videogame/service/betarea"
	"videogame/service/games/msgrouter"
	"videogame/service/games/player"
)

func (gm *GameTableDTX) SendPlayersCardInfo() {

	for _, pp := range gm.MapPlayer {
		//从数据库读取
		gm.SendToPlayerCardInfo(pp)
	}
}

// 通知 玩家 胜额
func (gm *GameTableDTX) SendPlayersWin() {

	wincount := 0

	for _, ppbet := range gm.MapPlayerBet {

		if !ppbet.IsWin() {
			continue
		}
		gm.SendToPlayerBetWin(ppbet)

		wincount++
	}

	//	fmt.Printf("发送有赚钱的玩家 %d 个\n", wincount)

}

// 发送给玩家游戏状态
func (gm *GameTableDTX) SendToPlayerGameStatus(player *player.Player) error {

	var gsStatus msg_struct.GameStatusMsg

	gm.GetGameStatus(&gsStatus)

	var gameGs msg.STOC_GameStatus

	gameGs.GameId = gm.GetGameId()
	gameGs.DeskNo = int32(gsStatus.DeskNo)
	gameGs.Set = int32(gsStatus.Set)
	gameGs.Round = int32(gsStatus.Round)
	gameGs.GameSatus = int32(gsStatus.GameStatus)
	gameGs.IsAnswered = int32(gsStatus.IsAnswered)
	gameGs.BetTime = int32(gm.GetBetSeconds()) //得到还有多少下注时间

	sendnum := player.SendToPlayer(player.Id, msg.MsgType_GameTable_Status, &gameGs)

	fmt.Printf("发送游戏状态给玩家 To Player :%s  gameId:%s, set:%d round:%d gamestatus:%d  betTime:%d \n", player.Id, gameGs.GameId,
		gameGs.Set, gameGs.DeskNo, gameGs.GameSatus, gameGs.BetTime)

	if sendnum != 1 {
		return errors.New("send to player error")
	}

	return nil
}

// 发送历史开答给玩家
func (gm *GameTableDTX) SendToPlayerHistoryAnswers(player *player.Player) error {

	var perioMsg msg.STOC_GameHistoryAnswers

	var answers msg_struct.GameDtxAnswersPeriodic

	gm.GetAnswerPeriodic(&answers)

	perioMsg.GameId = gm.GetGameId()
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

	sendnum := player.SendToPlayer(player.Id, msg.MsgType_GameTable_HistoryAnswers, &perioMsg)

	if sendnum != 1 {
		return errors.New("send to player error")
	}

	fmt.Printf("发送历史开答结果给玩家！！！ To Player :%s  answer:%s \n", player.Id, perioMsg.DataType)

	return nil
}

// SendToPlayerGameSetting 发送游戏配置给玩家
func (gm *GameTableDTX) SendToPlayerGameSetting(player *player.Player) error {

	var gameSetting msg.STOC_GameSetting
	var betAreas []*msg.BetArea

	digits := int(player.GetDigits())

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

		//下面这些以后改成从数据库读取，每个玩家是不一样的
		digits1 := int(player.GetDigits())

		//龙区域
		dragonArea := betAreadUtil.GetBetArea(betOdds.Bet_Dtx_Dragon, int(0.1*float32(digits)), 1000*digits1)
		betAreas = append(betAreas, dragonArea)

		//虎区域
		tigerArea := betAreadUtil.GetBetArea(betOdds.Bet_Dtx_Tiger, int(0.1*float32(digits)), 1000*digits1)
		betAreas = append(betAreas, tigerArea)

		//和区域
		tieArea := betAreadUtil.GetBetArea(betOdds.Bet_Dtx_Tie, int(0.1*float32(digits)), 1000*digits1)
		betAreas = append(betAreas, tieArea)

	}

	betSetting := msg.BetSetting{}

	betSetting.BetAreaNum = int32(len(betAreas))
	betSetting.BetArea = append(betSetting.BetArea, betAreas...)
	betSetting.Digits = int32(digits)

	gameSetting.GameId = gm.GetGameId()
	gameSetting.DeskNo = int32(gm.GameInfo.DeskNo)
	gameSetting.Currency = player.GetCurrency()
	gameSetting.DealerName = gm.GameInfo.GetDealerName() //荷管的名字

	gameSetting.BetNum = 3 //3个区域

	gameSetting.DataType = int32(msg.DataType_DTX)
	gameSetting.TableSetting = util.ProtoToAny(&betSetting)

	chiplist := player.GetChipsShow(player.GetDigits())

	gameSetting.ChipNum = int32(len(chiplist))
	gameSetting.ChipDigits = int32(player.GetDigits())

	//需要从数据库读取
	gameSetting.ChipPoints = append(gameSetting.ChipPoints, chiplist...)

	gameSetting.VideoUrl = gm.GameInfo.VideoURL

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

// SendToPlayerCardInfo 发送玩家的帐号属性信息
func (gm *GameTableDTX) SendToPlayerCardInfo(player *player.Player) error {

	var pyInfo msg.STOC_PlayerInfo

	pyInfo.UserId = player.UserId
	pyInfo.UserName = player.UserName

	pyInfo.Digits = int32(player.GetDigits())
	pyInfo.Credit = uint32(player.GetCredit())
	pyInfo.Currency = player.GetCurrency()

	player.SendMsg(msg.MsgType_PlayerInfo_Card, &pyInfo)

	fmt.Printf("发送给玩家余额 userid:%s , credit:%d , digit:%d  \n", pyInfo.UserId, player.GetCredit(), pyInfo.Digits)

	return nil
}

// 发送玩家胜额
func (gm *GameTableDTX) SendToPlayerBetWin(ppbet *PlayerBet) error {

	var ppwin msg.STOC_GamePlayerWin

	ppwin.Deskno = int32(gm.GetDeskNo())
	ppwin.GameId = gm.GetGameId()
	ppwin.UserId = ppbet.Player.Id
	ppwin.WinPoint = int32(ppbet.CurWin)
	ppwin.Digits = int32(ppbet.Digits)

	ppbet.Player.SendMsg(msg.MsgType_GameTable_PlayerWin, &ppwin)

	fmt.Printf("发送给玩家胜额 userid:%s , win:%d , digit:%d  \n", ppwin.UserId, ppwin.WinPoint, ppwin.Digits)
	log.Errorf("龙虎发送给玩家胜额 userid:%s , win:%d , digit:%d  \n", ppwin.UserId, ppwin.WinPoint, ppwin.Digits)

	return nil
}

func (gm *GameTableDTX) SendToPlayerBetted(player *player.Player) error {

	ppbet, ok := gm.MapPlayerBet[player.Id]

	if ok == false {
		return nil
	}
	if ppbet.TotalBet() == 0 {
		return nil
	}

	bets := ppbet.GetAllBets()

	var betmsg msg.STOC_GamePlayerBetted

	validbet := make([]*msg.BetAmout, 0)

	betlen := len(bets)
	digits := player.GetDigits()

	for i := 0; i < betlen; i++ {

		var betItem msg.BetAmout

		betItem.BetAreaId = int32(bets[i].BetAreaId)
		betItem.BetPoint = int32(bets[i].BetPoint)
		betItem.OrderKey = uint32(bets[i].OrderKey)

		validbet = append(validbet, &betItem)

		fmt.Printf("发送给玩家之前的投注额区域 id:%d bet:%d  orderkey:%d  \n", betItem.BetAreaId, betItem.BetPoint, betItem.OrderKey)
	}

	betmsg.BetItems = append(betmsg.BetItems, validbet...)

	betmsg.GameId = gm.GetGameId()
	betmsg.Deskno = int32(gm.GetDeskNo())
	betmsg.TotalBet = uint32(ppbet.TotalBet())
	betmsg.Digits = int32(digits)

	//发送投注
	player.SendMsg(msg.MsgType_GameTable_PlayerBetted, &betmsg)

	//fmt.Printf("发送给玩家之前的投注额 userid:%s betAmout: %d \n ", player.Id, betmsg.TotalBet/uint32(betmsg.Digits))

	log.Infof("发送给玩家之前的投注额gameId:%s,  userid:%s betAmout: %d \n ", gm.GetGameId(), player.Id, betmsg.TotalBet/uint32(betmsg.Digits))

	return nil
}

// 发送高低注信息给玩家
func (gm *GameTableDTX) SendToPlayerBetLimit(player *player.Player) error {

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

// 发送给玩家游戏状态
func (gm *GameTableDTX) CastToPlayerGameStatus(sender *msgrouter.MsgSender, gsStatus *msg_struct.GameStatusMsg) error {

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

	if sendnum != len(clientIds) {
		return errors.New("error send ")
	}

	return nil
}

// 广播开奖的牌
func (gm *GameTableDTX) CastToPlayerOpenAnswer(sender *msgrouter.MsgSender, asn *msg_struct.GameOpenAnswerDTX) error {

	var asnMs msg.STOC_GameAnswer

	dcard := msg.Card{}
	tcard := msg.Card{}

	dcard.Number = uint32(asn.DragonCard.Number)
	dcard.Color = uint32(asn.DragonCard.Color)
	tcard.Number = uint32(asn.TigerCard.Number)
	tcard.Color = uint32(asn.TigerCard.Color)

	cards := msg.CardDTX{}

	cards.TigerCard = &tcard
	cards.DragonCard = &dcard

	var ansDtx msg.AnswerDTX

	ansDtx.Winner = uint32(asn.Winner)
	ansDtx.Cards = &cards

	asnMs.DeskNo = int32(asn.DeskNo)
	asnMs.Set = int32(asn.Set)
	asnMs.Round = int32(asn.Round)
	asnMs.GameId = gm.GetGameId()
	asnMs.DataType = int32(msg.DataType_DTX)

	asnMs.Answer = util.ProtoToAny(&ansDtx)

	clientIds := gm.GetAllClientId()

	//通知客户端游戏状态改变了
	sendnum := sender.CastToClient(clientIds, msg.MsgType_GameTable_Answer, &asnMs)

	if sendnum != len(clientIds) {
		return errors.New("error send ")
	}

	return nil
}

// 广播历史开答列表
func (gm *GameTableDTX) CastToPlayerHistoryAnswers(sender *msgrouter.MsgSender, answers *msg_struct.GameDtxAnswersPeriodic) error {

	var perioMsg msg.STOC_GameHistoryAnswers

	var Winners msg.WinnersDTX

	perioMsg.GameId = gm.GetGameId()
	perioMsg.DeskNo = int32(answers.DeskNo)
	perioMsg.Set = int32(answers.Set)
	perioMsg.Round = int32(answers.Round)
	perioMsg.GameStatus = int32(answers.GameStatus)
	perioMsg.DataType = int32(msg.DataType_DTX)

	for i := 0; i < len(answers.Winners); i++ {
		Winners.WinnerItems = append(Winners.WinnerItems, int32(answers.Winners[i]))
	}
	//
	perioMsg.Answers = util.ProtoToAny(&Winners)

	clientIds := gm.GetAllClientId()

	//发送客户端游戏历史开答
	sendnum := sender.CastToClient(clientIds, msg.MsgType_GameTable_HistoryAnswers, &perioMsg)

	if sendnum != len(clientIds) {
		return errors.New("error send ")
	}

	return nil
}
