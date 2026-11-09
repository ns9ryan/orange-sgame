package gameTableBacx

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

	answer msg_struct.GameOpenAnswerBacx //当前的开答内容
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
func CalcBetWin(settleWin *AnswerSettle, betAreaId int, betAmout int) (int, int) {

	betid, bok := convertToBacxId(betAreaId)

	if bok == false {
		return 0, 0
	}

	win := settleWin.Winner
	betWin := -betAmout
	betEffective := betAmout

	switch betid {

	case PlayerId: //若投注了闲区域

		if win == PlayerId {
			//得到赔率
			odds, digits := betOdds.GetBetOdds(betOdds.Bet_Bacx_Player)
			win := betAmout * odds
			betWin = win / digits

		} else if win == TieId {
			betWin = 0 //开和就退回
		}

	case BankerId: //投注了庄区域

		if win == BankerId {
			//得到赔率
			odds, digits := betOdds.GetBetOdds(betOdds.Bet_Bacx_Banker)
			win := betAmout * odds
			betWin = win / digits

		} else if win == TieId {
			betWin = 0
		}

	case TieId:
		if win == TieId {
			//得到赔率
			odds, digits := betOdds.GetBetOdds(betOdds.Bet_Bacx_Tie)
			win := betAmout * odds
			betWin = win / digits
		}
	case PPairId: //闲对子
		if settleWin.PPair == true {
			odds, digits := betOdds.GetBetOdds(betOdds.Bet_Bacx_PPair)
			win := betAmout * odds
			betWin = win / digits
		}
	case BPairId: //庄对子
		if settleWin.BPair == true {
			odds, digits := betOdds.GetBetOdds(betOdds.Bet_Bacx_BPair)
			win := betAmout * odds
			betWin = win / digits
		}
	case BankerNocommId: //免佣的庄

		if win == BankerId {

			//若是以6点胜，则赔付只有一半
			if settleWin.SuperSix == true {

				//得到赔率
				odds, digits := betOdds.GetBetOdds(betOdds.Bet_Bacx_BankerNocomm)
				win := (betAmout * odds) / 2
				betWin = win / digits

			} else {

				//得到赔率
				odds, digits := betOdds.GetBetOdds(betOdds.Bet_Bacx_BankerNocomm)
				win := (betAmout * odds)
				betWin = win / digits
			}

		} else if win == TieId {
			betWin = 0
		}

	case SuperSixId: //超级Super

		if settleWin.SuperSix == true {
			odds, digits := betOdds.GetBetOdds(betOdds.Bet_Bacx_SuperSix)
			win := betAmout * odds
			betWin = win / digits
		}
	case AnyPairId: //任意对子
		if settleWin.AnyPair == true {
			odds, digits := betOdds.GetBetOdds(betOdds.Bet_Bacx_AnyPair)
			win := betAmout * odds
			betWin = win / digits
		}
	case PerfectPairId: //完美对子
		if settleWin.PerfectPair == true {
			odds, digits := betOdds.GetBetOdds(betOdds.Bet_Bacx_PerfectPair)
			win := betAmout * odds
			betWin = win / digits
		}
	case BigId: //大
		if settleWin.Big == true {
			odds, digits := betOdds.GetBetOdds(betOdds.Bet_Bacx_Big)
			win := betAmout * odds
			betWin = win / digits
		}
	case SmallId: //小

		if settleWin.Big == false {
			odds, digits := betOdds.GetBetOdds(betOdds.Bet_Bacx_Small)

			win := betAmout * odds
			betWin = win / digits
		}

	case PlayerBonusId: //闲龙宝

		//要先检测和
		if win == TieId {

			//检测是否例牌的8，9点和局
			if settleWin.PlayerPoint == 8 || settleWin.PlayerPoint == 9 {

				//若是例牌，则退回本金
				if settleWin.PlayerCardNum == 2 {
					betWin = 0
				}
			}

		} else if settleWin.PlayerBonus { //检测开奖是否闲龙宝范围

			odds, digits := getPlayerBonusOdds(settleWin.PlayerPoint, settleWin.PlayerCardNum, settleWin.BankPoint, settleWin.BankCardNum)

			if odds > 0 {
				win := betAmout * odds
				betWin = win / digits
			}
		}

	case BankerBonusId: //庄龙宝

		//一定要先检测和
		if win == TieId {

			//检测是否例牌的8，9点和局
			if settleWin.BankPoint == 8 || settleWin.BankPoint == 9 {

				//若是例牌，则退回本金
				if settleWin.BankCardNum == 2 {
					betWin = 0
				}
			}

		} else if settleWin.BankBonus {

			odds, digits := getBankerBonusOdds(settleWin.PlayerPoint, settleWin.PlayerCardNum, settleWin.BankPoint, settleWin.BankCardNum)

			if odds > 0 {
				win := betAmout * odds
				betWin = win / digits
			}

		}

	}

	return betWin, betEffective
}

func getPlayerBonusOdds(playPoint int, pcardNum int, bankPoint int, bcardNum int) (odds, digits int) {

	//处理闲家只有二家牌的情况
	if pcardNum == 2 {

		//8点或9点胜
		if playPoint == 8 || playPoint == 9 {

			//得到赔率
			odds, digits := betOdds.GetBonusOdds(betOdds.BonusWin8or9NRC)

			return odds, digits
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
		return 0, 1
	}

	//得到赔率
	odds, digits = betOdds.GetBonusOdds(winType)

	return odds, digits
}

// 庄龙宝结算
func getBankerBonusOdds(playPoint int, pcardNum int, bankPoint int, bcardNum int) (odds, digits int) {

	//若是和局，要看是例牌还是非例牌
	//if playPoint == bankPoint {
	//
	//	_, digits := betOdds.GetBonusOdds(betOdds.BonusWin8or9NRC)
	//
	//	if bcardNum == 2 {
	//
	//		//检测是否是例牌8，9平
	//		if bankPoint == 8 || bankPoint == 9 {
	//			return 0, digits
	//		}
	//	}
	//
	//	return -1, digits
	//}

	if bcardNum == 2 {

		//8点或9点胜
		if bankPoint == 8 || bankPoint == 9 {

			//得到赔率
			odds, digits := betOdds.GetBonusOdds(betOdds.BonusWin8or9NRC)

			return odds, digits
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
		return 0, 1 //此处是错的
	}

	//得到赔率
	odds, digits = betOdds.GetBonusOdds(winType)

	return odds, digits
}

// 计算每局的胜负
func (betImp *BetItemImp) CalcWinLost(settleWin *AnswerSettle, set int, round int, roundId string, answer *msg_struct.GameOpenAnswerBacx) {

	for _, betItem := range betImp.BetList {

		//计算每笔投注的胜负额
		betItem.WinLost, betItem.EffectiveBet = CalcBetWin(settleWin, betItem.BetAreaId, int(betItem.BetPoint))

		betItem.Winner = answer.Winner.Winner
		betItem.Val1 = answer.Winner.PPair
		betItem.Val2 = answer.Winner.BPair

		betItem.Answer_Timer = time.Now()
		betItem.Round_Id = roundId
		betItem.Round = round
		betItem.Set = set
	}

	betImp.answer = *answer
}

// 投注明细入库
func (betImp *BetItemImp) SaveToDB(gm *GameTableBacx) {

	bytes, _ := json.Marshal(&betImp.answer.BacxAnswerCard)

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
				log.Errorf("存储投注明细到数据库不成功 %v   ", betImp)
			}

			return false

		}, nil)
	}

}
