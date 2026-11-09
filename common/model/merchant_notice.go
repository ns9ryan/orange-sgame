package models

import (
	"gorm.io/plugin/dbresolver"
	"time"
	"videogame/common/global/variable"
)

// MerchantNotice 商户公告
type MerchantNotice struct {
	ID         uint       `gorm:"primarykey;type:int unsigned;comment:ID"`
	NoticeId   string     `gorm:"not null;comment:NoticeId"`
	MerchantId string     `gorm:"not null;comment:商户号"`
	Title      string     `gorm:"not null;comment:公告标题"`
	Content    string     `gorm:"comment:公告内容"`
	StartTime  *time.Time `gorm:"comment:公告开始时间"`
	EndTime    *time.Time `gorm:"comment:公告结束时间"`
}

// TableName 重写 GORM 的默认表名
func (MerchantNotice) TableName() string {
	return "or_merchant_notice" // 指定自定义表名为
}

// 读取指定的商户
func GetMerchantNotice(merchantId string) ([]*MerchantNotice, error) {

	var data []*MerchantNotice
	now := time.Now()

	err := variable.GormDbMysql.Clauses(dbresolver.Read).
		Where("merchant_id = ?", merchantId).
		Where("start_time <= ? AND end_time >= ?", now, now).
		Order("start_time DESC").
		Find(&data).Error

	if err != nil {
		return nil, err
	}
	return data, nil
}
