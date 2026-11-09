package httpgateservice

import (
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/util/gconv"
	"time"
	models "videogame/common/model"
	"videogame/common/proto/msg"
	"videogame/common/util"
)

// 查询今天数据
func (h *HttpGateModule) queryToday(c *gin.Context, userName string, startTime, endTime time.Time) {

	userWalletModel := &models.SUserWallet{}
	userWallet, err := userWalletModel.FindOne(userName)

	if err != nil {
		util.Fail(c, util.SystemError)
		return
	}

	digits := float64(userWallet.Digits)
	if userWallet.Digits == 0 { //预防除0
		digits = float64(1)
	}

	lastBet := gconv.String(float64(userWallet.LastBet) / digits)
	lastBetEffective := gconv.String(float64(userWallet.LastBetEffective) / digits)
	lastWater := gconv.String(float64(userWallet.LastWater) / digits)
	lastWinLost := gconv.String(float64(userWallet.LastWinLost) / digits)
	lastTip := gconv.String(float64(userWallet.LastTip) / digits)

	ret := msg.UserWinLostAccountRet{
		Username:     userName,
		Bet:          lastBet,
		BetEffective: lastBetEffective,
		Water:        lastWater,
		WinLost:      lastWinLost,
		Tip:          lastTip,
		StartTime:    startTime.Format(time.DateTime),
		EndTime:      endTime.Format(time.DateTime),
	}

	if ret.Username != "" {
		util.Success(c, &ret)
	} else {
		util.Success(c, nil)
	}

}

// 查询不包含今天的数据
func (h *HttpGateModule) queryBefore(c *gin.Context, userName string, startTime, endTime time.Time) {

	userAccountModel := &models.SUserAccount{}

	sumResult, err := userAccountModel.UserWinLostAccount(userName, startTime.Format(time.DateTime), endTime.Format(time.DateTime))
	if err != nil {
		util.Fail(c, util.SystemError)
	}

	digits := float64(sumResult.Digits)
	if sumResult.Digits == 0 { //预防除0
		digits = float64(1)
	}

	lastBet := gconv.String(float64(sumResult.LastBet) / digits)
	lastBetEffective := gconv.String(float64(sumResult.LastBetEffective) / digits)
	lastWater := gconv.String(float64(sumResult.LastWater) / digits)
	lastWinLost := gconv.String(float64(sumResult.LastWinLost) / digits)
	lastTip := gconv.String(float64(sumResult.LastTip) / digits)

	ret := msg.UserWinLostAccountRet{
		Username:     sumResult.Username,
		Bet:          lastBet,
		BetEffective: lastBetEffective,
		Water:        lastWater,
		WinLost:      lastWinLost,
		Tip:          lastTip,
		StartTime:    startTime.Format(time.DateTime),
		EndTime:      endTime.Format(time.DateTime),
	}

	if ret.Username != "" {
		util.Success(c, &ret)
	} else {
		util.Success(c, nil)
	}
}

// 查询之前包含今天数据
func (h *HttpGateModule) queryRange(c *gin.Context, userName string, startTime, endTime time.Time) {

	userWalletModel := &models.SUserWallet{}
	userWallet, err := userWalletModel.FindOne(userName)
	if err != nil {
		util.Fail(c, util.SystemError)
		return
	}

	userAccountModel := &models.SUserAccount{}

	sumResult, err := userAccountModel.UserWinLostAccount(userName, startTime.Format(time.DateTime), endTime.Format(time.DateTime))
	if err != nil {
		util.Fail(c, util.SystemError)
	}

	digits := float64(userWallet.Digits)
	if sumResult.Digits == 0 { //预防除0
		digits = float64(1)
	}

	lastBet := gconv.String(float64(userWallet.LastBet+sumResult.LastBet) / digits)
	lastBetEffective := gconv.String(float64(userWallet.LastBetEffective+sumResult.LastBetEffective) / digits)
	lastWater := gconv.String(float64(userWallet.LastWater+sumResult.LastWater) / digits)
	lastWinLost := gconv.String(float64(userWallet.LastWinLost+sumResult.LastWinLost) / digits)
	lastTip := gconv.String(float64(userWallet.LastTip+sumResult.LastTip) / digits)

	ret := msg.UserWinLostAccountRet{
		Username:     userWallet.Username,
		Bet:          lastBet,
		BetEffective: lastBetEffective,
		Water:        lastWater,
		WinLost:      lastWinLost,
		Tip:          lastTip,
		StartTime:    startTime.Format(time.DateTime),
		EndTime:      endTime.Format(time.DateTime),
	}

	if ret.Username != "" {
		util.Success(c, &ret)
	} else {
		util.Success(c, nil)
	}

}
