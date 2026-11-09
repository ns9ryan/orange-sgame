package msg_struct

const (
	Ok      = 0
	Fail    = 1
	OkMsg   = "Ok"
	FailMsg = "Fail"
)

// 每张桌的游戏ID
const (
	GAMEID_BACXA    = "BACX001"
	GAMEID_BACXB    = "BACX002"
	GAMEID_BACXC    = "BACX003"
	GAMEID_DTX      = "DTX001"
	GAMEID_ROT      = "ROT"
	GAMEID_SICBO    = "SICBO"
	GAMEID_FTAN     = "FTAN"
	GAMEID_SDIE     = "SDIE"
	GAMEID_BACXMUTI = "BACXMUTI"
	GAMEID_FSC      = "FSC"
)

// 游戏类型 ,每种类型有对应的类型名称
const (
	GameTypeAll   = "ALL "
	GameTypeBacx  = "BACX"
	GameTypeDtx   = "DTX"
	GameTypeRot   = "ROT"
	GameTypeSicbo = "SICBO"
	GameTypeFtan  = "FTAN"
	GameTypeSide  = "SDIE"
	GameTypeFSC   = "FSC"
)

// 服务器类型
const (
	SReg_Game    = 1 //游戏服
	SReg_Login   = 2 //登陆服 此系统为HttpGate
	SReg_Gate    = 3 //网关服 (附加参数有值)
	SReg_Lobby   = 4 //大厅服
	SReg_Wallet  = 5 //钱包服 ( 在附加参数中，有分单，多钱包）
	SReg_Bet     = 6 //投注服
	SReg_Collect = 7 //结算服
)

const (
	HistoryAnswersPerioSize = 20 //大厅下发给玩家路子的最大长度
)

const (
	Ping = 20
	Pong = 21

	//=============================这边主要是控端消息用============================================

	LoginReqCtrlProtocol = 200 //控端登录请求
	LoginResCtrlProtocol = 201 //控端登录返回

	GameServerTimeProtocol = 20000 //游戏服务器时间
	GameMsgFailedProtocol  = 20001 //游戏消息不成功

	GameDealerJoinProtocol       = 20002 //荷官进入工作
	GameDealerJoinProtocolResult = 30002 //荷官进入工作的回复消息

	GameStatusChangeProtocol    = 20010 //游戏状态的改变,控端上传服务器
	GameDtxOpenAnswerProtocol   = 20011 //龙虎游戏的开答消息，控端上传服务器
	GameBacxOpenAnswerProtocol  = 20012 //百家乐游戏的开答消息
	GameRotOpenAnswerProtocol   = 20013 //轮盘游戏的开答消息
	GameSicboOpenAnswerProtocol = 20014 //大小游戏的开答消息
	GameFtanOpenAnswerProtocol  = 20015 //番摊游戏的开答消息
	GameSdieOpenAnswerProtocol  = 20016 //色碟游戏的开答消息
	GameFscOpenAnswerProtocol   = 20017 //鱼虾蟹游戏的开答消息

	//下面这些消息是服务器主动发给控端
	GameDtxConfigProtocol         = 20100 //龙虎游戏的配置
	GameDtxStatusProtocol         = 20101 //游戏服务器的当前状态
	GameDtxAnswerPeriodicProtocol = 20102 //龙虎游戏的开奖数组列表消息

	GameBacxConfigProtocol         = 20110 //百家乐游戏的配置
	GameBacxStatusProtocol         = 20111 //游戏服务器的当前状态
	GameBacxAnswerPeriodicProtocol = 20112 //百家乐游戏的开奖数组列表消息

	GameRotConfigProtocol         = 20120 //轮盘游戏的配置
	GameRotStatusProtocol         = 20121 //轮盘游戏服务器的当前状态
	GameRotAnswerPeriodicProtocol = 20122 //轮盘游戏的开奖数组列表消息

	GameSicboConfigProtocol         = 20130 //大小游戏的配置
	GameSicboStatusProtocol         = 20131 //大小游戏服务器的当前状态
	GameSicboAnswerPeriodicProtocol = 20132 //大小游戏的开奖数组列表消息

	GameFtanConfigProtocol         = 20140 //番摊游戏的配置
	GameFtanStatusProtocol         = 20141 //番摊游戏服务器的当前状态
	GameFtanAnswerPeriodicProtocol = 20142 //番摊游戏的开奖数组列表消息

	GameSdieConfigProtocol         = 20150 //色碟游戏的配置
	GameSdieStatusProtocol         = 20151 //色碟游戏服务器的当前状态
	GameSdieAnswerPeriodicProtocol = 20152 //色碟游戏的开奖数组列表消息

	GameFscConfigProtocol         = 20160 //鱼虾蟹游戏的配置
	GameFscStatusProtocol         = 20161 //鱼虾蟹游戏服务器的当前状态
	GameFscAnswerPeriodicProtocol = 20162 //鱼虾蟹游戏的开奖数组列表消息

)
