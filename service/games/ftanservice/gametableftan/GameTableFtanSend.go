package gameTableFtan

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
func (gm *GameTableFtan) SendToPlayerGameSetting(player *player.Player) error {

	var gameSetting msg.STOC_GameSetting

	var betAreas []*msg.BetArea

	digits := player.GetDigits()

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

		lowBet := 10
		highBet := 10000

		//小大单双
		for i := betOdds.Bet_Ftan_Small_Big_And_Single_Double_Start; i <= betOdds.Bet_Ftan_Small_Big_And_Single_Double_End; i++ {
			oneArea := betAreadUtil.GetBetArea(i, lowBet, highBet)
			betAreas = append(betAreas, oneArea)
		}

		//番
		for i := betOdds.Bet_Ftan_Fan_Start; i <= betOdds.Bet_Ftan_Fan_End; i++ {
			oneArea := betAreadUtil.GetBetArea(i, lowBet, highBet)
			betAreas = append(betAreas, oneArea)
		}

		//角
		for i := betOdds.Bet_Ftan_Jiao_Start; i <= betOdds.Bet_Ftan_Jiao_End; i++ {
			oneArea := betAreadUtil.GetBetArea(i, lowBet, highBet)
			betAreas = append(betAreas, oneArea)
		}

		//念
		for i := betOdds.Bet_Ftan_Nian_Start; i <= betOdds.Bet_Ftan_Nian_End; i++ {
			oneArea := betAreadUtil.GetBetArea(i, lowBet, highBet)
			betAreas = append(betAreas, oneArea)
		}

		//三红
		for i := betOdds.Bet_Ftan_ThreeRed_Start; i <= betOdds.Bet_Ftan_ThreeRed_End; i++ {
			oneArea := betAreadUtil.GetBetArea(i, lowBet, highBet)
			betAreas = append(betAreas, oneArea)
		}

	}

	// 下注设定
	betSetting := msg.BetSetting{
		BetAreaNum: int32(len(betAreas)),
		BetArea:    betAreas,
		Digits:     int32(digits),
	}

	// 游戏设定
	chipList := player.GetChipsShow(player.GetDigits())

	gameSetting = msg.STOC_GameSetting{
		GameId:       gm.GetGameId(),
		DeskNo:       int32(gm.GameInfo.DeskNo),
		Currency:     player.GetCurrency(),
		DealerName:   gm.GameInfo.Dealer.DealerId, // 荷管的名字
		BetNum:       int32(len(betAreas)),
		DataType:     int32(msg.DataType_FTAN),
		TableSetting: util.ProtoToAny(&betSetting),
		ChipNum:      int32(len(chipList)),
		ChipDigits:   int32(digits),
	}

	//需要从数据库读取
	gameSetting.ChipPoints = append(gameSetting.ChipPoints, chipList...)

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

	fmt.Println("******* Send GameSetting To Player ..")

	return nil
}

// SendToPlayerBetLimit发送高低注信息给玩家
func (gm *GameTableFtan) SendToPlayerBetLimit(player *player.Player) error {

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

func (gm *GameTableFtan) SendPlayersCardInfo() {

	for _, pp := range gm.MapPlayer {
		//从数据库读取
		gm.SendToPlayerCardInfo(pp)
	}
}

// SendToPlayerHistoryAnswers 发送历史开答给玩家
func (gm *GameTableFtan) SendToPlayerHistoryAnswers(player *player.Player) error {
	var perioMsg msg.STOC_GameHistoryAnswers
	var answers msg_struct.FtanAnswersPeriodic

	// 读取历史开答列表
	gm.GetAnswerPeriodic(&answers)

	// 初始化历史消息
	perioMsg = msg.STOC_GameHistoryAnswers{
		GameId:     gm.GetGameId(),
		DeskNo:     int32(answers.DeskNo),
		Set:        int32(answers.Set),
		Round:      int32(answers.Round),
		GameStatus: int32(answers.GameStatus),
		DataType:   int32(msg.DataType_FTAN),
	}

	// 处理历史开奖数据
	start := len(answers.Winners) - ShowAnswerNum
	if start < 0 {
		start = 0
	}

	// 直接创建数组并填充，避免 append 多次调用
	answerArray := make([]uint32, 0)
	for i := start; i < len(answers.Winners); i++ {
		answerArray = append(answerArray, uint32(answers.Winners[i]))
	}

	aans := msg.WinnersFtan{
		Totalnum: uint32(len(answers.Winners)),
		Lastnum:  uint32(len(answerArray)),
	}

	aans.WinnerItems = append(aans.WinnerItems, answerArray...)

	// 序列化 Answers 并附加到消息
	perioMsg.Answers = util.ProtoToAny(&aans)

	// 发送数据
	if sendNum := player.SendToPlayer(player.Id, msg.MsgType_GameTable_HistoryAnswers, &perioMsg); sendNum != 1 {
		return errors.New("send to player error")
	}

	return nil
}

// 广播历史开答列表
func (gm *GameTableFtan) CastToPlayerHistoryAnswers(answers *msg_struct.FtanAnswersPeriodic) error {

	// 初始化历史消息
	perioMsg := msg.STOC_GameHistoryAnswers{
		GameId:     gm.GetGameId(),
		DeskNo:     int32(answers.DeskNo),
		Set:        int32(answers.Set),
		Round:      int32(answers.Round),
		GameStatus: int32(answers.GameStatus),
		DataType:   int32(msg.DataType_FTAN),
	}

	start := len(answers.Winners) - ShowAnswerNum //暂定100

	if start < 0 {
		start = 0
	}

	answerArray := make([]uint32, 0)
	for i := start; i < len(answers.Winners); i++ {
		answerArray = append(answerArray, uint32(answers.Winners[i]))
	}

	// 构建开奖历史数据
	aans := msg.WinnersFtan{
		Totalnum: uint32(len(answers.Winners)),
		Lastnum:  uint32(len(answerArray)),
	}

	aans.WinnerItems = append(aans.WinnerItems, answerArray...)

	// 序列化 Answers 并附加到消息
	perioMsg.Answers = util.ProtoToAny(&aans)

	// 获取所有客户端 ID 并广播消息
	clientIds := gm.GetAllClientId()
	sendNum := gm.Gs.GetMsgSender().CastToClient(clientIds, msg.MsgType_GameTable_HistoryAnswers, &perioMsg)

	// 检查发送数量是否匹配
	if sendNum != len(clientIds) {
		return errors.New("error sending history answers to clients")
	}

	return nil
}

// CastToPlayerOpenAnswer 广播游戏开奖
func (gm *GameTableFtan) CastToPlayerOpenAnswer(asn *msg_struct.FtanOpenAnswer) error {
	if asn == nil {
		return errors.New("invalid input: asn is nil")
	}

	// 检测开答是否在正确范围内
	if asn.Winner < 1 || asn.Winner > TotalAnswer {
		return errors.New("error win index")
	}

	// 构造开奖信息
	answerFtan := msg.AnswerFtan{
		Winner:    uint32(asn.Winner),
		WinnerIds: GlobalGetWinAreaIds(asn.Winner), //高量区域id
	}

	asnMs := msg.STOC_GameAnswer{
		DeskNo:   int32(asn.DeskNo),
		Set:      int32(asn.Set),
		Round:    int32(asn.Round),
		GameId:   gm.GetGameId(),
		DataType: int32(msg.DataType_FTAN),
		Answer:   util.ProtoToAny(&answerFtan),
	}

	// 获取所有客户端 ID 并发送消息
	clientIds := gm.GetAllClientId()
	sendnum := gm.Gs.GetMsgSender().CastToClient(clientIds, msg.MsgType_GameTable_Answer, &asnMs)

	// 记录日志
	//fmt.Printf("[CastToPlayerOpenAnswer] Sent to %d/%d clients | DeskNo: %d | Set: %d | Round: %d | Winner: %d\n",
	//	sendnum, len(clientIds), asnMs.DeskNo, asnMs.Set, asnMs.Round, asn.Winner)

	// 发送数量不匹配时返回错误
	if sendnum != len(clientIds) {
		return errors.New("failed to send to some clients")
	}

	return nil
}

// CastToPlayerGameStatus 广播游戏状态给玩家
func (gm *GameTableFtan) CastToPlayerGameStatus(gsStatus *msg_struct.GameStatusMsg) error {
	// 组装游戏状态消息
	gameGs := msg.STOC_GameStatus{
		GameId:     gm.GetGameId(),
		DeskNo:     int32(gsStatus.DeskNo),
		Set:        int32(gsStatus.Set),
		Round:      int32(gsStatus.Round),
		GameSatus:  int32(gsStatus.GameStatus),
		IsAnswered: int32(gsStatus.IsAnswered),
		BetTime:    int32(gm.GetBetSeconds()),
	}

	// 获取所有客户端 ID 并广播消息
	clientIds := gm.GetAllClientId()
	sendNum := gm.Gs.GetMsgSender().CastToClient(clientIds, msg.MsgType_GameTable_Status, &gameGs)

	// 校验发送数量
	if sendNum != len(clientIds) {
		return fmt.Errorf("failed to send game status: sent %d, expected %d", sendNum, len(clientIds))
	}

	//fmt.Printf("------>游戏状态---->玩家------>data=%v\n", gameGs)

	return nil
}

// 若玩家上线时发送之前投注未开答的投注额
func (gm *GameTableFtan) SendToPlayerBetted(player *player.Player) error {

	// 获取玩家投注信息
	ppbet, exists := gm.MapPlayerBet[player.Id]

	if !exists || ppbet.TotalBet() == 0 {
		return nil
	}

	var betmsg msg.STOC_GamePlayerBetted

	bets := ppbet.GetBetNode()
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

	log.Infof("发送给玩家之前的投注额gameId:%s,  userid:%s betAmout: %d \n ", gm.GetGameId(), player.Id, betmsg.TotalBet/uint32(betmsg.Digits))

	return nil
}

// SendToPlayerBetWin 发送玩家胜额
func (gm *GameTableFtan) SendToPlayerBetWin(ppbet *PlayerBet) error {
	if ppbet == nil || ppbet.Player == nil {
		return errors.New("invalid player bet data")
	}

	// 构造胜额消息
	ppwin := msg.STOC_GamePlayerWin{
		Deskno:   int32(gm.GetDeskNo()),
		GameId:   gm.GetGameId(),
		UserId:   ppbet.Player.Id,
		WinPoint: int32(ppbet.CurWin),
		Digits:   int32(ppbet.Digits),
	}

	// 发送消息
	ppbet.Player.SendMsg(msg.MsgType_GameTable_PlayerWin, &ppwin)

	// 记录日志
	//fmt.Printf("[SendToPlayerBetWin] UserID: %s | Win: %d | Digits: %d\n", ppwin.UserId, ppwin.WinPoint, ppwin.Digits)

	return nil
}

// 通知 玩家 胜额
func (gm *GameTableFtan) SendPlayersWin() {

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
