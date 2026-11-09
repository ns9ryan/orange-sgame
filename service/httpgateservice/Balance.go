package httpgateservice

import (
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"videogame/common/util"
)

type ServiceNode struct {
	ip   string
	port uint
	note string //gate1,gate2...
}

type BalanceInfo struct {
	Weight  int64
	NodeId  string
	GateUrl string
}

func (h *HttpGateModule) initMapBalance() {
	h.MapBalanceInfo = make(map[string]*BalanceInfo, 0)
}

// GetBestGateUrl 获取最小负载gateUrl
func (h *HttpGateModule) GetBestGateUrl() string {

	h.MapBalanceInfoMutex.Lock()
	defer h.MapBalanceInfoMutex.Unlock()

	if len(h.MapBalanceInfo) == 0 {

		log.Errorf("GETURL:ENMPTY  111")
		return h.getGateUrl(defaultLisenAddr)
	}

	var minBalance *BalanceInfo

	for _, balance := range h.MapBalanceInfo {
		if minBalance == nil || balance.Weight <= minBalance.Weight {
			minBalance = balance
		}
	}

	if minBalance.GateUrl == "" {

		log.Errorf("GETURL:ENMPTY 222 ")

		return h.getGateUrl(defaultLisenAddr)
	}

	log.Errorf("GETURL:ENMPTY 333,%s ", minBalance.GateUrl)

	return h.getGateUrl(minBalance.GateUrl)
}

// GetGateUrl 获取url
func (h *HttpGateModule) getGateUrl(lisenAddr string) string {
	wsProto := "ws"

	pem, err := util.GetGlobalPemFile()
	if pem.KeyFile != "" && pem.CertFile != "" && err == nil {
		wsProto = "wss"
	}

	return fmt.Sprintf("%s://%s", wsProto, lisenAddr)

}

func (h *HttpGateModule) UpBalance(nodeId string, increaseNum int64) {

	fmt.Printf("************* HttpGateModule UpBalance %s  num:%d \n", nodeId, increaseNum)

	h.MapBalanceInfoMutex.Lock()
	defer h.MapBalanceInfoMutex.Unlock()

	// 获取或初始化 balanceInfo
	balanceInfo, exists := h.MapBalanceInfo[nodeId]

	if !exists {

		balanceInfo = &BalanceInfo{
			Weight: 1, // 初始化为 1
		}
		h.MapBalanceInfo[nodeId] = balanceInfo

		return
	}

	// 更新权重
	balanceInfo.Weight += increaseNum
	if balanceInfo.Weight < 0 {
		balanceInfo.Weight = 0
	}
}
