package ftanservice

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
func (self *FtanService) RPCBetGame(betMsg *rpc.BetToGameRPC, ret *rpc.Int) error {

	if betMsg.BetMode == 0 { //游戏内的投注

		player := self.PlayersImp.GetPlayer(betMsg.UserId)

		//	fmt.Printf("FtanService :%v  PushEvent  \n", *betMsg)

		//若没有找到
		if player == nil {

			//需要重新建一个帐号

			log.Errorf("FtanService-->RPCBetGame 内存中无法查找到玩家:%s ", betMsg.UserId)

			var betResult msg.STOC_BetGame_Result

			betResult.Success = 1
			betResult.GameId = betMsg.GameId
			betResult.BetMode = int32(betMsg.BetMode)
			betResult.DeskNo = int32(self.GameTable.GetDeskNo())
			betResult.OrderKey = uint32(betMsg.OrderKey)

			betResult.Balance = int32(player.GetCredit())
			betResult.Digits = int32(player.GetDigits())
			betResult.ErrCode = int32(msg.ErrCode_BetErr_ServerPlayerIdErr)

			player.SendMsg(msg.MsgType_GameTable_BetResult, &betResult)

			return errors.New("player is nil")
		}

		//帐号被锁
		if player.Blocked == true {
			//return errors.New("card is blocked!!!")

			log.Warnf("FtanService-->RPCBetGame 玩家被锁，无法投注:%s", betMsg.UserId)

			var betResult msg.STOC_BetGame_Result

			betResult.Success = 1
			betResult.GameId = betMsg.GameId
			betResult.BetMode = int32(betMsg.BetMode)
			betResult.DeskNo = int32(self.GameTable.GetDeskNo())
			betResult.OrderKey = uint32(betMsg.OrderKey)

			betResult.Balance = int32(player.GetCredit())
			betResult.Digits = int32(player.GetDigits())
			betResult.ErrCode = int32(msg.ErrCode_BetErr_CardBlocked)

			player.SendMsg(msg.MsgType_GameTable_BetResult, &betResult)

			return nil
		}

		//帐号被暂停
		if player.Paused == true {

			//return errors.New("card is paused!!!")

			var betResult msg.STOC_BetGame_Result

			betResult.Success = 1
			betResult.GameId = betMsg.GameId
			betResult.BetMode = int32(betMsg.BetMode)
			betResult.DeskNo = int32(self.GameTable.GetDeskNo())
			betResult.OrderKey = uint32(betMsg.OrderKey)

			betResult.Balance = int32(player.GetCredit())
			betResult.Digits = int32(player.GetDigits())
			betResult.ErrCode = int32(msg.ErrCode_BetErr_CardPauseBet)

			player.SendMsg(msg.MsgType_GameTable_BetResult, &betResult)

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

			//log.Errorf("投注时间不对，已停止投注:gameId:%s ", betMsg.GameId)

			log.Warnf("投注时间不对，已停止投注:gameId:%s  userId:%s", betMsg.GameId, betMsg.UserId)

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
func (self *FtanService) RPCCancelBetGame(betMsg *rpc.BetToGameRPC, ret *rpc.Int) error {

	if betMsg.BetMode == 0 { //游戏内的投注

		player := self.PlayersImp.GetPlayer(betMsg.UserId)

		fmt.Printf("FtanService :%v  PushEvent  \n", betMsg)

		//若没有找到
		if player == nil {

			var betResult msg.STOC_CancelBetGame_Result

			betResult.Success = 1
			betResult.GameId = betMsg.GameId
			betResult.BetMode = int32(betMsg.BetMode)
			betResult.DeskNo = int32(self.GameTable.GetDeskNo())
			betResult.OrderKey = uint32(betMsg.OrderKey)

			betResult.Balance = int32(player.GetCredit())
			betResult.Digits = int32(player.GetDigits())
			betResult.ErrCode = int32(msg.ErrCode_BetErr_ServerPlayerIdErr)

			player.SendMsg(msg.MsgType_GameTable_CancelBetResult, &betResult)

			log.Errorf("撤消投注，内存查无此玩家:gameId:%s  userId :%s ", betMsg.GameId, betMsg.UserId)

			return errors.New("player is nil")
		}

		//帐号被锁
		if player.Blocked == true {

			var betResult msg.STOC_CancelBetGame_Result

			betResult.Success = 1
			betResult.GameId = betMsg.GameId
			betResult.BetMode = int32(betMsg.BetMode)
			betResult.DeskNo = int32(self.GameTable.GetDeskNo())
			betResult.OrderKey = uint32(betMsg.OrderKey)

			betResult.Balance = int32(player.GetCredit())
			betResult.Digits = int32(player.GetDigits())
			betResult.ErrCode = int32(msg.ErrCode_BetErr_CardBlocked)

			player.SendMsg(msg.MsgType_GameTable_CancelBetResult, &betResult)

			log.Warnf("撤消投注，帐号被锁了:gameId:%s  userId :%s ", betMsg.GameId, betMsg.UserId)

			//return errors.New("card is blocked!!!")
			return nil
		}

		//帐号被暂停
		if player.Paused == true {

			//return errors.New("card is paused!!!")
			var betResult msg.STOC_CancelBetGame_Result

			betResult.Success = 1
			betResult.GameId = betMsg.GameId
			betResult.BetMode = int32(betMsg.BetMode)
			betResult.DeskNo = int32(self.GameTable.GetDeskNo())
			betResult.OrderKey = uint32(betMsg.OrderKey)

			betResult.Balance = int32(player.GetCredit())
			betResult.Digits = int32(player.GetDigits())
			betResult.ErrCode = int32(msg.ErrCode_BetErr_CardPauseBet)

			player.SendMsg(msg.MsgType_GameTable_CancelBetResult, &betResult)

			log.Warnf("**撤消投注，帐号被暂停投注了！:gameId:%s  userId :%s ", betMsg.GameId, betMsg.UserId)

			return nil
		}

		//检测是否在投注状态
		if !self.GameTable.GameInfo.IsBetting() {

			var betResult msg.STOC_CancelBetGame_Result

			betResult.Success = 1
			betResult.GameId = betMsg.GameId
			betResult.BetMode = betMsg.BetMode
			betResult.DeskNo = int32(self.GameTable.GetDeskNo())
			betResult.OrderKey = uint32(betMsg.OrderKey)

			betResult.Balance = int32(player.GetCredit())
			betResult.Digits = int32(player.GetDigits())
			betResult.ErrCode = int32(msg.ErrCode_BetErr_NoBettingStatus)

			player.SendMsg(msg.MsgType_GameTable_CancelBetResult, &betResult)

			log.Warnf("**撤消投注时间不对，已停止投注:gameId:%s ,userId:%s ", betMsg.GameId, betMsg.UserId)

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
func (self *FtanService) RPCTipGame(tipMsg *rpc.TipToGameRPC, ret *rpc.Int) error {

	player := self.PlayersImp.GetPlayer(tipMsg.UserId)

	//fmt.Printf("**************FtanService :---> RPCTipGame :%v    \n", tipMsg)

	//若没有找到
	if player == nil {
		//需要重新建一个帐号
		var tipResult msg.STOC_TipMoneyResult

		tipResult.Success = 1
		tipResult.GameId = tipMsg.GameId
		tipResult.DeskNo = int32(self.GameTable.GetDeskNo())
		tipResult.OrderKey = uint32(tipMsg.OrderKey)
		tipResult.TipMoney = tipMsg.TipPoint
		tipResult.Digits = tipMsg.Digits

		tipResult.ErrCode = int32(msg.ErrCode_BetErr_ServerPlayerIdErr)

		player.SendMsg(msg.MsgType_GameTable_TipMoneyResult, &tipResult)

		return errors.New("FtanService-->RPCTipGame  player is nil")

	}

	//帐号被锁
	if player.Blocked == true {

		log.Warnf("**************FtanService :---> RPCTipGame  cardIs:%S is Blocked \n ", player.Id)

		var tipResult msg.STOC_TipMoneyResult

		tipResult.Success = 1
		tipResult.GameId = tipMsg.GameId
		tipResult.DeskNo = int32(self.GameTable.GetDeskNo())
		tipResult.OrderKey = uint32(tipMsg.OrderKey)
		tipResult.TipMoney = tipMsg.TipPoint
		tipResult.Digits = tipMsg.Digits

		tipResult.ErrCode = int32(msg.ErrCode_BetErr_CardBlocked)

		player.SendMsg(msg.MsgType_GameTable_TipMoneyResult, &tipResult)

		return nil
	}

	//投递到投注队列
	self.GetEventHandler().NotifyEvent(&event.Event{
		Type: eventid.Event_Player_TipMoney,
		Data: tipMsg,
	})

	return nil

}

// 用于玩家之间的聊天
func (self *FtanService) RPCTalkGame(talkMsg *rpc.GameTalkChat, ret *rpc.Int) error {

	//投递到事件处理队列
	self.GetEventHandler().NotifyEvent(&event.Event{
		Type: eventid.Event_Player_TalkChat,
		Data: talkMsg,
	})

	return nil
}
