package betnode

import (
	"time"
	"videogame/common/util"
)

// 压注的结构体
type BetNode struct {
	UserId   string //玩家的ID
	ClientId string //网络通信的uid

	ParentId     string //上级的ID 这个为上级商户的ID
	MerchantId   string //商户的ID，这个为顶级商户的ID
	MerchantCode string //商户码
	RootKey      uint   //商户的key, 它处在树顶层的值

	OrderKey int64  //定单的Key
	OrkeyUid string //定单的唯一ID

	GameId string //游戏id
	DeskNo int    //游戏桌号

	BetMode int //投注模式 0 游戏内投注，1 多台投注

	Currency string //货币种类

	BetAreaId int       //投注区域Id,每个区域唯一
	BetPoint  uint32    //投注的金额
	Digits    int32     //小数位数
	BetTime   time.Time //投注时间
}

// 打赏的结构体
type TipNode struct {
	UserId   string //玩家的ID
	ClientId string //网络通信的uid

	ParentId     string //它的上级商户的ID
	MerchantId   string //顶级商户的ID
	MerchatnCode string //商户码

	RootKey uint //商户的key, 它处在树顶层的值

	OrderKey int64  //定单的Key
	OrkeyUid string //定单的唯一ID

	GameId string //游戏id
	DeskNo int    //游戏桌号

	Currency string //货币种类

	TipPoint uint32 //投注的金额
	Digits   int32  //小数位数

	DealerId string //打赏的荷官

	TipTime time.Time //打赏时间

}

// 产生唯一的定单ID
func (betNode *BetNode) CreateOrderId(prifx string) {

	//gid, err :=    gonanoid.New()
	//
	//if err != nil {
	//	//出错处理
	//	return
	//}
	//
	//recId := prifx + gid

	//	recId := util.GenerateRandomString(16)
	recId := util.GenerateUniqueID(24, betNode.UserId)

	betNode.OrkeyUid = recId

}

type BetItem struct {
	BetNode

	Set      int    //场次
	Round    int    //局号
	Round_Id string //对应局数的全局ID

	Type    string //游戏类型
	SubType string //二级游戏类型

	EffectiveBet int //有效投注
	WinLost      int //胜负金额

	Winner int //开奖结果
	Val1   int //开奖参数1
	Val2   int //开奖参数2
	Val3   int //开奖参数3

	Answer_Timer time.Time //开奖时间
}

func (betItem *BetItem) Build(gameType string, subType string, betNode *BetNode) {

	betItem.UserId = betNode.UserId
	betItem.ClientId = betNode.ClientId

	betItem.ParentId = betNode.ParentId
	betItem.MerchantId = betNode.MerchantId
	betItem.MerchantCode = betNode.MerchantCode
	betItem.RootKey = betNode.RootKey

	betItem.BetPoint = betNode.BetPoint
	betItem.Digits = betNode.Digits
	betItem.BetTime = betNode.BetTime
	betItem.BetAreaId = betNode.BetAreaId
	betItem.GameId = betNode.GameId
	betItem.DeskNo = betNode.DeskNo
	betItem.OrderKey = betNode.OrderKey
	betItem.OrkeyUid = betNode.OrkeyUid
	betItem.Currency = betNode.Currency
	betItem.BetMode = betNode.BetMode

	betItem.Type = gameType
	betItem.SubType = subType

}
