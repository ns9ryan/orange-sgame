package msg_struct_crm

const (
	ErrParam   = 101 //参数错
	ErrIdError = 102 //消息ID错误
	ErrCluster = 103 //服务器内部出错
)

// 请求的消息头,每个结构体都有这个头
type RequestHeader struct {
	RequestId uint `json:"RequestId"` //请求的ID,每次请求这个ID值应不一样
	MsgId     uint `json:"MsgId"`     //消息Id
}

// 统一回复的消息体
type ResponseMsg struct {
	ResponseId uint `json:"ResponseId"` //哪个请求ID对应的回复
	MsgId      uint `json:"MsgId"`      //处理哪条消息的回复
	Success    uint `json:"Success"`    //处理的结果 1：成功 0：失败
	ErrCode    int  `json:"ErrCode"`    //错误码 ，上面的Err值之一

}

// 用于通知新建，更新商户的结构体
type MerchantItem struct {
	RequestHeader

	MerchantId string `json:"MerchantId"` //商户ID
	ParentId   string `json:"ParentId"`   //它上级商户的Id
	Status     int    `json:"Status"`     //商户状态
	BetPause   int    `json:"BetPause"`   //商户投注状态
}

// 模板ID
type PlanIdItem struct {
	RequestHeader

	Opt      int    `json:"Opt"` //操作动作  0:新建 1:更新 2:删除
	PlanId   string `json:"PlanId"`
	MinValue uint   `json:"MinValue"`
	MaxValue uint   `json:"MaxValue"`
	Digits   uint   `json:"Digits"`
}

// 更新玩家的数据
type PlayerItem struct {
	RequestHeader

	PlayerId string `json:"PlayerId"`
	ParentId string `json:"ParentId"`

	Status   int `json:"Status"`
	BetPause int `json:"BetPause"`
}

// 游戏配置更新
type GameConfigItem struct {
	RequestHeader

	Opt          int    `json:"Opt"`          //操作动作  0:新建 1:更新 2:删除
	GameId       string `json:"GameId"`       //游戏ID
	DeskNo       uint   `json:"DeskNo"`       //桌号
	TotalBetTime uint   `json:"TotalBetTime"` //下注时间，以秒为单位
	VideoUrl     string `json:"VideoUrl"`     //视讯游戏地址
	GameTitle    string `json:"GameTitle"`    //游戏标题

}
