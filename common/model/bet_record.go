package models

import (
	"encoding/json"
	"github.com/gogf/gf/v2/util/gconv"
	"gorm.io/plugin/dbresolver"
	"time"
	"videogame/common/betnode"
	"videogame/common/global/variable"
	"videogame/common/proto/msg"
)

/******************************************
  各投注有投注明细表

******************************************/

type SBetRecord struct {
	ID       uint   `gorm:"primarykey;type:int unsigned;comment:ID"`
	RecordId string `gorm:"size:32;not null;default:'';comment:投注明细表的ID"`

	UserId       string `gorm:"size:32;not null;default:'';comment:帐号的ID"`
	MerchantId   string `gorm:"size:32;not null;default:'';comment:商户ID"`
	MerchantCode string `gorm:"size:32;not null;default:'';comment:商户码"`
	ParentId     string `gorm:"size:32;not null;default:'';comment:上级的ID"`
	RootKey      uint   `gorm:"type:int unsigned;not null;default:0;comment:顶级Key"`

	GameId string `gorm:"size:30;not null;default:'';comment:游戏id"`

	DeskNo  int    `gorm:"type:tinyint unsigned;not null;default:0;comment:桌号"`
	Type    string `gorm:"size:20;not null;default:'';comment:游戏类型"`
	SubType string `gorm:"size:20;not null;default:'';comment:游戏类型二级"`

	Set     int    `gorm:"type:smallint unsigned;not null;default:0;comment:桌号"`
	Round   int    `gorm:"type:smallint unsigned;not null;default:0;comment:局号"`
	RoundId string `gorm:"size:24;not null;default:'';comment:游戏局数全局ID"`

	Currency string `gorm:"size:20;not null;default:'';comment:货币种类"`

	BetTime *time.Time `gorm:"default:NULL;comment:投注时间"`

	BetMode int `gorm:"type:smallint unsigned;not null;default:0;comment:投注的模式"`

	BetAreaId    int `gorm:"type:mediumint  unsigned;not null;default:0;comment:投注区域ID"`
	BetAmount    int `gorm:"type:int unsigned;not null;default:0;comment:投注额"`
	BetEffective int `gorm:"type:int unsigned;not null;default:0;comment:有效投注额"`

	WinLost int `gorm:"type:int ;not null;default:0;comment:胜负金额"`
	Digits  int `gorm:"type:mediumint unsigned;not null;default:0;comment:投注额的小数位"`

	Winner int `gorm:"type:tinyint unsigned;not null;default:0;comment:开奖结果,游戏种类不同"`

	Value1 int `gorm:"type:tinyint unsigned;not null;default:0;comment:开奖参数1"`
	Value2 int `gorm:"type:tinyint unsigned;not null;default:0;comment:开奖参数2"`
	Value3 int `gorm:"type:tinyint unsigned;not null;default:0;comment:开奖参数3"`

	AnswerTime *time.Time `gorm:"default:NULL;comment:开奖时间"`

	Answer json.RawMessage `gorm:"type:json;comment:开奖的详细内容"`

	BaseModel
}

// TableName 重写 GORM 的默认表名
func (SBetRecord) TableName() string {
	return "or_bet_record" // 指定自定义表名为
}

// 对象和方法的写法示例：
func (u *SBetRecord) Create() error {
	return variable.GormDbMysql.Clauses(dbresolver.Write).Create(u).Error // 使用指针类型 u
}

func (u *SBetRecord) Delete() error {
	return variable.GormDbMysql.Clauses(dbresolver.Write).Delete(u).Error // 使用指针类型 u
}

func (u *SBetRecord) Save() error {

	return variable.GormDbMysql.Clauses(dbresolver.Write).Save(u).Error
}

func (u *SBetRecord) Copy(item *betnode.BetItem) {

	u.RecordId = item.OrkeyUid
	u.RoundId = item.Round_Id
	u.DeskNo = item.DeskNo
	u.GameId = item.GameId
	u.Set = item.Set
	u.Round = item.Round

	u.UserId = item.UserId

	u.ParentId = item.ParentId
	u.MerchantId = item.MerchantId
	u.MerchantCode = item.MerchantCode
	u.RootKey = item.RootKey

	u.Currency = item.Currency
	u.Type = item.Type
	u.SubType = item.SubType

	u.BetAreaId = item.BetAreaId
	u.BetAmount = int(item.BetPoint)
	u.BetEffective = item.EffectiveBet

	u.Digits = int(item.Digits)

	u.BetTime = &item.BetTime
	u.AnswerTime = &item.Answer_Timer

	u.WinLost = item.WinLost

	u.Winner = item.Winner
	u.Value1 = item.Val1
	u.Value2 = item.Val2
	u.Value3 = item.Val3
}

func (u *SBetRecord) GetBetRecordList(userId string, deskNo int32, page int, pageSize int, startTime string, endTime string) (ret []*SBetRecord, total int64, currentPage int, size int, err error) {
	// 保证page和pageSize有默认值
	if page < 1 {
		page = 1
	}

	if pageSize <= 0 {
		pageSize = 20 // 默认每页20条
	} else if pageSize > 100 {
		pageSize = 100 // 最多每页100条
	}

	db := variable.GormDbMysql.Clauses(dbresolver.Read).Model(&SBetRecord{})

	// 加上时间条件（注意startTime、endTime如果传空，就不加）
	if startTime != "" {
		db = db.Where("bet_time >= ?", startTime)
	}
	if endTime != "" {
		db = db.Where("bet_time <= ?", endTime)
	}

	// 先查总数
	db.Where("user_id = ? and desk_no = ?", userId, deskNo).Count(&total)

	var list []*SBetRecord

	// 再查分页数据
	offset := (page - 1) * pageSize
	db.Where("user_id = ? and desk_no = ?", userId, deskNo).
		Order("bet_time DESC"). // 按照CreatedTime降序排列
		Limit(pageSize).
		Offset(offset).
		Find(&list)

	return list, total, page, pageSize, db.Error
}

func (u *SBetRecord) ToBetRecordProto(areaName string) (betRecord *msg.BetRecord) {
	var answerStr string
	if u.Answer != nil {
		answerStr = string(u.Answer)
	} else {
		answerStr = ""
	}

	var betTimeStr, answerTimeStr string
	if u.BetTime != nil {
		betTimeStr = u.BetTime.Format(time.DateTime)
	}
	if u.AnswerTime != nil {
		answerTimeStr = u.AnswerTime.Format(time.DateTime)
	}

	digits := float64(u.Digits)

	if u.Digits == 0 { //预防除0
		digits = float64(1)
	}

	betAmount := gconv.String(float64(u.BetAmount) / digits)
	betEffective := gconv.String(float64(u.BetEffective) / digits)
	winLost := gconv.String(float64(u.WinLost) / digits)

	betRecord = &msg.BetRecord{
		Id:           int32(u.ID),
		Set:          gconv.String(u.Set),
		Round:        gconv.String(u.Round),
		BetAreaName:  areaName,
		BetAmount:    betAmount,
		BetEffective: betEffective,
		WinLost:      winLost,
		Winner:       gconv.String(u.Winner),
		Answer:       answerStr,
		AnswerTime:   answerTimeStr,
		SubType:      u.SubType,
		BetTime:      betTimeStr,
		Type:         u.Type,
	}

	return betRecord
}
