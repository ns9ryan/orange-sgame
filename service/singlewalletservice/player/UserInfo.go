package player

// UserInfo 用户信息
type UserInfo struct {
	Id       	string //vg_users表字段：id
	UserId   	string
	UserName 	string

	Code 		string
	MerchantKey string

	//只是保存Redis检查用，后续高并发时可能要去掉
	Credit		uint
}

func (user *UserInfo) GetId() string {
	return user.Id
}

func (user *UserInfo) GetUserID() string {
	return user.UserId
}
