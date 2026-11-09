package models

import (
	"gorm.io/plugin/dbresolver"
	"videogame/common/global/variable"
)

// MerchantTemplateBetlimitArea  商户限红模板表结构体
type SMerchantTemplateBetlimitArea struct {
	ID           uint   `gorm:"primarykey;type:int unsigned;comment:ID"`
	Username     string `gorm:"size:32;not null;comment:商户账号"`
	TemplateId   string `gorm:"size:36;not null;default:'';comment:商户限红模板ID"`
	Name         string `gorm:"size:100;not null;default:'';comment:模板名称"`
	GameId       string `gorm:"size:36;not null;default:'';comment:游戏ID"`
	AreaId       string `gorm:"size:10;not null;not null;comment:区域ID"`
	MinBetAmount uint   `gorm:"type:int unsigned;not null;default:0;comment:最小投注金额"`
	MaxBetAmount uint   `gorm:"type:int unsigned;not null;default:0;comment:最大投注金额"`
	Digits       uint   `gorm:"type:int unsigned;not null;default:1;comment:投注的最小单位倍数"`
	PlanId       string `gorm:"size:36;not null;default:'';comment:限红方案ID"`

	//BaseModel
}

// TableName 重写 GORM 的默认表名
func (SMerchantTemplateBetlimitArea) TableName() string {
	return "or_merchant_template_betlimit_area" // 指定自定义表名为
}

// 对象和方法的写法示例：
func (u *SMerchantTemplateBetlimitArea) Create() error {
	return variable.GormDbMysql.Clauses(dbresolver.Write).Create(u).Error // 使用指针类型 u
}

// 读取限红模板的数据
func (u *SMerchantTemplateBetlimitArea) GetMerchantTemplateData(templateId string, gameId string) ([]SMerchantTemplateBetlimitArea, error) {

	var items []SMerchantTemplateBetlimitArea

	result := variable.GormDbMysql.Clauses(dbresolver.Read).Where("template_id = ?", templateId).Where("game_id = ?", gameId).Find(&items) // 使用 user_id

	return items, result.Error
}

// 读取模板Id
func (u *SMerchantTemplateBetlimitArea) GetMerchantTemplatePlanId(templateId string, gameId string) (string, error) {

	var item SMerchantTemplateBetlimitArea

	result := variable.GormDbMysql.Clauses(dbresolver.Read).Where("template_id = ?", templateId).Where("game_id = ?", gameId).First(&item)

	if result.Error != nil {

		return "", result.Error
	}

	return item.PlanId, result.Error

}
