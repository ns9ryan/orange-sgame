package httpgateservice

import (
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/node"
	"github.com/duanhf2012/origin/v2/service"
	"strings"
	"time"
	models "videogame/common/model"
	"videogame/common/util"
)

func init() {
	node.Setup(&HttpGateService{})
}

type HttpGateService struct {
	service.Service

	HttpGateModule HttpGateModule

	serviceNodeList []*ServiceNode //gate网关的配置信息列表，从数据库读取
}

func (gate *HttpGateService) OnInit() error {

	//添加模块
	gate.HttpGateModule = HttpGateModule{}
	gate.AddModule(&gate.HttpGateModule)

	// 注册监听节点连接事件
	gate.RegDiscoverListener(gate)

	//读取数据库或者全局gate配置文件的gate网关的配置地址和端口，（debug=true:读取全局的配置文件中的网关地址和端口）
	err := gate.setServiceNodeList()

	if err != nil {
		return fmt.Errorf("WSCfg are not configured")
	}

	//性能监控
	gate.OpenProfiler()
	gate.GetProfiler().SetOverTime(time.Millisecond * 100)
	gate.GetProfiler().SetMaxOverTime(time.Second * 10)

	return nil
}

func (gate *HttpGateService) setServiceNodeList() error {

	//读取全局的Debug模式的配置
	debug, err := util.GetGlobalDebug()

	if err != nil {
		return err
	}

	var err1 error

	if debug { //开启debug 读取全局配置文件的网关地址和端口，分配网关
		err1 = gate.setGateHostListFromConfig()
	} else { // 读取数据库的网关地址和端口，分配网关
		err1 = gate.setGateHostListFromDB()
	}

	return err1

}

func (gate *HttpGateService) setGateHostListFromDB() error {
	gate.serviceNodeList = make([]*ServiceNode, 0)

	serviceNodeList, err := models.GetServiceNodeList()
	if err != nil {
		log.Errorf("HttpGateService OnInit GetServiceNodeList err:", err.Error())
		return err
	}

	for _, serviceNode := range serviceNodeList {

		snode := &ServiceNode{
			ip:   serviceNode.Ip,
			port: serviceNode.Port,
			note: serviceNode.Note,
		}
		gate.serviceNodeList = append(gate.serviceNodeList, snode)
	}

	return nil
}

func (gate *HttpGateService) setGateHostListFromConfig() error {

	gate.serviceNodeList = make([]*ServiceNode, 0)

	list, err := util.GetGlobalGateServiceCfgList()

	if err != nil {
		log.Errorf("GetGlobalGateServiceCfgList err:", err.Error())
		return err
	}

	for _, l := range list {
		snode := &ServiceNode{
			ip:   l.Host,
			port: l.Port,
			note: l.NodeId,
		}
		gate.serviceNodeList = append(gate.serviceNodeList, snode)
	}

	return nil
}

// OnDiscoveryService 节点启动发现
func (gate *HttpGateService) OnDiscoveryService(nodeId string, serviceList []string) {
	_ = serviceList

	gate.HttpGateModule.MapBalanceInfoMutex.Lock()
	defer gate.HttpGateModule.MapBalanceInfoMutex.Unlock()

	//有新节点启动，从数据库重新获取一次网关的分配地址，
	err := gate.setServiceNodeList()

	if err != nil {
		log.Errorf("HttpGateService OnDiscoverService err:", err.Error())
		return
	}

	for _, serviceNode := range gate.serviceNodeList {

		if nodeId == serviceNode.note { //启动的节点如果是网关的节点，则加入分配

			listenAddr := fmt.Sprintf("%s:%d", strings.TrimSpace(serviceNode.ip), serviceNode.port)

			gate.HttpGateModule.MapBalanceInfo[nodeId] = &BalanceInfo{GateUrl: listenAddr, NodeId: nodeId}
			break
		}
	}

}

// OnUnDiscoveryService 节点停止发现
func (gate *HttpGateService) OnUnDiscoveryService(nodeId string, serviceList []string) {
	_ = serviceList

	gate.HttpGateModule.MapBalanceInfoMutex.Lock()
	defer gate.HttpGateModule.MapBalanceInfoMutex.Unlock()

	balanceInfo := gate.HttpGateModule.MapBalanceInfo[nodeId]
	if balanceInfo != nil {
		delete(gate.HttpGateModule.MapBalanceInfo, nodeId)
	}

}
