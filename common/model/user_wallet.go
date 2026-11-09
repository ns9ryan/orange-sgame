package models

import (
	"fmt"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
	"videogame/common/global/variable"
)

// PlayerWallet 玩家钱包表结构体
type SUserWallet struct {
	ID                 uint   `gorm:"primarykey;type:int unsigned;comment:ID"`
	MerchantId         string `gorm:"size:32;not null;comment:对应商户表的username"`
	Currency           string `gorm:"type:varchar(10);not null;comment:币种"`
	MerchantCurrencyId string `gorm:"size:50;not null;comment:商户币种唯一标识"`
	Username           string `gorm:"uniqueindex;size:30;not null;comment:玩家账号,每个玩家的唯一值"`
	Digits             uint   `gorm:"type:int unsigned;not null;comment:币种的最小单位倍数"`
	CreditSum          uint   `gorm:"type:int unsigned;not null;default:0;comment:累计充值额度，存储为最小单位"`
	Balance            uint   `gorm:"type:int unsigned;not null;default:0;comment:账户余额，存储为最小单位"`
	Tip                uint   `gorm:"type:int unsigned;not null;default:0;comment:总打赏金额，存储为最小单位"`
	TotalBet           uint   `gorm:"type:int unsigned;not null;default:0;comment:总投注金额，存储为最小单位"`
	TotalBetEffective  uint   `gorm:"type:int unsigned;not null;default:0;comment:总有效投注金额，存储为最小单位"`
	TotalWater         uint   `gorm:"type:int unsigned;not null;default:0;comment:总洗码金额，存储为最小单位"`
	TotalWinLost       int    `gorm:"type:int(10) not null;default:0;comment:总输赢金额，存储为最小单位"`
	LastBet            uint   `gorm:"type:int unsigned;not null;default:0;comment:上次投注金额，存储为最小单位"`
	LastBetEffective   uint   `gorm:"type:int unsigned;not null;default:0;comment:上次有效投注金额，存储为最小单位"`
	LastWater          uint   `gorm:"type:int unsigned;not null;default:0;comment:上次洗码金额，存储为最小单位"`
	LastWinLost        int    `gorm:"type:int(10) not null;default:0;comment:上次输赢金额，存储为最小单位"`
	LastTip            uint   `gorm:"type:int unsigned;not null;default:0;comment:上次打赏金额，存储为最小单位"`

	//BaseModel
}

// TableName 重写 GORM 的默认表名
func (SUserWallet) TableName() string {
	return "or_player_wallet" // 指定自定义表名为
}

func (u *SUserWallet) Save() error {
	return variable.GormDbMysql.Clauses(dbresolver.Write).Save(u).Error
}

func GetUserWalletByUid(uid string) (*SUserWallet, error) {

	var user SUserWallet

	result := variable.GormDbMysql.Clauses(dbresolver.Read).Where("UserName = ?", uid).First(&user) // 使用 user_id

	return &user, result.Error
}

func (u *SUserWallet) UpdateBalance(uid string, balance uint64) error {

	res := variable.GormDbMysql.Model(&SUserWallet{}).Where("Username=?", uid).Update("Balance", balance)

	if res.Error != nil {
		return fmt.Errorf("update wallet balance %s ,err:%s ", uid, res.Error)
	}

	return nil

}

func (u *SUserWallet) UpdateAccount(userId string, lastbet uint, lastbeteff uint, lastwater uint, lastwin int) error {

	err := variable.GormDbMysql.Model(&SUserWallet{}).
		Where("username = ?", userId).
		Updates(map[string]interface{}{
			"total_bet":           gorm.Expr("total_bet + ?", lastbet),
			"total_bet_effective": gorm.Expr("total_bet_effective + ?", lastbeteff),
			"total_water":         gorm.Expr("total_water + ?", lastwater),
			"total_win_lost":      gorm.Expr("total_win_lost + ?", lastwin),
			"last_bet":            gorm.Expr("last_bet + ?", lastbet),
			"last_bet_effective":  gorm.Expr("last_bet_effective + ?", lastbeteff),
			"last_water":          gorm.Expr("last_water + ?", lastwater),
			"last_win_lost":       gorm.Expr("last_win_lost + ?", lastwin),
		}).Error

	return err

}

// 更新今天的打赏金额
func (u *SUserWallet) UpdateTip(userId string, lasttip uint) error {

	err := variable.GormDbMysql.Model(&SUserWallet{}).
		Where("username = ?", userId).
		Updates(map[string]interface{}{
			"tip":      gorm.Expr("tip + ?", lasttip),
			"last_tip": gorm.Expr("last_tip + ?", lasttip),
		}).Error

	return err

}

// 读取有投注的玩家列表
func (u *SUserWallet) GetBettedUsers() ([]*SUserWallet, error) {

	var users []*SUserWallet

	//if err := variable.GormDbMysql.Select("username, merchant_key, merchant_currency_key,  digits, last_bet, last_bet_effective, last_water, last_win_lost").Where("last_bet>0").Find(&users).Error; err != nil {

	if err := variable.GormDbMysql.Where("last_bet >0 || last_tip>0 ").Find(&users).Error; err != nil {
		// 错误处理
		return nil, err
	}

	return users, nil
}

// 重置有投注的玩家
func (u *SUserWallet) ResetBettedUsers() error {

	err := variable.GormDbMysql.Model(&SUserWallet{}).
		Where("last_bet >0 ").Updates(map[string]interface{}{
		"last_bet":           0,
		"last_bet_effective": 0,
		"last_water":         0,
		"last_win_lost":      0,
	}).Error

	return err
}

// 重置有打赏的用户
func (u *SUserWallet) ResetTipUsers() error {

	err := variable.GormDbMysql.Model(&SUserWallet{}).
		Where("last_tip >0 ").Updates(map[string]interface{}{
		"last_tip": 0,
	}).Error

	return err

}

func (u *SUserWallet) FindOne(userName string) (*SUserWallet, error) {
	var userWallet SUserWallet

	db := variable.GormDbMysql.Where("username = ?", userName)

	if err := db.First(&userWallet).Error; err != nil {
		return nil, err
	}

	return &userWallet, nil
}
