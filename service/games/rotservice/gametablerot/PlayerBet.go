package gameTableRot

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
	RotSingleId = iota
	RotDoubleId
	RotThreeId
	RotFourId
	RotSixId
	RotTwelveId
	RotEighteenId
	RotNeighbourTwoId //邻注
	RotTiersId        //轮盘下脚注
	RotOrphelinsId    //轮盘孤注
	RotVoisinsId      //零旁注上角
	RotJeu0Id         //轮上零旁
)

type WinType uint64

// 投注的点数
type BetPoint struct {
	areaId   int     //投注区域
	betPoint float32 //投注额度

	betTime time.Time //投注时间
}

type CurBets struct { //当前区域总的投注额
	Bets   [MaxIdNum]WinType //轮盘区域投注额
	Digits int
}

// 注意如果在一个位置下了多注，在计算区域中奖总额的的时候，
// 是+=合成一注一起计算，但投注明细是分多注进数据库的
func (curBet *CurBets) Bet(betId int, betPoint uint32) {

	if betId >= 0 && betId < MaxIdNum {
		curBet.Bets[betId] += WinType(betPoint)
	}
}

// 玩家投注类
type PlayerBet struct {
	Player *player.Player

	CurBets    CurBets           //当前区域总投注
	CurAccount player.CurAccount //当局的帐目情况.2025.3.19 add

	BeforeBalance uint //投注前的余额
	AfterBalance  uint //投注后的余额

	CurWin WinType //当前胜的胜额
	Digits int     //小数点位数

	betNum int //投注区域个数,百的类型多样，压注区域个数不一样

	BetList []betnode.BetNode //投注的列表

}

// 转换全局的投注区域ID为本游戏压注区主次ID
func convertToRotId(areaId int) (rgnMain, rgnSub int, ret bool) {

	ret = true

	//现在没有对越界检查代码，后续需要增加
	if areaId >= betOdds.Bet_Rot_SingleStart && areaId < betOdds.Bet_Rot_DoubleStart {

		rgnMain = RotSingleId
		rgnSub = areaId - betOdds.Bet_Rot_SingleStart

	} else if areaId >= betOdds.Bet_Rot_DoubleStart && areaId < betOdds.Bet_Rot_ThreeStart {

		rgnMain = RotDoubleId
		rgnSub = areaId - betOdds.Bet_Rot_DoubleStart

	} else if areaId >= betOdds.Bet_Rot_ThreeStart && areaId < betOdds.Bet_Rot_FourStart {

		rgnMain = RotThreeId
		rgnSub = areaId - betOdds.Bet_Rot_ThreeStart

	} else if areaId >= betOdds.Bet_Rot_FourStart && areaId < betOdds.Bet_Rot_SixStart {

		rgnMain = RotFourId
		rgnSub = areaId - betOdds.Bet_Rot_FourStart

	} else if areaId >= betOdds.Bet_Rot_SixStart && areaId < betOdds.Bet_Rot_TwelveStart {

		rgnMain = RotSixId
		rgnSub = areaId - betOdds.Bet_Rot_SixStart

	} else if areaId >= betOdds.Bet_Rot_TwelveStart && areaId < betOdds.Bet_Rot_EighteenStart {

		rgnMain = RotTwelveId
		rgnSub = areaId - betOdds.Bet_Rot_TwelveStart

	} else if areaId >= betOdds.Bet_Rot_EighteenStart && areaId < betOdds.Bet_Rot_NeighbourTwoStart {

		rgnMain = RotEighteenId
		rgnSub = areaId - betOdds.Bet_Rot_EighteenStart

	} else if areaId >= betOdds.Bet_Rot_NeighbourTwoStart && areaId < betOdds.Bet_Rot_Tiers {

		rgnMain = RotNeighbourTwoId
		rgnSub = areaId - betOdds.Bet_Rot_NeighbourTwoStart

	} else if areaId == betOdds.Bet_Rot_Tiers {

		rgnMain = RotTiersId
		rgnSub = 0

	} else if areaId == betOdds.Bet_Rot_Orphelins {

		rgnMain = RotOrphelinsId
		rgnSub = 0

	} else if areaId == betOdds.Bet_Rot_Voisins {

		rgnMain = RotVoisinsId
		rgnSub = 0

	} else if areaId == betOdds.Bet_Rot_Jeu0 {

		rgnMain = RotJeu0Id
		rgnSub = 0

	} else {

		ret = false
	}

	return rgnMain, rgnSub, ret
}

func convertToRotIndex(areaId int) (index int, ret bool) {

	if areaId >= betOdds.Bet_Rot_Start && areaId <= betOdds.Bet_Rot_End {
		index = areaId - betOdds.Bet_Rot_Start
		ret = true
	} else {
		index = -1
		ret = false
	}

	return index, ret
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

	//p.betNum = MaxIdNum

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

func (p *PlayerBet) TotalBet() float64 {

	total := WinType(0)

	for i := 0; i < MaxIdNum; i++ {
		total += p.CurBets.Bets[i]
	}

	return float64(total)
}

// 读取有投注的区域及投注
func (p *PlayerBet) GetValidBets() map[int]uint32 {

	bets := make(map[int]uint32)

	//检测有投注的区域
	for i := 0; i < MaxIdNum; i++ {

		if p.CurBets.Bets[i] > 0 {
			id, ok := convertToRotIndex(i)
			if ok {
				bets[id] = uint32(p.CurBets.Bets[i])
			}
		}
	}

	return bets
}

// 投注
func (p *PlayerBet) Bet(areaId int, betPoint uint32) bool {

	var betid int

	betid, bok := convertToRotIndex(areaId)
	fmt.Printf("betid to index=%d\n", betid)
	if bok == false {
		return false
	}

	p.CurBets.Bet(betid, betPoint)

	return true
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
	betid, bok := convertToRotIndex(areaId)

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

func (p *PlayerBet) PushBet(betItem betnode.BetNode) {

	//增加投注队列
	p.BetList = append(p.BetList, betItem)

}

// 读取有中奖的区域id列表
func GlobalGetWinAreaIds(win int) []uint32 {

	ppbet := PlayerBet{}

	betids := ppbet.GetWinAreaIds(win)

	return betids
}

// 读取有中奖的区域id列表
func (p *PlayerBet) GetWinAreaIds(win int) []uint32 {

	betIds := make([]uint32, 0)

	//方案一，类似骰宝用模拟100筹码下注所有区域计算Winlost，得到命中区域列表
	//不采用，因轮盘需要排除显示中奖号码的紧临区域(避免UI显示太乱)，只显示中奖号码和红黑单双黑白零上旁注等重点；

	//方案二，直接读取0-36号的固定map[int][]uint32指
	if _, exists := WinnerIdsMap[win]; exists {
		betIds = WinnerIdsMap[win]
	}

	return betIds
}

// 检测是否小于低注
func (p *PlayerBet) IsLessLowBet(areaId int) bool {

	betid, bok := convertToRotIndex(areaId)

	if bok == false {
		return false
	}

	//先检测投注区域是否存在
	betlimit, ok := p.Player.SettingData.BetSet.BetAreas[areaId]

	if ok == false {
		return false
	}

	lowbet := uint64(float64(betlimit.LowBet) * float64(p.Player.GetDigits()))
	blowbet := false

	if p.CurBets.Bets[betid] < WinType(lowbet) {
		blowbet = true
	}

	return blowbet
}

// 检测是否超出高注
func (p *PlayerBet) IsOverHighBet(areaId int) bool {

	betid, bok := convertToRotIndex(areaId)

	if bok == false {
		return false
	}

	var bhighbet bool = false

	//检测高低注区域是否存在
	betLimit, ok := p.Player.SettingData.BetSet.BetAreas[areaId]

	if ok == false {
		return false
	}

	highbet := WinType(float64(betLimit.HighBet) * float64(p.Digits))

	if p.CurBets.Bets[betid] > highbet {
		bhighbet = true
	}

	return bhighbet
}

// 检测是否超出台红
func (p *PlayerBet) IsOverTableBet(areaId int) bool {

	return true
}

// 区域1结算， 用于单球压注
func (p *PlayerBet) SingleWin(win int, bet WinType, rgnSub int) WinType {
	fmt.Printf("ooooo SingleWin win=%d, bet=%d, rgnSub=%d\n", win, bet, rgnSub)

	//先检测是否投中此区域内的
	if !(rgnSub >= 0 && rgnSub < SingleIdNum) {
		return 0
	}

	//检测开奖是否正确的数值
	if !(win >= 0 && win <= 36) {
		return 0
	}

	//检测是否有中奖了
	if win == rgnSub {
		//得到赔率
		odds, digits := betOdds.GetBetOdds(betOdds.Bet_Rot_SingleStart)
		fmt.Printf("ooooo SingleWin 命中了,odds=%d, digits=%d\n", odds, digits)

		win := bet * WinType(odds)                          //64位
		winPoint := WinType(float64(win) / float64(digits)) //实际胜额
		winPoint = winPoint + WinType(bet)                  //本金加上

		return winPoint
	}

	fmt.Printf("ooooo SingleWin FINISH ooooo\n")
	return 0
}

// 区域2结算， 用于双区域压注
func (p *PlayerBet) DoubleWin(win int, bet WinType, rgnSub int) WinType {
	fmt.Printf("ooooo DoubleWin win=%d, bet=%f, rgnSub=%d\n", win, bet, rgnSub)
	//先检测是否投中此区域内的
	if !(rgnSub >= 0 && rgnSub < DoubeIdNum) {
		return 0
	}

	//检测开奖是否正确的数值
	if !(win >= 0 && win <= 36) {
		return 0
	}

	fmt.Printf("ooooo DoubleWin DoubleRots[rgnSub]=%v", DoubleRots[rgnSub])
	for i := 0; i < 2; i++ {
		//检测是否中双区域
		if win == DoubleRots[rgnSub][i] {

			//得到赔率
			odds, digits := betOdds.GetBetOdds(betOdds.Bet_Rot_DoubleStart)

			fmt.Printf("ooooo DoubleWin 命中了, odds=%d, digits=%d\n", odds, digits)

			win := bet * WinType(odds)                          //64位
			winPoint := WinType(float64(win) / float64(digits)) //实际胜额
			winPoint = winPoint + WinType(bet)                  //本金加上

			return winPoint
		}
	}

	fmt.Printf("ooooo DoubleWin FINISH ooooo\n")
	return 0
}

// 区域3结算， 用于三数压注
func (p *PlayerBet) ThreeWin(win int, bet WinType, rgnSub int) WinType {
	fmt.Printf("ooooo ThreeWin win=%d, bet=%f, rgnSub=%d\n", win, bet, rgnSub)
	//先检测是否投中此区域内的
	if !(rgnSub >= 0 && rgnSub < ThreeIdNum) {
		return 0
	}

	//检测开奖是否正确的数值
	if !(win >= 0 && win <= 36) {
		return 0
	}

	fmt.Printf("ooooo ThreeWin ThreeRots[rgnSub]=%v", ThreeRots[rgnSub])
	for i := 0; i < 3; i++ {
		//检测是否中三数区域
		if win == ThreeRots[rgnSub][i] {

			//得到赔率
			odds, digits := betOdds.GetBetOdds(betOdds.Bet_Rot_ThreeStart)

			fmt.Printf("ooooo ThreeWin 命中了, odds=%d, digits=%d\n", odds, digits)

			win := bet * WinType(odds)                          //64位
			winPoint := WinType(float64(win) / float64(digits)) //实际胜额
			winPoint = winPoint + WinType(bet)                  //本金加上

			return winPoint
		}
	}

	fmt.Printf("ooooo ThreeWin FINISH ooooo\n")
	return 0
}

// 区域4结算， 用于四数压注
func (p *PlayerBet) FourWin(win int, bet WinType, rgnSub int) WinType {
	fmt.Printf("ooooo FourWin win=%d, bet=%f, rgnSub=%d\n", win, bet, rgnSub)
	//先检测是否投中此区域内的
	if !(rgnSub >= 0 && rgnSub < FourIdNum) {
		return 0
	}

	//检测开奖是否正确的数值
	if !(win >= 0 && win <= 36) {
		return 0
	}

	fmt.Printf("ooooo FourWin FourRots[rgnSub]=%v", FourRots[rgnSub])
	for i := 0; i < 4; i++ {
		//检测是否中四数区域
		if win == FourRots[rgnSub][i] {

			//得到赔率
			odds, digits := betOdds.GetBetOdds(betOdds.Bet_Rot_FourStart)

			fmt.Printf("ooooo FourWin 命中了, odds=%d, digits=%d\n", odds, digits)

			win := bet * WinType(odds)                          //64位
			winPoint := WinType(float64(win) / float64(digits)) //实际胜额
			winPoint = winPoint + WinType(bet)                  //本金加上

			return winPoint
		}
	}

	fmt.Printf("ooooo FourWin FINISH ooooo\n")
	return 0
}

// 区域5结算， 用于六数压注
func (p *PlayerBet) SixWin(win int, bet WinType, rgnSub int) WinType {
	fmt.Printf("ooooo SixWin win=%d, bet=%f, rgnSub=%d\n", win, bet, rgnSub)
	//先检测是否投中此区域内的
	if !(rgnSub >= 0 && rgnSub < SixIdNum) {
		return 0
	}

	//检测开奖是否正确的数值
	if !(win >= 0 && win <= 36) {
		return 0
	}

	fmt.Printf("ooooo SixWin SixRots[rgnSub]=%v", SixRots[rgnSub])
	for i := 0; i < 6; i++ {
		//检测是否中六数区域
		if win == SixRots[rgnSub][i] {

			//得到赔率
			odds, digits := betOdds.GetBetOdds(betOdds.Bet_Rot_SixStart)

			fmt.Printf("ooooo SixWin 命中了, odds=%d, digits=%d\n", odds, digits)

			win := bet * WinType(odds)                          //64位
			winPoint := WinType(float64(win) / float64(digits)) //实际胜额
			winPoint = winPoint + WinType(bet)                  //本金加上

			return winPoint

		}
	}

	fmt.Printf("ooooo SixWin FINISH ooooo\n")
	return 0
}

// 区域6结算， 用于十二数压注
func (p *PlayerBet) TwelveWin(win int, bet WinType, rgnSub int) WinType {
	fmt.Printf("ooooo TwelveWin win=%d, bet=%f, rgnSub=%d\n", win, bet, rgnSub)
	//先检测是否投中此区域内的
	if !(rgnSub >= 0 && rgnSub < TwelveIdNum) {
		return 0
	}

	//检测开奖是否正确的数值
	if !(win >= 0 && win <= 36) {
		return 0
	}

	fmt.Printf("ooooo TwelveWin TwelveRots[rgnSub]=%v", TwelveRots[rgnSub])
	for i := 0; i < 12; i++ {
		//检测是否中十二数区域
		if win == TwelveRots[rgnSub][i] {

			//得到赔率
			odds, digits := betOdds.GetBetOdds(betOdds.Bet_Rot_TwelveStart)

			fmt.Printf("ooooo TwelveWin 命中了, odds=%d, digits=%d\n", odds, digits)

			win := bet * WinType(odds)                          //64位
			winPoint := WinType(float64(win) / float64(digits)) //实际胜额
			winPoint = winPoint + WinType(bet)                  //本金加上

			return winPoint
		}
	}

	fmt.Printf("ooooo TwelveWin FINISH ooooo\n")
	return 0
}

// 区域7结算， 用于十八数压注
func (p *PlayerBet) EighteenWin(win int, bet WinType, rgnSub int) WinType {
	fmt.Printf("ooooo EighteenWin win=%d, bet=%f, rgnSub=%d\n", win, bet, rgnSub)
	//先检测是否投中此区域内的
	if !(rgnSub >= 0 && rgnSub < EighteenIdNum) {
		return 0
	}

	//检测开奖是否正确的数值
	if !(win >= 0 && win <= 36) {
		return 0
	}

	fmt.Printf("ooooo EighteenWin EighteenRots[rgnSub]=%v", EighteenRots[rgnSub])
	for i := 0; i < 18; i++ {
		//检测是否中十八数区域
		if win == EighteenRots[rgnSub][i] {

			//得到赔率
			odds, digits := betOdds.GetBetOdds(betOdds.Bet_Rot_EighteenStart)

			fmt.Printf("ooooo EighteenWin 命中了, odds=%d, digits=%d\n", odds, digits)

			win := bet * WinType(odds)                          //64位
			winPoint := WinType(float64(win) / float64(digits)) //实际胜额
			winPoint = winPoint + WinType(bet)                  //本金加上

			return winPoint
		}
	}

	fmt.Printf("ooooo EighteenWin FINISH ooooo\n")
	return 0
}

// 区域8结算， 轮盘圈上的两邻居邻注的压注
// 说明：目前这版先实现每个数的左右两个邻居(NeighbourTwo)
// 下注：筹码需要是5的倍数，相当于下了5个单注
// 开奖：如果命中5个中的某个数字，赔率固定是1陪35
func (p *PlayerBet) NeighbourTwoWin(win int, bet WinType, rgnSub int) WinType {
	fmt.Printf("ooooo NeighbourTwoWin win=%d, bet=%f, rgnSub=%d\n", win, bet, rgnSub)
	//先检测是否投中此区域内的
	if !(rgnSub >= 0 && rgnSub < NeighbourTwoIdNum) {
		return 0
	}

	//检测开奖是否正确的数值
	if !(win >= 0 && win <= 36) {
		return 0
	}

	fmt.Printf("ooooo NeighbourTwoWin NeighbourTwoRots[rgnSub]=%v", NeighbourTwoRots[rgnSub])
	for i := 0; i < 5; i++ {
		//检测是否中邻注区域
		if win == NeighbourTwoRots[rgnSub][i] {

			//得到赔率
			odds, digits := betOdds.GetBetOdds(betOdds.Bet_Rot_NeighbourTwoStart)

			fmt.Printf("ooooo NeighbourTwoWin 命中了, odds=%d, digits=%d\n", odds, digits)

			betPart := bet / NeighbourTwoMulti                  //下注额除以5得到每注的下注额
			win := betPart * WinType(odds)                      //64位
			winPoint := WinType(float64(win) / float64(digits)) //实际胜额
			winPoint = winPoint + WinType(bet)                  //本金加上

			return winPoint
		}
	}

	fmt.Printf("ooooo NeighbourTwoWin FINISH ooooo\n")
	return 0
}

// 区域9结算， 轮盘圈上的轮盘下角注
// 说明：0对面的区域，包括30,11,36,13,27,33,16,24,5,10,23,8
// 下注：筹码需要是6的倍数，其中包括5,8及10,11及13,16及23,24及27,30及33,36六个双数注
// 开奖，如果命中其中的某个数字，赔率固定是1陪17
func (p *PlayerBet) TiersWin(win int, bet WinType) WinType {
	fmt.Printf("ooooo TiersWin win=%d, bet=%f\n", win, bet)
	//检测开奖是否正确的数值
	if !(win >= 0 && win <= 36) {
		return 0
	}

	//判断有没有中奖
	if _, exists := TiersRange[win]; exists {
		//得到赔率
		odds, digits := betOdds.GetBetOdds(betOdds.Bet_Rot_Tiers)

		fmt.Printf("ooooo TiersWin 命中了, odds=%d, digits=%d\n", odds, digits)

		betPart := bet / TiersMulti                         //下注额除以6得到每注的下注额
		win := betPart * WinType(odds)                      //64位
		winPoint := WinType(float64(win) / float64(digits)) //实际胜额
		winPoint = winPoint + WinType(bet)                  //本金加上

		return winPoint
	}

	fmt.Printf("ooooo TiersWin FINISH ooooo\n")
	return 0
}

// 区域10结算， 轮盘圈上的轮盘孤注
// 说明：轮盘侧边的6,34,17,9,31,14,20,1
// 下注：筹码需要是5的倍数，其中包括了1一个单注，和6,9及14,17及17,20及31,34四个双数注
// 开奖：如果开1则赔率是1陪35，如果不是1则赔率是1陪17
func (p *PlayerBet) OrphelinsWin(win int, bet WinType) WinType {
	fmt.Printf("ooooo TiersWin win=%d, bet=%f\n", win, bet)
	//检测开奖是否正确的数值
	if !(win >= 0 && win <= 36) {
		return 0
	}

	//判断有没有中奖
	if _, exists := OrphelinsRange[win]; exists {

		odds, digits := betOdds.GetBetOdds(betOdds.Bet_Rot_Orphelins)

		if win == 1 {

			//如果开1号则是1陪35
			odds = 35
			fmt.Printf("ooooo OrphelinsWin 命中了1号, odds=%d, digits=%d\n", odds, digits)

		} else if win == 17 {

			//如果开17号，因两个分注都包含了17号，相当于赔率17乘以2
			odds = odds * 2
			fmt.Printf("ooooo OrphelinsWin 命中了17号, odds=%d, digits=%d\n", odds, digits)

		} else {

			//不开1号和17号，默认是1陪17
			fmt.Printf("ooooo OrphelinsWin 命中了非1号, odds=%d, digits=%d\n", odds, digits)

		}

		betPart := bet / OrphelinsMulti                     //下注额除以5得到每注的下注额
		win := betPart * WinType(odds)                      //64位
		winPoint := WinType(float64(win) / float64(digits)) //实际胜额
		winPoint = winPoint + WinType(bet)                  //本金加上

		return winPoint
	}

	fmt.Printf("ooooo OrphelinsWin FINISH ooooo\n")
	return 0
}

// 区域11结算， 轮盘圈上的零旁注上角
// 说明：轮盘上0边上的25,2,21,4,19,15,32,0,26,3,35,12,28,7,29,18,22
// 下注：筹码需要是9的倍数，其中包括了0,2,3一个三数注(注意要两个筹码)、25,26,28,29一个四数注(注意要两个筹码)、4,7和12,15等5个双数注，加起来一共需要9个筹码
// 开奖：如果开0或2或3赔率是1陪11，如果开25或26或28或29赔率是1陪8，开其他数则赔率是1陪17
func (p *PlayerBet) VoisinsWin(win int, bet WinType) WinType {
	fmt.Printf("ooooo VoisinsWin win=%d, bet=%f\n", win, bet)
	//检测开奖是否正确的数值
	if !(win >= 0 && win <= 36) {
		return 0
	}

	//判断有没有中奖
	if _, exists := VoisinsRotsRange[win]; exists {

		odds, digits := betOdds.GetBetOdds(betOdds.Bet_Rot_Voisins)
		var betPart WinType

		if win == 0 || win == 2 || win == 3 {

			//如果开0,2,3号是1陪11，注意0,2,3这个三数注占了9注里的2/9注
			odds = 11
			betPart = (bet / VoisinsMulti) * 2 //下注额除以9乘2得到每注的下注额
			fmt.Printf("ooooo OrphelinsWin 命中了0,2,3号, odds=%d, digits=%d\n", odds, digits)

		} else if win == 25 || win == 26 || win == 28 || win == 29 {

			//如果开25,26,28,29是1陪8，注意25,26,28,29这个四数注也是占了9注里的2/9注
			odds = 8
			betPart = (bet / VoisinsMulti) * 2 //下注额除以9乘2得到每注的下注额
			fmt.Printf("ooooo OrphelinsWin 命中了25,26,28,29号, odds=%d, digits=%d\n", odds, digits)

		} else {

			//其他默认是1陪17
			betPart = (bet / VoisinsMulti) //下注额除以9得到每注的下注额
			fmt.Printf("ooooo VoisinsWin 命中了其他号, odds=%d, digits=%d\n", odds, digits)

		}

		win := betPart * WinType(odds)                      //64位
		winPoint := WinType(float64(win) / float64(digits)) //实际胜额
		winPoint = winPoint + WinType(bet)                  //本金加上

		return winPoint

	}

	fmt.Printf("ooooo VoisinsWin FINISH ooooo\n")
	return 0
}

// 区域12结算， 轮盘圈上的轮上零旁
// 说明：轮盘上0边上的15,32,0,26,3,35
// 下注：筹码需要是4的倍数，其中包括了26一个单注，0,3双数注，12,15双数注，32,35双数注
// 开奖：如果开26赔率是1陪35，如果不是26则是1陪17
func (p *PlayerBet) Jeu0Win(win int, bet WinType) WinType {
	fmt.Printf("ooooo Jeu0Win win=%d, bet=%f\n", win, bet)
	//检测开奖是否正确的数值
	if !(win >= 0 && win <= 36) {
		return 0
	}

	//判断有没有中奖
	fmt.Printf("ooooo Jeu0Win Jeu0Range=%v\n", Jeu0Range)
	if _, exists := Jeu0Range[win]; exists {

		fmt.Printf("ooooo exists!!!\n")
		odds, digits := betOdds.GetBetOdds(betOdds.Bet_Rot_Jeu0)

		if win == 26 {

			//如果开26号是1陪35
			odds = 35
			fmt.Printf("ooooo OrphelinsWin 命中了26号, odds=%d, digits=%d\n", odds, digits)

		} else {

			//其他默认是一陪17
			fmt.Printf("ooooo Jeu0Win 命中了非26号, odds=%d, digits=%d\n", odds, digits)

		}

		betPart := (bet / Jeu0Multi)                        //下注额除以4得到每注的下注额
		win := betPart * WinType(odds)                      //64位
		winPoint := WinType(float64(win) / float64(digits)) //实际胜额
		winPoint = winPoint + WinType(bet)                  //本金加上

		return winPoint
	}

	fmt.Printf("ooooo Jeu0Win FINISH ooooo\n")
	return 0
}

// 检测玩家是否有赢额
func (p *PlayerBet) IsWin() bool {

	return p.CurWin > 0
}

// 开奖并结算
func (p *PlayerBet) WinLost(win int) WinType {
	fmt.Printf("RotService Start WinLost Count .....winner=%d\n", win)

	var totalWin WinType = 0
	var winPoint WinType = 0

	//检测全部区域的投注
	for index, bet := range p.CurBets.Bets {

		if bet == 0 {
			continue
		}

		//得到实际的投注区域
		betId := betOdds.Bet_Rot_Start + index

		rgnMain, rgnSub, bok := convertToRotId(betId)

		fmt.Printf("\n\nZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZ WinLost winner=%d, index=%d, betId=%d, rgnMain=%d, rgnSub=%d\n", win, index, betId, rgnMain, rgnSub)
		if bok == false {
			continue
		}

		switch rgnMain {
		case RotSingleId:
			winPoint = p.SingleWin(win, bet, rgnSub)
		case RotDoubleId:
			winPoint = p.DoubleWin(win, bet, rgnSub)
		case RotThreeId:
			winPoint = p.ThreeWin(win, bet, rgnSub)
		case RotFourId:
			winPoint = p.FourWin(win, bet, rgnSub)
		case RotSixId:
			winPoint = p.SixWin(win, bet, rgnSub)
		case RotTwelveId:
			winPoint = p.TwelveWin(win, bet, rgnSub)
		case RotEighteenId:
			winPoint = p.EighteenWin(win, bet, rgnSub)
		case RotNeighbourTwoId:
			winPoint = p.NeighbourTwoWin(win, bet, rgnSub)
		case RotTiersId:
			winPoint = p.TiersWin(win, bet)
		case RotOrphelinsId:
			winPoint = p.OrphelinsWin(win, bet)
		case RotVoisinsId:
			winPoint = p.VoisinsWin(win, bet)
		case RotJeu0Id:
			winPoint = p.Jeu0Win(win, bet)
		}

		if winPoint > 0 {
			totalWin += winPoint
		}
	}

	p.CurWin = totalWin

	betwin := uint64(p.CurWin) - uint64(p.TotalBet())

	//更新帐目数据
	p.refreshAccountWater()
	p.refreshAccountWinLost(int(betwin))

	return WinType(totalWin)
}

// 更新帐号的洗码情况
func (p *PlayerBet) refreshAccountWater() {

	bet := uint(p.TotalBet())

	p.CurAccount.SetBet(bet)

	p.CurAccount.SetBet(bet)
	p.CurAccount.SetWater(bet)
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
func (p *PlayerBet) SaveBetCollectToDB(gm *GameTableRot, set int, round int, roundId string, answer *msg_struct.GameOpenAnswerRot) {

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
	betCollect.Type = msg_struct.GameTypeRot
	bytes, _ := json.Marshal(answer)

	betCollect.AnswerTime = &nowTime
	betCollect.Answer = bytes

	//betCollect.RecordId = util.NewUnionId()

	betCollect.RecordId = util.GenerateUniqueID(util.RecordIdLength, betCollect.UserId)

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
