package fscservice

import (
	"errors"
	"fmt"
	"github.com/duanhf2012/origin/v2/event"
	"github.com/duanhf2012/origin/v2/log"
	"videogame/common/eventid"
	"videogame/common/proto/msg"
	"videogame/common/proto/rpc"
)

// 用于投注服传递投注的RPC调用
func (self *FscService) RPCBetGame(betMsg *rpc.BetToGameRPC, ret *rpc.Int) error {

	if betMsg.BetMode == 0 { //游戏内的投注

		player := self.PlayersImp.GetPlayer(betMsg.UserId)

		fmt.Printf("FscService :%v  PushEvent  \n", *betMsg)

		//若没有找到
		if player == nil {
			//需要重新建一个帐号
		}

		//帐号被锁
		if player.Blocked == true {
			return errors.New("card is blocked!!!")
		}

		//帐号被暂停
		if player.Paused == true {
			return errors.New("card is paused!!!")
		}

		//检测是否在投注状态
		if !self.GameTable.GameInfo.IsBetting() {

			var betResult msg.STOC_BetGame_Result

			betResult.Success = 1
			betResult.GameId = betMsg.GameId
			betResult.BetMode = int32(betMsg.BetMode)
			betResult.DeskNo = int32(self.GameTable.GetDeskNo())
			betResult.OrderKey = uint32(betMsg.OrderKey)

			betResult.Balance = int32(player.GetCredit())
			betResult.Digits = int32(player.GetDigits())
			betResult.ErrCode = int32(msg.ErrCode_BetErr_NoBettingStatus)

			player.SendMsg(msg.MsgType_GameTable_BetResult, &betResult)

			log.Errorf("投注时间不对，已停止投注:gameId:%s ", betMsg.GameId)

			return nil
		}

		//投递到投注队列
		self.GetEventHandler().NotifyEvent(&event.Event{
			Type: eventid.Event_Player_Bet,
			Data: betMsg,
		})

	} else {

	}

	return nil

}

// 用于投注服传递投注的RPC调用
func (self *FscService) RPCCancelBetGame(betMsg *rpc.BetToGameRPC, ret *rpc.Int) error {

	if betMsg.BetMode == 0 { //游戏内的投注

		player := self.PlayersImp.GetPlayer(betMsg.UserId)

		fmt.Printf("LongHuService :%v  PushEvent  \n", betMsg)

		//若没有找到
		if player == nil {
			//需要重新建一个帐号
		}

		//帐号被锁
		if player.Blocked == true {
			return errors.New("card is blocked!!!")
		}

		//帐号被暂停
		if player.Paused == true {
			return errors.New("card is paused!!!")
		}

		//检测是否在投注状态
		if !self.GameTable.GameInfo.IsBetting() {

			var betResult msg.STOC_CancelBetGame_Result

			betResult.Success = 1
			betResult.GameId = betMsg.GameId
			betResult.BetMode = int32(betMsg.BetMode)
			betResult.DeskNo = int32(self.GameTable.GetDeskNo())
			betResult.OrderKey = uint32(betMsg.OrderKey)

			betResult.Balance = int32(player.GetCredit())
			betResult.Digits = int32(player.GetDigits())
			betResult.ErrCode = int32(msg.ErrCode_BetErr_NoBettingStatus)

			player.SendMsg(msg.MsgType_GameTable_CancelBetResult, &betResult)

			log.Errorf("撤消投注时间不对，已停止投注:gameId:%s ", betMsg.GameId)

			return nil
		}

		//投递到投注队列
		self.GetEventHandler().NotifyEvent(&event.Event{
			Type: eventid.Event_Player_CancelBet,
			Data: betMsg,
		})

	} else {

	}

	return nil

}

// 用于投注服传递打赏的RPC调用
func (self *FscService) RPCTipGame(tipMsg *rpc.TipToGameRPC, ret *rpc.Int) error {

	player := self.PlayersImp.GetPlayer(tipMsg.UserId)

	fmt.Printf("**************BacxService :---> RPCTipGame :%v    \n", tipMsg)

	//若没有找到
	if player == nil {
		//需要重新建一个帐号
		return errors.New("player is nil")
	}

	//帐号被锁
	if player.Blocked == true {

		fmt.Printf("**************BacxService :---> RPCTipGame  cardIs is Blocked \n ")

		var tipResult msg.STOC_TipMoneyResult

		tipResult.Success = 1
		tipResult.GameId = tipMsg.GameId
		tipResult.DeskNo = int32(self.GameTable.GetDeskNo())
		tipResult.OrderKey = uint32(tipMsg.OrderKey)
		tipResult.TipMoney = tipMsg.TipPoint
		tipResult.Digits = tipMsg.Digits

		tipResult.ErrCode = int32(msg.ErrCode_BetErr_CardBlocked)

		player.SendMsg(msg.MsgType_GameTable_TipMoneyResult, &tipResult)

		return errors.New("card is blocked!!!")
	}

	//投递到投注队列
	self.GetEventHandler().NotifyEvent(&event.Event{
		Type: eventid.Event_Player_TipMoney,
		Data: tipMsg,
	})

	return nil

}

// 用于玩家之间的聊天
func (self *FscService) RPCTalkGame(talkMsg *rpc.GameTalkChat, ret *rpc.Int) error {

	//投递到事件处理队列
	self.GetEventHandler().NotifyEvent(&event.Event{
		Type: eventid.Event_Player_TalkChat,
		Data: talkMsg,
	})

	return nil
}
