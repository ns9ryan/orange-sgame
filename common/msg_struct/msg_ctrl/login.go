package msg_ctrl

/*
   作用:控端和服务器之间的消息结构体传递
*/

// 控端登录服务器请求,对应id = LoginReqCtrlProtocol
type CtrlLoginReq struct {
	UserId   string `json:"userid"`
	Token    string `json:"token"`
	GameName string `json:"game_name"` //登录游戏服务名称
	Room     string `json:"room"`      //对应游戏的房间号，例如1，2，3

}
