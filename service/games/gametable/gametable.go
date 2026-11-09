package gameTable

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/duanhf2012/origin/v2/event"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/service"
	"github.com/duanhf2012/origin/v2/util/queue"
	"sync"
	"time"
	"videogame/common/betnode"
	"videogame/common/eventid"
	"videogame/common/global/variable"
	models "videogame/common/model"
	"videogame/common/msg_struct"
	"videogame/common/proto/msg"
	"videogame/common/proto/rpc"
	redisId "videogame/common/redis"
	"videogame/common/util"
	"videogame/service/games/player"
	factory "videogame/service/games/playerfactory"
	"videogame/service/interfacedef"
)

const (
	NormalDealer = 1  //普通荷官
	SuperDealer  = 99 //超级荷官
)

type TableCfg struct {
	GameId  string //游戏名称,唯一
	DeskNo  int    //游戏桌号
	BetTime int    //总的投注时间
}

type TableInfo struct {
	GameId string //游戏名称,唯一

	GameSubCode string //游戏的二级code

	DeskNo int //游戏桌号
	Set    int //游戏的场次
	Round  int //游戏的局数

	RoundId string //每局的唯一ID

	GameStatus int //游戏状态

	TotalBetTime int       //总的投注时间
	StartBetTime time.Time //开始下注的时间

	IsAnswered bool   //是否已开奖
	VideoURL   string //视频源的地址

	NoBetRound uint //最大的未下注局数需要提供

	Dealer msg_struct.DealerInfo //本桌的荷官

}

// 检测是否在投注状态
func (table *TableInfo) IsBetting() bool {
	return table.GameStatus == msg_struct.GameStatusBet
}

func (table *TableInfo) GetGameId() string {
	return table.GameId
}

// 读取荷官的id
func (table *TableInfo) GetDealerId() string {
	return table.Dealer.DealerId
}

// 读取荷官的名字
func (table *TableInfo) GetDealerName() string {
	return table.Dealer.Name
}

type ITable interface {
	OnStartBet()            //下注开始
	OnStopBet()             //停止下注
	OnBroadHistoryAnswers() //开答
	OnSetStart()            //开场第一局
	OnInterMission()        //洗牌

	OnSleep() //休场.2025.4.2

}

type Table struct {
}

type GameTable struct {
	service.Module

	PlayersImp *factory.PlayerFactoryModule
	Gs         interfacedef.IGSService
	Table      ITable

	GameInfo TableInfo //游戏桌子的信息

	StopBetTimer *time.Timer // 停止下注的定时器

	mapPlayerMutex sync.RWMutex              //读写玩家的锁变量
	MapPlayer      map[string]*player.Player //玩家列表

	mapBetMutex sync.Mutex //投注的读写锁

	AnswerBar []int //历史开答表

	MerchantAccountQueue *queue.SyncQueue //更新商户帐目数据的队列

}

func (gm *GameTable) Init(objectFactoryModule *factory.PlayerFactoryModule, table ITable) {

	gm.PlayersImp = objectFactoryModule
	gm.Table = table

	//初始化队列
	gm.MerchantAccountQueue = queue.NewSyncQueue()
}

func (gm *GameTable) OnInit() error {

	gm.Gs = gm.GetService().(interfacedef.IGSService)

	//初始化玩家的数组
	gm.MapPlayer = make(map[string]*player.Player, 2048)

	return nil
}

// 读取DB表的中游戏配置参数
func (gm *GameTable) ReadDBConfig(gamecode string) bool {

	var talbeDB models.SGameTable

	dbItem, err := talbeDB.GetGameTable(gamecode)

	gm.GameInfo.NoBetRound = 4 //暂定4局

	if err == nil {

		gm.GameInfo.DeskNo = int(dbItem.DeskNo)
		gm.GameInfo.TotalBetTime = int(dbItem.BetInterval)

		gm.GameInfo.GameSubCode = dbItem.CategorySub
		gm.GameInfo.VideoURL = dbItem.StreamUrl
		gm.GameInfo.NoBetRound = dbItem.UnbetRounds

		return true
	}

	return false
}

// 从redis读取桌子信息
func (gm *GameTable) ReadTableInfoRedis() {

	redisKey := redisId.RedisId_GameInfo
	filedkey := gm.GameInfo.GetGameId() + ":TableInfo"

	bytes, err := variable.RedisStorage.HGet(redisKey, filedkey)

	if err != nil {
		return
	}

	var tableInfo TableInfo

	err = json.Unmarshal([]byte(bytes), &tableInfo)

	if err != nil {
		return
	}

	//若是下注状态，则需要改到停止下注
	if tableInfo.GameStatus == msg_struct.GameStatusBet {
		tableInfo.GameStatus = msg_struct.GameStatusStopBet
	}

	gm.GameInfo = tableInfo

}

// 把桌号信息存到redis
func (gm *GameTable) SaveTableInfoToRedis() {

	//var tableInfo TableInfo

	bytes, err := json.Marshal(gm.GameInfo)

	if err != nil {
		return
	}

	redisKey := redisId.RedisId_GameInfo
	filedkey := gm.GameInfo.GetGameId() + ":TableInfo"

	variable.RedisStorage.HSet(redisKey, filedkey, string(bytes))
}

func (gm *GameTable) GetPlayerNum() int {

	return len(gm.MapPlayer)
}

func (gm *GameTable) GetPlayer(clientId string) (*player.Player, bool) {

	gm.mapPlayerMutex.RLock()
	defer gm.mapPlayerMutex.RUnlock()

	pp, ok := gm.MapPlayer[clientId]

	return pp, ok
}

func (gm *GameTable) GetTableConfig(gameconfig *msg_struct.GameToCtrlGameConfig) {

	gameconfig.GameId = gm.GameInfo.GameId
	gameconfig.DeskNo = gm.GameInfo.DeskNo
	gameconfig.TotalBetTime = gm.GameInfo.TotalBetTime
	gameconfig.VideoUrl = gm.GameInfo.VideoURL

}

// 设定投注时间
func (gm *GameTable) SetTableBetTime(totalbetTime int) {

	if gm.GameInfo.TotalBetTime != totalbetTime {
		gm.GameInfo.TotalBetTime = totalbetTime
	}

}

func (gm *GameTable) SetTableVideoUrl(videoUrl string) {

	if gm.GameInfo.VideoURL != videoUrl {
		gm.GameInfo.VideoURL = videoUrl
	}
}

// PlayerJoin 玩家加入
func (gm *GameTable) PlayerJoin(player *player.Player, gateNodeId string, clientId string) bool {

	cid := player.GetClientId() //得到这个玩家之前的clid

	gm.mapPlayerMutex.Lock()

	if cid != "" {
		delete(gm.MapPlayer, cid)
	}

	gm.MapPlayer[clientId] = player

	gm.mapPlayerMutex.Unlock()

	player.AttachConn(gateNodeId, clientId, "", "", true)
	player.SetOnline(true)

	log.Infof("*******玩家成功加入游戏桌：gameId:%s,  user:%s   gateid:%s , clientId:%s \n", gm.GetGameId(), player.UserId, gateNodeId, clientId)

	return true
}

// PlayerLeave 玩家离开
func (gm *GameTable) PlayerLeave(player *player.Player) bool {

	cid := player.GetClientId()

	if cid != "" {

		player.SetOnline(false)

		gm.RemovePlayer(cid)

		log.Infof("玩家离开游戏桌了 gameId:%s ,userId:%s \n", gm.GetGameId(), player.UserId)

		return true
	}

	return false
}

// RemovePlayer 删除玩家帐号,只是从游戏桌去除，但内存中还存在
func (gm *GameTable) RemovePlayer(clientId string) bool {

	if clientId == "" {
		return false
	}

	gm.mapPlayerMutex.Lock()

	delete(gm.MapPlayer, clientId)

	gm.mapPlayerMutex.Unlock()

	return true
}

// 读取游戏ID
func (gm *GameTable) GetGameId() string {
	return gm.GameInfo.GameId
}

// 读取游戏桌号
func (gm *GameTable) GetDeskNo() int {
	return gm.GameInfo.DeskNo
}

// 检测桌子是否在下注中
func (gm *GameTable) IsBetting() bool {

	return gm.GameInfo.GameStatus == msg_struct.GameStatusBet
}

func (gm *GameTable) IsStopBet() bool {
	return gm.GameInfo.GameStatus == msg_struct.GameStatusStopBet
}

// 更新荷官的信息
func (gm *GameTable) UpdateDealer(dealer *msg_struct.DealerInfo) {

	//更新荷官的信息
	gm.GameInfo.Dealer = *dealer

	//需要触发更换荷管的事件,通知大厅，及玩家
	var rpcDealer rpc.DealerUpdateInfo

	rpcDealer.DealerId = dealer.DealerId
	rpcDealer.DealerName = dealer.Name
	rpcDealer.Avator = dealer.Avator

	rpcDealer.GameId = gm.GameInfo.GameId
	rpcDealer.DeskNo = uint32(gm.GameInfo.DeskNo)

	//通知大厅游戏开答了
	nodeId := util.GetGameServiceName(util.LobbyService)

	if nodeId == "" {
		return
	}

	////通知大厅，开答了，这个开答方法，每个游戏对大厅调用都不一样，因为开答结构体不一样，后面优化可以改用proto,就可以统一下
	rpcFun := util.FormatRpcMethod(util.LobbyService, "RPCUpdateDealerInfo")

	//通知大厅服务，更新荷管的消息
	gm.Gs.GoNode(nodeId, rpcFun, &rpcDealer)

}

// IsAnswered 是否开答了
func (gm *GameTable) IsAnswered() int {

	if gm.GameInfo.IsAnswered == true {
		return 1
	}

	return 0
}

// SendToPlayerCardInfo 发送玩家的帐号属性信息
func (gm *GameTable) SendToPlayerCardInfo(player *player.Player) error {

	var pyInfo msg.STOC_PlayerInfo

	pyInfo.UserId = player.UserId
	pyInfo.UserName = player.UserName

	pyInfo.Digits = int32(player.GetDigits())
	pyInfo.Credit = uint32(player.GetCredit())

	pyInfo.Currency = player.GetCurrency()

	player.SendMsg(msg.MsgType_PlayerInfo_Card, &pyInfo)

	return nil
}

// 发送给玩家游戏状态
func (gm *GameTable) SendToPlayerGameStatus(player *player.Player) error {

	var gsStatus msg_struct.GameStatusMsg

	gm.GetGameStatus(&gsStatus)

	var gameGs msg.STOC_GameStatus

	gameGs.GameId = gm.GetGameId()
	gameGs.DeskNo = int32(gsStatus.DeskNo)
	gameGs.Set = int32(gsStatus.Set)
	gameGs.Round = int32(gsStatus.Round)
	gameGs.GameSatus = int32(gsStatus.GameStatus)
	gameGs.IsAnswered = int32(gsStatus.IsAnswered)
	gameGs.BetTime = int32(gm.GetBetSeconds())

	sendnum := player.SendToPlayer(player.Id, msg.MsgType_GameTable_Status, &gameGs)

	fmt.Printf("Send GameStatuts To Player :%s  gameId:%s, Deskno:%d, set:%d round:%d gamestatus:%d\n", player.Id, gameGs.GameId, gameGs.DeskNo,
		gameGs.Set, gameGs.DeskNo, gameGs.GameSatus)

	if sendnum != 1 {
		return errors.New("send to player error")
	}

	log.Infof("Send GameStatuts To Player :%s  gameId:%s, Deskno:%d, set:%d round:%d gamestatus: %d  Answer:%d \n", player.Id, gameGs.GameId, gameGs.DeskNo,
		gameGs.Set, gameGs.DeskNo, gameGs.GameSatus, gameGs.IsAnswered)

	return nil
}

// 读取当前游戏状态
func (gm *GameTable) GetGameStatus(msg *msg_struct.GameStatusMsg) {

	msg.GameId = gm.GetGameId()
	msg.GameStatus = gm.GameInfo.GameStatus
	msg.Set = gm.GameInfo.Set
	msg.Round = gm.GameInfo.Round
	msg.DeskNo = gm.GameInfo.DeskNo

	msg.RoundId = gm.GameInfo.RoundId //新增加的局数id

	if gm.GameInfo.IsAnswered == false {
		msg.IsAnswered = 0
	} else {
		msg.IsAnswered = 1
	}

	if gm.GameInfo.GameStatus == msg_struct.GameStatusBet {
		msg.CurrentTime = gm.GetStartBetElapsedSeconds()
	} else {
		msg.CurrentTime = -1
	}
}

// 设置游戏状态
func (gm *GameTable) SetGameStatus(set int, round int, status int) bool {

	var bUpdate bool

	if gm.GameInfo.Set != set {
		return false
	}

	//若是开始下注，则检测局数是否正确
	//if status != msg_struct.GameStatusBet {
	//	if gm.GameInfo.Round != round {
	//		return false
	//	}
	//}

	fmt.Printf("GameTalbeUpdateStatus %d \n  ", status)

	bUpdate = false

	switch status {

	case msg_struct.GameStatusSleep: //休场
		bUpdate = gm.OnGameStatusSleep()
	case msg_struct.GameStatusInterMission: //洗牌
		bUpdate = gm.OnGameStatusInterMission()
	case msg_struct.GameStatusBet: //下注开始
		bUpdate = gm.OnGameStatusStartBet()
	}

	//存储游戏状态到redis服务器
	gm.SaveTableInfoToRedis()

	//通知各个玩家当前游戏状态
	return bUpdate
}

// 开始下注,需要检查的条件
func (gm *GameTable) OnGameStatusStartBet() bool {

	var bstartbet bool

	bstartbet = false

	if (gm.GameInfo.GameStatus == msg_struct.GameStatusStopBet && gm.GameInfo.IsAnswered == true) ||
		(gm.GameInfo.GameStatus == msg_struct.GameStatusSleep) || (gm.GameInfo.GameStatus == msg_struct.GameStatusInterMission) {
		bstartbet = true
	}

	if bstartbet == false {
		return false
	}

	//需要清除很多数据
	gm.GameInfo.StartBetTime = time.Now()
	gm.GameInfo.GameStatus = msg_struct.GameStatusBet
	gm.GameInfo.IsAnswered = false
	gm.GameInfo.Round++ //局数此时要加1

	//如果局数是第一局，场次需要更新
	if gm.GameInfo.Round == 1 {
		gm.GameInfo.Set++
		//需要清除开答
		gm.AnswerBar = gm.AnswerBar[:0]

		if gm.Table != nil {

			gm.Table.OnSetStart()
			gm.Table.OnBroadHistoryAnswers()
		}
	}

	gid := util.GenerateRoundId(gm.GameInfo.DeskNo, gm.GameInfo.Set, gm.GameInfo.Round)

	gm.GameInfo.RoundId = gid

	timedelay := time.Duration(gm.GameInfo.TotalBetTime)*time.Second + 1

	//启动定时器，倒计时为0，停止下注
	gm.StopBetTimer = time.AfterFunc(timedelay, gm.StopBetEvent)

	//如果处在结算时间，这边是否需要把场次重置,代码以后开发，写入ｒｅｄｉｓ

	//广播通知玩家端当前游戏的状态
	gm.BroadGameStatus()

	if gm.Table != nil {
		gm.Table.OnStartBet()

	}

	return true
}

// 休场
func (gm *GameTable) OnGameStatusSleep() bool {

	if (gm.GameInfo.GameStatus == msg_struct.GameStatusStopBet && gm.GameInfo.IsAnswered == true) ||
		gm.GameInfo.GameStatus == msg_struct.GameStatusInterMission {

		gm.GameInfo.GameStatus = msg_struct.GameStatusSleep
		gm.GameInfo.Round = 0
		gm.GameInfo.Set = 0
		gm.GameInfo.IsAnswered = false

		if gm.Table != nil {
			gm.Table.OnSleep()
		}

		//通知更新游戏状态
		gm.BroadGameStatus()

		return true
	}

	return false
}

// 洗牌
func (gm *GameTable) OnGameStatusInterMission() bool {

	var breset bool

	breset = false

	//以下几种状态可以洗牌
	if gm.GameInfo.GameStatus == msg_struct.GameStatusSleep {
		breset = true
	} else if gm.GameInfo.GameStatus == msg_struct.GameStatusStopBet && gm.GameInfo.IsAnswered == true {
		breset = true
	} else if gm.GameInfo.Dealer.Level == SuperDealer {
		breset = true
	}

	if breset == true {

		gm.GameInfo.GameStatus = msg_struct.GameStatusInterMission
		gm.GameInfo.IsAnswered = false
		gm.GameInfo.Round = 0
		//需要清除相关数据

		if gm.Table != nil {
			gm.Table.OnInterMission()
		}

		//通知更新游戏
		gm.BroadGameStatus()
	}

	fmt.Println(" OnGameStatusInterMission pair  %v", breset)

	return breset
}

// 广播游戏状态
func (gm *GameTable) BroadGameStatus() {

	var gsStatus msg_struct.GameStatusMsg

	gm.GetGameStatus(&gsStatus)

	gm.GetEventHandler().NotifyEvent(&event.Event{
		Type: eventid.Event_Broad_GameStatus,
		Data: gsStatus,
	})

}

func (gm *GameTable) GetAllClientId() []string {

	var clientIds []string

	for clientId, _ := range gm.MapPlayer {
		clientIds = append(clientIds, clientId)
	}

	return clientIds
}

// 读取从开始下注经过的秒数
func (gm *GameTable) GetStartBetElapsedSeconds() int {

	var ret int

	//只有在投注状态下，才可以得到实际时间
	nowTime := time.Now()
	offSecond := nowTime.Sub(gm.GameInfo.StartBetTime)
	ret = int(offSecond.Seconds())

	//	fmt.Printf("********NowTime:%s -    StartBetTime:%s  \n", nowTime.String(), gm.GameInfo.StartBetTime.String())
	return ret
}

// 读取还有多少秒可以下注
func (gm *GameTable) GetBetSeconds() int {

	var ret int

	//只有在投注状态下，才可以得到实际时间
	if gm.GameInfo.GameStatus == msg_struct.GameStatusBet {

		nowTime := time.Now()

		offSecond := nowTime.Sub(gm.GameInfo.StartBetTime)

		ret = gm.GameInfo.TotalBetTime - int(offSecond.Seconds())

		fmt.Println("####################GetBetSeconds :%d  , gm.GameInfo.StartBetTime:%d", offSecond, gm.GameInfo.StartBetTime)

	} else {
		ret = -1
	}

	return ret
}

// 读取此桌允许的未下注的局数
func (gm *GameTable) GetNoBetRound() uint {

	return gm.GameInfo.NoBetRound
}

func (gm *GameTable) StopBetEvent() {

	if gm.Table != nil {

		gm.Table.OnStopBet()

		//把游戏状态存到redis
		gm.SaveTableInfoToRedis()
	}

}

// 增加商户的帐目数据
func (gm *GameTable) MerchantAddAccount(merchantId string, account *player.CurAccount) {

	merchantAccount := player.MerchantAccount{
		MerchantId:       merchantId,
		LastBet:          account.LocalBet,
		LastBetEffective: account.LocalBetEffice,
		LastWater:        account.LocalWater,
		LastWin:          account.LocalWin,
	}

	//商户据入队列
	if gm.MerchantAccountQueue != nil {
		gm.MerchantAccountQueue.Add(&merchantAccount)
	}

}

// 启动协程池存储商户的帐目
func (gm *GameTable) DoMerchantAccountQueueToDB(queueId int64) {

	if gm.MerchantAccountQueue == nil {
		return
	}

	//启用协程队列存储
	gm.AsyncDoByQueue(queueId, func() bool {

		mlen := gm.MerchantAccountQueue.Len()

		if mlen == 0 {
			return false
		}

		mapMerchantItems := make(map[string]*player.MerchantAccount)

		//从队列取出每个代理的数据，然后累加
		for gm.MerchantAccountQueue.Len() > 0 {

			//可以读取队列数据了
			item := gm.MerchantAccountQueue.Pop()

			if item == nil {
				continue
			}

			dbItem := item.(*player.MerchantAccount)

			if dbItem == nil {
				continue
			}

			if dbItem.MerchantId == "" {
				continue
			}

			//若这个代理还没有置位，则填充这个位置
			if mapMerchantItems[dbItem.MerchantId] == nil {
				mapMerchantItems[dbItem.MerchantId] = dbItem
			} else {

				mapMerchantItems[dbItem.MerchantId].LastBet += dbItem.LastBet
				mapMerchantItems[dbItem.MerchantId].LastBetEffective += dbItem.LastBetEffective
				mapMerchantItems[dbItem.MerchantId].LastWater += dbItem.LastWater
				mapMerchantItems[dbItem.MerchantId].LastWin += dbItem.LastWin
			}

			//检测它是否有上级代理商
			merchantNode, ok := player.GetMerchantNode(dbItem.MerchantId)

			if ok == false {

				log.Errorf("无法读取内存商户：%s", dbItem.MerchantId)

				//后面需要到数据库再次读取
				continue
			}

			//检测是根商户吗
			if merchantNode.ParentId == "root" {
				continue
			}

			if mapMerchantItems[merchantNode.ParentId] == nil {
				mapMerchantItems[merchantNode.ParentId] = &player.MerchantAccount{MerchantId: merchantNode.ParentId}
			}

			mapMerchantItems[merchantNode.ParentId].LastBet += dbItem.LastBet
			mapMerchantItems[merchantNode.ParentId].LastBetEffective += dbItem.LastBetEffective
			mapMerchantItems[merchantNode.ParentId].LastWater += dbItem.LastWater
			mapMerchantItems[merchantNode.ParentId].LastWin += dbItem.LastWin
		}

		var merchantTable models.MerchantCurrency

		totalbet := uint(0)
		totalbeteff := uint(0)
		totalwater := uint(0)
		totalwin := 0

		//哪个代理商累帐目数据入表
		for _, item := range mapMerchantItems {

			if item == nil {
				continue
			}

			//需要改成钱包服
			err := merchantTable.UpdateMerchantAccount(item.MerchantId, item.LastBet, item.LastBetEffective, item.LastWater, item.LastWin)

			if err != nil {
				log.Errorf("")
			}

			totalbet += item.LastBet
			totalbeteff += item.LastBetEffective
			totalwater += item.LastWater
			totalwin += item.LastWin

			log.Infof("memrid:%s bet :%d water:%d win:%d", item.MerchantId, item.LastBet, item.LastWater, item.LastWin)
		}

		//累计值存入当前游戏的帐目统计表，以后开发，这个存在不同币种数据

		return false

	}, nil)

}

// 增加商户的打赏金额
func (gm *GameTable) MerchantAddTip(merchantId string, tipMount uint) {

	var merchantTable models.MerchantCurrency

	err := merchantTable.UpdateMerchantTip(merchantId, tipMount)

	if err != nil {

		log.Errorf("update merchantTip err:%v \n", err)
	}

}

// 玩家打赏
func (gm *GameTable) PlayerTip(pplayer *player.Player, tipBar *betnode.TipNode) (bool, int) {

	//userId := tipBar.UserId

	if tipBar.Digits == 0 {
		tipBar.Digits = 1
	}

	recordId := util.NewUnionId()

	//从钱包服读取当前余额，并检测投注额是否大于余额
	//optdec := float32(tipBar.TipPoint) / float32(tipBar.Digits)

	//这边需要检测是单钱包API调用或者本地钱包处理
	optData := player.OptData{
		UserId:    tipBar.UserId,
		Point:     uint(tipBar.TipPoint),
		Digits:    uint(tipBar.Digits),
		Opt:       1,
		GameId:    gm.GameInfo.GameId,
		DeskNo:    gm.GameInfo.DeskNo,
		TransType: 2, //打赏订单
		TransId:   recordId,
		RoundId:   gm.GameInfo.RoundId,
	}

	log.Infof("gameId:%s 玩家打赏 userId:%s   tipMoney:%d Digits:%d 开始处理！", tipBar.GameId, tipBar.UserId, tipBar.TipPoint, tipBar.Digits)

	if pplayer.DecWalletBalance(&optData) == true {

		//fmt.Println("打赏扣款出错！！！")
		log.Info("****玩家打赏,扣款成功: ")

		//打赏更新到钱包服记录
		if pplayer.AddClustorWalletTip(&optData) == false {

			log.Infof("玩家打赏更新数据到钱包服出错: ")

			return false, 0
		}

		//存打赏明细到数据库
		var tipTable models.STipRecord

		tipTable.UserName = tipBar.UserId
		tipTable.ParentId = tipBar.ParentId
		tipTable.MerchantId = tipBar.MerchantId
		tipTable.MerchantCode = tipBar.MerchatnCode
		tipTable.DeskNo = tipBar.DeskNo
		tipTable.GameId = tipBar.GameId
		tipTable.Digits = int(tipBar.Digits)
		tipTable.Currency = tipBar.Currency
		tipTable.TipAmount = int(tipBar.TipPoint)
		tipTable.TipTime = &tipBar.TipTime
		tipTable.DealerId = tipBar.DealerId

		tipTable.RecordId = recordId

		//tipTable.Type = gm.GameInfo.Type
		tipTable.SubType = gm.GameInfo.GameSubCode
		tipTable.RoundId = gm.GameInfo.RoundId

		//tipTable.RecordId,是否需要启动协程池
		err := tipTable.Create()

		if err != nil {
			log.Infof("gameId:%s 玩家打赏 userId:%s   tipMoney:%d Digits:%d 钱包服扣款成功! 但数据库没有存成功", tipBar.GameId, tipBar.UserId, tipBar.TipPoint, tipBar.Digits)

		} else {

			//商户的打赏也得累加
			log.Infof("玩家打赏 gameId:%s  userId:%s   tipMoney:%d Digits:%d 钱包服扣款成功!", tipBar.GameId, tipBar.UserId, tipBar.TipPoint, tipBar.Digits)
		}

	} else {
		//fmt.Println("钱包服扣款出错！！！")

		log.Errorf("玩家打赏 gameId:%s, userId:%s   tipMoney:%d Digits:%d 钱包服扣款出错！！", tipBar.GameId, tipBar.UserId, tipBar.TipPoint, tipBar.Digits)

		return false, int(msg.ErrCode_BetErr_ServerTipErr)
	}

	return true, 0

}

// 统计没有下注的会员
func (gm *GameTable) StatistNoBetRound() {

	//全部会员统一加一个
	for _, pplayer := range gm.MapPlayer {
		pplayer.NoBetRoundInc() //没有投注的局数加1
	}
}

// 通知玩家超出没有下注的局数，需要退出游戏
func (gm *GameTable) NotifyPlayerLeaveTable(maxround uint) {

	for _, pplayer := range gm.MapPlayer {

		if pplayer.IsOverNoBetRound(int(maxround)) == false {
			//若没有超出不需要处理
			continue
		}

		var noBetMsg msg.STOC_NoBetRoundOver

		noBetMsg.NoBetRound = int32(maxround)
		noBetMsg.GameId = gm.GetGameId()
		noBetMsg.Deskno = int32(gm.GetDeskNo())

		pplayer.SendMsg(msg.MsgType_GameTable_NoBetRound, &noBetMsg)

		//超出了，需要通知玩家要退出游戏
		log.Infof("player Id:%s not bet over gameId:%s ", pplayer.UserId, gm.GetGameId())

	}

}
