package betAreadUtil

import (
	betOdds "videogame/common/betodds"
	"videogame/common/proto/msg"
)

// 得到压注区域的高低注
func GetBetArea(areaId int, lowbet, highbet int) *msg.BetArea {

	betid := areaId

	//得到此投注区域的赔率，及小数位数
	odds, digits := betOdds.GetBetOdds(betid)

	var betarea msg.BetArea

	betarea.AreaId = uint32(betid)

	betarea.Odds = uint32(odds)
	betarea.OddsDigits = uint32(digits)

	//高低注
	//	betarea.LowBet = uint64(lowbet)
	//betarea.HighBet = uint64(highbet)

	betarea.LowBet = uint32(lowbet)
	betarea.HighBet = uint32(highbet)

	return &betarea
}
