package msg_struct_crm

const (
	CrmMsgId_Define = 1000 //消息ID起始用

	CrmMsgId_CreateMerchant = CrmMsgId_Define + 1 //创建商户消息
	CrmMsgId_UpdateMerchant = CrmMsgId_Define + 2 //更新商户的消息

	CrmMsgId_CreatePlayer = CrmMsgId_Define + 3 //创建玩家的消息
	CrmMsgId_UpdatePlayer = CrmMsgId_Define + 4 //更新玩家的消息，包括限红模板ID

	CrmMsgId_OperatePlanId = CrmMsgId_Define + 5 //操作限红模板

	CrmMsgId_OperateGameConfig = CrmMsgId_Define + 6 //修改游戏配置

	CrmMsgId_KickOutPlayer = CrmMsgId_Define + 7 //把玩家踢出去

)
