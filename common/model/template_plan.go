package models

import (
	"gorm.io/plugin/dbresolver"
	"videogame/common/global/variable"
)

// TemplatePlan  限红方案表结构体
type STemplatePlan struct {
	ID          uint   `gorm:"primarykey;type:int unsigned;comment:ID"`
	CategorySub string `gorm:"size:32;not null;comment:游戏子分类"`
	Name        string `gorm:"size:100;not null;default:'';comment:方案名称"`
	PlanId      string `gorm:"uniqueindex;size:36;not null;default:'';comment:限红方案ID"`
	MinValue    uint   `gorm:"not null;default:0;comment:方案的最小投注值"`
	MaxValue    uint   `gorm:"not null;default:0;comment:方案的最大投注值"`
	Digits      uint   `gorm:"type:int unsigned;not null;default:1;comment:方案投注的最小单位倍数"`
}

// TableName 重写 GORM 的默认表名
func (STemplatePlan) TableName() string {
	return "or_template_plan" // 指定自定义表名为
}

// 读取全部的数据
func (u *STemplatePlan) GetAllsData() ([]*STemplatePlan, error) {

	var datas []*STemplatePlan

	if err := variable.GormDbMysql.Find(&datas).Error; err != nil {
		// 错误处理
		return nil, err
	}

	return datas, nil
}

// 读取指定ID的数据
func (u *STemplatePlan) GetPlanData(planId string) (*STemplatePlan, error) {

	var item STemplatePlan

	result := variable.GormDbMysql.Clauses(dbresolver.Read).Where("plan_id = ?", planId).First(&item) // 使用 user_id

	return &item, result.Error
}
