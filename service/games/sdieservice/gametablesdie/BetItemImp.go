package gameTableSdie

import (
	"encoding/json"
	"fmt"
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

	answer msg_struct.SdieOpenAnswer
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

// SingleDoubleAndSmallBigIdWin 单双,小大区域
func SingleDoubleAndSmallBigIdWin(winner int, bet int, rgnSub int) int {
	//先检测是否投中此区域内的
	if !(rgnSub >= 0 && rgnSub < SingleDoubleAndSmallBigNum) {
		return 0
	}

	var bWin bool
	for _, v := range SingleDoubleAndSmallBigMember[rgnSub] {
		if v == winner {
			bWin = true
			break
		}
	}

	if bWin {
		odds, digits := betOdds.GetBetOdds(betOdds.Bet_Sdie_Single) //得到赔率
		if digits == 0 {
			return 0 // 避免除零错误
		}

		betWin := bet * odds //64位

		winPoint := int(float64(betWin) / float64(digits)) //实际胜额

		return winPoint
	}

	return 0 // 未中奖
}
func FourWhiteFourRedIdWin(winner int, bet int, rgnSub int) int {
	//先检测是否投中此区域内的
	if !(rgnSub >= 0 && rgnSub < FourWhiteFourRedIdNum) {
		return 0
	}

	var bWin bool
	v := FourWhiteFourRedMember[rgnSub]
	if v == winner {
		bWin = true
	}

	if bWin {
		odds, digits := betOdds.GetBetOdds(betOdds.Bet_Sdie_0Red_4White) //得到赔率
		if digits == 0 {
			return 0 // 避免除零错误
		}

		betWin := bet * odds //64位

		winPoint := int(float64(betWin) / float64(digits)) //实际胜额

		return winPoint
	}

	return 0 // 未中奖
}

func ThreeWhiteThreeRedIdWin(winner int, bet int, rgnSub int) int {
	//先检测是否投中此区域内的
	if !(rgnSub >= 0 && rgnSub < ThreeWhiteThreeRedIdNum) {
		return 0
	}

	var bWin bool
	v := ThreeWhiteThreeRedMember[rgnSub]
	if v == winner {
		bWin = true
	}

	if bWin {
		odds, digits := betOdds.GetBetOdds(betOdds.Bet_Sdie_1Red_3White) //得到赔率
		if digits == 0 {
			return 0 // 避免除零错误
		}

		betWin := bet * odds //64位

		winPoint := int(float64(betWin) / float64(digits)) //实际胜额

		return winPoint
	}

	return 0 // 未中奖
}

func TwoWhiteTwoRedIdWin(winner int, bet int, rgnSub int) int {
	//先检测是否投中此区域内的
	if !(rgnSub >= 0 && rgnSub < TwoWhiteTwoRedIdNum) {
		return 0
	}

	var bWin bool
	if TwoWhiteTwoRedMember == winner {
		bWin = true
	}

	if bWin {
		odds, digits := betOdds.GetBetOdds(betOdds.Bet_Sdie_2Red_2White) //得到赔率
		if digits == 0 {
			return 0 // 避免除零错误
		}

		betWin := bet * odds //64位

		winPoint := int(float64(betWin) / float64(digits)) //实际胜额

		return winPoint
	}

	return 0 // 未中奖
}

func FourWhiteOrFourRedIdWin(winner int, bet int, rgnSub int) int {
	//先检测是否投中此区域内的
	if !(rgnSub >= 0 && rgnSub < FourWhiteOrFourRedIdNum) {
		return 0
	}

	var bWin bool
	for _, v := range FourWhiteOrFourRedMember {
		if v == winner {
			bWin = true
		}
	}

	if bWin {
		odds, digits := betOdds.GetBetOdds(betOdds.Bet_Sdie_4White_Or_4Red) //得到赔率
		if digits == 0 {
			return 0 // 避免除零错误
		}

		betWin := bet * odds //64位

		winPoint := int(float64(betWin) / float64(digits)) //实际胜额

		return winPoint
	}

	return 0 // 未中奖
}

// 计算胜负,返回参数，第一个为胜额，第二个有效投注
func CalcBetWin(winner int, betAreaId int, betAmout int) (int, int) {
	//检测开奖是否正确的开奖数组索引
	if !(winner >= 0 && winner < TotalAnswer) {
		return 0, 0
	}

	betWin := 0
	effectiveBet := betAmout

	rgnMain, rgnSub, bok := convertToSdieArea(betAreaId)

	if bok == false {
		return 0, 0
	}

	switch rgnMain {
	case SingleDoubleAndSmallBigId:
		betWin = SingleDoubleAndSmallBigIdWin(winner, betAmout, rgnSub)
	case FourWhiteFourRedId:
		betWin = FourWhiteFourRedIdWin(winner, betAmout, rgnSub)
	case ThreeWhiteThreeRedId:
		betWin = ThreeWhiteThreeRedIdWin(winner, betAmout, rgnSub)
	case TwoWhiteTwoRedId:
		betWin = TwoWhiteTwoRedIdWin(winner, betAmout, rgnSub)
	case FourWhiteOrFourRedId:
		betWin = FourWhiteOrFourRedIdWin(winner, betAmout, rgnSub)
	}

	if betWin == 0 {
		betWin = -betAmout
	}
	fmt.Printf("------------>计算投注每个投注明细的胜负------->rgnMain:%v,rgnSub:%v,betWin:%v\n", rgnMain, rgnSub, betWin)
	return betWin, effectiveBet
}

// 计算每局的胜负
func (betImp *BetItemImp) CalcWinLost(set int, round int, roundId string, answer *msg_struct.SdieOpenAnswer) {

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
func (betImp *BetItemImp) SaveToDB(gm *GameTableSdie) {

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
