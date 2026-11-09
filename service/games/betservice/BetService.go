package betService

import (
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/node"
	originRpc "github.com/duanhf2012/origin/v2/rpc"
	"github.com/duanhf2012/origin/v2/service"
	"google.golang.org/protobuf/proto"
	"sync"
	"videogame/common/proto/msg"
	"videogame/common/proto/rpc"
	"videogame/common/util"
	"videogame/common/util/typ"
	"videogame/service/games/betService/msgrouter"
	"videogame/service/games/player"
	factory "videogame/service/games/playerfactory"
	"videogame/service/interfacedef"
)

func init() {
	node.Setup(&BetService{})
}

type BetService struct {
	service.Service

	goroutineNum uint32 //启动协程的个数

	channelNum   int
	bCloseStatus bool

	mapPlayerMutex  sync.RWMutex              //专门用于读写安全的
	mapClientPlayer map[string]*player.Player //map[clientId]*Player

	PlayerImps *factory.PlayerFactoryModule

	msgReceiver msgrouter.MsgReceiver
	msgSender   msgrouter.MsgSender
}

func (self *BetService) OnInit() error {

	err := self.ReadCfg()

	if err != nil {
		return err
	}

	//初始化玩家容器
	self.mapClientPlayer = make(map[string]*player.Player, 2048)

	self.PlayerImps = factory.NewPlayerFactoryImp()

	self.RegRawRpc(util.RawRpcOnRecv, self.RpcOnRecvCallBack)
	self.RegRawRpc(util.RawRpcOnClose, self.RpcOnCloseCallBack)

	_, err = self.AddModule(self.PlayerImps)
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

	return nil

}

// ReadCfg 读取DB配置,配置在service.json中
func (self *BetService) ReadCfg() error {

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
func (self *BetService) RpcOnRecvCallBack(data []byte) {

	//解析转发过来的数据
	var rawInput rpc.RawInputArgs

	fmt.Printf("*****BetService..... RpcOnRecvCallBack \n ")

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

	msgType := msg.MsgType(rawInput.GetMsgType())

	//只处理投注消息，撤消投注消息
	if !(msgType == msg.MsgType_GameTable_Bet || msgType == msg.MsgType_GameTable_CancelBet) {

		//检测是否打赏，聊天消息
		if !(msgType == msg.MsgType_GameTable_TipMoney || msgType == msg.MsgType_GameTable_TalkChat) {
			log.SError("msgType is error:", msgType)
			return
		}
	}

	cliId := clientIdList[0]

	if msgType == msg.MsgType_GameTable_Bet {

		var betMsg msg.CTOS_BetGame

		err = proto.Unmarshal(rawInput.RawData, &betMsg)

		if err != nil {

			log.SError(err.Error())
			return
		}

		//调用投注方法
		self.OnBetToGame(cliId, &betMsg)

	} else if msgType == msg.MsgType_GameTable_CancelBet {

		var betMsg msg.CTOS_CancelBetGame

		err = proto.Unmarshal(rawInput.RawData, &betMsg)

		if err != nil {
			log.SError(err.Error())
			return
		}

		//调用撤消投注方法
		self.OnCancelBetToGame(cliId, &betMsg)

	} else if msgType == msg.MsgType_GameTable_TipMoney { //打赏金额

		var tipMsg msg.CTOS_TipMoney

		err = proto.Unmarshal(rawInput.RawData, &tipMsg)

		if err != nil {
			log.SError(err.Error())
			return
		}

		//调用撤消投注方法
		self.OnTipToGame(cliId, &tipMsg)

	} else if msgType == msg.MsgType_GameTable_TalkChat { //是否聊天消息

		var talkMsg msg.CTOS_TalkChat

		err = proto.Unmarshal(rawInput.RawData, &talkMsg)

		if err != nil {
			log.SError(err.Error())
			return
		}

		//调用撤消投注方法
		self.OnTalkToGame(cliId, &talkMsg)

	}

}

// 通知过来要客户端已关了
func (self *BetService) RpcOnCloseCallBack(rawInput []byte) {

	var rawInputArgs rpc.RawInputArgs

	err := proto.Unmarshal(rawInput, &rawInputArgs)
	if err != nil {
		log.Error("Unmarshal fail", log.ErrorField("err", err))
		return
	}

	for _, clientId := range rawInputArgs.ClientIdList {

		del := self.BeCloseClient(clientId)

		if del == true {
			fmt.Printf("投注服_______________玩家clientId:%s ,已经主动下线了。。。。 \n", clientId)
		} else {
			fmt.Printf("投注服_______________玩家clientId:%s ,已经主动下线了。。。但没有找到。 \n", clientId)
		}

	}

}

// BeCloseClient 被关闭
func (self *BetService) BeCloseClient(clientId string) bool {

	if clientId == "" {
		log.Error("clientId is empty")
		return false
	}

	ret := self.closeClient(clientId)

	return ret
}

func (self *BetService) OnStart() {
	//gs.AfterFunc(time.Second*2, gs.asyncPlayerListTimer)
}

func (self *BetService) OnRetire() {

}

func (self *BetService) OnRelease() {
}

func (self *BetService) attachConn(cid string, gateId string, player *player.Player) {

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

	player.AttachConn(gateId, cid, "", gateId, false)

}

func (self *BetService) GetGateNodeIdByClientId(clientId string) string {

	self.mapPlayerMutex.RLock()

	p := self.mapClientPlayer[clientId]

	self.mapPlayerMutex.RUnlock()

	if p == nil {
		return ""
	}

	return p.GetGateNodeId()
}

func (self *BetService) GetClientIdByPlayerId(playerId string) string {

	p := self.PlayerImps.GetPlayer(playerId)

	if p == nil {
		return ""
	}

	return p.GetClientId()
}

func (self *BetService) delClientPlayer(clientId string) {

	self.mapPlayerMutex.Lock()
	defer self.mapPlayerMutex.Unlock()

	delete(self.mapClientPlayer, clientId)

}

func (self *BetService) closeClient(clientId string) bool {

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
func (self *BetService) CloseClient(clientId string) {

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

func (self *BetService) DestroyPlayer(playerId string) bool {

	p := self.PlayerImps.GetPlayer(playerId)

	if p == nil {
		log.Error("cannot find player", log.String("playerId", playerId))
		return false
	}

	clientId := p.GetClientId()
	if clientId != util.IsEmpty {
		self.CloseClient(clientId)
	}
	self.PlayerImps.ReleasePlayer(p)

	return true
}

func (self *BetService) GetMsgSender() interfacedef.IMsgSender {
	return &self.msgSender
}

func (self *BetService) getClientsByMapClients(mapClents map[string]*player.Player) []string {

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

func (self *BetService) GetClients() []string {
	return self.getClientsByMapClients(self.mapClientPlayer)
}

func (self *BetService) GetMsgReceiver() interfacedef.IMsgReceiver {
	return &self.msgReceiver
}

func (self *BetService) GetPlayerTimer() interfacedef.IPlayerTimer {
	return self.PlayerImps
}
func (self *BetService) GetClientPlayer(clientID string) interfacedef.IPlayer {

	self.mapPlayerMutex.RLock()
	self.mapPlayerMutex.RUnlock()

	return self.mapClientPlayer[clientID]
}

func (self *BetService) GetAllClientId() []string {

	self.mapPlayerMutex.RLock()
	self.mapPlayerMutex.RUnlock()

	var clientIds []string
	for clientId, _ := range self.mapClientPlayer {
		clientIds = append(clientIds, clientId)
	}

	return clientIds
}

func (self *BetService) GetObjectFactoryModule() *factory.PlayerFactoryModule {
	return self.PlayerImps
}
