package walletService

import (
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/node"
	originRpc "github.com/duanhf2012/origin/v2/rpc"
	"github.com/duanhf2012/origin/v2/service"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"sync"
	"time"
	"videogame/common/proto/rpc"
	"videogame/common/util"
	"videogame/common/util/typ"
	"videogame/service/games/player"
	"videogame/service/interfacedef"
	"videogame/service/walletservice/msgrouter"
	factory "videogame/service/walletservice/playerfactory"
)

/***********************************************************************
统计结算服，对玩家每天投注数据，洗码，胜负的统计入库
**********************************************************************/

// 结算入库用
type AccountTime struct {
	Hour   int
	Minute int
	Second int

	//	accountTime     time.Time

	NowAccountTime time.Time //现在要结算的时间

	LastAccountTime time.Time //最近结算的时间

}

func init() {
	node.Setup(&CollectService{})
}

type CollectService struct {
	service.Service

	goroutineNum uint32 //启动协程的个数

	channelNum   int //队列内数组个数
	bCloseStatus bool

	bAccounted bool //是否结算过

	IsAccountWorking bool //是否正在结算工作

	mapPlayerMutex  sync.RWMutex              //专门用于读写安全的
	mapClientPlayer map[string]*player.Player //map[clientId]*Player

	playerFactory *factory.PlayerFactoryModule

	msgReceiver msgrouter.MsgReceiver
	msgSender   msgrouter.MsgSender

	AccountTimer AccountTime //结算时间对象

	AccountTickerId uint64 //结算定时器ID

	CheckTime time.Time //检测时间

}

func (self *CollectService) OnInit() error {

	err := self.ReadCfg()

	if err != nil {
		return err
	}

	//初始化玩家容器
	self.mapClientPlayer = make(map[string]*player.Player, 2048)
	self.playerFactory = factory.NewPlayerFactoryModule()

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

	self.CheckTime = time.Now()

	player.ResetReadMerchant()
	player.InitMerchantGroups()

	//每秒调用一次
	self.SafeNewTicker(&self.AccountTickerId, time.Second, nil, self.AccountToTick)

	//性能监控
	self.OpenProfiler()
	self.GetProfiler().SetOverTime(time.Millisecond * 100)
	self.GetProfiler().SetMaxOverTime(time.Second * 10)

	return nil
}

// ReadCfg 读取DB配置,配置在service.json中
func (self *CollectService) ReadCfg() error {

	iConfig := self.GetService().GetServiceCfg().(map[string]interface{})

	Hour, ok1 := iConfig["AccountHour"]
	Minute, ok2 := iConfig["AccountMinute"]
	Second, ok3 := iConfig["AccountSecond"]

	if ok1 == false || ok2 == false || ok3 == false {
		return errors.New("read config error ")
	}

	self.AccountTimer.Hour, _ = typ.ConvertToNumber[int](Hour)
	self.AccountTimer.Minute, _ = typ.ConvertToNumber[int](Minute)
	self.AccountTimer.Second, _ = typ.ConvertToNumber[int](Second)

	//初始化最近结算 时间为当前时间
	self.AccountTimer.LastAccountTime = time.Now()

	return nil
}

// 接收路由发过来的投注
func (self *CollectService) RpcOnRecvCallBack(data []byte) {

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
func (self *CollectService) RpcOnCloseCallBack(rawInput []byte) {

	//var rawInputArgs rpc.RawInputArgs
	//
	//err := proto.Unmarshal(rawInput, &rawInputArgs)
	//if err != nil {
	//	log.Error("Unmarshal fail", log.ErrorField("err", err))
	//	return
	//}
}

// BeCloseClient 被关闭
func (self *CollectService) BeCloseClient(clientId string) bool {

	if clientId == "" {
		log.Error("clientId is empty")
		return false
	}

	ret := self.closeClient(clientId)

	return ret
}

func (self *CollectService) OnStart() {
	//gs.AfterFunc(time.Second*2, gs.asyncPlayerListTimer)
}

func (self *CollectService) OnRetire() {

}

func (self *CollectService) OnRelease() {
}

func (self *CollectService) attachConn(cid string, gateId string, player *player.Player) {

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

	//player.AttachConn(gateId, cid, "")

}

func (self *CollectService) GetGateNodeIdByClientId(clientId string) string {

	self.mapPlayerMutex.RLock()

	p := self.mapClientPlayer[clientId]

	self.mapPlayerMutex.RUnlock()

	if p == nil {
		return ""
	}

	return p.GetGateNodeId()
}

func (self *CollectService) GetClientIdByPlayerId(playerId string) string {

	p := self.playerFactory.GetPlayer(playerId)

	if p == nil {
		return ""
	}

	return p.GetClientId()
}

func (self *CollectService) delClientPlayer(clientId string) {

	self.mapPlayerMutex.Lock()
	defer self.mapPlayerMutex.Unlock()

	delete(self.mapClientPlayer, clientId)

}

func (self *CollectService) closeClient(clientId string) bool {

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
func (self *CollectService) CloseClient(clientId string) {

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

func (self *CollectService) DestroyPlayer(playerId string) bool {

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

func (self *CollectService) GetMsgSender() interfacedef.IMsgSender {
	return &self.msgSender
}

func (self *CollectService) getClientsByMapClients(mapClents map[string]*player.Player) []string {

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

func (self *CollectService) GetClients() []string {
	return self.getClientsByMapClients(self.mapClientPlayer)
}

func (self *CollectService) GetMsgReceiver() interfacedef.IMsgReceiver {
	return &self.msgReceiver
}

func (self *CollectService) GetPlayerTimer() interfacedef.IPlayerTimer {
	return self.playerFactory
}
func (self *CollectService) GetClientPlayer(clientID string) interfacedef.IPlayer {

	self.mapPlayerMutex.RLock()
	self.mapPlayerMutex.RUnlock()

	return self.mapClientPlayer[clientID]
}

func (self *CollectService) GetAllClientId() []string {

	self.mapPlayerMutex.RLock()
	self.mapPlayerMutex.RUnlock()

	var clientIds []string
	for clientId, _ := range self.mapClientPlayer {
		clientIds = append(clientIds, clientId)
	}

	return clientIds
}
