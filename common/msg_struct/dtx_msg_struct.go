package msg_struct

// 发送给控端的游戏配置
type DTXToCtrlGameConfig struct {
	DeskNo       int    `json:"desk_no"`        //桌号
	GameId       string `json:"game_id"`        //游戏的唯一ID
	TotalBetTime int    `json:"total_bet_time"` //总压注时间秒为单位
	VideoUrl     string `json:"video_url"`      //视频流的地址
}

// id =GameStatusChangeProtocol = 20103
type DTXChangeGameStatus struct {
	DeskNo     int `json:"desk_no"`     //桌号
	Set        int `json:"set"`         //场次
	Round      int `json:"round"`       //局号
	GameStatus int `json:"game_status"` //游戏状态
}

// 龙虎游戏的牌
type DTXAnswerCard struct {
	DragonCard Card `json:"dcard"` //龙的牌
	TigerCard  Card `json:"tcard"` //虎的牌
}

//func( carddt  *DTXAnswerCard )  Is

// 开答结构体
type GameOpenAnswerDTX struct {
	GameId string `json:"game_id"` //coolz add 2025.1.24
	DeskNo int    `json:"desk_no"`
	Set    int    `json:"set"`
	Round  int    `json:"round"`
	Winner int    `json:"winner"` //开奖结果  0 龙，1 虎 2 平

	DTXAnswerCard `json:"answer_card"`
}

// 历史开答的结构体
type GameDtxAnswersPeriodic struct {
	GameId     string `json:"game_id"`     //2025.1.6 add
	DeskNo     int    `json:"desk_no"`     //桌号
	Set        int    `json:"set"`         //场次
	Round      int    `json:"round"`       //当前局数
	GameStatus int    `json:"game_status"` //当前游戏状态，参考前面游戏值
	WinnerNum  int    `json:"winner_num"`  //当前开奖数组个数
	Winners    []int  `json:"winners"`     //已开奖的历史开答路子( 0 龙， 1虎 ，2平)
}
