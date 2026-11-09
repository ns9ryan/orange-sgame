package gameTableSicbo

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

//=============================================================================================
// 骰子宝投注区域
// 主区域代号： 大小单双：0   特定双数：1 特定三数：2 任意三数：3 总点数：4 一对组合：5   数字赌注：6
//=============================================================================================

const (
	OneId = iota
	TwoSpecId
	ThreeSameId
	ThreeAnySameId
	SumTotalId
	PairId
	NumberId
)

type WinType uint64

// 转换全局的投注区域ID为本游戏压注区主次ID
func convertToSicboArea2D(areaId int) (rgnMain, rgnSub int, ret bool) {

	ret = true

	//现在没有对越界检查代码，后续需要增加
	if areaId >= betOdds.Bet_Sicbo_One_Big && areaId <= betOdds.Bet_Sicbo_One_Even {

		rgnMain = OneId
		rgnSub = areaId - betOdds.Bet_Sicbo_One_Big

	} else if areaId >= betOdds.Bet_Sicbo_TwoNum_1And2 && areaId <= betOdds.Bet_Sicbo_TwoNum_5And6 {

		rgnMain = TwoSpecId
		rgnSub = areaId - betOdds.Bet_Sicbo_TwoNum_1And2

	} else if areaId >= betOdds.Bet_Sicbo_ThreeSame_1All && areaId <= betOdds.Bet_Sicbo_ThreeSame_6All {

		rgnMain = ThreeSameId
		rgnSub = areaId - betOdds.Bet_Sicbo_ThreeSame_1All

	} else if areaId == betOdds.Bet_Sicbo_ThreeSame_Any {

		rgnMain = ThreeAnySameId
		rgnSub = 0 //全围只有一个区域，任意同样的三点一样

	} else if areaId >= betOdds.Bet_Sicbo_SumPoint_4 && areaId <= betOdds.Bet_Sicbo_SumPoint_17 {

		rgnMain = SumTotalId
		rgnSub = areaId - betOdds.Bet_Sicbo_SumPoint_4

	} else if areaId >= betOdds.Bet_Sicbo_PairSame_1 && areaId <= betOdds.Bet_Sicbo_PairSame_6 {

		rgnMain = PairId
		rgnSub = areaId - betOdds.Bet_Sicbo_PairSame_1

	} else if areaId >= betOdds.Bet_Sicbo_NumPoint_1 && areaId <= betOdds.Bet_Sicbo_NumPoint_6 {

		rgnMain = NumberId
		rgnSub = areaId - betOdds.Bet_Sicbo_NumPoint_1

	} else {
		ret = false
		rgnMain = -1
		rgnSub = -1
	}

	return rgnMain, rgnSub, ret
}

// 转换全局区域id 到索引值
func convertToSicboIndex(areaId int) (index int, ret bool) {

	if areaId >= betOdds.Bet_Sicbo_Start && areaId <= betOdds.Bet_Sicbo_End {
		index = areaId - betOdds.Bet_Sicbo_Start
		ret = true
	} else {
		index = -1
		ret = false
	}

	return index, ret
}

// 转换局部投注区域到全局id
func convertToGlobalId(areaId int) (int, bool) {

	if areaId >= 0 && areaId < MaxIdNum {

		idoff := betOdds.Bet_Sicbo_Start + areaId

		return idoff, true
	}

	return -1, false
}

type CurBets struct { //当前区域总的投注额
	Bets [MaxIdNum]WinType //轮盘区域投注额

	Digits int
}

func (curBet *CurBets) Bet(index int, bet WinType) bool {

	if index >= 0 && index < MaxIdNum {
		curBet.Bets[index] += bet
		return true
	}

	return false
}

// 玩家投注类
type PlayerBet struct {
	Player *player.Player

	CurBets    CurBets           //当前区域总投注
	CurAccount player.CurAccount //当前的帐目

	BeforeBalance uint //投注前的余额
	AfterBalance  uint //投注后的余额

	CurWin WinType
	Digits int

	BetList []betnode.BetNode //投注的列表
}

func (p *PlayerBet) Init(player *player.Player) {

	p.CurBets = CurBets{}
	p.Player = player
	p.Digits = int(player.GetDigits())
	p.CurBets.Digits = p.Digits
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

// 读取有投注的区域及投注,现版本只检测三个区域
func (p *PlayerBet) GetValidBets() map[int]uint64 {

	bets := make(map[int]uint64)

	//检测有投注的区域
	for i := 0; i < MaxIdNum; i++ {

		if p.CurBets.Bets[i] > 0 {
			id, ok := convertToGlobalId(i)
			if ok {
				bets[id] = uint64(p.CurBets.Bets[i])
			}
		}
	}

	return bets
}

func (p *PlayerBet) TotalBet() WinType {

	total := WinType(0)

	for i := 0; i < MaxIdNum; i++ {
		total += p.CurBets.Bets[i]
	}

	return total

}

func (p *PlayerBet) IsWin() bool {

	return p.CurWin > 0
}

func (p *PlayerBet) DupliBet(pp *PlayerBet) {

	p.Player = pp.Player
	p.CurBets = pp.CurBets

	p.Digits = pp.Digits

}

func (p *PlayerBet) GetAllBets() []betnode.BetNode {

	return p.BetList
}

// 拷贝投注明细数组
func (p *PlayerBet) GetBetNode() []betnode.BetNode {

	return p.BetList
}

// 复制一份压注拷备
func (p *PlayerBet) DuplicateBet(pp *PlayerBet) {

	p.CurBets = pp.CurBets

}

// 投注
func (p *PlayerBet) Bet(areaId int, betPoint uint32) bool {

	//需要把全局投注区域转换成本游戏对应的ID
	index, bok := convertToSicboIndex(areaId)

	if bok == false {
		return false
	}

	p.CurBets.Bet(index, WinType(betPoint))

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
	//digits := p.BetList[ilen].Digits

	//需要把全局投注区域转换成本游戏对应的ID
	index, bok := convertToSicboIndex(areaId)

	if bok == false {
		return betNode, false
	}

	//扣除点数
	p.CurBets.Bets[index] -= WinType(betPoint)

	betNode = p.BetList[ilen]

	//去除最后一个投注
	p.BetList = p.BetList[:ilen]

	return betNode, true
}

// 检测是否小于低注
func (p *PlayerBet) IsLessLowBet(areaId int) bool {

	index, iok := convertToSicboIndex(areaId)

	if iok == false {
		return false
	}

	//先检测投注区域是否存在
	betlimit, ok := p.Player.SettingData.BetSet.BetAreas[areaId]

	if ok == false {
		return false
	}

	lowbet := uint64(float64(betlimit.LowBet) * float64(p.Player.GetDigits()))
	blowbet := false

	if p.CurBets.Bets[index] < WinType(lowbet) {
		blowbet = true
	}

	return blowbet
}

// 检测是否超出高注
func (p *PlayerBet) IsOverHighBet(areaId int) bool {

	betid, bok := convertToSicboIndex(areaId)

	if bok == false {
		return false
	}

	//检测高低注区域是否存在
	betlimit, ok := p.Player.SettingData.BetSet.BetAreas[areaId]

	if ok == false {
		return false
	}

	//这个数是小数位已相城的
	highbet := uint64(float64(betlimit.HighBet) * float64(p.Digits))

	bhighbet := false

	//检测龙区域压注是否小于低注
	if p.CurBets.Bets[betid] > WinType(highbet) {
		bhighbet = true
	}

	return bhighbet
}

// 检测是否超出台红
func (p *PlayerBet) IsOverTableBet(areaId int) bool {

	return true
}

// 区域1结算， 用于大小单双压注
func (p *PlayerBet) OneIdWin(win int, bet WinType, rgnSub int) WinType {

	//先检测是否投中此区域内的
	if !(rgnSub >= 0 && rgnSub < OneIdNum) {
		return 0
	}

	//检测开奖是否正确的开奖数组索引
	if !(win >= 0 && win < TotalAnswer) {
		return 0
	}

	//先检测开奖三个点是否一样
	if SicboAnswerInds[win][0] == SicboAnswerInds[win][1] && SicboAnswerInds[win][0] == SicboAnswerInds[win][2] {
		return 0
	}
	//计算点数和
	pointnum := SicboAnswerInds[win][0] + SicboAnswerInds[win][1] + SicboAnswerInds[win][2]

	bwin := false

	switch rgnSub {

	case 0: //大
		if pointnum >= 11 && pointnum <= 17 {
			bwin = true
		}
	case 1: //小
		if pointnum >= 4 && pointnum <= 10 {
			bwin = true
		}
	case 2: //单
		if pointnum%2 == 1 {
			bwin = true
		}
	case 3: //双
		if pointnum%2 == 0 {
			bwin = true
		}

	}

	//检测是否有中奖了
	if bwin == true {
		//得到赔率
		odds, digits := betOdds.GetBetOdds(betOdds.Bet_Sicbo_One_Big)

		betwin := bet * WinType(odds) //64位

		winPoint := WinType(float64(betwin) / float64(digits)) //实际胜额

		winPoint = winPoint + WinType(bet) //本金加上

		return winPoint
	}

	return 0
}

// 区域2结算， 用于特定2个数组合
func (p *PlayerBet) TwoSpecIdWin(win int, bet WinType, rgnSub int) WinType {

	//先检测是否投中此区域内的
	if !(rgnSub >= 0 && rgnSub < TwoSpecIdNum) {
		return 0
	}

	//检测开奖是否正确的开奖数组索引
	if !(win >= 0 && win < TotalAnswer) {
		return 0
	}

	pointA := SicboAnswerInds[win][0]
	pointB := SicboAnswerInds[win][1]
	pointC := SicboAnswerInds[win][2]

	//先检测开奖三个点是否一样
	if pointA == pointB && pointC == pointB {
		return 0
	}

	point0 := SicboAnswerTwoInds[rgnSub][0]
	point1 := SicboAnswerTwoInds[rgnSub][1]

	danswer := 0
	bwin := false

	//先检测开答索引表中第一个数是否 特别的两个数之一
	if pointA == point0 || pointA == point1 {
		danswer = pointA
	}

	//检测开答索引表中第二个数是否 特别的两个数之一
	if pointB == point0 || pointB == point1 {

		//若是，则再检测第一个数是否也是其中之一
		if danswer != 0 {
			//两个数不能相等
			if pointB != danswer {
				bwin = true
			}
		} else {
			danswer = pointB
		}
	}

	//检测第三个数是否特别的两个数之一
	if pointC == point0 || pointC == point1 {

		//若是，则再检测第一个数是否也是其中之一
		if danswer != 0 {
			//两个数不能相等
			if pointC != danswer {
				bwin = true
			}
		}
	}

	//检测是否有中奖了
	if bwin == true {
		//得到赔率
		//fmt.Printf(" two win ok :  win:%d, answer:%v ,sub:%d   dict:[%v] success *********\n ", win, SicboAnswerInds[win], rgnSub, SicboAnswerTwoInds[rgnSub])

		odds, digits := betOdds.GetBetOdds(betOdds.Bet_Sicbo_TwoNum_1And2)

		betwin := bet * WinType(odds) //64位

		winPoint := WinType(float64(betwin) / float64(digits)) //实际胜额

		winPoint = winPoint + WinType(bet) //本金加上

		return winPoint
	}

	return 0
}

// 区域3结算， 用于特定3数点相同
func (p *PlayerBet) ThreeSameIdWin(win int, bet WinType, rgnSub int) WinType {

	//先检测是否投中此区域内的
	if !(rgnSub >= 0 && rgnSub < ThreeSameIdNum) {
		return 0
	}
	//检测开奖是否正确的开奖数组索引
	if !(win >= 0 && win < TotalAnswer) {
		return 0
	}
	pointA := SicboAnswerInds[win][0]
	pointB := SicboAnswerInds[win][1]
	pointC := SicboAnswerInds[win][2]

	//先检测开奖三个点是否一样
	if !(pointA == pointB && pointC == pointB) {
		return 0
	}

	var pointcount [6]int

	for i := 0; i < 6; i++ {
		if i+1 == pointA {
			pointcount[i]++
		}
		if i+1 == pointB {
			pointcount[i]++
		}
		if i+1 == pointC {
			pointcount[i]++
		}
	}

	bwin := false

	//检测哪个点数有三个一样的出现
	if pointcount[rgnSub] == 3 {
		bwin = true
	}

	if bwin == true {
		//得到赔率
		odds, digits := betOdds.GetBetOdds(betOdds.Bet_Sicbo_ThreeSame_1All)

		betwin := bet * WinType(odds) //64位

		winPoint := WinType(float64(betwin) / float64(digits)) //实际胜额

		winPoint = winPoint + WinType(bet) //本金加上

		fmt.Printf(" two win ok :  win:%d, answer:%v ,sub:[%d,%d,%d]  success *********\n ", win, SicboAnswerInds[win], rgnSub+1, rgnSub+1, rgnSub+1)

		return winPoint
	}

	return 0
}

// 区域4结算， 用于任意3数点相同
func (p *PlayerBet) ThreeAnySameIdWin(winner int, bet WinType, rgnSub int) WinType {

	//先检测是否投中此区域内的
	if !(rgnSub >= 0 && rgnSub < ThreeAnySameIdNum) {
		return 0
	}
	//检测开奖是否正确的开奖数组索引
	if !(winner >= 0 && winner < TotalAnswer) {
		return 0
	}
	pointA := SicboAnswerInds[winner][0]
	pointB := SicboAnswerInds[winner][1]
	pointC := SicboAnswerInds[winner][2]

	//先检测开奖三个点是否一样
	if !(pointA == pointB && pointC == pointB) {
		return 0
	}

	//得到赔率
	odds, digits := betOdds.GetBetOdds(betOdds.Bet_Sicbo_ThreeSame_Any)

	betwin := bet * WinType(odds) //64位

	winPoint := WinType(float64(betwin) / float64(digits)) //实际胜额

	winPoint = winPoint + WinType(bet) //本金加上

	fmt.Printf(" two win ok :  win:%d, answer:%v ,sub:[%d,%d,%d]  success *********\n ", winner, SicboAnswerInds[winner], rgnSub+1, rgnSub+1, rgnSub+1)

	return winPoint
}

// 区域5结算， 总数点和结算
func (p *PlayerBet) TotalPointNumWin(win int, bet WinType, rgnSub int) WinType {

	//先检测是否投中此区域内的
	if !(rgnSub >= 0 && rgnSub < SumTotalIdNum) {
		return 0
	}
	//检测开奖是否正确的开奖数组索引
	if !(win >= 0 && win < TotalAnswer) {
		return 0
	}

	//得到总点数
	pointTotal := SicboAnswerInds[win][0] + SicboAnswerInds[win][1] + SicboAnswerInds[win][2]

	if pointTotal == rgnSub+4 {

		betid := betOdds.Bet_Sicbo_SumPoint_4 + rgnSub
		//得到赔率
		odds, digits := betOdds.GetBetOdds(betid)
		betwin := bet * WinType(odds)                          //64位
		winPoint := WinType(float64(betwin) / float64(digits)) //实际胜额
		winPoint = winPoint + WinType(bet)                     //本金加上

		return winPoint
	}

	return 0
}

// 区域6结算， 指定对子中奖
func (p *PlayerBet) PairPointWin(win int, bet WinType, rgnSub int) WinType {

	//先检测是否投中此区域内的
	if !(rgnSub >= 0 && rgnSub < PairIdNum) {
		return 0
	}

	//检测开奖是否正确的开奖数组索引
	if !(win >= 0 && win < TotalAnswer) {
		return 0
	}

	pointA := SicboAnswerInds[win][0]
	pointB := SicboAnswerInds[win][1]
	pointC := SicboAnswerInds[win][2]

	//先检测开奖三个点是否一样,都不一样则不处理
	if pointA != pointB && pointC != pointB && pointA != pointC {
		return 0
	}

	var pointcount [6]int

	for i := 0; i < 6; i++ {
		if i+1 == pointA {
			pointcount[i]++
		}
		if i+1 == pointB {
			pointcount[i]++
		}
		if i+1 == pointC {
			pointcount[i]++
		}
	}

	if pointcount[rgnSub] >= 2 {

		//得到赔率
		odds, digits := betOdds.GetBetOdds(betOdds.Bet_Sicbo_PairSame_1)

		betwin := bet * WinType(odds)                          //64位
		winPoint := WinType(float64(betwin) / float64(digits)) //实际胜额
		winPoint = winPoint + WinType(bet)                     //本金加上

		return winPoint
	}

	return 0
}

// 区域7结算， 指定数字中奖
func (p *PlayerBet) NumberPointWin(win int, bet WinType, rgnSub int) WinType {

	//先检测是否投中此区域内的
	if !(rgnSub >= 0 && rgnSub < PairIdNum) {
		return 0
	}

	//检测开奖是否正确的开奖数组索引
	if !(win >= 0 && win < TotalAnswer) {
		return 0
	}

	var pointcount [6]int

	pointA := SicboAnswerInds[win][0] - 1
	pointB := SicboAnswerInds[win][1] - 1
	pointC := SicboAnswerInds[win][2] - 1

	pointcount[pointA]++
	pointcount[pointB]++
	pointcount[pointC]++

	multi := [3]int{1, 2, 3} //赔率，这个是包括本金的

	if pointcount[rgnSub] > 0 {

		//点数个数
		num := pointcount[rgnSub] - 1

		//得到胜额
		betwin := bet * WinType(multi[num])

		winPoint := betwin + WinType(bet) //本金加上

		return winPoint
	}

	return 0
}

// 开奖并结算
func (p *PlayerBet) WinLost(win int) WinType {

	totalWin := WinType(0)
	winPoint := WinType(0)

	for i := 0; i < MaxIdNum; i++ {

		if p.CurBets.Bets[i] > 0 {

			areadId := betOdds.Bet_Sicbo_Start + i

			rgnMain, rgnSub, ret := convertToSicboArea2D(areadId)

			if ret == false {
				continue
			}

			winPoint = 0

			switch rgnMain {
			case OneId: //大小单双
				winPoint = p.OneIdWin(win, p.CurBets.Bets[i], rgnSub)
			case TwoSpecId: //特别双数
				winPoint = p.TwoSpecIdWin(win, p.CurBets.Bets[i], rgnSub)
			case ThreeSameId: //指定三个数相同
				winPoint = p.ThreeSameIdWin(win, p.CurBets.Bets[i], rgnSub)
			case ThreeAnySameId: //任意三数
				winPoint = p.ThreeAnySameIdWin(win, p.CurBets.Bets[i], rgnSub)
			case SumTotalId: //总点数
				winPoint = p.TotalPointNumWin(win, p.CurBets.Bets[i], rgnSub)
			case PairId: //对子
				winPoint = p.PairPointWin(win, p.CurBets.Bets[i], rgnSub)
			case NumberId: //点数
				winPoint = p.NumberPointWin(win, p.CurBets.Bets[i], rgnSub)
			}

			if winPoint > 0 {
				totalWin += winPoint
			}
		}

	}

	p.CurWin = WinType(totalWin)

	betwin := uint64(p.CurWin - p.TotalBet())

	//更新帐目数据
	p.refreshAccountWater()
	p.refreshAccountWinLost(int(betwin))

	return WinType(totalWin)
}

// 更新帐号的洗码情况
func (p *PlayerBet) refreshAccountWater() {

	bet := uint(p.TotalBet())

	p.CurAccount.SetBet(bet)
	p.CurAccount.SetWater(bet)
	p.CurAccount.SetBetEffective(bet)

}

func (p *PlayerBet) refreshAccountWinLost(winLost int) {

	p.CurAccount.SetWin(winLost)
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

	for i := 0; i < MaxIdNum; i++ {

		areadId := betOdds.Bet_Sicbo_Start + i

		rgnMain, rgnSub, ret := convertToSicboArea2D(areadId)

		if ret == false {
			continue
		}

		winPoint := WinType(0)
		bet := WinType(100)

		switch rgnMain {
		case OneId: //大小单双
			winPoint = p.OneIdWin(win, bet, rgnSub)
		case TwoSpecId: //特别双数
			winPoint = p.TwoSpecIdWin(win, bet, rgnSub)
		case ThreeSameId: //指定三个数相同
			winPoint = p.ThreeSameIdWin(win, bet, rgnSub)
		case ThreeAnySameId: //任意三数
			winPoint = p.ThreeAnySameIdWin(win, bet, rgnSub)
		case SumTotalId: //总点数
			winPoint = p.TotalPointNumWin(win, bet, rgnSub)
		case PairId: //对子
			winPoint = p.PairPointWin(win, bet, rgnSub)
		case NumberId: //点数
			winPoint = p.NumberPointWin(win, bet, rgnSub)
		}

		//若有中奖则增加
		if winPoint > 0 {
			betIds = append(betIds, uint32(areadId))
		}

	}

	return betIds
}

func (p *PlayerBet) UpdateAccountToDB() {

	//p.Player.AccountData.AddAccount(&p.CurAccount)

	//存储到数据库
	p.Player.AccountData.AddAccountToDB(p.Player.Id, &p.CurAccount)

}

// 投注集合到数据库
func (p *PlayerBet) SaveBetCollectToDB(gm *GameTableSicbo, set int, round int, roundId string, answer *msg_struct.SicboOpenAnswer) {

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
	betCollect.Type = msg_struct.GameTypeSicbo
	bytes, _ := json.Marshal(answer)

	betCollect.AnswerTime = &nowTime
	betCollect.Answer = bytes

	//betCollect.RecordId = util.NewUnionId()
	betCollect.RecordId = util.GenerateUniqueID(util.RecordIdLength, betCollect.UserId)

	queueId := int64(2)

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
