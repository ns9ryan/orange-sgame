package walletService

import (
	"errors"
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/node"
	originRpc "github.com/duanhf2012/origin/v2/rpc"
	sysrpc "github.com/duanhf2012/origin/v2/rpc"
	"github.com/duanhf2012/origin/v2/service"
	"github.com/duanhf2012/origin/v2/util/queue"
	"google.golang.org/protobuf/proto"
	"sync"
	"videogame/common/proto/rpc"
	"videogame/common/util"
	"videogame/common/util/typ"
	"videogame/service/interfacedef"
	"videogame/service/walletservice/msgrouter"
	"videogame/service/walletservice/player"
	factory "videogame/service/walletservice/playerfactory"
)

/***********************************************************************
钱包服务，对帐号发起加减款
**********************************************************************/

type OptType int32

const (
	WOpt_Add  OptType = 0 //加点操作
	WOpt_Dec  OptType = 1 //减点操作
	WOpt_Get  OptType = 2 //读取操作
	WOpt_Init OptType = 3 //初始化钱包

)

const (
	DBData_Balance = 0 //更新余额的数据
	DBData_Account = 1 //更新帐目的数据
	DBData_Tip     = 2 //更新打赏的数据
)

// 需求操作队列结构体
type WalletRequest struct {
	request   *rpc.WalletOptReq //操作钱包请求
	responder sysrpc.Responder  //回复的请求
}

// 数据库操作队列结构体
type WalletDBItem struct {
	UserId string //玩家ID

	DBDataType int //上面的DBData之一

	Balance uint64 //玩家余额

	Bet          uint32 //投注
	BetEffective uint32 //有效投注
	Water        uint32 //洗码量
	WinLost      int32  //胜负数据

	Tip uint32 //打赏的数据，2025.4.15号增加

}

func init() {
	node.Setup(&WalletService{})
}

type WalletService struct {
	service.Service

	goroutineNum uint32 //启动协程的个数

	channelNum   int //队列内数组个数
	bCloseStatus bool

	ReadingTableStatus bool //是否正在读取数据表的数据，用于结算时，通知钱包服

	mapPlayerMutex  sync.RWMutex              //专门用于读写安全的
	mapClientPlayer map[string]*player.Player //map[clientId]*Player

	playerFactory *factory.PlayerFactoryModule

	msgReceiver msgrouter.MsgReceiver
	msgSender   msgrouter.MsgSender

	channelWalletOpts []chan WalletRequest //操作钱包的通道队列

	dbRefreshQueue *queue.SyncQueue //数据库更新队列,这个内部是已经有互锁了

}

func (self *WalletService) OnInit() error {

	err := self.ReadCfg()

	if err != nil {
		return err
	}

	//初始化玩家容器
	self.mapClientPlayer = make(map[string]*player.Player, 2048)
	self.playerFactory = factory.NewPlayerFactoryModule()

	//处理数据库更新的队列
	self.dbRefreshQueue = queue.NewSyncQueue()

	self.channelWalletOpts = make([]chan WalletRequest, self.goroutineNum)

	//启动多个协程处理
	for i := uint32(0); i < self.goroutineNum; i++ {

		self.channelWalletOpts[i] = make(chan WalletRequest, self.channelNum)

		go self.ExecuteWalletOpt(self.channelWalletOpts[i])
	}

	//启动存储数据库协程
	go self.ExecuteWalletSaveToDB()

	_, err = self.AddModule(self.playerFactory)
	if err != nil {
		return err
	}

	_, err = self.AddModule(&self.msgSender)
	if err != nil {
		return err
	}

	_, err = self.AddModule(&self.msgReceiver)
	if err != nil {
		return err
	}

	self.RegRawRpc(util.RawRpcOnRecv, self.RpcOnRecvCallBack)
	self.RegRawRpc(util.RawRpcOnClose, self.RpcOnCloseCallBack)

	return nil
}

// ReadCfg 读取DB配置,配置在service.json中
func (self *WalletService) ReadCfg() error {

	iConfig := self.GetService().GetServiceCfg().(map[string]interface{})

	goroutineNum, ok := iConfig["GoroutineNum"]

	if ok == false {
		return fmt.Errorf("LobbyService config is error!")
	}

	self.goroutineNum, _ = typ.ConvertToNumber[uint32](goroutineNum)

	channelNum, ok := iConfig["ChannelNum"]

	if ok == false {
		return fmt.Errorf("LobbyService config is error!")
	}

	self.channelNum, _ = typ.ConvertToNumber[int](channelNum)

	return nil
}

// 接收路由发过来的投注
func (self *WalletService) RpcOnRecvCallBack(data []byte) {

	//解析转发过来的数据
	var rawInput rpc.RawInputArgs

	fmt.Printf("*****WalletService..... RpcOnRecvCallBack \n ")

	err := proto.Unmarshal(data, &rawInput)

	if err != nil {
		log.SError("RpcOnRecvCallBack Unmarshal is error:", err.Error())
		return
	}

	clientIdList := rawInput.GetClientIdList()

	if len(clientIdList) == 0 {
		//收消息只可能有一个clientid
		log.SError("RpcOnRecvCallBack receive client len[", len(clientIdList), "] > 1")
		return
	}

	if err != nil || len(clientIdList) != 1 {
		log.SError("parse message is error:", err.Error())
		return
	}

}

// 通知过来要客户端已关了
func (self *WalletService) RpcOnCloseCallBack(rawInput []byte) {

	//var rawInputArgs rpc.RawInputArgs
	//
	//err := proto.Unmarshal(rawInput, &rawInputArgs)
	//if err != nil {
	//	log.Error("Unmarshal fail", log.ErrorField("err", err))
	//	return
	//}
}

// 初始化玩家的钱包
func (self *WalletService) InitPlayerWallet(userId string) {

	pplayer := self.playerFactory.GetPlayer(userId)

	//InitPlayerWallet通过这个接口进来的Init可能是商户对多钱包玩家的充值/提现请求
	if pplayer == nil {
		pplayer = self.playerFactory.NewPlayer(userId, &self.msgSender, self, &self.msgReceiver)

		fmt.Printf("钱包服 NewPlayer userID:%v  \n", userId)

		if pplayer == nil {
			return
		}
	}

	pplayer.InitBalance()
}

func (self *WalletService) BalanceOpt(userId string, point float32, opt int32) (float64, error) {

	fmt.Println("WalletService   BalanceOpt---------------------")

	pplayer := self.playerFactory.GetPlayer(userId)

	//若没有找到玩家，需要去大厅服读取相关数据
	if pplayer == nil {
		return 0, errors.New("not find player !!!")
	}

	upoint := uint64(point * float32(pplayer.Digits))
	balance := uint64(0)

	//加点操作
	if opt == int32(WOpt_Add) {

		balance = pplayer.BalanceAdd(upoint)

	} else if opt == int32(WOpt_Dec) { //减点操作

		if pplayer.BalanceIsEnough(upoint) == false {
			return 0, errors.New("balance is insufficient")
		}

		balance = pplayer.BalanceDec(upoint)

	} else {

		return 0, errors.New("wallet opt no supper!!!")
	}

	//更新数据库及redis,缓存先更新，数据延后1秒左右更新
	pplayer.SaveWalletToRedis()

	//入队列，让排队存入数据库
	dbItem := WalletDBItem{
		UserId:       userId,
		DBDataType:   DBData_Balance,
		Balance:      balance,
		Bet:          0,
		BetEffective: 0,
		Water:        0,
		WinLost:      0,
		Tip:          0,
	}

	if self.dbRefreshQueue != nil {
		self.dbRefreshQueue.Add(&dbItem)
	}

	return pplayer.GetBalanceF(), nil
}

// 读取玩家余额
func (self *WalletService) GetBalance(userId string) (float64, error) {

	fmt.Println("WalletService   GetBalance---------------------")

	pplayer := self.playerFactory.GetPlayer(userId)

	//若没有找到玩家，需要去大厅服读取相关数据
	if pplayer == nil {
		return 0, errors.New("not find player !!!")
	}

	return pplayer.GetBalanceF(), nil

}

// BeCloseClient 被关闭
func (self *WalletService) BeCloseClient(clientId string) bool {

	if clientId == "" {
		log.Error("clientId is empty")
		return false
	}

	ret := self.closeClient(clientId)

	return ret
}

func (self *WalletService) OnStart() {
	//gs.AfterFunc(time.Second*2, gs.asyncPlayerListTimer)
}

func (self *WalletService) OnRetire() {

}

func (self *WalletService) OnRelease() {
}

func (self *WalletService) attachConn(cid string, gateId string, player *player.Player) {

	if player == nil {
		log.SError("betservice .ResetConn req is nil")
		return
	}

	self.mapPlayerMutex.Lock()

	if player.GetClientId() != "" {
		delete(self.mapClientPlayer, player.GetClientId())
	}

	self.mapClientPlayer[cid] = player

	self.mapPlayerMutex.Unlock()

	player.AttachConn(gateId, cid, "")

}

func (self *WalletService) GetGateNodeIdByClientId(clientId string) string {

	self.mapPlayerMutex.RLock()

	p := self.mapClientPlayer[clientId]

	self.mapPlayerMutex.RUnlock()

	if p == nil {
		return ""
	}

	return p.GetGateNodeId()
}

func (self *WalletService) GetClientIdByPlayerId(playerId string) string {

	p := self.playerFactory.GetPlayer(playerId)

	if p == nil {
		return ""
	}

	return p.GetClientId()
}

func (self *WalletService) delClientPlayer(clientId string) {

	self.mapPlayerMutex.Lock()
	defer self.mapPlayerMutex.Unlock()

	delete(self.mapClientPlayer, clientId)

}

func (self *WalletService) closeClient(clientId string) bool {

	//查找Player对象
	self.mapPlayerMutex.RLock()
	p, ok := self.mapClientPlayer[clientId]
	self.mapPlayerMutex.RUnlock()

	if ok == false {
		log.SError("clientId ", clientId, " not found")
		return false
	}

	//设置为离线状态
	p.SetOnline(false)

	self.delClientPlayer(clientId)

	return true

}

// CloseClient 主动关闭连接
func (self *WalletService) CloseClient(clientId string) {

	if clientId == "" {
		log.Error("clientId is empty")
		return
	}
	nodeId := self.GetGateNodeIdByClientId(clientId)

	var rawInputArgs rpc.RawInputArgs
	rawInputArgs.ClientIdList = []string{clientId}
	rawInputBytes, err := proto.Marshal(&rawInputArgs)
	if err != nil {
		log.Error("lobbyservice.CloseClient proto.Marshal ", log.ErrorField("err", err), log.String("clientId", clientId))
		return
	}
	err = self.RawGoNode(originRpc.RpcProcessorPB, nodeId, util.RawRpcCloseClient, util.GateService, rawInputBytes)

	if err != nil {
		log.SError("lobbyservice.CloseClient RawGoNode err:", err.Error())
	}

	self.closeClient(clientId)
}

func (self *WalletService) DestroyPlayer(playerId string) bool {

	p := self.playerFactory.GetPlayer(playerId)

	if p == nil {
		log.Error("cannot find player", log.String("playerId", playerId))
		return false
	}

	clientId := p.GetClientId()
	if clientId != util.IsEmpty {
		self.CloseClient(clientId)
	}
	self.playerFactory.ReleasePlayer(p)

	return true
}

func (self *WalletService) GetMsgSender() interfacedef.IMsgSender {
	return &self.msgSender
}

func (self *WalletService) getClientsByMapClients(mapClents map[string]*player.Player) []string {

	self.mapPlayerMutex.RLock()
	self.mapPlayerMutex.RUnlock()

	var clients []string
	if len(mapClents) == 0 {
		return clients
	}
	for client := range mapClents {
		clients = append(clients, client)
	}
	return clients
}

func (self *WalletService) GetClients() []string {
	return self.getClientsByMapClients(self.mapClientPlayer)
}

func (self *WalletService) GetMsgReceiver() interfacedef.IMsgReceiver {
	return &self.msgReceiver
}

func (self *WalletService) GetPlayerTimer() interfacedef.IPlayerTimer {
	return self.playerFactory
}
func (self *WalletService) GetClientPlayer(clientID string) interfacedef.IPlayer {

	self.mapPlayerMutex.RLock()
	self.mapPlayerMutex.RUnlock()

	return self.mapClientPlayer[clientID]
}

func (self *WalletService) GetAllClientId() []string {

	self.mapPlayerMutex.RLock()
	self.mapPlayerMutex.RUnlock()

	var clientIds []string
	for clientId, _ := range self.mapClientPlayer {
		clientIds = append(clientIds, clientId)
	}

	return clientIds
}
