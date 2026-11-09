package models

import (
	"gorm.io/plugin/dbresolver"
	"videogame/common/global/variable"
)

// GameCategory  游戏分类表结构体
type SGameGroup struct {
	ID               uint   `gorm:"primarykey;type:int unsigned;comment:ID"`
	GameCategoryCode string `gorm:"size:32;not null;comment:分类"`
	SubCode          string `gorm:"uniqueindex;size:32;not null;comment:子分类"`
	Description      string `gorm:"size:100;not null;default:'';comment:描述"`
	Status           uint   `gorm:"type:tinyint unsigned;not null;default:1;comment:是否启用:0否,1是"`
	IsHide           uint   `gorm:"type:tinyint unsigned;not null;default:1;comment:是否稳场显示:0否,1是"`
}

// TableName 重写 GORM 的默认表名
func (SGameGroup) TableName() string {
	return "or_game_category" // 指定自定义表名为
}

// 对象和方法的写法示例：
func (u *SGameGroup) Create() error {
	return variable.GormDbMysql.Clauses(dbresolver.Write).Create(u).Error // 使用指针类型 u
}

func (u *SGameGroup) GetGameGroups() ([]*SGameGroup, error) {

	var groups []*SGameGroup

	if err := variable.GormDbMysql.Where("is_hide", 0).Find(&groups).Error; err != nil {
		// 错误处理
		return nil, err
	}

	return groups, nil
}
