package httpgateservice

import (
	"errors"
	"fmt"
	"github.com/duanhf2012/origin/v2/cluster"
	"github.com/duanhf2012/origin/v2/log"
	"strings"
	"videogame/common/proto/rpc"
	"videogame/common/util"
	"videogame/service/games/player"
)

// 新建的商户信息
func (self *HttpGateService) RPCCreateMerchant(merchantItem *rpc.MerchantItem, ret *rpc.Int) error {

	_, ok := player.GetMerchantNode(merchantItem.MerchantId)

	//若没有读取，需要到数据库去读取
	if ok == true {

		ret.Value = 0

		//若读取到了，则不处理
		return nil
	}

	merchantNode, retok := player.BuildMerchantNode(merchantItem.MerchantId)

	//若没有建成功
	if retok == false {

		ret.Value = 1

		return errors.New("build merchant error")
	}

	if merchantNode != nil {
		log.Infof("create merchantId:%s ,parentId:%d, ok ", merchantNode.MerchantId, merchantNode.ParentId)
	}

	return nil
}

func (self *HttpGateService) RPCUpdateMerchant(merchantItem *rpc.MerchantItem, ret *rpc.Int) error {

	merchantNode, ok := player.GetMerchantNode(merchantItem.MerchantId)

	if ok == false {

		//重新建一个
		merchantNode, ok = player.BuildMerchantNode(merchantItem.MerchantId)

		//若没有建成功
		if ok == false {

			ret.Value = 1

			return errors.New("updatte  merchant error")
		}
	}

	//重新去数据库读取此代理的数据
	merchantNode.Blocked = util.ConvertIntToBool(int(merchantItem.Status))
	merchantNode.BetPaused = util.ConvertIntToBool(int(merchantItem.BetPause))

	player.ReLoadMerchantNode(merchantItem.MerchantId)

	return nil
}

func (self *HttpGateService) RPCUpdateBalance(req *rpc.GateServiceBalance) error {

	if req == nil {
		log.SError("req is nil")
		return fmt.Errorf("invalid rpc request")
	}

	if !cluster.HasService(req.GetNodeId(), "GateService") {

		log.SError("nodeid is does not exist in GateService")

		return fmt.Errorf("invalid rpc request")
	}

	self.HttpGateModule.UpBalance(req.GetNodeId(), req.GetIncreaseNum())

	return nil
}

// RPCGetBestGateUrl 获取最小负载gateUrl
func (h *HttpGateService) RPCGetBestGateUrl(req *rpc.Int, ret *rpc.String) error {
	_ = req

	ret.Value = h.HttpGateModule.GetBestGateUrl()

	return nil
}

// 更新网关的IP
func (h *HttpGateService) RPCRegisterGateUrl(nodeId *rpc.String, ret *rpc.Int) error {

	if nodeId.Value == "" {

	}

	h.HttpGateModule.MapBalanceInfoMutex.Lock()
	defer h.HttpGateModule.MapBalanceInfoMutex.Unlock()

	//有新节点启动，从数据库重新获取一次网关的分配地址，
	err := h.setServiceNodeList()

	if err != nil {

		log.Errorf("HttpGateService RPCUpdateGateUrl err:", err.Error())
		ret.Value = 0

		return errors.New("errors ")
	}

	for _, serviceNode := range h.serviceNodeList {

		if nodeId.Value == serviceNode.note { //启动的节点如果是网关的节点，则加入分配

			listenAddr := fmt.Sprintf("%s:%d", strings.TrimSpace(serviceNode.ip), serviceNode.port)

			h.HttpGateModule.MapBalanceInfo[nodeId.Value] = &BalanceInfo{GateUrl: listenAddr, NodeId: nodeId.Value}
			ret.Value = 1

			break
		}
	}

	fmt.Printf("**** RPCRegisterGateUrl OK ,nodeId:%s \n", nodeId.Value)

	return nil
}
