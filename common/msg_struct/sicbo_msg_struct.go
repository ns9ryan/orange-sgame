package msg_struct

// id =GameStatusChangeProtocol = 20103
type SicboChangeGameStatus struct {
	DeskNo     int `json:"desk_no"`     //桌号
	Set        int `json:"set"`         //场次
	Round      int `json:"round"`       //局号
	GameStatus int `json:"game_status"` //游戏状态
}

// 发送给控端的游戏配置
type SicboWinner struct {
	Winner int `json:"winner"` //开奖结果  1--36
}

// 开答结构体
type SicboOpenAnswer struct {
	GameId string `json:"game_id"`
	DeskNo int    `json:"desk_no"`
	Set    int    `json:"set"`
	Round  int    `json:"round"`

	Winner int `json:"winner"`
}

// 历史开答的结构体
type SicboAnswersPeriodic struct {
	GameId     string `json:"game_id"`     //2025.1.6 add
	DeskNo     int    `json:"desk_no"`     //桌号
	Set        int    `json:"set"`         //场次
	Round      int    `json:"round"`       //当前局数
	GameStatus int    `json:"game_status"` //当前游戏状态，参考前面游戏值
	WinnerNum  int    `json:"winner_num"`  //当前开奖数组个数
	Winners    []int  `json:"winners"`     //已开奖的历史开答路子( 1--100 )
}
