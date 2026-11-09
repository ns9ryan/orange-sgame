package gameTableSdie

import (
	"fmt"
	"github.com/duanhf2012/origin/v2/event"
	"github.com/duanhf2012/origin/v2/log"
	"strconv"
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

const ShowAnswerNum = 100 //要发给玩家端最后多少局的开答数

type GameTableSdie struct {
	*gameTable.Table //这个不是一定要

	gameTable.GameTable

	MapPlayerBet map[string]*PlayerBet //玩家的投注

	AnswerBar []int //历史开答列表
}

// ReadConfig 读取游戏配置
func (gm *GameTableSdie) ReadConfig() {

	gm.GameInfo.GameId = msg_struct.GAMEID_SDIE
	gm.GameInfo.DeskNo = 8
	gm.GameInfo.Set = 0
	gm.GameInfo.Round = 0
	gm.GameInfo.GameStatus = 0
	gm.GameInfo.IsAnswered = false
	gm.GameInfo.TotalBetTime = 30

	//读取数据表的参数
	gm.ReadDBConfig(msg_struct.GAMEID_SDIE)

	//检测投注时间是否为0
	if gm.GameInfo.TotalBetTime == 0 {
		gm.GameInfo.TotalBetTime = 10
	}

}

func (gm *GameTableSdie) OnStartBet() {

	fmt.Printf("GameTableSdie OnStartBet\n")
}

func (gm *GameTableSdie) OnStopBet() {

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

	rpcFun := util.FormatRpcMethod(util.SdieService, "RPCStopBetToCtrlService")

	//通知控端
	err := gm.Gs.Call(rpcFun, &gsStatus, &ret)

	if err != nil {

		log.Error("Error Call SdieService RPCStopBetToCtrlService ")

		return
	}

	//通知各个玩家,游戏状态停止下注
	gm.BroadGameStatus()
}
func (gm *GameTableSdie) OnBroadHistoryAnswers() {

	//需要触发清除历史开答的通道,需要通知玩家及大厅
	gm.BroadGameHistoryAnswers()

	gm.NotifyLobbyUpateHistoryAnswers()

	fmt.Printf("GameTableSdie  OnBroadHistoryAnswers  ")
}

// 每一局开始需要处理的
func (gm *GameTableSdie) OnSetStart() {

	gm.AnswerBar = gm.AnswerBar[:0]

	if len(gm.MapPlayerBet) > 0 {
		gm.MapPlayerBet = make(map[string]*PlayerBet)
	}

}

func (gm *GameTableSdie) OnInterMission() {

}

// 休场的回调
func (gm *GameTableSdie) OnSleep() {

	fmt.Printf("GameTableSicbo OnSleep\n")

	gm.AnswerBar = gm.AnswerBar[:0]
	gm.MapPlayerBet = make(map[string]*PlayerBet)

	//清除redis上面的历史开答
	gm.ClearRedisHistoryAnswers()
}

// 从redis读取历史开答表
func (gm *GameTableSdie) ReadHistoryAnswersRedis() {

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
func (gm *GameTableSdie) SaveAnswerToRedis(answerBar *msg_struct.SdieOpenAnswer) {

	strInt := strconv.Itoa(answerBar.Winner)

	redisKey := redisId.RedisId_GameAnswers + gm.GameInfo.GetGameId()

	filedkey := fmt.Sprintf("round:%03d", answerBar.Round)

	variable.RedisStorage.HSet(redisKey, filedkey, strInt)

}

func (gm *GameTableSdie) ClearRedisHistoryAnswers() {

	redisKey := redisId.RedisId_GameAnswers + gm.GameInfo.GetGameId()

	variable.RedisStorage.Del(redisKey)

}

func (gm *GameTableSdie) InitGame() {

	gm.MapPlayerBet = make(map[string]*PlayerBet)
	gm.ReadConfig()

	//从redis读取数据,恢复现场
	gm.ReadTableInfoRedis()
	gm.ReadHistoryAnswersRedis()

}

// 是否已经开答
func (gm *GameTableSdie) IsOpenAnswer() bool {

	return gm.GameInfo.IsAnswered
}

// 开奖
func (gm *GameTableSdie) OpenAnswer(answer *msg_struct.SdieOpenAnswer) bool {

	//若已经开答过，则返回
	if gm.GameInfo.IsAnswered == true {
		return false
	}

	answer.GameId = gm.GetGameId()

	//开答结果入历史开答表
	gm.AnswerBar = append(gm.AnswerBar, answer.Winner)

	gm.GameInfo.IsAnswered = true

	//存当前开答到redis
	gm.SaveAnswerToRedis(answer)

	//广播开奖消息给玩家
	gm.BroadGameAnswer(answer)

	//通知大厅开答历史变了
	gm.NotifyLobbyUpateHistoryAnswers()

	////开奖结算,用事件的方法处理
	gm.GetEventHandler().NotifyEvent(&event.Event{
		Type: eventid.Event_Answer_Settle,
		Data: answer,
	})

	return true
}

// 清空开答的
func (gm *GameTableSdie) ResetAnswerPeriodic() {
	gm.AnswerBar = gm.AnswerBar[:0]
}

// 读取历史开答表
func (gm *GameTableSdie) GetAnswerPeriodic(answerBar *msg_struct.SdieAnswersPeriodic) {

	answerBar.Set = gm.GameInfo.Set
	answerBar.Round = gm.GameInfo.Round
	answerBar.DeskNo = gm.GameInfo.DeskNo
	answerBar.GameStatus = gm.GameInfo.GameStatus

	answerBar.Winners = append(answerBar.Winners, gm.AnswerBar...)
	answerBar.WinnerNum = len(answerBar.Winners)

}

// 广播游戏当前开答结果
func (gm *GameTableSdie) BroadGameAnswer(answer *msg_struct.SdieOpenAnswer) {

	//fmt.Printf("开答消息发送给玩家 :%v \n ", answer)

	//利用些框架的事件队列处理
	gm.GetEventHandler().NotifyEvent(&event.Event{
		Type: eventid.Event_Broad_Answer,
		Data: *answer,
	})

}

// 广播游戏历史开答列表，利用事件通知功能
func (gm *GameTableSdie) BroadGameHistoryAnswers() {

	var answerBar msg_struct.SdieAnswersPeriodic

	gm.GetAnswerPeriodic(&answerBar)

	gm.GetEventHandler().NotifyEvent(&event.Event{
		Type: eventid.Event_Broad_History_answers,
		Data: answerBar,
	})

}

// 更新大厅更新历史开答
func (gm *GameTableSdie) NotifyLobbyUpateHistoryAnswers() {

	var answerBar msg_struct.SdieAnswersPeriodic

	gm.GetAnswerPeriodic(&answerBar)

	//fmt.Printf("*************Notify To Lobby AnswerPeriodic :%v", answerBar.Winners)

	gm.GetEventHandler().NotifyEvent(&event.Event{
		Type: eventid.Event_NotifyLobby_History_Answers,
		Data: answerBar,
	})

}

// PlayerBet 玩家投注，返回投注总额，及是否投注成功
func (gm *GameTableSdie) PlayerBet(betbar *betnode.BetNode) (int, int) {

	log.Infof("压注sdie游戏[%s]:gameId:%s   userId:%s  betAreaId: %d , betMoney:%d Digits:%d 处理开始", betbar.OrkeyUid, gm.GameInfo.GameId, betbar.UserId, betbar.BetAreaId, betbar.BetPoint, betbar.Digits)

	ppbet, ok := gm.MapPlayerBet[betbar.UserId]

	bNew := false

	if betbar.Digits == 0 {
		betbar.Digits = 1
	}

	areadId := betbar.BetAreaId
	bet := betbar.BetPoint
	userId := betbar.UserId

	//若玩家投注列表还没有，则需要生成
	if ok == false {

		pp, ok1 := gm.GetPlayer(betbar.ClientId)

		if ok1 == false {
			//需要处理没有找到此玩家的代码
			log.Infof("压注sdie游戏:gameId:%s  userId:%s 不成功，内存队列找不到此玩家！ ", gm.GameInfo.GameId, betbar.UserId)

			return 0, int(msg.ErrCode_BetErr_ServerClusterErr)
		}

		//产生新的投注玩家，后面需要改用内存池方案
		newpbet := PlayerBet{}
		newpbet.Init(pp)
		ppbet = &newpbet
		bNew = true

		credit := pp.GetCredit()
		ppbet.SetBeforeBalance(credit)

	}

	curPlayerBet := PlayerBet{Player: ppbet.Player, Digits: int(ppbet.Player.GetDigits())}

	//拷贝一份投注用于检测
	curPlayerBet.DuplicateBet(ppbet)

	//压注增加
	if curPlayerBet.Bet(areadId, bet) == false {

		log.Errorf("玩家压注sdie游戏 user:%s, betId:%d , betpoint:%d ,failed!!!", ppbet.Player.UserId, areadId, bet)

		return 0, int(msg.ErrCode_BetErr_ServerBetIdErr)
	}

	//检测是否少于低注
	if curPlayerBet.IsLessLowBet(areadId) == true {

		log.Errorf("玩家压ftan游戏: userId:%s  betAreaId: %d , betMoney:%d Digits:%d 少于低注",
			betbar.UserId, betbar.BetAreaId, betbar.BetPoint, betbar.Digits)

		return 0, int(msg.ErrCode_BetErr_LessLowBetting)
	}

	//检测是否超出高注
	if curPlayerBet.IsOverHighBet(areadId) == true {
		log.Errorf("玩家压注ftan游戏: userId:%s  betAreaId: %d , betMoney:%d Digits:%d 高于高注",
			betbar.UserId, betbar.BetAreaId, betbar.BetPoint, betbar.Digits)
		return 0, int(msg.ErrCode_BetErr_OverHighBetting)
	}

	//从钱包服读取当前余额，并检测投注额是否大于余额

	//从钱包服读取当前余额，并检测投注额是否大于余额
	//decPoint := float32(bet) / float32(betbar.Digits)

	//这边需要检测是单钱包API调用或者本地钱包处理
	optData := player.OptData{
		UserId:    ppbet.Player.UserId,
		Point:     uint(bet),
		Digits:    uint(betbar.Digits),
		Opt:       1,
		GameId:    gm.GameInfo.GameId,
		DeskNo:    gm.GameInfo.DeskNo,
		TransType: 0, //下注订单
		TransId:   betbar.OrkeyUid,
		RoundId:   gm.GameInfo.RoundId,
	}

	if ppbet.Player.DecWalletBalance(&optData) == false {

		//	fmt.Println("钱包服扣款出错！！！")

		log.Errorf("玩家压注色盘游戏: userId:%s  betAreaId: %d , betMoney:%d Digits:%d  钱包服扣款出错！！！", betbar.UserId, betbar.BetAreaId, betbar.BetPoint, betbar.Digits)

		return 0, int(msg.ErrCode_BetErr_ServerClusterErr)

	}

	//检测都通过了，可以拷贝最新投注
	ppbet.DuplicateBet(&curPlayerBet)

	//若是新产生的，需要加入投注列表
	if bNew == true {

		gm.MapPlayerBet[userId] = ppbet

		//	fmt.Printf("----------->新产生了MapPlayerBet:%v \n", gm.MapPlayerBet)
	}

	//投注入队列
	ppbet.PushBet(*betbar)

	totalbet := int(ppbet.TotalBet())

	log.Infof("压注sdie游戏[%s]:gameId:%s   userId:%s  betAreaId: %d , betMoney:%d Digits:%d 成功！", betbar.OrkeyUid, gm.GameInfo.GameId, betbar.UserId, betbar.BetAreaId, betbar.BetPoint, betbar.Digits)
	log.Infof("*********************************************************")

	return totalbet, 0
}

// PlayerBet 撤消玩家压注,返回投注项，投注额度， 错误码
func (gm *GameTableSdie) PlayerCancelBet(userId string) (betnode.BetNode, int, int) {

	log.Infof("******PlayerCancelBet :gameId:%s,  userid:---%s 开始处理 ", gm.GameInfo.GameId, userId)

	ppbet, ok := gm.MapPlayerBet[userId]

	//若玩家投注列表还没有，则需要生成
	if ok == false {
		//需要处理没有找到此玩家的代码
		log.Errorf("******PlayerCancelBet :gameId:%s,  userid:---%s  faile 找不到此投注玩家的索引  ", gm.GameInfo.GameId, userId)

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

		//	fmt.Println("钱包服扣款出错！！！")

		log.Errorf("玩家撤消压注色盘游戏: userId:%s   钱包服扣款出错！！！", ppbet.Player.UserId)

		return betnode.BetNode{}, 0, int(msg.ErrCode_BetErr_ServerClusterErr)

	}

	//若总投注额=0，则清掉这个数据
	if ppbet.TotalBet() == 0 {
		delete(gm.MapPlayerBet, userId)
	}

	//回复给玩家这个投注额撤消成功了
	totalbet := int(ppbet.TotalBet())

	log.Infof("******PlayerCancelBet :gameId:%s,  userid:---%s 处理完成 ", gm.GameInfo.GameId, userId)

	return betNode, totalbet, 0

}

// SettleWinLost 结算输嬴
func (gm *GameTableSdie) SettleWinLost(answer *msg_struct.SdieOpenAnswer) {

	betItemImp := CreateBetItemImp(msg_struct.GameTypeSide, gm.GameInfo.GameSubCode) //投注明细入库用

	set := gm.GameInfo.Set
	round := gm.GameInfo.Round
	roundId := gm.GameInfo.RoundId
	winner := answer.Winner

	for _, ppbet := range gm.MapPlayerBet {

		win := ppbet.WinLost(winner)

		//更新玩家余额到数据库
		if win > 0 {
			if ppbet.Digits == 0 {
				ppbet.Digits = 1
			}
			//	winf := float32(win) / float32(ppbet.Digits)

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

			if ppbet.Player.AddWalletBalance(&optData) == false {
				log.Errorf("gameId:%s ,user:%s AddWalletBalance failed", optData.GameId, optData.UserId)
			}

		}

		credit := ppbet.Player.GetCredit()
		ppbet.SetAfterBalance(credit)

		//存储投注列到数据库(不是明细，是投注统计)
		ppbet.SaveBetCollectToDBAsync(gm, set, round, roundId)

		//更新玩家帐目到数据库
		ppbet.UpdateAccountToDB()

		//把每个玩家的投注明细列表复一份
		betItemImp.Copy(ppbet)

		//增加商户的帐目数据到队列
		gm.MerchantAddAccount(ppbet.Player.ParentId, &ppbet.CurAccount)

	}

	//计算投注每个投注明细的胜负
	betItemImp.CalcWinLost(gm.GameInfo.Set, gm.GameInfo.Round, gm.GameInfo.RoundId, answer)

	//投注明细入库,后面需要改成独立协程处理
	betItemImp.SaveToDB(gm)

	//把商户帐目存入数据，人多时需要改成发放另一服务器专门存储
	gm.DoMerchantAccountQueueToDB(3)

}

func (gm *GameTableSdie) ClearMapPlayerBet() {

	gm.MapPlayerBet = make(map[string]*PlayerBet)
}

// 检测玩家是否有投注
func (gm *GameTableSdie) IsPlayerBetted(userId string) bool {

	_, ok := gm.MapPlayerBet[userId]

	return ok

}

// 通知商户单钱包服务，游戏开答了
func (gm *GameTableSdie) NotifyWalletServiceUpdateTrans() {
	
	sendPlayer := make(map[string]bool, 0)

	for _, pp := range gm.MapPlayerBet {

    	_, ok := sendPlayer[pp.Player.UserId]
    	if ok {
    		log.Infof("之前的玩家已经发送过 pp.Player.UserId=%s\n", pp.Player.UserId)
    		continue
        }

		optData := player.OpenAnswerData{
			DeskNo:    gm.GameInfo.DeskNo,
			RoundId:   gm.GameInfo.RoundId,
		}

		pp.Player.NotifyWalletOpenAnswer(&optData)

		sendPlayer[pp.Player.UserId] = true

	}

}
