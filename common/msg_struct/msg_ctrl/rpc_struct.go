package msg_ctrl

/*
  本文件是用于服务器之间rpc 通信的传递消息的结构体
*/

// 用于gate通知控端服务器已成功
type CtrlLoginSuccess struct {
	UserId   string `json:"userId"`   //荷官用户ID
	ClientId string `json:"clientId"` //荷官ID
}

// LoginToCtrlService rpc请求登录到控端服务
type LoginToCtrlService struct {
	GateNodeId string
	ClientId   string
	UserId     string
	Ip         string
}

// 控端登录服务器的回复结构体, 对应id = LoginResCtrlProtocol
type CtrlLoginResponse struct {
	GameId   string `json:"gameId"`   //荷官ID
	ClientId string `json:"clientId"` //荷官ID
	UserId   string `json:"userid"`   //荷官用户ID
	UserName string `json:"username"` //荷管的名称
	Level    int    `json:"level"`    //荷管的权限,前端不需要处理
	Avator   string `json:"avator"`   //荷管图片的地址
}

// 荷官的属性
type CtrlDealerInfo struct {
	Success bool //是否成功读取

	GameId string //游戏的ID

	DearlerId string //荷官的ID
	Level     int    //荷官的级别
	Name      string //荷官的姓名
	Avator    string //荷官的图标地址
}
