package gameTableRot

import (
	"encoding/json"
	"sort"
	"time"
	"videogame/common/betnode"
	models "videogame/common/model"
	"videogame/common/msg_struct"
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
)

// 存储本游戏全部会员的每局的投注明细
type BetItemImp struct {
	GameType string //游戏类型
	SubType  string //二级游戏类型
	BetList []*betnode.BetItem //投注的列表

	answer msg_struct.GameOpenAnswerRot
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

// 计算胜负
func CalcBetWin(win int, betAreaId int, betAmout int) (int, int) {

	betWin := 0
	effectiveBet := betAmout

	fmt.Printf("RotService Start CalcBetWin .....winner=%d\n", win)

	var winPoint WinType = 0
	rgnMain, rgnSub, bok := convertToRotId(betAreaId)

	fmt.Printf("\n\nCCCCCCCCCCCCCCCCCCCCCCCCCCCCC WinLost winner=%d, betAreaId=%d, rgnMain=%d, rgnSub=%d\n", win, betAreaId, rgnMain, rgnSub)

	if bok == false {
		return 0, 0
	}

	p := PlayerBet{}

	switch rgnMain {
	case RotSingleId:
		winPoint = p.SingleWin(win, WinType(betAmout), rgnSub)
	case RotDoubleId:
		winPoint = p.DoubleWin(win, WinType(betAmout), rgnSub)
	case RotThreeId:
		winPoint = p.ThreeWin(win, WinType(betAmout), rgnSub)
	case RotFourId:
		winPoint = p.FourWin(win, WinType(betAmout), rgnSub)
	case RotSixId:
		winPoint = p.SixWin(win, WinType(betAmout), rgnSub)
	case RotTwelveId:
		winPoint = p.TwelveWin(win, WinType(betAmout), rgnSub)
	case RotEighteenId:
		winPoint = p.EighteenWin(win, WinType(betAmout), rgnSub)
	case RotNeighbourTwoId:
		winPoint = p.NeighbourTwoWin(win, WinType(betAmout), rgnSub)
	case RotTiersId:
		winPoint = p.TiersWin(win, WinType(betAmout))
	case RotOrphelinsId:
		winPoint = p.OrphelinsWin(win, WinType(betAmout))
	case RotVoisinsId:
		winPoint = p.VoisinsWin(win, WinType(betAmout))
	case RotJeu0Id:
		winPoint = p.Jeu0Win(win, WinType(betAmout))
	}

	if winPoint > 0 {
		betWin = int(winPoint - WinType(betAmout))
	}else{
		betWin = -betAmout
	}

	return betWin, effectiveBet
}

// 计算每局的胜负
func (betImp *BetItemImp) CalcWinLost(set int, round int, roundId string, answer *msg_struct.GameOpenAnswerRot) {

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
func (betImp *BetItemImp) SaveToDB(gm *GameTableRot) {
	answerJson := msg_struct.RotAnswerJson{
		Winner:   betImp.answer.Winner,
		IsRed:    IsRed(betImp.answer.Winner),
		IsBig:    IsBig(betImp.answer.Winner),
		IsSingle: IsSingle(betImp.answer.Winner),
	}

	bytes, _ := json.Marshal(answerJson)

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
