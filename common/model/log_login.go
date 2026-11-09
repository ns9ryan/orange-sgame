package models

import (
	"gorm.io/plugin/dbresolver"
	"time"
	"videogame/common/global/variable"
)

// LoginLog 登录日志表结构体
type SLoginLog struct {
	ID        uint       `gorm:"primarykey;type:int unsigned;comment:ID"`
	UserType  uint       `gorm:"type:tinyint unsigned;not null;comment:用户类型:1玩家,2商户"`
	Username  string     `gorm:"size:30;not null;default:'';comment:账号"`
	LoginTime *time.Time `gorm:"default:NULL;comment:登录时间"`
	LoginIp   string     `gorm:"size:50;comment:登录IP"`
	IsOk      uint       `gorm:"type:tinyint unsigned;not null;default:0;comment:是否成功:0否,1是"`
	Reason    string     `gorm:"not null;default:'';comment:备注"`

	RootKey    uint   `gorm:"type:int unsigned;default:0;comment:顶级Key"`
	MerchantId string `gorm:"size:32;comment:对应商户表的username"`

	BaseModel
}

// TableName 重写 GORM 的默认表名
func (SLoginLog) TableName() string {
	return "or_login_log" // 指定自定义表名为
}

// 对象和方法的写法示例：
func (u *SLoginLog) Create() error {

	return variable.GormDbMysql.Clauses(dbresolver.Write).Create(u).Error // 使用指针类型 u

}
