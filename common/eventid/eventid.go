package eventid

import "github.com/duanhf2012/origin/v2/event"

const (
	//自定义事件类型，必需从event.Sys_Event_User_Define开始
	//event.Sys_Event_User_Define以内给系统预留
	//以下事件用于GameTable和它的上级Service之间的调用

	Event_Broad_GameStatus            event.EventType = event.Sys_Event_User_Define + 1 //游戏状态改变事件
	Event_Broad_Answer                event.EventType = event.Sys_Event_User_Define + 2 //游戏开奖
	Event_Broad_History_answers       event.EventType = event.Sys_Event_User_Define + 3 //游戏开奖历史
	Event_NotifyLobby_History_Answers event.EventType = event.Sys_Event_User_Define + 4 //更新大厅历史开答表消息
	Event_Player_Bet                  event.EventType = event.Sys_Event_User_Define + 5 //玩家下注
	Event_Answer_Settle                               = event.Sys_Event_User_Define + 6 //开奖结算
	Event_Broad_PlayerCard                            = event.Sys_Event_User_Define + 7 //广播玩家帐号信息，主要是余额

	Event_Player_CancelBet = event.Sys_Event_User_Define + 8  //广播玩家帐号信息，主要是余额
	Event_Player_TipMoney  = event.Sys_Event_User_Define + 9  //打赏给荷官
	Event_Update_OnlineNum = event.Sys_Event_User_Define + 10 //更新玩家在线人数
	Event_Player_TalkChat  = event.Sys_Event_User_Define + 11 //玩家聊天

	Event_Core_CreateMerChant       = event.Sys_Event_User_Define + 20 //新建一个商户
	Event_Core_UpdateMerChant       = event.Sys_Event_User_Define + 21 //更新一个商户
	Event_Core_UpdateTemplatePlanId = event.Sys_Event_User_Define + 22 //更新限红模版ID
	Event_Core_OperateGameConfig    = event.Sys_Event_User_Define + 23 //更新一个游戏设置
)
