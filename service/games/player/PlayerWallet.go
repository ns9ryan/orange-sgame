package player

import (
	"encoding/json"
	"github.com/duanhf2012/origin/v2/log"
	"videogame/common/global/variable"
	models "videogame/common/model"
	redisId "videogame/common/redis"
)

// 玩家的钱包
type PlayerWallet struct {
	Currency string //币种

	Balance float64 //余额

	Credit uint //余额
	Digits uint //小数点数

}

// =============================================================================================
func (wallet *PlayerWallet) ResetZero() {

	wallet.Balance = 0
	wallet.Credit = 0
	wallet.Digits = 0
}

// 初始化钱包数据
func (wallet *PlayerWallet) InitWallet(currency string, credit uint, digits uint) {

	if digits == 0 {
		digits = 1
	}

	wallet.Currency = currency

	wallet.Credit = credit
	wallet.Digits = uint(digits)

	wallet.Balance = float64(credit) / float64(digits)
}

func (wallet *PlayerWallet) GetCurrency() string {

	return wallet.Currency
}

func (wallet *PlayerWallet) GetDigits() uint {

	if wallet.Digits == 0 {
		return 1
	}

	return wallet.Digits
}

func (wallet *PlayerWallet) GetCredit() uint {
	return wallet.Credit
}

func (wallet *PlayerWallet) GetBalance() float64 {
	return wallet.Balance
}

func (wallet *PlayerWallet) SetDigits(digits uint) {

	wallet.Digits = uint(digits)
}

func (wallet *PlayerWallet) SetCredit(credit uint) {
	wallet.Credit = credit
}

func (wallet *PlayerWallet) SetBalance(balance float64) {
	wallet.Balance = balance
}

func (wallet *PlayerWallet) AddBalance(balance float64) float64 {

	wallet.Balance += balance

	return wallet.Balance
}

func (wallet *PlayerWallet) AddCredit(credit uint) uint {

	wallet.Credit += credit
	return wallet.Credit
}

// 读取玩家钱包
func (wallet *PlayerWallet) ReadWalletData(userId string) bool {

	wallketKey := redisId.RedisId_PlayerWallet + userId

	//先从redis读取，看是否存在
	bytes, err := variable.RedisStorage.Get(wallketKey)

	if err == nil {

		var userWallet redisId.RedisUserWallet

		err = json.Unmarshal([]byte(bytes), &userWallet)

		if err != nil {
			//从数据库读取
		}
		wallet.InitWallet(userWallet.Currency, userWallet.Credit, userWallet.Digits)

		return true
	}

	//若不存在，则从数据库读取
	playerInfo, err1 := models.GetUserWalletByUid(userId)

	if err1 != nil {
		return false
	}
	credit := playerInfo.Balance
	digits := playerInfo.Digits
	currency := playerInfo.Currency

	wallet.InitWallet(currency, credit, digits)

	return true
}

// 存储钱包数据到redis,移到别地方处理了，此处不用
func (wallet *PlayerWallet) SaveToRedis(userId string) {

	walletKey := redisId.RedisId_PlayerWallet + userId

	var userWallet redisId.RedisUserWallet

	userWallet.UserId = userId
	userWallet.Currency = wallet.GetCurrency()
	userWallet.Credit = wallet.GetCredit()
	userWallet.Digits = wallet.GetDigits()

	bytes, err := json.Marshal(userWallet)

	if err != nil {
		return
	}

	// 尝试第一次设置
	if err = variable.RedisStorage.Set(walletKey, bytes, 0); err != nil {

		log.SError("-------->钱包服第一次写入redis失败，err:", err.Error())

		// 再尝试第二次
		if err = variable.RedisStorage.Set(walletKey, bytes, 0); err != nil {
			log.SError("-------->钱包服第二次写入redis失败，准备删除，err: ", err.Error())

			// 删除 Redis 数据
			if res, delErr := variable.RedisStorage.Del(walletKey); delErr != nil {
				log.SError("-------->钱包服删除redis失败，err:", delErr.Error())

			} else {
				if res <= 0 {
					log.SWarn("-------->钱包服删除redis失败，userId:", userWallet.UserId, " ,walletKey:", walletKey, "不存在")
				}
			}
		}
	}

}
