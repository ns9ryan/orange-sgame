package gameTableFsc

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
// 鱼虾蟹投注区域
// 主区域代号： 单一点数：0
//=============================================================================================

const (
	SinglePointId = iota
)

type WinType uint64

// 转换全局的投注区域ID为本游戏压注区主次ID
func convertToFscArea2D(areaId int) (rgnMain, rgnSub int, ret bool) {

	ret = true

	//现在没有对越界检查代码，后续需要增加
	if areaId >= betOdds.Bet_Fsc_Single_Point1 && areaId <= betOdds.Bet_Fsc_Single_Point6 {

		rgnMain = SinglePointId
		rgnSub = areaId - betOdds.Bet_Fsc_Single_Point1

	} else {
		ret = false
		rgnMain = -1
		rgnSub = -1
	}

	return rgnMain, rgnSub, ret
}

// 转换全局区域id 到索引值
func convertToFscIndex(areaId int) (index int, ret bool) {

	if areaId >= betOdds.Bet_Fsc_Start && areaId <= betOdds.Bet_Fsc_End {
		index = areaId - betOdds.Bet_Fsc_Start
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

		idoff := betOdds.Bet_Fsc_Start + areaId

		return idoff, true
	}

	return -1, false
}

type CurBets struct { //当前区域总的投注额
	Bets [MaxIdNum]WinType //鱼虾蟹区域投注额

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

	BeforeBalance uint //压注前余额
	AfterBalance  uint //压注后余额

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

// 设置投注前的余额
func (p *PlayerBet) SetBeforeBalance(balance uint) {
	p.BeforeBalance = balance
}

// 设置开奖后余额
func (p *PlayerBet) SetAfterBalance(balance uint) {
	p.AfterBalance = balance
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
	index, bok := convertToFscIndex(areaId)

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

	//需要把全局投注区域转换成本游戏对应的ID
	index, bok := convertToFscIndex(areaId)

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

	index, iok := convertToFscIndex(areaId)

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

	betid, bok := convertToFscIndex(areaId)

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

// 区域1结算， 用于单一数压注
func (p *PlayerBet) SinglePointIdWin(win int, bet WinType, rgnSub int) WinType {

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
		betwin := bet * WinType(odds) //64位

		winPoint := WinType(float64(betwin) / float64(digits)) //实际胜额

		winPoint = winPoint + bet //本金加上

		return winPoint
	}

	return 0
}

// GetWinPoint 计算玩家赢得的金额（含本金）
func (p *PlayerBet) GetWinPoint(winner int, rgnMain int, rgnSub int, bet WinType) WinType {
	var totalWin WinType

	//检测开奖是否正确的开奖数组索引
	if !(winner >= 0 && winner < TotalAnswer) {
		return 0
	}

	switch rgnMain {

	case 0:
		totalWin = p.SinglePointIdWin(winner, bet, rgnSub)
	}

	return totalWin
}

// 开奖并结算
func (p *PlayerBet) WinLost(win int) WinType {

	totalWin := WinType(0)
	winPoint := WinType(0)

	for index, bet := range p.CurBets.Bets { //检测全部区域的投注
		if bet == 0 {
			continue
		}

		//得到实际的投注区域
		betId := betOdds.Bet_Fsc_Start + index

		rgnMain, rgnSub, ret := convertToFscArea2D(betId)
		if ret == false {
			continue
		}

		winPoint = p.GetWinPoint(win, rgnMain, rgnSub, bet)
		if winPoint > 0 {
			totalWin += winPoint
		}
	}

	p.CurWin = totalWin

	betwin := uint64(p.CurWin - p.TotalBet())

	//更新帐目数据
	p.refreshAccountWater()
	p.refreshAccountWinLost(int(betwin))

	return totalWin
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

		areadId := betOdds.Bet_Fsc_Start + i

		rgnMain, rgnSub, ret := convertToFscArea2D(areadId)

		if ret == false {
			continue
		}

		winPoint := WinType(0)
		bet := WinType(100)

		switch rgnMain {
		case SinglePointId: //大小单双
			winPoint = p.SinglePointIdWin(win, bet, rgnSub)
		}

		//若有中奖则增加
		if winPoint > 0 {
			betIds = append(betIds, uint32(areadId))
		}

	}

	return betIds
}

func (p *PlayerBet) UpdateAccountToDB() {

	//存储到数据库
	p.Player.AccountData.AddAccountToDB(p.Player.Id, &p.CurAccount)

}

// 投注集合到数据库
func (p *PlayerBet) SaveBetCollectToDB(gm *GameTableFsc, set int, round int, roundId string, answer *msg_struct.FscOpenAnswer) {

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
	betCollect.Type = msg_struct.GameTypeFSC
	bytes, _ := json.Marshal(answer)

	betCollect.AnswerTime = &nowTime
	betCollect.Answer = bytes

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
