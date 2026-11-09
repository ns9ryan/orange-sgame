package lobbyservice

import (
	"github.com/duanhf2012/origin/v2/log"
	"github.com/pkg/errors"
	"videogame/common/proto/rpc"
	"videogame/common/util"
	"videogame/service/games/player"
	"videogame/service/games/templatelimit"
)

// 新建的商户信息
func (self *LobbyService) RPCCreateMerchant(merchantItem *rpc.MerchantItem, ret *rpc.Int) error {

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

// 更新商户信息
func (self *LobbyService) RPCUpdateMerchant(merchantItem *rpc.MerchantItem, ret *rpc.Int) error {

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

// 更新模板限红ID表
func (self *LobbyService) RPCUpdatePlanId(planItem *rpc.PlanIdItem, ret *rpc.Int) error {

	//新增加
	if planItem.Opt == 0 {

		_, ok := templatelimit.SetTemplatePlanData(planItem.PlanId)

		if ok == false {
			ret.Value = 1
			return errors.New("set template plan error")
		}

	} else if planItem.Opt == 1 {

		//更新
		_, ok := templatelimit.ReGetTemplatePlanData(planItem.PlanId)

		if ok == false {
			ret.Value = 1
			return errors.New("set template plan error")
		}

	} else if planItem.Opt == 2 {

		//删除不需要处理
	}

	return nil

}

// 踢除玩家
func (self *LobbyService) RPCKickPlayer(playerId *rpc.String, ret *rpc.Int) error {

	//查找此会员是否在线
	pplayer := self.PlayersImp.GetPlayer(playerId.Value)

	if pplayer == nil {
		return errors.New("player not found!!!")
	}

	return nil
}

// 用于投注服传递投注的RPC调用
func (self *LobbyService) RPCOperateGameConfig(setItem *rpc.GameConfigItem, ret *rpc.Int) error {

	log.Infof("LobbyService-->RPCOperateGameConfig ok ")

	self.GameLobby.UpdateTableInfo(setItem.GameId, setItem.VideoUrl, setItem.GameTitle)

	return nil

}
