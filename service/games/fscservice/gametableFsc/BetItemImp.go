package gameTableFsc

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

	answer msg_struct.FscOpenAnswer
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

// 区域1结算， 用于单一数压注
func SinglePointIdWin(win int, bet int, rgnSub int) int {

	//先检测是否投中此区域内的
	if !(rgnSub >= 0 && rgnSub < SinglePointIdNum) {
		return 0
	}

	//检测开奖是否正确的开奖数组索引
	if !(win >= 0 && win < TotalAnswer) {
		return 0
	}

	num := 0 // 0：表示未中奖 ，1：单骰 2：双骰 3：全骰
	for _, point := range FscAnswerInds[win] {
		if point == rgnSub+1 { //有押中
			num++
		}
	}

	if num > 0 { //中奖了

		//得到赔率
		odds, digits := betOdds.GetFscSinglePointOdds(num)
		betwin := bet * odds //64位

		winPoint := int(float64(betwin) / float64(digits)) //实际胜额

		return winPoint
	}

	return 0
}

// 计算胜负,返回参数，第一个为胜额，第二个有效投注
func CalcBetWin(win int, betAreaId int, betAmout int) (int, int) {

	betWin := 0
	effectiveBet := betAmout

	rgnMain, rgnSub, bok := convertToFscArea2D(betAreaId)

	if bok == false {
		return 0, 0
	}

	switch rgnMain {

	case SinglePointId: //若开奖是单一数区域
		betWin = SinglePointIdWin(win, betAmout, rgnSub)
	}

	if betWin == 0 {
		betWin = -betAmout
	}

	return betWin, effectiveBet
}

// 计算每局的胜负
func (betImp *BetItemImp) CalcWinLost(set int, round int, roundId string, answer *msg_struct.FscOpenAnswer) {

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
func (betImp *BetItemImp) SaveToDB(gm *GameTableFsc) {

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
