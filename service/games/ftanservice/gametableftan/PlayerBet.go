package gameTableFtan

import (
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

type WinType uint64

// 转换全局的投注区域ID为本游戏压注区主次ID
func convertToFtanArea(areaId int) (rgnMain, rgnSub int, ret bool) {

	ret = true

	//现在没有对越界检查代码，后续需要增加
	if areaId >= betOdds.Bet_Ftan_Small_Big_And_Single_Double_Start && areaId <= betOdds.Bet_Ftan_Small_Big_And_Single_Double_End {
		rgnMain = SmallBigAndSingleDoubleId
		rgnSub = areaId - betOdds.Bet_Ftan_Small_Big_And_Single_Double_Start

	} else if areaId >= betOdds.Bet_Ftan_Fan_Start && areaId <= betOdds.Bet_Ftan_Fan_End {
		rgnMain = FanId
		rgnSub = areaId - betOdds.Bet_Ftan_Fan_Start
	} else if areaId >= betOdds.Bet_Ftan_Jiao_Start && areaId <= betOdds.Bet_Ftan_Jiao_End {
		rgnMain = JiaoId
		rgnSub = areaId - betOdds.Bet_Ftan_Jiao_Start
	} else if areaId >= betOdds.Bet_Ftan_Nian_Start && areaId <= betOdds.Bet_Ftan_Nian_End {
		rgnMain = NianId
		rgnSub = areaId - betOdds.Bet_Ftan_Nian_Start

	} else if areaId >= betOdds.Bet_Ftan_ThreeRed_Start && areaId <= betOdds.Bet_Ftan_ThreeRed_End {
		rgnMain = ThreeRedId
		rgnSub = areaId - betOdds.Bet_Ftan_ThreeRed_Start
	}

	return rgnMain, rgnSub, ret
}

// 转换全局区域id 到索引值
func convertToFtanIndex(areaId int) (index int, ret bool) {

	if areaId >= betOdds.Bet_Ftan_Start && areaId <= betOdds.Bet_Ftan_End {
		index = areaId - betOdds.Bet_Ftan_Start
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

		idoff := betOdds.Bet_Ftan_Start + areaId

		return idoff, true
	}

	return -1, false
}

type CurBets struct { //当前区域总的投注额
	Bets [MaxIdNum]WinType //番摊区域投注额

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
	CurAccount player.CurAccount //当局的帐目情况.2025.3.19 add

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

func (p *PlayerBet) DupliBet(pp *PlayerBet) {

	p.Player = pp.Player
	p.CurBets = pp.CurBets

	p.Digits = pp.Digits

}

// 设置投注前的余额
func (p *PlayerBet) SetBeforeBalance(balance uint) {
	p.BeforeBalance = balance
}

// 设置开奖后余额
func (p *PlayerBet) SetAfterBalance(balance uint) {
	p.AfterBalance = balance
}

func (p *PlayerBet) ResetWin() {
	p.CurWin = 0
}

func (p *PlayerBet) IsWin() bool {

	return p.CurWin > 0
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
	index, bok := convertToFtanIndex(areaId)

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
	index, bok := convertToFtanIndex(areaId)

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

	betid, bok := convertToFtanIndex(areaId)

	if bok == false {
		return false
	}

	//先检测投注区域是否存在
	betlimit, ok := p.Player.SettingData.BetSet.BetAreas[areaId]

	if ok == false {
		return false
	}

	lowbet := uint64(float64(betlimit.LowBet) * float64(p.Player.GetDigits()))
	blowbet := false

	if p.CurBets.Bets[betid] < WinType(lowbet) {
		blowbet = true
	}

	return blowbet
}

// 检测是否超出高注
func (p *PlayerBet) IsOverHighBet(areaId int) bool {

	betid, bok := convertToFtanIndex(areaId)

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

// FanIdWin 番
func (p *PlayerBet) FanIdWin(winner int, bet WinType, rgnSub int) WinType {

	var winPoint WinType

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

		betwin := bet * WinType(odds) //64位

		winPoint = WinType(float64(betwin) / float64(digits)) //实际胜额

		winPoint = winPoint + bet //本金加上

		return winPoint
	}

	return winPoint
}

// JiaoIdWin 角
func (p *PlayerBet) JiaoIdWin(winner int, bet WinType, rgnSub int) WinType {

	//先检测是否投中此区域内的
	if !(rgnSub >= 0 && rgnSub < JiaoIdNum) {
		return 0
	}

	var winPoint WinType

	if JiaoMember[rgnSub][0] == winner || JiaoMember[rgnSub][1] == winner {

		odds, digits := betOdds.GetBetOdds(betOdds.Bet_Ftan_Jiao_Start) //得到赔率
		if digits == 0 {
			return 0 // 避免除零错误
		}

		betwin := bet * WinType(odds) //64位

		winPoint = WinType(float64(betwin) / float64(digits)) //实际胜额

		winPoint = winPoint + bet //本金加上

		return winPoint
	}

	return 0
}

// NianIdWin 念
func (p *PlayerBet) NianIdWin(winner int, bet WinType, rgnSub int) WinType {

	//先检测是否投中此区域内的
	if !(rgnSub >= 0 && rgnSub < NianIdNum) {
		return 0
	}

	var winPoint WinType

	if NianWinMember[rgnSub] == winner {

		odds, digits := betOdds.GetBetOdds(betOdds.Bet_Ftan_Nian_Start) //得到赔率
		if digits == 0 {
			return 0 // 避免除零错误
		}

		betwin := bet * WinType(odds) //64位

		winPoint = WinType(float64(betwin) / float64(digits)) //实际胜额

		winPoint = winPoint + bet //本金加上

		return winPoint
	} else if NianTieMember[rgnSub] == winner {

		return bet //和，退回本金int
	}

	return 0
}

// ThreeRedIdWin 三红
func (p *PlayerBet) ThreeRedIdWin(winner int, bet WinType, rgnSub int) WinType {
	//fmt.Printf("-------------------------->ThreeRedIdWin winner:%v-->bet:%v-->rgnSub:%v\n", winner, bet, rgnSub)
	//先检测是否投中此区域内的
	if !(rgnSub >= 0 && rgnSub < ThreeRedIdNum) {
		return 0
	}
	//fmt.Printf("-------------------------->TThreeRedIdWin Members:%v\n", ThreeRedMember[rgnSub])

	var winPoint WinType
	for _, v := range ThreeRedMember[rgnSub] {
		if v == winner {
			odds, digits := betOdds.GetBetOdds(betOdds.Bet_Ftan_ThreeRed_Start) //得到赔率
			if digits == 0 {
				return 0 // 避免除零错误
			}
			betwin := bet * WinType(odds) //64位

			winPoint = WinType(float64(betwin) / float64(digits)) //实际胜额

			winPoint = winPoint + bet //本金加上
			return winPoint
		}
	}
	return 0
}

// SmallBigAndSingleDoubleIdWin 小大单双区域
func (p *PlayerBet) SmallBigAndSingleDoubleIdWin(winner int, bet WinType, rgnSub int) WinType {
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

		betWin := bet * WinType(odds) //64位

		winPoint := WinType(float64(betWin) / float64(digits)) //实际胜额

		winPoint = winPoint + bet //本金加上

		return winPoint
	}

	return 0 // 未中奖
}

// GetWinPoint 计算玩家赢得的金额（含本金）
func (p *PlayerBet) GetWinPoint(winner int, rgnMain int, rgnSub int, bet WinType) WinType {
	var totalWin WinType

	//检测开奖是否正确的开奖数组索引
	if !(winner >= 1 && winner <= TotalAnswer) {
		return 0
	}

	switch rgnMain {

	case 0: // 小大,单双 (1 赔 0.95)
		totalWin = p.SmallBigAndSingleDoubleIdWin(winner, bet, rgnSub)
	case 1: // 番 (1 赔 2.85)
		totalWin = p.FanIdWin(winner, bet, rgnSub)
	case 2: // 角 (1 赔 0.95)
		totalWin = p.JiaoIdWin(winner, bet, rgnSub)
	case 3: // 稔 (1 赔 1.90)
		totalWin = p.NianIdWin(winner, bet, rgnSub)

	case 4: // 压三 (3 赔 1)
		totalWin = p.ThreeRedIdWin(winner, bet, rgnSub)

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
		betId := betOdds.Bet_Ftan_Start + index

		rgnMain, rgnSub, ret := convertToFtanArea(betId)
		if ret == false {
			continue
		}

		winPoint = p.GetWinPoint(win, rgnMain, rgnSub, bet)
		if winPoint > 0 {
			totalWin += winPoint
		}
	}

	fmt.Printf("---------------->番摊--------->开奖并结算------>Userid:%v,总输赢:%v\n", p.Player.Id, totalWin)

	p.CurWin = totalWin

	betwin := uint64(p.CurWin) - uint64(p.TotalBet())

	//更新帐目数据
	p.refreshAccountWater()
	p.refreshAccountWinLost(int(betwin))

	return totalWin
}

// 更新帐号的洗码情况
func (p *PlayerBet) refreshAccountWater() {

	bet := uint(p.TotalBet())

	p.CurAccount.SetBet(bet)

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

// GlobalGetWinAreaIds 读取有中奖的区域id列表
func GlobalGetWinAreaIds(win int) []uint32 {

	ppbet := PlayerBet{}

	betids := ppbet.GetWinAreaIds(win)

	return betids
}

// GetWinAreaIds 读取有中奖的区域id列表
func (p *PlayerBet) GetWinAreaIds(win int) []uint32 {

	betIds := make([]uint32, 0)

	for i := 0; i < MaxIdNum; i++ {

		areadId := betOdds.Bet_Ftan_Start + i

		rgnMain, rgnSub, ret := convertToFtanArea(areadId)

		if ret == false {
			continue
		}

		winPoint := WinType(0)
		bet := WinType(100)

		switch rgnMain {
		case SmallBigAndSingleDoubleId: //小大，单双
			winPoint = p.SmallBigAndSingleDoubleIdWin(win, bet, rgnSub)
		case FanId: //番
			winPoint = p.FanIdWin(win, bet, rgnSub)
		case JiaoId: //角
			winPoint = p.JiaoIdWin(win, bet, rgnSub)
		case NianId: //念
			winPoint = p.NianIdWin(win, bet, rgnSub)
		case ThreeRedId: //三红
			winPoint = p.ThreeRedIdWin(win, bet, rgnSub)
		}

		//若有中奖则增加
		if winPoint > 0 {
			betIds = append(betIds, uint32(areadId))
		}

	}

	return betIds
}

func (p *PlayerBet) UpdateAccountToDB() {

	//存储到钱包服数据库
	if p.Player.AddAccount(&p.CurAccount) == false {

		//若调用不成功，则自已更新到数据库
		p.Player.AccountData.AddAccountToDB(p.Player.Id, &p.CurAccount)

	}

}

func (p *PlayerBet) SaveBetCollectToDB(gm *GameTableFtan, set int, round int, roundId string, answer *msg_struct.GameOpenAnswerDTX) {

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
	betCollect.Type = msg_struct.GameTypeFtan
	betCollect.AnswerTime = &nowTime

	betCollect.RecordId = util.GenerateUniqueID(util.RecordIdLength, p.Player.UserId)

	err := betCollect.Create()

	if err != nil {
		//存数据库不成功
	}

}

// 投注集合到数据库
func (p *PlayerBet) SaveBetCollectToDBAsync(gm *GameTableFtan, set int, round int, roundId string) {

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
	betCollect.Type = msg_struct.GameTypeFtan
	betCollect.AnswerTime = &nowTime

	betCollect.RecordId = util.GenerateUniqueID(util.RecordIdLength, p.Player.UserId)

	//启用协程队列存储
	queueId := int64(2)
	gm.AsyncDoByQueue(queueId, func() bool {

		err := betCollect.Create()

		if err != nil {
			//存数据库不成功
			log.Errorf("SaveBetCollectToDBAsync 不成功，错误:%s", err.Error())
		}
		return false

	}, nil)

}

// TongIdWin 通
//func (p *PlayerBet) TongIdWin(winner int, bet WinType, rgnSub int) WinType {
//
//	//先检测是否投中此区域内的
//	if !(rgnSub >= 0 && rgnSub < TongIdNum) {
//		return 0
//	}
//
//	var winPoint WinType
//
//	if TongWinMember[rgnSub][0] == winner || TongWinMember[rgnSub][1] == winner {
//		odds, digits := betOdds.GetBetOdds(betOdds.Bet_Ftan_Tong_Start) //得到赔率
//		if digits == 0 {
//			return 0 // 避免除零错误
//		}
//
//		betwin := bet * WinType(odds) //64位
//
//		winPoint = WinType(float64(betwin) / float64(digits)) //实际胜额
//
//		winPoint = winPoint + bet //本金加上
//
//		return winPoint
//
//	} else if TongTieMember[rgnSub] != winner {
//
//		return bet //和，退回本金
//	}
//
//	return 0
//}

// ZhengIdWin 压正
//func (p *PlayerBet) ZhengIdWin(winner int, bet WinType, rgnSub int) WinType {
//	if rgnSub < 0 || rgnSub >= TongIdNum {
//		return 0
//	}
//
//	var winPoint WinType
//
//	var bwin bool
//
//	switch rgnSub {
//	case 0:
//		bwin = (winner == 1)
//		if winner != 3 {
//			return bet // 退还本金
//		}
//	case 1:
//		bwin = (winner == 2)
//		if winner != 4 {
//			return bet
//		}
//	case 2:
//		bwin = (winner == 3)
//		if winner != 1 {
//			return bet
//		}
//	case 3:
//		bwin = (winner == 4)
//		if winner != 2 {
//			return bet
//		}
//	}
//
//	if bwin {
//		odds, digits := betOdds.GetBetOdds(betOdds.Bet_Ftan_Zheng_Start)
//		if digits == 0 {
//			return 0 // 避免除零错误
//		}
//
//		winPoint = WinType(float64(bet) * float64(odds) / float64(digits))
//
//		return winPoint + bet // 中奖金额 + 本金
//	}
//
//	return 0 // 未中奖
//}
