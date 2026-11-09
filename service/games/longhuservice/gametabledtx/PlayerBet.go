package gameTableDTX

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
	DragonId  = 0 //龙区域id
	TigerId   = 1
	TieId     = 2
	TiePairId = 3 //同花和局
)

const MaxBetNumId = 4

type WinType uint64

// 转换全局的投注区域ID为本游戏压注区id
func convertToDTXId(areaId int) (int, bool) {

	var betid int
	var bok bool

	bok = true

	if areaId == betOdds.Bet_Dtx_Dragon {
		betid = DragonId
	} else if areaId == betOdds.Bet_Dtx_Tiger {
		betid = TigerId
	} else if areaId == betOdds.Bet_Dtx_Tie {
		betid = TieId
	} else if areaId == betOdds.Bet_Dtx_TiePair {
		betid = TiePairId
	} else {
		bok = false
	}
	return betid, bok
}

// 转换本游戏内的id为全局区域id
func convertToGlobalId(areaId int) (int, bool) {

	if areaId == DragonId {
		return betOdds.Bet_Dtx_Dragon, true
	}

	if areaId == TigerId {
		return betOdds.Bet_Dtx_Tiger, true
	}

	if areaId == TieId {
		return betOdds.Bet_Dtx_Tie, true
	}

	return 0, false
}

type CurBets struct { //当前区域总的投注额

	Bets   [MaxBetNumId]uint64
	Digits int
}

// 玩家投注类,每个玩家,每局都有投,会生成对应的类
type PlayerBet struct {
	Player     *player.Player
	CurBets    CurBets           //当前区域总投注
	CurAccount player.CurAccount //当局的帐目情况.2025.3.19 add

	BeforeBalance uint //投注前的余额
	AfterBalance  uint //投注后的余额

	CurWin WinType //当前胜的胜额
	Digits int     //小数点位数

	BetList []betnode.BetNode //每注投注的列表
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

// 复制一份压注拷备
func (p *PlayerBet) DuplicateBet(pp *PlayerBet) {

	p.CurBets = pp.CurBets

}

func (p *PlayerBet) TotalBet() uint64 {

	total := p.CurBets.Bets[DragonId] + p.CurBets.Bets[TigerId] + p.CurBets.Bets[TieId]

	return total
}

// 拷贝投注明细数组
func (p *PlayerBet) GetBetNode() []betnode.BetNode {

	return p.BetList
}

// 检测是否有打水,主要是对的检测
func (p *PlayerBet) IsCheckWaterBet(betId int) bool {

	//
	if betId == DragonId {
		//检测虎是否已经有压了
		if p.CurBets.Bets[TigerId] > 0 {
			return true
		}

	} else if betId == TigerId {

		//检测龙是否已经有压了
		if p.CurBets.Bets[DragonId] > 0 {
			return true
		}

	}

	return false
}

// 投注
func (p *PlayerBet) Bet(areaId int, betPoint uint32) bool {

	var betid int

	betid, bok := convertToDTXId(areaId)

	if bok == false {
		return false
	}

	//检测是否有打水
	//if p.IsCheckWaterBet(betid) == true {
	//	//return false
	//}

	p.CurBets.Bets[betid] += uint64(betPoint)

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

	//转换成本游戏内部id
	betid, bok := convertToDTXId(areaId)

	if bok == false {
		return betNode, false
	}

	//扣除点数
	p.CurBets.Bets[betid] -= uint64(betPoint)

	betNode = p.BetList[ilen]

	//去除最后一个投注
	p.BetList = p.BetList[:ilen]

	return betNode, true
}

func (p *PlayerBet) PushBet(betItem betnode.BetNode) {

	//增加投注队列
	p.BetList = append(p.BetList, betItem)

}

// 检测是否小于低注
func (p *PlayerBet) IsLessLowBet(areaId int) bool {

	betid, bok := convertToDTXId(areaId)

	if bok == false {
		return false
	}

	betlimit, ok := p.Player.SettingData.BetSet.BetAreas[areaId]

	if ok == false {
		return false
	}

	//得到低注
	lowbet := uint64(float64(betlimit.LowBet) * float64(p.Player.GetDigits()))

	var blowbet bool

	blowbet = false

	//检测龙区域压注是否小于低注
	if p.CurBets.Bets[betid] < lowbet {
		blowbet = true
	}

	return blowbet
}

// 检测是否超出高注
func (p *PlayerBet) IsOverHighBet(areaId int) bool {

	betid, bok := convertToDTXId(areaId)

	if bok == false {
		return false
	}

	betlimit, ok := p.Player.SettingData.BetSet.BetAreas[areaId]

	if ok == false {
		return false
	}

	highbet := uint64(float64(betlimit.HighBet) * float64(p.Digits))

	bhigh := false

	//检测龙区域压注是否小于低注
	if p.CurBets.Bets[betid] > highbet {
		bhigh = true
	}

	return bhigh
}

// 检测是否超出台红
func (p *PlayerBet) IsOverTableBet(areaId int) bool {

	return true
}

// 龙赢结算
func (p *PlayerBet) DragonWin() WinType {

	if p.CurBets.Bets[DragonId] > 0 {

		//得到赔率
		odds, digits := betOdds.GetBetOdds(betOdds.Bet_Dtx_Dragon)

		win := p.CurBets.Bets[DragonId] * uint64(odds) //64位

		winPoint := WinType(float64(win) / float64(digits)) //实际胜额

		winPoint = winPoint + WinType(p.CurBets.Bets[DragonId]) //本金加上

		return winPoint
	}

	return 0
}

// 虎赢结算
func (p *PlayerBet) TigerWin() WinType {

	if p.CurBets.Bets[TigerId] > 0 {

		//得到赔率
		odds, digits := betOdds.GetBetOdds(betOdds.Bet_Dtx_Tiger)

		win := p.CurBets.Bets[TigerId] * uint64(odds) //64位

		winPoint := WinType(float64(win) / float64(digits)) //实际胜额

		winPoint = winPoint + WinType(p.CurBets.Bets[TigerId]) //本金加上

		return winPoint
	}

	return 0

}

// 和赢结算
func (p *PlayerBet) TieWin() WinType {

	var winPoint WinType = 0

	if p.CurBets.Bets[TieId] > 0 {

		//得到赔率
		odds, digits := betOdds.GetBetOdds(betOdds.Bet_Dtx_Tie)

		win := p.CurBets.Bets[TieId] * uint64(odds)         //64位
		winPoint += WinType(float64(win) / float64(digits)) //实际胜额
		winPoint += WinType(p.CurBets.Bets[TieId])          //本金加上
	}

	if p.CurBets.Bets[DragonId] > 0 {

		//得到赔率
		winPoint += WinType(p.CurBets.Bets[DragonId]) / 2 //本金也加上
	}

	if p.CurBets.Bets[TigerId] > 0 {

		//得到赔率
		winPoint += WinType(p.CurBets.Bets[TigerId]) / 2
	}

	return winPoint
}

// 同花结算，现版本不用
func (p *PlayerBet) TiePairWin() WinType {

	return 0

}

// 计算胜负
func (p *PlayerBet) CalcWin(win int, betAreaId int, betAmout int) int {

	betWin := 0

	betid, bok := convertToDTXId(betAreaId)

	if bok == false {
		return 0
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

	return betWin
}

// 开奖计算胜负
func (p *PlayerBet) WinLost(win int) WinType {

	var winPoint WinType = 0

	p.ResetWin()

	switch win {
	case DragonId: //开答是long
		winPoint = p.DragonWin()
	case TigerId:
		winPoint = p.TigerWin()
	case TieId:
		winPoint = p.TieWin()
	case TiePairId:
		winPoint = p.TiePairWin()
	}

	p.CurWin = WinType(winPoint)

	betwin := uint64(p.CurWin) - p.TotalBet()

	//更新帐目数据
	p.refreshAccountWater(win)
	p.refreshAccountWinLost(int(betwin))

	return WinType(winPoint)
}

// 更新帐号的洗码情况
func (p *PlayerBet) refreshAccountWater(win int) {

	bet := uint(p.TotalBet())

	p.CurAccount.SetBet(bet)

	if win != TieId {

		p.CurAccount.SetWater(bet)
		p.CurAccount.SetBetEffective(bet)

	} else {

		//开和，有压注龙和虎只能算一半洗码
		water := uint((p.CurBets.Bets[DragonId] + p.CurBets.Bets[TigerId]) / 2)

		p.CurAccount.SetWater(water)
		p.CurAccount.SetBetEffective(water)
	}
}

func (p *PlayerBet) refreshAccountWinLost(winLost int) {

	p.CurAccount.SetWin(winLost)
}

func (p *PlayerBet) ResetWin() {
	p.CurWin = 0
}

// 检测玩家是否有赢额
func (p *PlayerBet) IsWin() bool {

	return p.CurWin > 0
}

func (p *PlayerBet) GetAllBets() []betnode.BetNode {

	return p.BetList
}

// 读取有投注的区域及投注,现版本只检测三个区域
func (p *PlayerBet) GetValidBets() map[int]uint64 {

	bets := make(map[int]uint64)

	if p.CurBets.Bets[DragonId] > 0 {

		id, ok := convertToGlobalId(DragonId)

		if ok == true {
			bets[id] = p.CurBets.Bets[DragonId]
		}
	}

	if p.CurBets.Bets[TigerId] > 0 {

		id, ok := convertToGlobalId(TigerId)

		if ok == true {
			bets[id] = p.CurBets.Bets[TigerId]
		}
	}

	if p.CurBets.Bets[TieId] > 0 {

		id, ok := convertToGlobalId(TieId)

		if ok == true {
			bets[id] = p.CurBets.Bets[TieId]
		}
	}

	return bets
}

func (p *PlayerBet) UpdateAccountToDB() {

	//存储到钱包服数据库
	if p.Player.AddAccount(&p.CurAccount) == false {

		//若调用不成功，则自已更新到数据库
		p.Player.AccountData.AddAccountToDB(p.Player.Id, &p.CurAccount)

	}

}

func (p *PlayerBet) SaveBetCollectToDB(gm *GameTableDTX, set int, round int, roundId string, answer *msg_struct.GameOpenAnswerDTX) {

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
	betCollect.Type = msg_struct.GameTypeDtx
	bytes, _ := json.Marshal(answer.DTXAnswerCard)

	betCollect.AnswerTime = &nowTime
	betCollect.Answer = bytes

	betCollect.RecordId = util.NewUnionId()

	err := betCollect.Create()

	if err != nil {
		//存数据库不成功
	}

}

// 投注集合到数据库
func (p *PlayerBet) SaveBetCollectToDBAsync(gm *GameTableDTX, set int, round int, roundId string, answer *msg_struct.GameOpenAnswerDTX) {

	betCount := len(p.BetList)
	betAmout := p.TotalBet()

	nowTime := time.Now()

	fmt.Println("SaveBetCollectToDBAsync-------------------")

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
	betCollect.Type = msg_struct.GameTypeDtx
	bytes, _ := json.Marshal(answer.DTXAnswerCard)

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
