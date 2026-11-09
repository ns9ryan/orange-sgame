package coreservice

import (
	"github.com/duanhf2012/origin/v2/node"
	"github.com/duanhf2012/origin/v2/service"
	"sync"
	"time"
	"videogame/common/eventid"
)

func init() {
	node.Setup(&CoreService{})
}

type CoreService struct {
	service.Service

	HttpModule HttpCoreModule

	serviceMapMutex sync.RWMutex            //用于下面的的存储
	ServiceMap      map[string]*ServiceNode //存储服务名用，每个游戏服都需要注册

}

func (core *CoreService) OnInit() error {

	//添加模块
	core.HttpModule = HttpCoreModule{}
	core.AddModule(&core.HttpModule)

	// 注册监听节点连接事件
	core.RegDiscoverListener(core)

	core.ServiceMap = make(map[string]*ServiceNode)

	//注册调用事件服务
	core.GetEventProcessor().RegEventReceiverFunc(eventid.Event_Core_CreateMerChant, core.GetEventHandler(), core.OnCreateMerchantEvent)
	core.GetEventProcessor().RegEventReceiverFunc(eventid.Event_Core_UpdateMerChant, core.GetEventHandler(), core.OnUpdateMerchantEvent)
	core.GetEventProcessor().RegEventReceiverFunc(eventid.Event_Core_UpdateTemplatePlanId, core.GetEventHandler(), core.OnOperateTemplatePlanId)
	core.GetEventProcessor().RegEventReceiverFunc(eventid.Event_Core_OperateGameConfig, core.GetEventHandler(), core.OnOperateGameConfigEvent)

	//性能监控
	core.OpenProfiler()
	core.GetProfiler().SetOverTime(time.Millisecond * 100)
	core.GetProfiler().SetMaxOverTime(time.Second * 10)

	return nil
}

// OnDiscoveryService 节点启动发现
func (core *CoreService) OnDiscoveryService(nodeId string, serviceList []string) {

	_ = serviceList

	//有新节点启动，从数据库重新获取一次网关的分配地址，
	//err := gate.setServiceNodeList()
	//
	//if err != nil {
	//	log.Errorf("HttpGateService OnDiscoverService err:", err.Error())
	//	return
	//}
	//
	//for _, serviceNode := range gate.serviceNodeList {
	//
	//
	//}

}

// OnUnDiscoveryService 节点停止发现
func (core *CoreService) OnUnDiscoveryService(nodeId string, serviceList []string) {
	_ = serviceList

}
