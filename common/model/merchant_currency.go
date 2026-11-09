package models

import (
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
	"videogame/common/global/variable"
)

//
//// MerchantCurrency 商户币种表结构体
//type MerchantCurrency struct {
//	ID                 uint   `gorm:"primarykey;type:int unsigned;comment:ID"`
//	Username           string `gorm:"size:32;not null;comment:商户账号"`
//	Currency           string `gorm:"type:varchar(10);not null;comment:币种"`
//	MerchantCurrencyId string `gorm:"uniqueindex;size:50;not null;comment:唯一标识"`
//
//	Digits uint `gorm:"type:int unsigned;not null;comment:币种的最小单位倍数"`
//
//	Status            uint `gorm:"type:tinyint unsigned;not null;default:1;comment:是否启用:0否,1是"`
//	TotalBet          uint `gorm:"type:int unsigned;not null;default:0;comment:总投注金额，存储为最小单位"`
//	TotalBetEffective uint `gorm:"type:int unsigned;not null;default:0;comment:总有效投注金额，存储为最小单位"`
//	TotalWater        uint `gorm:"type:int unsigned;not null;default:0;comment:总洗码金额，存储为最小单位"`
//	TotalWinLost      int  `gorm:"type:int(10) not null;default:0;comment:总输赢金额，存储为最小单位"`
//	BetPaused         uint `gorm:"type:tinyint unsigned;not null;default:1;comment:是否暂停投注（不开放）:0否,1是"`
//
//	CommissionRate uint `gorm:"type:int unsigned;not null;default:0;comment:成数"`
//	RollingRate    uint `gorm:"type:int unsigned;not null;default:0;comment:洗码比"`
//
//	//LimitPlayer        uint   `gorm:"type:int unsigned;not null;default:10000;comment:限制会员数量"`
//}

// MerchantCurrency 商户币种表结构体
type MerchantCurrency struct {
	ID                 uint   `gorm:"primarykey;type:int unsigned;comment:ID"`
	Username           string `gorm:"size:32;not null;comment:商户账号"`
	Currency           string `gorm:"type:varchar(10);not null;comment:币种"`
	MerchantCurrencyId string `gorm:"uniqueindex;size:50;not null;comment:唯一标识"`
	Digits             uint   `gorm:"type:int unsigned;not null;comment:币种的最小单位倍数"`
	Status             uint   `gorm:"type:tinyint unsigned;not null;default:1;comment:是否启用:0否,1是"`

	Tip               uint `gorm:"type:int unsigned;not null;default:0;comment:总打赏金额，存储为最小单位"`
	TotalBet          uint `gorm:"type:int unsigned;not null;default:0;comment:总投注金额，存储为最小单位"`
	TotalBetEffective uint `gorm:"type:int unsigned;not null;default:0;comment:总有效投注金额，存储为最小单位"`
	TotalWater        uint `gorm:"type:int unsigned;not null;default:0;comment:总洗码金额，存储为最小单位"`
	TotalWinLost      int  `gorm:"type:int(10) not null;default:0;comment:总输赢金额，存储为最小单位"`

	LastBet          uint `gorm:"type:int unsigned;not null;default:0;comment:上次投注金额，存储为最小单位"`
	LastBetEffective uint `gorm:"type:int unsigned;not null;default:0;comment:上次有效投注金额，存储为最小单位"`
	LastWater        uint `gorm:"type:int unsigned;not null;default:0;comment:上次洗码金额，存储为最小单位"`
	LastWinLost      int  `gorm:"type:int(10) not null;default:0;comment:上次输赢金额，存储为最小单位"`

	LastTip uint `gorm:"type:int unsigned;not null;default:0;comment:上次打赏金额，存储为最小单位"`

	BetPaused uint `gorm:"type:tinyint unsigned;not null;default:1;comment:是否暂停投注（不开放）:0否,1是"`

	CommissionRate uint `gorm:"type:int unsigned;not null;default:0;comment:成数"`
	RollingRate    uint `gorm:"type:int unsigned;not null;default:0;comment:洗码比"`

	LimitPlayer uint `gorm:"type:int unsigned;not null;comment:限制会员数量"`

	MaxTip uint `gorm:"not null;default:0;comment:荷官可接收的最大打赏金额，存储为最小单位"`
	MinTip uint `gorm:"not null;default:0;comment:荷官可接收的最小打赏金额，存储为最小单位"`

	TipOptions string `gorm:"not null;comment:打赏选项金额，以逗号隔开，存储为最小单位"`

	//BaseModel
}

// TableName 重写 GORM 的默认表名
func (MerchantCurrency) TableName() string {
	return "or_merchant_currency" // 指定自定义表名为
}

// 对象和方法的写法示例：
func (u *MerchantCurrency) Create() error {
	return variable.GormDbMysql.Clauses(dbresolver.Write).Create(u).Error // 使用指针类型 u
}

// 读取指定的商户的数据
func (u *MerchantCurrency) GetMerchantGroup(uid string) (*MerchantCurrency, error) {

	var user MerchantCurrency

	result := variable.GormDbMysql.Clauses(dbresolver.Read).Where("Username = ?", uid).First(&user) // 使用 user_id

	return &user, result.Error
}

// 更新商户的帐目金额
func (u *MerchantCurrency) UpdateMerchantAccount(userId string, lastbet, lastbeteff, lastwater uint, lastwin int) error {

	err := variable.GormDbMysql.Model(&MerchantCurrency{}).
		Where("username  = ?", userId).
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

// 更新商户的打赏金额
func (u *MerchantCurrency) UpdateMerchantTip(userId string, lasttip uint) error {

	err := variable.GormDbMysql.Model(&MerchantCurrency{}).
		Where("username = ?", userId).
		Updates(map[string]interface{}{
			"last_tip": gorm.Expr("last_tip + ?", lasttip),
		}).Error

	return err

}

// 清除商户当天投注大于0的帐号
func (u *MerchantCurrency) ResetZeroLastAccount() error {

	err := variable.GormDbMysql.Model(&MerchantCurrency{}).Where("last_bet >0 ").Updates(map[string]interface{}{
		"last_bet":           0,
		"last_bet_effective": 0,
		"last_water":         0,
		"last_win_lost":      0,
	}).Error

	return err
}

// 清除商户当天打赏大于0的帐号
func (u *MerchantCurrency) ResetZeroLastTip() error {

	err := variable.GormDbMysql.Model(&MerchantCurrency{}).Where("last_tip >0 ").Updates(map[string]interface{}{
		"last_tip": 0,
	}).Error

	return err
}

// 全局的读取商户的
func GetMerchantCurrencys() ([]*MerchantCurrency, error) {

	var groups []*MerchantCurrency

	if err := variable.GormDbMysql.Find(&groups).Error; err != nil {
		// 错误处理
		return nil, err
	}
	return groups, nil
}

// 读取指定的商户
func GetChildMerchantCurrencys(username string) ([]*MerchantCurrency, error) {

	var groups []*MerchantCurrency

	result := variable.GormDbMysql.Clauses(dbresolver.Read).Where("username = ?", username).Find(&groups) // 使用 user_id

	return groups, result.Error
}
