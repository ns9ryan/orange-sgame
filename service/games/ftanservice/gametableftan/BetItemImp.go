package gameTableFtan

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
	SubType  string //二级类型

	BetList []*betnode.BetItem //投注的列表

	answer msg_struct.FtanOpenAnswer
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

// SmallBigAndSingleDoubleIdWin 小大,单双
func SmallBigAndSingleDoubleIdWin(winner int, bet int, rgnSub int) int {
	//先检测是否投中此区域内的
	if !(rgnSub >= 0 && rgnSub < SmallBigAndSingleDoubleIdNum) {
		return 0
	}

	var bWin bool
	for _, v := range SmallBigMemberAndSingleDouble[rgnSub] {
		if v == winner {
			bWin = true
			break
		}
	}

	if bWin {
		odds, digits := betOdds.GetBetOdds(betOdds.Bet_Ftan_Small_Big_And_Single_Double_Start) //得到赔率
		if digits == 0 {
			return 0 // 避免除零错误
		}

		betWin := bet * odds //64位

		winPoint := int(float64(betWin) / float64(digits)) //实际胜额

		return winPoint
	}

	return 0 // 未中奖
}

// FanIdWin 番
func FanIdWin(winner int, bet int, rgnSub int) int {

	var winPoint int

	//先检测是否投中此区域内的
	if !(rgnSub >= 0 && rgnSub < FanIdNum) {
		return 0
	}

	if FanMember[rgnSub] == winner {
		//得到赔率
		odds, digits := betOdds.GetBetOdds(betOdds.Bet_Ftan_Fan_Start)
		if digits == 0 {
			return 0 // 避免除零错误
		}

		betwin := bet * odds

		winPoint = int(float64(betwin) / float64(digits)) //实际胜额

		return winPoint
	}

	return winPoint
}

// JiaoIdWin 角
func JiaoIdWin(winner int, bet int, rgnSub int) int {

	//先检测是否投中此区域内的
	if !(rgnSub >= 0 && rgnSub < JiaoIdNum) {
		return 0
	}

	var winPoint int

	if JiaoMember[rgnSub][0] == winner || JiaoMember[rgnSub][1] == winner {

		odds, digits := betOdds.GetBetOdds(betOdds.Bet_Ftan_Jiao_Start) //得到赔率
		if digits == 0 {
			return 0 // 避免除零错误
		}

		betwin := bet * odds //64位

		winPoint = int(float64(betwin) / float64(digits)) //实际胜额

		return winPoint
	}

	return 0
}

// NianIdWin 念
func NianIdWin(winner int, bet int, rgnSub int) int {

	//先检测是否投中此区域内的
	if !(rgnSub >= 0 && rgnSub < NianIdNum) {
		return 0
	}

	var winPoint int

	if NianWinMember[rgnSub] == winner {

		odds, digits := betOdds.GetBetOdds(betOdds.Bet_Ftan_Nian_Start) //得到赔率
		if digits == 0 {
			return 0 // 避免除零错误
		}

		betwin := bet * odds //64位

		winPoint = int(float64(betwin) / float64(digits)) //实际胜额

		return winPoint
	} else if NianTieMember[rgnSub] == winner {

		return -1
	}

	return 0
}

// ThreeRedIdWin 三红
func ThreeRedIdWin(winner int, bet int, rgnSub int) int {

	//先检测是否投中此区域内的
	if !(rgnSub >= 0 && rgnSub < ThreeRedIdNum) {
		return 0
	}

	for _, v := range ThreeRedMember[rgnSub] {
		if v == winner {
			odds, digits := betOdds.GetBetOdds(betOdds.Bet_Ftan_ThreeRed_Start) //得到赔率
			if digits == 0 {
				return 0 // 避免除零错误
			}
			betwin := bet * odds

			return int(float64(betwin) / float64(digits)) //实际胜额
		}
	}
	return 0
}

// 计算胜负,返回参数，第一个为胜额，第二个有效投注
func CalcBetWin(winner int, betAreaId int, betAmout int) (int, int) {
	//检测开奖是否正确的开奖数组索引
	if !(winner >= 1 && winner <= TotalAnswer) {
		return 0, 0
	}

	betWin := 0
	effectiveBet := betAmout

	rgnMain, rgnSub, bok := convertToFtanArea(betAreaId)

	if bok == false {
		return 0, 0
	}

	switch rgnMain {
	case SmallBigAndSingleDoubleId:
		betWin = SmallBigAndSingleDoubleIdWin(winner, betAmout, rgnSub)
	case FanId:
		betWin = FanIdWin(winner, betAmout, rgnSub)
	case JiaoId:
		betWin = JiaoIdWin(winner, betAmout, rgnSub)
	case ThreeRedId:
		betWin = ThreeRedIdWin(winner, betAmout, rgnSub)
	case NianId:
		betWin = NianIdWin(winner, betAmout, rgnSub)
	}

	if betWin == 0 {
		betWin = -betAmout
	} else if betWin == -1 {
		betWin = 0
	}

	return betWin, effectiveBet
}

// 计算每局的胜负
func (betImp *BetItemImp) CalcWinLost(set int, round int, roundId string, answer *msg_struct.FtanOpenAnswer) {

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
func (betImp *BetItemImp) SaveToDB(gm *GameTableFtan) {

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
