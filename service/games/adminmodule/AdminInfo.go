package adminmodule

// UserInfo 用户信息
type AdminInfo struct {
	Id       string //vg_users表字段：id
	UserName string
	Email    string
	Phone    string
	UserId   string // //邀请码
	Avatar   string
	RoomId   int    //房间号
	GSName   string //进入的游戏名称
}

func (user *AdminInfo) GetId() string {
	return user.Id
}

func (user *AdminInfo) GetUserID() string {
	return user.UserId
}

func (user *AdminInfo) GetRoomID() int {
	return user.RoomId
}

func (user *AdminInfo) GetGSName() string {
	return user.GSName
}
