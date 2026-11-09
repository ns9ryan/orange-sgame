package singleWalletService

import (
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/node"
	originRpc "github.com/duanhf2012/origin/v2/rpc"
	sysrpc "github.com/duanhf2012/origin/v2/rpc"
	"github.com/duanhf2012/origin/v2/service"
	"google.golang.org/protobuf/proto"
	"sync"
	"videogame/common/proto/rpc"
	"videogame/common/util"
	"videogame/common/util/typ"
	"videogame/service/interfacedef"
	"videogame/service/singlewalletservice/msgrouter"
	"videogame/service/singlewalletservice/player"
	factory "videogame/service/singlewalletservice/playerfactory"
	merchantapicall "videogame/service/singlewalletservice/merchantapicall"
	merchantapiimitate "videogame/service/singlewalletservice/merchantapiimitate"
	"github.com/duanhf2012/origin/v2/util/queue"
	"time"
)

/***********************************************************************
单钱包服务，向商户通过REST API发起查询余额和加减款操作
**********************************************************************/

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

// 数据库订单表操作队列结构体
type TransDBItem struct {
	UserId		string 	//玩家ID
	GameId		string 	//游戏SubCode
	DeskNo		int32 	//桌号
	RoundId		string 	//对局ID
	TransId		string 	//订单号
	Opt			int32	//加点还是减点
	DataType	int32	//订单详细操作类型
	Point 		float64 //订单金额
	Balance		float64 //操作后最新总金额
	Reason		string	//订单原因
	CreateTime	time.Time	//订单创建时间
	UpdateTime	time.Time	//订单更新时间
}

func init() {
	node.Setup(&SingleWalletService{})
}

type SingleWalletService struct {
	service.Service

	goroutineNum uint32 //启动协程的个数

	channelNum   int //队列内数组个数
	bCloseStatus bool

	mapPlayerMutex  sync.RWMutex              //专门用于读写安全的
	mapClientPlayer map[string]*player.Player //map[clientId]*Player

	playerFactory *factory.PlayerFactoryModule
	merchantApiCall *merchantapicall.MerchantApiModule

	msgReceiver msgrouter.MsgReceiver
	msgSender   msgrouter.MsgSender

	channelWalletOpts []chan WalletRequest //操作钱包的通道队列

	//用于监听在HTTP 9004端口上的商户单钱包服务的模拟
	MerchantApiModuleImitate merchantapiimitate.MerchantApiModuleImitate

	ReadingTableStatus bool //是否正在读取数据表的数据，用于结算时，通知钱包服
	dbRefreshQueue *queue.SyncQueue //数据库更新队列,这个内部是已经有互锁了
	dbTransQueue *queue.SyncQueue //数据库订单表更新队列,这个内部是已经有互锁了
}

func (self *SingleWalletService) OnInit() error {

	err := self.ReadCfg()

	if err != nil {
		return err
	}

	//初始化玩家容器
	self.mapClientPlayer = make(map[string]*player.Player, 2048)
	self.playerFactory = factory.NewPlayerFactoryModule()
	self.merchantApiCall = merchantapicall.NewMerchantApiModule()

	//处理数据库更新的队列
	self.dbRefreshQueue = queue.NewSyncQueue()

	//处理数据库订单表更新的队列
	self.dbTransQueue = queue.NewSyncQueue()

	self.channelWalletOpts = make([]chan WalletRequest, self.goroutineNum)

	//启动多个协程处理
	for i := uint32(0); i < self.goroutineNum; i++ {

		self.channelWalletOpts[i] = make(chan WalletRequest, self.channelNum)

		go self.ExecuteSingleWalletOpt(i, self.channelWalletOpts[i])
	}

	//启动存储数据库协程
	go self.ExecuteWalletSaveToDB()

	//启动订单存储数据库协程
	go self.ExecuteTransSaveToDB()

	_, err = self.AddModule(self.playerFactory)
	if err != nil {
		return err
	}

	_, err = self.AddModule(self.merchantApiCall)
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


	//用于监听在HTTP 9004端口上的商户单钱包服务的模拟
	self.MerchantApiModuleImitate = merchantapiimitate.MerchantApiModuleImitate{}
	self.AddModule(&self.MerchantApiModuleImitate)

	return nil
}

// ReadCfg 读取DB配置,配置在service.json中
func (self *SingleWalletService) ReadCfg() error {

	iConfig := self.GetService().GetServiceCfg().(map[string]interface{})

	goroutineNum, ok := iConfig["GoroutineNum"]

	fmt.Printf("==========ReadCfg get GoroutineNum=%d", goroutineNum)

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
func (self *SingleWalletService) RpcOnRecvCallBack(data []byte) {

	//解析转发过来的数据
	var rawInput rpc.RawInputArgs

	fmt.Printf("*****SingleWalletService..... RpcOnRecvCallBack \n ")

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
func (self *SingleWalletService) RpcOnCloseCallBack(rawInput []byte) {

	//var rawInputArgs rpc.RawInputArgs
	//
	//err := proto.Unmarshal(rawInput, &rawInputArgs)
	//if err != nil {
	//	log.Error("Unmarshal fail", log.ErrorField("err", err))
	//	return
	//}
}

// BeCloseClient 被关闭
func (self *SingleWalletService) BeCloseClient(clientId string) bool {

	if clientId == "" {
		log.Error("clientId is empty")
		return false
	}

	ret := self.closeClient(clientId)

	return ret
}

func (self *SingleWalletService) OnStart() {
	//gs.AfterFunc(time.Second*2, gs.asyncPlayerListTimer)
}

func (self *SingleWalletService) OnRetire() {

}

func (self *SingleWalletService) OnRelease() {
}

func (self *SingleWalletService) attachConn(cid string, gateId string, player *player.Player) {

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

func (self *SingleWalletService) GetGateNodeIdByClientId(clientId string) string {

	self.mapPlayerMutex.RLock()

	p := self.mapClientPlayer[clientId]

	self.mapPlayerMutex.RUnlock()

	if p == nil {
		return ""
	}

	return p.GetGateNodeId()
}

func (self *SingleWalletService) GetClientIdByPlayerId(playerId string) string {

	p := self.playerFactory.GetPlayer(playerId)

	if p == nil {
		return ""
	}

	return p.GetClientId()
}

func (self *SingleWalletService) delClientPlayer(clientId string) {

	self.mapPlayerMutex.Lock()
	defer self.mapPlayerMutex.Unlock()

	delete(self.mapClientPlayer, clientId)

}

func (self *SingleWalletService) closeClient(clientId string) bool {

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
func (self *SingleWalletService) CloseClient(clientId string) {

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

func (self *SingleWalletService) DestroyPlayer(playerId string) bool {

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

func (self *SingleWalletService) GetMsgSender() interfacedef.IMsgSender {
	return &self.msgSender
}

func (self *SingleWalletService) getClientsByMapClients(mapClents map[string]*player.Player) []string {

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

func (self *SingleWalletService) GetClients() []string {
	return self.getClientsByMapClients(self.mapClientPlayer)
}

func (self *SingleWalletService) GetMsgReceiver() interfacedef.IMsgReceiver {
	return &self.msgReceiver
}

func (self *SingleWalletService) GetPlayerTimer() interfacedef.IPlayerTimer {
	return self.playerFactory
}
func (self *SingleWalletService) GetClientPlayer(clientID string) interfacedef.IPlayer {

	self.mapPlayerMutex.RLock()
	self.mapPlayerMutex.RUnlock()

	return self.mapClientPlayer[clientID]
}

func (self *SingleWalletService) GetAllClientId() []string {

	self.mapPlayerMutex.RLock()
	self.mapPlayerMutex.RUnlock()

	var clientIds []string
	for clientId, _ := range self.mapClientPlayer {
		clientIds = append(clientIds, clientId)
	}

	return clientIds
}
