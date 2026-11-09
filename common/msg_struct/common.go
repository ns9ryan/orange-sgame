package msg_struct

// RetMsg 返回客户端统一结构
type RetMsg struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type MsgPong struct {
	NowTime string `json:"now_time"`
}

type MsgPing struct {
	NowTime string `json:"now_time"`
}

// ================================coolz=========================================================

const (
	GameStatusError        = -1 //状态错误，更新状态不成功时用
	GameStatusSleep        = 0  //休场
	GameStatusInterMission = 1  //洗牌
	GameStatusBet          = 2  //下注开始
	GameStatusStopBet      = 3  //停止下注
)

// 服务器时间
type GameServerTime struct {
	ServerTime string `json:"server_time"` //游戏服务器时间, //年月日，小时，分，秒，纳秒)
}

// 荷官加入此桌台
type GameDealerJoin struct {
	UserId string `json:"user_id"`
}

// 荷官加入此桌台的回复
type GameDealerJoinResult struct {
	RetCode string `json:"ret"` //  "ok" ,表示加入成功,"failed",不成功

	UserId     string `json:"user_id"`    //荷官ID
	UserName   string `json:"user_name"`  //荷官名字
	Supervisor string `json:"supervisor"` //上级主管名
}

// 发送给控端的游戏配置
type GameToCtrlGameConfig struct {
	DeskNo       int    `json:"desk_no"`        //桌号
	GameId       string `json:"game_id"`        //游戏的唯一ID
	TotalBetTime int    `json:"total_bet_time"` //总压注时间秒为单位
	VideoUrl     string `json:"video_url"`      //视频流的地址
}

// 游戏状态情况
type GameStatusMsg struct {
	GameId string `json:"game_id"` //游戏ID
	DeskNo int    `json:"desk_no"` //桌号
	Set    int    `json:"set"`     //场次
	Round  int    `json:"round"`   //局号

	RoundId string `json:"round_id"` //游戏全局的id  2025.3.6增加

	GameStatus  int `json:"game_status"`  //游戏状态
	CurrentTime int `json:"current_time"` //当前时间，表示游戏开始下注已经过的时间，秒数
	IsAnswered  int `json:"is_answered"`  //是否已开答
}

// 状态改变不成功
type GameChangeFailed struct {
	GameId string `json:"game_id"` //游戏ID  2025.1.22 add
	DeskNo int    `json:"desk_no"` //桌号
	Set    int    `json:"set"`     //场次
	Round  int    `json:"round"`   //局号
	Msgid  int    `json:"msgid"`   //哪条消息更新不成功
	Code   int    `json:"code"`    //错误码
}

type Card struct {
	Used   int  `json:"used"`   //是否使用0，没有使用，1 启用 2025.1.22 add
	Number byte `json:"number"` //牌的数值 （ 1---13）
	Color  byte `json:"color"`  //牌的色彩  （1---- 4）
}

// 牌数值是否想同,color表示是否检测色彩一样
func (card *Card) IsSame(ccard *Card, color bool) bool {

	if color == true {

		if (card.Number == ccard.Number) && (card.Color == ccard.Color) {
			return true
		}

		return false

	} else if card.Number == ccard.Number {

		return true
	}

	return false
}

// 牌的基本类msgid = GameDtxAnswerPeriodicProtocol
type GameAnswerCardBase struct {
	Winner   byte //胜者
	PPair    bool //闲对子
	BPair    bool //庄对子
	SuperSix bool //是否supersix
}

// 荷光的属性
type DealerInfo struct {
	DealerId string //荷官的ID
	Level    int    //荷官的级别
	Name     string //荷官的姓名
	Avator   string //荷官的图标地址
}

// 历史开答的结构体
type GameAnswersPerio struct {
	GameId string //游戏的ID
	DeskNo int    `json:"desk_no"` //桌号

	Set   int `json:"set"`   //场次
	Round int `json:"round"` //当前局数

	GameStatus int `json:"game_status"` //当前游戏状态，参考前面游戏值
	WinnerNum  int `json:"winner_num"`  //当前开奖数组个数

	GameType int //游戏类型

	Winners interface{} //已开奖的历史开答路子
}

// 每个游戏的状态
type LobbyGameSession struct {
	GameId string //游戏的ID
	DeskNo int    //桌号

	Set   int //场次
	Round int //当前局数

	GameStatus int //当前游戏状态，参考前面游戏值

	TotalBetTime int //总的投注时间
	StartBetTime int //开始下注的时间

	WinnerNum int `json:"winner_num"` //当前开奖数组个数

	GameType int //游戏类型

	Winners interface{} //已开奖的历史开答路子
}

// 玩家输入账号密码登录游戏，或者通过resource_id快速进入游戏的结构体
type LoginEnterLobby struct {
	Token      		string `json:"Token"`
	UserId     		string `json:"UserId"`
	WsAddr     		string `json:"WsAddr"`
	Language   		string `json:"Language"`
	DeviceType 		int32  `json:"DeviceType"`

	DirectIntoGame 	bool 	`json:"DirectIntoGame"` //是否直接进入游戏
	GroupId 		string  `json:"GroupId"`
	GameId 		    string  `json:"GameId"`
	DeskNo          int32   `json:"DeskNo"`
	BackUrl 		string  `json:"BackUrl"`
}
