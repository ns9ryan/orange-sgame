package models

import (
	"gorm.io/plugin/dbresolver"
	"time"
	"videogame/common/global/variable"
)

/******************************************
  各种打赏明细表
******************************************/

type STipRecord struct {
	ID       uint   `gorm:"primarykey;type:int unsigned;comment:ID"`
	RecordId string `gorm:"size:32;not null;default:'';comment:打赏表的ID"`

	UserName     string `gorm:"size:32;not null;default:'';comment:帐号的ID"`
	ParentId     string `gorm:"size:32;not null;default:'';comment:上级的ID"`
	MerchantId   string `gorm:"size:32;not null;default:'';comment:商户ID"`
	MerchantCode string `gorm:"size:32;not null;default:'';comment:商户码"`

	RootKey uint `gorm:"type:int unsigned;not null;default:0;comment:顶级Key"`

	GameId string `gorm:"size:30;not null;default:'';comment:游戏id"`
	DeskNo int    `gorm:"type:tinyint unsigned;not null;default:0;comment:桌号"`

	Type string `gorm:"size:20;not null;default:'';comment:游戏类型"`

	SubType string `gorm:"size:20;not null;default:'';comment:游戏类型二级"`
	RoundId string `gorm:"size:24;not null;default:'';comment:游戏局数全局ID"`

	Currency string `gorm:"size:20;not null;default:'';comment:货币种类"`

	TipAmount int `gorm:"type:int unsigned;not null;default:0;comment:打赏金额"`
	Digits    int `gorm:"type:mediumint unsigned;not null;default:0;comment:投注额的小数位"`

	DealerId string `gorm:"size:30;not null;default:'';comment:打赏给哪个荷官"`

	TipTime *time.Time `gorm:"default:NULL;comment:打赏时间"`
}

// TableName 重写 GORM 的默认表名
func (STipRecord) TableName() string {
	return "or_tip_record" // 指定自定义表名为
}

// 对象和方法的写法示例：
func (u *STipRecord) Create() error {

	return variable.GormDbMysql.Clauses(dbresolver.Write).Create(u).Error // 使用指针类型 u
}

func (u *STipRecord) Delete() error {

	return variable.GormDbMysql.Clauses(dbresolver.Write).Delete(u).Error // 使用指针类型 u
}

func (u *STipRecord) Save() error {

	return variable.GormDbMysql.Clauses(dbresolver.Write).Save(u).Error
}
