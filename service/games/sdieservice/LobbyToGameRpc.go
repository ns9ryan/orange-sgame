package sdieservice

import (
	"errors"
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"videogame/common/msg_struct"
	"videogame/common/proto/rpc"
)

// 大厅会发送这个调用进入游戏桌
func (self *SdieService) RPCJoinGame(joinmsg *rpc.JoinToGameRPC, ret *rpc.Int) error {

	p := self.PlayersImp.GetPlayer(joinmsg.UserId)

	//创建用户
	if p == nil {

		p = self.PlayersImp.NewPlayer(joinmsg.UserId, &self.msgSender, self, &self.msgReceiver)

		if p != nil {

			//初始化玩家的数据
			p.InitPlayerInfo()

			p.InitBetLimit(msg_struct.GameTypeSide, nil)

			//从钱包服读取玩家余额
			p.GetWalletBalance()
		}

		fmt.Printf("色碟游戏 ，初始化 玩家:%s 的数据完成! \n", joinmsg.UserId)
	}

	//读取余额
	self.GameTable.PlayerJoin(p, joinmsg.GateNodeId, joinmsg.ClientId)

	ret.Value = int32(self.GetPlayerNum())

	self.UpdateOnlinePlayerNum()

	return nil
}

// 大厅会发送这个离开进入游戏桌
func (self *SdieService) RPCLeaveGame(joinmsg *rpc.LeaveFromGameRPC, ret *rpc.Int) error {

	p := self.PlayersImp.GetPlayer(joinmsg.UserId)

	//创建用户
	if p == nil {

		fmt.Printf("玩家要求玩家游戏房间，但之前已经离开! \n")

		return nil
	}

	fmt.Printf("SdieService RPCLeaveGame userID:%v\n", p.Id)

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

	self.UpdateOnlinePlayerNum()

	return nil
}

// 通过RPC要求发送消息游戏配置，卡号信息等给玩家
func (self *SdieService) RPCSendGameInfo(joinmsg *rpc.JoinToGameRPC) error {

	p := self.PlayersImp.GetPlayer(joinmsg.UserId)

	if p == nil {

		log.Infof("player not found userid: ", joinmsg.UserId)

		return errors.New("player not found ")
	}

	//发送游戏配置信息，视频地址， 高低注信息
	self.GameTable.SendToPlayerGameSetting(p)

	//发送高低注
	self.GameTable.SendToPlayerBetLimit(p)

	//发送卡号信息
	self.GameTable.SendToPlayerCardInfo(p)

	//发送游戏状态信息
	self.GameTable.SendToPlayerGameStatus(p)

	//发送历史开答信息
	self.GameTable.SendToPlayerHistoryAnswers(p)

	//发送玩家之前有下注的情况
	self.GameTable.SendToPlayerBetted(p)

	fmt.Printf("****SdieService   RPCSendGameInfo userID:%v\n", joinmsg.UserId)

	return nil

}
