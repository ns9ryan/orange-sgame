package player

// 最新的投注帐目数据
type CurAccount struct {
	LocalBet       uint //当局的投注量
	LocalBetEffice uint //当局的有效投注量
	LocalWater     uint //当局的洗码量
	LocalWin       int  //当前的胜额
}

func (account *CurAccount) AddBet(bet uint) uint {

	account.LocalBet += bet
	return account.LocalBet
}

func (account *CurAccount) AddWater(water uint) uint {
	account.LocalWater += water
	return account.LocalWater
}

func (account *CurAccount) AddWin(winlost int) int {

	account.LocalWin += winlost
	return account.LocalWin
}

func (account *CurAccount) AddBetEffective(bet uint) uint {

	account.LocalBetEffice += bet
	return account.LocalBetEffice
}

func (account *CurAccount) SetBet(bet uint) {

	account.LocalBet = bet

}

func (account *CurAccount) SetWater(water uint) {
	account.LocalWater = water

}

func (account *CurAccount) SetWin(winlost int) {

	account.LocalWin = winlost
}

func (account *CurAccount) SetBetEffective(bet uint) {

	account.LocalBetEffice = bet

}
