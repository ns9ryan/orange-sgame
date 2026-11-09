package player

import (
	"encoding/json"
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"videogame/common/global/variable"
	models "videogame/common/model"
	redisId "videogame/common/redis"
)

// 玩家的帐目
type PlayerAccount struct {
	TotalBet          uint //总投注量
	TotalBetEffective uint //有效投注

	TotalWater uint //总洗码量
	TotalWin   int

	LastBet          uint
	LastBetEffective uint //有效投注
	LastWater        uint
	LastWin          int
}

//==================================================================================================================================

// 读取Redis中的数据,现在这个功能暂时没有什么
func (account *PlayerAccount) ReadAccountData(userId string) bool {

	accountKey := redisId.RedisId_PlayerAccount + userId

	//先从redis读取，看是否存在
	bytes, err := variable.RedisStorage.Get(accountKey)

	if err == nil {
		var redisAccount redisId.RedisUserAccount

		err = json.Unmarshal([]byte(bytes), &redisAccount)

		if err != nil {
			//从数据库读取
		}

		account.TotalWin = redisAccount.TotalWin
		account.TotalBet = redisAccount.TotalBet
		account.TotalBetEffective = redisAccount.TotalBetEffective
		account.TotalWater = redisAccount.TotalWater
		account.LastWin = redisAccount.LastWin

		return true
	}

	//若不存在，则从数据库读取
	accountInfo, err1 := models.GetUserWalletByUid(userId)

	if err1 != nil {
		return false
	}

	account.TotalBetEffective = accountInfo.TotalBetEffective
	account.TotalWin = accountInfo.TotalWinLost
	account.TotalWater = accountInfo.TotalWater
	account.TotalBet = accountInfo.TotalBet

	account.LastBetEffective = accountInfo.LastBetEffective
	account.LastBet = accountInfo.LastBet
	account.LastWin = accountInfo.LastWinLost
	account.LastWater = accountInfo.LastWater

	account.SaveToReids(userId)

	return true
}

// 存帐号到Redis中的数据,现在这个功能暂时没有什么
func (account *PlayerAccount) SaveToReids(userId string) bool {

	var redisAccount redisId.RedisUserAccount

	// 读取玩家钱包
	redisAccount.UserId = userId

	redisAccount.TotalWin = account.TotalWin
	redisAccount.TotalBet = account.TotalBet
	redisAccount.TotalWater = account.TotalWater
	redisAccount.TotalBetEffective = account.TotalBetEffective

	redisAccount.LastWin = account.LastWin
	redisAccount.LastBet = account.LastBet
	redisAccount.LastWater = account.LastWater
	redisAccount.LastBetEffective = account.LastBetEffective

	accountKey := redisId.RedisId_PlayerAccount + userId

	bytes, err := json.Marshal(redisAccount)

	if err != nil {
		return false
	}

	err = variable.RedisStorage.Set(accountKey, bytes, 0)

	return err == nil
}

// 帐号的帐目变化
func (account *PlayerAccount) AddAccount(curAccount *CurAccount) {

	account.LastBet += curAccount.LocalBet
	account.LastBetEffective += curAccount.LocalBetEffice
	account.LastWater += curAccount.LocalWater
	account.LastWin += curAccount.LocalWin
}

// 更新到钱包服的数据库
func (account *PlayerAccount) UpdateAccountToWalletDB(curAccount *CurAccount) {

}

func (account *PlayerAccount) AddAccountToDB(userId string, curAccount *CurAccount) error {

	//更新到数据库]
	var userWallet models.SUserWallet

	err := userWallet.UpdateAccount(userId, curAccount.LocalBet, curAccount.LocalBetEffice, curAccount.LocalWater,
		curAccount.LocalWin)

	if err != nil {
		fmt.Println(err.Error())
		log.Error(err.Error())
	}

	return err
}

// 更新商户的帐目信息
func (account *PlayerAccount) AddMerchantAccount(parentId string, curAccount *CurAccount) error {

	var merchantTable models.MerchantCurrency

	err := merchantTable.UpdateMerchantAccount(parentId, curAccount.LocalBet, curAccount.LocalBetEffice,
		curAccount.LocalWater, curAccount.LocalWin)

	if err != nil {
		fmt.Println(err.Error())
		log.Error(err.Error())
	}

	return err
}
