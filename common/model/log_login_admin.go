package models

import (
	"gorm.io/plugin/dbresolver"
	"time"
	"videogame/common/global/variable"
)

// ControlLoginLog 控端登录日志表结构体
type ControlLoginLog struct {
	ID        uint       `gorm:"primarykey;type:int unsigned;comment:ID"`
	Username  string     `gorm:"size:30;not null;default:'';comment:账号"`
	LoginTime *time.Time `gorm:"default:NULL;comment:登录时间"`
	LoginIp   string     `gorm:"size:50;comment:登录IP"`
	IsOk      uint       `gorm:"type:tinyint unsigned;not null;default:0;comment:是否成功:0否,1是"`
	Reason    string     `gorm:"not null;default:'';comment:备注"`

	BaseModel
}

// TableName 重写 GORM 的默认表名
func (ControlLoginLog) TableName() string {
	return "or_control_login_log" // 指定自定义表名为
}

// 对象和方法的写法示例：
func (u *ControlLoginLog) Create() error {

	return variable.GormDbMysql.Clauses(dbresolver.Write).Create(u).Error // 使用指针类型 u

}
