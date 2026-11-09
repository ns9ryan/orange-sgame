package models

import (
	"gorm.io/plugin/dbresolver"
	"videogame/common/global/variable"
)

// GameTable  游戏桌台表结构体
type SGameTable struct {
	ID               uint   `gorm:"primarykey;type:int unsigned;comment:ID"`
	GameTableCode    string `gorm:"uniqueindex;size:32;not null;comment:游戏桌台标识符（如 bacx001）"`
	GameCategoryCode string `gorm:"size:32;not null;comment:游戏分类"`
	CategorySub      string `gorm:"size:32;not null;comment:游戏子分类"`
	DeskNo           uint   `gorm:"type:int unsigned;not null;comment:桌号"`
	BetInterval      uint   `gorm:"type:int unsigned;not null;comment:投注间隔时间,秒,0表示不限制"`
	UnbetRounds      uint   `gorm:"type:int unsigned;not null;comment:允许未下注局数,局,0表示不限制"`
	Status           uint   `gorm:"type:tinyint unsigned;not null;default:0;comment:是否启用:0否,1是"`
	GameServiceName  string `gorm:"size:32;not null;default:'';comment:游戏服务名"`
	StreamUrl        string `gorm:"not null;default:'';comment:视频流地址"`
	Description      string `gorm:"size:100;not null;default:'';comment:游戏说明"`
}

// TableName 重写 GORM 的默认表名
func (SGameTable) TableName() string {
	return "or_game_table" // 指定自定义表名为
}

// 对象和方法的写法示例：
func (u *SGameTable) Create() error {
	return variable.GormDbMysql.Clauses(dbresolver.Write).Create(u).Error // 使用指针类型 u
}

func (u *SGameTable) GetGameTables() ([]*SGameTable, error) {

	var tables []*SGameTable

	if err := variable.GormDbMysql.Find(&tables).Error; err != nil {
		// 错误处理
		return nil, err
	}

	return tables, nil
}

// 读取指定的荷管的数据
func (u *SGameTable) GetGameTable(gameCode string) (*SGameTable, error) {

	var tableInfo SGameTable

	result := variable.GormDbMysql.Clauses(dbresolver.Read).Where("game_table_code = ?", gameCode).First(&tableInfo) // 使用 user_id

	return &tableInfo, result.Error
}
