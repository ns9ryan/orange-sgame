package lobbyservice

import (
	"github.com/duanhf2012/origin/v2/event"
	msg "videogame/common/proto/msg"
	"videogame/common/proto/rpc"
)

func (self *LobbyService) OnServiceGameStatusEvent(ev event.IEvent) {

	msgstatus, ok := (ev.(*event.Event).Data).(*msg.NotifyLobbyGameStatus)

	if !ok {
		return
	}
	//广播游戏状态
	self.CastToPlayersGameStatus(msgstatus)
}

// 服务监听事件回调,游戏开答事件,暂时不需要发给客户端了
func (self *LobbyService) OnServiceOpenAnswerEvent(ev event.IEvent) {

	msganswer, ok := (ev.(*event.Event).Data).(*msg.NotifyLobbyGameAnswer)

	//fmt.Printf("^^^^^^^^^^^^^^^^^^SicboService  OnServiceAnswerEvent  data:%+v  \n", asn)
	if !ok {
		return
	}

	//广播游戏状态
	self.CastToPlayersGameAnswer(msganswer)

}

// 通知发送历史开答表给大厅
func (self *LobbyService) OnServiceGameHistoryAnswersToLobbyEvent(ev event.IEvent) {

	msgAnswers, ok := (ev.(*event.Event).Data).(*msg.NotifyLobbyHistoryAnswers)

	//fmt.Printf("^^^^^^^^^^^^^^^^^^SicboService  OnServiceAnswerEvent  data:%+v  \n", asn)
	if !ok {
		return
	}

	self.CastToPlayerOneTableHistoryAnswers(msgAnswers)

}

// 通知发送在线玩家人数
func (self *LobbyService) OnServiceUpdateGameOnlineEvent(ev event.IEvent) {

	playerInfo, ok := (ev.(*event.Event).Data).(*rpc.GameOnlinePlayerNum)

	if !ok {
		return
	}

	self.CastToPlayersOneGamePlayerNum(playerInfo.GameId)

}
