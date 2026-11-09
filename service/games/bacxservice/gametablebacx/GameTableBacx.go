package gameTableBacx

import (
	"encoding/json"
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

type GameTableBacx struct {
	*gameTable.Table //这个不是一定要

	gameTable.GameTable

	MapPlayerBet map[string]*PlayerBet //玩家的投注

	answerCard msg_struct.BacxAnswerCard //最近开奖的牌内容

	AnswerBar []msg_struct.BacxWinner //历史开答列表
}

// ReadConfig 读取游戏配置
func (gm *GameTableBacx) ReadConfig() {

	gsname := gm.Gs.GetName()

	if gsname == util.BacxServiceA {
		gm.GameInfo.GameId = msg_struct.GAMEID_BACXA
		gm.GameInfo.DeskNo = 1
	} else if gsname == util.BacxServiceB {
		gm.GameInfo.GameId = msg_struct.GAMEID_BACXB
		gm.GameInfo.DeskNo = 2
	} else if gsname == util.BacxServiceC {
		gm.GameInfo.GameId = msg_struct.GAMEID_BACXC
		gm.GameInfo.DeskNo = 3
	}

	gm.GameInfo.Set = 0
	gm.GameInfo.Round = 0
	gm.GameInfo.GameStatus = 0
	gm.GameInfo.TotalBetTime = 30
	gm.GameInfo.IsAnswered = false

	//读取数据库表配置参数
	if gm.ReadDBConfig(gm.GameInfo.GameId) == false {

		log.Errorf("read bacx dbconfig error！！！")
	}

	//检测投注时间是否为0
	if gm.GameInfo.TotalBetTime == 0 {
		gm.GameInfo.TotalBetTime = 10
	}

}

// 从redis读取历史开答表
func (gm *GameTableBacx) ReadHistoryAnswersRedis() {

	//先读取游戏桌状态
	redisKey := redisId.RedisId_GameAnswers + gm.GameInfo.GetGameId()

	for i := 0; i < gm.GameInfo.Round; i++ {

		filedkey := fmt.Sprintf("round:%03d", i+1)

		bytes, err := variable.RedisStorage.HGet(redisKey, filedkey)

		if err != nil {
			//出错处理，没有路子
			continue
		}

		var answerBar msg_struct.BacxWinner

		err = json.Unmarshal([]byte(bytes), &answerBar)

		if err != nil {
			//出错处理
			continue
		}

		gm.AnswerBar = append(gm.AnswerBar, answerBar)
	}

	fmt.Println("成功读取路子信息。。。。。len:", len(gm.AnswerBar))

}

// 存历史开答表到redis
func (gm *GameTableBacx) SaveAnswerToRedis(answerBar *msg_struct.GameOpenAnswerBacx) {

	bytes, err := json.Marshal(answerBar.Winner)

	if err != nil {
		return
	}

	redisKey := redisId.RedisId_GameAnswers + gm.GameInfo.GetGameId()

	filedkey := fmt.Sprintf("round:%03d", answerBar.Round)

	variable.RedisStorage.HSet(redisKey, filedkey, string(bytes))

}

func (gm *GameTableBacx) ClearRedisHistoryAnswers() {

	redisKey := redisId.RedisId_GameAnswers + gm.GameInfo.GetGameId()

	variable.RedisStorage.Del(redisKey)

}

func (gm *GameTableBacx) OnStartBet() {

	fmt.Printf("GameTableBacx OnStartBet\n")

}

func (gm *GameTableBacx) OnStopBet() {

	fmt.Println("StopBetEvent ")

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

	rpcFun := util.FormatRpcMethod(gm.Gs.GetName(), "RPCStopBetToCtrlService")

	//通知控端
	err := gm.Gs.Call(rpcFun, &gsStatus, &ret)

	if err != nil {
		return
	}

	//通知各个玩家,游戏状态停止下注
	gm.BroadGameStatus()
}

// 休场的回调
func (gm *GameTableBacx) OnSleep() {

	fmt.Printf("GameTableBacx OnSleep\n")

	gm.AnswerBar = gm.AnswerBar[:0]
	gm.MapPlayerBet = make(map[string]*PlayerBet)

	//清除redis上面的历史开答
	gm.ClearRedisHistoryAnswers()

}

// 每一局开始需要处理的
func (gm *GameTableBacx) OnSetStart() {

	gm.AnswerBar = gm.AnswerBar[:0]
	gm.MapPlayerBet = make(map[string]*PlayerBet)

	//清除redis上面的历史开答
	gm.ClearRedisHistoryAnswers()

}

func (gm *GameTableBacx) OnInterMission() {

}

func (gm *GameTableBacx) InitGame() {

	gm.MapPlayerBet = make(map[string]*PlayerBet)

	gm.ReadConfig()

	//从redis读取数据
	gm.ReadTableInfoRedis()
	gm.ReadHistoryAnswersRedis()

}

// 是否已经开答
func (gm *GameTableBacx) IsOpenAnswer() bool {

	return gm.GameInfo.IsAnswered
}

func (gm *GameTableBacx) OnBroadHistoryAnswers() {

	//需要触发清除历史开答的通道,需要通知玩家及大厅
	gm.BroadGameHistoryAnswers()

	gm.NotifyLobbyUpateHistoryAnswers()

	fmt.Printf("GameTableBacx  OnBroadHistoryAnswers  ")
}

// 开奖
func (gm *GameTableBacx) OpenAnswer(answer *msg_struct.GameOpenAnswerBacx) bool {

	//若已经开答过，则返回
	if gm.GameInfo.IsAnswered == true {
		return false
	}

	answer.GameId = gm.GetGameId()

	//开答结果入历史开答表
	gm.AnswerBar = append(gm.AnswerBar, answer.Winner)

	gm.GameInfo.IsAnswered = true

	//log.Infof("GameTableBacx  openanswer winner :%d   AnswerBar :%v \n", answer.Winner, gm.AnswerBar)

	//存游戏桌状态到redis
	gm.SaveTableInfoToRedis()

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
func (gm *GameTableBacx) ResetAnswerPeriodic() {
	gm.AnswerBar = gm.AnswerBar[:0]
}

// 读取历史开答表
func (gm *GameTableBacx) GetAnswerPeriodic(answerBar *msg_struct.GameBacXAnswersPeriodic) {

	answerBar.Set = gm.GameInfo.Set
	answerBar.Round = gm.GameInfo.Round
	answerBar.DeskNo = gm.GameInfo.DeskNo
	answerBar.GameStatus = gm.GameInfo.GameStatus
	answerBar.GameId = gm.GameInfo.GetGameId()

	answerBar.Winners = append(answerBar.Winners, gm.AnswerBar...)
	answerBar.WinnerNum = len(answerBar.Winners)

	//fmt.Printf("BACXSERIVE    GetAnswerPeriodic :%v \n ", answerBar.Winners)
}

// 广播游戏当前开答结果
func (gm *GameTableBacx) BroadGameAnswer(answer *msg_struct.GameOpenAnswerBacx) {

	//fmt.Printf("开答消息发送给玩家 :%v \n ", answer)

	//利用些框架的事件队列处理
	gm.GetEventHandler().NotifyEvent(&event.Event{
		Type: eventid.Event_Broad_Answer,
		Data: answer,
	})

}

// 广播游戏历史开答列表，利用事件通知功能
func (gm *GameTableBacx) BroadGameHistoryAnswers() {

	var answerBar msg_struct.GameBacXAnswersPeriodic

	gm.GetAnswerPeriodic(&answerBar)

	//fmt.Printf("Broad To Client AnswerPeriodic :%v", answerBar.Winners)

	gm.GetEventHandler().NotifyEvent(&event.Event{
		Type: eventid.Event_Broad_History_answers,
		Data: &answerBar,
	})

}

// 更新大厅更新历史开答
func (gm *GameTableBacx) NotifyLobbyUpateHistoryAnswers() {

	var answerBar msg_struct.GameBacXAnswersPeriodic

	gm.GetAnswerPeriodic(&answerBar)

	//	fmt.Printf("*************Notify To Lobby AnswerPeriodic :%v", answerBar.Winners)

	gm.GetEventHandler().NotifyEvent(&event.Event{
		Type: eventid.Event_NotifyLobby_History_Answers,
		Data: &answerBar,
	})

}

func (gm *GameTableBacx) SaveBetToRedis(betbar *betnode.BetNode) {

	keystr := betbar.UserId + ":" + betbar.OrkeyUid

	variable.RedisStorage.HSet(gm.GetGameId(), keystr, strconv.Itoa(int(betbar.BetPoint)))

}

// PlayerBet 玩家投注，返回投注总额，及是否投注成功
func (gm *GameTableBacx) PlayerBet(betbar *betnode.BetNode) (int, int) {

	bNew := false

	if betbar.Digits == 0 {
		betbar.Digits = 1
	}

	userId := betbar.UserId
	clientId := betbar.ClientId
	areadId := betbar.BetAreaId
	bet := betbar.BetPoint

	ppbet, ok := gm.MapPlayerBet[userId]

	//fmt.Printf("PlayerBet : userid:---%s  areaid:%d, bet:%v  ---  \n ", userId, areadId, bet)
	//log.Infof("PlayerBet : userid:---%s  areaid:%d, bet:%v  ", userId, areadId, bet)

	log.Infof("压注百家乐游戏[%s]:gameId:%s   userId:%s  betAreaId: %d , betMoney:%d Digits:%d  Orkey:%d 处理开始", betbar.OrkeyUid, gm.GameInfo.GameId, betbar.UserId, betbar.BetAreaId, betbar.BetPoint, betbar.Digits, betbar.OrderKey)

	//若玩家投注列表还没有，则需要生成
	if ok == false {

		pp, ok1 := gm.GetPlayer(clientId)

		if ok1 == false {

			//需要处理没有找到此玩家的代码
			log.Infof("压注百家乐游戏:gameId:%s  userId:%s 不成功，内存队列找不到此玩家！ ", gm.GameInfo.GameId, betbar.UserId)

			return 0, int(msg.ErrCode_BetErr_ServerClusterErr)
		}

		//产生新的投注玩家，后面需要改用内存池方案
		newpbet := PlayerBet{}
		newpbet.Init(pp)
		ppbet = &newpbet
		bNew = true

		//设置余额
		credit := pp.GetCredit()
		ppbet.SetBeforeBalance(credit)
	}

	//先检测此ID是否符合本游戏内
	if ppbet.BetIdIsRight(areadId) == false {

		log.Errorf("压注百家乐游戏gameId:%s  user:%s, betId:%d , betpoint:%d ,failed!!! , 区域不支持", gm.GameInfo.GameId, ppbet.Player.UserId, areadId, bet)
		return 0, int(msg.ErrCode_BetErr_ServerBetIdErr)
	}

	supersix := ppbet.Player.GetSuperSix()

	//检测在关闭supersix情况下，是否有投注了超级6相关区域
	if supersix == 0 {

		if ppbet.SuperSixAreaIsBet(areadId) == true {

			log.Errorf("压注百家乐游戏gameId:%s  user:%s, betId:%d , betpoint:%d ,failed!!! , supersix=0，不能压这", gm.GameInfo.GameId, ppbet.Player.UserId, areadId, bet)
			return 0, int(msg.ErrCode_BetErr_ServerBetIdErr)
		}

	} else { //有打开了超级6区域，不能压普通庄

		if ppbet.BankerAreaIsBet(areadId) == true {

			log.Errorf("压注百家乐游戏gameId:%s  user:%s, betId:%d , betpoint:%d ,failed!!! , supersix=1，不能压这", gm.GameInfo.GameId, ppbet.Player.UserId, areadId, bet)
			return 0, int(msg.ErrCode_BetErr_ServerBetIdErr)
		}

	}

	curPlayerBet := PlayerBet{Player: ppbet.Player, Digits: int(ppbet.Player.GetDigits())}

	//拷贝一份投注用于检测
	curPlayerBet.DuplicateBet(ppbet)

	//压注增加
	if curPlayerBet.Bet(areadId, bet) == false {
		//log.Errorf("压注百家乐游戏gameId:%s  user:%s, betId:%d , betpoint:%d ,failed!!! , 闲庄对压", gm.GameInfo.GameId, ppbet.Player.UserId, areadId, bet)
		return 0, int(msg.ErrCode_BetErr_ServerBetIdErr)
	}

	//检测是否少于低注
	if curPlayerBet.IsLessLowBet(areadId) == true {
		log.Errorf("压注百家乐游戏 gameId:%s  user:%s, betId:%d , betpoint:%d ,failed!!! betLessLowBet ", gm.GameInfo.GameId, ppbet.Player.UserId, areadId, bet)
		return 0, int(msg.ErrCode_BetErr_LessLowBetting)
	}

	//检测是否超出高注
	if curPlayerBet.IsOverHighBet(areadId) == true {
		log.Errorf("压注百家乐游戏gameId:%s   user:%s, betId:%d , betpoint:%d ,failed!!! OverHighBet ", gm.GameInfo.GameId, ppbet.Player.UserId, areadId, bet)
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
		GameId:    gm.GetGameId(),
		DeskNo:    gm.GetDeskNo(),
		TransType: 0, //下注订单
		TransId:   betbar.OrkeyUid,
		RoundId:   gm.GameInfo.RoundId,
	}

	if ppbet.Player.DecWalletBalance(&optData) == false {

		log.Errorf("玩家压注百家乐游戏:gameId:%s, userId:%s  betAreaId: %d , betMoney:%d Digits:%d  钱包服扣款出错！！！", gm.GameInfo.GameId,
			betbar.UserId, betbar.BetAreaId, betbar.BetPoint, betbar.Digits)

		return 0, int(msg.ErrCode_BetErr_ServerClusterErr)
	}

	//检测都通过了，可以拷贝最新投注
	ppbet.DuplicateBet(&curPlayerBet)

	//若是新产生的，需要加入投注列表
	if bNew == true {

		gm.MapPlayerBet[userId] = ppbet

		//不能操作supersix状态
		ppbet.Player.SetSuperSixOpt(false)

		//fmt.Println("新产生了")
	}

	//投注入队列
	ppbet.PushBet(*betbar)

	//计算总投注
	totalbet := int(ppbet.TotalBet())

	//把投注存到redis中
	gm.SaveBetToRedis(betbar)

	log.Infof("压注百家乐游戏[%s]:gameId:%s  userId:%s  betAreaId: %d , betMoney:%d Digits:%d orderKey:%d,   余额:%d    成功！！！",
		betbar.OrkeyUid, gm.GameInfo.GameId, betbar.UserId, betbar.BetAreaId, betbar.BetPoint, betbar.Digits, betbar.OrderKey, ppbet.Player.GetCredit())

	log.Infof("**********************************************")

	//增加一条投注明细
	return totalbet, 0
}

// PlayerBet 撤消玩家压注,返回投注项，投注额度， 错误码
func (gm *GameTableBacx) PlayerCancelBet(userId string) (betnode.BetNode, int, int) {

	log.Infof("******PlayerCancelBet :gameId:%s,  userid:---%s  ", gm.GameInfo.GameId, userId)

	ppbet, ok := gm.MapPlayerBet[userId]

	//若玩家投注列表还没有，则需要生成
	if ok == false {
		//需要处理没有找到此玩家的代码
		log.Infof("******PlayerCancelBet :gameId:%s,  userid:---%s  faile 找不到此投注玩家的索引  ", gm.GameInfo.GameId, userId)

		return betnode.BetNode{}, 0, int(msg.ErrCode_BetErr_ServerEmptyBet)
	}

	betNode, bOk := ppbet.CancelLastBet()

	if bOk == false {

		log.Infof("******PlayerCancelBet :gameId:%s,  userid:---%s  faile 没有可撤消的  ", gm.GameInfo.GameId, userId)

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
		GameId:    gm.GetGameId(),
		DeskNo:    gm.GetDeskNo(),
		TransType: 1, //取消下注订单
		TransId:   betNode.OrkeyUid,
		RoundId:   gm.GameInfo.RoundId,
	}

	if ppbet.Player.AddWalletBalance(&optData) == false {

		//fmt.Println("钱包服扣款出错！！！")
		log.Errorf("玩家撤消压注百家乐游戏: gameId:%s , userId:%s   钱包服扣款出错！！！", gm.GameInfo.GameId, ppbet.Player.UserId)

		return betnode.BetNode{}, 0, int(msg.ErrCode_BetErr_ServerClusterErr)
	}

	//若总投注额=0，则清掉这个数据
	if ppbet.TotalBet() == 0 {

		//恢复可操作supersix
		ppbet.Player.SetSuperSixOpt(true)

		delete(gm.MapPlayerBet, userId)
	}

	//回复给玩家这个投注额撤消成功了
	totalbet := int(int(ppbet.TotalBet()))

	log.Infof("******PlayerCancelBet :gameId:%s,  userid:---%s  成功！！ ", gm.GameInfo.GameId, userId)

	//回复给玩家这个投注额撤消成功了
	return betNode, totalbet, 0
}

// SettleWinLost 结算输嬴,投注明细入库，结算数据入库
func (gm *GameTableBacx) SettleWinLost(asn *msg_struct.GameOpenAnswerBacx) {

	betItemImp := CreateBetItemImp(msg_struct.GameTypeBacx, gm.GameInfo.GameSubCode) //投注明细入库用

	set := gm.GameInfo.Set
	round := gm.GameInfo.Round
	roundId := gm.GameInfo.RoundId

	//处理开答结算
	var settle AnswerSettle

	settle.Winner = asn.Winner.Winner
	settle.PPair = util.ConvertIntToBool(asn.Winner.PPair)
	settle.BPair = util.ConvertIntToBool(asn.Winner.BPair)

	settle.Big = asn.IsBig()
	settle.AnyPair = asn.IsAnyPair()
	settle.SuperSix = asn.IsSuperSix()
	settle.PerfectPair = asn.IsPerfectPair()

	settle.PlayerBonus = asn.IsPlayerBonus()
	settle.BankBonus = asn.IsBankerBonus()

	settle.PlayerPoint = asn.GetPlayerPoint()
	settle.BankPoint = asn.GetBankerPoint()
	settle.PlayerCardNum = asn.GetPlayerNum()
	settle.BankCardNum = asn.GetBankerNum()

	//把每个会员的没下注局数都加1
	gm.StatistNoBetRound()

	for _, ppbet := range gm.MapPlayerBet {

		log.Infof("gameId:%s 结算玩家：%s 开始!!", gm.GetGameId(), ppbet.Player.UserId)

		supersix := ppbet.Player.GetSuperSix()
		win := WinType(0)

		//若没有开启supersix玩法，则按普通的规则算
		if supersix == 0 {
			win = ppbet.WinSettle(&settle)
		} else {
			win = ppbet.WinSettleWithSuperSix(&settle)
		}

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
				GameId:    gm.GetGameId(),
				DeskNo:    gm.GameInfo.DeskNo,
				TransType: 3, //结算订单
				TransId:   fmt.Sprintf("%s_%s", gm.GameInfo.RoundId, ppbet.Player.UserId),
				RoundId:   gm.GameInfo.RoundId,
			}

			if ppbet.Player.AddWalletBalance(&optData) == false {
				//处理不成功
				log.Errorf("钱包服务加点不成功！！！ gameId:%s", gm.GameInfo.GameId)
			}

		}

		//2025.5.23 Add 读取玩家余额
		credit := ppbet.Player.GetCredit()
		ppbet.SetAfterBalance(credit)

		log.Infof("存储投注统计数，SaveBetCollectToDB ")
		//存储投注统计数 到数据库(不是明细，是投注统计)
		ppbet.SaveBetCollectToDB(gm, set, round, roundId, asn)

		//把每个玩家的投注明细列表复一份
		betItemImp.Copy(ppbet)

		log.Infof("存储帐目数据，UpdateAccountToDB ")
		//把玩家帐目更新到数据库
		ppbet.UpdateAccountToDB()

		//增加商户的帐目数据到队列
		gm.MerchantAddAccount(ppbet.Player.ParentId, &ppbet.CurAccount)

		//恢复可操作supersix状态
		ppbet.Player.SetSuperSixOpt(true)

		//连续没下注的局数清0
		ppbet.Player.ResetNoBetRound()

		log.Infof("gameId:%s 结算玩家：%s 完成!!", gm.GetGameId(), ppbet.Player.UserId)

	}

	if len(gm.MapPlayerBet) > 0 {

		//计算投注每个投注明细的胜负
		betItemImp.CalcWinLost(&settle, set, round, roundId, asn)

		log.Infof("投注明细入库 gameId:%s  SaveToDB ", gm.GetGameId())

		//投注明细入库,后面需要改成独立协程处理
		betItemImp.SaveToDB(gm)

		//log.Infof("商户帐目存入数据，，DoMerchantAccountQueueToDB ")

		//把商户帐目存入数据，人多时需要改成发放另一服务器专门存储
		gm.DoMerchantAccountQueueToDB(3)
	}

}

func (gm *GameTableBacx) ClearMapPlayerBet() {

	gm.MapPlayerBet = make(map[string]*PlayerBet)
}

// 检测玩家是否有投注
func (gm *GameTableBacx) IsPlayerBetted(userId string) bool {

	_, ok := gm.MapPlayerBet[userId]

	return ok

}

// 通知商户单钱包服务，游戏开答了
func (gm *GameTableBacx) NotifyWalletServiceUpdateTrans() {

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
