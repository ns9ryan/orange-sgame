/*
   供本地RPC调用
*/

package lobbyservice

import (
	"errors"
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"videogame/common/proto/msg"
	"videogame/common/proto/rpc"
	"videogame/common/util"
)

// 玩家请求进入游戏房间,从gate收到此消息后，分发给指定的游戏
func (self *LobbyService) RPCJoinGame(joinmsg *rpc.JoinToGameRPC) error {

	//fmt.Printf("Local-->LobbyService--->RPCJoinGame user:%s,  gameId:%s ,deskNo:%s,  gateNodId:%s  \n ", joinmsg.UserId, joinmsg.GameId, joinmsg.DeskNo, joinmsg.GateNodeId)
	//log.Infof("Local-->LobbyService--->RPCJoinGame user:%s,  gameId:%s ,deskNo:%s,  gateNodId:%s  \n ", joinmsg.UserId, joinmsg.GameId, joinmsg.DeskNo, joinmsg.GateNodeId)

	gameName, ok := util.GetServerNameFromGameId(joinmsg.GameId)

	if ok == false {
		return errors.New("LobbyService.RPCJoinGame  GameId  no game  nodeId ")
	}

	//通知实际游戏服务器更新游戏状态
	nodeId := util.GetGameServiceName(gameName) //请求目的游戏服务的nodeId
	rpcFunc := util.FormatRpcMethod(gameName, "RPCJoinGame")

	if nodeId == "" {

		log.Errorf("*************LobbyService.RPCJoinGame  :%s  deskNo,%d find no game  nodeId !!!", gameName, joinmsg.DeskNo)

		return errors.New("LobbyService.RPCJoinGame  find no game  nodeId ")
	}

	//通知游戏的房间加入RPCJoinGame"
	err := self.AsyncCallNode(nodeId, rpcFunc, joinmsg, func(ret *rpc.Int, err error) {

		fmt.Printf("玩家请求进入游戏的服务器请求成功！:%s  userid:%s   \n", joinmsg.GameId, joinmsg.GetUserId())

		//发送成功加入游戏房间的消息给客户端
		var joinret msg.JoinToGameResult

		joinret.GameId = joinmsg.GameId
		joinret.DeskNo = joinmsg.DeskNo
		joinret.HallId = joinmsg.HallId

		if err != nil {
			//发送加入房间不成功的消息
			joinret.Success = 1
			joinret.Code = 0
		} else {
			joinret.Success = 0
			joinret.Code = int32(ret.Value)
		}

		//通知玩家端进入的结果
		self.msgSender.SendToPlayer(joinmsg.UserId, msg.MsgType_JoinGameResult, &joinret)

		//if err == nil {
		//
		//	//若是成功了，则需要通知游戏服务器发送游戏相关信息,
		//	rpcFunc = util.FormatRpcMethod(gameName, "RPCSendGameInfo")
		//
		//	//"gameName.RPCSendGameInfo", 通知游戏服发送相关的配置消息
		//	err = self.GoNode(nodeId, rpcFunc, joinmsg)
		//}

	})

	if err != nil {

		//发送不成功加入游戏房间的消息给客户端
		var joinret msg.JoinToGameResult

		joinret.GameId = joinmsg.GameId
		joinret.DeskNo = joinmsg.DeskNo
		joinret.Success = 1
		joinret.Code = 0
		joinret.HallId = joinmsg.HallId

		//通知玩家端进入的结果
		self.msgSender.SendToPlayer(joinmsg.UserId, msg.MsgType_JoinGameResult, &joinret)

		return errors.New("call 游戏服.JoinGame error ")
	}

	return nil

}

// 玩家要离开游戏房间,回到大厅
func (self *LobbyService) RPCLeaveGame(joinmsg *rpc.LeaveFromGameRPC) error {

	//	fmt.Printf(" LeaveGame gameId:%s ,deskNo:%s  userId:%s  \n ", joinmsg.GameId, joinmsg.DeskNo, joinmsg.UserId)

	gameName, ok := util.GetServerNameFromGameId(joinmsg.GameId)

	if ok == false {
		return errors.New("LobbyService.RPCLeaveGame  GameId  no game  nodeId ")
	}

	//通知实际游戏服务器更新游戏状态
	nodeId := util.GetGameServiceName(gameName) //请求目的游戏服务的nodeId

	if nodeId == "" {
		return errors.New("find no LobbyService.RPCLeaveGame  nodeId ")
	}

	rpcFunc := util.FormatRpcMethod(gameName, "RPCLeaveGame")

	//RPC通知离开游戏房间
	err := self.AsyncCallNode(nodeId, rpcFunc, joinmsg, func(ret *rpc.Int, err error) {

		//离开成功游戏房间的消息给客户端
		var joinret msg.LeaveFromGameResult

		joinret.GameId = joinmsg.GameId
		joinret.DeskNo = joinmsg.DeskNo
		joinret.Success = 0
		joinret.Code = int32(ret.Value)

		if err != nil {
			//发送加入房间不成功的消息
			joinret.Success = 1
			joinret.Code = 0
		}

		self.msgSender.SendToPlayer(joinmsg.UserId, msg.MsgType_ExitGameResult, &joinret)
	})

	if err != nil {

		var joinret msg.LeaveFromGameResult

		joinret.GameId = joinmsg.GameId
		joinret.DeskNo = joinmsg.DeskNo
		joinret.Success = 1
		joinret.Code = 0

		self.msgSender.SendToPlayer(joinmsg.UserId, msg.MsgType_ExitGameResult, &joinret)

		return errors.New("call 游戏服.LeaveFromGame error ")
	}

	return nil

}

// 玩家已进入房间了，玩家进入房间后，会发这条消息
func (self *LobbyService) RPCEnterGame(joinMsg *rpc.JoinToGameRPC) error {

	gameName, ok := util.GetServerNameFromGameId(joinMsg.GameId)

	if ok == false {
		return errors.New("LobbyService.RPCJoinGame  GameId  no game  nodeId ")
	}

	//通知实际游戏服务器更新游戏状态
	nodeId := util.GetGameServiceName(gameName) //请求目的游戏服务的nodeId

	//若是成功了，则需要通知游戏服务器发送游戏相关信息,
	rpcFunc := util.FormatRpcMethod(gameName, "RPCSendGameInfo")

	//"gameName.RPCSendGameInfo", 通知游戏服发送相关的配置消息
	err := self.GoNode(nodeId, rpcFunc, joinMsg)

	return err
}

// 玩家要设置本地SUPERSIX 的RPC调用,百家乐专用
func (self *LobbyService) RPCSuperSixSet(setmsg *rpc.GameSuperSixSet) error {

	//fmt.Printf(" LeaveGame gameId:%s ,deskNo:%s  userId:%s  \n ", setmsg.GameId, setmsg.DeskNo, setmsg.UserId)

	gameName, ok := util.GetServerNameFromGameId(setmsg.GameId)

	if ok == false {
		return errors.New("LobbyService.RPCSuperSixSet  GameId  no game  nodeId ")
	}

	//通知实际游戏服务器更新游戏状态
	nodeId := util.GetGameServiceName(gameName) //请求目的游戏服务的nodeId

	if nodeId == "" {
		return errors.New("find no LobbyService.RPCSuperSixSet  nodeId ")
	}

	rpcFunc := util.FormatRpcMethod(gameName, "RPCSuperSixSet")

	//发送对应的游戏服
	err := self.GoNode(nodeId, rpcFunc, setmsg)

	if err != nil {
		return errors.New("callnode game.RPCSuperSixSet  error ")
	}

	return nil

}
