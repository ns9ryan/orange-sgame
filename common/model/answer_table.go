package models

import (
	"encoding/json"
	"time"
)

type SGAnswerTable struct {
	ID      uint   `gorm:"primarykey;type:int unsigned;comment:ID"`
	GameId  string `gorm:"size:30;not null;default:'';comment:游戏id"`
	DeskNo  int    `gorm:"type:tinyint unsigned;not null;default:0;comment:桌号"`
	Type    string `gorm:"size:20;not null;default:'';comment:游戏类型"`
	Set     int    `gorm:"type:smallint unsigned;not null;default:0;comment:桌号"`
	Round   int    `gorm:"type:smallint unsigned;not null;default:0;comment:桌号"`
	RoundId string `gorm:"size:32;not null;default:'';comment:游戏的局数ID"`
	Win     int    `gorm:"type:tinyint unsigned;not null;default:0;comment:开奖结果"`

	Value1 int `gorm:"type:tinyint unsigned;not null;default:0;comment:开奖参数1"`
	Value2 int `gorm:"type:tinyint unsigned;not null;default:0;comment:开奖参数2"`
	Value3 int `gorm:"type:tinyint unsigned;not null;default:0;comment:开奖参数3"`

	AnswerTime *time.Time `gorm:"default:NULL;comment:开奖时间"`

	Dealer string          `gorm:"size:30;not null;default:'';comment:开奖的荷管"`
	Cards  json.RawMessage `gorm:"type:json;comment:牌的内容"`

	BaseModel
}
