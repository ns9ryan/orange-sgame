package models

import (
	//"fmt"
	//"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
	"videogame/common/global/variable"
	"time"
)

// SWalletTransPending 处理中的订单表结构体
type SWalletTransPending struct {
	ID                 int32   	`gorm:"primarykey;type:int unsigned;comment:ID"`
	UserId             string 	`gorm:"size:32;not null;comment:玩家账号,每个玩家的唯一值"`
	MerchantId         string 	`gorm:"size:32;not null;comment:玩家账号,每个玩家的唯一值"`
	WalletType		   int32    `gorm:"not null;comment:钱包类型0多钱包1单钱包"`

	TransId            string 	`gorm:"uniqueIndex;size:64;not null;comment:订单号"`
	GameId             string 	`gorm:"size:32;not null;comment:游戏SubCode"`
	DeskNo             int32   	`gorm:"not null;comment:桌号"`
	RoundId            string 	`gorm:"size:32;not null;comment:对局ID"`
	Opt                int32   	`gorm:"not null;comment:加点还是减点"`
	DataType           int32   	`gorm:"not null;comment:订单详细操作类型,0下注/1取消下注/2小费/3结算/4其他订单"`
	Point              float32	`gorm:"not null;comment:订单金额"`
	AfterBalance       float64  `gorm:"not null;comment:操作后最新总金额"`
	Reason             string 	`gorm:"size:32;comment:取消下注订单会写玩家撤销/网络异常/内部异常"`
	State			   string 	`gorm:"size:32;default:unset;comment:订单处理状态:unset未结算/set已结算/cancel已取消error异常未送达等"`
	CreateTime 		   *time.Time `gorm:"default:NULL;comment:创建时间"`
	UpdateTime 		   *time.Time `gorm:"default:NULL;comment:更新时间"`
}

// TableName 重写 GORM 的默认表名
func (SWalletTransPending) TableName() string {
	return "or_wallet_trans_pending" // 指定自定义表名为
}

func (u *SWalletTransPending) Save() error {
	return variable.GormDbMysql.Clauses(dbresolver.Write).Save(u).Error
}

func (u *SWalletTransPending) CancelBetById(transId string, reason string, balance float64, nowTime time.Time) error {

	res := variable.GormDbMysql.Model(&SWalletTransPending{}).
		Where("trans_id = ?", transId).
		Updates(map[string]interface{}{
			"state":     		"cancel",
			"reason": 	 		reason,
			"after_balance":	balance,
			"update_time": 		&nowTime,
		})

	if res.Error != nil {
		return res.Error
	}

	return nil
}

func (u *SWalletTransPending) UpdateTrans(roundId string) error {

	res := variable.GormDbMysql.Model(&SWalletTransPending{}).
		Where("round_id = ? AND state=?", roundId, "unset").
		Updates(map[string]interface{}{
			"state":     	"set",
		})

	if res.Error != nil {
		return res.Error
	}

	return nil
}

// SWalletTrans 历史订单表结构体
type SWalletTransHistory struct {
	ID                 int32   	`gorm:"primarykey;type:int unsigned;comment:ID"`
	UserId             string 	`gorm:"size:32;not null;comment:玩家账号,每个玩家的唯一值"`
	MerchantId         string 	`gorm:"size:32;not null;comment:玩家账号,每个玩家的唯一值"`
	WalletType		   int32    `gorm:"not null;comment:钱包类型0多钱包1单钱包"`

	TransId            string 	`gorm:"uniqueIndex;size:64;not null;comment:订单号"`
	GameId             string 	`gorm:"size:32;not null;comment:游戏SubCode"`
	DeskNo             int32   	`gorm:"not null;comment:桌号"`
	RoundId            string 	`gorm:"size:32;not null;comment:对局ID"`
	Opt                int32   	`gorm:"not null;comment:加点还是减点"`
	DataType           int32   	`gorm:"not null;comment:订单详细操作类型,0下注/1取消下注/2小费/3结算/4其他订单"`
	Point              float32	`gorm:"not null;comment:订单金额"`
	AfterBalance       float64  `gorm:"not null;comment:操作后最新总金额"`
	Reason             string 	`gorm:"size:32;comment:取消下注订单会写玩家撤销/网络异常/内部异常"`
	State			   string 	`gorm:"size:32;default:unset;comment:订单处理状态:unset未结算/set已结算/cancel已取消error异常未送达等"`
	CreateTime 		   *time.Time `gorm:"default:NULL;comment:创建时间"`
	UpdateTime 		   *time.Time `gorm:"default:NULL;comment:更新时间"`
}

// TableName 重写 GORM 的默认表名
func (SWalletTransHistory) TableName() string {
	return "or_wallet_trans_history" // 指定自定义表名为
}

func (u *SWalletTransHistory) Save() error {
	return variable.GormDbMysql.Clauses(dbresolver.Write).Save(u).Error
}
