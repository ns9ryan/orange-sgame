package coreservice

import (
	"github.com/duanhf2012/origin/v2/log"
	"videogame/common/proto/rpc"
)

// 注册服务
func (core *CoreService) RPCRegisterService(serviceItem *rpc.ServiceItem, ret *rpc.Int) error {

	core.serviceMapMutex.Lock()
	defer core.serviceMapMutex.Unlock()

	service, ok := core.ServiceMap[serviceItem.GameId]

	if ok == false {

		nodeService := ServiceNode{
			SType:       int(serviceItem.RegType),
			GameId:      serviceItem.GameId,
			ServiceName: serviceItem.ServiceName,
			IpAddr:      serviceItem.IpAddr,
			WalletType:  uint(serviceItem.WParam),
		}
		core.ServiceMap[serviceItem.GameId] = &nodeService

	} else {

		service.ServiceName = serviceItem.ServiceName
		service.IpAddr = serviceItem.IpAddr
		service.SType = int(serviceItem.RegType)
		service.GameId = serviceItem.GameId
		service.WalletType = uint(serviceItem.WParam)
	}

	log.Infof("HttpCoreModule-->RegisterService :%s  sericename:%s, wParam:%d ", serviceItem.GameId,
		serviceItem.ServiceName, serviceItem.WParam)

	ret.Value = 1

	return nil
}
