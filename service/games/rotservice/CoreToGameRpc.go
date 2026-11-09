package rotservice

import (
	"errors"
	"github.com/duanhf2012/origin/v2/log"
	"videogame/common/proto/rpc"
	"videogame/common/util"
	"videogame/service/games/player"
)

// 新建的商户信息
func (self *RotService) RPCCreateMerchant(merchantItem *rpc.MerchantItem, ret *rpc.Int) error {

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

func (self *RotService) RPCUpdateMerchant(merchantItem *rpc.MerchantItem, ret *rpc.Int) error {

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

// 用于投注服传递投注的RPC调用
func (self *RotService) RPCOperateGameConfig(setItem *rpc.GameConfigItem, ret *rpc.Int) error {

	if setItem.GameId != self.GameTable.GetGameId() {

		return errors.New("gameId not same ")
	}

	self.GameTable.SetTableBetTime(int(setItem.TotalBetTime))
	self.GameTable.SetTableVideoUrl(setItem.VideoUrl)

	//重新读取数据库表配置参数
	//if self.GameTable.ReadDBConfig(setItem.GameId) == false {
	//
	//	log.Errorf("read bacx dbconfig error！！！")
	//}

	log.Infof("BacxService-->RPCOperateGameConfig ok ")

	nodeId, rpcFunc := util.GetServiceRpcMethodName(util.LobbyService, "RPCOperateGameConfig")

	if nodeId != "" {

		//调用大厅
		self.GoNode(nodeId, rpcFunc, setItem)
	}

	//重新存到redis
	self.GameTable.SaveTableInfoToRedis()

	return nil

}
