package longhuservice

import (
	"encoding/json"
	"fmt"
	"github.com/duanhf2012/origin/v2/event"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/network/processor"
	"github.com/duanhf2012/origin/v2/node"
	"github.com/duanhf2012/origin/v2/service"
	"github.com/duanhf2012/origin/v2/sysmodule/netmodule/wsmodule"
	"google.golang.org/protobuf/proto"
	"time"
	"videogame/common/eventid"
	"videogame/common/msg_struct"
	netModule "videogame/common/netmodule"
	"videogame/common/proto/rpc"
	"videogame/common/systemdata"
	"videogame/common/util"
	gameTableDTX "videogame/service/games/longhuservice/gametabledtx"
	"videogame/service/games/msgrouter"
	"videogame/service/games/player"
	factory "videogame/service/games/playerfactory"
	"videogame/service/interfacedef"
)

func init() {

	node.Setup(&LongHuService{})

}

type LongHuService struct {
	service.Service

	//游戏桌子信息
	GameTable gameTableDTX.GameTableDTX

	PlayersImp *factory.PlayerFactoryModule

	msgSender   msgrouter.MsgSender
	msgReceiver msgrouter.MsgReceiver

	//移植的
	netModule netModule.INetModule
	adminCtrl AdminDtx

	pbRawProcessor processor.PBRawProcessor
	rawPackInfo    processor.PBRawPackInfo
}

func (self *LongHuService) OnInit() error {

	//初始化协程池
	self.OpenConcurrent(10, 100, 10000)

	self.PlayersImp = factory.NewPlayerFactoryImp()

	self.RegRawRpc(util.RawRpcOnRecv, self.msgReceiver.RpcOnRecvCallBack)
	self.RegRawRpc(util.RawRpcOnClose, self.RpcOnCloseCallBack)

	//初始化控端服务
	self.adminCtrl.Init(&self.pbRawProcessor, self, &self.adminCtrl)

	self.AddModule(&self.adminCtrl)

	iConfig := self.GetService().GetServiceCfg()

	if iConfig == nil {
		return fmt.Errorf("%s config is error", self.GetService().GetName())
	}

	mapTcpCfg := iConfig.(map[string]interface{})
	_, wsOk := mapTcpCfg["WSCfg"]

	if wsOk == true {

		var wsModule wsmodule.WSModule

		//wsModule.SetProcessor(&self.pbRawProcessor)
		wsCfg, err := self.readWSCfg()

		if err != nil {
			return err
		}

		wsModule.Init(wsCfg, &self.pbRawProcessor)

		wsModule.SetMessageType(2)

		self.AddModule(&wsModule)

		self.adminCtrl.SetNetModule(&wsModule)

		self.netModule = &wsModule

	}

	_, err := self.AddModule(self.PlayersImp)
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

	//增加游戏桌子模块
	self.GameTable = gameTableDTX.GameTableDTX{}
	//此项一定要
	self.GameTable.Init(self.PlayersImp, &self.GameTable)

	self.AddModule(&self.GameTable)

	//初始化游戏桌的配置
	self.GameTable.InitGame()
	self.InitEvent()

	//初始化全局抽码列表
	systemdata.InitChipsConfig()

	//初始化商户代理
	player.InitMerchantGroups()

	//性能监控
	self.OpenProfiler()
	self.GetProfiler().SetOverTime(time.Millisecond * 100)
	self.GetProfiler().SetMaxOverTime(time.Second * 10)

	var timerId uint64
	//启动在线玩家列表
	self.SafeNewTicker(&timerId, time.Second*30, nil, self.ReportInfo)

	var timerId2 uint64
	self.SafeNewTicker(&timerId2, time.Second*3, nil, self.RegisterServiceToCoreEvent)

	return self.netModule.Start()

}
func (gate *LongHuService) readWSCfg() (*wsmodule.WSCfg, error) {

	//1解析配置
	iConfig := gate.GetService().GetServiceCfg()
	if iConfig == nil {
		return nil, fmt.Errorf("%s config is error", gate.GetService().GetName())
	}

	mapWSCfg := iConfig.(map[string]interface{})
	sWSCfg, ok := mapWSCfg["WSCfg"]
	if ok == false {
		return nil, fmt.Errorf("%s.WSCfg config is error", gate.GetService().GetName())
	}

	var wsCfg wsmodule.WSCfg
	byteWSCfg, err := json.Marshal(sWSCfg)

	if err != nil {
		return nil, fmt.Errorf("%s.WSCfg config is error:%s", gate.GetService().GetName(), err.Error())
	}

	err = json.Unmarshal(byteWSCfg, &wsCfg)

	if err != nil {
		return nil, fmt.Errorf("%s.WSCfg config is error:%s", gate.GetService().GetName(), err.Error())
	}

	return &wsCfg, nil
}

// 初始化事件
func (self *LongHuService) InitEvent() {

	//初始化事件
	self.RegEventReceiverFunc(eventid.Event_Broad_GameStatus, self.GetEventHandler(), self.OnServiceGameStatusEvent)
	self.RegEventReceiverFunc(eventid.Event_Broad_Answer, self.GetEventHandler(), self.OnServiceOpenAnswerEvent)
	self.RegEventReceiverFunc(eventid.Event_Broad_History_answers, self.GetEventHandler(), self.OnServiceGameHistoryAnswersEvent)
	self.RegEventReceiverFunc(eventid.Event_NotifyLobby_History_Answers, self.GetEventHandler(), self.OnServiceGameHistoryAnswersToLobbyEvent)

	self.RegEventReceiverFunc(eventid.Event_Player_Bet, self.GetEventHandler(), self.OnServiceTableBetEvent)
	self.RegEventReceiverFunc(eventid.Event_Player_CancelBet, self.GetEventHandler(), self.OnServiceTableCancelBetEvent)

	self.RegEventReceiverFunc(eventid.Event_Answer_Settle, self.GetEventHandler(), self.OnServiceAnswerSettleEvent)
	self.RegEventReceiverFunc(eventid.Event_Broad_PlayerCard, self.GetEventHandler(), self.OnBroadPlayerWalletEvent)

	self.RegEventReceiverFunc(eventid.Event_Player_TipMoney, self.GetEventHandler(), self.OnServiceTipEvent)

	//通知玩家在线人数
	self.RegEventReceiverFunc(eventid.Event_Update_OnlineNum, self.GetEventHandler(), self.OnServiceNotifyPlayerNumEvent)

	//聊天
	self.RegEventReceiverFunc(eventid.Event_Player_TalkChat, self.GetEventHandler(), self.OnServiceTalkChatEvent)

}

func (self *LongHuService) OnStart() {}

func (self *LongHuService) OnRetire() {}

func (self *LongHuService) OnRelease() {}

func (self *LongHuService) RpcOnCloseCallBack(rawInput []byte) {

	var rawInputArgs rpc.RawInputArgs

	err := proto.Unmarshal(rawInput, &rawInputArgs)
	if err != nil {
		log.Error("Unmarshal fail", log.ErrorField("err", err))
		return
	}

	for _, clientId := range rawInputArgs.ClientIdList {

		ret := self.BeCloseClient(clientId)

		if ret == true {
			fmt.Printf("龙虎游戏__________________玩家clientId:%s ,已经主动下线了。。。。 \n", clientId)
		} else {
			fmt.Printf("龙虎游戏__________________玩家clientId:%s ,已经主动下线了。。但不在此服。。 \n", clientId)
		}

	}

}

// BeCloseClient玩家退出被关闭
func (self *LongHuService) BeCloseClient(clientId string) bool {

	if clientId == "" {
		log.Error("clientId is empty")
		return false
	}

	player, ok := self.GameTable.GetPlayer(clientId)

	if ok == true {

		//检测没有投注的帐号才可以关闭
		if self.GameTable.IsPlayerBetted(player.Id) == false {

			self.closeClient(player.ClientId)

			return true
		}
	}

	return false

}

func (self *LongHuService) closeClient(clientId string) {

	if clientId == "" {
		return
	}

	player, ok := self.GameTable.GetPlayer(clientId)

	if ok == false {
		return
	}

	//查找Player对象
	self.GameTable.RemovePlayer(clientId)

	self.DestroyPlayer(player.Id)

}

func (self *LongHuService) DestroyPlayer(playerId string) bool {

	p := self.PlayersImp.GetPlayer(playerId)

	if p == nil {
		log.Error("cannot find player", log.String("playerId", playerId))
		return false
	}
	self.PlayersImp.ReleasePlayer(p)

	return true
}

// // CloseClient 主动关闭连接
func (self *LongHuService) CloseClient(clientId string) {

	if clientId == "" {
		log.Error("clientId is empty")
		return
	}

	//现阶这个不需要了
	//nodeId := self.GetGateNodeIdByClientId(clientId)
	//var rawInputArgs rpc.RawInputArgs
	//
	//rawInputArgs.ClientIdList = []string{clientId}
	//rawInputBytes, err := proto.Marshal(&rawInputArgs)
	//
	//if err != nil {
	//	log.Error("longhu1service.CloseClient proto.Marshal ", log.ErrorField("err", err), log.String("clientId", clientId))
	//	return
	//}
	//err = self.RawGoNode(originRpc.RpcProcessorPB, nodeId, util.RawRpcCloseClient, util.GateService, rawInputBytes)
	//
	//if err != nil {
	//	log.SWarn("longhu1service.CloseClient RawGoNode err:", err.Error())
	//}

	//需要改成通知大厅，支除这个玩家
	self.closeClient(clientId)

}

func (self *LongHuService) GetMsgSender() interfacedef.IMsgSender {
	return &self.msgSender
}

func (self *LongHuService) GetMsgReceiver() interfacedef.IMsgReceiver {
	return &self.msgReceiver
}

func (self *LongHuService) GetPlayerTimer() interfacedef.IPlayerTimer {
	return self.PlayersImp
}
func (self *LongHuService) GetClientPlayer(clientID string) interfacedef.IPlayer {
	return self.GameTable.MapPlayer[clientID]
}

func (self *LongHuService) GetGateNodeIdByClientId(clientId string) string {
	p := self.GameTable.MapPlayer[clientId]
	if p == nil {
		return ""
	}
	return p.GetGateNodeId()
}

func (self *LongHuService) GetClientIdByPlayerId(playerId string) string {
	p := self.PlayersImp.GetPlayer(playerId)
	if p == nil {
		return ""
	}
	return p.GetClientId()
}

func (self *LongHuService) GetAllClientId() []string {

	var clientIds []string
	for clientId, _ := range self.GameTable.MapPlayer {
		clientIds = append(clientIds, clientId)
	}
	return clientIds
}

func (self *LongHuService) GetObjectFactoryModule() *factory.PlayerFactoryModule {
	return self.PlayersImp
}

func (self *LongHuService) GetPlayerNum() int {

	return self.GameTable.GetPlayerNum()

}

// 更新玩家人数
func (self *LongHuService) UpdateOnlinePlayerNum() {

	//通知大厅，更新人数
	self.GetEventHandler().NotifyEvent(&event.Event{
		Type: eventid.Event_Update_OnlineNum,
		Data: nil,
	})
}

func (self *LongHuService) ReportInfo(tickId uint64, usned interface{}) {

	if self.PlayersImp != nil {
		self.PlayersImp.ReportPlayerInfo(self.GameTable.GetGameId())
	}

}

// 向CoreServer 注册本服务
func (self *LongHuService) RegisterServiceToCoreEvent(timerId uint64, usned interface{}) {

	var req rpc.ServiceItem

	req.NodeId = node.GetNodeId()
	req.ServiceName = self.GetName()
	req.RegType = msg_struct.SReg_Game
	req.GameId = self.GameTable.GetGameId()
	req.DeskNo = int32(self.GameTable.GetDeskNo())
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
