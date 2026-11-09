package rotservice

import (
	"errors"
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"videogame/common/msg_struct"
	"videogame/common/proto/msg"
	"videogame/common/proto/rpc"
)

// 大厅会发送这个调用进入游戏桌
func (self *RotService) RPCJoinGame(joinmsg *rpc.JoinToGameRPC, ret *rpc.Int) error {
	fmt.Printf("RotService joinmsg:%v\n", joinmsg)
	p := self.PlayersImp.GetPlayer(joinmsg.UserId)

	//创建用户
	if p == nil {

		p = self.PlayersImp.NewPlayer(joinmsg.UserId, &self.msgSender, self, &self.msgReceiver)

		//	fmt.Printf("RotService NewPlayer userID:%v  \n", p.Id)

		if p != nil {

			//初始化玩家的数据
			p.InitPlayerInfo()

			p.InitBetLimit(msg_struct.GameTypeRot, nil)

			//从钱包服读取玩家余额
			p.GetWalletBalance()
		}

	}

	//读取余额
	self.GameTable.PlayerJoin(p, joinmsg.GateNodeId, joinmsg.ClientId)

	ret.Value = int32(self.GetPlayerNum())

	self.UpdateOnlinePlayerNum()

	return nil
}

// 大厅会发送这个离开进入游戏桌
func (self *RotService) RPCLeaveGame(joinmsg *rpc.LeaveFromGameRPC, ret *rpc.Int) error {

	p := self.PlayersImp.GetPlayer(joinmsg.UserId)

	//创建用户
	if p == nil {

		fmt.Printf("RotService  玩家要求玩家游戏房间，但之前已经离开! \n")

		return nil
	}

	fmt.Printf("RotService RPCLeaveGame userID:%v\n", p.Id)

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
func (self *RotService) RPCSendGameInfo(joinmsg *rpc.JoinToGameRPC) error {

	fmt.Printf("LobbyRpc --->RotService ---> RPCSendGameInfo userID:%v\n", joinmsg.UserId)

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

	p.SetRoomIn(true)

	//发送游戏配置信息，视频地址
	self.GameTable.SendToPlayerGameSetting(p)

	// 发送高低注信息
	self.GameTable.SendToPlayerBetLimit(p)

	//发送卡号信息
	self.GameTable.SendToPlayerCardInfo(p)

	//发送游戏状态信息
	self.GameTable.SendToPlayerGameStatus(p)

	//发送历史开答信息
	self.GameTable.SendToPlayerHistoryAnswers(p)

	//玩家掉线前，若有投注，在开答前上线，需要发送此消息
	self.GameTable.SendToPlayerBetted(p)

	fmt.Printf("****RotService RPCSendGameInfo userID:%v\n", joinmsg.UserId)

	return nil

}
