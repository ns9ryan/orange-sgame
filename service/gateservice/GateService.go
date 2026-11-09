package gateservice

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/network/processor"
	"github.com/duanhf2012/origin/v2/node"
	"github.com/duanhf2012/origin/v2/service"
	"github.com/duanhf2012/origin/v2/sysmodule/netmodule/wsmodule"
	"github.com/duanhf2012/origin/v2/util/queue"
	"google.golang.org/protobuf/proto"
	"time"
	netModule "videogame/common/netmodule"
	"videogame/common/proto/rpc"
	"videogame/common/util"
)

func init() {
	node.Setup(&GateService{})
}

type GateService struct {
	service.Service

	netModule      netModule.INetModule
	pbRawProcessor processor.PBRawProcessor
	msgRouter      MsgRouter
	rawPackInfo    processor.PBRawPackInfo

	bStopStatus bool //是否停止工作

	removeCidQueue *queue.SyncQueue //无效通道存储

}

func (gate *GateService) OnInit() error {

	gate.msgRouter.Init(&gate.pbRawProcessor)
	gate.AddModule(&gate.msgRouter)

	//用于存储无效通道的列队，定期清除
	gate.removeCidQueue = queue.NewSyncQueue()

	var wsCfg *wsmodule.WSCfg

	//读取全局的gate的配置
	wsCfg, err := gate.getWSCfg()

	if err == nil && wsCfg != nil {

		var wsModule wsmodule.WSModule

		wsModule.Init(wsCfg, &gate.pbRawProcessor)
		wsModule.SetMessageType(2)

		gate.AddModule(&wsModule)
		gate.msgRouter.SetNetModule(&wsModule)
		gate.netModule = &wsModule

	} else {
		return errors.New("WSCfg  are not configured")
	}

	// 注册监听节点连接事件
	gate.RegDiscoverListener(gate)

	gate.RegRawRpc(util.RawRpcMsgDispatch, gate.RawRpcDispatch)
	gate.RegRawRpc(util.RawRpcCloseClient, gate.RawCloseClient)

	go gate.ExecuteRemoveCidQueueLoop()

	return gate.netModule.Start()

}

func (gate *GateService) getWSCfg() (*wsmodule.WSCfg, error) {

	globalCfg, err := util.GetGlobalGateCfgByNodeId(node.GetNodeId())

	if err != nil || globalCfg == nil {
		return nil, errors.New("WSCfg are not configured")
	}

	//获取全局中的证书地址
	var keyFile string
	var certFile string

	pemCfg, err := util.GetGlobalPemFile()
	if err == nil && pemCfg != nil {
		keyFile = pemCfg.KeyFile
		certFile = pemCfg.CertFile
	}

	wsCfg := &wsmodule.WSCfg{
		ListenAddr:      globalCfg.ListenAddr,
		MaxConnNum:      globalCfg.MaxConnNum,
		PendingWriteNum: globalCfg.PendingWriteNum,
		MaxMsgLen:       globalCfg.MaxMsgLen,
		LittleEndian:    globalCfg.LittleEndian,
		KeyFile:         keyFile,
		CertFile:        certFile,
	}

	return wsCfg, nil
}

func (gate *GateService) readWSCfg() (*wsmodule.WSCfg, error) {

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

func (gate *GateService) SendMsg(clientId string, msgType uint16, rawMsg []byte) error {

	gate.rawPackInfo.SetPackInfo(msgType, rawMsg)
	bytes, err := gate.pbRawProcessor.Marshal(clientId, &gate.rawPackInfo)

	if err != nil {
		return err
	}

	err = gate.netModule.SendRawMsg(clientId, bytes)

	if err != nil {
		log.Debugf("SendMsg fail ", log.ErrorField("err", err), log.String("clientId", clientId))
	}

	return err
}

func (gate *GateService) RawRpcDispatch(rawInput []byte) {

	var rawInputArgs rpc.RawInputArgs

	err := proto.Unmarshal(rawInput, &rawInputArgs)

	if err != nil {
		log.SError("msg is error:%s", err.Error())
		return
	}

	for _, clientId := range rawInputArgs.ClientIdList {

		err = gate.SendMsg(clientId, uint16(rawInputArgs.MsgType), rawInputArgs.RawData)

		if err != nil {

			log.Debugf("SendRawMsg fail:", err.Error())

			//通知大厅，此通道下线了
			gate.removeCidQueue.Add(clientId)
		}
	}

}

func (gate *GateService) RawCloseClient(rawInput []byte) {

	var rawInputArgs rpc.RawInputArgs

	err := proto.Unmarshal(rawInput, &rawInputArgs)

	if err != nil {
		log.Error("msg is error", log.ErrorField("err", err))
		return
	}

	for _, clientId := range rawInputArgs.ClientIdList {
		gate.netModule.Close(clientId)
	}
}

// OnDiscoveryService 服务启动发现
func (gate *GateService) OnDiscoveryService(nodeId string, serviceList []string) {
	fmt.Printf("********gate.OnDiscoveryService nodeId:%s serviceList:%v\n", nodeId, serviceList)
}

// OnUnDiscoveryService 服务关闭发现
func (gate *GateService) OnUnDiscoveryService(nodeId string, serviceList []string) {
	fmt.Printf("******gate.OnUnDiscoveryService nodeId:%s serviceList:%v\n", nodeId, serviceList)
}

func (gate *GateService) ExecuteRemoveCidQueueLoop() {

	for gate.bStopStatus == false {

		if gate.removeCidQueue == nil {
			time.Sleep(5 * time.Second)
			continue
		}

		//若队列中数据没有要更新的
		if gate.removeCidQueue.Len() == 0 {
			time.Sleep(10 * time.Second)
			continue
		}

		clientList := make([]string, 0)

		for gate.removeCidQueue.Len() > 0 {

			obj := gate.removeCidQueue.Pop()

			if obj == nil {
				time.Sleep(3 * time.Second)
				continue
			}

			clientId, ok := obj.(string)

			if ok == false {
				continue
			}

			clientList = append(clientList, clientId)
		}

		log.Infof("gate.removeCidQueue push obj:%v", clientList)

		//通知游戏服，这个下线了
		gate.msgRouter.NotifyGameServiceRemoveClid(clientList)

	}

}
