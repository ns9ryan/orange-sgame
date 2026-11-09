package bacxservice

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
	"videogame/common/proto/rpc"
	"videogame/common/util"
	"videogame/service/games/bacxservice/gameTableBacx"
	"videogame/service/games/msgrouter"
	factory "videogame/service/games/playerfactory"
	"videogame/service/interfacedef"
)

func init() {

	//node.Setup(&BacxService{})
	node.SetupTemplate[BacxService]()

}

type BacxService struct {
	service.Service

	//游戏桌子信息
	GameTable gameTableBacx.GameTableBacx

	//玩家列表
	PlayersImp *factory.PlayerFactoryModule

	msgSender   msgrouter.MsgSender
	msgReceiver msgrouter.MsgReceiver

	//控端接收器
	adminCtrl AdminBacx

	pbRawProcessor processor.PBRawProcessor
	rawPackInfo    processor.PBRawPackInfo
}

func (self *BacxService) OnInit() error {

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

		wsCfg, err := self.readWSCfg()

		if err != nil {
			return err
		}

		wsModule.SetMessageType(2)

		wsModule.Init(wsCfg, &self.pbRawProcessor)

		self.AddModule(&wsModule)

		self.adminCtrl.SetNetModule(&wsModule)

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
	self.GameTable = gameTableBacx.GameTableBacx{}

	self.GameTable.Init(self.PlayersImp, &self.GameTable)
	self.AddModule(&self.GameTable)

	self.GameTable.InitGame()
	self.InitEvent()

	//性能监控
	self.OpenProfiler()
	self.GetProfiler().SetOverTime(time.Millisecond * 100)
	self.GetProfiler().SetMaxOverTime(time.Second * 10)

	var timerId uint64
	//启动输出在线玩家列表
	self.SafeNewTicker(&timerId, time.Second*30, nil, self.ReportInfo)

	var timerId2 uint64
	self.SafeNewTicker(&timerId2, time.Second*3, nil, self.RegisterServiceToCoreEvent)

	return self.adminCtrl.Start()
}

func (gate *BacxService) readWSCfg() (*wsmodule.WSCfg, error) {

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
func (self *BacxService) InitEvent() {

	//初始化事件
	self.RegEventReceiverFunc(eventid.Event_Broad_GameStatus, self.GetEventHandler(), self.OnServiceGameStatusEvent)
	self.RegEventReceiverFunc(eventid.Event_Broad_Answer, self.GetEventHandler(), self.OnServiceOpenAnswerEvent)
	self.RegEventReceiverFunc(eventid.Event_Broad_History_answers, self.GetEventHandler(), self.OnServiceGameHistoryAnswersEvent)
	self.RegEventReceiverFunc(eventid.Event_NotifyLobby_History_Answers, self.GetEventHandler(), self.OnServiceGameHistoryAnswersToLobbyEvent)

	self.RegEventReceiverFunc(eventid.Event_Player_Bet, self.GetEventHandler(), self.OnServiceTableBetEvent)
	self.RegEventReceiverFunc(eventid.Event_Player_CancelBet, self.GetEventHandler(), self.OnServiceTableCancelBetEvent)

	self.RegEventReceiverFunc(eventid.Event_Answer_Settle, self.GetEventHandler(), self.OnServiceAnswerSettleEvent)

	//打赏事件
	self.RegEventReceiverFunc(eventid.Event_Player_TipMoney, self.GetEventHandler(), self.OnServiceTipEvent)

	//通知玩家在线人数
	self.RegEventReceiverFunc(eventid.Event_Update_OnlineNum, self.GetEventHandler(), self.OnServiceNotifyPlayerNumEvent)

	//玩家聊天事件
	self.RegEventReceiverFunc(eventid.Event_Player_TalkChat, self.GetEventHandler(), self.OnServiceTalkChatEvent)

}

func (self *BacxService) OnStart() {}

func (self *BacxService) OnRetire() {}

func (self *BacxService) OnRelease() {}

func (self *BacxService) RpcOnCloseCallBack(rawInput []byte) {

	var rawInputArgs rpc.RawInputArgs

	err := proto.Unmarshal(rawInput, &rawInputArgs)

	if err != nil {
		log.Error("Unmarshal fail", log.ErrorField("err", err))
		return
	}

	for _, clientId := range rawInputArgs.ClientIdList {

		ret := self.BeCloseClient(clientId)

		if ret == true {
			fmt.Printf("百家乐gameId:%s 玩家 clientId:%s 收到网关通知下线，关闭成功！   \n", self.GameTable.GetGameId(), clientId)
		} else {
			fmt.Printf("百家乐gameId:%s  clientId:%s    收到网关通知下线, 但此玩家没有在这个服！！！\n", self.GameTable.GetGameId(), clientId)
		}
	}
}

// BeCloseClient玩家退出被关闭
func (self *BacxService) BeCloseClient(clientId string) bool {

	if clientId == "" {
		log.Error("clientId is empty")
		return false
	}

	player, ok := self.GameTable.GetPlayer(clientId)

	if ok == true {

		//玩家要是有投注就不能注销
		if self.GameTable.IsPlayerBetted(player.Id) == false {

			self.closeClient(player.ClientId)

			return true
		}

		log.Infof("玩家有下注：%s , 不能注销", player.Id)
		//这边需要增加代码，加入注销的队列
	}

	return false
}

func (self *BacxService) closeClient(clientId string) {

	if clientId == "" {
		return
	}

	player, ok := self.GameTable.GetPlayer(clientId)

	if ok == false {
		return
	}

	log.Infof("玩家:%s 注销ok!", player.Id)

	//删除玩家对象
	self.GameTable.RemovePlayer(clientId)

	self.DestroyPlayer(player.Id)

}

func (self *BacxService) DestroyPlayer(playerId string) bool {

	p := self.PlayersImp.GetPlayer(playerId)

	if p == nil {
		log.Error("cannot find player", log.String("playerId", playerId))
		return false
	}
	self.PlayersImp.ReleasePlayer(p)

	return true
}

// CloseClient 主动关闭连接
func (self *BacxService) CloseClient(clientId string) {

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

func (self *BacxService) GetMsgSender() interfacedef.IMsgSender {
	return &self.msgSender
}

func (self *BacxService) GetMsgReceiver() interfacedef.IMsgReceiver {
	return &self.msgReceiver
}

func (self *BacxService) GetPlayerTimer() interfacedef.IPlayerTimer {
	return self.PlayersImp
}
func (self *BacxService) GetClientPlayer(clientID string) interfacedef.IPlayer {
	return self.GameTable.MapPlayer[clientID]
}

func (self *BacxService) GetGateNodeIdByClientId(clientId string) string {
	p := self.GameTable.MapPlayer[clientId]
	if p == nil {
		return ""
	}
	return p.GetGateNodeId()
}

func (self *BacxService) GetClientIdByPlayerId(playerId string) string {
	p := self.PlayersImp.GetPlayer(playerId)
	if p == nil {
		return ""
	}
	return p.GetClientId()
}

func (self *BacxService) GetAllClientId() []string {

	var clientIds []string
	for clientId, _ := range self.GameTable.MapPlayer {
		clientIds = append(clientIds, clientId)
	}
	return clientIds
}

func (self *BacxService) GetObjectFactoryModule() *factory.PlayerFactoryModule {
	return self.PlayersImp
}

func (self *BacxService) GetPlayerNum() int {

	return self.GameTable.GetPlayerNum()

}

// 更新玩家人数
func (self *BacxService) UpdateOnlinePlayerNum() {

	//通知大厅，更新人数
	self.GetEventHandler().NotifyEvent(&event.Event{
		Type: eventid.Event_Update_OnlineNum,
		Data: nil,
	})
}

func (self *BacxService) ReportInfo(tickId uint64, usned interface{}) {

	if self.PlayersImp != nil {
		self.PlayersImp.ReportPlayerInfo(self.GameTable.GetGameId())
	}

}

// 向CoreServer 注册本服务
func (self *BacxService) RegisterServiceToCoreEvent(timerId uint64, usned interface{}) {

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
