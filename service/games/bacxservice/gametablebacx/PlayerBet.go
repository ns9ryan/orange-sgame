package gameTableBacx

import (
	"encoding/json"
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"time"
	"videogame/common/betnode"
	betOdds "videogame/common/betodds"
	models "videogame/common/model"
	"videogame/common/msg_struct"
	"videogame/common/util"
	"videogame/service/games/player"
)

const (
	PlayerId = iota //区域id
	BankerId
	TieId
	PPairId //闲对子
	BPairId //庄对子

	BankerNocommId //免佣的庄
	SuperSixId     //超级Super

	AnyPairId     //任意对子
	PerfectPairId //完美对子
	BigId         //大
	SmallId       //小
	PlayerBonusId //闲龙宝
	BankerBonusId //庄龙宝
)

const TotalBetNum = 13 //总的投注区域

type WinType uint64

type CurBets struct { //当前区域总的投注额

	Bets [TotalBetNum]WinType //每个区的投注额

	Digits int
}

// 开答结算胜负用
type AnswerSettle struct {
	Winner int  //胜者
	PPair  bool //闲对子
	BPair  bool //庄对子

	SuperSix    bool //是否庄6点胜
	AnyPair     bool //任意对子
	PerfectPair bool //完美对子
	Big         bool //大

	PlayerBonus bool //是否闲龙宝
	BankBonus   bool //是否庄龙宝

	PlayerPoint int //闲家点数
	BankPoint   int //庄家点数

	PlayerCardNum int //玩家点数
	BankCardNum   int //庄家点数
}

// 玩家投注类
type PlayerBet struct {
	Player *player.Player

	CurBets    CurBets           //当前区域总投注
	CurAccount player.CurAccount //当局的帐目情况.2025.3.19 add

	BeforeBalance uint //压注前余额
	AfterBalance  uint //压注后余额

	CurWin WinType //当前胜的胜额
	Digits int     //小数点位数

	betNum int //投注区域个数,百的类型多样，压注区域个数不一样

	BetList []betnode.BetNode //以后需要优化成存指針

}

// 转换全局的投注区域ID为本游戏压注区id
func convertToBacxId(areaId int) (int, bool) {

	var betid int
	var bok bool

	bok = false

	end := betOdds.Bet_Bacx_Start + TotalBetNum
	betid = -1

	for i := betOdds.Bet_Bacx_Start; i < end; i++ {

		if areaId == i {
			betid = i - betOdds.Bet_Bacx_Start
			bok = true
			break
		}
	}

	return betid, bok
}

// 转换局部投注区域到全局id
func convertToGlobalId(areaId int) (int, bool) {

	if areaId >= PlayerId && areaId <= BankerBonusId {

		idoff := betOdds.Bet_Bacx_Start + (areaId - PlayerId)

		return idoff, true
	}

	return -1, false
}

// 检测投注区域是否符合
func (curB *CurBets) Bet(betId int, betPoint uint32) {

	if !(betId >= PlayerId && betId <= BankerBonusId) {
		return
	}

	curB.Bets[betId] += WinType(betPoint)
}

// 返回参数 ：投注区域id, 投注额度， 成功是否
func (p *PlayerBet) CancelLastBet() (betnode.BetNode, bool) {

	betNode := betnode.BetNode{}

	if len(p.BetList) == 0 {
		return betNode, false
	}

	ilen := len(p.BetList) - 1

	areaId := p.BetList[ilen].BetAreaId
	betPoint := p.BetList[ilen].BetPoint

	//转换成本游戏内部id
	betid, bok := convertToBacxId(areaId)

	if bok == false {
		return betNode, false
	}

	//扣除点数
	p.CurBets.Bets[betid] -= WinType(betPoint)

	betNode = p.BetList[ilen]

	//去除最后一个投注
	p.BetList = p.BetList[:ilen]

	return betNode, true
}

func (p *PlayerBet) DupliBet(pp *PlayerBet) {

	p.Player = pp.Player
	p.CurBets = pp.CurBets

}

// 拷贝投注明细数组
func (p *PlayerBet) GetBetNode() []betnode.BetNode {
	return p.BetList
}

func (p *PlayerBet) Init(player *player.Player) {

	p.CurBets = CurBets{}
	p.Player = player
	p.Digits = int(player.GetDigits())
	p.CurBets.Digits = p.Digits

	p.betNum = TotalBetNum

	p.BetList = make([]betnode.BetNode, 0)

}

// 设置投注前的余额
func (p *PlayerBet) SetBeforeBalance(balance uint) {
	p.BeforeBalance = balance
}

// 设置开奖后余额
func (p *PlayerBet) SetAfterBalance(balance uint) {
	p.AfterBalance = balance
}

// 复制一份压注拷备
func (p *PlayerBet) DuplicateBet(pp *PlayerBet) {

	p.CurBets = pp.CurBets

}

func (p *PlayerBet) TotalBet() WinType {

	total := WinType(0)

	var betnum int = p.betNum

	for i := 0; i < betnum; i++ {
		total += p.CurBets.Bets[i]
	}

	return total
}

func (p *PlayerBet) ResetWin() {

	p.CurWin = 0
}

// 检测玩家是否有赢额
func (p *PlayerBet) IsWin() bool {
	return p.CurWin > 0
}

// 检测是否小于低注
func (p *PlayerBet) IsLessLowBet(areaId int) bool {

	betid, bok := convertToBacxId(areaId)

	if bok == false {
		return false
	}
	if !(betid >= PlayerId && betid <= BankerBonusId) {
		return false
	}
	//
	////先检测投注ID是否在正常信偏移内
	//idoff := areaId - betOdds.Bet_Bacx_Start
	//
	//if idoff < 0 || idoff > betOdds.Bet_Bacx_BankerBonus-betOdds.Bet_Bacx_Start {
	//	return false
	//}

	betlimit, ok := p.Player.SettingData.BetSet.BetAreas[areaId]

	lowbet := WinType(float64(betlimit.LowBet) * float64(p.Player.GetDigits()))

	if ok == false {
		return false
	}

	//检测投注额度是否小于低注
	if p.CurBets.Bets[betid] < WinType(lowbet) {

		return true
	}

	return false
}

// 检测是否超出高注
func (p *PlayerBet) IsOverHighBet(areaId int) bool {

	betid, bok := convertToBacxId(areaId)

	if bok == false {
		return false
	}

	if !(betid >= PlayerId && betid <= BankerBonusId) {
		return false
	}

	////先检测投注ID是否在正常信偏移内
	//idoff := areaId - betOdds.Bet_Bacx_Start
	//
	//if idoff < 0 || idoff > betOdds.Bet_Bacx_BankerBonus-betOdds.Bet_Bacx_Start {
	//	return false
	//}

	betlimit, ok := p.Player.SettingData.BetSet.BetAreas[areaId]

	if ok == false {
		return false
	}

	highbet := WinType(float64(betlimit.HighBet) * float64(p.Digits))

	bhigh := false

	//检测龙区域压注是否小于低注
	if p.CurBets.Bets[betid] > highbet {
		bhigh = true
	}

	return bhigh

	////庄闲投注额之差要小于限红
	//if math.Abs(float64(p.CurBets.DragonBet-p.CurBets.TigerBet)) > float64(p.Player.BetLimit.DragonArea.HighBet) {
	//	return true
	//}

}

// 检测是否超出台红
func (p *PlayerBet) IsOverTableBet(areaId int) bool {

	return true
}

// 读取有投注的区域及投注,现版本只检测三个区域
func (p *PlayerBet) GetValidBets() map[int]uint32 {

	bets := make(map[int]uint32)

	//检测有投注的区域
	for i := 0; i < p.betNum; i++ {

		if p.CurBets.Bets[i] > 0 {
			id, ok := convertToGlobalId(i)
			if ok {
				bets[id] = uint32(p.CurBets.Bets[i])
			}
		}
	}

	return bets
}

// 检测投注ID是否符合本游戏，百家乐现在多种玩法，每种游戏区域都不一样
func (p *PlayerBet) BetIdIsRight(areaId int) bool {

	if areaId >= betOdds.Bet_Bacx_Start && areaId <= betOdds.Bet_Bacx_PerfectPair {
		return true
	}

	return false
}

// 检测是否有打水,主要是对庄闲压注的检测
func (p *PlayerBet) IsCheckWaterBet(betId int) bool {

	//
	if betId == PlayerId {
		//检测庄是否已经有压了
		if p.CurBets.Bets[BankerId] > 0 {
			return true
		}

	} else if betId == BankerId {

		//检测庄是否已经有压了
		if p.CurBets.Bets[PlayerId] > 0 {
			return true
		}

	}

	return false
}

// 检测是否超级
func (p *PlayerBet) SuperSixAreaIsBet(areaId int) bool {

	betid, bok := convertToBacxId(areaId)

	if bok == false {
		return false
	}

	if betid == BankerNocommId {
		return true
	}

	if betid == SuperSixId {
		return true
	}

	return false
}

// 检测是否有压注普通的庄`
func (p *PlayerBet) BankerAreaIsBet(areaId int) bool {

	betid, bok := convertToBacxId(areaId)

	if bok == false {
		return false
	}

	if betid == BankerId {
		return true
	}

	return false
}

// 投注
func (p *PlayerBet) Bet(areaId int, betPoint uint32) bool {

	var betid int

	//全局区域ID转换成本地ID
	betid, bok := convertToBacxId(areaId)

	if bok == false {
		return false
	}

	////检测是否存在,闲,庄打水
	//if p.IsCheckWaterBet(betid) == true {
	//	return false
	//}

	p.CurBets.Bet(betid, betPoint)

	return true
}

func (p *PlayerBet) PushBet(betItem betnode.BetNode) {

	//增加投注队列
	p.BetList = append(p.BetList, betItem)

}

// 闲赢结算
func (p *PlayerBet) PlayerWin() WinType {

	if p.CurBets.Bets[PlayerId] > 0 {

		//得到赔率
		odds, digits := betOdds.GetBetOdds(betOdds.Bet_Bacx_Player)

		win := p.CurBets.Bets[PlayerId] * WinType(odds) //64位

		winPoint := WinType(float64(win) / float64(digits)) //实际胜额

		winPoint = winPoint + WinType(p.CurBets.Bets[PlayerId]) //本金加上

		return winPoint

	}

	return 0
}

// 庄赢结算
func (p *PlayerBet) BankerWin() WinType {

	if p.CurBets.Bets[BankerId] > 0 {

		//得到赔率
		odds, digits := betOdds.GetBetOdds(betOdds.Bet_Bacx_Banker)

		win := p.CurBets.Bets[BankerId] * WinType(odds) //64位

		winPoint := WinType(float64(win) / float64(digits)) //实际胜额

		winPoint = winPoint + WinType(p.CurBets.Bets[BankerId]) //本金加上

		return winPoint
	}

	return 0
}

// 免佣庄结算
func (p *PlayerBet) BankerNocommIdWin(supersix bool) WinType {

	if p.CurBets.Bets[BankerNocommId] > 0 {

		//得到赔率
		if supersix == false {

			odds, digits := betOdds.GetBetOdds(betOdds.Bet_Bacx_BankerNocomm)
			win := p.CurBets.Bets[BankerNocommId] * WinType(odds)         //64位
			winPoint := WinType(float64(win) / float64(digits))           //实际胜额
			winPoint = winPoint + WinType(p.CurBets.Bets[BankerNocommId]) //本金加上

			return winPoint

		} else {

			//若庄家是6点胜， 则赔率是1:0.5
			odds, digits := betOdds.GetBetOdds(betOdds.Bet_Bacx_BankerNocomm)
			win := p.CurBets.Bets[BankerNocommId] * WinType(odds)         //64位
			winPoint := WinType(float64(win)/float64(digits)) / 2         //被抽取一半压注金额
			winPoint = winPoint + WinType(p.CurBets.Bets[BankerNocommId]) //本金加上

			return winPoint
		}

	}

	return 0
}

// 和赢结算
func (p *PlayerBet) TieWin() WinType {

	winPoint := WinType(0)

	//检测和局区域有没有压注
	if p.CurBets.Bets[TieId] > 0 {

		//得到赔率
		odds, digits := betOdds.GetBetOdds(betOdds.Bet_Bacx_Tie)

		win := p.CurBets.Bets[TieId] * WinType(odds) //64位

		winPoint = WinType(float64(win) / float64(digits)) //实际胜额
		winPoint += WinType(p.CurBets.Bets[TieId])         //本金加上
	}

	//检测是否压闲了
	if p.CurBets.Bets[PlayerId] > 0 {

		//得到赔率
		win := p.CurBets.Bets[PlayerId] //64位

		winPoint += WinType(win) //实际胜额
	}

	//检测是否压闲了
	if p.CurBets.Bets[BankerId] > 0 {

		//得到赔率
		win := p.CurBets.Bets[BankerId] //64位

		winPoint += WinType(win) //实际胜额
	}

	return winPoint
}

// 开和结算
func (p *PlayerBet) TieWinNoCommon() WinType {

	winPoint := WinType(0)

	//检测和局区域有没有压注
	if p.CurBets.Bets[TieId] > 0 {

		//得到赔率
		odds, digits := betOdds.GetBetOdds(betOdds.Bet_Bacx_Tie)

		win := p.CurBets.Bets[TieId] * WinType(odds) //64位

		winPoint += WinType(float64(win) / float64(digits)) //实际胜额
		winPoint += WinType(p.CurBets.Bets[TieId])          //本金加上
	}

	//检测是否压闲了
	if p.CurBets.Bets[PlayerId] > 0 {

		//得到赔率
		win := p.CurBets.Bets[PlayerId] //64位

		winPoint += WinType(win) //实际胜额
	}

	//检测是否压闲了
	if p.CurBets.Bets[BankerNocommId] > 0 {

		//得到赔率
		win := p.CurBets.Bets[BankerNocommId] //64位

		winPoint += WinType(win) //实际胜额
	}

	return winPoint
}

// 开和，检测是否有压龙宝的情况
func (p *PlayerBet) TieWinBonus(PlayerPoint, PCardNum int, BankerPoint, BCardNum int) WinType {

	winPoint := WinType(0)

	//检测闲宝龙区域有没有压注
	if p.CurBets.Bets[PlayerBonusId] > 0 {

		if PlayerPoint == 8 || PlayerPoint == 9 {

			//检测是否例牌胜
			if PCardNum == 2 {
				//退回本金
				winPoint += WinType(p.CurBets.Bets[PlayerBonusId]) //本金加上
			}
		}
	}

	//检测庄龙宝区域有没有压注
	if p.CurBets.Bets[BankerBonusId] > 0 {

		if BankerPoint == 8 || BankerPoint == 9 {

			//检测是否例牌胜
			if BCardNum == 2 {
				//退回本金
				winPoint += WinType(p.CurBets.Bets[BankerBonusId]) //本金加上
			}
		}
	}

	return winPoint

}

// 闲对子结算
func (p *PlayerBet) PPairWin() WinType {

	if p.CurBets.Bets[PPairId] > 0 {

		//得到赔率
		odds, digits := betOdds.GetBetOdds(betOdds.Bet_Bacx_PPair)

		win := p.CurBets.Bets[PPairId] * WinType(odds) //64位

		winPoint := WinType(float64(win) / float64(digits)) //实际胜额

		winPoint = winPoint + WinType(p.CurBets.Bets[PPairId]) //本金加上

		return winPoint
	}
	return 0
}

// 庄对子结算
func (p *PlayerBet) BPairWin() WinType {

	if p.CurBets.Bets[BPairId] > 0 {

		//得到赔率
		odds, digits := betOdds.GetBetOdds(betOdds.Bet_Bacx_BPair)
		win := p.CurBets.Bets[BPairId] * WinType(odds)         //64位
		winPoint := WinType(float64(win) / float64(digits))    //实际胜额
		winPoint = winPoint + WinType(p.CurBets.Bets[BPairId]) //本金加上

		return winPoint
	}
	return 0
}

// SUPERSIX 结算
func (p *PlayerBet) SuperSixWin() WinType {

	betid := SuperSixId
	gid := betOdds.Bet_Bacx_SuperSix

	if p.CurBets.Bets[betid] > 0 {

		//得到赔率
		odds, digits := betOdds.GetBetOdds(gid)
		win := p.CurBets.Bets[betid] * WinType(odds)         //64位
		winPoint := WinType(float64(win) / float64(digits))  //实际胜额
		winPoint = winPoint + WinType(p.CurBets.Bets[betid]) //本金加上

		return winPoint
	}

	return 0
}

// 完美对子结算
func (p *PlayerBet) PerfectPairWin() WinType {

	betid := PerfectPairId
	gid := betOdds.Bet_Bacx_PerfectPair

	if p.CurBets.Bets[betid] > 0 {

		//得到赔率
		odds, digits := betOdds.GetBetOdds(gid)
		win := p.CurBets.Bets[betid] * WinType(odds)         //64位
		winPoint := WinType(float64(win) / float64(digits))  //实际胜额
		winPoint = winPoint + WinType(p.CurBets.Bets[betid]) //本金加上

		return winPoint
	}

	return 0
}

// 任意对子结算
func (p *PlayerBet) AnyPairWin() WinType {

	betid := AnyPairId
	gid := betOdds.Bet_Bacx_AnyPair

	if p.CurBets.Bets[betid] > 0 {

		//得到赔率
		odds, digits := betOdds.GetBetOdds(gid)
		win := p.CurBets.Bets[betid] * WinType(odds)         //64位
		winPoint := WinType(float64(win) / float64(digits))  //实际胜额
		winPoint = winPoint + WinType(p.CurBets.Bets[betid]) //本金加上

		return winPoint
	}

	return 0
}

// 大结算
func (p *PlayerBet) BigWin() WinType {

	betid := BigId
	gid := betOdds.Bet_Bacx_Big

	if p.CurBets.Bets[betid] > 0 {

		//得到赔率
		odds, digits := betOdds.GetBetOdds(gid)
		win := p.CurBets.Bets[betid] * WinType(odds)         //64位
		winPoint := WinType(float64(win) / float64(digits))  //实际胜额
		winPoint = winPoint + WinType(p.CurBets.Bets[betid]) //本金加上

		return winPoint
	}

	return 0
}

// 小结算
func (p *PlayerBet) SmallWin() WinType {

	betid := SmallId
	gid := betOdds.Bet_Bacx_Small

	if p.CurBets.Bets[betid] > 0 {

		//得到赔率
		odds, digits := betOdds.GetBetOdds(gid)
		win := p.CurBets.Bets[betid] * WinType(odds)         //64位
		winPoint := WinType(float64(win) / float64(digits))  //实际胜额
		winPoint = winPoint + WinType(p.CurBets.Bets[betid]) //本金加上

		return winPoint
	}

	return 0

}

// 闲龙宝结算
func (p *PlayerBet) PlayerBonusWin(playPoint int, pcardNum int, bankPoint int, bcardNum int) WinType {

	//若是和
	if playPoint == bankPoint {

		//检测是否例牌和局，若是的话退回下注
		if pcardNum == 2 {

			winPoint := WinType(p.CurBets.Bets[PlayerBonusId]) //本金加上

			return winPoint
		}

		return 0
	}

	//若闲家没有胜则返回
	if playPoint < bankPoint {
		return 0
	}

	//处理闲家只有二家牌的情况
	if pcardNum == 2 {

		//8点或9点胜
		if playPoint == 8 || playPoint == 9 {

			//得到赔率
			odds, digits := betOdds.GetBonusOdds(betOdds.BonusWin8or9NRC)

			win := p.CurBets.Bets[PlayerBonusId] * WinType(odds)         //64位
			winPoint := WinType(float64(win) / float64(digits))          //实际胜额
			winPoint = winPoint + WinType(p.CurBets.Bets[PlayerBonusId]) //本金加上

			return winPoint
		}
	}

	winPoint := playPoint - bankPoint
	winType := 0

	//胜4点
	if winPoint == 4 {
		winType = betOdds.BonusWin4Natural
	} else if winPoint == 5 {
		winType = betOdds.BonusWin5Natural
	} else if winPoint == 6 {
		winType = betOdds.BonusWin6Natural
	} else if winPoint == 7 {
		winType = betOdds.BonusWin7Natural
	} else if winPoint == 8 {
		winType = betOdds.BonusWin8Natural
	} else if winPoint == 9 {
		winType = betOdds.BonusWin9Natural
	} else {
		return 0
	}

	//得到赔率
	odds, digits := betOdds.GetBonusOdds(winType)

	win1 := p.CurBets.Bets[PlayerBonusId] * WinType(odds) //64位

	winPoint1 := WinType(float64(win1) / float64(digits)) //实际胜额

	winPoint1 = winPoint1 + WinType(p.CurBets.Bets[PlayerBonusId]) //本金加上

	return winPoint1

}

// 庄龙宝结算
func (p *PlayerBet) BankerBonusWin(playPoint int, pcardNum int, bankPoint int, bcardNum int) WinType {

	//若是和
	if playPoint == bankPoint {

		//检测是否例牌和局，若是的话退回下注
		if pcardNum == 2 {

			winPoint := WinType(p.CurBets.Bets[BankerBonusId]) //本金加上

			return winPoint
		}

		return 0
	}

	//若闲家胜，则不处理返回0
	if playPoint > bankPoint {
		return 0
	}

	if bcardNum == 2 {

		//8点或9点胜
		if bankPoint == 8 || bankPoint == 9 {

			//得到赔率
			odds, digits := betOdds.GetBonusOdds(betOdds.BonusWin8or9NRC)

			win := p.CurBets.Bets[BankerBonusId] * WinType(odds)         //64位
			winPoint := WinType(float64(win) / float64(digits))          //实际胜额
			winPoint = winPoint + WinType(p.CurBets.Bets[BankerBonusId]) //本金加上

			return winPoint
		}
	}

	winPoint := bankPoint - playPoint
	winType := 0

	//胜4点
	if winPoint == 4 {
		winType = betOdds.BonusWin4Natural
	} else if winPoint == 5 {
		winType = betOdds.BonusWin5Natural
	} else if winPoint == 6 {
		winType = betOdds.BonusWin6Natural
	} else if winPoint == 7 {
		winType = betOdds.BonusWin7Natural
	} else if winPoint == 8 {
		winType = betOdds.BonusWin8Natural
	} else if winPoint == 9 {
		winType = betOdds.BonusWin9Natural
	} else {
		return 0
	}

	//得到赔率
	odds, digits := betOdds.GetBonusOdds(winType)

	win1 := p.CurBets.Bets[BankerBonusId] * WinType(odds) //64位

	winPoint1 := WinType(float64(win1) / float64(digits)) //实际胜额

	winPoint1 = winPoint1 + WinType(p.CurBets.Bets[BankerBonusId]) //本金加上

	return winPoint1

}

func (p *PlayerBet) WinSettle(settle *AnswerSettle) WinType {

	var winPoint WinType = 0

	p.ResetWin()

	switch settle.Winner {
	case PlayerId: //开答是闲
		winPoint = p.PlayerWin()
	case BankerId: //开答是庄
		winPoint = p.BankerWin()
	case TieId:

		winPoint = p.TieWin()

		//计算是否有人压闲龙宝
		winPoint += p.TieWinBonus(settle.PlayerPoint, settle.PlayerCardNum, settle.BankPoint, settle.BankCardNum)

	}

	if settle.PPair == true {
		winPoint += p.PPairWin()
	}
	if settle.BPair == true {
		winPoint += p.BPairWin()
	}

	////是否庄家6点胜
	//if settle.SuperSix == true {
	//	winPoint += p.SuperSixWin()
	//}
	//是否任意对子胜
	if settle.AnyPair == true {
		winPoint += p.AnyPairWin()
	}
	//是否完美对子胜
	if settle.PerfectPair == true {
		winPoint += p.PerfectPairWin()
	}

	//是否大胜
	if settle.Big == true {
		winPoint += p.BigWin()
	} else {
		winPoint += p.SmallWin()
	}

	//若开奖不是和，计算闲，庄龙宝
	if settle.Winner != TieId {

		//是否闲龙宝
		if settle.PlayerBonus == true {
			winPoint += p.PlayerBonusWin(settle.PlayerPoint, settle.PlayerCardNum, settle.BankPoint, settle.BankCardNum)
		}

		//是否庄龙宝
		if settle.BankBonus == true {
			winPoint += p.BankerBonusWin(settle.PlayerPoint, settle.PlayerCardNum, settle.BankPoint, settle.BankCardNum)
		}
	}

	//是否庄龙宝

	totalbet := p.TotalBet()
	p.CurWin = winPoint

	betwin := int(winPoint - totalbet)

	p.refreshAccountWater(settle.Winner)
	p.refreshAccountWinLost(betwin)

	return WinType(winPoint)
}

func (p *PlayerBet) WinSettleWithSuperSix(settle *AnswerSettle) WinType {

	var winPoint WinType = 0

	p.ResetWin()

	switch settle.Winner {
	case PlayerId: //开答是闲
		winPoint = p.PlayerWin()
	case BankerId: //开答是庄
		winPoint = p.BankerNocommIdWin(settle.SuperSix)
	case TieId:
		winPoint = p.TieWinNoCommon()
		winPoint += p.TieWinBonus(settle.PlayerPoint, settle.PlayerCardNum, settle.BankPoint, settle.BankCardNum)

	}

	if settle.PPair == true {
		winPoint += p.PPairWin()
	}
	if settle.BPair == true {
		winPoint += p.BPairWin()
	}

	//是否庄家6点胜,检测是否压supersix
	if settle.SuperSix == true {
		winPoint += p.SuperSixWin()
	}
	//是否任意对子胜
	if settle.AnyPair == true {
		winPoint += p.AnyPairWin()
	}
	//是否完美对子胜
	if settle.PerfectPair == true {
		winPoint += p.PerfectPairWin()
	}

	//是否大胜
	if settle.Big == true {
		winPoint += p.BigWin()
	} else {
		winPoint += p.SmallWin()
	}

	if settle.Winner != TieId {
		//是否闲龙宝
		if settle.PlayerBonus == true {
			winPoint += p.PlayerBonusWin(settle.PlayerPoint, settle.PlayerCardNum, settle.BankPoint, settle.BankCardNum)
		}

		//是否庄龙宝
		if settle.BankBonus == true {
			winPoint += p.BankerBonusWin(settle.PlayerPoint, settle.PlayerCardNum, settle.BankPoint, settle.BankCardNum)
		}
	}

	//是否大
	totalbet := p.TotalBet()
	p.CurWin = winPoint

	betwin := int(winPoint - totalbet)

	p.refreshAccountWater(settle.Winner)
	p.refreshAccountWinLost(betwin)

	return WinType(winPoint)
}

// 更新帐号的洗码情况
func (p *PlayerBet) refreshAccountWater(win int) {

	bet := uint(p.TotalBet())

	p.CurAccount.SetBet(bet)

	water := uint(0)

	if win != TieId { //若没有开和，庄，闲，和则水钱全算

		water = uint(p.CurBets.Bets[PlayerId] + p.CurBets.Bets[BankerId] + p.CurBets.Bets[TieId])

	} else {

		//开和，只有压注和才算水
		water = uint(p.CurBets.Bets[TieId])
	}

	//其它区域都要算上水钱
	for i := PPairId; i <= BankerBonusId; i++ {
		water += uint(p.CurBets.Bets[i])
	}

	//设置洗码量
	p.CurAccount.SetWater(water)

	//设置有效投注
	p.CurAccount.SetBetEffective(bet)

}

func (p *PlayerBet) refreshAccountWinLost(winLost int) {

	p.CurAccount.SetWin(winLost)
}

func (p *PlayerBet) UpdateAccountToDB() {

	//存储到钱包服数据库，若钱包服调用不成功，自已再存储到数据库
	if p.Player.AddAccount(&p.CurAccount) == false {

		//更新到本地数据库
		p.Player.AccountData.AddAccountToDB(p.Player.Id, &p.CurAccount)
	}

	//更新到

}

// 投注集合到数据库
func (p *PlayerBet) SaveBetCollectToDBEx(gm *GameTableBacx, set int, round int, roundId string, beforeBalance uint, afterBalance uint, answer *msg_struct.GameOpenAnswerBacx) {

	betCount := len(p.BetList)
	betAmout := p.TotalBet()

	nowTime := time.Now()

	if betCount < 1 {
		return
	}

	var betCollect models.SBetAccountRecord

	betCollect.GameId = p.BetList[0].GameId
	betCollect.UserId = p.BetList[0].UserId
	betCollect.DeskNo = p.BetList[0].DeskNo

	betCollect.Currency = p.BetList[0].Currency
	betCollect.ParentId = p.BetList[0].ParentId
	betCollect.MerchantId = p.BetList[0].MerchantId
	betCollect.MerchantCode = p.BetList[0].MerchantCode
	betCollect.RootKey = p.BetList[0].RootKey

	betCollect.Digits = int(p.BetList[0].Digits)
	betCollect.Set = set
	betCollect.Round = round

	betCollect.BetAmount = int(betAmout)
	betCollect.BetCount = betCount
	betCollect.WinLost = int(p.CurWin)

	betCollect.BeforeBalance = beforeBalance
	betCollect.AfterBalance = afterBalance

	betCollect.BetTime = &p.BetList[0].BetTime

	betCollect.RoundId = roundId
	betCollect.Type = msg_struct.GameTypeBacx
	bytes, _ := json.Marshal(answer)

	betCollect.AnswerTime = &nowTime
	betCollect.Answer = bytes

	//betCollect.RecordId = util.NewUnionId()
	betCollect.RecordId = util.GenerateUniqueID(util.RecordIdLength, p.Player.UserId)

	queueId := int64(2) //使用第二个队列

	//启用协程队列存储
	gm.AsyncDoByQueue(queueId, func() bool {

		err := betCollect.Create()

		if err != nil {
			//存数据库不成功
			log.Errorf("SaveBetCollectToDBAsync 不成功，错误:%s", err.Error())
		}

		fmt.Println("SaveBetCollectToDBAsync===========")

		return false

	}, nil)

}

// 投注集合到数据库
func (p *PlayerBet) SaveBetCollectToDB(gm *GameTableBacx, set int, round int, roundId string, answer *msg_struct.GameOpenAnswerBacx) {

	betCount := len(p.BetList)
	betAmout := p.TotalBet()

	nowTime := time.Now()

	if betCount < 1 {
		return
	}

	var betCollect models.SBetAccountRecord

	betCollect.GameId = p.BetList[0].GameId
	betCollect.UserId = p.BetList[0].UserId
	betCollect.DeskNo = p.BetList[0].DeskNo

	betCollect.Currency = p.BetList[0].Currency
	betCollect.ParentId = p.BetList[0].ParentId
	betCollect.MerchantId = p.BetList[0].MerchantId
	betCollect.MerchantCode = p.BetList[0].MerchantCode
	betCollect.RootKey = p.BetList[0].RootKey

	betCollect.Digits = int(p.BetList[0].Digits)
	betCollect.Set = set
	betCollect.Round = round

	betCollect.BetAmount = int(betAmout)
	betCollect.BetCount = betCount
	betCollect.WinLost = int(p.CurWin)

	betCollect.BeforeBalance = p.BeforeBalance
	betCollect.AfterBalance = p.AfterBalance

	betCollect.BetTime = &p.BetList[0].BetTime

	betCollect.RoundId = roundId
	betCollect.Type = msg_struct.GameTypeBacx
	bytes, _ := json.Marshal(answer)

	betCollect.AnswerTime = &nowTime
	betCollect.Answer = bytes

	//betCollect.RecordId = util.NewUnionId()
	betCollect.RecordId = util.GenerateUniqueID(util.RecordIdLength, p.Player.UserId)

	queueId := int64(2) //使用第二个队列

	//启用协程队列存储
	gm.AsyncDoByQueue(queueId, func() bool {

		err := betCollect.Create()

		if err != nil {
			//存数据库不成功
			log.Errorf("SaveBetCollectToDBAsync 不成功，错误:%s", err.Error())
		}

		fmt.Println("SaveBetCollectToDBAsync===========")

		return false

	}, nil)

}
