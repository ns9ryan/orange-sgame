package lobbyservice

import (
	"errors"
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/node"
	originRpc "github.com/duanhf2012/origin/v2/rpc"
	"github.com/duanhf2012/origin/v2/service"
	"google.golang.org/protobuf/proto"
	"sync"
	"time"
	"videogame/common/eventid"
	"videogame/common/msg_struct"
	"videogame/common/proto/msg"
	"videogame/common/proto/rpc"
	"videogame/common/util"
	"videogame/common/util/typ"
	"videogame/service/games/lobbyservice/gamelobby"
	"videogame/service/games/lobbyservice/msgrouter"
	"videogame/service/games/player"
	factory "videogame/service/games/playerfactory"
	"videogame/service/games/templatelimit"
	"videogame/service/interfacedef"
)

func init() {
	node.Setup(&LobbyService{})
}

type LobbyService struct {
	service.Service

	goroutineNum uint32
	channelNum   int
	bCloseStatus bool

	TickerId uint64 //定时器

	mapPlayerMutex  sync.RWMutex              //专门用于读写安全的
	mapClientPlayer map[string]*player.Player //map[clientId]*Player

	PlayersImp *factory.PlayerFactoryModule

	msgReceiver msgrouter.MsgReceiver
	msgSender   msgrouter.MsgSender

	GameLobby *gamelobby.GameLobby
}

func (self *LobbyService) OnInit() error {

	err := self.ReadCfg()

	if err != nil {
		return err
	}

	//初始化协程池
	self.OpenConcurrent(10, 100, 10000)

	//初始化玩家容器
	self.mapClientPlayer = make(map[string]*player.Player, 2048)

	self.PlayersImp = factory.NewPlayerFactoryImp()

	self.RegRawRpc(util.RawRpcOnRecv, self.msgReceiver.RpcOnRecvCallBack)
	self.RegRawRpc(util.RawRpcOnClose, self.RpcOnCloseCallBack)

	// 注册监听节点连接事件
	//self.RegDiscoverListener( self )

	_, err = self.AddModule(self.PlayersImp)
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

	self.GameLobby = gamelobby.NewGameLobby()

	//启动大厅模块
	_, err = self.AddModule(self.GameLobby)
	if err != nil {
		return err
	}

	//初始化商户代理,已移到main.go
	//player.InitMerchantGroups()

	//初始化ID数据
	templatelimit.InitTemplatePlans()

	//注册gate发过来的三条消息
	msgrouter.RegMsgHandler(msg.MsgType_Ping, self.OnMsgPing)
	msgrouter.RegMsgHandler(msg.MsgType_JoinGame, self.OnMsgJoinGame)
	msgrouter.RegMsgHandler(msg.MsgType_ExitGame, self.OnMsgLeaveGame)
	msgrouter.RegMsgHandler(msg.MsgType_EnterGame, self.OnMsgEnterGame)
	msgrouter.RegMsgHandler(msg.MsgType_GameTable_SuperSixSet, self.OnMsgSuperSixSet)

	//初始化事件器
	self.InitEvent()

	//在线检测玩家定时器
	self.SafeNewTicker(&self.TickerId, time.Second*10, nil, self.PlayersCheckIsOnlineTimer)

	var timerId2 uint64
	self.SafeNewTicker(&timerId2, time.Second*3, nil, self.RegisterServiceToCoreEvent)

	return nil

}

// 初始化事件
func (self *LobbyService) InitEvent() {

	//注册游戏状态事件改变
	self.RegEventReceiverFunc(eventid.Event_Broad_GameStatus, self.GetEventHandler(), self.OnServiceGameStatusEvent)

	//注册游戏开答事件
	self.RegEventReceiverFunc(eventid.Event_Broad_Answer, self.GetEventHandler(), self.OnServiceOpenAnswerEvent)

	self.RegEventReceiverFunc(eventid.Event_Broad_History_answers, self.GetEventHandler(), self.OnServiceGameHistoryAnswersToLobbyEvent)

	self.RegEventReceiverFunc(eventid.Event_Update_OnlineNum, self.GetEventHandler(), self.OnServiceUpdateGameOnlineEvent)

}

// ReadCfg 读取DB配置,配置在service.json中
func (self *LobbyService) ReadCfg() error {

	iConfig := self.GetService().GetServiceCfg().(map[string]interface{})

	goroutineNum, ok := iConfig["GoroutineNum"]

	if ok == false {
		return fmt.Errorf("LobbyService config is error!")
	}
	self.goroutineNum, _ = typ.ConvertToNumber[uint32](goroutineNum)

	//channelNum, ok := iConfig["ChannelNum"]
	//
	//if ok == false {
	//	return fmt.Errorf("LobbyService config is error!")
	//}
	//
	//self.channelNum, _ = typ.ConvertToNumber[int](channelNum)

	return nil
}

// 路由通知过来要客户端已关了
func (self *LobbyService) RpcOnCloseCallBack(rawInput []byte) {

	var rawInputArgs rpc.RawInputArgs

	err := proto.Unmarshal(rawInput, &rawInputArgs)

	if err != nil {
		//	log.Error("Unmarshal fail", log.ErrorAttr("err", err))
		return
	}

	//log.Infof("LobbyService--RpcOnCloseCallBack  ")

	//增加代码通知其它服务器玩家退出了
	//
	for _, clientId := range rawInputArgs.ClientIdList {

		self.BeCloseClient(clientId)

		log.Infof("LobbyService--RpcOnCloseCallBack---%s ", clientId)
	}

}

func (self *LobbyService) OnStart() {

}

func (self *LobbyService) OnRetire() {

}

func (self *LobbyService) OnRelease() {
}

// 存储玩家内数数据到服务的玩家数组中
func (self *LobbyService) attachConn(req *rpc.LoginToLobbyReq, player *player.Player) {

	if req == nil {
		log.SError("lobbyservice.ResetConn req is nil")
		return
	}

	//	fmt.Printf("LobbyService....attachConn   ")

	self.mapPlayerMutex.Lock()

	if player.GetClientId() != "" {
		delete(self.mapClientPlayer, player.GetClientId())
	}

	//重新建索引
	self.mapClientPlayer[req.ClientId] = player

	self.mapPlayerMutex.Unlock()

	player.AttachConn(req.GateNodeId, req.ClientId, req.SessionId, req.Ip, false)

}

func (self *LobbyService) GetGateNodeIdByClientId(clientId string) string {

	self.mapPlayerMutex.RLock()

	p := self.mapClientPlayer[clientId]

	self.mapPlayerMutex.RUnlock()

	if p == nil {
		return ""
	}

	return p.GetGateNodeId()
}

func (self *LobbyService) GetClientIdByPlayerId(playerId string) string {

	p := self.PlayersImp.GetPlayer(playerId)

	if p == nil {
		return ""
	}

	return p.GetClientId()
}

// 删除服务内存储玩家的数组对应项
func (self *LobbyService) delClientPlayer(clientId string) {

	if clientId == util.IsEmpty {
		return
	}

	self.mapPlayerMutex.Lock()
	defer self.mapPlayerMutex.Unlock()

	delete(self.mapClientPlayer, clientId)

}

// 路由服务通知过来此通道需要被动关闭
func (self *LobbyService) BeCloseClient(clientId string) {

	if clientId == "" {
		log.Error("clientId is empty")
		return
	}

	self.closeClient(clientId)
}

// 服务器主动关闭连接
func (self *LobbyService) CloseClient(clientId string) {

	if clientId == "" {
		log.Error("clientId is empty")
		return
	}
	nodeId := self.GetGateNodeIdByClientId(clientId)

	var rawInputArgs rpc.RawInputArgs

	rawInputArgs.ClientIdList = []string{clientId}
	rawInputBytes, err := proto.Marshal(&rawInputArgs)

	if err != nil {
		//log.Error("lobbyservice.CloseClient proto.Marshal ", log.ErrorAttr("err", err), log.String("clientId", clientId))
		return
	}
	//发回gate通道玩家被服务器关了
	err = self.RawGoNode(originRpc.RpcProcessorPB, nodeId, util.RawRpcCloseClient, util.GateService, rawInputBytes)

	if err != nil {
		log.SError("lobbyservice.CloseClient RawGoNode err:", err.Error())
	}

	self.closeClient(clientId)
}

func (self *LobbyService) closeClient(clientId string) {

	//查找Player对象
	self.mapPlayerMutex.RLock()
	p, ok := self.mapClientPlayer[clientId]
	self.mapPlayerMutex.RUnlock()

	if ok == false {
		log.SError("clientId ", clientId, " not found")
		return
	}

	//设置为离线状态
	p.SetOnline(false)
	p.WriteOnlineFlag(0)

	//移除redis上面的信息
	p.RemoveRedisUserInfo()

	//删除服务器对应的列
	self.delClientPlayer(clientId)

	//通知下注服，此玩家下线
	var req rpc.LoginToLobbyReq

	req.UserId = p.GetId()
	req.ClientId = p.ClientId
	req.GateNodeId = p.GetGateNodeId()

	//通知投注服，关闭玩家
	nodeId := util.GetGameServiceName(util.BetService) //请求目的游戏服务的nodeId

	if nodeId != "" {
		rpcFunc := util.FormatRpcMethod(util.BetService, "RPCLeaveBet")
		self.GoNode(nodeId, rpcFunc, &req)
	}

	fmt.Printf("玩家 clientId:%s 已掉线\n", clientId)

	//从内存缓存中清除此玩家数据，不再存在
	self.DestroyPlayer(p.Id)

}

// 从内存中，清除此玩家
func (self *LobbyService) DestroyPlayer(playerId string) bool {

	p := self.PlayersImp.GetPlayer(playerId)

	if p == nil {
		log.Error("cannot find player", log.String("playerId", playerId))
		return false
	}

	self.PlayersImp.ReleasePlayer(p)

	return true
}

func (self *LobbyService) GetMsgSender() interfacedef.IMsgSender {
	return &self.msgSender
}

func (self *LobbyService) GetMsgReceiver() interfacedef.IMsgReceiver {
	return &self.msgReceiver
}

func (self *LobbyService) GetPlayerTimer() interfacedef.IPlayerTimer {
	return self.PlayersImp
}
func (self *LobbyService) GetClientPlayer(clientID string) interfacedef.IPlayer {

	self.mapPlayerMutex.RLock()
	defer self.mapPlayerMutex.RUnlock()

	return self.mapClientPlayer[clientID]

}
func (self *LobbyService) GetAllClientId() []string {

	var clientIds []string

	self.mapPlayerMutex.RLock()

	for clientId, _ := range self.mapClientPlayer {
		clientIds = append(clientIds, clientId)
	}

	self.mapPlayerMutex.RUnlock()

	return clientIds
}

// 通知钱包服，初始化
func (self *LobbyService) InitPlayerWalletService(req *rpc.LoginToLobbyReq) error {

	//通知钱包服，初始化玩家钱包
	nodeId := util.GetGameServiceName(util.WalletService)

	if nodeId != "" {

		rpcFunc := util.FormatRpcMethod(util.WalletService, "RPCWalletInit")
		err := self.GoNode(nodeId, rpcFunc, req)

		if err != nil {
			return errors.New("cluster error !!!")
		}

	} else {
		return errors.New("cluster error !!!")
	}

	return nil
}

func (self *LobbyService) InitPlayerSingleWalletService(req *rpc.LoginToLobbyReq) error {

	//通知钱包服，初始化玩家钱包
	nodeId := util.GetGameServiceName(util.SingleWalletService)

	if nodeId != "" {

		rpcFunc := util.FormatRpcMethod(util.SingleWalletService, "RPCWalletInit")
		err := self.GoNode(nodeId, rpcFunc, req)

		if err != nil {
			return errors.New("cluster error !!!")
		}

	} else {
		return errors.New("cluster error !!!")
	}

	return nil
}

// 初始化投注服
func (self *LobbyService) JoinPlayerBetService(req *rpc.LoginToLobbyReq) error {

	//通知投注服，玩家进入游戏了
	nodeId := util.GetGameServiceName(util.BetService)

	if nodeId != "" {

		rpcFunc := util.FormatRpcMethod(util.BetService, "RPCJoinBet")

		err := self.GoNode(nodeId, rpcFunc, req)

		if err != nil {
			return errors.New("cluster error !!!")
		}

	} else {
		return errors.New("cluster error !!!")
	}

	return nil
}

// 实时打印在线玩家信息
func (self *LobbyService) PrintOnlinePlayers() {

	self.mapPlayerMutex.Lock()

	for k, v := range self.mapClientPlayer {

		fmt.Printf("玩家 userId:%s  clintId:%s 上线ip:%s , 上线时间:%s ", k, v.ClientId, v.LoginIp)
	}

	self.mapPlayerMutex.Unlock()
}

// 间隔检测玩家是否在线保持心跳
func (self *LobbyService) PlayersCheckIsOnlineTimer(utimer uint64, value interface{}) {

	if len(self.mapClientPlayer) == 0 {
		return
	}

	fmt.Printf("======================检测玩家在线情况=========================================================\n")

	self.mapPlayerMutex.RLock()

	ind := 0

	for _, player := range self.mapClientPlayer {

		//检测玩家是否下线
		if player.GetOnline() == false {

			fmt.Printf("玩家 userId:%s  clintId:%s   已下线！！！\n", player.Id, player.ClientId)
			continue
		}

		tms := player.LoginTime.String()

		fmt.Printf("玩家:%d userId:%s  clintId:%s 上线ip:%s , 上线时间:%s  \n", ind, player.Id, player.ClientId, player.LoginIp, tms)

		if player.PingIsTimeOut() == true {
			fmt.Printf("***玩家 userId:%s 已掉线了 \n", player.Id)

			player.SetOnline(false)
		}

		ind++
	}

	self.mapPlayerMutex.RUnlock()

	fmt.Printf("===============================================================================\n")

}

// 向CoreServer 注册本服务
func (self *LobbyService) RegisterServiceToCoreEvent(timerId uint64, usned interface{}) {

	var req rpc.ServiceItem

	req.NodeId = node.GetNodeId()
	req.ServiceName = self.GetName()
	req.RegType = msg_struct.SReg_Lobby
	req.GameId = "Lobby"
	req.DeskNo = 0
	//req.IpAddr =

	nodeId, rpcFun := util.GetServiceRpcMethodName(util.CoreService, "RPCRegisterService")

	if nodeId == "" {
		return
	}

	err := self.AsyncCallNode(nodeId, rpcFun, &req, func(ret *rpc.Int, err error) {

		if err != nil {
			log.Errorf("BacxService RegisterServiceToCoreEvent GoNode fail:%s", err.Error())
			return
		}

		if ret.Value == 1 {
			self.CancelTimerId(&timerId)
		}

	})

	if err != nil {
		fmt.Printf("BacxService RegisterServiceToCoreEvent GoNode fail:%s", err.Error())
	}

}
