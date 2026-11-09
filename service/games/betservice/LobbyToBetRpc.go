package betService

import (
	"fmt"
	"videogame/common/proto/rpc"
)

// 大厅会发送这个调用,当有玩家登陆游戏时，大厅会调用此RPC
func (self *BetService) RPCJoinBet(req *rpc.LoginToLobbyReq) error {

	p := self.PlayerImps.GetPlayer(req.UserId)

	fmt.Printf("**************Lobby-->BetService RPCJoinBet \n")

	//创建用户
	if p == nil {

		p = self.PlayerImps.NewPlayer(req.UserId, &self.msgSender, self, &self.msgReceiver)

		fmt.Printf("投注服有玩家加入  NewPlayer userID:%v  \n", p.Id)

	} else {

	}

	self.attachConn(req.ClientId, req.GateNodeId, p)

	//从redis或数据库读取玩家的额度

	return nil
}

// 大厅会发送这个调用,当有玩家退出游戏时，大厅会调用此RPC
func (self *BetService) RPCLeaveBet(req *rpc.LoginToLobbyReq, ret *rpc.Int) error {

	p := self.PlayerImps.GetPlayer(req.UserId)

	//创建用户
	if p == nil {
		fmt.Printf("玩家要求玩家游戏房间，但之前已经离开! \n")
		return nil
	}

	fmt.Printf("投注服有玩家离开 BetService RPCLeaveBet userID:%v\n", p.Id)

	self.delClientPlayer(req.ClientId)
	self.PlayerImps.ReleasePlayer(p)

	return nil
}
