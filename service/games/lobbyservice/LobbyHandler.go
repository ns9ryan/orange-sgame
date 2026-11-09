package lobbyservice

import (
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"videogame/common/proto/msg"
	"videogame/common/proto/rpc"
	"videogame/common/util"
	"videogame/service/games/player"
)

func (self *LobbyService) OnMsgPing(player *player.Player, msg *msg.MsgNil) {

	fmt.Println("大厅服接收到玩家的Ping...")

	if player == nil {
		return
	}

	//刷新ping的时间值
	player.UpdatePingTime()

}

// 进入游戏房间
func (self *LobbyService) OnMsgJoinGame(player *player.Player, joinMsg *msg.JoinToGame) {

	if player == nil {
		return
	}

	if joinMsg == nil {
		return
	}

	//fmt.Printf("LobbyService*******************OnMsgJoinGame ok :%s  gameId:%s deskNo:%d", joinMsg.UserId, joinMsg.GameId, joinMsg.DeskNo)
	log.Infof("lobbyserivce OnMsgJoinGame ok User :%s  gameId:%s deskNo:%s", joinMsg.UserId, joinMsg.GameId, joinMsg.DeskNo)

	//var ret int
	var joinrpc rpc.JoinToGameRPC

	joinrpc.UserId = player.Id
	joinrpc.ClientId = player.ClientId
	joinrpc.DeskNo = joinMsg.DeskNo
	joinrpc.GameId = joinMsg.GameId
	joinrpc.GateNodeId = player.GetGateNodeId()
	joinrpc.HallId = joinMsg.HallId

	rpcFunc := util.FormatRpcMethod(self.GetName(), "RPCJoinGame")

	//本地rpc调用，通知"LobbyService.RPCJoinGame"
	err := self.Go(rpcFunc, &joinrpc)

	if err != nil {

		log.Errorf("go lobbyService.RPCJoinGame error:%v", err)
		return
	}

	//更新人数
	//	self.GameLobby.GamePlayerNumAdd(joinMsg.GameId, 1)

	//广播游戏人数
	//	self.CastToPlayersOneGamePlayerNum(joinMsg.GameId)

}

func (self *LobbyService) OnMsgLeaveGame(player *player.Player, joinMsg *msg.LeaveFromGame) {

	if player == nil {
		return
	}
	if joinMsg == nil {
		return
	}

	//fmt.Println("OnMsgLeaveGame ok :%s  gameId:%s deskNo:%d", joinMsg.UserId, joinMsg.GameId, joinMsg.DeskNo)
	log.Infof("lobbyservice OnMsgLeaveGame ok :%s  gameId:%s deskNo:%s", joinMsg.UserId, joinMsg.GameId, joinMsg.DeskNo)

	var leaverpc rpc.LeaveFromGameRPC

	leaverpc.UserId = player.Id
	leaverpc.ClientId = player.ClientId
	leaverpc.DeskNo = joinMsg.DeskNo
	leaverpc.GameId = joinMsg.GameId
	leaverpc.GateNodeId = player.GetGateNodeId()

	rpcFunc := util.FormatRpcMethod(self.GetName(), "RPCLeaveGame")

	//本地rpc调用，通知
	err := self.Go(rpcFunc, &leaverpc)

	if err != nil {
		log.Errorf("go lobbyService.RPCJoinGame error:%v", err)
		return
	}

	//更新人数
	self.GameLobby.GamePlayerNumAdd(joinMsg.GameId, -1)

	//广播游戏人数
	self.CastToPlayersOneGamePlayerNum(joinMsg.GameId)

}

func (self *LobbyService) OnMsgEnterGame(player *player.Player, joinMsg *msg.EnterToGame) {

	if player == nil {
		return
	}

	if joinMsg == nil {
		return
	}

	//fmt.Printf("LobbyService*******************OnMsgEnterGame ok :%s  gameId:%s deskNo:%s \n", joinMsg.UserId, joinMsg.GameId, joinMsg.DeskNo)

	//var ret int
	var joinrpc rpc.JoinToGameRPC

	joinrpc.UserId = player.Id
	joinrpc.ClientId = player.ClientId
	joinrpc.DeskNo = joinMsg.DeskNo
	joinrpc.GameId = joinMsg.GameId
	joinrpc.GateNodeId = player.GetGateNodeId()

	rpcFunc := util.FormatRpcMethod(self.GetName(), "RPCEnterGame")

	//本地rpc调用，通知"LobbyService.RPCEnterGame"
	err := self.Go(rpcFunc, &joinrpc)

	if err != nil {
		return
	}

}

// 百家乐SUPERSIX的设置
func (self *LobbyService) OnMsgSuperSixSet(player *player.Player, setmsg *msg.CTOS_GameSuperSixSet) {

	if player == nil {
		return
	}
	if setmsg == nil {
		return
	}

	var supersixSet rpc.GameSuperSixSet

	supersixSet.DeskNo = setmsg.DeskNo
	supersixSet.GameId = setmsg.GameId
	supersixSet.SuperSixStatus = setmsg.Supersix

	supersixSet.UserId = player.Id
	supersixSet.ClientId = player.ClientId
	supersixSet.GateNodeId = player.GetGateNodeId()

	rpcFunc := util.FormatRpcMethod(self.GetName(), "RPCSuperSixSet")

	//本地rpc调用，通知
	err := self.Go(rpcFunc, &supersixSet)

	if err != nil {
		return
	}

	log.Infof("RPCSuperSixSet ok :userId %s  gameId:%s deskNo:%d", player.Id, setmsg.GameId, setmsg.DeskNo)

}
