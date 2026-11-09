package gameTableSicbo

import (
	"encoding/json"
	"github.com/duanhf2012/origin/v2/log"
	"sort"
	"time"
	"videogame/common/betnode"
	betOdds "videogame/common/betodds"
	models "videogame/common/model"
	"videogame/common/msg_struct"
)

// 存储本游戏全部会员的每局的投注明细
type BetItemImp struct {
	GameType string //游戏类型
	SubType  string //二级游戏类型

	BetList []*betnode.BetItem //投注的列表

	answer msg_struct.SicboOpenAnswer
}

func CreateBetItemImp(gameType string, subType string) *BetItemImp {
	return &BetItemImp{GameType: gameType, SubType: subType}
}

func (betImp *BetItemImp) Copy(ppbet *PlayerBet) {

	betNodes := ppbet.GetBetNode()

	for _, betNode := range betNodes {
		betImp.PushItem(&betNode)
	}

}

func (betImp *BetItemImp) PushItem(node *betnode.BetNode) {

	betItem := betnode.BetItem{} //以后需要优化成内存池方式

	betItem.Build(betImp.GameType, betImp.SubType, node)

	betImp.BetList = append(betImp.BetList, &betItem)
}

// 按时间排列
func (betImp *BetItemImp) SortAsce() {

	sort.SliceStable(betImp.BetList, func(i, j int) bool {

		if betImp.BetList[i].BetTime.Before(betImp.BetList[j].BetTime) == true {
			return true
		}
		return false
	})
}

// 区域1结算， 用于大小单双压注
func OneIdWin(win int, bet int, rgnSub int) int {

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

		betwin := bet * odds //64位

		winPoint := int(float64(betwin) / float64(digits)) //实际胜额

		return winPoint
	}

	return 0
}

// 区域2结算， 用于特定2个数组合
func TwoSpecIdWin(win int, bet int, rgnSub int) int {

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
		odds, digits := betOdds.GetBetOdds(betOdds.Bet_Sicbo_TwoNum_1And2)
		win := bet * int(odds)                          //64位
		winPoint := int(float64(win) / float64(digits)) //实际胜额

		return winPoint
	}

	return 0
}

// 区域3结算， 用于特定3数点相同
func ThreeSameIdWin(win int, bet int, rgnSub int) int {

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
		betwin := bet * int(odds)                          //64位
		winPoint := int(float64(betwin) / float64(digits)) //实际胜额

		//	fmt.Printf(" two win ok :  win:%d, answer:%v ,sub:[%d,%d,%d]  success *********\n ", win, SicboAnswerInds[win], rgnSub+1, rgnSub+1, rgnSub+1)

		return winPoint
	}

	return 0
}

// 区域4结算， 用于任意3数点相同
func ThreeAnySameIdWin(winner int, bet int, rgnSub int) int {

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
	betwin := bet * int(odds)                          //64位
	winPoint := int(float64(betwin) / float64(digits)) //实际胜额

	//fmt.Printf(" two win ok :  win:%d, answer:%v ,sub:[%d,%d,%d]  success *********\n ", win, SicboAnswerInds[win], rgnSub+1, rgnSub+1, rgnSub+1)

	return winPoint
}

// 区域5结算， 总数点和结算
func TotalPointNumWin(win int, bet int, rgnSub int) int {

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
		betwin := bet * int(odds)                          //64位
		winPoint := int(float64(betwin) / float64(digits)) //实际胜额

		return winPoint
	}

	return 0
}

// 区域6结算， 指定对子中奖
func PairPointWin(win int, bet int, rgnSub int) int {

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

		betwin := bet * int(odds)                          //64位
		winPoint := int(float64(betwin) / float64(digits)) //实际胜额

		return winPoint
	}

	return 0
}

// 区域7结算， 指定数字中奖
func NumberPointWin(win int, bet int, rgnSub int) int {

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

	multi := [3]int{1, 2, 3} //赔率

	if pointcount[rgnSub] > 0 {

		//点数个数
		num := pointcount[rgnSub] - 1

		//得到赔率
		betwin := bet * multi[num]

		return betwin
	}

	return 0
}

// 计算胜负,返回参数，第一个为胜额，第二个有效投注
func CalcBetWin(win int, betAreaId int, betAmout int) (int, int) {

	betWin := 0
	effectiveBet := betAmout

	rgnMain, rgnSub, bok := convertToSicboArea2D(betAreaId)

	if bok == false {
		return 0, 0
	}

	switch rgnMain {

	case OneId: //若开奖是龙
		betWin = OneIdWin(win, betAmout, rgnSub)
	case TwoSpecId:
		betWin = TwoSpecIdWin(win, betAmout, rgnSub)
	case ThreeSameId: //区域3
		betWin = ThreeSameIdWin(win, betAmout, rgnSub)
	case ThreeAnySameId: //区域4
		betWin = ThreeAnySameIdWin(win, betAmout, rgnSub)
	case SumTotalId: //区域 5
		betWin = TotalPointNumWin(win, betAmout, rgnSub)
	case PairId: //区域6
		betWin = PairPointWin(win, betAmout, rgnSub)
	case NumberId: //区域7
		betWin = NumberPointWin(win, betAmout, rgnSub)
	}

	if betWin == 0 {
		betWin = -betAmout
	}

	return betWin, effectiveBet
}

// 计算每局的胜负
func (betImp *BetItemImp) CalcWinLost(set int, round int, roundId string, answer *msg_struct.SicboOpenAnswer) {

	for _, betItem := range betImp.BetList {

		//计算胜负
		betItem.WinLost, betItem.EffectiveBet = CalcBetWin(answer.Winner, betItem.BetAreaId, int(betItem.BetPoint))

		betItem.Winner = answer.Winner
		betItem.Answer_Timer = time.Now()
		betItem.Round_Id = roundId
		betItem.Round = round
		betItem.Set = set
	}

	betImp.answer = *answer

}

// 投注明细入库
func (betImp *BetItemImp) SaveToDB(gm *GameTableSicbo) {

	bytes, _ := json.Marshal(&betImp.answer.Winner)

	queueId := int64(1)

	//启用独立协程存储
	for _, betItem := range betImp.BetList {

		gm.AsyncDoByQueue(queueId, func() bool {

			betRecord := models.SBetRecord{}

			betRecord.Copy(betItem)
			betRecord.Answer = bytes

			err := betRecord.Create()

			if err != nil {

				//存数据表不成功
				log.Errorf("存储数据库不成功 %v", betImp)
			}

			return false

		}, nil)
	}

}
