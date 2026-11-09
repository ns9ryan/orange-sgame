package lobbyservice

import (
	"errors"
	"github.com/duanhf2012/origin/v2/event"
	"videogame/common/eventid"
	"videogame/common/msg_struct"
	"videogame/common/proto/msg"
	"videogame/common/proto/rpc"
	player "videogame/service/games/player"
)

// 接收游戏端发送过来的游戏状态的改动
func (self *LobbyService) RPCUpdateGameStatus(msg *msg.NotifyLobbyGameStatus, ret *rpc.Int) error {

	//fmt.Println("**********LobbyService RPCUpdateGameStatus... ")
	//fmt.Println("*********LobbyService RPCUpdateGameStatus...****** ")

	//更新游戏态
	self.GameLobby.UpdateGameStatus(msg)

	//通过事件队列通知更新玩家端
	self.GetEventHandler().NotifyEvent(&event.Event{
		Type: eventid.Event_Broad_GameStatus,
		Data: msg,
	})

	return nil
}

func (self *LobbyService) RPCOpenGameAnswer(ansmsg *msg.NotifyLobbyGameAnswer, ret *rpc.Int) error {

	//fmt.Printf("****************LobbyService RPCOpenGameAnswer..aa:%d  *************\n ", ansmsg.Set)
	//fmt.Printf("****************LobbyService RPCOpenGameAnswer... gameId:%s, dataType:%d \n", ansmsg.GameId, ansmsg.GameType)

	//通过事件队列通知更新玩家端
	self.GetEventHandler().NotifyEvent(&event.Event{
		Type: eventid.Event_Broad_Answer,
		Data: ansmsg,
	})

	return nil
}

func (self *LobbyService) RPCUpdateGameAnswersPerio(asp *msg.NotifyLobbyHistoryAnswers, ret *rpc.Int) error {

	self.GameLobby.UpdateGameAnswers(asp)

	//通过事件队列通知更新玩家端
	self.GetEventHandler().NotifyEvent(&event.Event{
		Type: eventid.Event_Broad_History_answers,
		Data: asp,
	})

	//log.Debugf("**********************LobbyService RPCUpdateGameAnswersPerio************************ DataType:%d \n", asp.DataType)

	return nil

}

// 更新荷管的信息
func (self *LobbyService) RPCUpdateDealerInfo(info *rpc.DealerUpdateInfo, ret *rpc.Int) error {

	var dealerInfo msg_struct.DealerInfo

	dealerInfo.DealerId = info.DealerId
	dealerInfo.Name = info.DealerName
	dealerInfo.Avator = info.Avator
	dealerInfo.Level = 0

	self.GameLobby.UpdateDealerInfo(info.GameId, &dealerInfo)

	return nil
}

// 其它游戏读取玩家信息
func (self *LobbyService) RPCGetPlayer(cliId *rpc.String, clientInfo *rpc.ClientInfoRPC) error {

	pp := self.GetClientPlayer(cliId.Value)

	if pp == nil {
		return errors.New("find not player ")
	}

	pplayer, ok := pp.(*player.Player)

	if !ok {
		return errors.New("not player ")
	}

	clientInfo.ClientId = cliId.Value
	clientInfo.UserId = pplayer.Id
	clientInfo.GateNodeId = pplayer.GetGateNodeId()

	//fmt.Println("LobbyService --->RPCGetPlayer, clintId:%s ---UserId:%d", clientInfo.ClientId, clientInfo.GateNodeId)

	return nil

}

// 更新在线玩家人数
func (self *LobbyService) RPCUpdateOnlinePlayerNum(playerInfo *rpc.GameOnlinePlayerNum, ret *rpc.Int) error {

	self.GameLobby.UpdateGamePlayerNum(playerInfo.GameId, int(playerInfo.OnlineNum))

	//用于事件通知
	self.GetEventHandler().NotifyEvent(&event.Event{
		Type: eventid.Event_Update_OnlineNum,
		Data: playerInfo,
	})

	return nil
}

//// 读取商户的资料,玩家的上级商户
//func (self *LobbyService) RPCGetMerchant(userId *rpc.String, merchantInfo *rpc.MerchantInfoRPC) error {
//
//	//
//
//	return nil
//
//}
