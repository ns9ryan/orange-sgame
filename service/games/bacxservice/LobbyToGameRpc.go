package bacxservice

import (
	"errors"
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"videogame/common/msg_struct"
	"videogame/common/proto/msg"
	"videogame/common/proto/rpc"
)

// 大厅会发送这个调用进入游戏桌
func (self *BacxService) RPCJoinGame(joinmsg *rpc.JoinToGameRPC, ret *rpc.Int) error {

	p := self.PlayersImp.GetPlayer(joinmsg.UserId)

	//若玩家不存在，则创建用户
	if p == nil {
		p = self.PlayersImp.NewPlayer(joinmsg.UserId, &self.msgSender, self, &self.msgReceiver)
	}

	if p != nil {

		p.InitPlayerInfo()

		p.InitBetLimit(msg_struct.GameTypeBacx, nil)

		//从钱包服读取玩家余额
		p.GetWalletBalance()
	}

	//玩家加入桌子
	self.GameTable.PlayerJoin(p, joinmsg.GateNodeId, joinmsg.ClientId)

	ret.Value = int32(self.GetPlayerNum())

	//更新玩家人数
	self.UpdateOnlinePlayerNum()

	return nil
}

// 大厅会发送这个离开进入游戏桌
func (self *BacxService) RPCLeaveGame(joinmsg *rpc.LeaveFromGameRPC, ret *rpc.Int) error {

	p := self.PlayersImp.GetPlayer(joinmsg.UserId)

	//创建用户
	if p == nil {

		log.Errorf("玩家进百家乐游戏gameId:%s ,在内存中找不到这个玩家了 :userId:%s \n", self.GameTable.GetGameId(), joinmsg.UserId)

		return nil
	}

	//fmt.Printf("BacxService RPCLeaveGame userID:%v\n", p.Id)

	ok := self.GameTable.PlayerLeave(p)

	if ok == true {

		//需要检测玩家是否有投注,有投注话不能删除，需要进册除队列等待
		if self.GameTable.IsPlayerBetted(p.Id) == false {

			log.Infof("gameId:%s 玩家正常请求退出，没有下注，可以正常清除！！！", self.GameTable.GetGameId())

			self.PlayersImp.ReleasePlayer(p)

		} else {

			//进入册除队列中
			log.Infof("gameId:%s  玩家正常请求退出，但它现在有下注，不能删除它！！！", self.GameTable.GetGameId())
		}

	}

	ret.Value = int32(self.GetPlayerNum())

	//更新玩家人数
	self.UpdateOnlinePlayerNum()

	return nil
}

// 通过RPC要求发送消息游戏配置，卡号信息等给玩家
func (self *BacxService) RPCSendGameInfo(joinmsg *rpc.JoinToGameRPC) error {

	p := self.PlayersImp.GetPlayer(joinmsg.UserId)

	if p == nil {

		var enterFailed msg.EnterToGameFailed

		enterFailed.GameId = joinmsg.GameId
		enterFailed.DeskNo = joinmsg.DeskNo
		enterFailed.Code = 1

		p.SendMsg(msg.MsgType_EnterGameFailed, &enterFailed)

		log.Infof("player not found userid: ", joinmsg.UserId)

		return errors.New("player not found ")
	}

	//进入房间了
	p.SetRoomIn(true)

	//发送游戏配置信息，视频地址， 高低注信息
	self.GameTable.SendToPlayerGameSetting(p)

	//发送高低注
	self.GameTable.SendToPlayerBetLimit(p)

	//发送卡号信息
	self.GameTable.SendToPlayerCardInfo(p)

	//发送supersix 开关状态
	self.GameTable.SendToPlayerSuperSix(p)

	//发送游戏状态信息
	self.GameTable.SendToPlayerGameStatus(p)

	//发送历史开答信息
	self.GameTable.SendToPlayerHistoryAnswers(p)

	//发送玩家之前有下注的情况
	self.GameTable.SendToPlayerBetted(p)

	fmt.Printf("****BacxService   RPCSendGameInfo userID:%v\n", joinmsg.UserId)

	return nil

}

// 通过RPC设置SupSerSix开关
func (self *BacxService) RPCSuperSixSet(setmsg *rpc.GameSuperSixSet) error {

	log.Infof("收到 RPCSuperSixSet :%s , %d ", setmsg.UserId, setmsg.SuperSixStatus)
	//fmt.Printf("RPCSuperSixSet :%s , %d ", setmsg.UserId, setmsg.SuperSixStatus)

	p := self.PlayersImp.GetPlayer(setmsg.UserId)

	if p == nil {

		log.Infof("player not found userid: ", setmsg.UserId)

		return errors.New("player not found ")
	}

	//状态一样，不用改
	if setmsg.SuperSixStatus == int32(p.GetSuperSix()) {

	} else if self.GameTable.IsBetting() == true { //检测是否处在投注状态

		//若玩家有投注不能修改投注状态
		if self.GameTable.IsPlayerBetted(p.Id) == true { //检测玩家是否有投注，有投注了不能改

		} else { //若没有投注，则可以修改

			p.SetSuperSix(int(setmsg.SuperSixStatus))

		}

	} else {
		//修改SuperSix状态
		p.SetSuperSix(int(setmsg.SuperSixStatus))
	}

	self.GameTable.SendToPlayerSuperSix(p)

	return nil
}
