package lobbyservice

import (
	"errors"
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	models "videogame/common/model"
	"videogame/common/proto/msg"
	"videogame/service/games/templatelimit"
)

// 发送大厅相关的配置信息给玩家
func (self *LobbyService) SendToPlayerLobbySetting(userId string) error {

	return nil
}

// 发送玩家帐号信息
func (self *LobbyService) SendToPlayerUserInfo(userId string) error {

	var playInfo msg.STOC_PlayerInfo

	p := self.PlayersImp.GetPlayer(userId)

	if p == nil {

		return errors.New("player not exist")
	}

	//读取钱包服的余额
	if p.GetWalletBalance() == false {
		return errors.New("insufficient balance cluster error!!")
	}

	playInfo.UserId = p.UserId
	playInfo.UserName = p.UserName
	playInfo.Currency = p.GetCurrency()

	//	playInfo.Credit = int64(p.GetCredit())
	//	playInfo.Digits = int32(p.GetDigits())

	playInfo.Credit = uint32(p.GetCredit())
	playInfo.Digits = int32(p.GetDigits())

	p.SendMsg(msg.MsgType_PlayerInfo_Card, &playInfo)

	fmt.Printf("***********Send Player:%s credit:%d , Digits:%d \n", playInfo.UserId, playInfo.Credit, playInfo.Digits)

	return nil
}

// 发送游戏的历史开答
func (self *LobbyService) SendToPlayerAnswers(userId string) error {

	p := self.PlayersImp.GetPlayer(userId)

	if p == nil {
		return errors.New("player not exist")
	}

	//发送各游戏的开答列表给玩家
	self.GameLobby.SendHistoryAnswerToPlayer(p)

	return nil
}

// 广播发送游戏分组信息
func (self *LobbyService) SendToPlayerGameGroups(userId string) error {

	var gameGroups msg.STOC_LobbyGameGroups

	var templateArea models.SMerchantTemplateBetlimitArea

	p := self.PlayersImp.GetPlayer(userId)

	if p == nil {
		return errors.New("player not exist")
	}

	//1.先发送游戏分组的组别信息

	//gameGroups.GroupId = append(gameGroups.GroupId, []string{"BACX", "DTX", "ROT", "SICBO", "FTAN", "SDIE"}...)
	for _, group := range self.GameLobby.GameGroup {

		gameGroups.GroupId = append(gameGroups.GroupId, group.GroupName)
	}

	gameGroups.GroupNum = int32(len(gameGroups.GroupId))

	//先发送游戏分组信息
	p.SendMsg(msg.MsgType_Lobby_GameGroup, &gameGroups)

	gdeskNo := 1

	//var gameNum = [6]int{3, 1, 1, 1, 1, 1}

	//发送每组内的游戏信息
	for i := 0; i < int(gameGroups.GroupNum); i++ {

		var gameRooms msg.STOC_LobbyGameRooms

		gameRooms.GroupId = gameGroups.GroupId[i]

		//先得到多少数组
		gameNum := len(self.GameLobby.MapGameGroup[gameRooms.GroupId])

		//每个游戏组不一样
		for j := 0; j < gameNum; j++ {

			room := msg.GameRoom{}

			gameId := self.GameLobby.MapGameGroup[gameRooms.GroupId][j]

			tableInfo, ok := self.GameLobby.MapGameTable[gameId]

			if !ok {
				continue
			}

			//读取模板限红
			PlanId, err := templateArea.GetMerchantTemplatePlanId(p.BetLimitTemplateId, gameId)

			if err != nil {
				log.Warnf("gameId = %s,uid = %s,BetLimitTemplateId = %s ,读取模板限红失败", gameId, p.UserName, p.BetLimitTemplateId)
				continue
			}

			planItem, ok1 := templatelimit.GetTemplatePlanData(PlanId)

			if !ok1 {

				log.Warnf("PlanId = %s 限红方案未设置进内存", PlanId)

				//再次初始化限红模板(解决新增限红方案时，出现丢游戏桌的问题)
				planItem, ok1 = templatelimit.SetTemplatePlanData(PlanId)

				if !ok1 {
					log.Errorf("gameId = %s,uid = %s,BetLimitTemplateId = %s PlanId = %s,限红方案设置进内存失败", gameId, p.UserName, p.BetLimitTemplateId, PlanId)
					continue
				}

				planItem, ok1 = templatelimit.GetTemplatePlanData(PlanId)

			}

			room.RoomNo = int32(i)
			room.DeskNo = int32(tableInfo.DeskNo)

			room.GameId = gameId
			room.Title = tableInfo.RoomInfo.Title
			room.VideoURL = tableInfo.RoomInfo.VideoURL
			room.DealerName = tableInfo.Dealer.Name
			room.DealerUrl = tableInfo.Dealer.Avator
			room.IsDisable = 0

			room.MinBet = uint32(planItem.MinBet)
			room.MaxBet = uint32(planItem.MaxBet)
			room.Digits = uint32(planItem.Digits)

			gameRooms.GameRooms = append(gameRooms.GameRooms, &room)

			gdeskNo++
		}

		if i == int(gameGroups.GroupNum-1) {
			gameRooms.IsLast = 1
		} else {
			gameRooms.IsLast = 0
		}

		//发送每个分组内的游戏信息
		p.SendMsg(msg.MsgType_Lobby_GameRoom, &gameRooms)

	}

	return nil
}

// 发送通知消息
func (self *LobbyService) SendToPlayerNoticMessage(userId string) error {

	if len(self.GameLobby.NoticMessage) == 0 {
		return nil
	}

	p := self.PlayersImp.GetPlayer(userId)

	if p == nil {
		return errors.New("player not exist")
	}

	var noticMsg msg.NotifyLobbyNoticMessage

	noticMsg.MsgType = 0
	noticMsg.Message = self.GameLobby.NoticMessage

	p.SendMsg(msg.MsgType_Lobby_NoticeMessage, &noticMsg)

	return nil
}

// 发送每个游戏的玩家人数
func (self *LobbyService) SendToPlayerGamePlayerNum(userId string) error {

	p := self.PlayersImp.GetPlayer(userId)

	if p == nil {
		return errors.New("player not exist")
	}

	//发送各游戏的人数给玩家
	self.GameLobby.SendGamePlayerNumToPlayer(p)

	return nil
}

// 广播游戏的历史开答
func (self *LobbyService) CastToPlayerHistoryAnswers() error {

	//发送各游戏的开答列表给玩家
	for _, pp := range self.mapClientPlayer {
		self.GameLobby.SendHistoryAnswerToPlayer(pp)
	}

	return nil
}

// 广播游戏的历史开答
func (self *LobbyService) CastToPlayerOneTableHistoryAnswers(answers *msg.NotifyLobbyHistoryAnswers) error {

	//发送各游戏的开答列表给玩家
	for _, pp := range self.mapClientPlayer {
		pp.SendMsg(msg.MsgType_Lobby_GameHistoryAnswers, answers)
	}

	//广播更新的历史记录
	fmt.Printf("广播%s 的游戏历路开答路子给玩家\n", answers.GameId)

	return nil
}

// 广播发送当前在线总玩家的人数
func (self *LobbyService) CastToPlayersPlayerNum() error {

	var playerNum msg.STOC_PlayerNum

	playerNum.DeskNo = 0
	playerNum.GameId = "lobby"
	playerNum.OnlineNum = int32(self.PlayersImp.GetPlayerNum())

	clientIds := self.GetAllClientId()

	sendnum := self.GetMsgSender().CastToClient(clientIds, msg.MsgType_GameInfo_PlayNum, &playerNum)

	fmt.Printf("************** 在线人数:%d \n", sendnum)

	if sendnum == len(clientIds) {
		return nil
	}

	return errors.New("send playernum error !")
}

// 广播指定游戏桌的人数
func (self *LobbyService) CastToPlayersOneGamePlayerNum(gameId string) error {

	var playerNum msg.STOC_PlayerNum

	table, ok := self.GameLobby.GetTableInfo(gameId)

	if !ok {
		return errors.New("gameId no exist  !!!")
	}

	playerNum.DeskNo = int32(table.DeskNo)
	playerNum.GameId = table.GameId
	playerNum.OnlineNum = int32(table.PlayerNum)

	clientIds := self.GetAllClientId()

	self.GetMsgSender().CastToClient(clientIds, msg.MsgType_GameInfo_PlayNum, &playerNum)

	//log.Debugf("Broad CastToPlayersOneGamePlayerNum ")

	return nil
}

// 广播游戏状态
func (self *LobbyService) CastToPlayersGameStatus(gss *msg.NotifyLobbyGameStatus) {

	clientIds := self.GetAllClientId()

	self.GetMsgSender().CastToClient(clientIds, msg.MsgType_Lobby_GameStatus, gss)

	//	log.Debugf("Broad CastToPlayersGameStatus ")

}

// 广播当前开答的消息
func (self *LobbyService) CastToPlayersGameAnswer(gans *msg.NotifyLobbyGameAnswer) {

	clientIds := self.GetAllClientId()

	self.GetMsgSender().CastToClient(clientIds, msg.MsgType_Lobby_GameAnswer, gans)

	//	log.Debugf("Broad CastToPlayersGameAnswer ")

}
