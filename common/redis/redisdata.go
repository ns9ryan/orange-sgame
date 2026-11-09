package redisId

// redis中存储玩家的信息数据
type RedisUserInfo struct {
	UserId    string `json:"userid"`
	UserName  string `json:"username"`
	Password  string `json:"password"`
	AvatarUrl string `json:"avatar_url"`

	TopMerchantId string `json:"TopMerchantId"` //最上级商户的ID
	MerchantId    string `json:"merchantid"`    //上级商户的ID
	MerchantCode  string `json:"merchantcode"`  //商户的Code
	RootKey       uint   `json:"rootkey"`

	Email string `json:"email"`
	Phone string `json:"phone"`

	UserType   uint `json:"user_type"`   //玩家种类
	MoneyType  uint `json:"money_type"`  //玩家类型
	WalletType uint `json:"wallet_type"` //钱包类型
	Blocked    bool `json:"block"`       //是否锁定
	Paused     bool `json:"paused"`      //

	LoginIpFirst string `json:"login_ip_first"` //首次进入的IP
	LoginIP      string `json:"loginip"`        //最近登陆的IP

	MaxWin uint `json:"maxwin"` //最大盈额
	Level  uint `json:"level"`  //玩家级别

	ChipsList          string `json:"chip_list"`  //它的抽码列表
	BetLimitTemplateId string `json:"templateid"` //投注区域的模板

	MinTip uint `json:"mintip"` //最小打赏金额
	MaxTip uint `json:"maxtip"` //最大打赏金额
	Digits uint `json:"digits"` //打赏金额的小数位

	UpdateTime int64 `json:"updateTime"` //更新的时间，以UNIX时间为准
}

// 玩家的钱包数据
type RedisUserWallet struct {
	UserId      string `json:"userid"`
	Credit      uint   `json:"credit"`   //玩家余额
	Currency    string `json:"currency"` //货币
	RewardAmout uint   `json:"reward"`   //打赏的金额

	Digits uint `json:"digits"` //小数位数
}

// 玩家的投注帐目数据
type RedisUserAccount struct {
	UserId string `json:"userid"`
	Digits int    `json:"digits"` //小数位数

	TotalBet          uint `json:"total_bet"`           //总投注
	TotalBetEffective uint `json:"total_bet_effective"` //有效投注
	TotalWater        uint `json:"total_water"`         //总的洗码
	TotalWin          int  `json:"total_win"`           //总洗码

	LastBet          uint `json:"last_bet"`              //最近一天投注额
	LastBetEffective uint `json:"total_water_effective"` //有效投注
	LastWater        uint `json:"last_water"`            //

	LastWin int `json:"last_win"`
}

// 商户的信息数据
type RedisMerchantInfo struct {
	MerchantId   string `json:"merchantid"`   //商户的ID
	MerchantName string `json:"merchantname"` //商户的名字

	WalletType int `json:"wallet_type"` //钱包的类型

}

// 抽码的列表信息
type RedisChipsList struct {
	Id         int
	ChipValue  int
	ChipDigits int
	ChipFValue float32
	ChipName   string
	Status     int
}
