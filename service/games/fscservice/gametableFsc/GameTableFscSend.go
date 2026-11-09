package gameTableFsc

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
func (gm *GameTableFsc) SendToPlayerGameSetting(player *player.Player) error {

	var gameSetting msg.STOC_GameSetting

	var betAreas []*msg.BetArea

	betAreas = make([]*msg.BetArea, 0)

	digits := int(player.GetDigits())

	//从数据库读取高低注，若读不成功，刚自已初始化
	betAreasTemp, err := player.GetBetLimit(player.BetLimitTemplateId, gm.GetGameId())

	if err == nil {

		for _, item := range betAreasTemp {

			//赔率内部处理了
			areaitem := betAreadUtil.GetBetArea(int(item.AreaId), int(item.LowBet), int(item.HighBet))
			betAreas = append(betAreas, areaitem)
		}

	} else {

		//单一点数
		for i := betOdds.Bet_Fsc_Single_Point1; i <= betOdds.Bet_Fsc_Single_Point6; i++ {
			oneArea := betAreadUtil.GetBetArea(i, 10, 1000)
			betAreas = append(betAreas, oneArea)
		}

	}

	betSetting := msg.BetSetting{}

	betSetting.BetAreaNum = int32(len(betAreas))
	betSetting.BetArea = append(betSetting.BetArea, betAreas...)
	betSetting.Digits = int32(digits)

	//下面这些以后改成从数据库读取，每个玩家是不一样的
	gameSetting.GameId = gm.GetGameId()
	gameSetting.DeskNo = int32(gm.GameInfo.DeskNo)
	gameSetting.Currency = player.GetCurrency()
	gameSetting.DealerName = gm.GameInfo.Dealer.DealerId //荷管的名字

	gameSetting.BetNum = int32(len(betAreas))

	gameSetting.DataType = int32(msg.DataType_FSC)
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

	//fmt.Printf("*******Send GameSetting ToPlayer ..  \n")

	return nil
}

// 发送高低注信息给玩家
func (gm *GameTableFsc) SendToPlayerBetLimit(player *player.Player) error {

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

func (gm *GameTableFsc) SendPlayersCardInfo() {

	for _, pp := range gm.MapPlayer {
		//从数据库读取
		gm.SendToPlayerCardInfo(pp)
	}
}

// 发送历史开答给玩家
func (gm *GameTableFsc) SendToPlayerHistoryAnswers(player *player.Player) error {

	var perioMsg msg.STOC_GameHistoryAnswers

	var answers msg_struct.FscAnswersPeriodic

	//先读取历史开答列表
	gm.GetAnswerPeriodic(&answers)

	perioMsg.GameId = gm.GetGameId()
	perioMsg.DeskNo = int32(answers.DeskNo)
	perioMsg.Set = int32(answers.Set)
	perioMsg.Round = int32(answers.Round)
	perioMsg.GameStatus = int32(answers.GameStatus)
	perioMsg.DataType = int32(msg.DataType_FSC)

	var aans msg.WinnersFsc

	start := len(answers.Winners) - ShowAnswerNum //暂定100

	if start < 0 {
		start = 0
	}

	answerarray := make([]uint32, 0)
	for i := start; i < len(answers.Winners); i++ {
		answerarray = append(answerarray, uint32(answers.Winners[i]))
	}
	aans.Totalnum = int32(len(answers.Winners))
	aans.WinnerItems = append(aans.WinnerItems, answerarray...)
	aans.Lastnum = int32(len(aans.WinnerItems))
	perioMsg.Answers = util.ProtoToAny(&aans)

	sendnum := player.SendToPlayer(player.Id, msg.MsgType_GameTable_HistoryAnswers, &perioMsg)

	if sendnum != 1 {
		return errors.New("send to player error")
	}

	return nil
}

// 广播历史开答列表
func (gm *GameTableFsc) CastToPlayerHistoryAnswers(answers *msg_struct.FscAnswersPeriodic) error {

	var perioMsg msg.STOC_GameHistoryAnswers

	perioMsg.GameId = gm.GetGameId()
	perioMsg.DeskNo = int32(answers.DeskNo)
	perioMsg.Set = int32(answers.Set)
	perioMsg.Round = int32(answers.Round)
	perioMsg.GameStatus = int32(answers.GameStatus)
	perioMsg.DataType = int32(msg.DataType_FSC)

	var aans msg.WinnersFsc

	start := len(answers.Winners) - ShowAnswerNum //暂定100

	if start < 0 {
		start = 0
	}

	answerarray := make([]uint32, 0)
	for i := start; i < len(answers.Winners); i++ {
		answerarray = append(answerarray, uint32(answers.Winners[i]))
	}
	aans.Totalnum = int32(len(answers.Winners))
	aans.WinnerItems = append(aans.WinnerItems, answerarray...)
	aans.Lastnum = int32(len(aans.WinnerItems))
	perioMsg.Answers = util.ProtoToAny(&aans)

	clientIds := gm.GetAllClientId()

	//发送客户端游戏历史开答
	sendnum := gm.Gs.GetMsgSender().CastToClient(clientIds, msg.MsgType_GameTable_HistoryAnswers, &perioMsg)

	if sendnum != len(clientIds) {
		return errors.New("error send ")
	}

	return nil
}

// 广播游戏开奖
func (gm *GameTableFsc) CastToPlayerOpenAnswer(asn *msg_struct.FscOpenAnswer) error {

	var asnMs msg.STOC_GameAnswer
	var answer msg.AnswerFsc

	//检测开答是否在正确的
	if !(asn.Winner >= 0 && asn.Winner < TotalAnswer) {
		return errors.New("error win index !")
	}

	answer.Winner = uint32(asn.Winner)

	answer.FscA = uint32(FscAnswerInds[asn.Winner][0])
	answer.FscB = uint32(FscAnswerInds[asn.Winner][1])
	answer.FscC = uint32(FscAnswerInds[asn.Winner][2])

	betIds := GlobalGetWinAreaIds(asn.Winner)

	answer.WinnerIds = make([]uint32, 0)
	answer.WinnerIds = append(answer.WinnerIds, betIds...)

	asnMs.DeskNo = int32(asn.DeskNo)
	asnMs.Set = int32(asn.Set)
	asnMs.Round = int32(asn.Round)
	asnMs.GameId = gm.GetGameId()
	asnMs.DataType = int32(msg.DataType_FSC)

	asnMs.Answer = util.ProtoToAny(&answer)

	clientIds := gm.GetAllClientId()

	//通知客户端游戏状态改变了
	sendnum := gm.Gs.GetMsgSender().CastToClient(clientIds, msg.MsgType_GameTable_Answer, &asnMs)

	if sendnum != len(clientIds) {
		return errors.New("error send ")
	}

	return nil
}

// 广播游戏状态给玩家
func (gm *GameTableFsc) CastToPlayerGameStatus(gsStatus *msg_struct.GameStatusMsg) error {

	var gameGs msg.STOC_GameStatus

	gameGs.GameId = gm.GetGameId()
	gameGs.DeskNo = int32(gsStatus.DeskNo)
	gameGs.Set = int32(gsStatus.Set)
	gameGs.Round = int32(gsStatus.Round)
	gameGs.GameSatus = int32(gsStatus.GameStatus)
	gameGs.IsAnswered = int32(gsStatus.IsAnswered)
	gameGs.BetTime = int32(gm.GetBetSeconds())

	clientIds := gm.GetAllClientId()

	sendnum := gm.Gs.GetMsgSender().CastToClient(clientIds, msg.MsgType_GameTable_Status, &gameGs)

	if sendnum != len(clientIds) {
		return errors.New("error send ")
	}

	return nil
}

// 若玩家上线时发送之前投注未开答的投注额
func (gm *GameTableFsc) SendToPlayerBetted(player *player.Player) error {

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
	betmsg.TotalBet = uint32(ppbet.TotalBet())
	betmsg.Digits = int32(player.GetDigits())

	//发送投注
	player.SendMsg(msg.MsgType_GameTable_PlayerBetted, &betmsg)

	//	fmt.Printf("发送给玩家之前的投注额 userid:%s betAmout: %d \n ", player.Id, betmsg.TotalBet/uint32(betmsg.Digits))
	log.Infof("发送给玩家之前的投注额gameId:%s,  userid:%s betAmout: %d \n ", gm.GetGameId(), player.Id, betmsg.TotalBet/uint32(betmsg.Digits))

	return nil
}

// 发送玩家胜额
func (gm *GameTableFsc) SendToPlayerBetWin(ppbet *PlayerBet) error {

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
func (gm *GameTableFsc) SendPlayersWin() {

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
