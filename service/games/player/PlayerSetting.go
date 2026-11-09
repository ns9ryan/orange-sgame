package player

type BetSetting struct {
	BetAreas map[int]BetAreaF //区域高低注的列表

	LowBet  int //区域最低注
	HighBet int //区域最高注
	Digits  int //小数位数
}

// 玩家的高低注的系统配置
type UserSetting struct {
	SuperVisible bool //这个按钮是否可见
	SuperSixOpt  bool //是否可操控这个按钮状态 ，百家乐游戏专用
	SuperSix     int  //转换supersix玩法  0:关闭  1：开启

	BetSet BetSetting //全部投注区域配置

}
