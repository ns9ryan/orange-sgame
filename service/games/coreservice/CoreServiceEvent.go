package coreservice

import (
	"github.com/duanhf2012/origin/v2/event"
	"github.com/duanhf2012/origin/v2/log"
	"videogame/common/msg_struct"
	"videogame/common/msg_struct_crm"
	"videogame/common/proto/rpc"
	"videogame/common/util"
)

// 操作游戏配置事件
func (core *CoreService) OnOperateGameConfigEvent(ev event.IEvent) {

	setItem, ok := (ev.(*event.Event).Data).(*msg_struct_crm.GameConfigItem)

	if ok == false {
		return
	}
	log.Infof("OnOperateGameConfigEvent  GameId:%s", setItem.GameId)

	serviceNode, ok := core.ServiceMap[setItem.GameId]

	if ok == false {
		return
	}

	nodeId := util.GetGameServiceName(serviceNode.ServiceName) //请求目的游戏服务的nodeId
	rpcFunc := util.FormatRpcMethod(serviceNode.ServiceName, "RPCOperateGameConfig")

	if nodeId == "" {
		return
	}

	var setReq rpc.GameConfigItem

	setReq.GameId = setItem.GameId
	setReq.TotalBetTime = int32(setItem.TotalBetTime)
	setReq.GameTitle = setItem.GameTitle
	setReq.VideoUrl = setItem.VideoUrl

	//调用对应的游戏服务器
	err := core.GoNode(nodeId, rpcFunc, &setReq)

	if err != nil {

		log.Errorf("Call nodeId:%s %s failed", nodeId, rpcFunc)
	}

}

// 建新的商户
func (core *CoreService) OnCreateMerchantEvent(ev event.IEvent) {

	merchantItem, ok := (ev.(*event.Event).Data).(*msg_struct_crm.MerchantItem)

	if ok == false {
		return
	}

	var createReq rpc.MerchantItem

	createReq.MerchantId = merchantItem.MerchantId
	createReq.ParentId = merchantItem.ParentId
	createReq.BetPause = int32(merchantItem.BetPause)
	createReq.Status = int32(merchantItem.Status)

	for _, serviceNode := range core.ServiceMap {

		rpcFunc := ""
		nodeId := ""

		if serviceNode.SType == msg_struct.SReg_Game ||
			serviceNode.SType == msg_struct.SReg_Lobby ||
			serviceNode.SType == msg_struct.SReg_Login {

		} else if serviceNode.SType == msg_struct.SReg_Wallet {
			//若是钱包服，需要通知单钱包

		} else {
			continue
		}

		nodeId, rpcFunc = util.GetServiceRpcMethodName(serviceNode.ServiceName, "RPCCreateMerchant")

		if nodeId == "" {
			continue
		}

		//通知对应的服务器
		err := core.GoNode(nodeId, rpcFunc, &createReq)

		if err != nil {

			log.Errorf("CoreService-->OnCreateMerchantEvent  Call nodeId:%s %s failed", nodeId, rpcFunc)
			return
		}

	}

}

// 更新商户
func (core *CoreService) OnUpdateMerchantEvent(ev event.IEvent) {

	merchantItem, ok := (ev.(*event.Event).Data).(*msg_struct_crm.MerchantItem)

	if ok == false {
		return
	}

	var createReq rpc.MerchantItem

	createReq.MerchantId = merchantItem.MerchantId
	createReq.ParentId = merchantItem.ParentId
	createReq.BetPause = int32(merchantItem.BetPause)
	createReq.Status = int32(merchantItem.Status)

	for _, serviceNode := range core.ServiceMap {

		rpcFunc := ""
		nodeId := ""

		if serviceNode.SType == msg_struct.SReg_Game ||
			serviceNode.SType == msg_struct.SReg_Lobby ||
			serviceNode.SType == msg_struct.SReg_Login {

		} else if serviceNode.SType == msg_struct.SReg_Wallet {
			//若是钱包服，需要通知单钱包

		} else {
			continue
		}

		nodeId, rpcFunc = util.GetServiceRpcMethodName(serviceNode.ServiceName, "RPCUpdateMerchant")

		if nodeId == "" {
			continue
		}

		//通知对应的服务器
		err := core.GoNode(nodeId, rpcFunc, &createReq)

		if err != nil {

			log.Errorf("CoreService-->OnUpdateMerchantEvent  Call nodeId:%s %s failed", nodeId, rpcFunc)
			return
		}

	}

}

// 操作商户的限红模板ID
func (core *CoreService) OnOperateTemplatePlanId(ev event.IEvent) {

	PlanIdItem, ok := (ev.(*event.Event).Data).(*msg_struct_crm.PlanIdItem)

	if ok == false {
		return
	}

	var planReq rpc.PlanIdItem

	planReq.PlanId = PlanIdItem.PlanId
	planReq.Opt = int32(PlanIdItem.Opt)
	planReq.Digits = uint32(PlanIdItem.Digits)
	planReq.MaxValue = uint32(PlanIdItem.MaxValue)
	planReq.MinValue = uint32(PlanIdItem.MinValue)

	for _, serviceNode := range core.ServiceMap {

		rpcFunc := ""
		nodeId := ""

		if serviceNode.SType != msg_struct.SReg_Lobby {
			continue
		}

		nodeId, rpcFunc = util.GetServiceRpcMethodName(serviceNode.ServiceName, "RPCUpdatePlanId")

		if nodeId == "" {
			continue
		}

		//通知对应的服务器
		err := core.GoNode(nodeId, rpcFunc, &planReq)

		if err != nil {

			log.Errorf("CoreService-->OnOperateTemplatePlanId  Call nodeId:%s %s failed", nodeId, rpcFunc)
			return
		}

		break

	}

}

// 操作游戏配置
func (core *CoreService) OnOperateGameConfig(ev event.IEvent) {

	ConfigItem, ok := (ev.(*event.Event).Data).(*msg_struct_crm.GameConfigItem)

	if ok == false {
		return
	}

	var configReq rpc.GameConfigItem

	configReq.GameId = ConfigItem.GameId
	configReq.DeskNo = int32(ConfigItem.DeskNo)
	configReq.GameTitle = ConfigItem.GameTitle
	configReq.TotalBetTime = int32(ConfigItem.TotalBetTime)
	configReq.VideoUrl = ConfigItem.VideoUrl

	for _, serviceNode := range core.ServiceMap {

		rpcFunc := ""
		nodeId := ""

		//只有lobby,game才接收
		if serviceNode.SType != msg_struct.SReg_Lobby {
			if serviceNode.SType != msg_struct.SReg_Game {
				continue
			}
		}

		nodeId, rpcFunc = util.GetServiceRpcMethodName(serviceNode.ServiceName, "RPCOperateGameConfig")

		if nodeId == "" {
			continue
		}

		//通知对应的服务器
		err := core.GoNode(nodeId, rpcFunc, &configReq)

		if err != nil {

			log.Errorf("CoreService-->OnOperateGameConfig  Call nodeId:%s %s failed", nodeId, rpcFunc)
			return
		}

		break

	}

}
