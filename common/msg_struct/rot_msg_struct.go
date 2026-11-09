package msg_struct

// // 发送给控端的游戏配置，不再每个游戏一个，公共GameToCtrlGameConfig
// type RotToCtrlGameConfig struct {
// 	DeskNo int    `json:"desk_no"` //桌号
// 	GameId string `json:"game_id"` //游戏的唯一ID

// 	TotalBetTime int    `json:"total_bet_time"` //总压注时间秒为单位
// 	VideoUrl     string `json:"video_url"`      //视频流的地址
// }

// 轮盘开奖的红黑大小单双的详情，用于数据库记录
type RotAnswerJson struct {
	Winner 		int `json:"winner"`
	IsRed		int `json:"is_red"`
	IsBig       int `json:"is_big"`
	IsSingle    int `json:"is_single"`
}

// id =GameStatusChangeProtocol = 20103
type RotChangeGameStatus struct {
	DeskNo     int `json:"desk_no"`     //桌号
	Set        int `json:"set"`         //场次
	Round      int `json:"round"`       //局号
	GameStatus int `json:"game_status"` //游戏状态
}

type RotWinner struct {
	Winner int `json:"winner"` //开奖结果  1--36
}

// 开答结构体
type GameOpenAnswerRot struct {
	GameId string `json:"game_id"`
	DeskNo int    `json:"desk_no"`
	Set    int    `json:"set"`
	Round  int    `json:"round"`

	Winner int `json:"winner"`
}

// 历史开答的结构体
type GameRotAnswersPeriodic struct {
	GameId     string `json:"game_id"`     //2025.1.6 add
	DeskNo     int    `json:"desk_no"`     //桌号
	Set        int    `json:"set"`         //场次
	Round      int    `json:"round"`       //当前局数
	GameStatus int    `json:"game_status"` //当前游戏状态，参考前面游戏值
	WinnerNum  int    `json:"winner_num"`  //当前开奖数组个数
	Winners    []int  `json:"winners"`     //已开奖的历史开答路子( 1--36 )
}
