package models

import (
	"gorm.io/plugin/dbresolver"
	"videogame/common/global/variable"
)

// 抽码数据库
type SChipsConfig struct {
	ID     uint   `gorm:"primarykey;type:int unsigned;comment:ID"`
	Name   string `gorm:"size:32;not null;comment:筹码名称"`
	Value  uint   `gorm:"type:int unsigned;not null;comment:面额"`
	Digits uint   `gorm:"type:int unsigned;not null;default:1;comment:筹码的最小单位倍数"`
	Status uint   `gorm:"type:tinyint unsigned;not null;default:1;comment:是否启用:0否,1是"`
}

// TableName 重写 GORM 的默认表名
func (SChipsConfig) TableName() string {
	return "or_system_chip" // 指定自定义表名为
}

// 对象和方法的写法示例：
func (u *SChipsConfig) Create() error {
	return variable.GormDbMysql.Clauses(dbresolver.Write).Create(u).Error // 使用指针类型 u
}

func (u *SChipsConfig) GetGhipConfig() ([]*SChipsConfig, error) {

	var chipsList []*SChipsConfig

	if err := variable.GormDbMysql.Find(&chipsList).Error; err != nil {
		// 错误处理
		return nil, err
	}

	return chipsList, nil
}
