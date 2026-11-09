package player

// UserInfo 用户信息
type UserInfo struct {
	Id       string //vg_users表字段：id
	UserId   string
	UserName string
}

func (user *UserInfo) GetId() string {
	return user.Id
}

func (user *UserInfo) GetUserID() string {
	return user.UserId
}
