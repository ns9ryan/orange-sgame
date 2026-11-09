package gameTableDTX

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

	answer msg_struct.GameOpenAnswerDTX
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

// 计算胜负,返回参数，第一个为胜额，第二个有效投注
func CalcBetWin(win int, betAreaId int, betAmout int) (int, int) {

	betWin := 0
	effectiveBet := betAmout

	betid, bok := convertToDTXId(betAreaId)

	if bok == false {
		return 0, 0
	}

	switch betid {

	case DragonId: //若开奖是龙

		if win == DragonId {
			//得到赔率
			odds, digits := betOdds.GetBetOdds(betOdds.Bet_Dtx_Dragon)
			win := betAmout * odds
			betWin = win / digits

		} else if win == TigerId {
			betWin = -betAmout
		} else if win == TieId { //开和退一半水

			betWin = -betAmout / 2
			effectiveBet = betAmout / 2
		}

	case TigerId:

		if win == TigerId {
			//得到赔率
			odds, digits := betOdds.GetBetOdds(betOdds.Bet_Dtx_Tiger)
			win := betAmout * odds
			betWin = win / digits

		} else if win == DragonId {

			betWin = -betAmout

		} else if win == TieId {

			betWin = -betAmout / 2
			effectiveBet = betAmout / 2
		}

	case TieId:

		if win == TieId {
			//得到赔率
			odds, digits := betOdds.GetBetOdds(betOdds.Bet_Dtx_Tie)
			win := betAmout * odds
			betWin = win / digits

		} else {
			betWin = -betAmout
		}
	}

	return betWin, effectiveBet
}

// 计算每局的胜负
func (betImp *BetItemImp) CalcWinLost(set int, round int, roundId string, answer *msg_struct.GameOpenAnswerDTX) {

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
func (betImp *BetItemImp) SaveToDB(gm *GameTableDTX) {

	bytes, _ := json.Marshal(&betImp.answer.DTXAnswerCard)

	//for _, betItem := range betImp.BetList {
	//
	//	betRecord := models.SBetRecord{}
	//
	//	betRecord.Copy(betItem)
	//	betRecord.Answer = bytes
	//
	//	err := betRecord.Create()
	//
	//	if err != nil {
	//		//存数据表不成功
	//	}
	//}

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
