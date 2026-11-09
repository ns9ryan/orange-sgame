package gameTableDTX

import (
	"fmt"
	"github.com/duanhf2012/origin/v2/event"
	"github.com/duanhf2012/origin/v2/log"
	"strconv"
	"time"
	"videogame/common/betnode"
	"videogame/common/eventid"
	"videogame/common/global/variable"
	"videogame/common/msg_struct"
	"videogame/common/proto/msg"
	redisId "videogame/common/redis"
	"videogame/common/util"
	"videogame/service/games/gameTable"
	"videogame/service/games/player"
)

type GameTableDTX struct {
	*gameTable.Table //这个不是一定要

	gameTable.GameTable

	MapPlayerBet map[string]*PlayerBet //玩家的投注
}

// ReadConfig 读取游戏配置
func (gm *GameTableDTX) ReadConfig() {

	gm.GameInfo.GameId = msg_struct.GAMEID_DTX

	gm.GameInfo.DeskNo = 4
	gm.GameInfo.Set = 0
	gm.GameInfo.Round = 0
	gm.GameInfo.GameStatus = 0
	gm.GameInfo.IsAnswered = false
	gm.GameInfo.TotalBetTime = 10
	gm.GameInfo.GameStatus = 0

	//去读取数据库的配置参数
	gm.ReadDBConfig(msg_struct.GAMEID_DTX)

	//检测投注时间是否为0
	if gm.GameInfo.TotalBetTime == 0 {
		gm.GameInfo.TotalBetTime = 10
	}

	//var merchantTable models.MerchantCurrency

	//merchantTable.UpdateMerchantAccount("merchant001", 100, 100, 100, -50)

}

// 从redis读取历史开答表
func (gm *GameTableDTX) ReadHistoryAnswersRedis() {

	//先读取游戏桌状态
	redisKey := redisId.RedisId_GameAnswers + gm.GameInfo.GetGameId()

	for i := 0; i < gm.GameInfo.Round; i++ {

		filedkey := fmt.Sprintf("round:%03d", i+1)

		bytes, err := variable.RedisStorage.HGet(redisKey, filedkey)

		if err != nil {
			//出错处理，没有路子
			continue
		}

		answer, err1 := strconv.Atoi(bytes)

		if err1 != nil {
			continue
		}

		gm.AnswerBar = append(gm.AnswerBar, answer)
	}

	fmt.Println("成功读取路子信息。。。。。len:", len(gm.AnswerBar))

}

// 存历史开答表到redis
func (gm *GameTableDTX) SaveAnswerToRedis(answerBar *msg_struct.GameOpenAnswerDTX) {

	strInt := strconv.Itoa(answerBar.Winner)

	redisKey := redisId.RedisId_GameAnswers + gm.GameInfo.GetGameId()

	filedkey := fmt.Sprintf("round:%03d", answerBar.Round)

	variable.RedisStorage.HSet(redisKey, filedkey, strInt)

}

func (gm *GameTableDTX) ClearRedisHistoryAnswers() {

	redisKey := redisId.RedisId_GameAnswers + gm.GameInfo.GetGameId()

	variable.RedisStorage.Del(redisKey)

}

func (gm *GameTableDTX) InitGame() {

	gm.MapPlayerBet = make(map[string]*PlayerBet)

	gm.ReadConfig() //读取数据库的配置

	//从redis读取数据,恢复现场
	gm.ReadTableInfoRedis()

	//从redis读取历名开答表
	gm.ReadHistoryAnswersRedis()

}

// IsAnswered 是否开答了
func (gm *GameTableDTX) IsAnswered() int {

	if gm.GameInfo.IsAnswered == true {
		return 1
	}

	return 0
}

func (gm *GameTableDTX) GetPlayerNum() int {

	return len(gm.MapPlayer)
}

func (gm *GameTableDTX) GetTableConfig(gameconfig *msg_struct.DTXToCtrlGameConfig) {

	gameconfig.GameId = gm.GameInfo.GameId
	gameconfig.DeskNo = gm.GameInfo.DeskNo
	gameconfig.TotalBetTime = gm.GameInfo.TotalBetTime
	gameconfig.VideoUrl = gm.GameInfo.VideoURL

}

// // 每一局开始需要处理的
func (gm *GameTableDTX) OnSetStart() {

	gm.AnswerBar = gm.AnswerBar[:0]
	gm.MapPlayerBet = make(map[string]*PlayerBet)

	//清除redis上面的历史开答
	gm.ClearRedisHistoryAnswers()

}

func (gm *GameTableDTX) OnStartBet() {

	fmt.Printf("GameTableBacx OnStartBet\n")
}

// 洗牌的回调
func (gm *GameTableDTX) OnInterMission() {

}

// 休场的回调
func (gm *GameTableDTX) OnSleep() {

}

func (gm *GameTableDTX) OnStopBet() {

	fmt.Println("StopBetEvent ")

	gm.GameInfo.GameStatus = msg_struct.GameStatusStopBet

	var gsStatus msg_struct.GameStatusMsg

	gsStatus.GameStatus = gm.GameInfo.GameStatus
	gsStatus.Set = gm.GameInfo.Set
	gsStatus.Round = gm.GameInfo.Round
	gsStatus.CurrentTime = -1
	gsStatus.DeskNo = gm.GameInfo.DeskNo
	gsStatus.RoundId = gm.GameInfo.RoundId //new add 2025.3.6

	if gm.GameInfo.IsAnswered == false {
		gsStatus.IsAnswered = 0
	} else {
		gsStatus.IsAnswered = 1
	}

	var ret int

	rpcFun := util.FormatRpcMethod(gm.Gs.GetName(), "RPCStopBetToCtrlService")

	//通知控端
	err := gm.Gs.Call(rpcFun, &gsStatus, &ret)

	if err != nil {
		return
	}

	//通知各个玩家,游戏状态停止下注
	gm.BroadGameStatus()
}

// 读取还有多少秒可以下注
func (gm *GameTableDTX) GetBetSeconds() int {

	var ret int

	//只有在投注状态下，才可以得到实际时间
	if gm.GameInfo.GameStatus == msg_struct.GameStatusBet {

		nowTime := time.Now()
		offSecond := nowTime.Sub(gm.GameInfo.StartBetTime)
		ret = gm.GameInfo.TotalBetTime - int(offSecond.Seconds())

		//fmt.Printf("********NowTime:%s -    StartBetTime:%s  \n", nowTime.String(), gm.GameInfo.StartBetTime.String())

	} else {

		//fmt.Printf("GetBetSeconds ....gamestatus .%d ", gm.GameInfo.GameStatus)

		ret = -1
	}

	return ret
}

// 停止下注事件
func (gm *GameTableDTX) StopBetEvent() {

	//fmt.Println("StopBetEvent ")

	gm.GameInfo.GameStatus = msg_struct.GameStatusStopBet

	var gsStatus msg_struct.GameStatusMsg

	gsStatus.GameStatus = gm.GameInfo.GameStatus
	gsStatus.Set = gm.GameInfo.Set
	gsStatus.Round = gm.GameInfo.Round
	gsStatus.CurrentTime = -1
	gsStatus.DeskNo = gm.GameInfo.DeskNo

	if gm.GameInfo.IsAnswered == false {
		gsStatus.IsAnswered = 0
	} else {
		gsStatus.IsAnswered = 1
	}

	var ret int

	rpcFun := util.FormatRpcMethod(util.LongHuService, "RPCStopBetToCtrlService")

	//通知控端,本地RPC调用
	err := gm.Gs.Call(rpcFun, &gsStatus, &ret)

	if err != nil {
		return
	}

	//通知各个玩家,游戏状态停止下注
	gm.BroadGameStatus()

}

// // 是否已经开答
func (gm *GameTableDTX) IsOpenAnswer() bool {

	return gm.GameInfo.IsAnswered
}

// 开奖
func (gm *GameTableDTX) OpenAnswer(answer *msg_struct.GameOpenAnswerDTX) bool {

	//若已经开答，则返回
	if gm.GameInfo.IsAnswered == true {
		return false
	}

	answer.GameId = gm.GetGameId()

	//通知玩家，处理结算
	gm.AnswerBar = append(gm.AnswerBar, answer.Winner)

	gm.GameInfo.IsAnswered = true

	//fmt.Printf("GameTable  openanswer winner :%d   AnswerBar :%v \n", answer.Winner, gm.AnswerBar)

	//存游戏桌状态到redis
	gm.SaveTableInfoToRedis()

	//存当前开答到redis
	gm.SaveAnswerToRedis(answer)

	//广播开奖消息给玩家
	gm.BroadGameAnswer(answer)

	//通知大厅开答历史变了
	gm.NotifyLobbyUpateHistoryAnswers()

	//开奖结算
	gm.GetEventHandler().NotifyEvent(&event.Event{
		Type: eventid.Event_Answer_Settle,
		Data: answer,
	})

	return true
}

// 读取历史开答表
func (gm *GameTableDTX) GetAnswerPeriodic(answerBar *msg_struct.GameDtxAnswersPeriodic) {

	answerBar.Set = gm.GameInfo.Set
	answerBar.Round = gm.GameInfo.Round
	answerBar.DeskNo = gm.GameInfo.DeskNo
	answerBar.GameStatus = gm.GameInfo.GameStatus

	answerBar.Winners = append(answerBar.Winners, gm.AnswerBar...)
	answerBar.WinnerNum = len(answerBar.Winners)

	//	fmt.Printf("GetAnswerPeriodic :%v", answerBar.Winners)
}

// 广播游戏当前开答结果
func (gm *GameTableDTX) BroadGameAnswer(answer *msg_struct.GameOpenAnswerDTX) {

	//fmt.Printf("开答消息发送给玩家 :%v", answer)

	gm.GetEventHandler().NotifyEvent(&event.Event{
		Type: eventid.Event_Broad_Answer,
		Data: *answer,
	})
}

// 广播游戏历史开答列表，利用事件通知功能
func (gm *GameTableDTX) BroadGameHistoryAnswers() {

	var answerBar msg_struct.GameDtxAnswersPeriodic

	gm.GetAnswerPeriodic(&answerBar)

	//fmt.Printf("Broad To Client AnswerPeriodic :%v", answerBar.Winners)

	gm.GetEventHandler().NotifyEvent(&event.Event{
		Type: eventid.Event_Broad_History_answers,
		Data: answerBar,
	})
}

// 更新大厅更新历史开答
func (gm *GameTableDTX) NotifyLobbyUpateHistoryAnswers() {

	var answerBar msg_struct.GameDtxAnswersPeriodic

	gm.GetAnswerPeriodic(&answerBar)

	//fmt.Printf("*************Notify To Lobby AnswerPeriodic :%v", answerBar.Winners)

	gm.GetEventHandler().NotifyEvent(&event.Event{
		Type: eventid.Event_NotifyLobby_History_Answers,
		Data: answerBar,
	})

}

// 广播游戏状态
func (gm *GameTableDTX) BroadGameStatus() {

	var gsStatus msg_struct.GameStatusMsg

	gm.GetGameStatus(&gsStatus)

	gm.GetEventHandler().NotifyEvent(&event.Event{
		Type: eventid.Event_Broad_GameStatus,
		Data: gsStatus,
	})
}

// 广播玩家的钱包
func (gm *GameTableDTX) BroadPlayerWallet() {

	gm.GetEventHandler().NotifyEvent(&event.Event{
		Type: eventid.Event_Broad_PlayerCard,
	})

}

// PlayerBet 玩家投注，返回投注总额，及是否投注成功
func (gm *GameTableDTX) PlayerBet(betbar *betnode.BetNode) (int, int) {

	ppbet, ok := gm.MapPlayerBet[betbar.UserId]

	bNew := false

	if betbar.Digits == 0 {
		betbar.Digits = 1
	}

	areadId := betbar.BetAreaId
	bet := betbar.BetPoint
	userId := betbar.UserId

	log.Infof("压注龙虎游戏[%s]:gameId:%s   userId:%s  betAreaId: %d , betMoney:%d Digits:%d 处理开始", betbar.OrkeyUid, gm.GameInfo.GameId, betbar.UserId, betbar.BetAreaId, betbar.BetPoint, betbar.Digits)

	//若玩家投注列表还没有，则需要生成
	if ok == false {

		pp, ok1 := gm.GetPlayer(betbar.ClientId)

		if ok1 == false {
			//需要处理没有找到此玩家的代码
			log.Infof("压注龙虎游戏:gameId:%s  userId:%s 不成功，内存队列找不到此玩家！ ", gm.GameInfo.GameId, betbar.UserId)
			return 0, int(msg.ErrCode_BetErr_ServerClusterErr)
		}

		//产生新的投注玩家，后面需要改用内存池方案
		newpbet := PlayerBet{}
		newpbet.Init(pp)
		ppbet = &newpbet
		bNew = true

		//2025.5.23
		credit := pp.GetCredit()
		ppbet.SetBeforeBalance(credit)

	}

	curPlayerBet := PlayerBet{Player: ppbet.Player, Digits: int(ppbet.Player.GetDigits())}

	//拷贝一份投注用于检测
	curPlayerBet.DuplicateBet(ppbet)

	//压注增加,内部有检测是否有打水
	if curPlayerBet.Bet(areadId, bet) == false {

		log.Errorf("玩家压注龙虎游戏 user:%s, betId:%d , betpoint:%d ,failed!!!", ppbet.Player.UserId, areadId, bet)

		return 0, int(msg.ErrCode_BetErr_ServerBetIdErr)
	}

	//检测是否少于低注
	if curPlayerBet.IsLessLowBet(areadId) == true {

		log.Errorf("玩家压注龙虎游戏: userId:%s  betAreaId: %d , betMoney:%d Digits:%d 少于低注", betbar.UserId, betbar.BetAreaId, betbar.BetPoint, betbar.Digits)

		return 0, int(msg.ErrCode_BetErr_LessLowBetting)
	}

	//检测是否超出高注
	if curPlayerBet.IsOverHighBet(areadId) == true {

		log.Errorf("玩家压注龙虎游戏: userId:%s  betAreaId: %d , betMoney:%d Digits:%d 高于高注", betbar.UserId, betbar.BetAreaId, betbar.BetPoint, betbar.Digits)

		return 0, int(msg.ErrCode_BetErr_OverHighBetting)
	}

	//从钱包服读取当前余额，并检测投注额是否大于余额
	//decPoint := float32(bet) / float32(betbar.Digits)

	//这边需要检测是单钱包API调用或者本地钱包处理
	optData := player.OptData{
		UserId:    ppbet.Player.UserId,
		Point:     uint(bet),
		Digits:    uint(betbar.Digits),
		Opt:       1,
		GameId:    gm.GameInfo.GetGameId(),
		DeskNo:    gm.GameInfo.DeskNo,
		TransType: 0, //下注订单
		TransId:   betbar.OrkeyUid,
		RoundId:   gm.GameInfo.RoundId,
	}

	if ppbet.Player.DecWalletBalance(&optData) == false {

		//fmt.Println("钱包服扣款出错！！！")
		log.Errorf("玩家压注龙虎游戏: userId:%s  betAreaId: %d , betMoney:%d Digits:%d  钱包服扣款出错！！！", betbar.UserId, betbar.BetAreaId, betbar.BetPoint, betbar.Digits)
		return 0, int(msg.ErrCode_BetErr_ServerClusterErr)
	}

	//检测都通过了，可以拷贝最新投注
	ppbet.DuplicateBet(&curPlayerBet)

	//若是新产生的，需要加入投注列表
	if bNew == true {
		gm.MapPlayerBet[userId] = ppbet
		//fmt.Println("新产生了")
	}

	//投注入队列
	ppbet.PushBet(*betbar)
	totalbet := int(ppbet.TotalBet())

	log.Infof("压注龙虎游戏[%s]:gameId:%s   userId:%s  betAreaId: %d , betMoney:%d Digits:%d 成功！", betbar.OrkeyUid, gm.GameInfo.GameId, betbar.UserId, betbar.BetAreaId, betbar.BetPoint, betbar.Digits)

	log.Infof("*********************************************************")

	return totalbet, 0
}

// PlayerBet 撤消玩家压注,返回投注项，投注额度， 错误码
func (gm *GameTableDTX) PlayerCancelBet(userId string) (betnode.BetNode, int, int) {

	log.Infof("******GameTableDTX-->PlayerCancelBet :gameId:%s,  userid:---%s 开始处理！！ ", gm.GameInfo.GameId, userId)

	ppbet, ok := gm.MapPlayerBet[userId]

	//若玩家投注列表还没有，则需要生成
	if ok == false {

		log.Errorf("******PlayerCancelBet :gameId:%s,  userid:---%s  faile 找不到此投注玩家的索引  ", gm.GameInfo.GameId, userId)

		//需要处理没有找到此玩家的代码
		return betnode.BetNode{}, 0, int(msg.ErrCode_BetErr_ServerEmptyBet)
	}

	betNode, bOk := ppbet.CancelLastBet()

	if bOk == false {

		log.Errorf("******PlayerCancelBet :gameId:%s,  userid:---%s  faile 没有可撤消的  ", gm.GameInfo.GameId, userId)

		return betnode.BetNode{}, 0, int(msg.ErrCode_BetErr_ServerEmptyBet)
	}

	//从钱包服读取当前余额，并检测投注额是否大于余额
	//bet := float32(betNode.BetPoint) / float32(betNode.Digits)

	//这边需要检测是单钱包API调用或者本地钱包处理
	optData := player.OptData{
		UserId:    ppbet.Player.UserId,
		Point:     uint(betNode.BetPoint),
		Digits:    uint(betNode.Digits),
		Opt:       0,
		GameId:    gm.GameInfo.GameId,
		DeskNo:    gm.GameInfo.DeskNo,
		TransType: 1, //取消下注订单
		TransId:   betNode.OrkeyUid,
		RoundId:   gm.GameInfo.RoundId,
	}

	if ppbet.Player.AddWalletBalance(&optData) == false {

		//fmt.Println("钱包服扣款出错！！！")
		log.Errorf("玩家撤消压注龙虎游戏: userId:%s   钱包服扣款出错！！！", ppbet.Player.UserId)

		return betnode.BetNode{}, 0, int(msg.ErrCode_BetErr_ServerClusterErr)

	}

	//若总投注额=0，则清掉这个数据
	if ppbet.TotalBet() == 0 {
		delete(gm.MapPlayerBet, userId)
	}

	//回复给玩家这个投注额撤消成功了
	totalbet := int(ppbet.TotalBet())

	log.Infof("******GameTableDTX-->PlayerCancelBet :gameId:%s,  userid:---%s 处理成功！ ", gm.GameInfo.GameId, userId)

	return betNode, totalbet, 0

}

// SettleWinLost 结算输嬴
func (gm *GameTableDTX) SettleWinLost(answer *msg_struct.GameOpenAnswerDTX) {

	betItemImp := CreateBetItemImp(msg_struct.GameTypeDtx, gm.GameInfo.GameSubCode) //投注明细入库用

	set := gm.GameInfo.Set
	round := gm.GameInfo.Round
	roundId := gm.GameInfo.RoundId
	winner := answer.Winner

	//把每个会员的没下注局数都加1
	gm.StatistNoBetRound()

	for _, ppbet := range gm.MapPlayerBet {

		win := ppbet.WinLost(winner)

		//更新玩家余额到数据库
		if win > 0 {

			if ppbet.Digits == 0 {
				ppbet.Digits = 1
			}
			//winf := float32(win) / float32(ppbet.Digits)

			optData := player.OptData{
				UserId:    ppbet.Player.UserId,
				Point:     uint(win),
				Digits:    uint(ppbet.Digits),
				Opt:       0,
				GameId:    gm.GameInfo.GameId,
				DeskNo:    gm.GameInfo.DeskNo,
				TransType: 3, //结算订单
				TransId:   fmt.Sprintf("%s_%s", gm.GameInfo.RoundId, ppbet.Player.UserId),
				RoundId:   gm.GameInfo.RoundId,
			}

			//胜的点数加入玩家余额
			if ppbet.Player.AddWalletBalance(&optData) == false {

				log.Errorf("结算加点出错:gameId:%s ,userId:%s", optData.GameId, optData.UserId)
			}
		}

		credit := ppbet.Player.GetCredit()
		ppbet.SetAfterBalance(credit)

		//存储投注列到数据库(不是明细，是投注统计)
		ppbet.SaveBetCollectToDBAsync(gm, set, round, roundId, answer)

		//把每个玩家的投注明细列表复一份
		betItemImp.Copy(ppbet)

		//更新玩家帐目到数据库
		ppbet.UpdateAccountToDB()

		//增加商户的帐目数据到队列
		gm.MerchantAddAccount(ppbet.Player.GetParentId(), &ppbet.CurAccount)

		//连续没下注的局数清0
		ppbet.Player.ResetNoBetRound()

		log.Infof("gameId:%s 结算玩家：%s 完成!!", gm.GetGameId(), ppbet.Player.UserId)
	}

	//有人有投注才进入
	if len(gm.MapPlayerBet) > 0 {

		//计算投注每个投注明细的胜负
		betItemImp.CalcWinLost(set, round, roundId, answer)

		//投注明细入库,后面需要改成独立协程处理
		betItemImp.SaveToDB(gm)

		//把商户帐目存入数据，人多时需要改成发放另一服务器专门存储
		gm.DoMerchantAccountQueueToDB(3)

	}

}

func (gm *GameTableDTX) ClearMapPlayerBet() {

	gm.MapPlayerBet = make(map[string]*PlayerBet)
}

// 检测玩家是否有投注
func (gm *GameTableDTX) IsPlayerBetted(userId string) bool {

	_, ok := gm.MapPlayerBet[userId]

	return ok
}

func (gm *GameTableDTX) OnBroadHistoryAnswers() {

	//需要触发清除历史开答的通道,需要通知玩家及大厅
	gm.BroadGameHistoryAnswers()

	gm.NotifyLobbyUpateHistoryAnswers()

	//	fmt.Printf("GameTableBacx  OnBroadHistoryAnswers  ")
}

// 通知商户单钱包服务，游戏开答了
func (gm *GameTableDTX) NotifyWalletServiceUpdateTrans() {

	sendPlayer := make(map[string]bool, 0)

	for _, pp := range gm.MapPlayerBet {

		_, ok := sendPlayer[pp.Player.UserId]
		if ok {
			log.Infof("之前的玩家已经发送过 pp.Player.UserId=%s\n", pp.Player.UserId)
			continue
		}

		optData := player.OpenAnswerData{
			DeskNo:  gm.GameInfo.DeskNo,
			RoundId: gm.GameInfo.RoundId,
		}

		pp.Player.NotifyWalletOpenAnswer(&optData)

		sendPlayer[pp.Player.UserId] = true

	}

}
