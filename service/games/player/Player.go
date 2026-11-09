package player

import (
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"google.golang.org/protobuf/proto"
	"time"
	betOdds "videogame/common/betodds"
	"videogame/common/msg_struct"
	"videogame/common/proto/msg"
	"videogame/service/interfacedef"
)

type Player struct {
	interfacedef.IPlayerTimer
	interfacedef.IGSService
	interfacedef.IMsgSender
	interfacedef.IMsgReceiver
	interfacedef.IPlayer

	PoolObj
	DataInfo
	UserInfo

	pingTime time.Time //ping时间
	msgTime  time.Time //最近一次消息的时间

	TickerId uint64 //用于计数器的ID

}

// 钱包数据
type OptData struct {
	UserId string //用户id
	GameId string //增加gameId

	Point  uint //操作金额
	Digits uint //小数位

	Opt       int    //0加点，1减点
	DeskNo    int    //桌号，1-3百家乐4龙虎5轮盘等
	TransType int    //订单类型，0下注订单，1取消下注订单  2小费订单，3结算订单，4其他订单，最终会int转为订单说明文字调用商户API
	TransId   string //订单id，下注用OrkeyUid，小费用小费生成id， 取消下注用OrkeyUid，结算用roundId
	RoundId   string //对局id
}

// 主要用于通知商户单钱包服务，游戏开答了，让他们更新相关玩家订单的状态
type OpenAnswerData struct {
	UserId  string //用户id
	DeskNo  int    //桌号，1-3百家乐4龙虎5轮盘等
	RoundId string //对局id
}

func (p *Player) Init(id string, sender interfacedef.IMsgSender, gsService interfacedef.IGSService, playerTimer interfacedef.IPlayerTimer, receive interfacedef.IMsgReceiver) {
	p.IMsgSender = sender
	p.IGSService = gsService
	p.IPlayerTimer = playerTimer
	p.IMsgReceiver = receive
	p.Id = id
	p.UserId = id

	//这个一定要初始化，不然后面写入出错
	p.SettingData.BetSet.BetAreas = make(map[int]BetAreaF)

	//测试用
	p.SetSuperSixOpt(true)

	p.UpdatePingTime()

}

func (p *Player) Reset() {
	*p = Player{}
}

// 初始化游戏压注区
func (p *Player) InitBetLimit(gameTypye string, cbInit func()) {

	if gameTypye == msg_struct.GameTypeAll {

		p.ReadDtxLimit()
		p.ReadBacxLimit()
		p.ReadRotLimit()
		p.ReadSicboLimit()
		p.ReadFtanLimit()
		p.ReadSdieLimit()
		p.ReadFscLimit()

	} else if gameTypye == msg_struct.GameTypeBacx {

		p.ReadBacxLimit()

	} else if gameTypye == msg_struct.GameTypeDtx {
		p.ReadDtxLimit()
	} else if gameTypye == msg_struct.GameTypeRot {
		p.ReadRotLimit()
	} else if gameTypye == msg_struct.GameTypeSicbo {
		p.ReadSicboLimit()
	} else if gameTypye == msg_struct.GameTypeFtan {
		p.ReadFtanLimit()
	} else if gameTypye == msg_struct.GameTypeSide {
		p.ReadSdieLimit()
	} else if gameTypye == msg_struct.GameTypeFSC {
		p.ReadFscLimit()
	}

	//自定议的高低注
	if cbInit != nil {
		cbInit()
	}

}

func (p *Player) SendMsg(msgType msg.MsgType, message proto.Message) int {
	return p.IMsgSender.SendToClient(p.GetClientId(), msgType, message)
}

func (p *Player) Destroy() {

	p.DestroyPlayer(p.GetId())

}

func (p *Player) ReadDtxLimit() {

	if p.SettingData.BetSet.BetAreas == nil {
		return
	}

	//从redis读取高低注，或没有则从数据库读取
	p.SettingData.BetSet.BetAreas[betOdds.Bet_Dtx_Dragon] = BetAreaF{betOdds.Bet_Dtx_Dragon, 0.1, 100}
	p.SettingData.BetSet.BetAreas[betOdds.Bet_Dtx_Tiger] = BetAreaF{betOdds.Bet_Dtx_Tiger, 0.1, 100}
	p.SettingData.BetSet.BetAreas[betOdds.Bet_Dtx_Tie] = BetAreaF{betOdds.Bet_Dtx_Tie, 0.1, 100}

	p.SettingData.BetSet.Digits = int(p.WalletData.GetDigits())

	p.SettingData.BetSet.LowBet = int(0.1 * float64(p.SettingData.BetSet.Digits))
	p.SettingData.BetSet.HighBet = 100 * p.SettingData.BetSet.Digits

}

func (p *Player) ReadBacxLimit() {

	//最后需要改成从redis,数据库 读取游戏玩家的配置
	p.SettingData.BetSet.BetAreas[betOdds.Bet_Bacx_Player] = BetAreaF{betOdds.Bet_Bacx_Player, 0.1, 100}
	p.SettingData.BetSet.BetAreas[betOdds.Bet_Bacx_Banker] = BetAreaF{betOdds.Bet_Bacx_Banker, 0.1, 100}
	p.SettingData.BetSet.BetAreas[betOdds.Bet_Bacx_Tie] = BetAreaF{betOdds.Bet_Bacx_Tie, 0.1, 100}
	p.SettingData.BetSet.BetAreas[betOdds.Bet_Bacx_PPair] = BetAreaF{betOdds.Bet_Bacx_PPair, 0.1, 100}
	p.SettingData.BetSet.BetAreas[betOdds.Bet_Bacx_BPair] = BetAreaF{betOdds.Bet_Bacx_BPair, 0.1, 100}
	p.SettingData.BetSet.BetAreas[betOdds.Bet_Bacx_BankerNocomm] = BetAreaF{betOdds.Bet_Bacx_BankerNocomm, 0.1, 100}
	p.SettingData.BetSet.BetAreas[betOdds.Bet_Bacx_SuperSix] = BetAreaF{betOdds.Bet_Bacx_SuperSix, 0.1, 1000}
	p.SettingData.BetSet.BetAreas[betOdds.Bet_Bacx_AnyPair] = BetAreaF{betOdds.Bet_Bacx_AnyPair, 0.1, 1000}
	p.SettingData.BetSet.BetAreas[betOdds.Bet_Bacx_PerfectPair] = BetAreaF{betOdds.Bet_Bacx_PerfectPair, 0.1, 1000}
	p.SettingData.BetSet.BetAreas[betOdds.Bet_Bacx_Big] = BetAreaF{betOdds.Bet_Bacx_Big, 0.1, 1000}
	p.SettingData.BetSet.BetAreas[betOdds.Bet_Bacx_Small] = BetAreaF{betOdds.Bet_Bacx_Small, 0.1, 1000}
	p.SettingData.BetSet.BetAreas[betOdds.Bet_Bacx_PlayerBonus] = BetAreaF{betOdds.Bet_Bacx_PlayerBonus, 0.1, 1000}
	p.SettingData.BetSet.BetAreas[betOdds.Bet_Bacx_BankerBonus] = BetAreaF{betOdds.Bet_Bacx_BankerBonus, 0.1, 1000}
	p.SettingData.BetSet.Digits = int(p.WalletData.GetDigits())

	p.SettingData.BetSet.LowBet = int(0.1 * float64(p.SettingData.BetSet.Digits))
	p.SettingData.BetSet.HighBet = 100 * p.SettingData.BetSet.Digits

}

func (p *Player) ReadRotLimit() {

	for i := betOdds.Bet_Rot_Start; i <= betOdds.Bet_Rot_EighteenDouble; i++ {
		p.SettingData.BetSet.BetAreas[i] = BetAreaF{uint32(i), 0.1, 100}
	}

	for i := betOdds.Bet_Rot_NeighbourTwo0; i <= betOdds.Bet_Rot_NeighbourTwo26; i++ {
		p.SettingData.BetSet.BetAreas[i] = BetAreaF{uint32(i), 0.5, 100}
	}

	p.SettingData.BetSet.BetAreas[betOdds.Bet_Rot_Tiers] = BetAreaF{uint32(betOdds.Bet_Rot_Tiers), 0.6, 100}
	p.SettingData.BetSet.BetAreas[betOdds.Bet_Rot_Orphelins] = BetAreaF{uint32(betOdds.Bet_Rot_Orphelins), 0.5, 100}
	p.SettingData.BetSet.BetAreas[betOdds.Bet_Rot_Voisins] = BetAreaF{uint32(betOdds.Bet_Rot_Voisins), 0.9, 100}
	p.SettingData.BetSet.BetAreas[betOdds.Bet_Rot_Jeu0] = BetAreaF{uint32(betOdds.Bet_Rot_Jeu0), 0.4, 100}

	p.SettingData.BetSet.Digits = int(p.WalletData.GetDigits())

	p.SettingData.BetSet.LowBet = int(0.1 * float64(p.SettingData.BetSet.Digits))
	p.SettingData.BetSet.HighBet = 100 * p.SettingData.BetSet.Digits

}

func (p *Player) ReadSicboLimit() {

	for i := betOdds.Bet_Sicbo_Start; i <= betOdds.Bet_Sicbo_End; i++ {
		p.SettingData.BetSet.BetAreas[i] = BetAreaF{uint32(i), 10, 100}
	}

	p.SettingData.BetSet.Digits = int(p.WalletData.GetDigits())

	p.SettingData.BetSet.LowBet = int(0.1 * float64(p.SettingData.BetSet.Digits))
	p.SettingData.BetSet.HighBet = 100 * p.SettingData.BetSet.Digits

}

func (p *Player) ReadFtanLimit() {

	for i := betOdds.Bet_Ftan_Start; i <= betOdds.Bet_Ftan_End; i++ {
		p.SettingData.BetSet.BetAreas[i] = BetAreaF{uint32(i), 10, 100}
	}

	p.SettingData.BetSet.Digits = int(p.WalletData.GetDigits())

	p.SettingData.BetSet.LowBet = int(0.1 * float64((p.SettingData.BetSet.Digits)))
	p.SettingData.BetSet.HighBet = 100 * p.SettingData.BetSet.Digits

}

func (p *Player) ReadSdieLimit() {
	for i := betOdds.Bet_Sdie_Start; i <= betOdds.Bet_Sdie_End; i++ {
		p.SettingData.BetSet.BetAreas[i] = BetAreaF{uint32(i), 10, 100}
	}

	p.SettingData.BetSet.Digits = int(p.WalletData.GetDigits())

	p.SettingData.BetSet.LowBet = int(0.1 * float64((p.SettingData.BetSet.Digits)))
	p.SettingData.BetSet.HighBet = 100 * p.SettingData.BetSet.Digits

}

func (p *Player) ReadFscLimit() {
	for i := betOdds.Bet_Fsc_Start; i <= betOdds.Bet_Fsc_End; i++ {
		p.SettingData.BetSet.BetAreas[i] = BetAreaF{uint32(i), 10, 100}
	}

	p.SettingData.BetSet.Digits = int(p.WalletData.GetDigits())

	p.SettingData.BetSet.LowBet = int(0.1 * float64((p.SettingData.BetSet.Digits)))
	p.SettingData.BetSet.HighBet = 100 * p.SettingData.BetSet.Digits

}

func (p *Player) Ping() {

	p.pingTime = time.Now()

	var pong msg.MsgPong

	pong.NowTime = uint32(time.Now().Unix())

	p.SendMsg(msg.MsgType_Pong, &pong)

}

//func (p *Player) GetCredit() int {
//	return int(p.WalletData.GetCredit())
//}

func (p *Player) GetCredit() uint {
	return p.WalletData.GetCredit()
}

func (p *Player) GetDigits() uint {

	return p.WalletData.GetDigits()

}

// 从钱包服读取余额
func (p *Player) GetWalletBalance() bool {

	if p.WalletType == MultWallet {
		return p.getClustorWalletBalance()
	} else {
		return p.getApiWalletBalance()
	}

}

// 钱包服加点
func (p *Player) AddWalletBalance(data *OptData) bool {

	data.Opt = 0 //这边修改

	if p.WalletType == MultWallet {
		return p.optClusterWalletBalance(data)
	} else {
		return p.optApiWalletBalance(data)
	}

}

// 钱包减点服务
func (p *Player) DecWalletBalance(data *OptData) bool {

	data.Opt = 1 //这边修改

	if p.WalletType == MultWallet {
		return p.optClusterWalletBalance(data)
	} else {
		return p.optApiWalletBalance(data)
	}

}

func (p *Player) AddAccount(curAccount *CurAccount) bool {
	return p.AddBetAccount(curAccount.LocalBet, curAccount.LocalBetEffice, curAccount.LocalWater, curAccount.LocalWin)
}

// 设置余额
func (p *Player) SetBalance(balance float64) {

	credit := uint(balance * float64(p.WalletData.GetDigits()))

	p.WalletData.SetBalance(balance)
	p.WalletData.SetCredit(credit)

}

func (p *Player) AddBalance(bet float32) {

	betex := uint(bet * float32(p.WalletData.GetDigits()))

	p.WalletData.AddBalance(float64(bet))
	p.WalletData.AddCredit(betex)

}

func (p *Player) DecBalance(bet float32) {

	bet = -bet

	betex := uint(bet * float32(p.WalletData.GetDigits()))

	p.WalletData.AddBalance(float64(bet))
	p.WalletData.AddCredit(betex)

}

// 读取低注
func (p *Player) GetAreaLowBet(areadId int) float64 {

	limit, ok := p.SettingData.BetSet.BetAreas[areadId]

	if !ok {
		return 0
	}

	return limit.LowBet

}

// 读取高注
func (p *Player) GetAreaHighBet(areadId int) float64 {

	limit, ok := p.SettingData.BetSet.BetAreas[areadId]

	if !ok {
		return 0
	}

	return limit.HighBet
}

// 初始化玩家的钱包设置信息
func (p *Player) InitPlayerSetting() bool {

	//读取本地缓存的钱包数据
	if p.WalletData.ReadWalletData(p.Id) == false {
		return false
	}

	return true

}

// 初始化玩家数据
func (p *Player) InitPlayerInfo() bool {

	//读取玩家的属性信息
	if p.ReadUserInfo() == false {
		return false
	}
	return p.InitPlayerSetting()
}

// 货币种类
func (p *Player) GetCurrency() string {

	return p.WalletData.GetCurrency()
}

// 刷新接收ping的时间
func (p *Player) UpdatePingTime() {
	p.pingTime = time.Now()
}

// 检测ping值是否超时
func (p *Player) PingIsTimeOut() bool {

	nowTime := time.Now()

	//检测是否超出20秒
	if nowTime.Unix()-p.pingTime.Unix() > int64(30) {

		fmt.Println("*********pingTime:", p.pingTime)
		fmt.Println("*********nowTime:", nowTime)

		log.Infof("玩家userId:%s 掉线了", p.Id)

		return true
	}

	//5分钟后释放,这个不需要
	//if time.Now().Sub(p.pingTime) > 300*time.Second {
	//	p.Destroy()
	//}

	return false

}

// 主要用于通知商户单钱包服务，游戏开答了，让他们更新相关玩家订单的状态
func (p *Player) NotifyWalletOpenAnswer(data *OpenAnswerData) bool {

	if p.WalletType == MultWallet {
		fmt.Println("多钱包，暂不涉及更新加款扣款订单状态")
		return true
	} else {
		return p.optApiWalletOpenAnswer(data)
	}

}
