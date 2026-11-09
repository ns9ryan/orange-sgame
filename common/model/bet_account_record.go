package models

import (
	"encoding/json"
	"gorm.io/plugin/dbresolver"
	"time"
	"videogame/common/global/variable"
)

/******************************************
  各投注有投注明细表

******************************************/

type SBetAccountRecord struct {
	ID       uint   `gorm:"primarykey;type:int unsigned;comment:ID"`
	RecordId string `gorm:"size:32;not null;default:'';comment:投注帐目表的ID"`

	UserId string `gorm:"size:32;not null;default:'';comment:帐号的ID"`

	ParentId     string `gorm:"size:32;not null;default:'';comment:上级商户ID"`
	MerchantId   string `gorm:"size:32;not null;default:'';comment:最顶级商户ID"`
	MerchantCode string `gorm:"size:32;not null;default:'';comment:商户Code"`

	RootKey uint `gorm:"type:int unsigned;not null;default:0;comment:顶级Key"`

	GameId string `gorm:"size:30;not null;default:'';comment:游戏id"`
	DeskNo int    `gorm:"type:tinyint unsigned;not null;default:0;comment:桌号"`
	Type   string `gorm:"size:20;not null;default:'';comment:游戏类型"`

	Set   int `gorm:"type:smallint unsigned;not null;default:0;comment:桌号"`
	Round int `gorm:"type:smallint unsigned;not null;default:0;comment:局号"`

	RoundId string `gorm:"size:24;not null;default:'';comment:游戏局数全局ID"`

	Currency string `gorm:"size:20;not null;default:'';comment:货币种类"`

	BetTime *time.Time `gorm:"default:NULL;comment:投注时间"`

	BetAmount int `gorm:"type:int unsigned;not null;default:0;comment:投注额"`
	BetCount  int `gorm:"type:int unsigned;not null;default:0;comment:投注区域数"`
	WinLost   int `gorm:"type:int ;not null;default:0;comment:胜负金额"`
	Digits    int `gorm:"type:mediumint unsigned;not null;default:0;comment:投注额的小数位"`

	BeforeBalance uint `gorm:"type:int unsigned;not null;default:0;comment:投注前的余额"`
	AfterBalance  uint `gorm:"type:int unsigned;not null;default:0;comment:结算后的余额"`

	Winner int `gorm:"type:tinyint unsigned;not null;default:0;comment:开奖结果,游戏种类不同"`

	Value1 int `gorm:"type:tinyint unsigned;not null;default:0;comment:开奖参数1"`
	Value2 int `gorm:"type:tinyint unsigned;not null;default:0;comment:开奖参数2"`
	Value3 int `gorm:"type:tinyint unsigned;not null;default:0;comment:开奖参数3"`

	AnswerTime *time.Time      `gorm:"default:NULL;comment:开奖时间"`
	Answer     json.RawMessage `gorm:"type:json;comment:开奖的详细内容"`

	BaseModel
}

// TableName 重写 GORM 的默认表名
func (SBetAccountRecord) TableName() string {
	return "or_bet_account_record" // 指定自定义表名为
}

// 对象和方法的写法示例：
func (u *SBetAccountRecord) Create() error {

	return variable.GormDbMysql.Clauses(dbresolver.Write).Create(u).Error // 使用指针类型 u
}

func (u *SBetAccountRecord) Delete() error {
	return variable.GormDbMysql.Clauses(dbresolver.Write).Delete(u).Error // 使用指针类型 u
}

func (u *SBetAccountRecord) Save() error {

	return variable.GormDbMysql.Clauses(dbresolver.Write).Save(u).Error
}
