package models

import (
	"gorm.io/plugin/dbresolver"
	"time"
	"videogame/common/global/variable"
)

type SUserAccount struct {
	ID uint `gorm:"primarykey;type:int unsigned;comment:ID"`

	AccountDate *time.Time `gorm:"default:NULL;comment:结算的归属日 "`
	Username    string     `gorm:"size:30;not null;comment:玩家账号,每个玩家的唯一值"`

	ParentId   string `gorm:"size:32;not null;comment:它上级的商户Username"`
	MerchantId string `gorm:"size:32;not null;comment:最上级商户表的username"`
	RootKey    uint   `gorm:"type:int unsigned;not null;default:0;comment:顶级Key"`

	Currency string `gorm:"size:20;not null;comment:货币的代码"`

	Digits           uint `gorm:"type:int unsigned;not null;comment:币种的最小单位倍数"`
	LastBet          uint `gorm:"type:int unsigned;not null;default:0;comment:上次投注金额，存储为最小单位"`
	LastBetEffective uint `gorm:"type:int unsigned;not null;default:0;comment:上次有效投注金额，存储为最小单位"`
	LastWater        uint `gorm:"type:int unsigned;not null;default:0;comment:上次洗码金额，存储为最小单位"`
	LastWinLost      int  `gorm:"type:int(10) not null;default:0;comment:上次输赢金额，存储为最小单位"`
	LastTip          uint `gorm:"type:int unsigned;not null;default:0;comment:最近一天的打赏金额，存储为最小单位"`

	StarTime *time.Time `gorm:"default:NULL;comment:结算区间起始时间"`
	EndTime  *time.Time `gorm:"default:NULL;comment:结算区间终止时间"`
}

type SumResult struct {
	Username   string `gorm:"column:username"`
	MerchantId string `gorm:"column:merchant_id"`
	Currency   string `gorm:"column:currency"`
	Digits     uint   `gorm:"column:digits"`

	LastBet          uint `gorm:"column:last_bet"`
	LastBetEffective uint `gorm:"column:last_bet_effective"`
	LastWater        uint `gorm:"column:last_water"`
	LastWinLost      int  `gorm:"column:last_win_lost"`
	LastTip          uint `gorm:"column:last_tip"`
}

// TableName 重写 GORM 的默认表名
func (SUserAccount) TableName() string {
	return "or_player_dayaccount" // 指定自定义表名为
}

func (u *SUserAccount) Save() error {

	return variable.GormDbMysql.Clauses(dbresolver.Write).Save(u).Error
}

// 对象和方法的写法示例：
func (u *SUserAccount) Create() error {
	return variable.GormDbMysql.Clauses(dbresolver.Write).Create(u).Error // 使用指针类型 u
}

func (u *SUserAccount) UserWinLostAccount(userName, startTime, endTime string) (*SumResult, error) {

	var result SumResult

	db := variable.GormDbMysql.Clauses(dbresolver.Read).Model(&SUserAccount{}).
		Select(`
			username,
			merchant_id,
			currency,
			digits,
			SUM(last_bet) as last_bet,
			SUM(last_bet_effective) as last_bet_effective,
			SUM(last_water) as last_water,
			SUM(last_win_lost) as last_win_lost,
			SUM(last_tip) as last_tip
		`).
		Where("username = ?", userName)

	// 判断 startTime 是否非空，且 endTime 非空，才添加时间过滤条件
	if startTime != "" && endTime != "" {
		db = db.Where("account_date BETWEEN ? AND ?", startTime, endTime)
	}

	err := db.Group("username, merchant_id, currency, digits").Scan(&result).Error
	if err != nil {
		return nil, err
	}

	return &result, nil
}
