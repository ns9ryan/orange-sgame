package models

import (
	"gorm.io/plugin/dbresolver"
	"time"
	"videogame/common/global/variable"
)

type SMerchantAccount struct {
	ID uint `gorm:"primarykey;type:int unsigned;comment:ID"`

	AccountDate *time.Time `gorm:"default:NULL;comment:结算归属日期"`

	MerchantId string `gorm:"size:32;not null;comment:最上级商户表的username"`
	ParentId   string `gorm:"size:32;not null;comment:上级商户的userName"`
	RootKey    uint   `gorm:"type:int unsigned;not null;default:0;comment:顶级Key"`

	Currency string `gorm:"size:20;not null;comment:货币的代码"`
	Digits   uint   `gorm:"type:int unsigned;not null;comment:币种的最小单位倍数"`

	Bet          uint `gorm:"type:int unsigned;not null;default:0;comment:当天投注金额，存储为最小单位"`
	BetEffective uint `gorm:"type:int unsigned;not null;default:0;comment:当天有效投注金额，存储为最小单位"`
	Water        uint `gorm:"type:int unsigned;not null;default:0;comment:当天洗码金额，存储为最小单位"`
	WinLost      int  `gorm:"type:int(10) not null;default:0;comment:当天输赢金额，存储为最小单位"`
	Tip          uint `gorm:"type:int unsigned;not null;default:0;comment:当天打赏金额，存储为最小单位"`

	CommissionRate uint `gorm:"type:int unsigned;not null;default:0;comment:成数"`
	RollingRate    uint `gorm:"type:int unsigned;not null;default:0;comment:洗码比"`

	StarTime *time.Time `gorm:"default:NULL;comment:结算区间起始时间"`
	EndTime  *time.Time `gorm:"default:NULL;comment:结算区间起始时间"`
}

// TableName 重写 GORM 的默认表名
func (SMerchantAccount) TableName() string {
	return "or_merchant_dayaccount" // 指定自定义表名为
}

func (u *SMerchantAccount) Save() error {

	return variable.GormDbMysql.Clauses(dbresolver.Write).Save(u).Error
}

// 对象和方法的写法示例：
func (u *SMerchantAccount) Create() error {
	return variable.GormDbMysql.Clauses(dbresolver.Write).Create(u).Error // 使用指针类型 u
}

func (u *SMerchantAccount) Delete() error {

	return variable.GormDbMysql.Delete(u).Error

}
